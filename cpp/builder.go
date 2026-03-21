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

func wrapTarget(ctx *Context, raw *target.Target[compiler.Config]) *Target {
	return &Target{ctx: ctx, raw: raw}
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
	ctx *Context
	raw *target.Target[compiler.Config]
}

func (t *Target) Raw() *target.Target[compiler.Config] { return t.raw }

func (t *Target) Name() string { return t.raw.Name() }

func (t *Target) Kind() target.Kind { return t.raw.Kind() }

func (t *Target) PublicCfg() compiler.Config { return t.raw.PublicCfg() }

func (t *Target) PrivateCfg() compiler.Config { return t.raw.PrivateCfg() }

func (t *Target) ResolvedConfig() compiler.Config { return t.raw.ResolvedConfig() }

func (t *Target) PublicConfig(cfg compiler.Config) *Target {
	t.raw.PublicConfig(cfg)
	return t
}

func (t *Target) PrivateConfig(cfg compiler.Config) *Target {
	t.raw.PrivateConfig(cfg)
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
