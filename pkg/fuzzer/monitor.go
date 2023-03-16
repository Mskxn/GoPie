package fuzzer

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"toolkit/pkg/feedback"
	"toolkit/pkg/seed"
)

var (
	maxWorker   = 5
	initTurnCnt = 10
	maxFuzzCnt  = 1000
	singleCrash = true
	debug       = false
)

type Monitor struct {
	etimes int32
	max    int32
}

type RunContext struct {
	In  Input
	Out Output
}

func (m *Monitor) Start(bin string, fn string) (bool, []string) {
	if m.max == int32(0) {
		m.max = int32(maxFuzzCnt)
	}
	ch := make(chan RunContext, maxWorker)
	cancel := make(chan struct{})

	var wg sync.WaitGroup
	dowork := func() {
		defer wg.Done()
		for {
			c := GetGlobalCorpus().Get()
			e := Executor{}
			in := Input{
				c:    c,
				cmd:  bin,
				args: []string{"-test.v", "-test.run", fn},
			}
			o := e.Run(in)
			atomic.AddInt32(&m.etimes, 1)
			ch <- RunContext{In: in, Out: o}
			select {
			case <-cancel:
				break
			default:
			}
		}
	}

	for i := 0; i < maxWorker; i++ {
		go dowork()
	}

	for {
		if m.etimes > m.max {
			close(cancel)
			return false, []string{}
		}
		ctx := <-ch
		var inputc string
		if ctx.In.c != nil {
			inputc = ctx.In.c.ToString()
		} else {
			inputc = "empty chain"
		}
		if debug {
			log.Printf("[WORKER] recv work : %s", inputc)
		}
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			detail := []string{inputc, strconv.FormatInt(int64(atomic.LoadInt32(&m.etimes)), 10), ctx.Out.O}
			if debug {
				log.Println(inputc, detail[1])
			}
			if singleCrash {
				close(cancel)
				return true, detail
			}
		}
		op_st, orders := feedback.ParseLog(ctx.Out.Trace)
		cov := feedback.Log2Cov(orders)
		if atomic.LoadInt32(&m.etimes) < int32(initTurnCnt) {
			seeds := seed.SRDOAnalysis(op_st)
			seeds = append(seeds, seed.SODRAnalysis(op_st)...)
			GetGlobalCorpus().UpdateSeed(seeds)
		}
		ok := feedback.GetGlobalCov().Merge(&cov)
		if ok && ctx.In.c != nil {
			m := Mutator{}
			ncs := m.mutate(*ctx.In.c)
			GetGlobalCorpus().Update(ncs)
		}
	}
}
