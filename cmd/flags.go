package cmd

import (
	flags "github.com/jessevdk/go-flags"
	"os"
)

var Opts struct {
	File string `long:"file" description:"Instrument single go source file"`
	Out  string `long:"out" description:"Output instrumented golang source file to the given file. Only allow when instrumenting single golang source file"`
}

func ParseFlags() {
	if _, err := flags.Parse(&Opts); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}
}
