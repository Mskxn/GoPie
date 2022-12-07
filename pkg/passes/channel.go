package passes

import (
	"go/ast"
	"go/token"
	"log"
	"strconv"

	"golang.org/x/tools/go/ast/astutil"
)

// ChResPass, Channel Record Pass. This pass instrumented at
// following four channel related operations:
// send, recv, make, close

var (
	ChannelNeedInst = "ChannelNeedInst"
	YieldImportName = "yield"
	YieldImportPath = "toolkit/pkg/yield"
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
		AddImport(iCtx.FS, iCtx.AstFile, YieldImportName, YieldImportPath)
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
			intID := int(getNewOpID())
			newCall := NewArgCallExpr("yield", "Yield", []ast.Expr{&ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.Itoa(intID),
			}})
			c.InsertBefore(newCall)
			iCtx.SetMetadata(ChannelNeedInst, true)

		// channel recv operation
		case *ast.ExprStmt:
			if unaryExpr, ok := concrete.X.(*ast.UnaryExpr); ok {
				if unaryExpr.Op == token.ARROW { // This is a receive operation
					intID := int(getNewOpID())
					newCall := NewArgCallExpr("yield", "Yield", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.Itoa(intID),
					}})
					c.InsertBefore(newCall)
					iCtx.SetMetadata(ChannelNeedInst, true)
				}
			} else if callExpr, ok := concrete.X.(*ast.CallExpr); ok { // like `close(ch)` or `mu.Lock()`
				if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok { // like `close(ch)`
					// channel close operation
					if funcIdent.Name == "close" {
						intID := int(getNewOpID())
						newCall := NewArgCallExpr("yield", "Yield", []ast.Expr{&ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.Itoa(intID),
						}})
						c.InsertBefore(newCall)
						iCtx.SetMetadata(ChannelNeedInst, true)
					}
				}
			}
		}

		return true
	}
}
