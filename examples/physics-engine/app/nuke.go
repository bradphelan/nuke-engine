package app

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
	"physics_engine/simulation"
)

var Def = target.Define(func(ctx *cpp.Context, self target.Def[*cpp.Context, compiler.Config]) *target.Target[compiler.Config] {
	dir := self.Dir(ctx.RootDir)
	return cpp.NewExe(self.Name(), ctx.Builder, ctx.Config).
		LinkPublic(mathcore.Def.Resolve(ctx)).
		LinkPrivate(simulation.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
