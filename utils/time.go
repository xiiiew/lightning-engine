package utils

import "time"

// NowUnixMilli 当前毫秒时间戳
func NowUnixMilli() int64 {
	return time.Now().UnixMilli()
}
