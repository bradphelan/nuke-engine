package msvc

import (
	"strings"
	"testing"

	"github.com/bradphelan/nuke-engine/compiler"
)

func TestBackendCompileFlags(t *testing.T) {
	if !hasMSVC() {
		t.Skip("MSVC not installed")
	}

	tc, err := Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	b := NewBackend(tc)

	cfg := compiler.New().
		WithStandard(compiler.Cpp20).
		WithBuildType(compiler.Release).
		WithDefine("FOO=1").
		WithIncludeDir(`C:\myinc`)

	cmd, args, env := b.CompileArgs(cfg, "main.cpp", "main.obj")

	if cmd != tc.CL {
		t.Errorf("cmd = %q, want %q", cmd, tc.CL)
	}

	want := map[string]bool{
		"/nologo":       true,
		"/c":            true,
		"/EHsc":         true,
		"/showIncludes": true,
		"/std:c++20":    true,
		"/O2":           true,
		"/DNDEBUG":      true,
		"/MD":           true,
		"/DFOO=1":       true,
		`/IC:\myinc`:    true,
	}

	joined := strings.Join(args, " ")
	for flag := range want {
		if !strings.Contains(joined, flag) {
			t.Errorf("missing flag %q in args: %v", flag, args)
		}
	}

	// Check source and object are present.
	if args[len(args)-1] != "main.cpp" {
		t.Errorf("last arg = %q, want main.cpp", args[len(args)-1])
	}
	if args[len(args)-2] != "/Fomain.obj" {
		t.Errorf("second-to-last arg = %q, want /Fomain.obj", args[len(args)-2])
	}

	// env must not be empty.
	if len(env) == 0 {
		t.Error("env is empty")
	}
}

func TestParseDiscoveredDeps(t *testing.T) {
	b := &Backend{tc: &Toolchain{}}

	output := "main.cpp\r\n" +
		"Note: including file:   C:\\sdk\\stdio.h\r\n" +
		"Note: including file:     C:\\sdk\\corecrt.h\r\n" +
		"some other line\r\n" +
		"Note: including file: D:\\proj\\myheader.h\r\n"

	deps := b.ParseDiscoveredDeps(output)

	if len(deps) != 3 {
		t.Fatalf("got %d deps, want 3: %v", len(deps), deps)
	}

	expected := []string{
		`C:\sdk\stdio.h`,
		`C:\sdk\corecrt.h`,
		`D:\proj\myheader.h`,
	}
	for i, want := range expected {
		if deps[i] != want {
			t.Errorf("deps[%d] = %q, want %q", i, deps[i], want)
		}
	}
}
