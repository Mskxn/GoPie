package bug

import (
	"log"
	"testing"
)

var testlog = "[WORKER] [WORKER 10] CRASH ({661424963617}, {296352743425})({296352743425}, {661424963621})({661424963621}, {661424963622})({661424963622}, {661424963599})({661424963599}, {661424963609})({661424963609}, {661424963662})    599\n" +
	"=== RUN   TestJournalingNoLocals_1\n" +
	"=== PAUSE TestJournalingNoLocals_1\n" +
	"=== CONT  TestJournalingNoLocals_1\n" +
	"[COVERED] {661424963617, 296352743425}\n" +
	"ched.go:275: found unexpected goroutines:\n" +
	"[Goroutine 124 in state select, with github.com/ethereum/go-ethereum/core/txpool.(*TxPool).scheduleReorgLoop on top of the stack:\n" +
	"oroutine 124 [select]:\n" +
	"ithub.com/ethereum/go-ethereum/core/txpool.(*TxPool).scheduleReorgLoop(0xc0003be900)\n" +
	"/tool/go-ethereum/core/txpool/txpool.go:1264 +0x2f4\n" +
	"reated by github.com/ethereum/go-ethereum/core/txpool.NewTxPool\n" +
	"/tool/go-ethereum/core/txpool/txpool.go:324 +0x6dc\n" +
	"\n" +
	"oroutine 125 in state select, with github.com/ethereum/go-ethereum/core/txpool.(*TxPool).loop on top of the stack:\n" +
	"oroutine 125 [select]:\n" +
	"ithub.com/ethereum/go-ethereum/core/txpool.(*TxPool).loop(0xc0003be900)\n" +
	"/tool/go-ethereum/core/txpool/txpool.go:368 +0x2d7\n" +
	"reated by github.com/ethereum/go-ethereum/core/txpool.NewTxPool\n" +
	"/tool/go-ethereum/core/txpool/txpool.go:341 +0x9a5\n" +
	"]\n" +
	"--- FAIL: TestJournalingNoLocals_1 (22.51s)\n"

func TestTopF(t *testing.T) {
	fs := TopF(testlog)
	for _, f := range fs {
		t.Log(f)
	}
}

func TestBugSet(t *testing.T) {
	bs := NewBugSet()
	log.Printf("%v", bs.Add(TopF(testlog)))
	log.Printf("%v", bs.Add(TopF(testlog)))
}
