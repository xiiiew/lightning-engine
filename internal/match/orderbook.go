package match

import (
	"errors"
	"fmt"
	"github.com/shopspring/decimal"
	"lightning-engine/internal/status"
	"lightning-engine/models"
	"lightning-engine/mq"
	"lightning-engine/pqueue/skiplist"
	"lightning-engine/utils"
	"time"
)

// Orderbook 盘口订单簿
type Orderbook struct {
	pair string
	bid  *skiplist.SkipListDesc // bid从大到小排列
	ask  *skiplist.SkipList     // ask从小到大排列

	mBid map[string]decimal.Decimal // bid订单id对应的score
	mAsk map[string]decimal.Decimal // ask订单id对应的score

	mq       mq.IMQ
	chAdd    chan models.Order // order channel 异步顺序处理订单
	chCancel chan string       // order_id channel 异步顺序处理订单
	status   *status.Status    // 程序退出状态
}

func NewOrderbook(status *status.Status, pair string, mq mq.IMQ) (*Orderbook, error) {
	if mq == nil {
		return nil, ErrMq
	}
	bid, err := skiplist.NewSkipListDesc()
	if err != nil {
		return nil, err
	}
	ask, err := skiplist.NewSkipList()
	if err != nil {
		return nil, err
	}
	return &Orderbook{
		pair:     pair,
		bid:      bid,
		ask:      ask,
		mBid:     make(map[string]decimal.Decimal),
		mAsk:     make(map[string]decimal.Decimal),
		mq:       mq,
		chAdd:    make(chan models.Order, 1000000),
		chCancel: make(chan string, 1000000),
		status:   status,
	}, nil
}

// Add 异步挂单
func (ob *Orderbook) Add(order *models.Order) error {
	ob.status.Add(1)
	defer ob.status.Done()
	select {
	case ob.chAdd <- *order:
		return nil
	case <-time.After(time.Second):
		return ErrTimeout
	case <-ob.status.Context().Done():
		return ErrClosed
	}
}

// Cancel 异步撤单
func (ob *Orderbook) Cancel(id string) error {
	ob.status.Add(1)
	defer ob.status.Done()
	select {
	case ob.chCancel <- id:
		return nil
	case <-time.After(time.Second):
		return ErrTimeout
	case <-ob.status.Context().Done():
		return ErrClosed
	}
}

// Begin 开始撮合
func (ob *Orderbook) Begin() {
	defer ob.status.Done()
	for {
		select {
		case order := <-ob.chAdd:
			ob.add(order)
		case orderId := <-ob.chCancel:
			ob.cancel(orderId)
		case <-ob.status.Context().Done():
			return
		}
	}
}

// add 挂单
func (ob *Orderbook) add(order models.Order) error {
	switch order.Side {
	case models.Buy:
		return ob.addBid(order)
	case models.Sell:
		return ob.addAsk(order)
	}
	return ErrOrderSide
}

// addBid 挂bid
func (ob *Orderbook) addBid(order models.Order) error {
	switch order.Type {
	case models.Limit:
		return ob.addBidLimit(order)
	case models.Market:
		return ob.addBidMarket(order)
	}
	return ErrOrderType
}

// addAsk 挂ask
func (ob *Orderbook) addAsk(order models.Order) error {
	switch order.Type {
	case models.Limit:
		return ob.addAskLimit(order)
	case models.Market:
		return ob.addAskMarket(order)
	}
	return ErrOrderType
}

// addBidLimit 挂bid限价单
func (ob *Orderbook) addBidLimit(order models.Order) error {
	switch order.TimeInForce {
	case models.TimeInForceGTC:
		return ob.addBidLimitGTC(order)
	case models.TimeInForceIOC:
		return ob.addBidLimitIOC(order)
	case models.TimeInForceFOK:
		return ob.addBidLimitFOK(order)
	}
	return ErrOrderTimeInForce
}

// addAskLimit 挂ask限价单
func (ob *Orderbook) addAskLimit(order models.Order) error {
	switch order.TimeInForce {
	case models.TimeInForceGTC:
		return ob.addAskLimitGTC(order)
	case models.TimeInForceIOC:
		return ob.addAskLimitIOC(order)
	case models.TimeInForceFOK:
		return ob.addAskLimitFOK(order)
	}
	return ErrOrderTimeInForce
}

