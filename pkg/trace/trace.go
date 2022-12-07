package trace

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type routineInfo struct {
	id     int64
	start  time.Time
	end    time.Time
	pos    string
	finish bool
}

type AllInfos struct {
	m sync.Map
}

var allInfos AllInfos

func (info *AllInfos) add(pos string) int64 {
	id := rand.Int63()
	info.m.Store(id, &routineInfo{id, time.Now(), time.Now(), pos, false})
	return id
}

func (info *AllInfos) del(id int64) {
	if v, ok := info.m.Load(id); ok {
		if vv, ok := v.(*routineInfo); ok {
			vv.finish = true
			vv.end = time.Now()
		}
	}
}

func GoStart(pos string) int64 {
	return allInfos.add(pos)
}

func GoEnd(id int64) {
	allInfos.del(id)
}

func Check(t *testing.T) {
	hangs := make([]string, 0)
	allInfos.m.Range(func(key, value any) bool {
		if v, ok := value.(*routineInfo); ok && v.finish {
			s := fmt.Sprintf("[LEAK] Create at : %v", v.pos)
			hangs = append(hangs, s)
		}
		return true
	})

	for _, s := range hangs {
		t.Log(s)
	}
	if len(hangs) != 0 {
		t.Fatal()
	}
	allInfos.m = sync.Map{}
}
