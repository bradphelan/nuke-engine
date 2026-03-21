package app

import (
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
	"physics_engine/simulation"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.Exe().
		LinkPublic(mathcore.Def).
		LinkPrivate(simulation.Def).
		Sources(self.Glob("src/**/*.cpp"))
})
