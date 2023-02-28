package fuzzer

import (
	"fmt"
	"math/rand"
	"sync"
)

type Entry struct {
	Opid uint64
}

type Pair struct {
	Prev Entry
	Next Entry
}

func (e *Pair) ToString() string {
	return fmt.Sprintf("(%v, %v)", e.Prev, e.Next)
}

func NewPair(op1, op2 uint64) *Pair {
	return &Pair{
		Prev: Entry{op1},
		Next: Entry{op2},
	}
}

type Chain struct {
	item []*Pair
}

func (c *Chain) Copy() *Chain {
	res := Chain{item: []*Pair{}}
	for _, p := range c.item {
		res.add(p)
	}
	return &res
}

func (c *Chain) ToString() string {
	res := ""
	for _, e := range c.item {
		res += e.ToString()
	}
	return res
}

func NewChain(g *Chain, t *Pair) *Chain {
	nc := *g
	nc.add(t)
	return &nc
}

func (c *Chain) Len() int {
	return len(c.item)
}

func (c *Chain) G() *Chain {
	if c.Len() <= 1 {
		return nil
	}
	return &Chain{c.item[0 : len(c.item)-1]}
}

func (c *Chain) T() *Pair {
	if c.Len() == 0 {
		return nil
	}
	return c.item[len(c.item)-1]
}

func (c *Chain) pop() *Pair {
	if c.Len() == 0 {
		return nil
	}
	last := c.item[len(c.item)-1]
	c.item = c.item[0 : len(c.item)-1]
	return last
}

func (c *Chain) add(e *Pair) {
	c.item = append(c.item, e)
}

func (c *Chain) merge(cc *Chain) {
	p1 := 0
	p2 := 0
	for {
		if p1 < c.Len() && p2 < cc.Len() {
			if rand.Int()%2 == 0 {
				p1 += 1
			} else {
				p2 += 1
			}
			if p1 < c.Len() && p2 < cc.Len() && rand.Int()%2 == 0 {
				temp := c.item[p1]
				c.item[p1] = c.item[p2]
				c.item[p2] = temp
			}
		}
	}
}

type Corpus struct {
	gm map[string]Chain
	tm map[string]Pair
	// store the orders of last or next op
	orders map[uint64]map[uint64]struct{}
	gmu    sync.RWMutex
	tmu    sync.RWMutex
	omu    sync.Mutex
}

var once sync.Once
var GlobalCorpus Corpus

func (cp *Corpus) Init() {
	once.Do(func() {
		GlobalCorpus.gm = make(map[string]Chain)
		GlobalCorpus.tm = make(map[string]Pair)
		GlobalCorpus.orders = make(map[uint64]map[uint64]struct{})
	})
}

func (cp *Corpus) GExist(chain Chain) bool {
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	if _, ok := cp.gm[chain.ToString()]; ok {
		return true
	}
	return false
}

func (cp *Corpus) TExist(e Pair) bool {
	cp.tmu.RLock()
	defer cp.tmu.RUnlock()
	if _, ok := cp.tm[e.ToString()]; ok {
		return true
	}
	return false
}

func (cp *Corpus) GUpdate(chain Chain) bool {
	if cp.GExist(chain) {
		return false
	}
	cp.gmu.Lock()
	defer cp.gmu.Unlock()
	cp.gm[chain.ToString()] = chain
	return true
}

func (cp *Corpus) TUpdate(e Pair) bool {
	if cp.TExist(e) {
		return false
	}
	cp.tmu.Lock()
	defer cp.tmu.Unlock()
	cp.tm[e.ToString()] = e
	return true
}

func (cp *Corpus) Next(opid uint64) []uint64 {
	cp.omu.Lock()
	defer cp.omu.Unlock()
	if v, ok := cp.orders[opid]; ok {
		res := make([]uint64, 0)
		for k, _ := range v {
			res = append(res, k)
		}
		return res
	}
	return nil
}

func (cp *Corpus) AddNext(opid uint64, next uint64) {
	cp.omu.Lock()
	defer cp.omu.Unlock()
	if _, ok := cp.orders[opid]; !ok {
		cp.orders[opid] = make(map[uint64]struct{})
	}
	cp.orders[opid][next] = struct{}{}
}
