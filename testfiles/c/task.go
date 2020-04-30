package task

import (
	"time"

	"cloud.google.com/go/datastore"
)

//go:generate repo_generator Task

type Task struct {
	Desc    string         `datastore:"description"`
	Created time.Time      `datastore:"created"`
	Done    bool           `datastore:"done"`
	ID      *datastore.Key `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
}