// addBidMarket 挂bid市价单
func (ob *Orderbook) addBidMarket(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// order.Amount > 0
	for ob.ask.First() != nil && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.ask.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // ask.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.ask.Delete(first.Score(), first.Value().GetId())
			}
		} else { // ask.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.ask.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addAskMarket 挂ask市价单
func (ob *Orderbook) addAskMarket(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// order.Amount > 0
	for ob.bid.First() != nil && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.bid.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // bid.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.bid.Delete(first.Score(), first.Value().GetId())
			}
		} else { // bid.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.bid.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addBidLimitGTC 挂bid限价GTC订单
func (ob *Orderbook) addBidLimitGTC(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// ask.first.Score <= order.Price && order.Amount > 0
	for ob.ask.First() != nil && ob.ask.First().Score().LessThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.ask.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // ask.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.ask.Delete(first.Score(), first.Value().GetId())
			}
		} else { // ask.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.ask.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		ob.bid.Insert(order.Price, &order)
		ob.mBid[order.Id] = order.Price
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addAskLimitGTC 挂ask限价GTC订单
func (ob *Orderbook) addAskLimitGTC(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// bid.first.Score >= order.Price && order.Amount > 0
	for ob.bid.First() != nil && ob.bid.First().Score().GreaterThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.bid.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // bid.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.bid.Delete(first.Score(), first.Value().GetId())
			}
		} else { // bid.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.bid.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		ob.ask.Insert(order.Price, &order)
		ob.mAsk[order.Id] = order.Price
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addBidLimitIOC 挂bid限价IOC订单
func (ob *Orderbook) addBidLimitIOC(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// ask.first.Score <= order.Price && order.Amount > 0
	for ob.ask.First() != nil && ob.ask.First().Score().LessThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.ask.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // ask.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.ask.Delete(first.Score(), first.Value().GetId())
			}
		} else { // ask.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.ask.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addAskLimitIOC 挂ask限价IOC订单
func (ob *Orderbook) addAskLimitIOC(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// bid.first.Score >= order.Price && order.Amount > 0
	for ob.bid.First() != nil && ob.bid.First().Score().GreaterThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.bid.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // bid.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.bid.Delete(first.Score(), first.Value().GetId())
			}
		} else { // bid.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.bid.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addBidLimitFOK 挂bid限价FOK订单
func (ob *Orderbook) addBidLimitFOK(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// 判断能否全部成交
	amount := order.Amount
	first = ob.ask.First()
	for first != nil && first.Score().LessThanOrEqual(order.Price) && amount.GreaterThan(decimal.Zero) {
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // ask.first.Amount >= order.Amount
			amount = amount.Sub(amount)
		} else {
			amount = amount.Sub(first.Value().GetAmount())
			first = first.Next(0)
		}
	}
	if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0, 撤销
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
		ob.PushTrades(trades...)
		return nil
	}

	// ask.first.Score <= order.Price && order.Amount > 0
	for ob.ask.First() != nil && ob.ask.First().Score().LessThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.ask.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // ask.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.ask.Delete(first.Score(), first.Value().GetId())
			}
		} else { // ask.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.ask.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// addAskLimitFOK 挂ask限价FOK订单
func (ob *Orderbook) addAskLimitFOK(order models.Order) error {
	trades := make([]models.Trade, 0)
	first := &skiplist.SkipListNode{}

	// 判断能否全部成交
	amount := order.Amount
	first = ob.bid.First()
	for first != nil && first.Score().GreaterThanOrEqual(order.Price) && amount.GreaterThan(decimal.Zero) {
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // bid.first.Amount >= order.Amount
			amount = amount.Sub(amount)
		} else {
			amount = amount.Sub(first.Value().GetAmount())
			first = first.Next(0)
		}
	}
	if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0, 撤销
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
		ob.PushTrades(trades...)
		return nil
	}

	// bid.first.Score >= order.Price && order.Amount > 0
	for ob.bid.First() != nil && ob.bid.First().Score().GreaterThanOrEqual(order.Price) && order.Amount.GreaterThan(decimal.Zero) {
		first = ob.bid.First()
		if first.Value().GetAmount().GreaterThanOrEqual(order.Amount) { // bid.first.Amount >= order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           order.Amount.String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)

			// 判断first剩余数量
			amount := first.Value().GetAmount().Sub(order.Amount)
			order.Amount = order.Amount.Sub(order.Amount)
			if amount.GreaterThan(decimal.Zero) { // 剩余数量 > 0
				first.Value().SetAmount(amount)
			} else { // 剩余数量 <= 0
				ob.bid.Delete(first.Score(), first.Value().GetId())
			}
		} else { // bid.first.Amount < order.Amount
			trade := models.Trade{
				Id:               utils.GenTradeId(),
				Pair:             order.Pair,
				MakerId:          first.Value().GetId(),
				TakerId:          order.Id,
				MakerUser:        first.Value().GetUserId(),
				TakerUser:        order.UserId,
				Price:            first.Score().String(),
				Amount:           first.Value().GetAmount().String(),
				TakerOrderSide:   order.Side,
				TakerOrderType:   order.Type,
				TakerTimeInForce: order.TimeInForce,
				Ts:               utils.NowUnixMilli(),
			}
			trades = append(trades, trade)
			order.Amount = order.Amount.Sub(first.Value().GetAmount())

			// 删除first
			ob.bid.Delete(first.Score(), first.Value().GetId())
		}
	}

	// 判断order是否完全成交
	if order.Amount.GreaterThan(decimal.Zero) {
		trade := models.Trade{
			Id:               utils.GenTradeId(),
			Pair:             order.Pair,
			MakerId:          order.Id,
			TakerId:          order.Id,
			MakerUser:        order.UserId,
			TakerUser:        order.UserId,
			Price:            order.Price.String(),
			Amount:           order.Amount.String(),
			TakerOrderSide:   order.Side,
			TakerOrderType:   models.Cancel,
			TakerTimeInForce: order.TimeInForce,
			Ts:               utils.NowUnixMilli(),
		}
		trades = append(trades, trade)
	}

	if len(trades) > 0 {
		ob.PushTrades(trades...)
	}
	return nil
}

