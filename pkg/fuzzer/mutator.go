package fuzzer

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
}

func (m *Mutator) mutatet(e *Entry) []*Entry {
	// TODO
}

// filter the output of mutator by rules
func (m *Mutator) filter(chain *Chain) bool {
	// TODO
}
