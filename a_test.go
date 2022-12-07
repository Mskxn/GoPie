/*
 * Project: cockroach
 * Issue or PR  : https://github.com/cockroachdb/cockroach/pull/10214
 * Buggy version: 7207111aa3a43df0552509365fdec741a53f873f
 * fix commit-id: 27e863d90ab0660494778f1c35966cc5ddc38e32
 * Flaky: 3/100
 * Description: This deadlock is caused by different order when acquiring
 * coalescedMu.Lock() and raftMu.Lock(). The fix is to refactor sendQueuedHeartbeats()
 * so that cockroachdb can unlock coalescedMu before locking raftMu.
 */
package testdata

import (
	"sync"
	"testing"
	sched "toolkit/pkg/sched"
	"unsafe"
)

type Store struct {
	coalescedMu struct {
		sync.Mutex
		heartbeatResponses []int
	}
	mu struct {
		replicas map[int]*Replica
	}
}

func (s *Store) sendQueuedHeartbeats() {
	sched.WE(1, 3)
	s.coalescedMu.Lock()
	sched.SE(1, // LockA acquire
		1)
	defer s.coalescedMu.Unlock() // LockA release
	for i := 0; i < len(s.coalescedMu.heartbeatResponses); i++ {
		s.sendQueuedHeartbeatsToNode() // LockB
	}
}

func (s *Store) sendQueuedHeartbeatsToNode() {
	for i := 0; i < len(s.mu.replicas); i++ {
		r := s.mu.replicas[i]
		r.reportUnreachable() // LockB
	}
}

type Replica struct {
	raftMu sync.Mutex
	mu     sync.Mutex
	store  *Store
}

func (r *Replica) reportUnreachable() {
	sched.WE(2, 3)
	r.raftMu.Lock()
	sched. // LockB acquire
		SE(2, 1)
	//+time.Sleep(time.Nanosecond)
	defer r.raftMu.Unlock()
	// LockB release
}

func (r *Replica) tick() {
	sched.WE(3, 3)
	r.raftMu.Lock()
	sched. // LockB acquire
		SE(3, 1)
	defer r.raftMu.Unlock()
	r.tickRaftMuLocked()
	// LockB release
}

func (r *Replica) tickRaftMuLocked() {
	sched.WE(4, 3)
	r.mu.Lock()
	sched.SE(4, 1)
	defer r.mu.Unlock()
	if r.maybeQuiesceLocked() {
		return
	}
}
func (r *Replica) maybeQuiesceLocked() bool {
	for i := 0; i < 2; i++ {
		if !r.maybeCoalesceHeartbeat() {
			return true
		}
	}
	return false
}
func (r *Replica) maybeCoalesceHeartbeat() bool {
	msgtype := uintptr(unsafe.Pointer(r)) % 3
	switch msgtype {
	case 0, 1, 2:
		sched.WE(5, 3)
		r.store.coalescedMu.Lock()
		sched. // LockA acquire
			SE(5, 1)
	default:
		return false
	}
	sched.WE(6, 6)
	r.store.coalescedMu.Unlock()
	sched. // LockA release
		SE(6, 2)
	return true
}

func TestCockroach10214(t *testing.T) {
	store := &Store{}
	responses := &store.coalescedMu.heartbeatResponses
	*responses = append(*responses, 1, 2)
	store.mu.replicas = make(map[int]*Replica)

	rp1 := &Replica{
		store: store,
	}
	rp2 := &Replica{
		store: store,
	}
	store.mu.replicas[0] = rp1
	store.mu.replicas[1] = rp2

	go func() {
		store.sendQueuedHeartbeats()
	}()

	go func() {
		rp1.tick()
	}()
}
func FuzzGenCockroach10214(f *testing.F) {
	f.Add(uint64(32), uint64(64))
	f.Fuzz(func(t *testing.T, sched_wait_bitmap, sched_send_bitmap uint64) {
		defer sched.Leakcheck(t)
		sched.SetEnableWait(sched_wait_bitmap)
		sched.SetEnableSend(sched_send_bitmap)
		go func() {
			store := &Store{}
			responses := &store.coalescedMu.heartbeatResponses
			*responses = append(*responses, 1, 2)
			store.mu.replicas = make(map[int]*Replica)

			rp1 := &Replica{
				store: store,
			}
			rp2 := &Replica{
				store: store,
			}
			store.mu.replicas[0] = rp1
			store.mu.replicas[1] = rp2

			go func() {
				store.sendQueuedHeartbeats()
			}()

			go func() {
				rp1.tick()
			}()
		}()
	})
}
