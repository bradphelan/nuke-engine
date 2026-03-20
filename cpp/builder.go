package cpp

import (
	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/rule"
	"github.com/bradphelan/nuke-engine/target"
)

type CppBuilder struct {
	backend  compiler.Backend
	buildDir string
}

func NewBuilder(backend compiler.Backend, buildDir string) *CppBuilder {
	return &CppBuilder{backend: backend, buildDir: buildDir}
}

func (b *CppBuilder) Backend() compiler.Backend { return b.backend }
func (b *CppBuilder) BuildDir() string          { return b.buildDir }

func (b *CppBuilder) EmptyConfig() compiler.Config {
	return compiler.New()
}

func (b *CppBuilder) Compile(name string, cfg compiler.Config, sources artifact.ArtifactSet) artifact.ArtifactSet {
	return rule.NewTransformSet(sources, compiler.NewCompileRule(b.backend, cfg, name, b.buildDir))
}

func (b *CppBuilder) StaticLib(name string, cfg compiler.Config, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return rule.NewTransformSet(inputs, compiler.NewStaticLibRule(b.backend, cfg, name, b.buildDir))
}

func (b *CppBuilder) SharedLib(name string, cfg compiler.Config, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return rule.NewTransformSet(inputs, compiler.NewSharedLibRule(b.backend, cfg, name, b.buildDir))
}

func (b *CppBuilder) Exe(name string, cfg compiler.Config, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return rule.NewTransformSet(inputs, compiler.NewExeRule(b.backend, cfg, name, b.buildDir))
}

// Convenience constructors

func NewStaticLib(name string, builder *CppBuilder, baseCfg compiler.Config) *target.Target[compiler.Config] {
	return target.NewStaticLib(name, baseCfg, builder)
}

func NewSharedLib(name string, builder *CppBuilder, baseCfg compiler.Config) *target.Target[compiler.Config] {
	return target.NewSharedLib(name, baseCfg, builder)
}

func NewExe(name string, builder *CppBuilder, baseCfg compiler.Config) *target.Target[compiler.Config] {
	return target.NewExe(name, baseCfg, builder)
}
