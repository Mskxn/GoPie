package fuzzer

import (
	"math/rand"
	"sync"
	"toolkit/pkg"
)

type Chain struct {
	item []*pkg.Pair
}

func (c *Chain) Copy() *Chain {
	res := Chain{item: []*pkg.Pair{}}
	if c == nil {
		return &res
	}
	for _, p := range c.item {
		res.add(p)
	}
	return &res
}

func (c *Chain) ToString() string {
	res := ""
	if c == nil {
		return res
	}
	if c.item == nil {
		return res
	}
	for _, e := range c.item {
		if e == nil {
			continue
		}
		res += e.ToString()
	}
	return res
}

func NewChain(g *Chain, t *pkg.Pair) *Chain {
	if g == nil && t == nil {
		return nil
	}
	nc := g.Copy()
	nc.add(t)
	return nc
}

func (c *Chain) Len() int {
	if c == nil {
		return 0
	}
	return len(c.item)
}

func (c *Chain) G() *Chain {
	if c.Len() <= 1 {
		return nil
	}
	return &Chain{c.item[0 : len(c.item)-1]}
}

func (c *Chain) T() *pkg.Pair {
	if c.Len() == 0 {
		return nil
	}
	return c.item[len(c.item)-1]
}

func (c *Chain) pop() *pkg.Pair {
	if c.Len() == 0 {
		return nil
	}
	last := c.item[len(c.item)-1]
	c.item = c.item[0 : len(c.item)-1]
	return last
}

func (c *Chain) add(e *pkg.Pair) {
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
	gm   map[string]*Chain
	tm   map[string]*pkg.Pair
	hash sync.Map
	gmu  sync.RWMutex
	tmu  sync.RWMutex
}

var once sync.Once
var GlobalCorpus Corpus

func init() {
	GlobalCorpus = Corpus{}
	GlobalCorpus.Init()
}

func (cp *Corpus) Init() {
	once.Do(func() {
		GlobalCorpus.gm = make(map[string]*Chain)
		GlobalCorpus.tm = make(map[string]*pkg.Pair)
	})
}

func (cp *Corpus) Get() *Chain {
	// TODO
	for _, v := range cp.gm {
		for _, vv := range cp.tm {
			c := NewChain(v, vv)
			if _, ok := cp.hash.LoadOrStore(c.ToString(), struct{}{}); !ok {
				return c
			}
		}
	}
	return NewChain(nil, nil)
}

func (cp *Corpus) GExist(chain *Chain) bool {
	if chain == nil {
		return false
	}
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	if _, ok := cp.gm[chain.ToString()]; ok {
		return true
	}
	return false
}

func (cp *Corpus) TExist(e *pkg.Pair) bool {
	if e == nil {
		return false
	}
	cp.tmu.RLock()
	defer cp.tmu.RUnlock()
	if _, ok := cp.tm[e.ToString()]; ok {
		return true
	}
	return false
}

func (cp *Corpus) GUpdate(chain *Chain) bool {
	if cp.GExist(chain) {
		return false
	}
	cp.gmu.Lock()
	defer cp.gmu.Unlock()
	cp.gm[chain.ToString()] = chain
	return true
}

func (cp *Corpus) TUpdate(e *pkg.Pair) bool {
	if cp.TExist(e) {
		return false
	}
	cp.tmu.Lock()
	defer cp.tmu.Unlock()
	cp.tm[e.ToString()] = e
	return true
}

func (cp *Corpus) Update(chs []*Chain) bool {
	ok := false
	for _, c := range chs {
		if cp.GUpdate(c.G()) {
			ok = true
		}
		if cp.TUpdate(c.T()) {
			ok = true
		}
	}
	return ok
}

func (cp *Corpus) UpdateSeed(seeds []*pkg.Pair) {
	chs := make([]*Chain, 0)
	for _, seed := range seeds {
		chs = append(chs, &Chain{
			item: []*pkg.Pair{seed, seed},
		})
	}
	cp.Update(chs)
}

func GetGlobalCorpus() *Corpus {
	return &GlobalCorpus
}
