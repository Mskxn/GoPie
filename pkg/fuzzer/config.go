package fuzzer

import "toolkit/pkg/bug"

type Config struct {
	Bin string
	Fn  string

	MaxWorker    int
	MaxExecution int
	SingleCrash  bool

	LogLevel string
	LogCh    chan string

	UseFeedBack     bool
	UseCoveredSched bool
	UseStates       bool
	UseAnalysis     bool
	UseMutate       bool

	TimeOut        int
	RecoverTimeOut int
	InitTurnCnt    int
	MaxQuit        int

	BugSet *bug.BugSet
}

func DefaultConfig() *Config {
	c := &Config{
		MaxWorker:       5,
		MaxExecution:    100000,
		SingleCrash:     false,
		LogLevel:        "normal",
		UseFeedBack:     true,
		UseStates:       true,
		UseCoveredSched: true,
		UseAnalysis:     true,
		UseMutate:       true,
		TimeOut:         20,
		RecoverTimeOut:  100,
		InitTurnCnt:     10,
		MaxQuit:         500,
	}
	return c
}

func GokerConfig() *Config {
	c := &Config{
		MaxWorker:       5,
		MaxExecution:    100000,
		SingleCrash:     true,
		LogLevel:        "normal",
		UseFeedBack:     true,
		UseStates:       true,
		UseCoveredSched: true,
		UseAnalysis:     true,
		UseMutate:       true,
		TimeOut:         5,
		RecoverTimeOut:  100,
		InitTurnCnt:     10,
		MaxQuit:         100000,
	}
	return c
}

func NewConfig(bin, fn string, logCh chan string, bugset *bug.BugSet, typ string) *Config {
	var c *Config
	switch typ {
	case "goker":
		c = GokerConfig()
	default:
		c = DefaultConfig()
	}
	c.Bin = bin
	c.Fn = fn
	c.LogCh = logCh
	c.BugSet = bugset
	return c
}
