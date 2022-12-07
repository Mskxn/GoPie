package passes

import (
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ast/astutil"
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
