// Package testutils provides utilities for testing.
package testutils

import (
	"encoding/json"
	"io"
	"net/http"
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

// AssertResponse asserts that the response status code and body are equal to the expected values.
func AssertResponse(t *testing.T, got *http.Response, status int, body []byte) {
	t.Helper()
	t.Cleanup(func() { _ = got.Body.Close() })
	gb, err := io.ReadAll(got.Body)
	if err != nil {
		t.Fatal(err)
	}
	if got.StatusCode != status {
		t.Fatalf("want status %d, but got %d, body: %q", status, got.StatusCode, gb)
	}

	if len(gb) == 0 && len(body) == 0 {
		// 期待としても実体としてもレスポンスボディがないので
		// AssertJSONを呼ぶ必要はない。
		return
	}
	AssertJSON(t, body, gb)
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
