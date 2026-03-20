package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/physics"
	"physics_engine/renderer"
)

var Def = target.Define(func(ctx *cpp.Context, self target.Def[*cpp.Context, compiler.Config]) *target.Target[compiler.Config] {
	dir := self.Dir(ctx.RootDir)
	return cpp.NewSharedLib(self.Name(), ctx.Builder, ctx.Config).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("SIMULATION_EXPORTS")).
		LinkPrivate(physics.Def.Resolve(ctx)).
		LinkPrivate(renderer.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
