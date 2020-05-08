// +build emulator

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func execTest(t *testing.T) {
	t.Helper()

	b, err := exec.Command("go", "test", "./tests", "-v", "-tags", "internal").CombinedOutput()

	if err != nil {
		t.Fatalf("go test failed: %+v(%s)", err, string(b))
	}
}

func TestGenerator(t *testing.T) {
	root, err := os.Getwd()

	if err != nil {
		t.Fatalf("failed to getwd: %+v", err)
	}

	t.Run("int64", func(t *testing.T) {
		if err := os.Chdir(filepath.Join(root, "testfiles/a")); err != nil {
			t.Fatalf("chdir failed: %+v", err)
		}

		if err := run("Task"); err != nil {
			t.Fatalf("failed to generate for testfiles/a: %+v", err)
		}

		if err := run("Name"); err != nil {
			t.Fatalf("failed to generate for testfiles/a: %+v", err)
		}

		execTest(t)
	})

	t.Run("string", func(t *testing.T) {
		if err := os.Chdir(filepath.Join(root, "testfiles/b")); err != nil {
			t.Fatalf("chdir failed: %+v", err)
		}

		if err := run("Task"); err != nil {
			t.Fatalf("failed to generate for testfiles/b: %+v", err)
		}

		execTest(t)
	})

	t.Run("datastore.Key", func(t *testing.T) {
		if err := os.Chdir(filepath.Join(root, "testfiles/c")); err != nil {
			t.Fatalf("chdir failed: %+v", err)
		}

		if err := run("Task"); err != nil {
			t.Fatalf("failed to generate for testfiles/c: %+v", err)
		}

		execTest(t)
	})

}
