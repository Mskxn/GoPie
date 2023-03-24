package fuzzer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"toolkit/pkg"
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
		for {
			var c *Chain
			if usefeedback {
				c = corpus.GGet()
				if debug {
					logCh <- fmt.Sprintf("[Corpus] SIZE %v, GET %s", corpus.Size(), c.ToString())
				}
			}
			e := Executor{}
			in := Input{
				c:              c,
				cmd:            bin,
				args:           []string{"-test.v", "-test.run", fn},
				timeout:        timeout,
				recovertimeout: recovertimeout,
			}
			o := e.Run(in)
			atomic.AddInt32(&m.etimes, 1)
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
		var inputc string
		if ctx.In.c != nil {
			inputc = ctx.In.c.ToString()
		} else {
			inputc = "empty chain"
		}
		if debug {
			logCh <- fmt.Sprintf("[WORKER %v] recv work : %s", wid, inputc)
		}
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			// ignore normal test fail
			if strings.Contains(ctx.Out.O, "panic") || strings.Contains(ctx.Out.O, "found unexpected goroutines") {
				new := !bugset.Add(bug.TopF(ctx.Out.O), fn)
				if new {
					detail := []string{inputc, strconv.FormatInt(int64(atomic.LoadInt32(&m.etimes)), 10), ctx.Out.O}
					if normal {
						logCh <- fmt.Sprintf("[WORKER %v] CRASH %s\t%s\n%s", wid, inputc, detail[1], ctx.Out.O)
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
		schedres := ""
		coverdinput := Chain{item: []*pkg.Pair{}}
		if inputc != "empty chain" {
			idx := 0
			for _, p := range ctx.In.c.item {
				if idx < len(schedcov) {
					cov := schedcov[idx]
					if p.Prev.Opid == cov[0] && p.Next.Opid == cov[1] {
						schedres += fmt.Sprintf("\033[1;31;40m%s\033[0m", p.ToString())
						coverdinput.item = append(coverdinput.item, p)
						idx += 1
					}
				} else {
					schedres += p.ToString()
				}
			}
		}

		cov := feedback.Log2Cov(orders)
		score := cov.Score()
		if len(schedcov) != 0 {
			score += len(schedcov) * 10
		}
		if score > m.maxscore {
			m.maxscore = score
		}
		energy := int(float64(score) / float64(m.maxscore+1) * 5)

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
		if (init || ok) && inputc != "empty chain" {
			if info {
				logCh <- fmt.Sprintf("[WORKER %v] NEW %s\tscore: [%v/%v]", wid, schedres, score, m.maxscore)
			}
			if coverdinput.Len() != 0 {
				m := Mutator{Cov: fncov}
				ncs := m.mutate(ctx.In.c, energy)
				if debug {
					logCh <- fmt.Sprintf("[WORKER %v] MUTATE %s", wid, coverdinput.ToString())
				}
				corpus.GSUpdate(ncs)
			}
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
