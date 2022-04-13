package match

import (
	"fmt"
	"github.com/shopspring/decimal"
	"lightning-engine/internal/status"
	"lightning-engine/models"
	"lightning-engine/mq"
	"lightning-engine/utils"
	"strconv"
	"testing"
)

var (
	mp    *MatchPool
	pairs = []string{"BTC/USDT", "ETH/USDT"}
	pair  = "BTC/USDT"
)

func TestMain(m *testing.M) {
	mp, _ = NewMatchPool(status.NewStatus(), pairs, &mq.YourMq{})
	m.Run()
}

func PrintMP() {
	fmt.Println("************** orderbook **************")
	// ask
	ask := make([]string, 0)
	first := mp.pool[pair].ask.First()
	for first != nil {
		ask = append(ask, fmt.Sprintf("%s\t\t%s\n", first.Score(), first.Value().GetAmount()))
		first = first.Next(0)
	}
	for i := len(ask) - 1; i >= 0; i-- {
		fmt.Print(ask[i])
	}
	fmt.Println("--------------------")

	// bid
	first = mp.pool[pair].bid.First()
	for first != nil {
		fmt.Printf("%s\t\t%s\n", first.Score(), first.Value().GetAmount())
		first = first.Next(0)
	}
}
func TestOrderbookGTC(t *testing.T) {
	TestOrderbook_AddBidLimitGTC(t)
	TestOrderbook_AddAskLimitGTC(t)
	PrintMP()
}
func TestOrderbookIOC(t *testing.T) {
	//TestOrderbook_AddAskLimitGTC(t)
	//TestOrderbook_AddBidLimitIOC(t)

	TestOrderbook_AddBidLimitGTC(t)
	TestOrderbook_AddAskLimitIOC(t)
	PrintMP()
}
func TestOrderbookFOK(t *testing.T) {
	//TestOrderbook_AddAskLimitGTC(t)
	//TestOrderbook_AddBidLimitFOK(t)

	TestOrderbook_AddBidLimitGTC(t)
	TestOrderbook_AddAskLimitFOK(t)
	PrintMP()
}
func TestOrderbookMarket(t *testing.T) {
	//TestOrderbook_AddAskLimitGTC(t)
	//TestOrderbook_AddBidMarket(t)

	TestOrderbook_AddBidLimitGTC(t)
	TestOrderbook_AddAskMarket(t)
	PrintMP()
}
func TestOrderbook_AddBidLimitGTC(t *testing.T) {
	for i := 100; i < 120; i++ {
		order := &models.Order{
			Id:          strconv.Itoa(i),
			UserId:      1,
			Pair:        pair,
			Price:       decimal.NewFromInt(int64(i)),
			Amount:      decimal.NewFromInt(int64(i)),
			Side:        models.Buy,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		err := mp.pool[pair].Add(order)
		if err != nil {
			t.Error(err)
		}
	}
}
func TestOrderbook_AddAskLimitGTC(t *testing.T) {
	for i := 110; i < 141; i++ {
		order := &models.Order{
			Id:          strconv.Itoa(i),
			UserId:      2,
			Pair:        pair,
			Price:       decimal.NewFromInt(int64(i)),
			Amount:      decimal.NewFromInt(int64(i)),
			Side:        models.Sell,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		err := mp.pool[pair].Add(order)
		if err != nil {
			t.Error(err)
		}
	}
}
func TestOrderbook_AddBidLimitIOC(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.NewFromInt(500),
		Amount:      decimal.NewFromInt(110),
		Side:        models.Buy,
		Type:        models.Limit,
		TimeInForce: models.TimeInForceIOC,
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_AddAskLimitIOC(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.NewFromInt(110),
		Amount:      decimal.NewFromInt(110000),
		Side:        models.Sell,
		Type:        models.Limit,
		TimeInForce: models.TimeInForceIOC,
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_AddBidLimitFOK(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.NewFromInt(110),
		Amount:      decimal.NewFromInt(111),
		Side:        models.Buy,
		Type:        models.Limit,
		TimeInForce: models.TimeInForceFOK,
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_AddAskLimitFOK(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.NewFromInt(110),
		Amount:      decimal.NewFromInt(120),
		Side:        models.Sell,
		Type:        models.Limit,
		TimeInForce: models.TimeInForceFOK,
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_AddBidMarket(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.Zero,
		Amount:      decimal.NewFromInt(10),
		Side:        models.Buy,
		Type:        models.Market,
		TimeInForce: "",
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_AddAskMarket(t *testing.T) {
	order := &models.Order{
		Id:          "100",
		UserId:      1,
		Pair:        pair,
		Price:       decimal.Zero,
		Amount:      decimal.NewFromInt(10),
		Side:        models.Sell,
		Type:        models.Market,
		TimeInForce: "",
	}
	err := mp.pool[pair].Add(order)
	if err != nil {
		t.Error(err)
	}
}
func TestOrderbook_Cancel(t *testing.T) {
	for i := 100; i < 120; i++ {
		order := &models.Order{
			Id:          strconv.Itoa(i),
			UserId:      1,
			Pair:        pair,
			Price:       decimal.NewFromInt(int64(i)),
			Amount:      decimal.NewFromInt(int64(i)),
			Side:        models.Buy,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		err := mp.pool[pair].Add(order)
		if err != nil {
			t.Error(err)
		}
	}

	err := mp.pool[pair].Cancel("100")
	if err != nil {
		t.Error(err)
	}
	PrintMP()
}

// 测试撮合性能，需注释推送成交(改成了异步，所以时间很短，但不是真实时间)
func TestOrderbook_TS_ADD_MATCH(t *testing.T) {
	size := 100000
	orders := make([]models.Order, size)
	for i := 0; i < size; i++ {
		order := models.Order{
			Id:          strconv.Itoa(i),
			UserId:      1,
			Pair:        pair,
			Price:       decimal.NewFromInt(int64(i)),
			Amount:      decimal.NewFromInt(100000),
			Side:        models.Buy,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		orders[i] = order
	}
	begin := utils.NowUnixMilli()
	for _, order := range orders {
		mp.pool[pair].Add(&order)
	}
	ts1 := utils.NowUnixMilli() - begin
	t.Logf("插入%d条数据: %dms", size, ts1)

	for i := 0; i < size; i++ {
		order := models.Order{
			Id:          strconv.Itoa(i),
			UserId:      1,
			Pair:        pair,
			Price:       decimal.NewFromInt(1),
			Amount:      decimal.NewFromFloat(100000),
			Side:        models.Sell,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		orders[i] = order
	}
	begin = utils.NowUnixMilli()
	for _, order := range orders {
		mp.pool[pair].Add(&order)
	}
	ts2 := utils.NowUnixMilli() - begin
	t.Logf("撮合%d条数据: %dms", size, ts2)
}

// 测试撤单性能，需注释推送成交(改成了异步，所以时间很短，但不是真实时间)
func TestOrderbook_TS_ADD_CANCEL(t *testing.T) {
	size := 100000
	orders := make([]models.Order, size)
	for i := 0; i < size; i++ {
		order := models.Order{
			Id:          strconv.Itoa(i),
			UserId:      1,
			Pair:        pair,
			Price:       decimal.NewFromInt(int64(i)),
			Amount:      decimal.NewFromInt(100000),
			Side:        models.Buy,
			Type:        models.Limit,
			TimeInForce: models.TimeInForceGTC,
		}
		orders[i] = order
	}
	begin := utils.NowUnixMilli()
	for _, order := range orders {
		mp.pool[pair].Add(&order)
	}
	ts1 := utils.NowUnixMilli() - begin
	t.Logf("插入%d条数据: %dms", size, ts1)

	begin = utils.NowUnixMilli()
	for i := 0; i < size; i++ {
		mp.pool[pair].Cancel(strconv.Itoa(i))
	}
	ts2 := utils.NowUnixMilli() - begin
	t.Logf("撤销%d条数据: %dms", size, ts2)
}
