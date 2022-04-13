package match

import "errors"

var (
	ErrMq               = errors.New("mq cannot nil")
	ErrTimeout          = errors.New("timeout")
	ErrClosed           = errors.New("match server closed")
	ErrOrderSide        = errors.New("order side error (buy/sell)")
	ErrOrderType        = errors.New("order type error (limit/market)")
	ErrOrderTimeInForce = errors.New("order timeInForce error (GTC/IOC/FOK)")
	ErrOrderId          = errors.New("order id error")
	ErrPair             = errors.New("pair error")
)
