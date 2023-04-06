package main

import (
	"fmt"
	"toolkit/cmd"
	"toolkit/pkg/bug"
	"toolkit/pkg/fuzzer"
)

func RQ1(path string) {
	resCh := make(chan string, 10000)
	logCh := make(chan string, 10000)
	bugset := bug.NewBugSet()

	bins := cmd.ListFiles(path, func(s string) bool {
		return true
	})

	dowork := func(bin string, fn string) {
		m := fuzzer.Monitor{}

		cfg := fuzzer.NewConfig(bin, fn, logCh, bugset, "goker")

		ok, detail := m.Start(cfg, &fuzzer.Visitor{})
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", bin, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", bin)
		}
		resCh <- res
	}

	for _, bin := range bins {
		go func(bin string) {
			fmt.Printf("[FUZZER] Start %s\n", bin)
			tests := ListTests(bin)
			for _, test := range tests {
				go dowork(bin, test)
			}
		}(bin)
	}

	cnt := len(bins)
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s", len(bins)-cnt+1, len(bins), v)
			cnt -= 1
			if cnt == 0 {
				return
			}
		case v := <-logCh:
			fmt.Printf("[WORKER] %s\n", v)
		default:
		}
	}
}
