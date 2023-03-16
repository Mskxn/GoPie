package fuzzer

import "testing"

func TestMonitor(t *testing.T) {
	m := Monitor{}
	m.Start("C:\\Users\\Msk\\GolandProjects\\toolkit\\toolkit.test.exe", "_1")
}
