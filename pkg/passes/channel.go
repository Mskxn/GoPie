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
	ChannelNeedInst   = "ChannelNeedInst"
	ChannelImportName = "sched"
	ChannelImportPath = "toolkit/pkg/sched"
)

type ChRecPass struct {
}

func RunChannelPass(in, out string) error {
	p := ChRecPass{}
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

func (p *ChRecPass) Before(iCtx *InstContext) {
	iCtx.SetMetadata(ChannelNeedInst, false)
}

func (p *ChRecPass) After(iCtx *InstContext) {
	need, _ := iCtx.GetMetadata(ChannelNeedInst)
	needinst := need.(bool)
	if needinst {
		AddImport(iCtx.FS, iCtx.AstFile, ChannelImportName, ChannelImportPath)
	}
}

func (p *ChRecPass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *ChRecPass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {

		// channel send operation
		case *ast.SendStmt:
			id := getNewOpID()
			id_map[concrete.Pos()] = id
			e := uint64(sched.W_RECV)

			before := NewArgCallExpr("sched", "WE", []ast.Expr{&ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.FormatUint(id, 10),
			}, &ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.FormatUint(e, 10),
			}})
			c.InsertBefore(before)

			e = sched.S_SEND
			after := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.FormatUint(id, 10),
			}, &ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.FormatUint(e, 10),
			}})
			c.InsertAfter(after)

			iCtx.SetMetadata(ChannelNeedInst, true)

		// channel recv operation
		case *ast.ExprStmt:
			if unaryExpr, ok := concrete.X.(*ast.UnaryExpr); ok {
				if unaryExpr.Op == token.ARROW { // This is a receive operation
					id := getNewOpID()
					id_map[concrete.Pos()] = id
					e := uint64(sched.W_SEND)

					before := NewArgCallExpr("sched", "WE", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(id, 10),
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(e, 10),
					}})
					c.InsertBefore(before)

					e = sched.S_RECV
					after := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(id, 10),
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(e, 10),
					}})
					c.InsertAfter(after)

					iCtx.SetMetadata(ChannelNeedInst, true)
				}
			} else if callExpr, ok := concrete.X.(*ast.CallExpr); ok { // like `close(ch)` or `mu.Lock()`
				if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok { // like `close(ch)`
					// channel close operation
					if funcIdent.Name == "close" {
						id := getNewOpID()
						id_map[concrete.Pos()] = id
						e := uint64(sched.W_CLOSE)

						before := NewArgCallExpr("sched", "WE", []ast.Expr{&ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(id, 10),
						}, &ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(e, 10),
						}})
						c.InsertBefore(before)

						e = sched.S_CLOSE
						after := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(id, 10),
						}, &ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(e, 10),
						}})
						c.InsertAfter(after)

						iCtx.SetMetadata(ChannelNeedInst, true)
					}
				}
			}
		}

		return true
	}
}
