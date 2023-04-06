package main

import (
	"fmt"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/fuzzer"
)

func Full(path string, llevel string) {
	resCh := make(chan string, 1000)
	logCh := make(chan string, 1000)

	// control
	nolimit := make(chan struct{})
	close(nolimit)

	fuzzfn := func(v *fuzzer.Visitor, cfg *fuzzer.Config) {
		m := fuzzer.Monitor{}
		ok, detail := m.Start(cfg, v, nolimit)
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", cfg.Fn, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", cfg.Fn)
		}
		resCh <- res
	}

	bins := cmd.ListFiles(path, func(s string) bool {
		return true
	})

	bin2tests := make(map[string][]string)
	bin2vst := make(map[string]*fuzzer.Visitor)
	bugset := bug.NewBugSet()

	// bind tests and visitor to bins
	for _, bin := range bins {
		tests := ListTests(bin)
		bin2tests[bin] = tests

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
			cfg.MaxQuit = 100
			cfg.MaxExecution = 100000
			cfg.LogLevel = llevel
			go fuzzfn(bin2vst[bin], cfg)
			total += 1
		}
	}

	defer fmt.Printf("[Fuzzer] Finish\n")
	cnt := 0
	for {
		select {
		case v := <-logCh:
			fmt.Printf("[WORKER] %s\n", v)
		case res := <-resCh:
			fmt.Printf(res)
			cnt += 1
			if cnt == total {
				return
			}
		default:
		}
	}
}
