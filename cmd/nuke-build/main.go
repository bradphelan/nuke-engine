package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// version is set at build time via -ldflags="-X main.version=..."
var version = "dev"

// generatedBootstrap is written to build/_nuke/main.go.
// It imports the project's top-level nuke.go package and calls Build().
const generatedBootstrap = `package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/compiler/msvc"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/engine"
	"github.com/bradphelan/nuke-engine/project"

	app "IMPORT_PATH"
)

func main() {
	tree := flag.Bool("tree", false, "print dependency tree and exit")
	treeGraphvis := flag.Bool("tree-graphvis", false, "print dependency DAG as Graphviz DOT and exit")
	treeGraphviz := flag.Bool("tree-graphviz", false, "alias for --tree-graphvis")
	compileCommands := flag.Bool("compile-commands", false, "write compile_commands.json and exit")
	compileCommandsJSON := flag.Bool("compile-commands.json", false, "alias for --compile-commands")
	flag.Parse()

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

	if *compileCommands || *compileCommandsJSON {
		outPath := filepath.Join(rootDir, "compile_commands.json")
		if err := engine.WriteCompileCommandsJSON(outPath, t.ArtifactSet()); err != nil {
			log.Fatal("compile_commands generation failed:", err)
		}
		fmt.Printf("Wrote: %s\n", filepath.ToSlash(outPath))
		return
	}

	if *treeGraphvis || *treeGraphviz {
		engine.PrintTreeGraphviz(os.Stdout, t.ArtifactSet())
		return
	}

	if *tree {
		engine.PrintTree(os.Stdout, t.ArtifactSet())
		return
	}

	p, err := project.Open(buildDir, backend, baseCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	result, stats, err := p.BuildWithStats(context.Background(), t.ArtifactSet())
	if err != nil {
		log.Fatal("Build failed:", err)
	}

	fmt.Printf("Stats: cache_hits=%d cache_misses=%d rules_executed=%d outputs=%d\n",
		stats.CacheHits, stats.CacheMisses, stats.RulesExecuted, len(result))
	if stats.RulesExecuted == 0 {
		fmt.Println("No tasks need running (up-to-date).")
	} else {
		fmt.Printf("Tasks executed: %d\n", stats.RulesExecuted)
	}

	fmt.Printf("Built: %s\n", result[0].URI())
}
`

func main() {
	// Manually parse only our own flags so that unrecognised flags are
	// forwarded to build.exe / build.out rather than causing an error.
	var projectDir, buildDir string
	var showVersion bool

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		s := args[i]
		name, val, hasEq := strings.Cut(strings.TrimLeft(s, "-"), "=")
		switch name {
		case "project":
			if hasEq {
				projectDir = val
			} else if i+1 < len(args) {
				projectDir = args[i+1]
				i++
			}
		case "build-dir":
			if hasEq {
				buildDir = val
			} else if i+1 < len(args) {
				buildDir = args[i+1]
				i++
			}
		case "version":
			showVersion = true
		case "h", "help":
			fmt.Fprintln(os.Stderr, "Usage: nuke-build --project <dir> [--build-dir <dir>] [--version]")
			fmt.Fprintln(os.Stderr, "All other flags are forwarded to build.exe / build.out.")
			os.Exit(0)
		}
	}

	if showVersion {
		fmt.Println(version)
		return
	}

	if projectDir == "" {
		log.Fatal("--project is required")
	}

	absProject, err := filepath.Abs(projectDir)
	if err != nil {
		log.Fatal(err)
	}

	absBuild := buildDir
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
	bootstrap = strings.ReplaceAll(bootstrap,
		`buildDir := filepath.Dir(rootDir) + "/build"`,
		fmt.Sprintf("buildDir := %q", filepath.ToSlash(absBuild)))
	if err := os.WriteFile(filepath.Join(stageDir, "main.go"), []byte(bootstrap), 0644); err != nil {
		log.Fatal(err)
	}

	// Compile the bootstrap into build.exe (Windows) / build.out (Linux/macOS)
	builderName := "build.exe"
	if runtime.GOOS != "windows" {
		builderName = "build.out"
	}
	builderExe := filepath.Join(absBuild, builderName)
	build := exec.Command("go", "build", "-o", builderExe, ".")
	build.Dir = stageDir
	build.Stdout = os.Stdout
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		log.Fatal("go build failed:", err)
	}

	// Run the builder — forward all unrecognised flags transparently
	run := exec.Command(builderExe, forwardedArgs()...)
	run.Stdout = os.Stdout
	run.Stderr = os.Stderr
	if err := run.Run(); err != nil {
		log.Fatal("builder failed:", err)
	}
}

// forwardedArgs returns os.Args[1:] with nuke-build-specific flags removed.
// Everything else is passed through verbatim to the builder binary.
func forwardedArgs() []string {
	var result []string
	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		s := strings.TrimLeft(args[i], "-")
		name, _, hasEq := strings.Cut(s, "=")
		switch name {
		case "project", "build-dir":
			// Skip flag; if value is in the next arg, skip that too.
			if !hasEq && i+1 < len(args) {
				i++
			}
		case "version":
			// Boolean; no following value.
		default:
			result = append(result, args[i])
		}
	}
	return result
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
