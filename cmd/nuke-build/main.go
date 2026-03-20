package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// generatedBootstrap is written to build/_nuke/main.go.
// It imports the project's top-level nuke.go package and calls Build().
const generatedBootstrap = `package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/compiler/msvc"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/project"

	app "IMPORT_PATH"
)

func main() {
	rootDir := ROOT_DIR
	buildDir := filepath.Dir(rootDir) + "/build"

	tc, err := msvc.Discover()
	if err != nil {
		log.Fatal("MSVC not found:", err)
	}

	backend := msvc.NewBackend(tc)
	builder := cpp.NewBuilder(backend, buildDir)
	baseCfg := compiler.New().WithStandard(compiler.Cpp20).WithBuildType(compiler.Debug)

	t := app.Build(builder, baseCfg, rootDir)

	p, err := project.Open(buildDir, backend, baseCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	result, err := p.Build(context.Background(), t.ArtifactSet())
	if err != nil {
		log.Fatal("Build failed:", err)
	}

	fmt.Printf("Built: %s\n", result[0].URI())
}
`

func main() {
	projectDir := flag.String("project", "", "path to project directory containing nuke.go")
	buildDir := flag.String("build-dir", "", "output directory (default: <project-parent>/build)")
	flag.Parse()

	if *projectDir == "" {
		log.Fatal("--project is required")
	}

	absProject, err := filepath.Abs(*projectDir)
	if err != nil {
		log.Fatal(err)
	}

	absBuild := *buildDir
	if absBuild == "" {
		absBuild = filepath.Join(filepath.Dir(absProject), "build")
	} else {
		if absBuild, err = filepath.Abs(absBuild); err != nil {
			log.Fatal(err)
		}
	}

	if _, err := os.Stat(filepath.Join(absProject, "nuke.go")); err != nil {
		log.Fatalf("no nuke.go found in %s", absProject)
	}

	// rootDir is the parent of the project dir (e.g. physics-engine root)
	rootDir := filepath.Dir(absProject)

	// Repo root is two levels above rootDir (physics-engine -> examples -> repo)
	repoRoot := filepath.Dir(filepath.Dir(rootDir))

	// Determine the import path for the project's nuke.go package.
	// We read it from the project's go.mod module declaration.
	projectModName := readModuleName(absProject)
	if projectModName == "" {
		log.Fatalf("could not determine module name from %s/go.mod", absProject)
	}

	// Stage dir lives inside build — never touches the source tree
	stageDir := filepath.Join(absBuild, "_nuke")
	if err := os.MkdirAll(stageDir, 0755); err != nil {
		log.Fatal(err)
	}

	// Write go.mod for the bootstrap runner
	gomod := "module nuke_runner\n\ngo 1.26.1\n"
	if err := os.WriteFile(filepath.Join(stageDir, "go.mod"), []byte(gomod), 0644); err != nil {
		log.Fatal(err)
	}

	// Collect all sibling nuke module dirs (dirs containing nuke.go + go.mod)
	siblings := collectNukeDirs(rootDir, absProject)

	// Write go.work with absolute use paths: stageDir + siblings + repoRoot
	var w strings.Builder
	w.WriteString("go 1.26.1\n\nuse (\n")
	w.WriteString("\t.\n") // stage dir
	for _, s := range siblings {
		fmt.Fprintf(&w, "\t%s\n", filepath.ToSlash(s))
	}
	// Include the project dir itself (the nuke.go package to import)
	fmt.Fprintf(&w, "\t%s\n", filepath.ToSlash(absProject))
	// Include nuke-engine repo root to satisfy github.com/bradphelan/nuke-engine imports
	fmt.Fprintf(&w, "\t%s\n", filepath.ToSlash(repoRoot))
	w.WriteString(")\n")
	if err := os.WriteFile(filepath.Join(stageDir, "go.work"), []byte(w.String()), 0644); err != nil {
		log.Fatal(err)
	}

	// Write the bootstrap main.go, substituting import path and rootDir
	bootstrap := strings.ReplaceAll(generatedBootstrap, "IMPORT_PATH", projectModName)
	bootstrap = strings.ReplaceAll(bootstrap, "ROOT_DIR", fmt.Sprintf("%q", filepath.ToSlash(rootDir)))
	// Remove the unused filepath import if rootDir is hardcoded
	bootstrap = strings.ReplaceAll(bootstrap, "\t\"path/filepath\"\n", "")
	bootstrap = strings.ReplaceAll(bootstrap,
		`buildDir := filepath.Dir(rootDir) + "/build"`,
		fmt.Sprintf("buildDir := %q", filepath.ToSlash(absBuild)))
	if err := os.WriteFile(filepath.Join(stageDir, "main.go"), []byte(bootstrap), 0644); err != nil {
		log.Fatal(err)
	}

	// Compile the bootstrap into nuke-builder.exe
	builderExe := filepath.Join(absBuild, "nuke-builder.exe")
	build := exec.Command("go", "build", "-o", builderExe, ".")
	build.Dir = stageDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		log.Fatal("go build failed:", err)
	}

	// Run the builder — it uses nuke-engine to compile the C++ project
	run := exec.Command(builderExe)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		log.Fatal("builder failed:", err)
	}

	fmt.Println("nuke-build: done")
}

func readModuleName(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}

func collectNukeDirs(rootDir, excludeDir string) []string {
	var result []string
	entries, err := os.ReadDir(rootDir)
	if err != nil {
		return result
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		dir := filepath.Join(rootDir, e.Name())
		if dir == excludeDir {
			continue
		}
		if _, err := os.Stat(filepath.Join(dir, "nuke.go")); err == nil {
			result = append(result, dir)
		}
	}
	return result
}
