package passes

import (
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
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
						var e2 uint64
						switch selectorExpr.Sel.Name {
						case "Lock":
							e = sched.S_LOCK
							e2 = sched.W_LOCK
						case "RLock":
							e = sched.S_RLOCK
							e2 = sched.W_LOCK
						case "RUnlock":
							e = sched.S_RUNLOCK
							e2 = sched.W_UNLOCK
						case "Unlock":
							e = sched.S_UNLOCK
							e2 = sched.W_UNLOCK
						default:
							matched = false
							e = 0
							e2 = 0
						}

						if matched {
							id := iCtx.GetNewOpId()
							Add(concrete.Pos(), id)
							mu := selectorExpr.X
							p_mu := &ast.UnaryExpr{
								Op: token.AND,
								X:  mu,
							}
							before := GenInstCall("InstMutexBF", p_mu, id, e)
							c.InsertBefore(before)
							after := GenInstCall("InstMutexAF", p_mu, id, e2)
							c.InsertAfter(after)
							iCtx.SetMetadata(LockNeedInst, true)
						}
					}
				}
			}
		case *ast.DeferStmt:
			callExpr := concrete.Call
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				if SelectorCallerHasTypes(iCtx, selectorExpr, true, "sync.Mutex", "*sync.Mutex", "sync.RWMutex", "*sync.RWMutex") {
					var matched bool = true
					var e uint64
					var e2 uint64
					switch selectorExpr.Sel.Name {
					case "Lock":
						e = sched.S_LOCK
						e2 = sched.W_LOCK
					case "RLock":
						e = sched.S_RLOCK
						e2 = sched.W_LOCK
					case "RUnlock":
						e = sched.S_RUNLOCK
						e2 = sched.W_UNLOCK
					case "Unlock":
						e = sched.S_UNLOCK
						e2 = sched.W_UNLOCK
					default:
						matched = false
						e = 0
						e2 = 0
					}

					if matched {
						id := iCtx.GetNewOpId()
						Add(concrete.Pos(), id)

						mu := selectorExpr.X
						p_mu := &ast.UnaryExpr{
							Op: token.AND,
							X:  mu,
						}
						before := GenInstCall("InstMutexBF", p_mu, id, e)
						after := GenInstCall("InstMutexAF", p_mu, id, e2)

						body := &ast.BlockStmt{List: []ast.Stmt{
							before,
							&ast.ExprStmt{callExpr},
							after,
						}}

						deferStmt := &ast.DeferStmt{
							Call: &ast.CallExpr{
								Fun: &ast.FuncLit{
									Type: &ast.FuncType{Params: &ast.FieldList{List: nil}},
									Body: body,
								},
								Args: []ast.Expr{},
							},
						}
						c.Replace(deferStmt)
						iCtx.SetMetadata(LockNeedInst, true)
					}
				}
			}
			return false
		}
		return true
	}
}
