package sched

import (
	"bufio"
	"fmt"
	"go.uber.org/goleak"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var event sync.Map
var timeout, recovertimeout time.Duration

var config *Config
var cancel chan struct{}
var once sync.Once

const (
	debugSched  = true
	ignoreOrder = false
)

func init() {
	config = NewConfig()
	cancel = make(chan struct{})
	timeout = time.Second * 20
	recovertimeout = time.Second * 2
	if s := os.Getenv("TIMEOUT"); s != "" {
		t, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			timeout = time.Duration(t) * time.Second
		}
	}
	if s := os.Getenv("RECOVER_TIMEOUT"); s != "" {
		t, err := strconv.ParseInt(s, 10, 32)
		if err == nil {
			recovertimeout = time.Duration(t) * time.Millisecond
		}
	}
}

func (c *Config) findPrev(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if int(c.oidx) < len(c.orders) && i == c.orders[c.oidx][1] {
		return c.orders[c.oidx][0]
	}
	return 0
}

func (c *Config) findNext(i uint64) uint64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if int(c.oidx) < len(c.orders) && i == c.orders[c.oidx][0] {
		return c.orders[c.oidx][1]
	}
	return 0
}

func ParsePair(s string) {
	// TODO
	config.mu.Lock()
	defer config.mu.Unlock()
	var prev, next uint64
	for {
		left := strings.Index(s, "(")
		right := strings.Index(s, ")")
		if left < right {
			_, err := fmt.Sscanf(s[left:right+1], "({%v}, {%v})", &prev, &next)
			if err == nil {
				config.waitmap[next] += 1
				config.orders = append(config.orders, []uint64{prev, next})
			}
		}
		if right+1 < len(s) {
			s = s[right+1 : len(s)]
		} else {
			break
		}
	}
}

func ParseInput() {
	str_pairs := os.Getenv("Input")
	if str_pairs == "" {
		return
	}
	ParsePair(str_pairs)
}

func SetTimeout(s int) {
	timeout = time.Second * time.Duration(s)
}

func (c *Config) doWait(id uint64) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if v, ok := c.waitmap[id]; ok {
		if v == 0 {
			return false
		} else {
			return true
		}
	}
	return false
}

func (c *Config) waitDec(id uint64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if v, ok := c.waitmap[id]; ok {
		if v == 0 {
			delete(config.waitmap, id)
		}
		if v > 0 {
			c.waitmap[id] -= 1
		}
	}
}

/*
TODO : change the bitmap, to statisfy real project
1. about fragment, we need a map to map id to its search space, if we want to use current fuzzer, the space should no bigger than 64
*/
func InstChBF[T any | chan T | <-chan T | chan<- T](id uint64, o T) {
	if !config.doWait(id) {
		return
	}
	pid := config.findPrev(id)
	if pid == 0 {
		return
	}
	for {
		if _, ok := event.LoadAndDelete(pid); ok {
			atomic.AddInt32(&config.oidx, 1)
			config.waitDec(id)
			fmt.Printf("[COVERED] {%v, %v}\n", pid, id)
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
	event.Store(id, struct{}{})
}

func InstMutexBF(id uint64, o any) {
	if !config.doWait(id) {
		return
	}
	pid := config.findPrev(id)
	if pid == 0 {
		return
	}
	for {
		if _, ok := event.LoadAndDelete(pid); ok {
			atomic.AddInt32(&config.oidx, 1)
			config.waitDec(id)
			fmt.Printf("[COVERED] {%v, %v}\n", pid, id)
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
	once.Do(func() {
		close(cancel)
		time.Sleep(recovertimeout)
		baseCheck(t)
	})
}

func readlines(filename string) []string {
	res := make([]string, 0)
	if _, err := os.Stat(filename); err != nil {
		return res
	}
	f, err := os.Open(filename)
	if err != nil {
		log.Fatalf("open file error: %v\n", err.Error())
		return []string{}
	}
	// remember to close the file at the end of the program
	defer f.Close()
	// read the file line by line using scanner
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		// do something with a line
		res = append(res, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return res
	}
	return res
}

func baseCheck(t *testing.T) {
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
		goleak.IgnoreTopFunction("testing.(*F).Fuzz.func1"),
		goleak.IgnoreTopFunction("testing.runFuzzTests"),
		goleak.IgnoreTopFunction("testing.runFuzzing"),
		goleak.IgnoreTopFunction("os/signal.NotifyContext.func1"),
		goleak.IgnoreTopFunction("sched.InstMutexBF"),
		goleak.IgnoreTopFunction("sched.InstMutexAF"),
		goleak.IgnoreTopFunction("sched.InstChBF[...]"),
		goleak.IgnoreTopFunction("sched.InstChAF[...]"),
		goleak.IgnoreTopFunction("sched.(*Config).findNext"),
		goleak.IgnoreTopFunction("go.uber.org/goleak/internal/stack.getStackBuffer"),
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/eth/gasprice.NewOracle.func1"),
		goleak.IgnoreTopFunction("github.com/ethereum/go-ethereum/core.(*txSenderCacher).cache"),
		goleak.IgnoreTopFunction("testing.tRunner.func1"),
	}
	others := readlines("./goleak.config")
	for _, o := range others {
		goleak.IgnoreTopFunction(o)
	}
	goleak.VerifyNone(t, opts...)
}
