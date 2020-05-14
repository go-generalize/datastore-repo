package main

import (
	dstags "github.com/go-generalize/ds-repo/tools/datastore_tag_linter"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		dstags.Analyzer,
	)
}
