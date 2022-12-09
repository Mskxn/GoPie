package passes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strconv"
	"toolkit/pkg/sched"
)

// ChResPass, Channel Record Pass. This pass instrumented at
// following four channel related operations:
// send, recv, make, close

var (
	SelectInstNeed   = "SelectNeedInst"
	SelectImportName = "sched"
	SelectImportPath = "toolkit/pkg/sched"
)

type SelectPass struct {
}

func RunSelectPass(in, out string) error {
	p := SelectPass{}
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

func (p *SelectPass) Before(iCtx *InstContext) {
	iCtx.SetMetadata(ChannelNeedInst, false)
}

func (p *SelectPass) After(iCtx *InstContext) {
	need, _ := iCtx.GetMetadata(ChannelNeedInst)
	needinst := need.(bool)
	if needinst {
		AddImport(iCtx.FS, iCtx.AstFile, SelectImportName, SelectImportPath)
	}
}

func (p *SelectPass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *SelectPass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {

		// channel send operation
		case *ast.SelectStmt:
			cases := concrete.Body.List
			for _, x := range cases {
				if _, ok := x.(*ast.CommClause); !ok {
					return true
				}
				comm, _ := x.(*ast.CommClause)
				switch comm.Comm.(type) {
				case *ast.ExprStmt: // recv
					id := getNewOpID()
					id_map[concrete.Pos()] = id
					e := uint64(sched.S_RECV)
					newCall := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(id, 10),
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(e, 10),
					}})
					comm.Body = append([]ast.Stmt{newCall}, comm.Body...)
				case *ast.SendStmt: // send
					id := getNewOpID()
					id_map[concrete.Pos()] = id
					e := uint64(sched.S_SEND)
					newCall := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(id, 10),
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(e, 10),
					}})
					comm.Body = append([]ast.Stmt{newCall}, comm.Body...)
				}
			}
		}

		return true
	}
}
