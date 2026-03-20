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

	"demo/mathcore"
	"demo/physics"
	"demo/renderer"
	"demo/simulation"
)

func main() {
	tc, err := msvc.Discover()
	if err != nil {
		log.Fatal("MSVC not found: ", err)
	}

	rootDir, _ := filepath.Abs(".")
	buildDir := filepath.Join(rootDir, "build", "debug")

	backend := msvc.NewBackend(tc)
	builder := cpp.NewBuilder(backend, buildDir)
	baseCfg := compiler.New().WithStandard(compiler.Cpp20).WithBuildType(compiler.Debug)

	mc := mathcore.Target(builder, baseCfg, rootDir)
	ph := physics.Target(builder, baseCfg, rootDir, mc)
	rn := renderer.Target(builder, baseCfg, rootDir, mc)
	sm := simulation.Target(builder, baseCfg, rootDir, ph, rn)
	ap := AppTarget(builder, baseCfg, rootDir, mc, sm)

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
