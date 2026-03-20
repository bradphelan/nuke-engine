package cpp

import (
	"path/filepath"
	"runtime"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/target"
)

// Target is the concrete target type used by C++ module definitions.
type Target = target.Target[compiler.Config]

// DefBuilder declares one module target using the shared build Context.
type DefBuilder func(ctx *Context, self Def) *Target

// Def is a package-level target declaration.
// The canonical name is derived from the declaring package directory.
type Def struct {
	name  string
	build DefBuilder
}

// Define creates a module target declaration. The target name is derived from
// the caller package directory name, so authors do not repeat it in code.
func Define(build DefBuilder) Def {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("cpp.Define: could not resolve caller")
	}
	pkgDir := filepath.Base(filepath.Dir(file))
	return Def{name: pkgDir, build: build}
}

func (d Def) Name() string { return d.name }

// Dir returns the module directory under the current project root.
func (d Def) Dir(ctx *Context) string {
	return filepath.Join(ctx.RootDir, d.name)
}

// Resolve materializes this target exactly once per Context.
func (d Def) Resolve(ctx *Context) *Target {
	if d.build == nil {
		panic("cpp.Def.Resolve: nil build function")
	}
	return ctx.Once(d.name, func() *Target {
		return d.build(ctx, d)
	})
}

// StaticLib creates a static-library target using this Def's canonical name.
func (d Def) StaticLib(ctx *Context) *Target {
	return NewStaticLib(d.name, ctx.Builder, ctx.Config)
}

// SharedLib creates a shared-library target using this Def's canonical name.
func (d Def) SharedLib(ctx *Context) *Target {
	return NewSharedLib(d.name, ctx.Builder, ctx.Config)
}

// Exe creates an executable target using this Def's canonical name.
func (d Def) Exe(ctx *Context) *Target {
	return NewExe(d.name, ctx.Builder, ctx.Config)
}
