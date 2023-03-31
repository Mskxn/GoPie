package testdata

import (
	sched "sched"
	"sync"
	"testing"
	"time"
)

type simpleTokenTTLKeeper struct {
	stopc chan struct{}
	donec chan struct{}
	mu    sync.Mutex
}

func (tm *simpleTokenTTLKeeper) stop() {
	select {
	case tm.stopc <- struct{}{}:
		sched.InstChAF(1005022347270, tm.stopc)
	case <-tm.donec:
		sched.InstChAF(1005022347271, tm.donec)
	}
	sched.InstChBF(1005022347267, tm.donec)
	<-tm.donec
	sched.InstChAF(1005022347267, tm.donec)
}

func (tm *simpleTokenTTLKeeper) run() {
	tokenTicker := time.NewTicker(1 * time.Second)
	defer func() {
		tokenTicker.Stop()
		close(tm.donec)
	}()
	for {
		select {
		case <-tokenTicker.C:
			sched.InstChAF(1005022347272, tokenTicker.C)
			sched.InstMutexBF(1005022347274, &tm.mu)
			tm.mu.Lock()
			sched.InstMutexAF(1005022347274, &tm.mu)
			sched.InstMutexBF(1005022347275, &tm.mu)
			tm.mu.Unlock()
			sched.InstMutexAF(1005022347275, &tm.mu)
		case <-tm.stopc:
			sched.InstChAF(1005022347273, tm.stopc)
			return
		}
	}
}

func TestClose(t *testing.T) {
	tm := simpleTokenTTLKeeper{
		donec: make(chan struct{}),
		stopc: make(chan struct{}),
		mu:    sync.Mutex{},
	}
	go tm.run()
	go tm.stop()
}
func TestClose_1(t *testing.T) {
	defer sched.Leakcheck(t)
	sched.ParseInput()
	done_xxx := sched.GetDone()
	timeout_xxx := sched.GetTimeout()
	go func() {
		defer sched.Done(done_xxx)
		tm := simpleTokenTTLKeeper{
			donec: make(chan struct{}),
			stopc: make(chan struct{}),
			mu:    sync.Mutex{},
		}
		go tm.run()
		go tm.stop()
	}()
	select {
	case <-timeout_xxx:
	case <-done_xxx:
	}
}
