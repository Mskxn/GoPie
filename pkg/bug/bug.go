package bug

import (
	"strings"
	"sync"
	"sync/atomic"
)

func TopF(s string) []string {
	ss := strings.Split(s, "\n")
	topfunc := make([]string, 0)
	for _, line := range ss {
		if strings.Contains(line, "on top of the stack:") {
			idx := strings.LastIndex(line, " on top of the stack:")
			if idx != -1 {
				idx2 := strings.Index(line, "with ")
				if idx2 != -1 {
					f := line[idx2+5 : idx]
					if strings.Contains(s, "github.com") {
						topfunc = append(topfunc, f)
					}
				}
			}
		}
	}
	return topfunc
}

type BugSet struct {
	m   sync.Map
	cnt uint32
}

func NewBugSet() *BugSet {
	return &BugSet{}
}

func (bs *BugSet) Add(fs []string, fn string) bool {
	for _, f := range fs {
		_, exist := bs.m.LoadOrStore(fn+f, struct{}{})
		if exist {
			return true
		}
	}
	atomic.AddUint32(&bs.cnt, 1)
	return false
}

func (bs *BugSet) Size() uint32 {
	return atomic.LoadUint32(&bs.cnt)
}
