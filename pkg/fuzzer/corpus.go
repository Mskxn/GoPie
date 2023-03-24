package fuzzer

import (
	"fmt"
	"math/rand"
	"sync"
	"toolkit/pkg"
)

const allowDup = false
const BANMAX = 100

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
	gm     map[string]*Chain
	tm     map[string]*pkg.Pair
	preban map[string]uint64
	ban    map[string]struct{}
	allow  map[string]struct{}
	hash   sync.Map
	gmu    sync.RWMutex
	tmu    sync.RWMutex
	bmu    sync.RWMutex
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
		GlobalCorpus.ban = make(map[string]struct{})
		GlobalCorpus.preban = make(map[string]uint64)
		GlobalCorpus.allow = make(map[string]struct{})
	})
}

func NewCorpus() *Corpus {
	corpus := &Corpus{}
	corpus.gm = make(map[string]*Chain)
	corpus.tm = make(map[string]*pkg.Pair)
	corpus.ban = make(map[string]struct{})
	corpus.preban = make(map[string]uint64)
	corpus.allow = make(map[string]struct{})
	return corpus
}

func (cp *Corpus) Get() *Chain {
	// TODO
	cp.gmu.RLock()
	cp.tmu.RLock()
	defer cp.tmu.RUnlock()
	defer cp.gmu.RUnlock()
	for _, v := range cp.gm {
		for _, vv := range cp.tm {
			c := NewChain(v, vv)
			if _, ok := cp.hash.LoadOrStore(c.ToString(), struct{}{}); !ok {
				return c
			}
			if allowDup {
				return c
			}
		}
	}
	return NewChain(nil, nil)
}

func (cp *Corpus) GGet() *Chain {
	// TODO
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	for _, v := range cp.gm {
		if _, ok := cp.hash.LoadOrStore(v.ToString(), struct{}{}); !ok {
			return v
		}
		if allowDup {
			return v
		}
	}
	return NewChain(nil, nil)
}

func (cp *Corpus) Ban(ps [][]uint64) {
	cp.bmu.Lock()
	defer cp.bmu.Unlock()
	for _, p := range ps {
		s := fmt.Sprintf("{%v, %v}", p[0], p[1])
		if _, ok := cp.ban[s]; ok {
			continue
		}
		if _, ok := cp.allow[s]; ok {
			continue
		}
		if _, ok := cp.preban[s]; !ok {
			cp.preban[s] = 0
		}
		cp.preban[s] += 1
		if cp.preban[s] >= BANMAX {
			cp.ban[s] = struct{}{}
			delete(cp.preban, s)
		}
	}
}

func (cp *Corpus) Allow(ps [][]uint64) {
	cp.bmu.Lock()
	defer cp.bmu.Unlock()
	for _, p := range ps {
		s := fmt.Sprintf("{%v, %v}", p[0], p[1])
		cp.allow[s] = struct{}{}
	}
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

func (cp *Corpus) GSUpdate(chains []*Chain) bool {
	var ok bool
	for _, chain := range chains {
		t := cp.GUpdate(chain)
		if t {
			ok = true
		}
	}
	return ok
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

func (cp *Corpus) GUpdateSeed(seeds []*pkg.Pair) {
	chs := make([]*Chain, 0)
	for _, seed := range seeds {
		chs = append(chs, &Chain{
			item: []*pkg.Pair{seed, seed},
		})
	}
	cp.GSUpdate(chs)
}

func GetGlobalCorpus() *Corpus {
	return &GlobalCorpus
}

func (cp *Corpus) Size() int {
	cp.gmu.RLock()
	defer cp.gmu.RUnlock()
	return len(cp.gm)
}
