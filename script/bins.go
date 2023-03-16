package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func bins() {
	resCh := make(chan string, 100)
	dowork := func(path string) {
		dirs := strings.Split(path, "/")
		fn := dirs[len(dirs)-1]
		fn = fn[0:len(fn)-len("_test.go")] + ".exe"
		opath := "./testdata/project/bins/" + fn
		command := exec.Command("C:\\Users\\Msk\\go\\go1.19\\bin\\go", "test", "-o", opath, "-c", path)
		var out, out2 bytes.Buffer
		command.Stdout = &out
		command.Stderr = &out2
		err := command.Run()
		if err == nil {
			resCh <- fmt.Sprintf("Handle\t%s OK", opath)
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
			fmt.Printf("[%v/%v]\t%s\n", len(Bins)-all+1, len(Bins), v)
			all -= 1
			if all == 0 {
				return
			}
		default:
		}
	}
}
