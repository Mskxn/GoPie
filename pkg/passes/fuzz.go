package passes

import (
	"go/ast"
	"go/token"
	"log"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

func RunFuzzPass(in, out string) error {
	p := fuzzPass{}
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

type fuzzPass struct {
}

var (
	FuzzNeedInst   = "NEED_FUZZ_INST"
	FuzzImportName = "sched"
	FuzzImportPath = "toolkit/pkg/sched"
)

func (p *fuzzPass) Before(ctx *InstContext) {
	ctx.SetMetadata(NeedInst, false)
}

func (p *fuzzPass) After(ctx *InstContext) {
	if v, ok := ctx.GetMetadata(NeedInst); ok && v.(bool) {
		AddImport(ctx.FS, ctx.AstFile, ImportName, ImportPath)
	}
}

func (p *fuzzPass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
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
				testname := name
				testfunc := concrete
				fuzzDecl := genFuzzDecl(testname, testfunc)
				iCtx.AstFile.Decls = append(iCtx.AstFile.Decls, fuzzDecl)
			}
		}
		return true
	}
}

func (p *fuzzPass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}

func genFuzzDecl(name string, fn *ast.FuncDecl) *ast.FuncDecl {
	fuzzname := "FuzzGen" + name[4:]

	fuzztyp := &ast.FuncType{
		Params: &ast.FieldList{
			List: []*ast.Field{fn.Type.Params.List[0]},
		},
	}
	fuzztyp.Params.List = append(fuzztyp.Params.List,
		&ast.Field{
			Names: []*ast.Ident{&ast.Ident{Name: "sched_wait_bitmap"}, &ast.Ident{Name: "sched_send_bitmap"}},
			Type:  &ast.Ident{Name: "uint64"},
		})

	testgo := &ast.GoStmt{Call: &ast.CallExpr{
		Fun: &ast.FuncLit{
			Type: &ast.FuncType{
				Params: &ast.FieldList{
					List: []*ast.Field{},
				},
			},
			Body: fn.Body,
		},
	}}
	checker := NewDeferExpr("sched", "Leakcheck", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.IDENT,
			Value: "t",
		},
	})
	setenable := NewArgCall("sched", "SetEnableWait", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.IDENT,
			Value: "sched_wait_bitmap",
		},
	})

	setenable2 := NewArgCall("sched", "SetEnableSend", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.IDENT,
			Value: "sched_send_bitmap",
		},
	})

	fuzzbody := &ast.BlockStmt{List: []ast.Stmt{
		checker,
		&ast.ExprStmt{setenable},
		&ast.ExprStmt{setenable2},
		testgo,
	}}

	fuzzdecl := &ast.FuncDecl{
		Name: &ast.Ident{Name: fuzzname},
		Type: &ast.FuncType{
			Params: &ast.FieldList{
				List: []*ast.Field{
					&ast.Field{
						Names: []*ast.Ident{
							&ast.Ident{Name: "f"},
						},
						Type: &ast.StarExpr{
							X: &ast.SelectorExpr{
								X: &ast.Ident{
									Name: "testing",
								},
								Sel: &ast.Ident{
									Name: "F",
								},
							},
						},
					},
				},
			},
		},
		Body: &ast.BlockStmt{
			List: []ast.Stmt{
				&ast.ExprStmt{
					NewArgCall("f", "Add", []ast.Expr{
						NewArgCall("", "uint64", []ast.Expr{
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "32",
							},
						}),
						NewArgCall("", "uint64", []ast.Expr{
							&ast.BasicLit{
								Kind:  token.INT,
								Value: "64",
							},
						}),
					}),
				},
				&ast.ExprStmt{
					NewArgCall("f", "Fuzz", []ast.Expr{
						&ast.FuncLit{
							Type: fuzztyp,
							Body: fuzzbody,
						},
					}),
				},
			},
		},
	}
	return fuzzdecl
}
