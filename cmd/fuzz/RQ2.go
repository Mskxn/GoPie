package main

import (
	"fmt"
	"time"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func RQ2(bin string, fn string) {
	sumCh := make(chan string, 1000000)
	bugCh := make(chan string, 1000000)
	nolimit := make(chan struct{}, 0)
	close(nolimit)
	dowork := func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
		m := fuzzer.Monitor{}
		m.Start(cfg, v, nolimit)
	}

	oneCase := func(id string, useanalysis, usefeedback, useguide, usemutate bool) []*fuzzer.Visitor {
		newBugset := bug.NewBugSet()
		sharedCov := feedback.NewCov()
		sharedCorpus := fuzzer.NewCorpus()
		sharedScore := int32(10)
		logCh := make(chan string, 10000)
		v := &fuzzer.Visitor{
			V_cov:    sharedCov,
			V_corpus: sharedCorpus,
			V_score:  &sharedScore,
		}
		newCfg := fuzzer.NewConfig(bin, fn, logCh, newBugset, "normal")
		newCfg.UseAnalysis = useanalysis
		newCfg.UseFeedBack = usefeedback
		newCfg.UseMutate = usemutate
		newCfg.MaxQuit = newCfg.MaxExecution
		newCfg.MaxWorker = 24

		go dowork(v, newCfg)
		go func() {
			for {
				log := <-logCh
				bugCh <- fmt.Sprintf("%v [%v] %v", time.Now(), id, log)
			}
		}()
		ticket := time.NewTicker(10 * time.Second)
		for {
			<-ticket.C
			o1sum, o2sum, statesum := sharedCov.Size()
			coveredsched := sharedCorpus.SchedCnt()
			totalrun := sharedCorpus.FetchCnt()
			// _ := atomic.LoadInt32(&sharedScore)
			sumCh <- fmt.Sprintf("[%s]\t%v\t%v\t%v\t%v\t%v\n", id, o1sum, o2sum, coveredsched, statesum, totalrun)
		}
	}

	go oneCase("FULL", true, true, true, true)
	// go oneCase("-An", false, true, true, true, true)
	go oneCase("-FB", true, false, true, true)
	// go oneCase("-SC", true, true, false, true, true)
	go oneCase("-GU", true, true, false, true)
	go oneCase("-MU", true, true, true, false)

	for {
		select {
		case v := <-sumCh:
			fmt.Printf(v)
		case v := <-bugCh:
			fmt.Printf(v)
		}
	}
}
