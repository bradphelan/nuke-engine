package project

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/compiler/msvc"
	"github.com/bradphelan/nuke-engine/rule"
)

// writeFile is a test helper that creates a file with the given content,
// creating parent directories as needed.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	os.MkdirAll(filepath.Dir(path), 0755)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// discoverMSVC returns an MSVC backend and toolchain, or skips the test if
// MSVC is not installed.
func discoverMSVC(t *testing.T) *msvc.Backend {
	t.Helper()
	tc, err := msvc.Discover()
	if err != nil {
		t.Skipf("MSVC not available: %v", err)
	}
	return msvc.NewBackend(tc)
}

// --- mockProtocRule: .proto -> .pb.cc + .pb.h ---

type mockProtocRule struct {
	outputDir string
}

func (r *mockProtocRule) ID() string { return "mock-protoc" }

func (r *mockProtocRule) Apply(_ context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	os.MkdirAll(r.outputDir, 0755)
	var outputs []artifact.Artifact

	for _, input := range inputs {
		uri := input.URI()
		base := filepath.Base(strings.TrimPrefix(uri, "file://"))
		stem := strings.TrimSuffix(base, filepath.Ext(base))

		// Generate .pb.h
		hPath := filepath.Join(r.outputDir, stem+".pb.h")
		hContent := fmt.Sprintf(
			"#pragma once\nint %s_value();\n", stem)
		if err := os.WriteFile(hPath, []byte(hContent), 0644); err != nil {
			return nil, nil, err
		}
		outputs = append(outputs, artifact.NewFileArtifact(hPath))

		// Generate .pb.cc
		ccPath := filepath.Join(r.outputDir, stem+".pb.cc")
		ccContent := fmt.Sprintf(
			"#include \"%s.pb.h\"\nint %s_value() { return 42; }\n", stem, stem)
		if err := os.WriteFile(ccPath, []byte(ccContent), 0644); err != nil {
			return nil, nil, err
		}
		outputs = append(outputs, artifact.NewFileArtifact(ccPath))
	}

	return outputs, nil, nil
}

// TestProtobufChaining proves multi-stage artifact chaining:
// .proto -> mock protoc -> .pb.cc/.pb.h -> compile -> .obj -> link -> .exe
func TestProtobufChaining(t *testing.T) {
	backend := discoverMSVC(t)

	tmpDir := t.TempDir()
	protoDir := filepath.Join(tmpDir, "proto")
	srcDir := filepath.Join(tmpDir, "src")
	buildDir := filepath.Join(tmpDir, "build")
	genDir := filepath.Join(buildDir, "gen")

	// Create .proto files (content doesn't matter — mock protoc ignores it)
	writeFile(t, protoDir, "greeter.proto", `syntax = "proto3";`)

	// Create main.cpp that calls the generated function
	writeFile(t, srcDir, "main.cpp", `
#include "greeter.pb.h"
int main() { return greeter_value() == 42 ? 0 : 1; }
`)

	// Include the gen dir so #include "greeter.pb.h" resolves
	cfg := compiler.New().
		WithStandard(compiler.Cpp17).
		WithBuildType(compiler.Debug).
		WithIncludeDir(genDir)

	p, err := Open(buildDir, backend, cfg)
	if err != nil {
		t.Fatalf("Open project: %v", err)
	}
	defer p.Close()

	ctx := context.Background()

	// Build the DAG
	protos := artifact.Glob(protoDir, "**/*.proto")
	generated := rule.NewTransformSet(protos, &mockProtocRule{outputDir: genDir})
	generatedSrc := generated.Filter("**/*.cc")
	handSrc := artifact.Glob(srcDir, "**/*.cpp")
	allSrc := handSrc.Add(generatedSrc)
	objs := p.CompileObjects("app", allSrc)
	exe := p.Executable("app", objs)

	result, err := p.Build(ctx, exe)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one output artifact")
	}

	// Verify the exe exists on disk
	exeURI := result[0].URI()
	exePath := strings.TrimPrefix(exeURI, "file://")
	exePath = filepath.FromSlash(exePath)
	if _, err := os.Stat(exePath); err != nil {
		t.Fatalf("exe not found on disk at %s: %v", exePath, err)
	}
	t.Logf("exe built at %s", exePath)
}

// TestSidecarDepInvalidation verifies that modifying a header (sidecar dep)
// changes its fingerprint, proving the engine would cache-miss on next build.
func TestSidecarDepInvalidation(t *testing.T) {
	backend := discoverMSVC(t)

	tmpDir := t.TempDir()
	srcDir := filepath.Join(tmpDir, "src")
	buildDir := filepath.Join(tmpDir, "build")

	// Create initial header and source
	writeFile(t, srcDir, "header.h", `#pragma once
#define VALUE 1
`)
	writeFile(t, srcDir, "main.cpp", `#include "header.h"
int main() { return VALUE; }
`)

	cfg := compiler.New().
		WithStandard(compiler.Cpp17).
		WithBuildType(compiler.Debug).
		WithIncludeDir(srcDir)

	// Use CompileRule directly (not through Project) to get discovered deps
	cr := compiler.NewCompileRule(backend, cfg, "sidecar-test", buildDir)

	ctx := context.Background()
	srcPath := filepath.Join(srcDir, "main.cpp")
	inputs := []artifact.Artifact{artifact.NewFileArtifact(srcPath)}

	// First compile — discover deps
	_, discovered, err := cr.Apply(ctx, inputs)
	if err != nil {
		t.Fatalf("first compile failed: %v", err)
	}

	// Find header.h among discovered deps
	var headerArtifact artifact.Artifact
	for _, d := range discovered {
		if strings.HasSuffix(d.URI(), "header.h") {
			headerArtifact = d
			break
		}
	}
	if headerArtifact == nil {
		t.Fatal("header.h not found in discovered dependencies")
	}

	fpBefore := headerArtifact.Fingerprint()
	t.Logf("header fingerprint before: %s", fpBefore)

	// Modify header.h
	writeFile(t, srcDir, "header.h", `#pragma once
#define VALUE 999
`)

	// Create a new FileArtifact for the modified file to get fresh fingerprint
	headerPath := strings.TrimPrefix(headerArtifact.URI(), "file://")
	headerPath = filepath.FromSlash(headerPath)
	freshHeader := artifact.NewFileArtifact(headerPath)
	fpAfter := freshHeader.Fingerprint()
	t.Logf("header fingerprint after: %s", fpAfter)

	if fpBefore == fpAfter {
		t.Fatal("fingerprint should have changed after modifying header.h")
	}
	t.Log("sidecar dep fingerprint changed — engine would cache-miss")
}
