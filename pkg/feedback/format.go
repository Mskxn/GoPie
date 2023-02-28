package feedback

import (
	"strconv"
	"strings"
	"toolkit/pkg/fuzzer"
)

/*
[FBSDK] makechan: chan=0xc00006a090; elemsize=8; dataqsiz=5
[FBSDK] Chansend: chan=0xc00006a090; elemsize=8; dataqsiz=5; qcount=1
[FB] chan: 0xc00006a090; id :841813590017;
[FBSDK] chanclose: chan=0xc00006a090
[FB] chan: 0xc00006a090; id :841813590018;
[FB] mutex: 0xc000089f3c; id :841813590019;
[FB] mutex: 0xc000089f3c; id :841813590020;
*/
func implParse(s string) (bool, []uint64) {
	res := make([]uint64, 0)
	ss := strings.Split(s, ";")
	for _, v := range ss {
		if len(v) == 0 || v == "\n" || v == "\t" || v == " " {
			continue
		}
		if !strings.Contains(v, "=") {
			return false, nil
		}
		value := strings.Split(v, "=")[1]
		var base int
		if len(value) >= 3 && value[0:2] == "0x" {
			base = 16
			value = value[2 : len(value)-1]
		} else {
			base = 10
		}
		if value[len(value)-1] == ';' {
			value = value[0 : len(value)-1]
		}
		i, err := strconv.ParseUint(value, base, 64)
		if err != nil {
			return false, nil
		}
		res = append(res, i)
	}
	return true, res
}

func parseLine(s string) (bool, uint64, OpAndStatus) {
	if !strings.HasPrefix(s, "[FB") {
		return false, 0, OpAndStatus{}
	}
	ss := strings.Split(s, ":")
	if len(ss) != 2 {
		return false, 0, OpAndStatus{}
	}
	head := ss[0]
	ok, values := implParse(ss[1])
	if !ok {
		return false, 0, OpAndStatus{}
	}
	var typ uint64
	switch head {
	case "[FBSDK] Chansend", "[FBSDK] chanrecv":
		if head == "[FBSDK] Chansend" {
			typ = Chansend
		} else {
			typ = Chanrecv
		}
		return true, values[0], OpAndStatus{
			opid:   0,
			oid:    values[0],
			status: ObjectStatus{isCh: true, dataqsize: int(values[2]), qcount: int(values[3])},
			gid:    values[4],
			typ:    typ,
		}
	case "[FBSDK] makechan":
		return true, values[0], OpAndStatus{
			opid:   0,
			oid:    values[0],
			status: ObjectStatus{isCh: true, dataqsize: int(values[2]), qcount: 0},
			typ:    Chanmake,
		}
	case "[FBSDK] chanclose":
		return true, values[0], OpAndStatus{
			opid:   0,
			oid:    values[0],
			status: ObjectStatus{isCh: true, closed: true},
			gid:    values[1],
			typ:    Chanclose,
		}
	case "[FB] chan":
		return true, values[0], OpAndStatus{
			opid:   values[1],
			oid:    values[0],
			status: ObjectStatus{isCh: true},
		}
	case "[FB] mutex":
		var islocked bool
		if values[2] == 1 {
			islocked = true
			typ = Lock
		} else {
			typ = Unlock
		}
		return true, values[0], OpAndStatus{
			opid:   values[1],
			oid:    values[0],
			status: ObjectStatus{isCh: false, locked: islocked},
			gid:    values[3],
			typ:    typ,
		}
	default:
		return false, 0, OpAndStatus{}
	}
}

func ParseLog(s string) (map[uint64][]OpAndStatus, []OpAndStatus) {
	lines := strings.Split(s, "\n")
	fblines := make([]string, 0)
	// global orders, which used to generate coverage
	orders := make([]OpAndStatus, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "[FB") {
			fblines = append(fblines, line)
		}
	}

	m := make(map[uint64][]OpAndStatus, 0)

	add := func(oid uint64, info OpAndStatus) {
		if _, ok := m[oid]; !ok {
			m[oid] = make([]OpAndStatus, 0)
		}
		// keep the order of logs
		m[oid] = append(m[oid], info)
		orders = append(orders, info)
	}

	// filter, FBSDK->FB for chan and [FB] mutex for mutex
	for i, line := range lines {
		if strings.HasPrefix(line, "[FB] mutex") {
			ok, oid, info := parseLine(line)
			if !ok {
				continue
			}
			add(oid, info)
		} else {
			if strings.HasPrefix(line, "[FBSDK]") {
				if i+1 < len(lines)-1 && strings.HasPrefix(lines[i+1], "[FB]") {
					ok, oid, info := parseLine(line)
					ok2, oid2, info2 := parseLine(lines[i+1])
					if !ok || !ok2 || (oid != oid2 && !info.status.isCh) {
						continue
					}
					info.opid = info2.opid
					add(oid, info)
				}
			}
		}
	}
	return m, orders
}

func Log2Cov(ops []OpAndStatus) Cov {
	cov := NewCov()
	l := len(ops)
	for i := 0; i < l-1; i++ {
		j := i + 1
		cov.UpdateP("", *fuzzer.NewPair(ops[i].opid, ops[j].opid))
	}
	for _, op := range ops {
		status := op.status.IsCritical()
		if status != 0 {
			cov.UpdateC(OpID(op.opid), ToStatus(status))
		}
	}
	return cov
}

type Fragments struct {
	m map[uint64]uint64
}

func (f *Fragments) Size() int {
	return len(f.m)
}

func (f *Fragments) Root(x uint64) uint64 {
	if x == 0 {
		return 0
	}
	for {
		if v, ok := f.m[x]; !ok {
			return 0
		} else {
			if v != x {
				x = v
			}
		}
	}
	return x
}

func (f *Fragments) IsSame(x uint64, y uint64) bool {
	return f.Root(x) != 0 && f.Root(x) == f.Root(y)
}

func (f *Fragments) Exist(x uint64) bool {
	_, ok := f.m[x]
	return ok
}

func (f *Fragments) Uint(x uint64, y uint64) uint64 {
	if !f.Exist(x) {
		f.Add(x)
	}
	if !f.Exist(y) {
		f.Add(y)
	}
	r1 := f.Root(x)
	r2 := f.Root(y)
	f.m[r2] = r1
	return r1
}

func (f *Fragments) Add(x uint64) {
	f.m[x] = x
}

func NewFragments(m map[uint64][]OpAndStatus) Fragments {
	// TODO
	return Fragments{}
}
