package passes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"log"
	"strconv"
	"toolkit/pkg/trace"
)

// ChResPass, Channel Record Pass. This pass instrumented at
// following four channel related operations:
// send, recv, make, close

type ChMakePass struct {
}

func RunMakeChannelPass(in, out string) error {
	p := ChMakePass{}
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

func (p *ChMakePass) Before(iCtx *InstContext) {
	iCtx.SetMetadata(ChannelNeedInst, false)
}

func (p *ChMakePass) After(iCtx *InstContext) {
}

func (p *ChMakePass) GetPostApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *ChMakePass) GetPreApply(iCtx *InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.AssignStmt:
			rhs := concrete.Rhs
			if len(rhs) == 0 {
				return true
			}
			if call, ok := rhs[0].(*ast.CallExpr); ok {
				if funcIdent, ok := call.Fun.(*ast.Ident); ok {
					if funcIdent.Name == "make" {
						if len(call.Args) > 0 && len(call.Args) < 3 {
							if _, ok := call.Args[0].(*ast.ChanType); ok {
								if len(call.Args) == 1 {
									trace.ChanInfos.Add(concrete.Pos(), 0)
								} else {
									if b, ok := call.Args[1].(*ast.BasicLit); ok {
										if b.Kind == token.INT {
											size, err := strconv.Atoi(b.Value)
											if err == nil {
												trace.ChanInfos.Add(concrete.Pos(), size)
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}

		return true
	}
}
