package main

import (
	"fmt"
	"time"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func RQ2(bin string) {
	sumCh := make(chan string, 10000)
	nolimit := make(chan struct{}, 0)
	close(nolimit)
	dowork := func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
		m := fuzzer.Monitor{}
		m.Start(cfg, v, nolimit)
	}

	tests := ListTests(bin)

	oneCase := func(id string, useanalysis, usefeedback, usesched, usestate, usemutate bool) []*fuzzer.Visitor {
		newBugset := bug.NewBugSet()
		sharedCov := feedback.NewCov()
		sharedCorpus := fuzzer.NewCorpus()
		logCh := make(chan string, 10000)
		for _, test := range tests {
			v := &fuzzer.Visitor{
				V_cov:    sharedCov,
				V_corpus: sharedCorpus,
			}

			newCfg := fuzzer.NewConfig(bin, test, logCh, newBugset, "normal")
			newCfg.UseAnalysis = useanalysis
			newCfg.UseStates = usestate
			newCfg.UseCoveredSched = usesched
			newCfg.UseFeedBack = usefeedback
			newCfg.UseStates = usemutate

			go dowork(v, newCfg)
			go func() {
				for {
					<-logCh
				}
			}()
		}
		ticket := time.NewTicker(10 * time.Second)
		for {
			<-ticket.C
			pairsum, statesum := sharedCov.Size()
			coveredsched := sharedCorpus.SchedCnt()
			totalrun := sharedCorpus.FetchCnt()
			sumCh <- fmt.Sprintf("[%s]\t%v\t%v\t%v\t%v\n", id, pairsum, statesum, coveredsched, totalrun)
		}
	}

	go oneCase("FULL", true, true, true, true, true)
	go oneCase("-An", false, true, true, true, true)
	go oneCase("-FB", true, false, true, true, true)
	go oneCase("-SC", true, true, false, true, true)
	go oneCase("-ST", true, true, true, false, true)
	go oneCase("-MU", true, true, true, true, false)

	for {
		select {
		case v := <-sumCh:
			fmt.Printf(v)
		}
	}
}
