package feedback

import (
	"strconv"
	"strings"
)

/*
[FBSDK] makechan: chan=0xc00006a090; elemsize=8; dataqsiz=5
[FBSDK] chansend: chan=0xc00006a090; elemsize=8; dataqsiz=5; qcount=1
[FB] chan: 0xc00006a090; id :841813590017;
[FBSDK] chanclose: chan=0xc00006a090
[FB] chan: 0xc00006a090; id :841813590018;
[FB] mutex: 0xc000089f3c; id :841813590019;
[FB] mutex: 0xc000089f3c; id :841813590020;
*/

type Status struct {
	isCh bool
	// for channel
	dataqsize int
	qcount    int
	closed    bool
	// for mutex
	locked bool
}

type OpAndStatus struct {
	opid   uint64
	status Status
}

func implParse(s string) (bool, []uint64) {
	res := make([]uint64, 0)
	ss := strings.Split(s, ";")
	for _, v := range ss {
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

	switch head {
	case "[FBSDK] chansend", "[FBSDK] chanrecv":
		return true, values[0], OpAndStatus{
			opid:   0,
			status: Status{isCh: true, dataqsize: int(values[2]), qcount: int(values[3])},
		}
	case "[FBSDK] makechan":
		return true, values[0], OpAndStatus{
			opid:   0,
			status: Status{isCh: true, dataqsize: int(values[2]), qcount: 0},
		}
	case "[FBSDK] chanclose":
		return true, values[0], OpAndStatus{
			opid:   0,
			status: Status{isCh: true, closed: true},
		}
	case "[FB] chan":
		return true, values[0], OpAndStatus{
			opid:   values[1],
			status: Status{isCh: true},
		}
	case "[FB] mutex":
		var islocked bool
		if values[2] == 1 {
			islocked = true
		}
		return true, values[0], OpAndStatus{
			opid:   values[1],
			status: Status{isCh: false, locked: islocked},
		}
	default:
		return false, 0, OpAndStatus{}
	}
}

func ParseLog(s string) map[uint64][]OpAndStatus {
	lines := strings.Split(s, "\n")
	fblines := make([]string, 0)
	for _, line := range lines {
		if strings.HasPrefix(line, "[FB") {
			fblines = append(fblines, line)
		}
	}

	m := make(map[uint64][]OpAndStatus, 0)

	for _, line := range fblines {
		ok, addr, info := parseLine(line)
		if !ok {
			continue
		}
		if _, ok := m[addr]; !ok {
			m[addr] = make([]OpAndStatus, 0)
		}
		m[addr] = append(m[addr], info)
	}
	return m
}
