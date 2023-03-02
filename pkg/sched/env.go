package sched

import "sync"

type Config struct {
	sendmap  map[uint64]uint64
	waitmap  map[uint64]uint64
	tracemap map[uint64]uint64
	mu       sync.RWMutex
}

func NewConfig() *Config {
	config := Config{}
	config.sendmap = make(map[uint64]uint64)
	config.waitmap = make(map[uint64]uint64)
	config.tracemap = make(map[uint64]uint64)
	return &config
}
