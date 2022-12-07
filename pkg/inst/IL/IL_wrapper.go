package IL

import (
	"go/ast"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"toolkit/pkg/inst"
)

func RunFLWrapperPass(in, out string, wp *ILWrapperPass) error {
	p := wp
	iCtx, err := inst.NewInstContext(in)
	if err != nil {
		log.Fatalf("Analysis source code failed %v", err)
	}
	p.Before(iCtx)
	iCtx.AstFile = astutil.Apply(iCtx.AstFile, p.GetPreApply(iCtx), p.GetPostApply(iCtx)).(*ast.File)
	p.After(iCtx)

	inst.DumpAstFile(iCtx.FS, iCtx.AstFile, out)
	return nil
}

type ILWrapperPass struct {
	FBefore func(iCtx *inst.InstContext, node ast.Node) *ast.ExprStmt
	FAfter  func(iCtx *inst.InstContext, node ast.Node) *ast.ExprStmt
	Dowrap  func(iCtx *inst.InstContext, node ast.Node) bool
	inst.Import
}

func (p *ILWrapperPass) Before(ctx *inst.InstContext) {
	if p.Need == "" {
		return
	}
	ctx.SetMetadata(p.Need, false)
}

func (p *ILWrapperPass) After(ctx *inst.InstContext) {
	if p.Need == "" {
		return
	}
	if v, ok := ctx.GetMetadata(p.Need); ok && v.(bool) {
		inst.AddImport(ctx.FS, ctx.AstFile, p.Name, p.Path)
	}
}

func (p *ILWrapperPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		n := c.Node()
		if !p.Dowrap(iCtx, n) {
			return true
		}
		switch st := n.(type) {
		case *ast.GoStmt:
			p.WrapGoStmt(st, iCtx)
			iCtx.SetMetadata(p.Need, true)
		default:
			if p.FBefore != nil {
				before := p.FBefore(iCtx, n)
				if before != nil {
					c.InsertBefore(before)
					iCtx.SetMetadata(p.Need, true)
				}
			}
			if p.FAfter != nil {
				after := p.FAfter(iCtx, n)
				if after != nil {
					c.InsertAfter(after)
					iCtx.SetMetadata(p.Need, true)
				}
			}
		}

		return true
	}
}

func (p *ILWrapperPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *ILWrapperPass) WrapGoStmt(stmt *ast.GoStmt, iCtx *inst.InstContext) {
	call := stmt.Call
	switch fun := call.Fun.(type) {
	case *ast.FuncLit:
		if p.FBefore != nil {
			before := p.FBefore(iCtx, stmt)
			if before != nil {
				fun.Body.List = append([]ast.Stmt{before}, fun.Body.List...)
				iCtx.SetMetadata(p.Need, true)
			}
		}
		if p.FAfter != nil {
			after := p.FAfter(iCtx, stmt)
			if after != nil {
				fun.Body.List = append(fun.Body.List, after)
				iCtx.SetMetadata(p.Need, true)
			}
		}
	case *ast.Ident:
		newClosure := ast.FuncLit{
			Type: &ast.FuncType{
				Params:  nil, //&ast.FieldList{List: []*ast.Field{}},
				Results: nil, //&ast.FieldList{List: []*ast.Field{}},
			},
			Body: &ast.BlockStmt{
				List: []ast.Stmt{
					&ast.ExprStmt{call},
				},
			},
		}
		if p.FBefore != nil {
			before := p.FBefore(iCtx, stmt)
			if before != nil {
				newClosure.Body.List = append([]ast.Stmt{before}, newClosure.Body.List...)
				iCtx.SetMetadata(p.Need, true)
			}
		}
		if p.FAfter != nil {
			after := p.FAfter(iCtx, stmt)
			if after != nil {
				newClosure.Body.List = append(newClosure.Body.List, after)
				iCtx.SetMetadata(p.Need, true)
			}
		}
		stmt.Call = &ast.CallExpr{Fun: &newClosure}
	}
}
