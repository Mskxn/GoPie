package fuzzer

import (
	"testing"
	"time"
)

type Fuzzer struct {
	target  func(*testing.T, uint64, uint64)
	timeout time.Duration
}
