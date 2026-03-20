package project

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/engine"
	"github.com/bradphelan/nuke-engine/rule"
)

type Project struct {
	buildDir string
	backend  compiler.Backend
	config   compiler.Config
	engine   *engine.Engine
}

func Open(buildDir string, backend compiler.Backend, config compiler.Config) (*Project, error) {
	os.MkdirAll(buildDir, 0755)
	cacheDir := filepath.Join(buildDir, ".nuke-cache")
	os.MkdirAll(cacheDir, 0755)

	eng, err := engine.New(filepath.Join(cacheDir, "cache.db"))
	if err != nil {
		return nil, err
	}
	return &Project{
		buildDir: buildDir,
		backend:  backend,
		config:   config,
		engine:   eng,
	}, nil
}

func (p *Project) Close() error { return p.engine.Close() }

func (p *Project) Build(ctx context.Context, target artifact.ArtifactSet) ([]artifact.Artifact, error) {
	return p.engine.Resolve(ctx, target)
}

func (p *Project) BuildDir() string { return p.buildDir }

// --- Fluent target builders ---

func (p *Project) CompileObjects(name string, sources artifact.ArtifactSet) artifact.ArtifactSet {
	r := compiler.NewCompileRule(p.backend, p.config, name, p.buildDir)
	return rule.NewTransformSet(sources, r)
}

func (p *Project) StaticLib(name string, objects artifact.ArtifactSet) artifact.ArtifactSet {
	r := compiler.NewStaticLibRule(p.backend, p.config, name, p.buildDir)
	return rule.NewTransformSet(objects, r)
}

func (p *Project) SharedLib(name string, objects artifact.ArtifactSet) artifact.ArtifactSet {
	r := compiler.NewSharedLibRule(p.backend, p.config, name, p.buildDir)
	return rule.NewTransformSet(objects, r)
}

func (p *Project) Executable(name string, objects artifact.ArtifactSet) artifact.ArtifactSet {
	r := compiler.NewExeRule(p.backend, p.config, name, p.buildDir)
	return rule.NewTransformSet(objects, r)
}
