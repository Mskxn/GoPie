package cmd

import (
	"testing"
)

func TestListFiles(t *testing.T) {
	for _, file := range ListFiles("../testdata/project/blocking/") {
		t.Log(file)
	}
}
