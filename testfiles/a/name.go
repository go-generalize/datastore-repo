package task

import (
	"time"
)

//go:generate repo_generator Name
//go:generate gofmt -w ./

// Name 拡張インデックスあり
type Name struct {
	ID      int64     `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
	Created time.Time `datastore:"created"`
	// supported indexer tags word: e/equal(Default), l/like, p/prefix,
	// TODO s/suffix
	Desc    string   `datastore:"description" indexer:"l"`
	Desc2   string   `datastore:"description2" indexer:"p"`
	Done    bool     `datastore:"done"`
	Count   int      `datastore:"count"`
	Indexes []string `datastore:"indexes"`
}
