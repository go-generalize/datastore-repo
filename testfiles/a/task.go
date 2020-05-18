package task

import (
	"time"
)

//go:generate datastore-repo Task
//go:generate gofmt -w ./

// Task 拡張インデックスなし
type Task struct {
	ID         int64     `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
	Desc       string    `datastore:"description"`
	Created    time.Time `datastore:"created"`
	Done       bool      `datastore:"done"`
	Done2      bool      `datastore:"done2"`
	Count      int       `datastore:"count"`
	Count64    int64     `datastore:"count64"`
	NameList   []string  `datastore:"nameList"`
	Proportion float64   `datastore:"proportion"`
	Flag       Flag      `datastore:"flag"` // NG
}

type Flag bool
