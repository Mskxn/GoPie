package main

import "testing"

func TestLite(t *testing.T) {
	Lite("C:\\Users\\Msk\\GolandProjects\\toolkit\\testdata\\testdata.test.exe", "", "debug", 5, 50, 32)
}

func TestRQ2(t *testing.T) {
	RQ2("C:\\Users\\Msk\\GolandProjects\\toolkit\\testdata\\testdata.test.exe")
}

func TestFull(t *testing.T) {
	Full("C:\\Users\\Msk\\GolandProjects\\toolkit\\testdata\\", "debug", 5)
}
