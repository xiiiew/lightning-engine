package mq

import "lightning-engine/models"

// IMQ
// 消息队列接口，撮合引擎只撮合盘口订单，成交单需推送到消息队列，下游服务消费队列，并进行业务处理。
// 下游处理包括不限于：落盘成交单、落盘委托单、用户资金操作、k线处理等。
// 需根据对应项目使用的消息队列，编写对应的实现类。
type IMQ interface {
	PushTrade(...models.Trade) // 推送成交单。注：成交单包括已取消的委托单，需要特殊处理。
}
