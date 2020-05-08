package task

import (
	"time"
)

//go:generate repo_generator Name github.com/go-generalize/repo_generator/testfiles/b

type Task struct {
	Desc    string    `datastore:"description"`
	Created time.Time `datastore:"created"`
	Done    bool      `datastore:"done"`
	ID      string    `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
}
