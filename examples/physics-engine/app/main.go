package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/compiler/msvc"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/project"
)

func main() {
	tc, err := msvc.Discover()
	if err != nil {
		log.Fatal("MSVC not found: ", err)
	}

	// rootDir should be the physics-engine directory (parent of app)
	absPath, _ := filepath.Abs(".")
	rootDir := filepath.Dir(absPath)
	buildDir := filepath.Join(rootDir, "build", "debug")

	backend := msvc.NewBackend(tc)
	builder := cpp.NewBuilder(backend, buildDir)
	baseCfg := compiler.New().WithStandard(compiler.Cpp20).WithBuildType(compiler.Debug)

	ap := Build(builder, baseCfg, rootDir)

	p, err := project.Open(buildDir, backend, baseCfg)
	if err != nil {
		log.Fatal(err)
	}
	defer p.Close()

	result, err := p.Build(context.Background(), ap.ArtifactSet())
	if err != nil {
		log.Fatal("Build failed: ", err)
	}

	fmt.Printf("Built: %s\n", result[0].URI())
}
