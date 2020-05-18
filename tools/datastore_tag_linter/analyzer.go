package dstags

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/fatih/structtag"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "dstags",
	Doc:      Doc,
	Run:      run,
	Requires: []*analysis.Analyzer{},
}

// Doc - analyzerの説明
const Doc = `Datastore tags linter for datastore-repo`

func getDatastoreAlias(inspect *inspector.Inspector) string {
	nodeFilter := []ast.Node{
		(*ast.GenDecl)(nil),
	}
	alias := ""
	inspect.Preorder(nodeFilter, func(n ast.Node) {
		decl := n.(*ast.GenDecl)

		if decl.Tok != token.IMPORT {
			return
		}

		for _, s := range decl.Specs {
			spec := s.(*ast.ImportSpec)

			if spec.Path.Value == `"cloud.google.com/go/datastore"` {
				if spec.Name != nil && spec.Name.Name != "" {
					alias = spec.Name.Name
				} else {
					alias = "datastore"
				}
			}
		}
	})

	return alias
}

func hasDatastoreKeyTag(lit *ast.BasicLit) bool {
	if lit == nil {
		return false
	}

	tags, err := structtag.Parse(strings.Trim(lit.Value, "`"))

	if err != nil {
		return false
	}

	_, err = tags.Get("datastore_key")

	if err != nil {
		return false
	}

	return true
}

func hasDatastoreTag(lit *ast.BasicLit) bool {
	if lit == nil {
		return false
	}

	tags, err := structtag.Parse(strings.Trim(lit.Value, "`"))

	if err != nil {
		return false
	}

	tag, err := tags.Get("datastore")

	if err != nil {
		return false
	}

	if tag.String() != `datastore:"-"` {
		return false
	}

	return true
}

func expectDatastoreKey(field *ast.Field, dspkg string) bool {
	stared, ok := field.Type.(*ast.StarExpr)
	if !ok {
		return false
	}

	ident, ok := stared.X.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	x, ok := ident.X.(*ast.Ident)

	if !ok {
		return false
	}

	if x.Name != dspkg || ident.Sel.Name != "Key" {
		return false
	}

	return true
}

func expectString(field *ast.Field) bool {
	ident, ok := field.Type.(*ast.Ident)

	if !ok {
		return false
	}

	if ident.Name != "string" {
		return false
	}

	return true
}

func expectInt64(field *ast.Field) bool {
	ident, ok := field.Type.(*ast.Ident)

	if !ok {
		return false
	}

	if ident.Name != "int64" {
		return false
	}

	return true
}

func runOnFile(pass *analysis.Pass, file *ast.File) {
	inspect := inspector.New([]*ast.File{file})

	dspkg := getDatastoreAlias(inspect)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		strc := n.(*ast.StructType)

		datastoreKeyTag := false

		for _, field := range strc.Fields.List {
			if !hasDatastoreKeyTag(field.Tag) {
				continue
			}

			datastoreKeyTag = true

			if !hasDatastoreTag(field.Tag) {
				pass.Reportf(field.Pos(), `datastore key should have datastore:"" tag`)
			}

			if !(expectDatastoreKey(field, dspkg) ||
				expectString(field) ||
				expectInt64(field)) {
				pass.Reportf(field.Pos(), `available types for datastore key is *datastore.Key, string, int64`)
			}
		}

		if !datastoreKeyTag {
			pass.Reportf(n.Pos(), `struct for datastore should have a field for the key`)
		}
	})
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		runOnFile(pass, file)
	}

	return nil, nil
}
