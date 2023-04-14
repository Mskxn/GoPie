package feedback

import (
	"fmt"
	"math"
	"sync"
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
	F uint64
	E uint64
	C uint64
	L uint64
	U uint64
}

func (s *Status) ToU64() uint64 {
	var res = uint64(0)
	res |= s.F
	res |= s.E
	res |= s.C
	res |= s.L
	res |= s.U
	return res
}

func Merge(s Status, ss Status) Status {
	return ToStatus(s.ToU64() | ss.ToU64())
}

func ToStatus(s uint64) Status {
	res := Status{}
	res.F = s & ChanFull
	res.E = s & ChanEmpty
	res.C = s & ChanClosed
	res.L = s & MutexLocked
	res.U = s & MutexUnlocked
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
	// ps map[string]fuzzer.Pair
	cs map[OpID]Status
	// the ops next to each op
	orders map[uint64]map[uint64]struct{}
	rel    map[uint64]map[uint64]struct{}

	mu sync.Mutex

	// for orders
	omu sync.Mutex
	rmu sync.Mutex
}

var GlobalCov *Cov

func init() {
	SetGlobalCov()
}

func SetGlobalCov() {
	GlobalCov = NewCov()
}

func GetGlobalCov() *Cov {
	return GlobalCov
}

func NewCov() *Cov {
	cov := Cov{}
	// cov.ps = make(map[string]fuzzer.Pair)
	cov.cs = make(map[OpID]Status)
	cov.rel = make(map[uint64]map[uint64]struct{})
	cov.orders = make(map[uint64]map[uint64]struct{})
	return &cov
}

func (c *Cov) ToString() string {
	pairs := "Pairs: \n"
	cs := "Status: \n"
	score := "Score: "
	var cnt = 0

	/*
		for k, _ := range c.ps {
			pairs += fmt.Sprintf("[%v] %s\n", cnt, k)
			cnt += 1
		}
	*/
	for k, v := range c.orders {
		for kk, _ := range v {
			pairs += fmt.Sprintf("[%v] (%v, %v)\n", cnt, k, kk)
			cnt += 1
		}
	}

	cnt = 0
	for opid, c := range c.cs {
		cs += fmt.Sprintf("[%v] %v (%s)\n", cnt, opid, c.ToString())
		cnt += 1
	}
	score += fmt.Sprintf("%v\n", c.Score(true))
	return pairs + "\n" + cs + "\n" + score + "\n"
}

func (c *Cov) GetStatus(k OpID) (Status, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	s, ok := c.cs[k]
	return s, ok
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

	for k, v := range cc.cs {
		if c.UpdateC(k, v) {
			interesting = true
		}
	}

	if c.OMerge(cc) {
		interesting = true
	}
	return interesting
}

func (c *Cov) Score(usest bool) int {
	psl, csl := c.Size()
	var score int
	if usest {
		score = int(math.Log(float64(psl))) + csl*10
	} else {
		score = int(math.Log(float64(psl)))
	}
	return score
}

func (c *Cov) Size() (int, int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// psl := len(c.ps)

	psl := 2
	for _, v := range c.orders {
		psl += len(v)
	}

	csl := len(c.cs)
	return psl, csl
}

func (c *Cov) Next(opid uint64) []uint64 {
	c.omu.Lock()
	defer c.omu.Unlock()
	if v, ok := c.orders[opid]; ok {
		res := make([]uint64, 0)
		for k, _ := range v {
			res = append(res, k)
		}
		return res
	}
	return nil
}

func (c *Cov) NextR(opid uint64) []uint64 {
	c.omu.Lock()
	defer c.omu.Unlock()
	if v, ok := c.rel[opid]; ok {
		res := make([]uint64, 0)
		for k, _ := range v {
			res = append(res, k)
		}
		return res
	}
	return nil
}

func (c *Cov) UpdateR(covered [][]uint64) {
	c.rmu.Lock()
	defer c.rmu.Unlock()
	for _, v := range covered {
		if _, ok := c.rel[v[0]]; !ok {
			c.rel[v[0]] = make(map[uint64]struct{})
		}
		if _, ok := c.rel[v[1]]; !ok {
			c.rel[v[1]] = make(map[uint64]struct{})
		}
		c.rel[v[0]][v[1]] = struct{}{}
		c.rel[v[1]][v[0]] = struct{}{}
	}
}

func (c *Cov) UpdateO(opid uint64, next uint64) {
	c.omu.Lock()
	defer c.omu.Unlock()
	if _, ok := c.orders[opid]; !ok {
		c.orders[opid] = make(map[uint64]struct{})
	}
	c.orders[opid][next] = struct{}{}
}

func (c *Cov) OMerge(cc *Cov) bool {
	c.omu.Lock()
	defer c.omu.Unlock()
	var interesting bool
	for opid, nexts := range cc.orders {
		for nextid, _ := range nexts {
			if _, ok := c.orders[opid]; !ok {
				c.orders[opid] = make(map[uint64]struct{})
				interesting = true
			}
			if _, ok := c.orders[opid][nextid]; !ok {
				interesting = true
			}
			c.orders[opid][nextid] = struct{}{}
		}
	}
	return interesting
}
