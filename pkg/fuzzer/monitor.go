package fuzzer

import (
	"fmt"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"toolkit/pkg/bug"
	"toolkit/pkg/feedback"
	"toolkit/pkg/seed"
)

var (
	debug  = false
	info   = false
	normal = true
)

type Monitor struct {
	etimes   int32
	max      int32
	maxscore int
}

type RunContext struct {
	In  Input
	Out Output
}

var workerID uint32

func (m *Monitor) Start(cfg *Config, visitor *Visitor, ticket chan struct{}) (bool, []string) {
	if m.max == int32(0) {
		m.max = int32(cfg.MaxExecution)
	}
	m.maxscore = 10
	switch cfg.LogLevel {
	case "debug":
		debug = true
		info = true
	case "info":
		info = true
	default:
	}
	var fncov *feedback.Cov
	var corpus *Corpus

	if visitor.V_cov == nil {
		fncov = feedback.NewCov()
	} else {
		fncov = visitor.V_cov
	}

	if visitor.V_corpus == nil {
		corpus = NewCorpus()
	} else {
		corpus = visitor.V_corpus
	}
	visitor.V_corpus = corpus
	visitor.V_cov = fncov

	wid := atomic.AddUint32(&workerID, 1)
	ch := make(chan RunContext, cfg.MaxWorker*10)
	cancel := make(chan struct{})
	quit := cfg.MaxQuit

	dowork := func() {
		for {
			var c, ht *Chain
			if cfg.UseMutate {
				c, ht = corpus.Get()
			} else {
				corpus.IncFetchCnt()
			}
			e := Executor{}
			in := Input{
				c:              c,
				ht:             ht,
				cmd:            cfg.Bin,
				args:           []string{"-test.v", "-test.run", cfg.Fn, "-test.timeout", "1m"},
				timeout:        cfg.TimeOut,
				recovertimeout: cfg.RecoverTimeOut,
			}
			// _, ok := <-ticket
			atomic.AddInt32(&m.etimes, 1)
			timeout := time.After(1 * time.Minute)
			done := make(chan int)
			var o *Output
			go func() {
				t := e.Run(in)
				o = &t
				close(done)
			}()
			select {
			case <-done:
			case <-timeout:
			}
			// if ok {
			//	ticket <- struct{}{}
			//}
			if o == nil {
				continue
			}
			if debug {
				cfg.LogCh <- fmt.Sprintf("%s\t[EXECUTOR] Finish, USE %s", time.Now().String(), o.Time.String())
			}
			ch <- RunContext{In: in, Out: *o}
			select {
			case <-cancel:
				break
			default:
			}
		}
	}

	for i := 0; i < cfg.MaxWorker; i++ {
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
			cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] Input: %s\nAttack: %s", time.Now().String(), wid, inputc, htc)
		}
		// global corpus is not thread safe now
		if ctx.Out.Err != nil {
			// ignore normal test fail
			if ctx.Out.Time < time.Duration(cfg.TimeOut)*time.Second &&
				(strings.Contains(ctx.Out.O, "panic") || strings.Contains(ctx.Out.O, "found unexpected goroutines")) {
				tfs := bug.TopF(ctx.Out.O)
				exist := cfg.BugSet.Exist(tfs, cfg.Fn)
				if !exist {
					detail := []string{inputc, strconv.FormatInt(int64(atomic.LoadInt32(&m.etimes)), 10), ctx.Out.O}
					if normal {
						cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] CRASH [%v] \n %s\n%s\n%s\n%s", time.Now().String(), cfg.BugSet.Size(), inputc, htc, detail[1], ctx.Out.O)
					}
					if debug {
						topfs := ""
						for _, f := range tfs {
							topfs += f + "\n"
						}
						cfg.LogCh <- fmt.Sprintf("%s\t[BUG] [%s] TopF : \n%s", time.Now().String(), cfg.Fn, topfs)
					}
					if cfg.SingleCrash {
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
		score := cov.Score(cfg.UseStates)
		if len(schedcov) != 0 && cfg.UseCoveredSched {
			score += (len(schedcov) / (ctx.In.c.Len())) * len(schedcov) * 10
		}
		if score > m.maxscore {
			m.maxscore = score
		}
		energy := int(float64(score+1) / float64(m.maxscore+1) * 100)
		if debug {
			cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] score : %v\tenergy %v", time.Now().String(), wid, score, energy)
		}

		var init bool
		if atomic.LoadInt32(&m.etimes) < int32(cfg.InitTurnCnt) {
			init = true
			if cfg.UseAnalysis {
				seeds := seed.SRDOAnalysis(op_st)
				seeds = append(seeds, seed.SODRAnalysis(op_st)...)
				if debug {
					if len(seeds) != 0 {
						cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] %v SEEDS %s ...", time.Now().String(), wid, len(seeds), seeds[0].ToString())
					}
				}
				corpus.GUpdateSeed(seeds)
			}
			seeds := seed.RandomSeed(op_st)
			if debug {
				if len(seeds) != 0 {
					cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] %v SEEDS %s ...", time.Now().String(), wid, len(seeds), seeds[0].ToString())
				}
			}
			corpus.GUpdateSeed(seeds)
		}
		ok := fncov.Merge(cov)
		if (init || ok) && inputc != "empty chain" && coveredinput.Len() != 0 {
			corpus.IncSchedCnt(schedres)
			if info {
				cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] NEW score: [%v/%v] Input:%s\t Attack:%s", time.Now().String(), wid, score, m.maxscore, schedres, attackres)
			}
			m := Mutator{Cov: fncov}
			var ncs []*Chain
			var hts map[uint64]map[uint64]struct{}
			if cfg.UseFeedBack {
				if cfg.UseCoveredSched {
					ncs, hts = m.mutate(coveredinput, energy)
				} else {
					ncs, hts = m.mutate(ctx.In.c, energy)
				}
				if debug {
					cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] MUTATE %s", time.Now().String(), wid, coveredinput.ToString())
				}
				corpus.Update(ncs, hts)
			}
			quit = cfg.MaxQuit
		} else {
			quit -= 1
			if quit <= 0 {
				if info {
					cfg.LogCh <- fmt.Sprintf("%s\t[WORKER %v] Fuzzing seems useless, QUIT", time.Now().String(), wid)
				}
				close(cancel)
				return false, []string{}
			}
		}
	}
}
