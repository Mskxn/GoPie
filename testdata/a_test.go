package kubernetes10182

import (
	sched "sched"
	"sync"
	"testing"
)

type statusManager struct {
	podStatusesLock  sync.RWMutex
	podStatusChannel chan bool
}

func (s *statusManager) Start() {
	go func() {
		for i := 0; i < 2; i++ {
			s.syncBatch()
		}
	}()
}

func (s *statusManager) syncBatch() {
	sched.InstChBF(1005022347265, s.podStatusChannel)
	<-s.podStatusChannel
	sched.InstChAF(1005022347265, s.podStatusChannel)
	s.DeletePodStatus()
}

func (s *statusManager) DeletePodStatus() {
	sched.InstMutexBF(1005022347267, &s.podStatusesLock)
	s.podStatusesLock.Lock()
	sched.InstMutexAF(1005022347267, &s.podStatusesLock)
	defer func() {
		sched.InstMutexBF(1005022347268, &s.podStatusesLock)
		s.podStatusesLock.Unlock()
		sched.InstMutexAF(1005022347268, &s.podStatusesLock)
	}()
}

func (s *statusManager) SetPodStatus() {
	sched.InstChBF(1005022347266, s.podStatusChannel)
	s.podStatusChannel <- true
	sched.InstChAF(1005022347266, s.podStatusChannel)
	sched.InstMutexBF(1005022347269, &s.podStatusesLock)
	s.podStatusesLock.Lock()
	sched.InstMutexAF(1005022347269, &s.podStatusesLock)
	defer func() {
		sched.InstMutexBF(1005022347270, &s.podStatusesLock)
		s.podStatusesLock.Unlock()
		sched.InstMutexAF(1005022347270, &s.podStatusesLock)
	}()
}

func NewStatusManager() *statusManager {
	return &statusManager{
		podStatusChannel: make(chan bool),
	}
}

// / G1 						G2							G3
// / s.Start()
// / s.syncBatch()
// / 						s.SetPodStatus()
// / <-s.podStatusChannel
// / 						s.podStatusesLock.Lock()
// / 						s.podStatusChannel <- true
// / 						s.podStatusesLock.Unlock()
// / 						return
// / s.DeletePodStatus()
// / 													s.podStatusesLock.Lock()
// / 													s.podStatusChannel <- true
// / s.podStatusesLock.Lock()
// / -----------------------------G1,G3 deadlock----------------------------
func TestKubernetes10182(t *testing.T) {
	s := NewStatusManager()
	go s.Start()
	go s.SetPodStatus() // G2
	go s.SetPodStatus() // G3
}
func TestKubernetes10182_1(t *testing.T) {
	defer sched.Leakcheck(t)
	sched.ParseInput()
	done_xxx := sched.GetDone()
	timeout_xxx := sched.GetTimeout()
	go func() {
		defer sched.Done(done_xxx)
		s := NewStatusManager()
		go s.Start()
		go s.SetPodStatus()
		go s.SetPodStatus()
	}()
	select {
	case <-timeout_xxx:
	case <-done_xxx:
	}
}
