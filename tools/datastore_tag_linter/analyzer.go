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
const Doc = `Datastore tags linter for repo_generator`

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

func runOnFile(pass *analysis.Pass, file *ast.File) {
	inspect := inspector.New([]*ast.File{file})

	dspkg := getDatastoreAlias(inspect)

	nodeFilter := []ast.Node{
		(*ast.StructType)(nil),
	}

	inspect.Preorder(nodeFilter, func(n ast.Node) {
		strc := n.(*ast.StructType)

		for _, field := range strc.Fields.List {
			stared, ok := field.Type.(*ast.StarExpr)
			if !ok {
				continue
			}

			ident, ok := stared.X.(*ast.SelectorExpr)
			if !ok {
				continue
			}

			x, ok := ident.X.(*ast.Ident)

			if !ok {
				continue
			}

			if x.Name != dspkg || ident.Sel.Name != "Key" {
				continue
			}

			if !hasDatastoreTag(field.Tag) {
				pass.Reportf(field.Pos(), `*datastore.Key should have datastore:"-" tag`)
			}
		}
	})
}

func run(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		runOnFile(pass, file)
	}

	return nil, nil
}
