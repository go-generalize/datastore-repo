# datastore-repo

Cloud Datastoreで利用されるコードを自動生成する

### Installation
```console
$ go get github.com/go-generalize/datastore-repo
```

### Usage

```go
package task

import (
	"time"
)

//go:generate datastore-repo Task

type Task struct {
	ID      int64          `datastore:"-" datastore_key:""` // supported type: string, int64, *datastore.Key
	Desc    string         `datastore:"description"`
	Done    bool           `datastore:"done"`
	Created time.Time      `datastore:"created"`
}
```
`//go:generate` から始まる行を書くことでdatastore向けのclientを自動生成するようになる。

また、structの中で一つの要素は必ず`datastore_key:""`を持った要素が必要となっている。  
この要素の型は `int64`、 `string`、`*datastore.Key` のいずれかである必要がある。

`*datastore.Key`をkeyとした場合、Put時にkeyを `nil` としている場合は自動でkeyの割り当てが行われる。  
`int64`及び `string` の場合は自動生成されないため明示的に設定する必要がある。

この状態で`go generate` を実行すると`_gen.go`で終わるファイルにクライアントが生成される。

## DataStore Tag Linter
### Installation
```console
$ go get github.com/go-generalize/datastore-repo/tools/datastore_tag_linter/cmd/dstags
```
[README](https://github.com/go-generalize/datastore-repo/blob/master/tools/datastore_tag_linter/README.md)

## License
- Under the MIT License
- Copyright (C) 2020 go-generalize