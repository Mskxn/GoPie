package fuzzer

import "math"

type Stmt string

type Pair struct {
	prev Stmt
	next Stmt
}

type Status struct {
	ops []Stmt
	// channel full, channel empty, channel closed ...
	s uint64
}

type Cov struct {
	ps map[uint64]Pair
	cs map[uint64]Status
}

func NewGlobalCov() {
	cov := Cov{}
	cov.ps = make(map[uint64]Pair)
	cov.cs = make(map[uint64]Status)
}

func (c *Cov) Update(cc *Cov) bool {
	var interesting bool = false
	for k, v := range cc.ps {
		if _, ok := c.ps[k]; !ok {
			c.ps[k] = v
			interesting = true
		}
	}
	for k, v := range cc.cs {
		if _, ok := c.cs[k]; !ok {
			c.cs[k] = v
			interesting = true
		}
	}
	return interesting
}

func (c *Cov) Score() int {
	psl := len(c.ps)
	csl := len(c.cs)
	return int(math.Log(float64(psl))) + csl*10
}
