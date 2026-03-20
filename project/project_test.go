package project

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
)

// mockBackend satisfies compiler.Backend without needing real tools
type mockBackend struct{}

func mockCommandArgs(msg string) (string, []string) {
	if runtime.GOOS == "windows" {
		return "cmd", []string{"/c", "echo", msg}
	}
	return "sh", []string{"-c", "echo " + msg}
}

func (m *mockBackend) ID() string { return "mock:1.0" }
func (m *mockBackend) CompileArgs(cfg compiler.Config, src, obj string) (string, []string, []string) {
	cmd, args := mockCommandArgs("compiled")
	return cmd, args, nil
}
func (m *mockBackend) StaticLibArgs(cfg compiler.Config, out string, objs []string) (string, []string, []string) {
	cmd, args := mockCommandArgs("archived")
	return cmd, args, nil
}
func (m *mockBackend) SharedLibArgs(cfg compiler.Config, out string, objs, libs []string) (string, []string, []string) {
	cmd, args := mockCommandArgs("linked-dll")
	return cmd, args, nil
}
func (m *mockBackend) ExeArgs(cfg compiler.Config, out string, objs, libs []string) (string, []string, []string) {
	cmd, args := mockCommandArgs("linked-exe")
	return cmd, args, nil
}
func (m *mockBackend) ParseDiscoveredDeps(output string) []string { return nil }

func TestProjectAccessors(t *testing.T) {
	dir := t.TempDir()
	cfg := compiler.New().WithStandard(compiler.Cpp20)
	mb := &mockBackend{}
	p, _ := Open(filepath.Join(dir, "build"), mb, cfg)
	defer p.Close()
	if p.Config().Standard() != compiler.Cpp20 { t.Fatal("config") }
	if p.Backend().ID() != "mock:1.0" { t.Fatal("backend") }
}

func TestProjectBuildDir(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build", "debug")

	p, err := Open(buildDir, &mockBackend{}, compiler.New())
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// Build dir should exist
	if _, err := os.Stat(buildDir); err != nil {
		t.Fatal("build dir should be created")
	}
	// Cache should be inside build dir
	if _, err := os.Stat(filepath.Join(buildDir, ".nuke-cache", "cache.db")); err != nil {
		t.Fatal("cache db should be inside build dir")
	}
}

func TestProjectFluentAPI(t *testing.T) {
	dir := t.TempDir()
	buildDir := filepath.Join(dir, "build")

	p, err := Open(buildDir, &mockBackend{}, compiler.New().WithStandard(compiler.Cpp20))
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	// Just verify the fluent API constructs the DAG without panicking
	src := artifact.NewSliceSet([]artifact.Artifact{})
	objs := p.CompileObjects("mylib", src)
	lib := p.StaticLib("mylib", objs)
	dll := p.SharedLib("mylib", objs)
	exe := p.Executable("app", objs)

	// Resolving empty inputs should succeed
	ctx := context.Background()
	_, err = p.Build(ctx, lib)
	if err != nil {
		t.Fatalf("StaticLib build failed: %v", err)
	}
	_, err = p.Build(ctx, dll)
	if err != nil {
		t.Fatalf("SharedLib build failed: %v", err)
	}
	_, err = p.Build(ctx, exe)
	if err != nil {
		t.Fatalf("Exe build failed: %v", err)
	}
}
