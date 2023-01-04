package leak_checker

import (
	"go/ast"
	"go/token"
	"toolkit/pkg/inst/passes"
	"toolkit/pkg/insttool"
)

func RunPass(in, out string) {
	need := "NEED-goleak"
	path := "toolkit/pkg/goleak"
	name := "goleak"

	imp := insttool.Import{name, path, need}

	call := "VerfiyNone"
	after := insttool.NewDeferExpr(name, call, []ast.Expr{&ast.BasicLit{
		ValuePos: 0,
		Kind:     token.IDENT,
		Value:    "t",
	}})

	before := (*ast.Stmt)(nil)

	dowrap := func(n ast.Node) bool {
		return insttool.IsTestFunc(n)
	}

	wp := passes.NewWrapperPass(
		before,
		after,
		dowrap,
		imp,
	)

	passes.RunFLWrapperPass(in, out, wp)
}
