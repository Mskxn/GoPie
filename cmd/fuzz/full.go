package main

import (
	"fmt"
	"runtime"
	"sync"
	"time"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func Full(path string, llevel string, feature string, maxworker int) {
	resCh := make(chan string, 100000)
	logCh := make(chan string, 100000)
	// control
	max := 48
	if maxworker != 0 {
		max = maxworker
	}
	limit := make(chan struct{}, max*2)
	for i := 0; i < max; i++ {
		limit <- struct{}{}
	}

	bin2tests := make(map[string][]string)
	bin2vst := make(map[string]*fuzzer.Visitor)
	bin2cnt := make(map[string]int)
	bugset := bug.NewBugSet()

	bins := cmd.ListFiles(path, func(s string) bool {
		return true
	})

	// bind tests and visitor to bins
	for _, bin := range bins {
		tests := ListTests(bin)
		bin2tests[bin] = tests
		bin2cnt[bin] = len(tests)

		cov := feedback.NewCov()
		corpus := fuzzer.NewCorpus()
		v := &fuzzer.Visitor{
			V_cov:    cov,
			V_corpus: corpus,
		}

		bin2vst[bin] = v
	}

	total := 0
	for bin, tests := range bin2tests {
		vst := bin2vst[bin]
		var wg sync.WaitGroup
		for _, test := range tests {
			cfg := fuzzer.DefaultConfig()
			// shared bugset
			cfg.BugSet = bugset
			cfg.Bin = bin
			cfg.Fn = test
			cfg.MaxWorker = 2
			cfg.TimeOut = 30
			cfg.RecoverTimeOut = 200
			cfg.LogCh = logCh
			cfg.MaxQuit = 64
			cfg.MaxExecution = 10000
			cfg.LogLevel = llevel
			if feature == "mu" {
				cfg.UseMutate = false
			}
			if feature == "fb" {
				cfg.UseFeedBack = false
			}
			wg.Add(1)
			go func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
				defer wg.Done()
				m := &fuzzer.Monitor{}
				ok, detail := m.Start(cfg, v, limit)
				var res string
				if ok {
					res = fmt.Sprintf("%s\tFAIL\t%s\n", cfg.Fn, detail[1])
				} else {
					res = fmt.Sprintf("%s\tPASS\n", cfg.Fn)
				}
				resCh <- res
			}(vst, cfg)
			total += 1
		}
		wg.Wait()
		delete(bin2vst, bin)
		runtime.GC()
	}

	defer fmt.Printf("%v [Fuzzer] Finish\n", time.Now().String())
	cnt := 0
	for {
		select {
		case v := <-resCh:
			fmt.Printf("%v [%v/%v]\t%s", time.Now().String(), cnt+1, total, v)
			cnt += 1
			if cnt == total {
				return
			}
		case v := <-logCh:
			fmt.Printf("%v [WORKER] %s\n", time.Now().String(), v)
		default:
		}
	}
}
