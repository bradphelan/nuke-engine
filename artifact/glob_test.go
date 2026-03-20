package artifact

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func createTestTree(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	files := []string{
		"src/core/a.cpp",
		"src/core/b.cpp",
		"src/test/c_test.cpp",
		"src/main.cpp",
		"include/header.h",
	}
	for _, f := range files {
		p := filepath.Join(dir, f)
		os.MkdirAll(filepath.Dir(p), 0755)
		os.WriteFile(p, []byte("// "+f), 0644)
	}
	return dir
}

func TestGlobSetResolve(t *testing.T) {
	dir := createTestTree(t)
	g := NewGlobSet(dir, "**/*.cpp")
	arts, err := g.Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 4 {
		t.Fatalf("expected 4 .cpp files, got %d", len(arts))
	}
}

func TestGlobSetIsLazy(t *testing.T) {
	g := NewGlobSet("/nonexistent/path", "**/*.cpp")
	_ = g // no panic = lazy
}

func TestGlobSetSubFilter(t *testing.T) {
	dir := createTestTree(t)
	all := NewGlobSet(dir, "**/*.cpp")
	tests := NewGlobSet(dir, "**/*_test.cpp")
	src := all.Sub(tests)
	arts, _ := src.Resolve(context.Background())
	if len(arts) != 3 {
		t.Fatalf("expected 3 non-test cpp files, got %d", len(arts))
	}
}
