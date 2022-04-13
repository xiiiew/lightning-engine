package utils

import (
	"strconv"
)

// GenTradeId 生成成交单id
func GenTradeId() string {
	return strconv.Itoa(int(NowUnixMilli()))
}
