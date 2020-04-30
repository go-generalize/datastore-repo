package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/structtag"
	"github.com/iancoleman/strcase"
)

func main() {
	l := len(os.Args)
	if l < 2 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	if err := run(os.Args[1]); err != nil {
		log.Fatal(err.Error())
	}
}

func run(structName string) error {
	fs := token.NewFileSet()
	pkgs, err := parser.ParseDir(fs, ".", nil, parser.AllErrors)

	if err != nil {
		panic(err)
	}

	for name, v := range pkgs {
		if strings.HasSuffix(name, "_test") {
			continue
		}

		return traverse(v, fs, structName)
	}

	return nil
}

func traverse(pkg *ast.Package, fs *token.FileSet, structName string) error {
	gen := &generator{PackageName: pkg.Name}
	for name, file := range pkg.Files {
		gen.FileName = strings.TrimSuffix(filepath.Base(name), ".go")
		gen.GeneratedFileName = gen.FileName + "_gen"

		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok != token.TYPE {
				continue
			}

			for _, spec := range genDecl.Specs {
				// 型定義
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				name := typeSpec.Name.Name

				if name != structName {
					continue
				}

				// structの定義
				structType, ok := typeSpec.Type.(*ast.StructType)
				if !ok {
					continue
				}
				gen.StructName = name

				return generate(gen, fs, structType)
			}
		}
	}

	return fmt.Errorf("no such struct: %s", structName)
}

func generate(gen *generator, fs *token.FileSet, structType *ast.StructType) error {
	for _, field := range structType.Fields.List {
		// structの各fieldを調査

		if field.Tag == nil {
			continue
		}

		pos := fs.Position(field.Pos()).String()

		tags, err := structtag.Parse(strings.Trim(field.Tag.Value, "`"))

		if err != nil {
			log.Printf(
				"%s: tag for %s in struct %s in %s",
				pos, field.Names[0].Name, gen.StructName, gen.GeneratedFileName+".go",
			)

			continue
		}

		_, err = tags.Get("datastore_key")

		if err != nil {
			continue
		}

		dsTag, err := tags.Get("datastore")

		// datastore タグが存在しないか-になっていない
		if err != nil || strings.Split(dsTag.Value(), ",")[0] != "-" {
			return fmt.Errorf("%s: key field for datastore should have datastore:\"-\" tag", pos)
		}

		if len(field.Names) != 1 {
			return fmt.Errorf("%s: datastore_key tag can be set to only one field", pos)
		}

		gen.KeyFieldName = field.Names[0].Name
		gen.KeyFieldType = getTypeName(field.Type)

		if gen.KeyFieldType != "int64" &&
			gen.KeyFieldType != "string" &&
			!strings.HasSuffix(gen.KeyFieldType, ".Key") {
			return fmt.Errorf("%s: supported key types are int64, string, *datastore.Key", pos)
		}

		gen.KeyValueName = strcase.ToLowerCamel(field.Names[0].Name)
	}

	fp, err := os.Create(gen.GeneratedFileName + ".go")

	if err != nil {
		panic(err)
	}

	gen.generate(
		fp,
	)

	fp.Close()

	return nil
}
