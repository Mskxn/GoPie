package testdata

import "testing"

func fn() {
	return
}

func TestFn(t *testing.T) {
	fn()
	a := 1
	a = a + 1
}
