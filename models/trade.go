package models

type Trade struct {
	Id               string `json:"i"`  // 成交单id
	Pair             string `json:"P"`  // 交易对
	MakerId          string `json:"mi"` // maker订单id
	TakerId          string `json:"ti"` // taker订单id
	MakerUser        int64  `json:"mu"` // maker用户id
	TakerUser        int64  `json:"tu"` // taker用户id
	Price            string `json:"p"`  // 成交价
	Amount           string `json:"a"`  // 成交数量
	TakerOrderSide   string `json:"s"`  // taker订单方向 buy/sell
	TakerOrderType   string `json:"t"`  // taker订单类型 limit/market/cancel
	TakerTimeInForce string `json:"f"`  // taker订单有效时间,type为limit时才生效 GTC/IOC/FOK
	Ts               int64  `json:"ts"` // 成交时间
}
