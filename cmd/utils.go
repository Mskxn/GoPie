package cmd

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"toolkit/pkg/inst"
	"toolkit/pkg/utils/gofmt"
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

func HandleSrcFile(src string, reg *inst.PassRegistry, passes []string) error {
	iCtx, err := inst.NewInstContext(src)
	if err != nil {
		return err
	}

	err = inst.Run(iCtx, reg, passes)
	if err != nil {
		return err
	}

	var dst string
	if Opts.Out != "" {
		dst = Opts.Out
	} else {
		// dump AST in-place
		dst = iCtx.File

	}
	err = inst.DumpAstFile(iCtx.FS, iCtx.AstFile, dst)
	if err != nil {
		return err
	}

	// check if output is valid, revert if error happened
	if gofmt.HasSyntaxError(dst) {
		// we simply ignored the instrumented result,
		// and revert the file content back to original version.
		err = ioutil.WriteFile(dst, iCtx.OriginalContent, 0777)
		if err != nil {
			log.Panicf("failed to recover file '%s'", dst)
		}
		log.Printf("recovered '%s' from syntax error\n", dst)
	}

	return nil
}
