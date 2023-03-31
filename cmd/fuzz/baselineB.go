package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"toolkit/cmd"
)

const (
	maxParallel = 5
)

func BaselineB(dir string) {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		cnt := 1
		bound := 1000
		var bt string
		for {
			command := exec.Command(path, "-test.v", "-test.run", "_1$")
			var out, out2 bytes.Buffer
			command.Stdout = &out
			command.Stderr = &out2
			err := command.Run()
			if err != nil {
				// log.Printf("%s", out.String())
				bt = out2.String()
				break
			}
			if cnt > bound {
				resCh <- fmt.Sprintf("%s\tPASS\t%v", path, cnt)
				return
			}
			cnt += 1
		}
		resCh <- fmt.Sprintf("%s\tFAIL\t%v\t%s", path, cnt, bt)
	}

	bins := cmd.ListFiles(dir, func(s string) bool {
		return true
	})

	all := len(bins)
	for _, p := range bins {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(bins)-all+1, len(bins), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
