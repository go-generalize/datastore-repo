package dstags_test

import (
	"testing"

	dstags "github.com/go-generalize/repo_generator/tools/datastore_tag_linter"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestRun(t *testing.T) {
	testdata := analysistest.TestData()
	analysistest.Run(t, testdata, dstags.Analyzer, "a")
}
