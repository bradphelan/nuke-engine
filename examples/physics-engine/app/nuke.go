package app

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
	"physics_engine/simulation"
)

var Def = cpp.Define(func(ctx *cpp.Context, self cpp.Def) *cpp.Target {
	dir := self.Dir(ctx)
	return self.Exe(ctx).
		LinkPublic(mathcore.Def.Resolve(ctx)).
		LinkPrivate(simulation.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
