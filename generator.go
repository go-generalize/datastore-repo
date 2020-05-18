package main

import (
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	_ "github.com/go-generalize/datastore-repo/statik"
	"github.com/iancoleman/strcase"
	"github.com/rakyll/statik/fs"
)

var statikFS http.FileSystem

func init() {
	var err error
	statikFS, err = fs.New()
	if err != nil {
		log.Fatal(err)
	}
}

type IndexesInfo struct {
	Comment   string
	ConstName string
	Label     string
	Method    string
}

type FieldInfo struct {
	DsTag     string
	Field     string
	FieldType string
	Indexes   []*IndexesInfo
}

type ImportInfo struct {
	Name string
}

type generator struct {
	PackageName       string
	ImportName        string
	ImportList        []ImportInfo
	GeneratedFileName string
	FileName          string
	StructName        string

	RepositoryStructName    string
	RepositoryInterfaceName string

	KeyFieldName string
	KeyFieldType string
	KeyValueName string // lower camel case

	FieldInfos []*FieldInfo

	EnableIndexes       bool
	FieldInfoForIndexes *FieldInfo
	BoolCriteriaCnt     int
	SliceExist          bool
}

func (g *generator) setting() {
	g.RepositoryInterfaceName = g.StructName + "Repository"
	g.RepositoryStructName = strcase.ToLowerCamel(g.RepositoryInterfaceName)
	g.buildConditions()
}

func (g *generator) buildConditions() {
	for _, field := range g.FieldInfos {
		switch field.FieldType {
		case "time.Time":
			g.ImportList = append(g.ImportList, ImportInfo{"time"})
		}
	}
}

func (g *generator) generate(writer io.Writer) {
	g.setting()
	funcMap := setFuncMap()
	contents := getFileContents("gen")

	t := template.Must(template.New("Template").Funcs(funcMap).Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateLabel(writer io.Writer) {
	contents := getFileContents("label")

	t := template.Must(template.New("TemplateLabel").Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateConstant(writer io.Writer) {
	contents := getFileContents("constant")

	t := template.Must(template.New("TemplateConstant").Parse(contents))

	if err := t.Execute(writer, g); err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func setFuncMap() template.FuncMap {
	return template.FuncMap{
		"Parse": func(fieldType string) string {
			if strings.HasPrefix(fieldType, "[]") {
				fieldType = fieldType[2:]
			}
			fn := "Int"
			switch fieldType {
			case typeInt:
			case typeInt64:
				fn = "Int64"
			case typeFloat64:
				fn = "Float64"
			case typeString:
				fn = "String"
			case typeBool:
				fn = "Bool"
			default:
				panic("invalid types")
			}
			return fn
		},
		"HasPrefixSlice": func(types string) bool {
			return strings.HasPrefix(types, "[]")
		},
	}
}
