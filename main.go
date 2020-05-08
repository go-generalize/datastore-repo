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
	"golang.org/x/xerrors"
)

var ImportName string

func main() {
	l := len(os.Args)
	if l < 3 {
		fmt.Println("You have to specify the struct name of target")
		os.Exit(1)
	}

	ImportName = os.Args[2]

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
				gen.LowerStructName = strcase.ToLowerCamel(name)

				return generate(gen, fs, structType)
			}
		}
	}

	return fmt.Errorf("no such struct: %s", structName)
}

func uppercaseExtraction(name string) (lower string) {
	for _, x := range name {
		if 65 <= x && x <= 90 {
			lower += string(x + 32)
		}
	}
	return
}

const queryLabel = "QueryLabel"

func generate(gen *generator, fs *token.FileSet, structType *ast.StructType) error {
	dupMap := make(map[string]int)
	filedLabel := gen.StructName + queryLabel
	for _, field := range structType.Fields.List {
		// structの各fieldを調査
		if len(field.Names) != 1 {
			return xerrors.New("`field.Names` must have only one element")
		}
		name := field.Names[0].Name

		if field.Tag == nil {
			continue
		}

		pos := fs.Position(field.Pos()).String()

		tags, err := structtag.Parse(strings.Trim(field.Tag.Value, "`"))

		if err != nil {
			log.Printf(
				"%s: tag for %s in struct %s in %s",
				pos, name, gen.StructName, gen.GeneratedFileName+".go",
			)

			continue
		}

		if name == "Indexes" {
			gen.EnableIndexes = true
			continue
		}

		_, err = tags.Get("datastore_key")
		if err != nil {
			f := func() string {
				u := uppercaseExtraction(name)
				if _, ok := dupMap[u]; !ok {
					dupMap[u] = 1
				} else {
					dupMap[u]++
					u = fmt.Sprintf("%s%d", u, dupMap[u])
				}
				return u
			}
			fieldInfo := &FieldInfo{
				Field:     name,
				FieldType: getTypeName(field.Type),
				Indexes:   make([]*IndexesInfo, 0),
			}
			dsTag, err := tags.Get("datastore")
			if err != nil {
				fieldInfo.DsTag = fieldInfo.Field
			} else {
				fieldInfo.DsTag = strings.Split(dsTag.Value(), ",")[0]
			}
			ft, err := tags.Get("filter")
			if err != nil || fieldInfo.FieldType != "string" {
				idx := &IndexesInfo{
					ConstName: filedLabel + name,
					Label:     f(),
					Method:    "Add",
				}
				idx.Comment = fmt.Sprintf("%s %s", idx.ConstName, name)
				if fieldInfo.FieldType != "string" {
					idx.Method += "Something"
				}
				fieldInfo.Indexes = append(fieldInfo.Indexes, idx)
			} else {
				filters := strings.Split(ft.Value(), ",")
				for _, fil := range filters {
					idx := &IndexesInfo{
						ConstName: filedLabel + name,
						Label:     f(),
						Method:    "Add",
					}
					switch fil {
					case "p", "prefix": // 前方一致 (AddPrefix)
						idx.Method += "Prefix"
						idx.ConstName += "Prefix"
						idx.Comment = fmt.Sprintf("%s %s前方一致", idx.ConstName, name)
					case "s", "suffix": /* TODO 後方一致
						idx.Method += "Suffix"
						idx.ConstName += "Suffix"
						idx.Comment = fmt.Sprintf("%s %s後方一致", idx.ConstName, name)*/
					case "e", "equal": // 完全一致 (Add) Default
						idx.Comment = fmt.Sprintf("%s %s", idx.ConstName, name)
					case "l", "like": // 部分一致
						idx.Method += "Biunigrams"
						idx.ConstName += "Like"
						idx.Comment = fmt.Sprintf("%s %s部分一致", idx.ConstName, name)
					default:
						continue
					}
					fieldInfo.Indexes = append(fieldInfo.Indexes, idx)
				}
			}

			gen.FieldInfos = append(gen.FieldInfos, fieldInfo)
			continue
		}

		dsTag, err := tags.Get("datastore")

		// datastore タグが存在しないか-になっていない
		if err != nil || strings.Split(dsTag.Value(), ",")[0] != "-" {
			return fmt.Errorf("%s: key field for datastore should have datastore:\"-\" tag", pos)
		}

		gen.KeyFieldName = name
		gen.KeyFieldType = getTypeName(field.Type)

		if gen.KeyFieldType != "int64" &&
			gen.KeyFieldType != "string" &&
			!strings.HasSuffix(gen.KeyFieldType, ".Key") {
			return fmt.Errorf("%s: supported key types are int64, string, *datastore.Key", pos)
		}

		gen.KeyValueName = strcase.ToLowerCamel(name)
	}

	{
		fp, err := os.Create(gen.GeneratedFileName + ".go")
		if err != nil {
			panic(err)
		}
		defer fp.Close()

		gen.generate(fp)
	}

	{
		if !exists("configs") {
			if err := os.Mkdir("configs", 0777); err != nil {
				return err
			}
		}
		path := "configs/" + strcase.ToLowerCamel(gen.StructName) + "_label.go"
		fp, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		gen.generateLabel(fp)
	}

	{
		fp, err := os.Create("configs/constant.go")
		if err != nil {
			panic(err)
		}
		defer fp.Close()
		gen.generateConstant(fp)
	}

	return nil
}

func exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}
