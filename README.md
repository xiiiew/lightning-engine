# lightning-engine

一套高性能的、纯内存撮合的数字货币交易所撮合系统。lightning-engine未使用redis等其他组件来辅助撮合，而是专门针对撮合系统，实现了升序和降序排列的跳表。从性能上来讲，lightning-engine要远大于基于redis其他撮合系统。

lightning-engine通过gRPC接收请求（挂单、撤单等），挂单时只接收撮合需要的订单字段（价格、数量、类型、方向等，其他业务相关字段需上游服务缓存）。lightning-engine专注于盘口管理和撮合订单，具体的业务处理应由下游服务处理（补全订单字段落盘、成交单处理、用户资金处理等）。lightning-engine将成交单（包括撤销单）推送到MQ，下游服务订阅相应的Topic，进行相应的处理。

## 订单类型

| 订单类型 | 描述                    | lightning-engine | lightning-engine-pro |
|--------|-----------------------|------------------|------------------|
| GTC限价单 | 成交的部分立即成交，不能成交的部分挂在盘口 | 支持               | 支持               |
| IOC限价单 | 成交的部分立即成交，不能成交的部分撤销   | 支持               | 支持               |
| FOK限价单 | 若不能完全成交，则全部撤销         | 支持               | 支持               |
| 市价单    | 以对手价成交，不能成交的部分，撤销     | 支持               | 支持               |

## 接口

| 接口   | 版本  | lightning-engine | lightning-engine-pro |
|------|-----|----------------|----------------------|
| 挂单   | v1  | 支持             | 支持                   |
| 撤单   | v1  | 支持             | 支持                   |
| 查询深度 | v1  | 支持             | 支持                   |

## example使用

```shell
go run example/main.go
☁  lightning-engine [master] ⚡  go run example/main.go
2024/04/22 18:04:22 [RPC] :8080
2024/04/22 18:04:22 启动监听终端信号成功
send buy Result:{Msg:"success"} <nil>
send sell Result:{Msg:"success"} <nil>
2024/04/22 18:04:23 成交单： [{Id:1713780263144 Pair:BTC-USDT MakerId:1 TakerId:2 MakerUser:2 TakerUser:2 Price:21000 Amount:1 TakerOrderSide:sell TakerOrderType:limit TakerTimeInForce:GTC Ts:1713780263144}]
send buy Result:{Msg:"success"} <nil>
2024/04/22 18:04:24 成交单： [{Id:1713780264146 Pair:BTC-USDT MakerId:1 TakerId:2 MakerUser:2 TakerUser:2 Price:21000 Amount:1 TakerOrderSide:sell TakerOrderType:limit TakerTimeInForce:GTC Ts:1713780264146}]
send sell Result:{Msg:"success"} <nil>
send buy Result:{Msg:"success"} <nil>
2024/04/22 18:04:25 成交单： [{Id:1713780265147 Pair:BTC-USDT MakerId:1 TakerId:2 MakerUser:2 TakerUser:2 Price:21000 Amount:1 TakerOrderSide:sell TakerOrderType:limit TakerTimeInForce:GTC Ts:1713780265147}]
send sell Result:{Msg:"success"} <nil>
send buy Result:{Msg:"success"} <nil>
2024/04/22 18:04:26 成交单： [{Id:1713780266149 Pair:BTC-USDT MakerId:1 TakerId:2 MakerUser:2 TakerUser:2 Price:21000 Amount:1 TakerOrderSide:sell TakerOrderType:limit TakerTimeInForce:GTC Ts:1713780266149}]
send sell Result:{Msg:"success"} <nil>
^C2024/04/22 18:04:26 handle signal: interrupt
2024/04/22 18:04:26 正在安全退出服务...
2024/04/22 18:04:26 安全退出完成
```


## Contact Us

个人邮箱: xiiiew@qq.com