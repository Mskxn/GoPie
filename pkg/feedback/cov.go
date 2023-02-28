package feedback

import (
	"fmt"
	"math"
	"sync"
	"toolkit/pkg/fuzzer"
)

type OpID uint64

type Status struct {
	/*
		F ChanFull
		E ChanEmpty
		C ChanClosed
		L MutexLocked
		U MutexUnlocked
	*/
	f uint64
	e uint64
	c uint64
	l uint64
	u uint64
}

func (s *Status) ToU64() uint64 {
	var res = uint64(0)
	res |= s.f
	res |= s.e
	res |= s.c
	res |= s.l
	res |= s.u
	return res
}

func Merge(s Status, ss Status) Status {
	return ToStatus(s.ToU64() | ss.ToU64())
}

func ToStatus(s uint64) Status {
	res := Status{}
	res.f = s & ChanFull
	res.e = s & ChanEmpty
	res.c = s & ChanClosed
	res.l = s & MutexLocked
	res.u = s & MutexUnlocked
	return res
}

func (s Status) ToString() string {
	us := s.ToU64()
	res := ""

	red := func(s string) string {
		return fmt.Sprintf("\033[1;31;40m%s\033[0m", s)
	}

	mark := func(us uint64, bitmask uint64, s string) string {
		if us&bitmask != 0 {
			return red(s)
		} else {
			return s
		}
	}

	res += mark(us, ChanFull, "F")
	res += mark(us, ChanEmpty, "E")
	res += mark(us, ChanClosed, "C")
	res += mark(us, MutexLocked, "L")
	res += mark(us, MutexUnlocked, "U")
	return res
}

type Cov struct {
	ps map[string]fuzzer.Pair
	cs map[OpID]Status
	mu sync.Mutex
}

func NewCov() Cov {
	cov := Cov{}
	cov.ps = make(map[string]fuzzer.Pair)
	cov.cs = make(map[OpID]Status)
	return cov
}

func (c *Cov) ToString() string {
	pairs := "Pairs: \n"
	cs := "Status: \n"
	score := "Score: "
	var cnt = 0
	for k, _ := range c.ps {
		pairs += fmt.Sprintf("[%v] %s\n", cnt, k)
		cnt += 1
	}
	cnt = 0
	for opid, c := range c.cs {
		cs += fmt.Sprintf("[%v] %v (%s)\n", cnt, opid, c.ToString())
		cnt += 1
	}
	score += fmt.Sprintf("%v\n", c.Score())
	return pairs + "\n" + cs + "\n" + score + "\n"
}

func (c *Cov) UpdateP(k string, pair fuzzer.Pair) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if k == "" {
		k = pair.ToString()
	}
	if _, ok := c.ps[k]; !ok {
		c.ps[k] = pair
		return true
	}
	return false
}

func (c *Cov) UpdateC(k OpID, v Status) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if st, ok := c.cs[k]; !ok {
		c.cs[k] = v
		return true
	} else {
		if st.ToU64() != v.ToU64() {
			c.cs[k] = Merge(st, v)
			return true
		}
	}
	return false
}

func (c *Cov) Merge(cc *Cov) bool {
	var interesting bool = false
	for k, v := range cc.ps {
		if c.UpdateP(k, v) {
			interesting = true
		}
	}
	for k, v := range cc.cs {
		if c.UpdateC(k, v) {
			interesting = true
		}
	}
	return interesting
}

func (c *Cov) Score() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	psl := len(c.ps)
	csl := len(c.cs)
	return int(math.Log(float64(psl))) + csl*10
}
