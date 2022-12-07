package passes

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"sync/atomic"
	"toolkit/pkg/inst"

	"golang.org/x/tools/go/ast/astutil"
)

func AddImport(fs *token.FileSet, ast *ast.File, name string, path string) error {
	for _, vecImportSpec := range astutil.Imports(fs, ast) {
		for _, importSpec := range vecImportSpec {
			if importSpec != nil && importSpec.Name != nil {

				if importSpec.Name.Name == name && importSpec.Path.Value == path { // instrumented before
					// ideomatic
					return nil
				}
			}

		}
	}

	ok := astutil.AddNamedImport(fs, ast, name, path)
	if !ok {
		return fmt.Errorf("failed to add import %s %s", name, path)
	}
	return nil
}

// DumpAstFile serialized AST to given file
func DumpAstFile(fset *token.FileSet, astFile *ast.File, dstFile string) error {
	if astFile == nil {
		return fmt.Errorf("found nil ast file for %s", dstFile)
	}
	fi, err := os.Stat(dstFile)
	var mode fs.FileMode
	if err != nil {
		// return any error except not exist
		if !os.IsNotExist(err) {
			return err
		} else {
			// if file not exist, use default mode
			mode = 0666
		}
	} else {
		// if file exist, use same mode
		mode = fi.Mode()
	}
	w, err := os.OpenFile(dstFile, os.O_CREATE|os.O_WRONLY, mode)
	defer w.Close()
	if err != nil {
		return err
	}
	err = format.Node(w, fset, astFile)
	if err != nil {
		log.Fatalf("%v", err.Error())
		return err
	}
	return nil
}

func NewArgCall(strPkg, strCallee string, vecExprArg []ast.Expr) *ast.CallExpr {
	newIdentPkg := &ast.Ident{
		NamePos: token.NoPos,
		Name:    strPkg,
		Obj:     nil,
	}
	newIdentCallee := &ast.Ident{
		NamePos: token.NoPos,
		Name:    strCallee,
		Obj:     nil,
	}
	var newCall *ast.CallExpr
	if strPkg != "" {
		fun := &ast.SelectorExpr{
			X:   newIdentPkg,
			Sel: newIdentCallee,
		}
		newCall = &ast.CallExpr{
			Fun:      fun,
			Lparen:   token.NoPos,
			Args:     vecExprArg,
			Ellipsis: token.NoPos,
			Rparen:   token.NoPos,
		}
	} else {
		newCall = &ast.CallExpr{
			Fun:      newIdentCallee,
			Lparen:   token.NoPos,
			Args:     vecExprArg,
			Ellipsis: token.NoPos,
			Rparen:   token.NoPos,
		}
	}
	return newCall
}

func NewArgCallExpr(strPkg, strCallee string, vecExprArg []ast.Expr) *ast.ExprStmt {
	newCall := NewArgCall(strPkg, strCallee, vecExprArg)
	newExpr := &ast.ExprStmt{X: newCall}
	return newExpr
}

func NewDeferExpr(strPkg, strCallee string, vecExprArg []ast.Expr) *ast.DeferStmt {
	newCall := NewArgCall(strPkg, strCallee, vecExprArg)
	newExpr := &ast.DeferStmt{Call: newCall}
	return newExpr
}

var (
	TraceImportPath = "toolkit/pkg/trace"
	TraceImportName = "trace"
	IDName          = "trace_goroutine_id_xxx"
)

// NewInstContext creates a InstContext by given Golang source file
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

var id_cnt uint64 = 0

func getNewOpID() uint64 {
	return atomic.AddUint64(&id_cnt, 1)
}

func getSelectorCallerType(iCtx *inst.InstContext, selExpr *ast.SelectorExpr) string {
	if callerIdent, ok := selExpr.X.(*ast.Ident); ok {
		if callerIdent.Obj != nil {
			if objStmt, ok := callerIdent.Obj.Decl.(*ast.AssignStmt); ok {
				if objIdent, ok := objStmt.Lhs[0].(*ast.Ident); ok {
					if to := iCtx.Type.Defs[objIdent]; to == nil || to.Type() == nil {
						return ""
					} else {
						return to.Type().String()
					}

				}
			}
		}
	}

	return ""
}

func SelectorCallerHasTypes(iCtx *inst.InstContext, selExpr *ast.SelectorExpr, trueIfUnknown bool, tys ...string) bool {
	t := getSelectorCallerType(iCtx, selExpr)
	if t == "" && trueIfUnknown {
		return true
	}
	for _, ty := range tys {
		if ty == t {
			return true
		}
	}

	return false
}

func getContextString(iCtx *inst.InstContext, c *astutil.Cursor) string {
	return iCtx.FS.Position(c.Node().Pos()).String()
}
