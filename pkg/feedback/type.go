package feedback

import "fmt"

const (
	Chanmake = iota
	Chansend
	Chanrecv
	Chanclose
	Lock
	Unlock
)

const (
	ChanFull = 1 << iota
	ChanEmpty
	ChanClosed
	MutexLocked
	MutexUnlocked
)

type ObjectStatus struct {
	isCh bool
	// for channel
	dataqsize int
	qcount    int
	closed    bool
	// for mutex
	locked bool
}

func (s *ObjectStatus) IsCritical() uint64 {
	if s.isCh {
		res := uint64(0)
		if s.qcount == 0 {
			res |= ChanEmpty
		}
		if s.qcount == s.dataqsize {
			res |= ChanFull
		}
		if s.closed {
			res = ChanClosed
		}
		return res
	}

	// mutex
	if s.locked {
		return MutexLocked
	}
	return MutexUnlocked
}

type OpAndStatus struct {
	opid   uint64
	oid    uint64
	gid    uint64
	typ    uint64
	status ObjectStatus
}

func (ops *OpAndStatus) ToString() string {
	var s string
	switch ops.typ {
	case Chanmake:
		s = fmt.Sprintf("make(chan, %v)", ops.status.dataqsize)
	case Chansend:
		s = fmt.Sprintf("%v: v -> chan, (%v/%v)", ops.gid, ops.status.qcount, ops.status.dataqsize)
	case Chanrecv:
		s = fmt.Sprintf("%v: <- chan, (%v/%v)", ops.gid, ops.status.qcount, ops.status.dataqsize)
	case Chanclose:
		s = fmt.Sprintf("%v: close(chan)", ops.gid)
	case Lock:
		s = fmt.Sprintf("%v: mu.lock", ops.gid)
	case Unlock:
		s = fmt.Sprintf("%v: mu.unlock", ops.gid)
	default:
	}
	return s
}
