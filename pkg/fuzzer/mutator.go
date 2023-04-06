package fuzzer

import (
	"math/rand"
	"toolkit/pkg"
	"toolkit/pkg/feedback"
)

const (
	BOUND       = 5
	MUTATEBOUND = 1024
)

type Mutator struct {
	Cov *feedback.Cov
}

func (m *Mutator) mutate(chain *Chain, energy int) ([]*Chain, map[uint64]map[uint64]struct{}) {
	// TODO : energy
	gs := m.mutateg(chain, energy)
	ht := make(map[uint64]map[uint64]struct{})
	for _, g := range gs {
		m.mutatet(g, ht)
	}
	return gs, ht
}

func (m *Mutator) mutateg(chain *Chain, energy int) []*Chain {
	// TODO
	if chain == nil {
		return []*Chain{}
	}
	if chain.Len() == 0 {
		return []*Chain{}
	}
	res := make([]*Chain, 0)
	set := make(map[string]*Chain, 0)
	set[chain.ToString()] = chain
	if chain.Len() == 1 {
		nc := &Chain{[]*pkg.Pair{&pkg.Pair{chain.T().Next, chain.T().Prev}}}
		set[nc.ToString()] = nc
	}
	for (rand.Int() % 150) < energy {
		for _, chain := range set {
			tset := make(map[string]*Chain, 0)
			// reduce the length
			if chain.Len() >= 2 {
				nc := chain.Copy()
				nc.pop()
				tset[nc.ToString()] = nc
				nc2 := chain.Copy()
				nc2.item = nc2.item[1:len(nc2.item)]
				tset[nc2.ToString()] = nc2
			}

			// increase the length
			if chain.Len() <= BOUND {
				if rand.Int()%2 == 1 {
					lastopid := chain.T().Next.Opid
					nexts := m.Cov.Next(lastopid)
					for _, next := range nexts {
						for _, next2 := range m.Cov.Next(next) {
							if lastopid != next2 {
								nc := chain.Copy()
								nc.add(pkg.NewPair(lastopid, next2))
								tset[nc.ToString()] = nc
							}
						}
					}
				} else {
					nc := chain.Copy()
					nc.add(GetGlobalCorpus().GetC())
					tset[nc.ToString()] = nc
				}
			}
			for k, v := range tset {
				set[k] = v
			}
		}
		if len(set) > MUTATEBOUND {
			break
		}
	}

	// merge two chain
	// TODO
	for _, v := range set {
		if m.filter(v) {
			res = append(res, v)
		}
	}

	return res
}

// mutatet find out possible attack pairs for each pair in the EC
func (m *Mutator) mutatet(c *Chain, ht map[uint64]map[uint64]struct{}) {
	for _, p := range c.item {
		nop := p.Next
		if _, ok := ht[nop.Opid]; !ok {
			ht[nop.Opid] = make(map[uint64]struct{}, 0)
		}

		// hang attack
		ht[nop.Opid][nop.Opid] = struct{}{}

		st, ok := m.Cov.GetStatus(feedback.OpID(nop.Opid))
		if !ok {
			return
		}
		stn := st.ToU64()

		typefilter := func(stbit, tbit uint64) {
			if stn&stbit != 0 {
				next, ok := m.Cov.NextTyp(nop.Opid, tbit, nil)
				if ok {
					ht[nop.Opid][next] = struct{}{}
				}
			}
		}

		// find status by typ
		statusfilter := func(stbit, sbit uint64) {
			if stn&stbit != 0 {
				next, ok2 := m.Cov.NextStatus(nop.Opid, stbit, nil)
				if ok2 {
					ht[nop.Opid][next] = struct{}{}
				}
			}
		}
		rule := feedback.RuleMap
		for st2, op := range rule {
			statusfilter(st2, op)
			typefilter(st2, op)
		}
	}
}

// filter the output of mutator by rules
func (m *Mutator) filter(chain *Chain) bool {
	for i := 0; i < chain.Len()-1; i++ {
		p1 := chain.item[i]
		p2 := chain.item[i+1]
		if p1.Next.Opid != p2.Prev.Opid {
			return false
		}
	}
	return true
}
