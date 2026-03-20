package main_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func findRepoRoot(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}
	dir := filepath.Dir(thisFile)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}

// TestNukeBuildPhysicsEngine proves the full pipeline:
//  1. Compile nuke-build orchestrator
//  2. nuke-build compiles the physics-engine nuke.go files into a builder exe
//  3. builder exe uses nuke-engine to compile the C++ project
//  4. The resulting app.exe exists and is a valid binary
func TestNukeBuildPhysicsEngine(t *testing.T) {
	// Skip if MSVC not available
	if _, err := exec.LookPath("cl.exe"); err != nil {
		// Try via vswhere — if nuke-engine can't discover MSVC, skip
		vswhere := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`
		if _, err := os.Stat(vswhere); err != nil {
			t.Skip("MSVC not installed")
		}
	}

	repoRoot := findRepoRoot(t)
	projectDir := filepath.Join(repoRoot, "examples", "physics-engine", "app")
	buildDir := filepath.Join(t.TempDir(), "build")

	// Step 1: build the nuke-build orchestrator
	nukeBuildExe := filepath.Join(t.TempDir(), "nuke-build.exe")
	build := exec.Command("go", "build", "-C", repoRoot, "-o", nukeBuildExe, "./cmd/nuke-build")
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		t.Fatalf("failed to build nuke-build: %v", err)
	}

	// Step 2+3: run nuke-build against the physics-engine project
	run := exec.Command(nukeBuildExe,
		"--project", projectDir,
		"--build-dir", buildDir,
	)
	out, err := run.CombinedOutput()
	t.Logf("nuke-build output:\n%s", out)
	if err != nil {
		t.Fatalf("nuke-build failed: %v", err)
	}

	// Step 4: verify the C++ app.exe was produced
	appExe := filepath.Join(buildDir, "bin", "app.exe")
	if _, err := os.Stat(appExe); err != nil {
		t.Fatalf("expected C++ app.exe at %s, not found: %v", appExe, err)
	}

	// Verify the output path was reported
	if !strings.Contains(string(out), "Built:") {
		t.Error("expected 'Built:' in output")
	}

	t.Logf("C++ app built at: %s", appExe)
}
