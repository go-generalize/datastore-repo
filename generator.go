package main

import (
	"io"
	"log"
	"text/template"
)

type IndexesInfo struct {
	Comment    string
	Field      string
	Label      string
	SearchItem string // TODO `前方一致` や `部分一致` など (5/7 unused)
}

type generator struct {
	PackageName       string
	GeneratedFileName string
	FileName          string
	StructName        string

	GoGenerate              string
	RepositoryStructName    string
	RepositoryInterfaceName string

	KeyFieldName string
	KeyFieldType string

	KeyValueName string // lower camel case

	ConstMapForIndexes []IndexesInfo
}

func (g *generator) setting() {
	g.GoGenerate = "go:generate"
	g.RepositoryInterfaceName = g.StructName + "Repository"
	g.setRepositoryStructName()
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

func (g *generator) setRepositoryStructName() {
	name := g.RepositoryInterfaceName
	prefix := name[:1]
	r := []rune(prefix)[0]
	if 65 <= r && r <= 90 {
		prefix = string(r + 32)
	}
	g.RepositoryStructName = prefix + name[1:]
}

func (g *generator) generate(writer io.Writer) {
	g.setting()
	funcMap := template.FuncMap{
		"Parse": func(field, fieldType string) string {
			fn := ".Int()"
			switch fieldType {
			case "int":
			case "int64":
				fn = ".Int64()"
			default:
				panic("invalid types")
			}
			return field + fn
		},
	}
	t := template.Must(template.New("tmpl").Funcs(funcMap).Parse(tmpl))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

func (g *generator) generateConstant(writer io.Writer) {
	g.setting()
	t := template.Must(template.New("tmpl").Parse(tmplConst))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

const tmplConst = `// THIS FILE IS A GENERATED CODE. EDIT OK
package configs

const (
{{- range .ConstMapForIndexes }}
	// {{ .Comment }}検索用ラベル
	{{ .Field }} = "{{ .Label }}"
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
	"golang.org/x/xerrors"
)

//{{ .GoGenerate }} mockgen -source {{ .GeneratedFileName }}.go -destination mock_{{ .GeneratedFileName }}/mock_{{ .GeneratedFileName }}.go

type {{ .RepositoryInterfaceName }} interface {
	// Single
	Get(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) (*{{ .StructName }}, error)
	Insert(ctx context.Context, subject *{{ .StructName }}) ({{ .KeyFieldType }}, error)
	Update(ctx context.Context, subject *{{ .StructName }}) error
	Delete(ctx context.Context, subject *{{ .StructName }}) error
	DeleteBy{{ .KeyFieldName }}(ctx context.Context, {{ .KeyValueName }} {{ .KeyFieldType }}) error
	// Multiple
	GetMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) ([]*{{.StructName}}, error)
	InsertMulti(ctx context.Context, subjects []*{{.StructName}}) ([]{{.KeyFieldType}}, error)
	UpdateMulti(ctx context.Context, subjects []*{{.StructName}}) error
	DeleteMulti(ctx context.Context, subjects []*{{.StructName}}) error
	DeleteMultiBy{{.KeyFieldName}}s(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) error
}

type {{ .RepositoryStructName }} struct {
	kind            string
	datastoreClient *datastore.Client
}

func New{{ .RepositoryInterfaceName }}(datastoreClient *datastore.Client) {{ .RepositoryInterfaceName }} {
	return &{{ .RepositoryStructName }}{
		kind:            "{{ .StructName }}",
		datastoreClient: datastoreClient,
	}
}

func (repo *{{.RepositoryStructName}}) getKeys(subjects ...*{{.StructName}}) ([]*datastore.Key, error) {
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
			key = datastore.IncompleteKey(repo.kind, nil)
		}
		keys = append(keys, key)
{{- end }}
	}

	return keys, nil
}

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
{{if eq .KeyFieldType "int64" }}
	subject.{{ .KeyFieldName }} = key.ID
{{ else if eq .KeyFieldType "string" }}
	subject.{{ .KeyFieldName }} = key.Name
{{ else }}
	subject.{{ .KeyFieldName }} = key
{{ end }}
	return subject, nil
}

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

	key, err := repo.datastoreClient.Put(ctx, keys[0], subject)
	if err != nil {
		return zero, err
	}
{{if eq .KeyFieldType "int64" }}
	return key.ID, nil
{{ else if eq .KeyFieldType "string" }}
	return key.Name, nil
{{ else }}
	return key, nil
{{- end }}
}

func (repo *{{ .RepositoryStructName }}) Update(ctx context.Context, subject *{{ .StructName }}) error {
	if _, err := repo.Get(ctx, subject.{{ .KeyFieldName }}); err == datastore.ErrNoSuchEntity {
		return err
	}

	keys, err := repo.getKeys(subject)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	if _, err := repo.datastoreClient.Put(ctx, keys[0], subject); err != nil {
		return err
	}

	return nil
}

func (repo *{{ .RepositoryStructName }}) Delete(ctx context.Context, subject *{{ .StructName }}) error {
	keys, err := repo.getKeys(subject)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	return repo.datastoreClient.Delete(ctx, keys[0])
}

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

func (repo *{{ .RepositoryStructName }}) GetMulti(ctx context.Context, {{ .KeyValueName }}s []{{ .KeyFieldType }}) ([]*{{ .StructName }}, error) {
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
	vessels := make([]*{{ .StructName }}, len({{ .KeyValueName }}s))
	err := repo.datastoreClient.GetMulti(ctx, keys, vessels)

	for i := range vessels {
		if vessels[i] != nil {
{{- if eq .KeyFieldType "int64" }}
			vessels[i].{{ .KeyFieldName }} = keys[i].ID
{{- else if eq .KeyFieldType "string" }}
			vessels[i].{{ .KeyFieldName }} = keys[i].Name
{{- else }}
			vessels[i].{{ .KeyFieldName }} = keys[i]
{{- end }}
		}
	}

	return vessels, err
}

func (repo *{{ .RepositoryStructName }}) InsertMulti(ctx context.Context, subjects []*{{ .StructName }}) ([]{{ .KeyFieldType }}, error) {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return nil, xerrors.Errorf("error in getKeys method: %w", err)
	}

	var cnt int
	if err := repo.datastoreClient.GetMulti(ctx, keys, make([]*{{ .StructName }}, 0, len(subjects))); err != nil {
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

	resKeys, err := repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return nil, err
	}

	vessels := make([]{{ .KeyFieldType }}, 0, len(resKeys))
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

func (repo *{{ .RepositoryStructName }}) UpdateMulti(ctx context.Context, subjects []*{{ .StructName }}) error {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	if err := repo.datastoreClient.GetMulti(ctx, keys, make([]*{{ .StructName }}, 0, len(subjects))); err != nil {
		if _, ok := err.(datastore.MultiError); ok {
			return err
		}
	}

	_, err = repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return err
	}

	return nil
}

func (repo *{{ .RepositoryStructName }}) DeleteMulti(ctx context.Context, subjects []*{{ .StructName }}) error {
	keys, err := repo.getKeys(subjects...)
	if err != nil {
		return xerrors.Errorf("error in getKeys method: %w", err)
	}

	return repo.datastoreClient.DeleteMulti(ctx, keys)
}

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
