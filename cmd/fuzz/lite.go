package main

import (
	"fmt"
	"toolkit/pkg/bug"
	"toolkit/pkg/fuzzer"
)

func Lite(bin string, llevel string, timeout, recovertimeout int, maxworker int) {
	resCh := make(chan string, 100)
	logCh := make(chan string, 100)
	bugset := bug.NewBugSet()

	dowork := func(bin string, fn string) {
		m := fuzzer.Monitor{MaxWorker: maxworker}
		ok, detail := m.Start(bin, fn, llevel, logCh, true, timeout, recovertimeout, bugset)
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", bin, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", bin)
		}
		resCh <- res
	}
	fmt.Printf("[FUZZER] Start %s\n", bin)
	tests := ListTests(bin)
	for _, test := range tests {
		fmt.Printf("[WORKER] Start %s\n", test)
		go dowork(bin, test)
	}
	cnt := len(tests)
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s", len(tests)-cnt+1, len(tests), v)
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
