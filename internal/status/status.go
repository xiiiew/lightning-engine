package status

import (
	"context"
	"sync"
)

// 服务运行状态
type Status struct {
	wg     sync.WaitGroup // 等待所有goroutine安全退出
	ctx    context.Context
	cancel func()
}

func NewStatus() *Status {
	ctx, cancel := context.WithCancel(context.Background())
	return &Status{
		wg:     sync.WaitGroup{},
		ctx:    ctx,
		cancel: cancel,
	}
}

// Context context
func (self *Status) Context() context.Context {
	return self.ctx
}

// Stop 退出程序
func (self *Status) Stop() {
	self.cancel()
}

// Add 添加退出等待
func (self *Status) Add(delta int) {
	self.wg.Add(delta)
}

// Done 减少退出等待
func (self *Status) Done() {
	self.wg.Done()
}

// Wait 等待执行完成
func (self *Status) Wait() {
	self.wg.Wait()
}
