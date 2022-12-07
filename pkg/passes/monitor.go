package passes

import (
	"go/ast"
	"go/token"
	"log"

	"golang.org/x/tools/go/ast/astutil"
)

func RunMonitorPass(in, out string) error {
	p := MonitorPass{}
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

type MonitorPass struct {
}

var (
	MonitorNeedInst   = "NeedMonitor"
	MonitorImportName = "trace"
	MonitorImportPath = "toolkit/pkg/trace"
)

func (p *MonitorPass) Before(ctx *InstContext) {
	ctx.SetMetadata(MonitorImportPath, false)
}

func (p *MonitorPass) After(ctx *InstContext) {
	if v, ok := ctx.GetMetadata(MonitorNeedInst); ok && v.(bool) {
		AddImport(ctx.FS, ctx.AstFile, MonitorImportName, MonitorImportPath)
	}
}

func (p *MonitorPass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.GoStmt:
			WrapGoStmt(concrete, iCtx)
			iCtx.SetMetadata(MonitorNeedInst, true)
		}
		return true
	}
}

func (p *MonitorPass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}

func WrapGoStmt(stmt *ast.GoStmt, iCtx *InstContext) {
	call := stmt.Call

	pos := iCtx.FS.Position(stmt.Pos())

	id_ident := ast.Ident{Name: IDName}
	callbefore := NewArgCall(TraceImportName, "GoStart", []ast.Expr{
		&ast.BasicLit{
			Kind:  token.STRING,
			Value: "\"" + pos.String() + "\"",
		},
	})
	callafter := NewDeferExpr(TraceImportName, "GoEnd", []ast.Expr{
		&id_ident,
	})
	assign := ast.AssignStmt{Lhs: []ast.Expr{&id_ident}, Rhs: []ast.Expr{callbefore}, Tok: token.DEFINE}

	switch fun := call.Fun.(type) {
	case *ast.FuncLit:
		fun.Body.List = append([]ast.Stmt{&assign, callafter}, fun.Body.List...)
		iCtx.SetMetadata(MonitorNeedInst, true)
	case *ast.Ident:
		newClosure := ast.FuncLit{
			Type: &ast.FuncType{
				Params:  nil, //&ast.FieldList{List: []*ast.Field{}},
				Results: nil, //&ast.FieldList{List: []*ast.Field{}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&assign,
					callafter,
					&ast.ExprStmt{call},
				},
			},
		}
		stmt.Call = &ast.CallExpr{Fun: &newClosure}
		iCtx.SetMetadata(MonitorNeedInst, true)
	}
}
