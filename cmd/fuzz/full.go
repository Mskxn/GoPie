package main

import (
	"fmt"
	"sync"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func Full(path string, llevel string, feature string, maxworker int) {
	resCh := make(chan string, 1000)
	logCh := make(chan string, 1000)

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
	var mu sync.Mutex
	bugset := bug.NewBugSet()

	fuzzfn := func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
		<-limit
		defer func() {
			limit <- struct{}{}
		}()
		m := &fuzzer.Monitor{}
		ok, detail := m.Start(cfg, v, limit)
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", cfg.Fn, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", cfg.Fn)
		}
		resCh <- res
		mu.Lock()
		defer mu.Unlock()
		if cnt, ok := bin2cnt[cfg.Bin]; ok {
			if cnt == 1 {
				delete(bin2vst, cfg.Bin)
			} else {
				bin2cnt[cfg.Bin] -= 1
			}
		}
	}

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
		for _, test := range tests {
			cfg := fuzzer.DefaultConfig()
			// shared bugset
			cfg.BugSet = bugset
			cfg.Bin = bin
			cfg.Fn = test
			cfg.MaxWorker = 2
			cfg.TimeOut = 30
			cfg.RecoverTimeOut = 1000
			cfg.LogCh = logCh
			cfg.MaxQuit = 64
			cfg.MaxExecution = 100000
			cfg.LogLevel = llevel
			if feature == "mu" {
				cfg.UseMutate = false
			}
			if feature == "fb" {
				cfg.UseFeedBack = false
			}
			go fuzzfn(bin2vst[bin], cfg)
			total += 1
		}
	}

	defer fmt.Printf("[Fuzzer] Finish\n")
	cnt := 0
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s", cnt+1, total, v)
			cnt += 1
			if cnt == total {
				return
			}
		case v := <-logCh:
			fmt.Printf("[WORKER] %s\n", v)
		default:
		}
	}
}
