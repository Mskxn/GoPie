package main

import (
	"fmt"
	"toolkit/pkg/fuzzer"
)

func lite() {
	resCh := make(chan string, 10)

	dowork := func(bin string) {
		m := fuzzer.Monitor{}
		ok, detail := m.Start(bin, "_1")
		var res string
		if ok {
			res = fmt.Sprintf("%s\tFAIL\t%s\n", bin, detail[1])
		} else {
			res = fmt.Sprintf("%s\tPASS\n", bin)
		}
		resCh <- res
	}

	for _, bin := range Bins {
		go dowork(bin)
	}
	cnt := len(Bins)
	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s", len(Bins)-cnt+1, len(Bins), v)
			cnt -= 1
			if cnt == 0 {
				return
			}
		default:
		}
	}
}
