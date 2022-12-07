package main

import (
	"log"
	"toolkit/cmd"
	"toolkit/pkg/passes"
)

func main() {
	cmd.ParseFlags()
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
	passes.RunFuzzPass(out, out)
}
