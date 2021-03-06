package main

import (
	"go/ast"
	"io/ioutil"
	"log"
	"regexp"
)

const (
	biunigrams   = "Biunigrams"
	prefix       = "Prefix"
	queryLabel   = "QueryLabel"
	typeString   = "string"
	typeInt      = "int"
	typeInt64    = "int64"
	typeFloat64  = "float64"
	typeBool     = "bool"
	typeTime     = "time.Time"
	datastoreKey = "*datastore.Key"
)

var (
	fieldLabel  string
	valueCheck  = regexp.MustCompile("^[0-9a-zA-Z_]+$")
	supportType = []string{
		typeBool,
		typeString,
		typeInt,
		typeInt64,
		typeFloat64,
		typeTime,
		datastoreKey,
	}
)

func getFileContents(name string) string {
	fp, err := statikFS.Open("/" + name + ".go.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	defer fp.Close()
	contents, err := ioutil.ReadAll(fp)
	if err != nil {
		log.Fatal(err)
	}
	return string(contents)
}

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
