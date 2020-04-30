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

	KeyFieldName string
	KeyFieldType string

	KeyValueName string // lower camel case
}

func (g *generator) generate(writer io.Writer) {
	t := template.Must(template.New("tmpl").Parse(tmpl))

	err := t.Execute(writer, g)

	if err != nil {
		log.Printf("failed to execute template: %+v", err)
	}
}

const tmpl = `// THIS FILE IS A GENERATED CODE. DO NOT EDIT
package {{.PackageName}}

import (
	"context"

	"cloud.google.com/go/datastore"
)

const kind = "{{.StructName}}"

` + `//go:generate mockgen -source {{.GeneratedFileName}}.go -destination mock_{{.GeneratedFileName}}/mock_{{.GeneratedFileName}}.go
type Repository interface {
	Update(ctx context.Context, subject *{{.StructName}}) error
	Put(ctx context.Context, subject *{{.StructName}}) ({{.KeyFieldType}}, error)
	Get(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) (*{{.StructName}}, error)
	GetMulti(ctx context.Context, {{.KeyValueName}}s []{{.KeyFieldType}}) ([]*{{.StructName}}, error)
	Delete(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) error
}

type repository struct {
	datastoreClient *datastore.Client
}

func NewRepository(datastoreClient *datastore.Client) Repository {
	return &repository{datastoreClient: datastoreClient}
}

func (repo *repository) Update(ctx context.Context, subject *{{.StructName}}) error {
{{if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(kind, subject.{{.KeyFieldName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(kind, subject.{{.KeyFieldName}}, nil)
{{else}}
	key := subject.{{.KeyFieldName}}
{{end}}
	_, err := repo.datastoreClient.Put(ctx, key, subject)
	if err != nil {
		return err
	}

	return nil
}

func (repo *repository) Put(ctx context.Context, subject *{{.StructName}}) ({{.KeyFieldType}}, error) {
{{if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(kind, subject.{{.KeyFieldName}}, nil)
	zero := int64(0)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(kind, subject.{{.KeyFieldName}}, nil)
	zero := ""
{{else}}
	key := subject.{{.KeyFieldName}}
	if key == nil {
		key = datastore.IncompleteKey(kind, nil)
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
{{end}}
}

func (repo *repository) Get(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) (*{{.StructName}}, error) {
{{if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(kind, {{.KeyValueName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(kind, {{.KeyValueName}}, nil)
{{else}}
	key := {{.KeyValueName}}
{{end}}
	subject := &{{.StructName}}{}
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

func (repo *repository) GetMulti(
	ctx context.Context,
	{{.KeyValueName}}s []{{.KeyFieldType}},
) ([]*{{.StructName}}, error) {
{{if eq .KeyFieldType "int64"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.IDKey(kind, {{.KeyValueName}}s[i], nil))
	}
{{else if eq .KeyFieldType "string"}}
	keys := make([]*datastore.Key, 0, len({{.KeyValueName}}s))

	for i := range {{.KeyValueName}}s {
		keys = append(keys, datastore.NameKey(kind, {{.KeyValueName}}s[i], nil))
	}
{{else}}
	keys := {{.KeyValueName}}s
{{end}}
	vessels := make([]*{{.StructName}}, len({{.KeyValueName}}s))

	err := repo.datastoreClient.GetMulti(ctx, keys, vessels)

	for i := range vessels {
		if vessels[i] != nil {
{{if eq .KeyFieldType "int64"}}
			vessels[i].{{.KeyFieldName}} = keys[i].ID
{{else if eq .KeyFieldType "string"}}
			vessels[i].{{.KeyFieldName}} = keys[i].Name
{{else}}
			vessels[i].{{.KeyFieldName}} = keys[i]
{{end}}
		}
	}

	return vessels, err
}

func (repo *repository) Delete(ctx context.Context, {{.KeyValueName}} {{.KeyFieldType}}) error {
{{if eq .KeyFieldType "int64"}}
	key := datastore.IDKey(kind, {{.KeyValueName}}, nil)
{{else if eq .KeyFieldType "string"}}
	key := datastore.NameKey(kind, {{.KeyValueName}}, nil)
{{else}}
	key := {{.KeyValueName}}
{{end}}
	return repo.datastoreClient.Delete(ctx, key)
}
`
