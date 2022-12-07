package leak_checker

import (
	"go/ast"
	"go/token"
	"toolkit/pkg/inst"
	"toolkit/pkg/inst/FL"
)

func RunPass(in, out string) {
	need := "NEED-goleak"
	path := "toolkit/pkg/goleak"
	name := "goleak"

	imp := inst.Import{name, path, need}

	call := "VerfiyNone"
	after := inst.NewDeferExpr(name, call, []ast.Expr{&ast.BasicLit{
		ValuePos: 0,
		Kind:     token.IDENT,
		Value:    "t",
	}})

	before := (*ast.Stmt)(nil)

	dowrap := func(n ast.Node) bool {
		return inst.IsTestFunc(n)
	}

	wp := FL.NewWrapperPass(
		before,
		after,
		dowrap,
		imp,
	)

	FL.RunFLWrapperPass(in, out, wp)
}
