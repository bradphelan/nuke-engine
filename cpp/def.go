package cpp

import (
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/target"
)

var (
	registryMu sync.Mutex
	registry   []Def
)

// Def is a C++ facade over the generic target.Def declaration primitive.
type Def struct {
	inner target.Def[*Context, compiler.Config]
}

// Self is a context-bound helper created when a Def is being realized.
// It avoids callback boilerplate by binding a declaration to the current build
// context for the duration of a single build callback.
type Self struct {
	def Def
	ctx *Context
}

// DefBuilder declares one C++ module target using the bound helper.
type DefBuilder func(self Self) *Target

// Define creates a C++ module target declaration.
func Define(build DefBuilder) Def {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("cpp.Define: could not resolve caller")
	}
	pkgDir := filepath.Base(filepath.Dir(file))
	inner := target.DefineNamed(pkgDir, func(ctx *Context, base target.Def[*Context, compiler.Config]) *target.Target[compiler.Config] {
		return build(Self{def: Def{inner: base}, ctx: ctx}).Raw()
	})
	def := Def{inner: inner}
	registerDef(def)
	return def
}

func registerDef(def Def) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry = append(registry, def)
}

// RegisteredDefs returns all declared C++ defs imported into the current
// process. Importing a package containing `var Def = cpp.Define(...)` is
// enough to register it.
func RegisteredDefs() []Def {
	registryMu.Lock()
	defer registryMu.Unlock()
	out := append([]Def(nil), registry...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name() < out[j].Name()
	})
	return out
}

// ResolveRegistered resolves every registered declaration against ctx.
func ResolveRegistered(ctx *Context) []*Target {
	defs := RegisteredDefs()
	resolved := make([]*Target, 0, len(defs))
	for _, def := range defs {
		resolved = append(resolved, def.Resolve(ctx))
	}
	return resolved
}

func (d Def) Name() string { return d.inner.Name() }

// Dir returns the module directory under the current project root.
func (d Def) Dir(ctx *Context) string { return d.inner.Dir(ctx.RootDir) }

// Glob resolves a module-relative glob like "src/**/*.cpp" without requiring
// callers to spell out a base directory and a separate pattern.
func (d Def) Glob(ctx *Context, pattern string) artifact.ArtifactSet {
	baseDir, relPattern := splitModuleGlob(d.Dir(ctx), pattern)
	return artifact.Glob(baseDir, relPattern)
}

// Resolve materializes this target exactly once per context.
func (d Def) Resolve(ctx *Context) *Target { return wrapTarget(ctx, d.inner.Resolve(ctx)) }

// StaticLib creates a static-library target using this Def's canonical name.
func (d Def) StaticLib(ctx *Context) *Target {
	return wrapTarget(ctx, target.NewStaticLib(d.Name(), ctx.Config, ctx.Builder))
}

// SharedLib creates a shared-library target using this Def's canonical name.
func (d Def) SharedLib(ctx *Context) *Target {
	return wrapTarget(ctx, target.NewSharedLib(d.Name(), ctx.Config, ctx.Builder))
}

// Exe creates an executable target using this Def's canonical name.
func (d Def) Exe(ctx *Context) *Target {
	return wrapTarget(ctx, target.NewExe(d.Name(), ctx.Config, ctx.Builder))
}

func (s Self) Name() string { return s.def.Name() }

func (s Self) RootDir() string { return s.ctx.RootDir }

func (s Self) Dir() string { return s.def.Dir(s.ctx) }

func (s Self) Glob(pattern string) artifact.ArtifactSet { return s.def.Glob(s.ctx, pattern) }

func (s Self) Resolve(def Def) *Target { return def.Resolve(s.ctx) }

func (s Self) StaticLib() *Target { return s.def.StaticLib(s.ctx) }

func (s Self) SharedLib() *Target { return s.def.SharedLib(s.ctx) }

func (s Self) Exe() *Target { return s.def.Exe(s.ctx) }

func splitModuleGlob(moduleDir, pattern string) (baseDir string, relPattern string) {
	cleanPattern := filepath.ToSlash(filepath.Clean(pattern))
	if cleanPattern == "." {
		return moduleDir, "*"
	}
	parts := strings.Split(cleanPattern, "/")
	rootCount := 0
	for rootCount < len(parts) && !strings.ContainsAny(parts[rootCount], "*?[]{}") {
		rootCount++
	}
	baseDir = moduleDir
	if rootCount > 0 {
		baseDir = filepath.Join(append([]string{moduleDir}, parts[:rootCount]...)...)
	}
	relParts := parts[rootCount:]
	if len(relParts) == 0 {
		relPattern = "**"
	} else {
		relPattern = strings.Join(relParts, "/")
	}
	return baseDir, relPattern
}
