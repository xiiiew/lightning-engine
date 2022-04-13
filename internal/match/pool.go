package match

import (
	"lightning-engine/internal/status"
	"lightning-engine/models"
	"lightning-engine/mq"
)

// MatchPool 撮合池
type MatchPool struct {
	pool map[string]*Orderbook
}

func NewMatchPool(status *status.Status, pairs []string, mq mq.IMQ) (*MatchPool, error) {
	mp := MatchPool{}
	mp.pool = make(map[string]*Orderbook)
	for _, p := range pairs {
		ob, err := NewOrderbook(status, p, mq)
		if err != nil {
			return nil, err
		}
		status.Add(1)
		go ob.Begin()
		mp.pool[p] = ob
	}
	return &mp, nil
}

// AddOrder 挂单
func (mp *MatchPool) AddOrder(order *models.Order) error {
	if _, ok := mp.pool[order.Pair]; !ok {
		return ErrPair
	}
	return mp.pool[order.Pair].Add(order)
}

// CancelOrder 撤单
func (mp *MatchPool) CancelOrder(pair string, id string) error {
	if _, ok := mp.pool[pair]; !ok {
		return ErrPair
	}
	return mp.pool[pair].Cancel(id)
}
