package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/physics"
	"physics_engine/renderer"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.SharedLib().
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "include")).WithDefine("SIMULATION_EXPORTS")).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "src"))).
		LinkPrivate(physics.Def).
		LinkPrivate(renderer.Def).
		Sources(self.Glob("src/**/*.cpp"))
})
