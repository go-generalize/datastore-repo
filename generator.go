package main

import (
	"io"
	"log"
	"text/template"
)

type generator struct {
	PackageName       string
	GeneratedFileName string
	FileName          string
	StructName        string

	RepositoryStructName    string
	RepositoryInterfaceName string

	KeyFieldName string
	KeyFieldType string

	KeyValueName string // lower camel case
}

func (g *generator) generate(writer io.Writer) {
	g.RepositoryInterfaceName = g.StructName + "Repository"
	g.setRepositoryStructName()
	t := template.Must(template.New("tmpl").Parse(tmpl))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
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

const tmpl = `// THIS FILE IS A GENERATED CODE. DO NOT EDIT
package {{.PackageName}}

import (
	"context"

	"cloud.google.com/go/datastore"
)

//go:generate mockgen -source {{.GeneratedFileName}}.go -destination mock_{{.GeneratedFileName}}/mock_{{.GeneratedFileName}}.go

type {{.RepositoryInterfaceName}} interface {
	Get(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) (*{{.StructName}}, error)
	GetMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) ([]*{{.StructName}}, error)
	Insert(ctx context.Context, subject *{{.StructName}}) ({{.KeyFieldType}}, error)
	InsertMulti(ctx context.Context, subjects []*{{.StructName}}) ([]{{.KeyFieldType}}, error)
	Update(ctx context.Context, subject *{{.StructName}}) error
	UpdateMulti(ctx context.Context, subjects []*{{.StructName}}) error
	Delete(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) error
	DeleteMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) error
}

type {{.RepositoryStructName}} struct {
	kind            string
	datastoreClient *datastore.Client
}

func New{{.RepositoryInterfaceName}}(datastoreClient *datastore.Client) {{.RepositoryInterfaceName}} {
	return &{{.RepositoryStructName}}{
		kind:            "{{.StructName}}",
		datastoreClient: datastoreClient,
	}
}

func (repo *{{.RepositoryStructName}}) Get(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) (*{{.StructName}}, error) {
{{- if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(repo.kind, {{.KeyValueName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(repo.kind, {{.KeyValueName}}, nil)
{{else}}
	key := {{.KeyValueName}}
{{end}}
	subject := new({{.StructName}})
	err := repo.datastoreClient.Get(ctx, key, subject)
	if err != nil {
		return nil, err
	}
{{if eq .KeyFieldType "int64"}}
	subject.{{.KeyFieldName}} = key.ID
{{else if eq .KeyFieldType "string"}}
	subject.{{.KeyFieldName}} = key.Name
{{else}}
	subject.{{.KeyFieldName}} = key
{{end}}
	return subject, nil
}

func (repo *{{.RepositoryStructName}}) GetMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) ([]*{{.StructName}}, error) {
{{- if eq .KeyFieldType "int64"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.IDKey(repo.kind, {{.KeyValueName}}s[i], nil))
	}
{{else if eq .KeyFieldType "string"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.NameKey(repo.kind, {{.KeyValueName}}s[i], nil))
	}
{{else}}
	keys := {{.KeyValueName}}s
{{end}}
	vessels := make([]*{{.StructName}}, len({{.KeyValueName}}s))
	err := repo.datastoreClient.GetMulti(ctx, keys, vessels)

	for i := range vessels {
		if vessels[i] != nil {
{{- if eq .KeyFieldType "int64"}}
			vessels[i].{{.KeyFieldName}} = keys[i].ID
{{- else if eq .KeyFieldType "string"}}
			vessels[i].{{.KeyFieldName}} = keys[i].Name
{{- else}}
			vessels[i].{{.KeyFieldName}} = keys[i]
{{- end}}
		}
	}

	return vessels, err
}

func (repo *{{.RepositoryStructName}}) Insert(ctx context.Context, subject *{{.StructName}}) ({{.KeyFieldType}}, error) {
{{- if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(repo.kind, subject.{{.KeyFieldName}}, nil)
	zero := int64(0)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(repo.kind, subject.{{.KeyFieldName}}, nil)
	zero := ""
{{else}}
	key := subject.{{.KeyFieldName}}
	if key == nil {
		key = datastore.IncompleteKey(repo.kind, nil)
	}
	var zero {{.KeyFieldType}}
{{end}}
	key, err := repo.datastoreClient.Put(ctx, key, subject)
	if err != nil {
		return zero, err
	}
{{if eq .KeyFieldType "int64"}}
	return key.ID, nil
{{else if eq .KeyFieldType "string"}}
	return key.Name, nil
{{else}}
	return key, nil
{{end -}}
}

func (repo *{{.RepositoryStructName}}) InsertMulti(ctx context.Context, subjects []*{{.StructName}}) ([]{{.KeyFieldType}}, error) {
	keys := make([]*datastore.Key, 0, len(subjects))
{{- if eq .KeyFieldType "int64"}}
	for _, subject := range subjects {
		keys = append(keys, datastore.IDKey(repo.kind, subject.ID, nil))
	}
{{else if eq .KeyFieldType "string"}}
	for _, subject := range subjects {
		keys = append(keys, datastore.NameKey(repo.kind, subject.ID, nil))
	}
{{else}}
	for _, subject := range subjects {
		keys = append(keys, subject.ID)
	}
{{end}}
	keys, err := repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return nil, err
	}

	vessels := make([]{{.KeyFieldType}}, 0, len(keys))
	for i := range keys {
		if keys[i] != nil {
{{- if eq .KeyFieldType "int64"}}
			vessels[i] = keys[i].ID
{{- else if eq .KeyFieldType "string"}}
			vessels[i] = keys[i].Name
{{- else}}
			vessels[i] = keys[i]
{{- end}}
		}
	}

	return vessels, err
}

func (repo *{{.RepositoryStructName}}) Update(ctx context.Context, subject *{{.StructName}}) error {
{{- if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(repo.kind, subject.{{.KeyFieldName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(repo.kind, subject.{{.KeyFieldName}}, nil)
{{else}}
	key := subject.{{.KeyFieldName}}
{{end}}
	_, err := repo.datastoreClient.Put(ctx, key, subject)
	if err != nil {
		return err
	}

	return nil
}

func (repo *{{.RepositoryStructName}}) UpdateMulti(ctx context.Context, subjects []*{{.StructName}}) error {
	keys := make([]*datastore.Key, 0, len(subjects))
{{- if eq .KeyFieldType "int64"}}
	for _, subject := range subjects {
		keys = append(keys, datastore.IDKey(repo.kind, subject.ID, nil))
	}
{{else if eq .KeyFieldType "string"}}
	for _, subject := range subjects {
		keys = append(keys, datastore.NameKey(repo.kind, subject.ID, nil))
	}
{{else}}
	for _, subject := range subjects {
		keys = append(keys, subject.ID)
	}
{{end}}
	_, err := repo.datastoreClient.PutMulti(ctx, keys, subjects)
	if err != nil {
		return err
	}

	return nil
}

func (repo *{{.RepositoryStructName}}) Delete(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) error {
{{- if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(repo.kind, {{.KeyValueName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(repo.kind, {{.KeyValueName}}, nil)
{{else}}
	key := {{.KeyValueName}}
{{end}}
	return repo.datastoreClient.Delete(ctx, key)
}

func (repo *{{.RepositoryStructName}}) DeleteMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) error {
{{- if eq .KeyFieldType "int64"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.IDKey(repo.kind, {{.KeyValueName}}s[i], nil))
	}
{{else if eq .KeyFieldType "string"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.NameKey(repo.kind, {{.KeyValueName}}s[i], nil))
	}
{{else}}
	keys := {{.KeyValueName}}s
{{end}}
	return repo.datastoreClient.DeleteMulti(ctx, keys)
}
`
