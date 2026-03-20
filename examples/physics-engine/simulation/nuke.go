package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/physics"
	"physics_engine/renderer"
)

var Def = cpp.Define(func(ctx *cpp.Context, self cpp.Def) *cpp.Target {
	dir := self.Dir(ctx)
	return self.SharedLib(ctx).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("SIMULATION_EXPORTS")).
		LinkPrivate(physics.Def.Resolve(ctx)).
		LinkPrivate(renderer.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
