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
		out = "a_test.go"
	}
	if cmd.Opts.File == "" {
		log.Fatalf("Need source file")
	}
	passes.RunLeakPass(cmd.Opts.File, out)
}
