package inst

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token"
	"io/fs"
	"os"
	"strings"
	"sync/atomic"

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
	newCallSelector := &ast.SelectorExpr{
		X:   newIdentPkg,
		Sel: newIdentCallee,
	}
	newCall := &ast.CallExpr{
		Fun:      newCallSelector,
		Lparen:   token.NoPos,
		Args:     vecExprArg,
		Ellipsis: token.NoPos,
		Rparen:   token.NoPos,
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

var id_cnt int64 = 0

func getNewOpID() int64 {
	return atomic.AddInt64(&id_cnt, 1)
}

func IsTestFunc(n ast.Node) bool {
	switch concrete := n.(type) {
	case *ast.FuncDecl:
		name := concrete.Name.Name
		params := concrete.Type.Params.List

		if !strings.HasPrefix(name, "Test") {
			return false
		}

		if len(params) != 1 {
			return true
		}

		check_ok := false
		names := params[0].Names
		if len(names) != 1 || names[0].Name != "t" {
			return false
		}

		if v, ok := params[0].Type.(*ast.StarExpr); ok {
			if vv, ok := v.X.(*ast.SelectorExpr); ok {
				if vvv, ok := vv.X.(*ast.Ident); ok {
					if vv.Sel.Name == "T" && vvv.Name == "testing" {
						check_ok = true
					}
				}
			}
		}

		return check_ok
	default:
		return false
	}
}
