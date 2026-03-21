package cpp

import (
	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/rule"
	"github.com/bradphelan/nuke-engine/target"
)

type configScope int

const (
	configPublic configScope = iota
	configPrivate
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

func wrapTarget(ctx *Context, raw *target.Target[compiler.Config]) *Target {
	return &Target{ctx: ctx, raw: raw, scope: configPublic}
}

// Collect returns the unique target graph reachable from one or more roots.
func Collect(roots ...*Target) []*Target {
	seen := make(map[*target.Target[compiler.Config]]struct{})
	var out []*Target
	for _, root := range roots {
		if root == nil {
			continue
		}
		for _, raw := range target.Collect(root.raw) {
			if _, ok := seen[raw]; ok {
				continue
			}
			seen[raw] = struct{}{}
			out = append(out, wrapTarget(root.ctx, raw))
		}
	}
	return out
}

// Target is a C++-specific fluent wrapper around the generic target.Target.
// It carries optional context so declarations can link Defs directly.
type Target struct {
	ctx   *Context
	raw   *target.Target[compiler.Config]
	scope configScope
}

func (t *Target) Raw() *target.Target[compiler.Config] { return t.raw }

func (t *Target) Name() string { return t.raw.Name() }

func (t *Target) Kind() target.Kind { return t.raw.Kind() }

func (t *Target) PublicCfg() compiler.Config { return t.raw.PublicCfg() }

func (t *Target) PrivateCfg() compiler.Config { return t.raw.PrivateCfg() }

func (t *Target) ResolvedConfig() compiler.Config { return t.raw.ResolvedConfig() }

func (t *Target) Public() *Target {
	t.scope = configPublic
	return t
}

func (t *Target) Private() *Target {
	t.scope = configPrivate
	return t
}

func (t *Target) applyConfig(cfg compiler.Config) {
	if t.scope == configPrivate {
		t.raw.PrivateConfig(cfg)
		return
	}
	t.raw.PublicConfig(cfg)
}

func (t *Target) WithStandard(s compiler.Standard) *Target {
	t.applyConfig(compiler.New().WithStandard(s))
	return t
}

func (t *Target) WithBuildType(bt compiler.BuildType) *Target {
	t.applyConfig(compiler.New().WithBuildType(bt))
	return t
}

func (t *Target) WithDefine(d string) *Target {
	t.applyConfig(compiler.New().WithDefine(d))
	return t
}

func (t *Target) WithIncludeDir(dir string) *Target {
	t.applyConfig(compiler.New().WithIncludeDir(dir))
	return t
}

func (t *Target) WithCFlags(flags ...string) *Target {
	t.applyConfig(compiler.New().WithCFlags(flags...))
	return t
}

func (t *Target) WithLDFlags(flags ...string) *Target {
	t.applyConfig(compiler.New().WithLDFlags(flags...))
	return t
}

func (t *Target) ExportAllSymbols(b bool) *Target {
	t.applyConfig(compiler.New().ExportAllSymbols(b))
	return t
}

func (t *Target) PublicConfig(configure func(compiler.Config) compiler.Config) *Target {
	if configure == nil {
		panic("cpp.Target PublicConfig requires a non-nil configure function")
	}
	t.raw.PublicConfig(configure(compiler.New()))
	t.scope = configPublic
	return t
}

func (t *Target) PrivateConfig(configure func(compiler.Config) compiler.Config) *Target {
	if configure == nil {
		panic("cpp.Target PrivateConfig requires a non-nil configure function")
	}
	t.raw.PrivateConfig(configure(compiler.New()))
	t.scope = configPrivate
	return t
}

func (t *Target) Sources(sources artifact.ArtifactSet) *Target {
	t.raw.Sources(sources)
	return t
}

func (t *Target) ArtifactSet() artifact.ArtifactSet { return t.raw.ArtifactSet() }

func (t *Target) linkTarget(dep any, public bool) *Target {
	var resolved *Target
	switch value := dep.(type) {
	case *Target:
		resolved = value
	case Def:
		if t.ctx == nil {
			panic("cpp.Target link with Def requires context")
		}
		resolved = value.Resolve(t.ctx)
	default:
		panic("cpp.Target link expects *cpp.Target or cpp.Def")
	}
	if public {
		t.raw.LinkPublic(resolved.raw)
	} else {
		t.raw.LinkPrivate(resolved.raw)
	}
	return t
}

func (t *Target) LinkPublic(dep any) *Target {
	return t.linkTarget(dep, true)
}

func (t *Target) LinkPrivate(dep any) *Target {
	return t.linkTarget(dep, false)
}

// Convenience constructors

func NewStaticLib(name string, builder *CppBuilder, baseCfg compiler.Config) *Target {
	return wrapTarget(nil, target.NewStaticLib(name, baseCfg, builder))
}

func NewSharedLib(name string, builder *CppBuilder, baseCfg compiler.Config) *Target {
	return wrapTarget(nil, target.NewSharedLib(name, baseCfg, builder))
}

func NewExe(name string, builder *CppBuilder, baseCfg compiler.Config) *Target {
	return wrapTarget(nil, target.NewExe(name, baseCfg, builder))
}
