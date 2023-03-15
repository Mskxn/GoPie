package fuzzer

import (
	"log"
	"sync"
	"sync/atomic"
	"toolkit/pkg/feedback"
	"toolkit/pkg/seed"
)

const (
	maxWorker   = 5
	initTurnCnt = 10
	singleCrash = true
	maxFuzzCnt  = 1000
)

var runtimes int32

type Monitor struct {
}

type RunContext struct {
	In  Input
	Out Output
}

func (m *Monitor) Start(bin string, fn string) {
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
			atomic.AddInt32(&runtimes, 1)
			ch <- RunContext{In: in, Out: o}
			select {
			case <-cancel:
				return
			default:
			}
		}
	}

	wg.Add(maxWorker)
	for i := 0; i < maxWorker; i++ {
		go dowork()
	}

	for {
		if runtimes > maxFuzzCnt {
			close(cancel)
			wg.Wait()
			break
		}
		ctx := <-ch
		var inputc string
		if ctx.In.c != nil {
			inputc = ctx.In.c.ToString()
		} else {
			inputc = "empty chain"
		}
		log.Printf("[WORKER] recv work : %s", inputc)
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			log.Printf("[WORKER] find bug : \n %s", ctx.Out.O)
			if singleCrash {
				close(cancel)
				wg.Wait()
				break
			}
		}
		op_st, orders := feedback.ParseLog(ctx.Out.Trace)
		cov := feedback.Log2Cov(orders)
		if atomic.LoadInt32(&runtimes) < initTurnCnt {
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
