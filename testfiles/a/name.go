package task

import (
	"time"
)

//go:generate repo_generator Name github.com/go-generalize/repo_generator/testfiles/a
//go:generate gofmt -w ./

type Name struct {
	ID      int64     `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
	Created time.Time `datastore:"created"`
	Desc    string    `datastore:"description"`
	Done    bool      `datastore:"done"`
	Count   int       `datastore:"count"`
}
