package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func baselineB() {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		cnt := 0
		bound := 100
		for {
			command := exec.Command(path, "-test.v", "-test.run", "_1")
			var out, out2 bytes.Buffer
			command.Stdout = &out
			command.Stderr = &out2
			err := command.Run()
			if err != nil {
				if strings.Contains(out.String(), "FAIL") || strings.Contains(out.String(), "fatal") {
					// log.Printf("%s", out.String())
					break
				}
			}
			if cnt > bound {
				resCh <- fmt.Sprintf("%s\tPASS\t%v", path, cnt)
				return
			}
			cnt += 1
		}
		resCh <- fmt.Sprintf("%s\tFAIL\t%v", path, cnt)
	}

	all := len(Bins)
	for _, p := range Bins {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(Bins)-all+1, len(Bins), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
