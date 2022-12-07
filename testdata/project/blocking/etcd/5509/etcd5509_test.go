package etcd5509

import (
	"context"
	"fmt"
	goleak "go.uber.org/goleak"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

var event uint32
var enablesend uint64
var enablewait uint64

const (
	LOCK = iota
	UNLOCK
	RLOCK
	RUNLOCK
	SEND
	RECVED
	CLOSED
	WGADD
	WGDONE
	WGWAIT
)

func init() {
}

func Leakcheck(t *testing.T) {
	if event == 0 {
		event = (1 << 32) - 1
		time.Sleep(time.Millisecond * 100)
	}
	opts := []goleak.Option{
		goleak.IgnoreTopFunction("time.Sleep"),
	}
	goleak.VerifyNone(t, opts...)
}

func WE(id int, e int) {
	if enablewait&(1<<id) == 0 {
		return
	}
	for {
		evt := atomic.LoadUint32(&event)
		if evt&(1<<e) != 0 {
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

var ErrConnClosed error

type Client struct {
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

func init() {
	enablesend = 0
	enablewait = 0
	event = 0
	i := rand.Int() % 7
	j := rand.Int() % 7
	if i == j {
		return
	}
	i = 3
	j = 2
	enablewait |= uint64(1 << j)
	enablesend |= uint64(1 << i)
}

func (c *Client) Close() {
	WE(0, LOCK)
	c.mu.Lock()
	SE(0, LOCK)
	defer c.mu.Unlock()
	if c.cancel == nil {
		return
	}
	c.cancel()
	c.cancel = nil
	WE(1, UNLOCK)
	c.mu.Unlock()
	SE(1, UNLOCK)
	WE(2, LOCK|RLOCK)
	c.mu.Lock() // block here
	SE(2, LOCK|RLOCK)
}

type remoteClient struct {
	client *Client
	mu     sync.Mutex
}

func (r *remoteClient) acquire(ctx context.Context) error {
	for {
		WE(3, RLOCK)
		r.client.mu.RLock()
		SE(3, RLOCK)
		closed := r.client.cancel == nil
		WE(4, LOCK)
		r.mu.Lock()
		SE(4, LOCK)
		WE(5, UNLOCK)
		r.mu.Unlock()
		SE(5, UNLOCK)
		if closed {
			return ErrConnClosed // Missing RUnlock before return
		}
		WE(6, RUNLOCK)
		r.client.mu.RUnlock()
		SE(6, RUNLOCK)
	}
}

type kv struct {
	rc *remoteClient
}

func (kv *kv) Get(ctx context.Context) error {
	return kv.Do(ctx)
}

func (kv *kv) Do(ctx context.Context) error {
	for {
		err := kv.do(ctx)
		if err == nil {
			return nil
		}
		return err
	}
}

func (kv *kv) do(ctx context.Context) error {
	err := kv.getRemote(ctx)
	return err
}

func (kv *kv) getRemote(ctx context.Context) error {
	return kv.rc.acquire(ctx)
}

type KV interface {
	Get(ctx context.Context) error
	Do(ctx context.Context) error
}

func NewKV(c *Client) KV {
	return &kv{rc: &remoteClient{
		client: c,
	}}
}
func TestEtcd5509(t *testing.T) {
	defer Leakcheck(t)
	done := make(chan int)
	go func() {
		defer close(done)
		ctx, cancel := context.WithCancel(context.TODO())
		cli := &Client{
			ctx:    ctx,
			cancel: cancel,
		}
		kv := NewKV(cli)
		donec := make(chan struct{})
		go func() {
			defer close(donec)
			err := kv.Get(context.TODO())
			if err != nil && err != ErrConnClosed {
				fmt.Println("Expect ErrConnClosed")
			}
		}()

		cli.Close()

		<-donec
	}()

	tchan := time.After(time.Second * 2)
	select {
	case <-tchan:
	case <-done:
	}
}
