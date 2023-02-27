package fuzzer

import "math/rand"

type Entry struct {
	prev Stmt
	next Stmt
}

type Chain struct {
	item []*Entry
}

func NewChain(g *Chain, t *Entry) *Chain {
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

func (c *Chain) T() *Entry {
	if c.Len() == 0 {
		return nil
	}
	return c.item[len(c.item)-1]
}

func (c *Chain) pop() *Entry {
	if c.Len() == 0 {
		return nil
	}
	last := c.item[len(c.item)-1]
	c.item = c.item[0 : len(c.item)-1]
	return last
}

func (c *Chain) add(e *Entry) {
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
	chains []Chain
}
