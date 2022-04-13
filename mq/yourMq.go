package mq

import (
	"lightning-engine/models"
	"log"
)

type YourMq struct {
}

func NewYourMq() IMQ {
	return &YourMq{}
}

func (mq *YourMq) PushTrade(trades ...models.Trade) {
	// 根据自己使用的队列，实现IMQ接口相应的方法
	log.Printf("成交单： %+v\n", trades)
}
