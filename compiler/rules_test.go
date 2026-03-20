package compiler

import (
	"strings"
	"testing"
)

func TestURIToPath(t *testing.T) {
	got := URIToPath("file:///src/main.cpp")
	// On Windows this should be \src\main.cpp (filepath.FromSlash)
	if got == "" {
		t.Fatal("should not be empty")
	}
	if got == "file:///src/main.cpp" {
		t.Fatal("should strip file:// prefix")
	}
	if strings.HasPrefix(got, "file://") {
		t.Fatalf("URI prefix not stripped, got: %s", got)
	}
}

func TestURIToPath_NoPrefix(t *testing.T) {
	// Paths without prefix should pass through unchanged (just slash conversion)
	got := URIToPath("/src/main.cpp")
	if got == "" {
		t.Fatal("should not be empty")
	}
}
