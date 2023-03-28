package fuzzer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/seed"
)

var (
	initTurnCnt = 100
	maxFuzzCnt  = 10000
	quitcnt     = 500
	singleCrash = false
	debug       = false
	info        = false
	normal      = true
)

type Monitor struct {
	etimes    int32
	max       int32
	maxscore  int
	MaxWorker int
}

type RunContext struct {
	In  Input
	Out Output
}

var workerID uint32

func (m *Monitor) Start(bin string, fn string, llevel string, logCh chan string, usefeedback bool, timeout, recovertimeout int, bugset *bug.BugSet) (bool, []string) {
	if m.max == int32(0) {
		m.max = int32(maxFuzzCnt)
	}
	if m.MaxWorker == 0 {
		m.MaxWorker = 16
	}
	m.maxscore = 10
	switch llevel {
	case "debug":
		debug = true
		info = true
	case "info":
		info = true
	default:
	}
	corpus := NewCorpus()
	fncov := feedback.NewCov()

	wid := atomic.AddUint32(&workerID, 1)
	ch := make(chan RunContext, m.MaxWorker*10)
	cancel := make(chan struct{})
	quit := quitcnt

	var wg sync.WaitGroup
	dowork := func() {
		defer wg.Done()
		defer atomic.AddInt32(&m.etimes, 1)
		for {
			var c, ht *Chain
			if usefeedback {
				c, ht = corpus.Get()
				if debug {
					logCh <- fmt.Sprintf("[Corpus] SIZE %v, GET %s, ATTACK %s", corpus.GSize(), c.ToString(), ht.ToString())
				}
			}
			e := Executor{}
			in := Input{
				c:              c,
				ht:             ht,
				cmd:            bin,
				args:           []string{"-test.v", "-test.run", fn},
				timeout:        timeout,
				recovertimeout: recovertimeout,
			}
			o := e.Run(in)
			ch <- RunContext{In: in, Out: o}
			select {
			case <-cancel:
				break
			default:
			}
		}
	}

	for i := 0; i < m.MaxWorker; i++ {
		go dowork()
	}

	for {
		if m.etimes > m.max {
			close(cancel)
			return false, []string{}
		}
		ctx := <-ch
		var inputc, htc string
		if ctx.In.c != nil {
			inputc = ctx.In.c.ToString()
		} else {
			inputc = "empty chain"
		}
		if ctx.In.ht != nil {
			htc = ctx.In.ht.ToString()
		} else {
			htc = "empty chain"
		}
		if debug {
			logCh <- fmt.Sprintf("[WORKER %v] Input: %s\nAttack: %s", wid, inputc, htc)
		}
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			// ignore normal test fail
			if strings.Contains(ctx.Out.O, "panic") || strings.Contains(ctx.Out.O, "found unexpected goroutines") {
				tfs := bug.TopF(ctx.Out.O)
				exist := bugset.Exist(tfs, fn)
				if !exist {
					detail := []string{inputc, strconv.FormatInt(int64(atomic.LoadInt32(&m.etimes)), 10), ctx.Out.O}
					if normal {
						logCh <- fmt.Sprintf("[WORKER %v] CRASH [%v] \n %s\n%s\n%s\n%s", wid, bugset.Size(), inputc, htc, detail[1], ctx.Out.O)
					}
					if debug {
						topfs := ""
						for _, f := range tfs {
							topfs += f + "\n"
						}
						logCh <- fmt.Sprintf("[BUG] [%s] TopF : \n%s", fn, topfs)
					}
					if singleCrash {
						close(cancel)
						return true, detail
					}
				}
			}
		}
		op_st, orders := feedback.ParseLog(ctx.Out.Trace)
		schedcov := feedback.ParseCovered(ctx.Out.O)
		schedres, coveredinput := ColorCovered(ctx.Out.O, ctx.In.c)
		attackres, _ := ColorCovered(ctx.Out.O, ctx.In.ht)

		cov := feedback.Log2Cov(orders)
		score := cov.Score()
		if len(schedcov) != 0 {
			score += (len(schedcov) / (ctx.In.c.Len())) * len(schedcov) * 10
		}
		if score > m.maxscore {
			m.maxscore = score
		}
		energy := int(float64(score+1) / float64(m.maxscore+1) * 100)
		if debug {
			logCh <- fmt.Sprintf("[WORKER %v] score : %v\tenergy %v", wid, score, energy)
		}

		var init bool
		if atomic.LoadInt32(&m.etimes) < int32(initTurnCnt) {
			init = true
			seeds := seed.SRDOAnalysis(op_st)
			seeds = append(seeds, seed.SODRAnalysis(op_st)...)
			if debug {
				if len(seeds) != 0 {
					logCh <- fmt.Sprintf("[WORKER %v] %v SEEDS %s ...", wid, len(seeds), seeds[0].ToString())
				}
			}
			corpus.GUpdateSeed(seeds)
		}
		ok := fncov.Merge(cov)
		if (init || ok) && inputc != "empty chain" && coveredinput.Len() != 0 {
			if info {
				logCh <- fmt.Sprintf("[WORKER %v] NEW score: [%v/%v] Input:%s\t Attack:%s", wid, score, m.maxscore, schedres, attackres)
			}
			m := Mutator{Cov: fncov}
			ncs, hts := m.mutate(coveredinput, energy)
			if debug {
				logCh <- fmt.Sprintf("[WORKER %v] MUTATE %s", wid, coveredinput.ToString())
			}
			corpus.Update(ncs, hts)
			quit = quitcnt
		} else {
			quit -= 1
			if quit <= 0 {
				if info {
					logCh <- fmt.Sprintf("[WORKER %v] Fuzzing seems useless, QUIT", wid)
				}
				return false, []string{}
			}
		}
	}
}
