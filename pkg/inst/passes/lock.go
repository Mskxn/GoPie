package passes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"strconv"
	"toolkit/pkg/inst"
	"toolkit/pkg/sched"
)

var (
	LockNeedInst   = "LockNeedInst"
	LockImportName = "sched"
	LockImportPath = "toolkit/pkg/sched"
)

type LockPass struct {
}

func (p *LockPass) Before(iCtx *inst.InstContext) {
	iCtx.SetMetadata(LockNeedInst, false)
}

func (p *LockPass) After(iCtx *inst.InstContext) {
	need, _ := iCtx.GetMetadata(LockNeedInst)
	needinst := need.(bool)
	if needinst {
		inst.AddImport(iCtx.FS, iCtx.AstFile, LockImportName, LockImportPath)
	}
}

func (p *LockPass) GetPostApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return nil
}

func (p *LockPass) GetPreApply(iCtx *inst.InstContext) func(*astutil.Cursor) bool {
	return func(c *astutil.Cursor) bool {
		defer func() {
			if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
				// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic
			}
		}()

		switch concrete := c.Node().(type) {
		case *ast.ExprStmt:
			if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
				if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
					if SelectorCallerHasTypes(iCtx, selectorExpr, true, "sync.Mutex", "*sync.Mutex", "sync.RWMutex", "*sync.RWMutex") {
						var matched bool = true
						var e uint64
						switch selectorExpr.Sel.Name {
						case "Lock":
							e = sched.S_LOCK
						case "RLock":
							e = sched.S_RLOCK
						case "RUnlock":
							e = sched.S_RUNLOCK
						case "Unlock":
							e = sched.S_UNLOCK
						default:
							matched = false
							e = 0
						}

						if matched {
							id := iCtx.GetNewOpId()
							Add(concrete.Pos(), id)
							newCall := NewArgCallExpr("sched", "WE", []ast.Expr{&ast.BasicLit{
								ValuePos: 0,
								Kind:     token.INT,
								Value:    strconv.FormatUint(id, 10),
							}, &ast.BasicLit{
								ValuePos: 0,
								Kind:     token.INT,
								Value:    strconv.FormatUint(e, 10),
							}})
							c.InsertBefore(newCall)
							newCall = NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
								ValuePos: 0,
								Kind:     token.INT,
								Value:    strconv.FormatUint(id, 10),
							}, &ast.BasicLit{
								ValuePos: 0,
								Kind:     token.INT,
								Value:    strconv.FormatUint(e, 10),
							}})
							c.InsertAfter(newCall)
							iCtx.SetMetadata(LockNeedInst, true)
						}
					}
				}
			}
		}
		return true
	}
}
