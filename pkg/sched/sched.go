package sched

import (
	goleak "go.uber.org/goleak"
	"sync/atomic"
	"testing"
	"time"
)

var event uint32
var enablesend uint64
var enablewait uint64
var timeout time.Duration

const (
	EMPTY = iota
	LOCK
	UNLOCK
	RLOCK
	RUNLOCK
	SEND
	RECV
	CLOSE
	WGADD
	WGDONE
	WGWAIT
)

const (
	W_LOCK    = LOCK | UNLOCK
	W_UNLOCK  = UNLOCK | RUNLOCK
	S_LOCK    = LOCK
	S_UNLOCK  = UNLOCK
	S_RLOCK   = RLOCK
	S_RUNLOCK = RUNLOCK
	W_SEND    = SEND | CLOSE
	W_RECV    = RECV
	W_CLOSE   = CLOSE
	S_SEND    = SEND
	S_RECV    = RECV
	S_CLOSE   = CLOSE
)

func init() {
	timeout = time.Second * 2
}

func SetEnableWait(en uint64) {
	enablewait = en
}

func SetEnableSend(en uint64) {
	enablesend = en
}

func SetTimeout(ms int) {
	timeout = time.Duration(ms * 1000000)
}

func WE(id int, e int) {
	if enablewait&(1<<id) == 0 {
		return
	}
	for {
		evt := atomic.LoadUint32(&event)
		if evt&(1<<e) != 0 {
			atomic.StoreUint32(&event, evt & ^(1<<e))
			break
		}
	}
}

func SE(id int, e int) {
	if enablesend&(1<<id) == 0 {
		return
	}
	evt := atomic.LoadUint32(&event)
	evt |= (1 << e)
	atomic.StoreUint32(&event, evt)
}

func Leakcheck(t *testing.T) {
	if event == 0 {
		event = (1 << 32) - 1
		time.Sleep(timeout)
	}
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("testing.(*F).Fuzz.func1"),
		goleak.IgnoreTopFunction("testing.runFuzzTests"),
		goleak.IgnoreTopFunction("testing.runFuzzing"),
		goleak.IgnoreTopFunction("os/signal.NotifyContext.func1"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.WE"),
	}
	goleak.VerifyNone(t, opts...)
}

func Check(t *testing.T) {
	tchan := time.After(timeout)
	select {
	case <-tchan:
		Leakcheck(t)
	}
}
