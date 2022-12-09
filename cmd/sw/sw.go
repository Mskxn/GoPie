package main

import (
	"log"
	"strings"
	"sync"
	"toolkit/cmd"
	"toolkit/pkg/passes"
)

func main() {
	cmd.ParseFlags()
	if cmd.Opts.Dir != "" {
		files := cmd.ListFiles(cmd.Opts.Dir)
		var wg sync.WaitGroup
		for _, file := range files {
			if strings.HasSuffix(file, ".go") {
				log.Println("Inst " + file)
				wg.Add(1)
				go func(file string) {
					passes.RunSWPass(file, file)
					passes.RunChannelPass(file, file)
					passes.RunSelectPass(file, file)
					passes.RunFuzzPass(file, file)
					wg.Done()
				}(file)
			}
		}
		wg.Wait()
	} else {
		out := ""
		if cmd.Opts.Out == "" {
			out = cmd.Opts.File
		} else {
			out = cmd.Opts.Out
		}
		if cmd.Opts.File == "" {
			log.Fatalf("Need source file")
		}
		passes.RunSWPass(cmd.Opts.File, out)
		passes.RunChannelPass(out, out)
		passes.RunSelectPass(out, out)
		passes.RunFuzzPass(out, out)
	}
}
