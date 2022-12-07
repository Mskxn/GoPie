package passes

import (
	"go/ast"
	"go/token"
	"strconv"
	"toolkit/pkg/inst"
	"toolkit/pkg/inst/IL"
	"toolkit/pkg/sched"
)

var im = inst.Import{
	Need: "SW_NEED_INST",
	Name: "sched",
	Path: "toolkit/pkg/sched",
}

var wrapper = IL.ILWrapperPass{
	FBefore: before,
	FAfter:  after,
	Dowrap:  dowrap,
}

var id_map map[token.Pos]uint64

func init() {
	wrapper.Need = im.Need
	wrapper.Name = im.Name
	wrapper.Path = im.Path

	id_map = make(map[token.Pos]uint64)
}

func RunSWPass(in, out string) {
	IL.RunFLWrapperPass(in, out, &wrapper)
}

func before(iCtx *inst.InstContext, node ast.Node) *ast.ExprStmt {
	switch concrete := node.(type) {
	case *ast.ExprStmt:
		if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				var matched bool = true
				var e uint64
				switch selectorExpr.Sel.Name {
				case "Lock", "RLOCK":
					e = sched.W_LOCK
				case "RUnlock", "Unlock":
					e = sched.W_UNLOCK
				default:
					matched = false
					e = 0
				}

				if matched {
					id := getNewOpID()
					id_map[node.Pos()] = id
					newCall := NewArgCallExpr("sched", "WE", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(id, 10),
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.FormatUint(e, 10),
					}})
					return newCall
				}
			}

		}
	}
	return nil
}

func after(iCtx *inst.InstContext, node ast.Node) *ast.ExprStmt {
	switch concrete := node.(type) {
	case *ast.ExprStmt:
		if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				var matched bool = true
				var e uint64
				switch selectorExpr.Sel.Name {
				case "Lock":
					e = sched.S_LOCK
				case "RLOCK":
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
					if id, ok := id_map[node.Pos()]; ok {
						newCall := NewArgCallExpr("sched", "SE", []ast.Expr{&ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(id, 10),
						}, &ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    strconv.FormatUint(e, 10),
						}})
						return newCall
					}
				}
			}
		}
	}
	return nil
}

func dowrap(iCtx *inst.InstContext, node ast.Node) bool {
	switch concrete := node.(type) {
	case *ast.ExprStmt:
		if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
			if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				if SelectorCallerHasTypes(iCtx, selectorExpr, true, "sync.Mutex", "*sync.Mutex", "sync.RWMutex", "*sync.RWMutex") {
					return true
				}
			}
		}
	}
	return false
}
