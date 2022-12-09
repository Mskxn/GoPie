package cmd

import (
	"os"
	"path/filepath"
)

func ListFiles(d string) []string {
	var files []string

	err := filepath.Walk(d, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		panic(err)
	}
	return files
}

func RunTest(path string):
	
