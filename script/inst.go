package main

import (
	"bytes"
	"fmt"
	"os/exec"
)

func inst() {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		command := exec.Command("C:\\Users\\Msk\\GolandProjects\\toolkit\\bin\\go_build_toolkit_cmd_sw.exe", "--file", path)
		var out, out2 bytes.Buffer
		command.Stdout = &out
		command.Stderr = &out2
		err := command.Run()
		if err == nil {
			resCh <- fmt.Sprintf("Handle\t%s OK", path)
		} else {
			resCh <- fmt.Sprintf("Handle\t%s FAIL", path)
		}
	}

	all := len(InstPaths)
	for _, p := range InstPaths {
		go dowork(p)
	}

	for {
		select {
		case v := <-resCh:
			fmt.Println(v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
