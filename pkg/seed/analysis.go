package seed

import (
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func SODRAnalysis(m map[uint64][]feedback.OpAndStatus) []fuzzer.Pair {
	//Same object operated in different routines
	res := make([]fuzzer.Pair, 0)
	for _, l := range m {
		size := len(l)
		for i := 0; i < size; i++ {
			for j := i + 1; j < size; j++ {
				if l[i].Gid != l[j].Gid {
					// Do not need to do a reverse, mutator can do it.
					res = append(res, fuzzer.Pair{
						Prev: fuzzer.Entry{
							Opid: l[i].Opid,
						},
						Next: fuzzer.Entry{
							Opid: l[j].Opid,
						},
					})
				}
			}
		}
	}
	return res
}

func SRDOAnalysis(m map[uint64][]feedback.OpAndStatus) []fuzzer.Pair {
	//Same routine operate different objects
	visit := make(map[string]fuzzer.Pair, 0)

	// switch to routine view
	m2 := make(map[uint64][]feedback.OpAndStatus)
	for _, l := range m {
		size := len(l)
		for i := 0; i < size; i++ {
			gid := l[i].Gid
			if _, ok := m2[gid]; !ok {
				m2[gid] = make([]feedback.OpAndStatus, 0)
			}
			m2[gid] = append(m2[gid], l[i])
		}
	}

	for _, l := range m2 {
		size := len(l)
		for i := 0; i < size; i++ {
			for j := i + 1; j < size; j++ {
				if l[i].Oid != l[j].Oid || l[i].Typ != l[j].Typ {
					// Find different objects operated in same routine, which we easily consider
					// other operations on these objects are possibly related
					oid1 := l[i].Oid
					oid2 := l[j].Oid
					if l1, ok := m[oid1]; ok {
						if l2, ok2 := m[oid2]; ok2 {
							for i, op1 := range l1 {
								for j, op2 := range l2 {
									if op1.Gid != op2.Gid {
										p := fuzzer.Pair{
											Prev: fuzzer.Entry{
												Opid: l[i].Opid,
											},
											Next: fuzzer.Entry{
												Opid: l[j].Opid,
											},
										}
										visit[p.ToString()] = p
									}
								}
							}
						}
					}
				}
			}
		}
	}

	res := make([]fuzzer.Pair, 0)
	for _, v := range visit {
		res = append(res, v)
	}
	return res
}