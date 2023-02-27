package sched

import (
	goleak "go.uber.org/goleak"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"
	"toolkit/pkg/log"
)

var event uint32
var enablesend uint64
var enablewait uint64
var istraced uint64
var timeout time.Duration

type IDMap struct {
	m  map[uint64]uint64
	mu sync.RWMutex
}

type LogEntry struct {
	opid uint64
	eid  uint32
	addr uint64
}

var idmap IDMap

const (
	debugSched = true
)

func (idm *IDMap) get(i uint64) uint64 {
	idm.mu.RLock()
	defer idm.mu.RUnlock()
	if _, ok := idm.m[i]; !ok {
		return i & ((uint64(1) << 32) - 1)
	}
	return idm.m[i]
}

func (idm *IDMap) set(mm map[uint64]uint64) {
	idm.mu.Lock()
	defer idm.mu.Unlock()
	idm.m = make(map[uint64]uint64, 0)
	for k, v := range mm {
		idm.m[k] = v
	}
}

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
	W_LOCK    = (1 << LOCK) | (1 << UNLOCK)
	W_UNLOCK  = (1 << UNLOCK) | (1 << RUNLOCK)
	S_LOCK    = 1 << LOCK
	S_UNLOCK  = 1 << UNLOCK
	S_RLOCK   = 1 << RLOCK
	S_RUNLOCK = 1 << RUNLOCK
	W_SEND    = (1 << SEND) | (1 << CLOSE)
	W_RECV    = 1 << RECV
	W_CLOSE   = (1 << CLOSE) | (1 << RECV)
	S_SEND    = 1 << SEND
	S_RECV    = 1 << RECV
	S_CLOSE   = (1 << CLOSE) | (1 << SEND)
)

func init() {
	timeout = time.Second * 2
	if debugSched {
		istraced = ^(uint64(0))
	}
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

/*
TODO : change the bitmap, to statisfy real project
1. about fragment, we need a map to map id to its search space, if we want to use current fuzzer, the space should no bigger than 64
*/
func InstChBF[T any](id uint64, e uint32, o chan T) {
	if debugSched {
		// print("chan: ", o, "; id :", id, ";\n")
	}
	mid := idmap.get(id)
	if istraced&(1<<mid) != 0 {
		// log the unmapped id
		log.LOG.LogWithTime(id, e, o)
	}
	if enablewait&(1<<mid) == 0 {
		return
	}
	for {
		evt := atomic.LoadUint32(&event)
		if evt&e != 0 {
			atomic.StoreUint32(&event, evt&(^e))
			break
		}
	}
}

func InstChAF[T any](id uint64, e uint32, o chan T) {
	if debugSched {
		print("[FB] chan: obj=", o, "; id=", id, ";\n")
	}
	mid := idmap.get(id)
	if istraced&(1<<mid) != 0 {
		log.LOG.LogWithTime(id, e, o)
	}
	if enablesend&(1<<mid) == 0 {
		return
	}
	evt := atomic.LoadUint32(&event)
	evt |= e
	atomic.StoreUint32(&event, evt)
}

func InstMutexBF(id uint64, e uint32, o *sync.Mutex) {
	if debugSched {
		// print("mutex: ", &o, "; id :", id, ";\n")
	}
	mid := idmap.get(id)
	if istraced&(1<<mid) != 0 {
		// log the unmapped id
		log.LOG.LogWithTime(id, e, o)
	}
	if enablewait&(1<<mid) == 0 {
		return
	}
	for {
		evt := atomic.LoadUint32(&event)
		if evt&e != 0 {
			atomic.StoreUint32(&event, evt&(^e))
			break
		}
	}
}

func InstMutexAF(id uint64, e uint32, o *sync.Mutex) {
	if debugSched {
		var islocked int
		if o.IsLocked() {
			islocked = 1
		} else {
			islocked = 0
		}
		print("[FB] mutex: obj=", &o, "; id=", id, "; locked=", islocked, "; gid=", runtime.Goid(), "\n")
	}
	mid := idmap.get(id)
	if istraced&(1<<mid) != 0 {
		log.LOG.LogWithTime(id, e, o)
	}
	if enablesend&(1<<mid) == 0 {
		return
	}
	evt := atomic.LoadUint32(&event)
	evt |= e
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
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstMutexBF"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstChBF"),
		goleak.IgnoreTopFunction("go.uber.org/goleak/internal/stack.getStackBuffer"),
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
