package fuzzer

import (
	"toolkit/pkg"
	"toolkit/pkg/feedback"
)

const (
	BOUND = 5
)

type Mutator struct {
}

func (m *Mutator) mutate(chain Chain) []*Chain {
	res := make([]*Chain, 0)
	gs := m.mutateg(chain.G())
	ts := m.mutatet(chain.T())
	for _, g := range gs {
		for _, t := range ts {
			nc := NewChain(g, t)
			ok := m.filter(nc)
			if ok {
				res = append(res, nc)
			}
		}
	}
	return res
}

func (m *Mutator) mutateg(chain *Chain) []*Chain {
	// TODO
	if chain == nil {
		return []*Chain{}
	}
	// reduce the length
	res := make([]*Chain, 0)
	if chain.Len() >= 1 {
		nc := chain.Copy()
		nc.pop()
		res = append(res, nc)
	}

	// increase the length
	if chain.Len() < BOUND {
		lastopid := chain.T().Next.Opid
		nexts := feedback.GlobalCov.Next(lastopid)
		for _, next := range nexts {
			nc := chain.Copy()
			nc.add(pkg.NewPair(lastopid, next))
			res = append(res, nc)
		}
	}

	// merge two chain
	// TODO

	return res
}

func (m *Mutator) mutatet(e *pkg.Pair) []*pkg.Pair {
	// TODO
	res := make([]*pkg.Pair, 0)
	st, ok := feedback.GlobalCov.GetStatus(feedback.OpID(e.Prev.Opid))
	if !ok {
		return res
	}

	stn := st.ToU64()

	// reverse
	// res = append(res, &Pair{e.Next, e.Prev})

	/*
		TODO : shit code,too many duplicate search works, but don't want to fix now
	*/

	// find type by status
	typefilter := func(stbit, tbit uint64) {
		if stn&stbit != 0 {
			next, ok := feedback.GlobalCov.NextTyp(e.Prev.Opid, tbit, nil)
			if ok {
				res = append(res, &pkg.Pair{Prev: e.Prev, Next: pkg.Entry{Opid: next}})
			}
		}
	}

	// find status by typ
	statusfilter := func(stbit, sbit uint64) {
		if stn&stbit != 0 {
			next, ok := feedback.GlobalCov.NextStatus(e.Next.Opid, stbit, nil)
			if ok {
				res = append(res, &pkg.Pair{Prev: pkg.Entry{Opid: next}, Next: e.Next})
			}
		}
	}

	/*
		typefilter(feedback.ChanFull, feedback.Chansend)
		typefilter(feedback.ChanEmpty, feedback.Chanrecv)
		typefilter(feedback.ChanClosed, feedback.Chansend|feedback.Chanclose)
		typefilter(feedback.MutexLocked, feedback.Lock)
		typefilter(feedback.MutexUnlocked, feedback.Unlock)

		statusfilter(feedback.ChanFull, feedback.Chansend)
		statusfilter(feedback.ChanEmpty, feedback.Chanrecv)
		statusfilter(feedback.ChanClosed, feedback.Chansend|feedback.Chanclose)
		statusfilter(feedback.MutexLocked, feedback.Lock)
		statusfilter(feedback.MutexUnlocked, feedback.Unlock)
	*/

	rule := feedback.RuleMap
	for st, op := range rule {
		statusfilter(st, op)
		typefilter(st, op)
	}

	return res
}

// filter the output of mutator by rules
func (m *Mutator) filter(chain *Chain) bool {
	// TODO
	if chain.Len() <= 1 {
		return true
	}
	g := chain.G()
	t := chain.T()

	// get g status
	status := feedback.Status{}
	for _, e := range g.item {
		id := e.Prev.Opid
		s, ok := feedback.GlobalCov.GetStatus(feedback.OpID(id))
		if ok {
			status = feedback.Merge(status, s)
		}
	}

	// check t
	stn := status.ToU64()
	if typ, ok := feedback.GlobalCov.GetTyp(feedback.OpID(t.Next.Opid)); ok {
		for s, op := range feedback.RuleMap {
			if s&stn != 0 {
				if typ&op != 0 {
					return true
				}
			}
		}
	}
	return false
}
