package sched

import (
	"fmt"
	"go.uber.org/goleak"
	"os"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"
)

var event sync.Map
var timeout time.Duration

var config *Config
var cancel chan struct{}

const (
	debugSched = true
)

func init() {
	config = NewConfig()
	cancel = make(chan struct{})
	timeout = time.Second * 2
}

func (c *Config) waiter(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if w, ok := config.waitmap[i]; ok {
		return w
	}
	return 0
}

func (c *Config) sender(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if w, ok := config.sendmap[i]; ok {
		return w
	}
	return 0
}

func ParsePair(s string) map[uint64]uint64 {
	// TODO
	res := make(map[uint64]uint64, 0)
	var prev, next uint64
	for {
		left := strings.Index(s, "(")
		right := strings.Index(s, ")")
		if left < right {
			_, err := fmt.Sscanf(s[left:right+1], "(%v,%v)", &prev, &next)
			if err == nil {
				res[prev] = next
			}
		}
		if right+1 < len(s) {
			s = s[right+1 : len(s)]
		} else {
			break
		}
	}
	return res
}

func ParseInput() {
	str_pairs := os.Getenv("Input")
	if str_pairs == "" {
		return
	}
	pairs := ParsePair(str_pairs)
	config.mu.Lock()
	defer config.mu.Unlock()
	for prev, next := range pairs {
		config.sendmap[prev] = next
		config.waitmap[next] = prev
	}
}

func SetTimeout(s int) {
	timeout = time.Second * time.Duration(s)
}

/*
TODO : change the bitmap, to statisfy real project
1. about fragment, we need a map to map id to its search space, if we want to use current fuzzer, the space should no bigger than 64
*/
func InstChBF[T any | chan T | <-chan T | chan<- T](id uint64, o T) {
	sid := config.sender(id)
	if sid == 0 {
		return
	}
	for {
		if _, ok := event.LoadAndDelete(sid); ok {
			return
		}
		select {
		case <-cancel:
			return
		default:

		}
	}
	return
}

func InstChAF[T any | chan T | <-chan T | chan<- T](id uint64, o T) {
	if debugSched {
		print("[FB] chan: obj=", o, "; id=", id, ";\n")
	}
	wid := config.waiter(id)
	if wid == 0 {
		return
	}
	event.Store(id, struct{}{})
}

func InstMutexBF(id uint64, o any) {
	sid := config.sender(id)
	if sid == 0 {
		return
	}
	for {
		if _, ok := event.LoadAndDelete(sid); ok {
			return
		}
		select {
		case <-cancel:
			return
		default:

		}
	}
	return
}

func InstMutexAF(id uint64, o any) {
	if debugSched {
		var islocked int
		var locked bool
		var mid uint64
		switch mu := o.(type) {
		case *sync.Mutex:
			locked = mu.IsLocked()
			mid = mu.ID()
		case *sync.RWMutex:
			locked = mu.IsLocked()
			mid = mu.ID()
		}
		if locked {
			islocked = 1
		} else {
			islocked = 0
		}
		print("[FB] mutex: obj=", mid, "; id=", id, "; locked=", islocked, "; gid=", runtime.Goid(), "\n")
	}
	wid := config.waiter(id)
	if wid == 0 {
		return
	}
	event.Store(id, struct{}{})
}

func GetDone() chan struct{} {
	return make(chan struct{})
}

func GetTimeout() <-chan time.Time {
	return time.After(timeout)
}

func Done(ch chan struct{}) {
	close(ch)
}

func Leakcheck(t *testing.T) {
	close(cancel)
	time.Sleep(timeout)
	baseCheck(t)
}

func baseCheck(t *testing.T) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("testing.(*F).Fuzz.func1"),
		goleak.IgnoreTopFunction("testing.runFuzzTests"),
		goleak.IgnoreTopFunction("testing.runFuzzing"),
		goleak.IgnoreTopFunction("os/signal.NotifyContext.func1"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstMutexBF"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstMutexAF"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstChBF[...]"),
		goleak.IgnoreTopFunction("toolkit/pkg/sched.InstChAF[...]"),
		goleak.IgnoreTopFunction("go.uber.org/goleak/internal/stack.getStackBuffer"),
	}
	goleak.VerifyNone(t, opts...)
}
