package main

import (
	"io"
	"log"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
)

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
	funcMap := template.FuncMap{
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
	t := template.Must(template.New("tmpl").Funcs(funcMap).Parse(tmpl))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateLabel(writer io.Writer) {
	t := template.Must(template.New("tmplLabel").Parse(tmplLabel))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateConstant(writer io.Writer) {
	t := template.Must(template.New("tmplConst").Parse(tmplConst))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

const tmplConst = `// THIS FILE IS A GENERATED CODE. DO NOT EDIT
package {{ .PackageName }}

import (
	"log"
	"strconv"
)

type BoolCriteria string

const (
	BoolCriteriaEmpty BoolCriteria = ""
	BoolCriteriaTrue  BoolCriteria = "true"
	BoolCriteriaFalse BoolCriteria = "false"
)

func (src BoolCriteria) Bool() bool {
	return src == BoolCriteriaTrue
}

type NumericCriteria string

const (
	NumericCriteriaEmpty NumericCriteria = ""
	NumericCriteriaBase  NumericCriteria = "0"
)

func (str NumericCriteria) Parse(i interface{}) NumericCriteria {
	switch x := i.(type) {
	case int:
		return NumericCriteria(strconv.Itoa(x))
	case int64:
		return NumericCriteria(strconv.FormatInt(x, 10))
	case float64:
		return NumericCriteria(strconv.FormatFloat(x, 'f', -1, 64))
	default:
		log.Println("invalid NumericCriteria value")
	}
	return NumericCriteriaEmpty
}

func (str NumericCriteria) Int() int {
	i32, err := strconv.Atoi(string(str))
	if err != nil {
		return -1
	}
	return i32
}

func (str NumericCriteria) Int64() int64 {
	i64, err := strconv.ParseInt(string(str), 10, 64)
	if err != nil {
		return -1
	}
	return i64
}

func (str NumericCriteria) Float64() float64 {
	f64, err := strconv.ParseFloat(string(str), 64)
	if err != nil {
		return -1
	}
	return f64
}
`

const tmplLabel = `// THIS FILE IS A GENERATED CODE. EDIT OK
package {{ .PackageName }}

const (
{{- range $fi := .FieldInfos }}
	{{- range $idx := $fi.Indexes }}
	// {{ $idx.Comment }}検索用ラベル
	{{ $idx.ConstName }} = "{{ $idx.Label }}"
	{{- end }}
{{- end }}
)
`

// nolint:lll
const tmpl = `// THIS FILE IS A GENERATED CODE. DO NOT EDIT
package {{ .PackageName }}

import (
	"context"
{{- range .ImportList }}
	"{{ .Name }}"
{{- end }}

	"cloud.google.com/go/datastore"
{{- if and (eq .SliceExist true) (eq .EnableIndexes false) }}
	"github.com/go-utils/dedupe"
{{- end }}
{{- if eq .KeyFieldType "*datastore.Key" }}
	"github.com/google/uuid"
{{- end }}
{{- if eq .EnableIndexes true }}
	"github.com/knightso/xian"
{{- end }}
	"golang.org/x/xerrors"
)

{{ $generate := "go:generate" }}
//{{ $generate }} mockgen -source {{ .GeneratedFileName }}.go -destination mock_{{ .GeneratedFileName }}/mock_{{ .GeneratedFileName }}.go

type {{ .RepositoryInterfaceName }} interface {
	// Single
	Get(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) (*{{ .StructName }}, error)
	Insert(ctx context.Context, subject *{{ .StructName }}) ({{ .KeyFieldType }}, error)
	Update(ctx context.Context, subject *{{ .StructName }}) error
	Delete(ctx context.Context, subject *{{ .StructName }}) error
	DeleteBy{{ .KeyFieldName }}(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) error
	// Multiple
	GetMulti(ctx context.Context, {{ .KeyValueName }}s []{{ .KeyFieldType }}) ([]*{{ .StructName }}, error)
	InsertMulti(ctx context.Context, subjects []*{{ .StructName }}) ([]{{ .KeyFieldType }}, error)
	UpdateMulti(ctx context.Context, subjects []*{{ .StructName }}) error
	DeleteMulti(ctx context.Context, subjects []*{{ .StructName }}) error
	DeleteMultiBy{{ .KeyFieldName }}s(ctx context.Context, {{ .KeyValueName }}s []{{ .KeyFieldType }}) error
	// List
	List(ctx context.Context, req *{{ .StructName }}ListReq, q *datastore.Query) ([]*{{ .StructName }}, error)
	// misc
	GetKindName() string
}

type {{ .RepositoryStructName }} struct {
	kind            string
	datastoreClient *datastore.Client
}

// New{{ .RepositoryInterfaceName }} constructor
func New{{ .RepositoryInterfaceName }}(datastoreClient *datastore.Client) {{ .RepositoryInterfaceName }} {
	return &{{ .RepositoryStructName }}{
		kind:            "{{ .StructName }}",
		datastoreClient: datastoreClient,
	}
}

// GetKindName KindName getter
func (repo *{{ .RepositoryStructName }}) GetKindName() string {
	return repo.kind
}

// getKeys Entityからkeyを取得する
func (repo *{{ .RepositoryStructName }}) getKeys(subjects ...*{{ .StructName }}) ([]*datastore.Key, error) {
	keys := make([]*datastore.Key, 0, len(subjects))
	for _, subject := range subjects {
		key	:= subject.{{ .KeyFieldName }}
{{- if eq .KeyFieldType "int64" }}
		if key == 0 {
			return nil, xerrors.New("ID must be set")
		}
		keys = append(keys, datastore.IDKey(repo.kind, key, nil))
{{- else if eq .KeyFieldType "string" }}
		if key == "" {
			return nil, xerrors.New("ID must be set")
		}
		keys = append(keys, datastore.NameKey(repo.kind, key, nil))
{{- else }}
		if key == nil {
			key = datastore.NameKey(repo.kind, uuid.New().String(), nil)
		}
		keys = append(keys, key)
{{- end }}
	}

	return keys, nil
}

{{- if eq .EnableIndexes true }}
// saveIndexes 拡張フィルタを保存する
func (repo *{{ .RepositoryStructName }}) saveIndexes(subjects ...*{{ .StructName }}) error {
	for _, subject := range subjects {
		idx := xian.NewIndexes({{ .StructName }}IndexesConfig)
{{- range $fi := .FieldInfos }}
{{- $PrefixIsSlice := HasPrefixSlice $fi.FieldType}}
{{- range $idx := $fi.Indexes }}
{{- if or (eq $fi.FieldType "bool") (eq $fi.FieldType "int" ) (eq $fi.FieldType "int64" ) (eq $fi.FieldType "float64" ) }}
		idx.{{ $idx.Method }}({{ $idx.ConstName }}, subject.{{ $fi.Field }})
{{- else if eq $fi.FieldType "string" }}
{{- if eq $idx.Method "AddPrefix" }}
		idx.{{ $idx.Method }}es({{ $idx.ConstName }}, subject.{{ $fi.Field }})
{{- else }}
		idx.{{ $idx.Method }}({{ $idx.ConstName }}, subject.{{ $fi.Field }})
{{- end }}
{{- else if eq $fi.FieldType "time.Time" }}
		idx.{{ $idx.Method }}({{ $idx.ConstName }}, subject.{{ $fi.Field }}.Unix())
{{- else if eq $PrefixIsSlice true }}
		idx.{{ $idx.Method }}({{ $idx.ConstName }}, subject.{{ $fi.Field }})
{{- end }}
{{- end }}
{{- end }}
		built, err := idx.Build()
		if err != nil {
			return err
		}
		subject.Indexes = built
	}

	return nil
}

// {{ .StructName }}IndexesConfig {{ .StructName }}用のIndexesConfigを設定する
var {{ .StructName }}IndexesConfig = &xian.Config{
	// IgnoreCase Case insensitive
	//   └──大文字小文字を区別しない
	IgnoreCase:         true,
	// SaveNoFiltersIndex 検索時にフィルタを設定しない場合、この拡張フィルタなし検索用インデックスのEquality Filterが自動で適用される
	//   ├── falseで、拡張フィルタのあり・なしの両パターンの検索がある場合、カスタムインデックスを両パターン分用意しておく必要がある
	//   └── trueにしておくことでカスタムインデックスを半分に節約することができる
	SaveNoFiltersIndex: true,
}
{{- end }}

// {{ .StructName }}ListReq List取得時に渡すリクエスト
// └─ bool/int(64)|float64 は stringの独自型で渡す(BoolCriteria | NumericCriteria)
type {{ .StructName }}ListReq struct {
{{- range .FieldInfos }}
{{- if eq .FieldType "bool" }}
	{{ .Field }} BoolCriteria
{{- else if or (eq .FieldType "int") (eq .FieldType "int64") (eq .FieldType "float64" ) }}
	{{ .Field }} NumericCriteria
{{- else }}
	{{ .Field }} {{ .FieldType }}
{{- end }}
{{- end }}
}

// List datastore.Queryを使用し条件抽出をする
//  └─ 第3引数はNOT/OR/IN/RANGEなど、より複雑な条件を適用したいときにつける
//      └─ 基本的にnilを渡せば良い
// BUG(54mch4n) 潜在的なバグがあるかもしれない
func (repo *{{ .RepositoryStructName }}) List(ctx context.Context, req *{{ .StructName }}ListReq, q *datastore.Query) ([]*{{ .StructName }}, error) {
	if q == nil {
		q = datastore.NewQuery(repo.kind)
	}
{{ $Enable := .EnableIndexes }}
{{- if eq $Enable true }}
	filters := xian.NewFilters({{ .StructName }}IndexesConfig)
{{- end }}
{{- range $fi := .FieldInfos }}
{{- $PrefixIsSlice := HasPrefixSlice $fi.FieldType}}
{{- if eq $fi.FieldType "bool" }}
	if req.{{ $fi.Field }} != "" {
{{- if eq $Enable true }}
{{- range $idx := $fi.Indexes }}
		filters.{{ $idx.Method }}({{ $idx.ConstName }}, req.{{ $fi.Field }})
{{- end }}
{{- else }}
		q = q.Filter("{{ $fi.DsTag }} =", req.{{ $fi.Field }}.Bool())
{{- end }}
	}
{{- else if eq $fi.FieldType "string" }}
	if req.{{ $fi.Field }} != "" {
{{- if eq $Enable true }}
{{- range $idx := $fi.Indexes }}
		filters.{{ $idx.Method }}({{ $idx.ConstName }}, req.{{ $fi.Field }})
{{- end }}
{{- else }}
		q = q.Filter("{{ $fi.DsTag }} =", req.{{ $fi.Field }})
{{- end }}
	}
{{- else if or (eq $fi.FieldType "int") (eq $fi.FieldType "int64") (eq $fi.FieldType "float64" ) }}
	if req.{{ $fi.Field }} != NumericCriteriaEmpty {
{{- if eq $Enable true }}
{{- range $idx := $fi.Indexes }}
		filters.{{ $idx.Method }}({{ $idx.ConstName }}, req.{{ $fi.Field }}.{{ Parse $fi.FieldType }}())
{{- end }}
{{- else }}
		q = q.Filter("{{ $fi.DsTag }} =", req.{{ $fi.Field }}.{{ Parse $fi.FieldType }}())
{{- end }}
	}
{{- else if eq $fi.FieldType "time.Time" }}
	if !req.{{ $fi.Field }}.IsZero() {
{{- if eq $Enable true }}
{{- range $idx := $fi.Indexes }}
		filters.{{ $idx.Method }}({{ $idx.ConstName }}, req.{{ $fi.Field }}.Unix())
{{- end }}
{{- else }}
		q = q.Filter("{{ $fi.DsTag }} =", req.{{ $fi.Field }})
{{- end }}
	}
{{- else if eq $PrefixIsSlice true }}
	if len(req.{{ $fi.Field }}) > 0 {
{{- if eq $Enable true }}
{{- range $idx := $fi.Indexes }}
		filters.{{ $idx.Method }}({{ $idx.ConstName }}, req.{{ $fi.Field }})
{{- end }}
{{- else }}
		dedupe.Do(&req.{{ $fi.Field }})
		for _, x := range req.{{ $fi.Field }} {
			q = q.Filter("{{ $fi.DsTag }} =", x)
		}
{{- end }}
	}
{{- end }}
{{- end }}
{{ if eq $Enable true }}
	built, err := filters.Build()
	if err != nil {
		return nil, err
	}

	for _, f := range built {
		q = q.Filter("{{ .FieldInfoForIndexes.DsTag }} =", f)
	}
{{- end }}
	subjects := make([]*{{ .StructName }}, 0)
	keys, err := repo.datastoreClient.GetAll(ctx, q, &subjects)
	if err != nil {
		return nil, err
	}

	for i, k := range keys {
		if k != nil {
{{- if eq .KeyFieldType "int64" }}
			subjects[i].ID = k.ID
{{- else if eq .KeyFieldType "string" }}
			subjects[i].ID = k.Name
{{- else }}
			subjects[i].ID = k
{{- end }}
		}
	}

	return subjects, nil
}

// Get 処理中の {{ .StructName }} の取得処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) Get(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) (*{{ .StructName }}, error) {
{{- if eq .KeyFieldType "int64" }}
	key := datastore.IDKey(repo.kind, {{ .KeyValueName }}, nil)
{{ else if eq .KeyFieldType "string" }}
	key := datastore.NameKey(repo.kind, {{ .KeyValueName }}, nil)
{{ else }}
	key := {{ .KeyValueName }}
{{ end }}
	subject := new({{ .StructName }})
	err := repo.datastoreClient.Get(ctx, key, subject)
	if err != nil {
		return nil, err
	}
{{ if eq .KeyFieldType "int64" }}
	subject.{{ .KeyFieldName }} = key.ID
{{ else if eq .KeyFieldType "string" }}
	subject.{{ .KeyFieldName }} = key.Name
{{ else }}
	subject.{{ .KeyFieldName }} = key
{{ end }}
	return subject, nil
}

// Insert 処理中の {{ .StructName }} の登録処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) Insert(ctx context.Context, subject *{{ .StructName }}) ({{ .KeyFieldType }}, error) {
{{- if eq .KeyFieldType "int64" }}
	zero := int64(0)
{{ else if eq .KeyFieldType "string" }}
	zero := ""
{{ else }}
	var zero {{ .KeyFieldType }}
{{ end }}
	keys, err := repo.getKeys(subject)
	if err != nil {
		return zero, xerrors.Errorf("error in getKeys method: %w", err)
	}
{{ if eq .EnableIndexes true }}
	if err := repo.saveIndexes(subject); err != nil {
		return zero, xerrors.Errorf("error in saveIndexes method: %w", err)
	}
{{ end }}
	key, err := repo.datastoreClient.Put(ctx, keys[0], subject)
	if err != nil {
		return zero, err
	}
{{ if eq .KeyFieldType "int64" }}
	return key.ID, nil
{{- else if eq .KeyFieldType "string" }}
	return key.Name, nil
{{- else }}
	return key, nil
{{- end }}
}

// Update 処理中の {{ .StructName }} の更新処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) Update(ctx context.Context, subject *{{ .StructName }}) error {
	if _, err := repo.Get(ctx, subject.{{ .KeyFieldName }}); err == datastore.ErrNoSuchEntity {
		return err
	}

	keys, err := repo.getKeys(subject)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}
{{ if eq .EnableIndexes true }}
	if err := repo.saveIndexes(subject); err != nil {
		return xerrors.Errorf("error in saveIndexes method: %w", err)
	}
{{ end }}
	if _, err := repo.datastoreClient.Put(ctx, keys[0], subject); err != nil {
		return err
	}

	return nil
}

// Delete 処理中の {{ .StructName }} の削除処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) Delete(ctx context.Context, subject *{{ .StructName }}) error {
	keys, err := repo.getKeys(subject)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	return repo.datastoreClient.Delete(ctx, keys[0])
}

// DeleteBy{{ .KeyFieldName }} 処理中の {{ .StructName }} の{{ .KeyFieldName }}から削除処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) DeleteBy{{ .KeyFieldName }}(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) error {
{{- if eq .KeyFieldType "int64" }}
	key := datastore.IDKey(repo.kind, {{ .KeyValueName }}, nil)
{{- else if eq .KeyFieldType "string" }}
	key := datastore.NameKey(repo.kind, {{ .KeyValueName }}, nil)
{{- else }}
	key := {{ .KeyValueName }}
{{ end }}
	return repo.datastoreClient.Delete(ctx, key)
}

// GetMulti 処理中の {{ .StructName }} の一括取得処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) GetMulti(ctx context.Context, {{ .KeyValueName }}s []{{ .KeyFieldType }}) (subjects []*{{ .StructName }}, err error) {
{{- if eq .KeyFieldType "int64" }}
	keys := make([]*datastore.Key, 0, len({{ .KeyValueName }}s))

	for i := range {{ .KeyValueName }}s {
		keys = append(keys, datastore.IDKey(repo.kind, {{ .KeyValueName }}s[i], nil))
	}
{{ else if eq .KeyFieldType "string" }}
	keys := make([]*datastore.Key, 0, len({{ .KeyValueName }}s))

	for i := range {{ .KeyValueName }}s {
		keys = append(keys, datastore.NameKey(repo.kind, {{ .KeyValueName }}s[i], nil))
	}
{{ else }}
	keys := {{ .KeyValueName }}s
{{ end }}
	subjects = make([]*{{ .StructName }}, len({{ .KeyValueName }}s))
	err = repo.datastoreClient.GetMulti(ctx, keys, subjects)

	for i := range subjects {
		if subjects[i] != nil {
{{- if eq .KeyFieldType "int64" }}
			subjects[i].{{ .KeyFieldName }} = keys[i].ID
{{- else if eq .KeyFieldType "string" }}
			subjects[i].{{ .KeyFieldName }} = keys[i].Name
{{- else }}
			subjects[i].{{ .KeyFieldName }} = keys[i]
{{- end }}
		}
	}

	return
}

// InsertMulti 処理中の {{ .StructName }} の一括挿入処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) InsertMulti(ctx context.Context, subjects []*{{ .StructName }}) ([]{{ .KeyFieldType }}, error) {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return nil, xerrors.Errorf("error in getKeys method: %w", err)
	}

	var cnt int
	if err := repo.datastoreClient.GetMulti(ctx, keys, make([]*{{ .StructName }}, len(subjects))); err != nil {
		if errs, ok := err.(datastore.MultiError); ok {
			for _, err := range errs {
				if err == datastore.ErrNoSuchEntity {
					cnt++
				}
			}
		}
	}

	if len(subjects) != cnt {
		return nil, xerrors.Errorf("already exist. (%d)", len(subjects)-cnt)
	}
{{ if eq .EnableIndexes true }}
	if err := repo.saveIndexes(subjects...); err != nil {
		return nil, xerrors.Errorf("error in saveIndexes method: %w", err)
	}
{{ end }}
	resKeys, err := repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return nil, err
	}

	vessels := make([]{{ .KeyFieldType }}, len(resKeys))
	for i := range resKeys {
		if keys[i] != nil {
{{- if eq .KeyFieldType "int64" }}
			vessels[i] = resKeys[i].ID
{{- else if eq .KeyFieldType "string" }}
			vessels[i] = resKeys[i].Name
{{- else }}
			vessels[i] = resKeys[i]
{{- end }}
		}
	}

	return vessels, err
}

// UpdateMulti 処理中の {{ .StructName }} の一括更新処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) UpdateMulti(ctx context.Context, subjects []*{{ .StructName }}) error {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	if err := repo.datastoreClient.GetMulti(ctx, keys, make([]*{{ .StructName }}, len(subjects))); err != nil {
		if _, ok := err.(datastore.MultiError); ok {
			return err
		}
	}
{{ if eq .EnableIndexes true }}
	if err := repo.saveIndexes(subjects...); err != nil {
		return xerrors.Errorf("error in saveIndexes method: %w", err)
	}
{{ end }}
	_, err = repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return err
	}

	return nil
}

// DeleteMulti 処理中の {{ .StructName }} の一括削除処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) DeleteMulti(ctx context.Context, subjects []*{{ .StructName }}) error {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	return repo.datastoreClient.DeleteMulti(ctx, keys)
}

// DeleteMultiBy{{ .KeyFieldName }}s 処理中の {{ .StructName }} の{{ .KeyFieldName }}群を元に一括削除処理一切の責任を持ち、これを行う
func (repo *{{ .RepositoryStructName }}) DeleteMultiBy{{ .KeyFieldName }}s(ctx context.Context, {{ .KeyValueName }}s []{{ .KeyFieldType }}) error {
{{- if eq .KeyFieldType "int64" }}
	keys := make([]*datastore.Key, 0, len({{ .KeyValueName }}s))

	for i := range {{ .KeyValueName }}s {
		keys = append(keys, datastore.IDKey(repo.kind, {{ .KeyValueName }}s[i], nil))
	}
{{ else if eq .KeyFieldType "string" }}
	keys := make([]*datastore.Key, 0, len({{ .KeyValueName }}s))

	for i := range {{ .KeyValueName }}s {
		keys = append(keys, datastore.NameKey(repo.kind, {{ .KeyValueName }}s[i], nil))
	}
{{ else }}
	keys := {{ .KeyValueName }}s
{{ end }}
	return repo.datastoreClient.DeleteMulti(ctx, keys)
}
`
