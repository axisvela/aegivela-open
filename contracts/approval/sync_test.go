package approval

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func fixturePaths(baseDir, subDir string) []string {
	entries, err := os.ReadDir(filepath.Join(baseDir, subDir))
	if err != nil {
		t.Fatalf("cannot read fixtures dir %s/%s: %v", baseDir, subDir, err)
	}
	var paths []string
	for _, e := range entries {
		if !e.IsDir() && (filepath.Ext(e.Name()) == ".json") {
			paths = append(paths, filepath.Join(subDir, e.Name()))
		}
	}
	return paths
}

func readFixture(t *testing.T, baseDir, path string) map[string]json.RawMessage {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(baseDir, path))
	if err != nil {
		t.Fatalf("cannot read fixture %s: %v", path, err)
	}
	var result map[string]json.RawMessage
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("cannot parse fixture %s: %v", path, err)
	}
	return result
}

func TestApprovalFixturesConformance(t *testing.T) {
	baseDir := "/d/profile/paper-code/Aegivela/contracts"

	validPaths := fixturePaths(baseDir, "approval/v1alpha2/fixtures/valid")
	invalidPaths := fixturePaths(baseDir, "approval/v1alpha2/fixtures/invalid")

	for _, p := range validPaths {
		fixture := readFixture(t, baseDir, p)
		if _, ok := fixture["request"]; !ok {
			t.Fatalf("valid fixture %s missing 'request' field", p)
		}
		if _, ok := fixture["response"]; !ok {
			t.Fatalf("valid fixture %s missing 'response' field", p)
		}
	}

	for _, p := range invalidPaths {
		fixture := readFixture(t, baseDir, p)
		if _, ok := fixture["request"]; !ok {
			t.Fatalf("invalid fixture %s missing 'request' field", p)
		}
	}
}

func TestRevocFixturesConformance(t *testing.T) {
	baseDir := "/d/profile/paper-code/Aegivela/contracts"

	validPaths := fixturePaths(baseDir, "revocation/v1alpha1/fixtures/valid")
	invalidPaths := fixturePaths(baseDir, "revocation/v1alpha1/fixtures/invalid")

	for _, p := range validPaths {
		fixture := readFixture(t, baseDir, p)
		if _, ok := fixture["request"]; !ok {
			t.Fatalf("valid revocation fixture %s missing 'request' field", p)
		}
		if _, ok := fixture["response"]; !ok {
			t.Fatalf("valid revocation fixture %s missing 'response' field", p)
		}
	}

	for _, p := range invalidPaths {
		fixture := readFixture(t, baseDir, p)
		if _, ok := fixture["request"]; !ok {
			t.Fatalf("invalid revocation fixture %s missing 'request' field", p)
		}
	}
}