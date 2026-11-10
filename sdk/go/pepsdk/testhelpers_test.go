package pepsdk

import (
	"os"
	"path/filepath"
	"testing"
)

// repositoryRoot finds the checkout root by walking up to the directory that
// contains go.work, so tests work from both backend/pepsdk and sdk/go/pepsdk.
func repositoryRoot(t *testing.T) string {
	t.Helper()
	directory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(directory, "go.work")); err == nil {
			return directory
		}
		parent := filepath.Dir(directory)
		if parent == directory {
			t.Fatal("repository root containing go.work not found")
		}
		directory = parent
	}
}