// cancel 撤单
func (ob *Orderbook) cancel(id string) error {
	ob.status.Add(1)
	defer ob.status.Done()
	if score, ok := ob.mBid[id]; ok {
		return ob.cancelBid(score, id)
	} else if score, ok := ob.mAsk[id]; ok {
		return ob.cancelAsk(score, id)
	} else {
		return ErrOrderId
	}
}

// cancelBid 撤销bid
func (ob *Orderbook) cancelBid(score decimal.Decimal, id string) error {
	node, _ := ob.bid.Find(score, id)
	if node == nil {
		return ErrOrderId
	}

	order, ok := node.Value().(*models.Order)
	if !ok {
		return errors.New(fmt.Sprintf("node value cannot convert to Order."))
	}

	ob.bid.Delete(score, id)
	trade := models.Trade{
		Id:               utils.GenTradeId(),
		Pair:             order.Pair,
		MakerId:          order.Id,
		TakerId:          order.Id,
		MakerUser:        order.UserId,
		TakerUser:        order.UserId,
		Price:            order.Price.String(),
		Amount:           order.Amount.String(),
		TakerOrderSide:   order.Side,
		TakerOrderType:   models.Cancel,
		TakerTimeInForce: order.TimeInForce,
		Ts:               utils.NowUnixMilli(),
	}
	ob.PushTrades(trade)
	delete(ob.mBid, id)
	return nil
}

// cancelAsk 撤销ask
func (ob *Orderbook) cancelAsk(score decimal.Decimal, id string) error {
	node, _ := ob.ask.Find(score, id)
	if node == nil {
		return ErrOrderId
	}

	order, ok := node.Value().(*models.Order)
	if !ok {
		return errors.New(fmt.Sprintf("node value cannot convert to Order."))
	}

	ob.ask.Delete(score, id)
	trade := models.Trade{
		Id:               utils.GenTradeId(),
		Pair:             order.Pair,
		MakerId:          order.Id,
		TakerId:          order.Id,
		MakerUser:        order.UserId,
		TakerUser:        order.UserId,
		Price:            order.Price.String(),
		Amount:           order.Amount.String(),
		TakerOrderSide:   order.Side,
		TakerOrderType:   models.Cancel,
		TakerTimeInForce: order.TimeInForce,
		Ts:               utils.NowUnixMilli(),
	}
	ob.PushTrades(trade)
	delete(ob.mAsk, id)
	return nil
}

// PushTrades 推送成交单
func (ob *Orderbook) PushTrades(trades ...models.Trade) {
	ob.mq.PushTrade(trades...)
}
