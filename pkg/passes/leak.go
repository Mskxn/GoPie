package passes

import (
	"go/ast"
	"go/token"
	"log"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func RunLeakPass(in, out string) error {
	p := leakPass{}
	iCtx, err := NewInstContext(in)
	if err != nil {
		log.Fatalf("Analysis source code failed %v", err)
	}
	p.Before(iCtx)
	iCtx.AstFile = astutil.Apply(iCtx.AstFile, p.GetPreApply(iCtx), p.GetPostApply(iCtx)).(*ast.File)
	p.After(iCtx)

	DumpAstFile(iCtx.FS, iCtx.AstFile, out)
	return nil
}

type leakPass struct {
}

var (
	NeedInst   = "NeedGoleak"
	ImportName = "trace"
	ImportPath = "toolkit/pkg/trace"
)

func (p *leakPass) Before(ctx *InstContext) {
	ctx.SetMetadata(NeedInst, false)
}

func (p *leakPass) After(ctx *InstContext) {
	if v, ok := ctx.GetMetadata(NeedInst); ok && v.(bool) {
		AddImport(ctx.FS, ctx.AstFile, ImportName, ImportPath)
	}
}

func (p *leakPass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.FuncDecl:
			name := concrete.Name.Name
			params := concrete.Type.Params.List

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
			if check_ok && strings.HasPrefix(name, "Test") {
				newCall := NewDeferExpr(ImportName, "Check", []ast.Expr{&ast.BasicLit{
					ValuePos: 0,
					Kind:     token.IDENT,
					Value:    "t",
				}})
				concrete.Body.List = append([]ast.Stmt{newCall}, concrete.Body.List...)
				iCtx.SetMetadata(NeedInst, true)
			}
		}
		return true
	}
}

func (p *leakPass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}
