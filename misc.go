package main

import (
	"go/ast"
)

func getTypeName(typ ast.Expr) string {
	switch v := typ.(type) {
	case *ast.SelectorExpr:
		return getTypeName(v.X) + "." + v.Sel.Name

	case *ast.Ident:

		return v.Name

	case *ast.StarExpr:
		return "*" + getTypeName(v.X)

	default:
		return ""
	}
}
