package target

import (
	"path/filepath"
	"runtime"
)

type resolveContext[C interface{ Merge(C) C }] interface {
	Once(name string, fn func() *Target[C]) *Target[C]
}

// DefBuilder declares one module target using a shared build context.
type DefBuilder[Ctx resolveContext[C], C interface{ Merge(C) C }] func(ctx Ctx, self Def[Ctx, C]) *Target[C]

// Def is a package-level target declaration.
// The canonical name is derived from the declaring package directory.
type Def[Ctx resolveContext[C], C interface{ Merge(C) C }] struct {
	name  string
	build DefBuilder[Ctx, C]
}

// Define creates a module target declaration. The target name is derived from
// the caller package directory name, so authors do not repeat it in code.
func Define[Ctx resolveContext[C], C interface{ Merge(C) C }](build DefBuilder[Ctx, C]) Def[Ctx, C] {
	_, file, _, ok := runtime.Caller(1)
	if !ok {
		panic("target.Define: could not resolve caller")
	}
	pkgDir := filepath.Base(filepath.Dir(file))
	return Def[Ctx, C]{name: pkgDir, build: build}
}

func (d Def[Ctx, C]) Name() string { return d.name }

// Dir returns the module directory under the provided project root.
func (d Def[Ctx, C]) Dir(rootDir string) string {
	return filepath.Join(rootDir, d.name)
}

// Resolve materializes this target exactly once per context.
func (d Def[Ctx, C]) Resolve(ctx Ctx) *Target[C] {
	if d.build == nil {
		panic("target.Def.Resolve: nil build function")
	}
	return ctx.Once(d.name, func() *Target[C] {
		return d.build(ctx, d)
	})
}