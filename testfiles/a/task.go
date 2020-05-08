package task

import (
	"time"
)

//go:generate repo_generator Task github.com/go-generalize/repo_generator/testfiles/a
//go:generate gofmt -w ./

type Task struct {
	ID      int64     `datastore:"-" datastore_key:""`     // supported type: string, int64, *datastore.Key
	Desc    string    `datastore:"description" filter:"l"` // supported word: m/matching(Default), l/like, p/prefix, TODO s/suffix
	Created time.Time `datastore:"created"`
	Done    bool      `datastore:"done"`
	Done2   bool      `datastore:"done2"`
	Count   int       `datastore:"count"`
	Count64 int64     `datastore:"count64"`
	Indexes []string  `datastore:"indexes"` // for XIAN
}
