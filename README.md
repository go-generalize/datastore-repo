## repo_generator

Cloud Datastoreで利用されるコードを自動生成する

### Installation
```console
$ go get github.com/go-generalize/repo_generator
```

### Usage

例: infra/datastore/task/task.go
```go
package task

import (
	"time"
)

//go:generate repo_generator Task

type Task struct {
	Desc    string         `datastore:"description"`
	Created time.Time      `datastore:"created"`
	Done    bool           `datastore:"done"`
	ID      int            `datastore:"-" datastore_key:""`
}
```
`//go:generate` から始まる行を書くことでdatastore向けのclientを自動生成するようになる。

また、structの中で一つの要素は必ず`datastore_key:""`を持った要素が必要となっている。
この要素の型は `int64`、 `string`、`*datastore.Key` のいずれかである必要がある。

`int64`をkeyとした場合、Put時にkeyを `0` としている場合は自動でkeyの割り当てが行われ、`0` 以外を指定するとその値をkeyとする要素が生成される。`string`及び `*datastore.Key` の場合は自動生成されないため明示的に設定する必要がある。

この状態で`go generate` を実行すると`_gen.go`で終わるファイルにクライアントが生成される。

### License
- Under the MIT License
- Copyright (C) 2020 go-generalize