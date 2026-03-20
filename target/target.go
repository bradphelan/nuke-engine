package target

import "github.com/bradphelan/nuke-engine/artifact"

// Kind enumerates the types of build targets.
type Kind int

const (
	StaticLibrary Kind = iota
	SharedLibrary
	Executable
)

// Builder abstracts how targets compile and link. C++ layer implements this.
type Builder[C any] interface {
	Compile(name string, cfg C, sources artifact.ArtifactSet) artifact.ArtifactSet
	StaticLib(name string, cfg C, inputs artifact.ArtifactSet) artifact.ArtifactSet
	SharedLib(name string, cfg C, inputs artifact.ArtifactSet) artifact.ArtifactSet
	Exe(name string, cfg C, inputs artifact.ArtifactSet) artifact.ArtifactSet
	EmptyConfig() C
}

type dependency[C interface{ Merge(C) C }] struct {
	target *Target[C]
	public bool
}

// Target represents a build target parameterised by a configuration type C.
// C must support merging so that transitive configuration propagation works.
type Target[C interface{ Merge(C) C }] struct {
	name       string
	kind       Kind
	baseCfg    C
	publicCfg  C
	privateCfg C
	sources    artifact.ArtifactSet
	deps       []dependency[C]
	builder    Builder[C]
}

func newTarget[C interface{ Merge(C) C }](name string, kind Kind, baseCfg C, builder Builder[C]) *Target[C] {
	empty := builder.EmptyConfig()
	return &Target[C]{
		name:       name,
		kind:       kind,
		baseCfg:    baseCfg,
		publicCfg:  empty,
		privateCfg: empty,
		builder:    builder,
	}
}

// NewStaticLib creates a new static-library target.
func NewStaticLib[C interface{ Merge(C) C }](name string, baseCfg C, builder Builder[C]) *Target[C] {
	return newTarget(name, StaticLibrary, baseCfg, builder)
}

// NewSharedLib creates a new shared-library target.
func NewSharedLib[C interface{ Merge(C) C }](name string, baseCfg C, builder Builder[C]) *Target[C] {
	return newTarget(name, SharedLibrary, baseCfg, builder)
}

// NewExe creates a new executable target.
func NewExe[C interface{ Merge(C) C }](name string, baseCfg C, builder Builder[C]) *Target[C] {
	return newTarget(name, Executable, baseCfg, builder)
}

// Name returns the target's name.
func (t *Target[C]) Name() string { return t.name }

// Kind returns the target's kind.
func (t *Target[C]) Kind() Kind { return t.kind }

// PublicCfg returns the public configuration accumulated on this target.
func (t *Target[C]) PublicCfg() C { return t.publicCfg }

// PrivateCfg returns the private configuration accumulated on this target.
func (t *Target[C]) PrivateCfg() C { return t.privateCfg }

// PublicConfig merges additional public configuration into the target.
func (t *Target[C]) PublicConfig(cfg C) *Target[C] {
	t.publicCfg = t.publicCfg.Merge(cfg)
	return t
}

// PrivateConfig merges additional private configuration into the target.
func (t *Target[C]) PrivateConfig(cfg C) *Target[C] {
	t.privateCfg = t.privateCfg.Merge(cfg)
	return t
}

// Sources sets the source artifact set for this target.
func (t *Target[C]) Sources(s artifact.ArtifactSet) *Target[C] {
	t.sources = s
	return t
}

// LinkPublic adds a public dependency. Its transitive public config will be
// visible to consumers of this target.
func (t *Target[C]) LinkPublic(d *Target[C]) *Target[C] {
	t.deps = append(t.deps, dependency[C]{target: d, public: true})
	return t
}

// LinkPrivate adds a private dependency. Its transitive public config is used
// when compiling this target but is NOT propagated to consumers.
func (t *Target[C]) LinkPrivate(d *Target[C]) *Target[C] {
	t.deps = append(t.deps, dependency[C]{target: d, public: false})
	return t
}

// TransitivePublicConfig returns the merged public configuration that
// consumers of this target inherit. It includes this target's own public
// config plus the transitive public config of all PUBLIC dependencies.
func (t *Target[C]) TransitivePublicConfig() C {
	cfg := t.publicCfg
	for _, d := range t.deps {
		if d.public {
			cfg = cfg.Merge(d.target.TransitivePublicConfig())
		}
	}
	return cfg
}

// ResolvedConfig returns the full configuration used to compile this target's
// own sources: baseCfg + publicCfg + privateCfg + transitive public configs
// from all (public and private) dependencies.
func (t *Target[C]) ResolvedConfig() C {
	cfg := t.baseCfg
	cfg = cfg.Merge(t.publicCfg)
	cfg = cfg.Merge(t.privateCfg)
	for _, d := range t.deps {
		cfg = cfg.Merge(d.target.TransitivePublicConfig())
	}
	return cfg
}

// ArtifactSet compiles sources, collects dependency artifacts, and links
// everything into the final output artifact set for this target.
func (t *Target[C]) ArtifactSet() artifact.ArtifactSet {
	cfg := t.ResolvedConfig()
	objects := t.builder.Compile(t.name, cfg, t.sources)

	var linkInputs artifact.ArtifactSet = objects
	for _, d := range t.deps {
		linkInputs = linkInputs.Add(d.target.ArtifactSet())
	}

	switch t.kind {
	case StaticLibrary:
		return t.builder.StaticLib(t.name, cfg, linkInputs)
	case SharedLibrary:
		return t.builder.SharedLib(t.name, cfg, linkInputs)
	case Executable:
		return t.builder.Exe(t.name, cfg, linkInputs)
	}
	panic("unknown target kind")
}

// Collect returns the unique target graph reachable from root in preorder.
func Collect[C interface{ Merge(C) C }](root *Target[C]) []*Target[C] {
	if root == nil {
		return nil
	}
	seen := make(map[*Target[C]]struct{})
	var out []*Target[C]
	var visit func(*Target[C])
	visit = func(t *Target[C]) {
		if _, ok := seen[t]; ok {
			return
		}
		seen[t] = struct{}{}
		out = append(out, t)
		for _, dep := range t.deps {
			visit(dep.target)
		}
	}
	visit(root)
	return out
}
