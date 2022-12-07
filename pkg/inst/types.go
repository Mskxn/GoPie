package inst

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"log"
)

type InstPass interface {
	// Deps returns a list of dependent passes
	// Deps() []string
	Before(iCtx *InstContext)
	GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool
	GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool
	After(iCtx *InstContext)
}

// InstContext contains all information needed to instrument one single Golang source code.
type InstContext struct {
	File            string
	OriginalContent []byte
	FS              *token.FileSet
	AstFile         *ast.File
	Type            *types.Info
	Metadata        map[string]interface{} // user can set custom metadata come along with instrumentation context
}

func (i *InstContext) SetMetadata(key string, value interface{}) {
	i.Metadata[key] = value
}

func (i *InstContext) GetMetadata(key string) (interface{}, bool) {
	val, exist := i.Metadata[key]
	return val, exist
}

func NewInstContext(goSrcFile string) (*InstContext, error) {
	oldSource, err := ioutil.ReadFile(goSrcFile)
	if err != nil {
		return nil, err
	}

	fs := token.NewFileSet()
	astF, err := parser.ParseFile(fs, goSrcFile, oldSource, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	conf := types.Config{
		Importer: importer.ForCompiler(fs, "source", nil),
		Error: func(err error) {
			log.Printf("'%s' type checker: %s", goSrcFile, err)
		},
	}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
		Defs:  make(map[*ast.Ident]types.Object),
		Uses:  make(map[*ast.Ident]types.Object),
	}

	conf.Check(astF.Name.Name, fs, []*ast.File{astF}, info)

	return &InstContext{
		File:            goSrcFile,
		OriginalContent: oldSource,
		FS:              fs,
		Type:            info,
		AstFile:         astF,
		Metadata:        make(map[string]interface{}),
	}, nil
}

type Import struct {
	Name string
	Path string
	Need string
}
