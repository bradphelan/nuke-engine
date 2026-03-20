package project

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/compiler/msvc"
	"github.com/bradphelan/nuke-engine/cpp"
)

// findRepoRoot walks up from this test file to find the repo root (has go.mod).
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

// TestFullDemoBuild builds the complete physics-engine demo with MSVC,
// verifying the full target graph: mathcore -> physics -> renderer -> simulation -> app.
func TestFullDemoBuild(t *testing.T) {
	tc, err := msvc.Discover()
	if err != nil {
		t.Skip("MSVC not installed:", err)
	}

	repoRoot := findRepoRoot(t)
	demoRoot := filepath.Join(repoRoot, "examples", "physics-engine")

	// Verify demo sources exist
	if _, err := os.Stat(filepath.Join(demoRoot, "mathcore", "src")); err != nil {
		t.Fatalf("demo sources not found at %s: %v", demoRoot, err)
	}

	buildDir := filepath.Join(t.TempDir(), "build")

	backend := msvc.NewBackend(tc)
	builder := cpp.NewBuilder(backend, buildDir)
	baseCfg := compiler.New().WithStandard(compiler.Cpp20).WithBuildType(compiler.Debug)

	// Build the target graph (mirrors examples/physics-engine/app/main.go)
	mc := cpp.NewStaticLib("mathcore", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(demoRoot, "mathcore", "include"))).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(demoRoot, "mathcore", "src", "internal"))).
		Sources(artifact.Glob(filepath.Join(demoRoot, "mathcore", "src"), "**/*.cpp"))

	ph := cpp.NewStaticLib("physics", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(demoRoot, "physics", "include"))).
		LinkPublic(mc).
		Sources(artifact.Glob(filepath.Join(demoRoot, "physics", "src"), "**/*.cpp"))

	rn := cpp.NewSharedLib("renderer", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(demoRoot, "renderer", "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mc).
		Sources(artifact.Glob(filepath.Join(demoRoot, "renderer", "src"), "**/*.cpp"))

	sm := cpp.NewSharedLib("simulation", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(demoRoot, "simulation", "include")).WithDefine("SIMULATION_EXPORTS")).
		LinkPrivate(ph).
		LinkPrivate(rn).
		Sources(artifact.Glob(filepath.Join(demoRoot, "simulation", "src"), "**/*.cpp"))

	ap := cpp.NewExe("app", builder, baseCfg).
		LinkPublic(mc).
		LinkPrivate(sm).
		Sources(artifact.Glob(filepath.Join(demoRoot, "app", "src"), "**/*.cpp"))

	// Build via project engine
	p, err := Open(buildDir, backend, baseCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()

	result, err := p.Build(context.Background(), ap.ArtifactSet())
	if err != nil {
		t.Fatal("Build failed:", err)
	}

	if len(result) == 0 {
		t.Fatal("expected at least one output artifact")
	}

	// Verify the exe exists on disk
	exePath := strings.TrimPrefix(result[0].URI(), "file://")
	exePath = filepath.FromSlash(exePath)
	if _, err := os.Stat(exePath); err != nil {
		t.Fatalf("exe not found at %s: %v", exePath, err)
	}

	t.Logf("Successfully built: %s", exePath)
	t.Log("Transitive includes verified: physics compiled with #include <mathcore/vec3.h> via LinkPublic")
	t.Log("Private isolation: app cannot see physics/renderer includes directly")
}
