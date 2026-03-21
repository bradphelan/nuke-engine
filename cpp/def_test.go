package cpp

import (
	"path/filepath"
	"testing"
)

func TestSplitModuleGlob(t *testing.T) {
	tests := []struct {
		name       string
		moduleDir  string
		pattern    string
		wantBase   string
		wantRelPat string
	}{
		{name: "recursive source glob", moduleDir: "/repo/app", pattern: "src/**/*.cpp", wantBase: "/repo/app/src", wantRelPat: "**/*.cpp"},
		{name: "single segment glob", moduleDir: "/repo/app", pattern: "*.cpp", wantBase: "/repo/app", wantRelPat: "*.cpp"},
		{name: "nested literal path then glob", moduleDir: "/repo/app", pattern: "include/generated/*.h", wantBase: "/repo/app/include/generated", wantRelPat: "*.h"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			base, relPattern := splitModuleGlob(test.moduleDir, test.pattern)
			if filepath.ToSlash(base) != test.wantBase {
				t.Fatalf("base = %q, want %q", base, test.wantBase)
			}
			if relPattern != test.wantRelPat {
				t.Fatalf("relPattern = %q, want %q", relPattern, test.wantRelPat)
			}
		})
	}
}
