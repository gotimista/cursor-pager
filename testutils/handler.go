// Package testutils provides utilities for testing.
package testutils

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

// AssertJSON asserts that the JSON bytes are equal.
func AssertJSON(t *testing.T, want, got []byte) {
	t.Helper()

	var jw, jg any
	if err := json.Unmarshal(want, &jw); err != nil {
		t.Fatalf("cannot unmarshal want %q: %v", want, err)
	}
	if err := json.Unmarshal(got, &jg); err != nil {
		t.Fatalf("cannot unmarshal got %q: %v", got, err)
	}
	if diff := cmp.Diff(jg, jw); diff != "" {
		t.Errorf("got differs: (-got +want)\n%s", diff)
	}
}

// LoadFile reads a file and returns its content.
func LoadFile(t *testing.T, path string) []byte {
	t.Helper()

	b, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return b
}
