package main

import (
	"fmt"
	"os"
	"toolkit/script/data_analysis"
)

func main() {
	task := os.Args[1]
	switch task {
	case "baselineA":
		baselineA()
	case "baselineB":
		baselineB()
	case "lite":
		lite()
	case "inst":
		inst()
	case "bins":
		bins()
	case "graph":
		data_analysis.GenGraph()
	default:
		fmt.Println("error argument" + " " + task)
	}
}
