package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"toolkit/cmd"
)

func BaselineA(dir string) {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		cnt := 1
		bound := 1000
		for {
			command := exec.Command("go", "test", "-timeout=1s", "-count=1", path)
			var out, out2 bytes.Buffer
			command.Stdout = &out
			command.Stderr = &out2
			err := command.Run()
			if err != nil {
				if strings.Contains(out.String(), "FAIL") || strings.Contains(out.String(), "fatal") || strings.Contains(out.String(), "error") {
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

	files := cmd.ListFiles(dir, func(s string) bool {
		return strings.HasSuffix(s, "_test.go")
	})
	all := len(files)
	for _, p := range files {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Printf("[%v/%v]\t%s\n", len(files)-all+1, len(files), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
