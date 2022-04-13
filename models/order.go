package models

import (
	"github.com/shopspring/decimal"
)

const (
	Buy  = "buy"
	Sell = "sell"

	Limit  = "limit"
	Market = "market"
	Cancel = "cancel"

	TimeInForceGTC = "GTC" // 订单一直有效，知道被成交或者取消
	TimeInForceIOC = "IOC" // 无法立即成交的部分就撤销
	TimeInForceFOK = "FOK" // 无法全部立即成交就撤销
)

// Order 订单, 实现INodeValue接口，存放在节点中
type Order struct {
	Id          string          `json:"i"` // 订单id
	UserId      int64           `json:"u"` // 用户id
	Pair        string          `json:"P"` // 交易对
	Price       decimal.Decimal `json:"p"` // 价格
	Amount      decimal.Decimal `json:"a"` // 数量
	Side        string          `json:"s"` // 订单方向 buy/sell
	Type        string          `json:"t"` // 订单类型 limit/market
	TimeInForce string          `json:"f"` // 订单有效时间,type为limit时才生效 GTC/IOC/FOK
}

func (o *Order) GetId() string {
	return o.Id
}

func (o *Order) GetAmount() decimal.Decimal {
	return o.Amount
}

func (o *Order) GetUserId() int64 {
	return o.UserId
}

func (o *Order) SetAmount(d decimal.Decimal) {
	o.Amount = d
}
