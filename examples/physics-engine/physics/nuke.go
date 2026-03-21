package physics

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.StaticLib().
		WithIncludeDir(filepath.Join(self.Dir(), "include")).
		Private().
		WithIncludeDir(filepath.Join(self.Dir(), "src")).
		LinkPublic(mathcore.Def).
		Sources(self.Glob("src/**/*.cpp"))
})
