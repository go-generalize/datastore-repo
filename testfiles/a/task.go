package task

import (
	"time"
)

//go:generate repo_generator Task

type Task struct {
	Desc    string    `datastore:"description"`
	Created time.Time `datastore:"created"`
	Done    bool      `datastore:"done"`
	ID      int64     `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
}
