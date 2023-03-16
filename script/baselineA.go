package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func baselineA() {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		cnt := 0
		bound := 1000
		for {
			command := exec.Command("C:\\Users\\Msk\\go\\go1.19\\bin\\go", "test", "-timeout=1s", "-count=1", path)
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

	all := len(Paths)
	for _, p := range Paths {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(Paths)-all+1, len(Paths), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
