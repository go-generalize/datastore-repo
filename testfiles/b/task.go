package task

import (
	"time"
)

//go:generate ds-repo Task
//go:generate gofmt -w ./

type Task struct {
	Desc    string    `datastore:"description"`
	Created time.Time `datastore:"created"`
	Done    bool      `datastore:"done"`
	ID      string    `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
}
