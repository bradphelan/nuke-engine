package physics

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.StaticLib().
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "include"))).
		LinkPublic(mathcore.Def).
		Sources(self.Glob("src/**/*.cpp"))
})
