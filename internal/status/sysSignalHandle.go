package status

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type SysSignalHandle struct {
	status *Status
}

func NewSysSignalHandle(status *Status) *SysSignalHandle {
	return &SysSignalHandle{status: status}
}

// Begin 监听退出信号
func (h *SysSignalHandle) Begin() {
	ss := newSignalSet()

	ss.register(syscall.SIGINT, h.dealSysSignal)
	ss.register(syscall.SIGTERM, h.dealSysSignal)
	ss.register(syscall.SIGQUIT, h.dealSysSignal)

	c := make(chan os.Signal)
	var sigs []os.Signal
	for sig := range ss.m {
		sigs = append(sigs, sig)
	}
	signal.Notify(c, sigs...)

	log.Println("启动监听终端信号成功")
	for {
		sig := <-c
		if err := ss.handle(sig, nil); err != nil {
			log.Println("unknown signal received: ", sig)
		}
	}
}

// 信号处理
func (h *SysSignalHandle) dealSysSignal(s os.Signal, arg interface{}) {
	msg := fmt.Sprintf("handle signal: %v", s)
	log.Println(msg)
	log.Println("正在安全退出服务...")
	h.status.Stop()
	h.status.Wait()

	log.Println("安全退出完成")
	os.Exit(0)
}

type signalHandler func(s os.Signal, arg interface{})

type signalSet struct {
	m map[os.Signal]signalHandler
}

func newSignalSet() *signalSet {
	ss := new(signalSet)
	ss.m = make(map[os.Signal]signalHandler)
	return ss
}

func (self *signalSet) register(s os.Signal, handler signalHandler) {
	self.m[s] = handler
}

func (self *signalSet) handle(sig os.Signal, arg interface{}) error {
	if _, ok := self.m[sig]; ok {
		self.m[sig](sig, arg)
		return nil
	}
	return fmt.Errorf("no handler available for signal %v", sig)
}
