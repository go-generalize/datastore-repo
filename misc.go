package main

import (
	"go/ast"
	"regexp"
)

func getTypeName(typ ast.Expr) string {
	switch v := typ.(type) {
	case *ast.SelectorExpr:
		return getTypeName(v.X) + "." + v.Sel.Name

	case *ast.Ident:
		return v.Name

	case *ast.StarExpr:
		return "*" + getTypeName(v.X)

	case *ast.ArrayType:
		return "[]" + getTypeName(v.Elt)

	default:
		return ""
	}
}

const (
	biunigrams  = "Biunigrams"
	prefix      = "Prefix"
	queryLabel  = "QueryLabel"
	typeString  = "string"
	typeInt     = "int"
	typeInt64   = "int64"
	typeFloat64 = "float64"
	typeBool    = "bool"
	typeTime    = "time.Time"
)

var (
	fieldLabel  string
	valueCheck  = regexp.MustCompile("^[0-9a-zA-Z_]+$")
	builtInType = []string{
		"uint8",
		"uint16",
		"uint32",
		"uint64",
		"int8",
		"int16",
		"int32",
		"float32",
		"complex64",
		"complex128",
		"uint",
		"uintptr",
		"byte",
		"rune",
		typeBool,
		typeString,
		typeInt,
		typeInt64,
		typeFloat64,
	}
	supportType = []string{
		typeBool,
		typeString,
		typeInt,
		typeInt64,
		typeFloat64,
		typeTime,
	}
)
