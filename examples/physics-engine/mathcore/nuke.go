package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/cpp"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.StaticLib().
		WithIncludeDir(filepath.Join(self.Dir(), "include")).
		Private().
		WithIncludeDir(filepath.Join(self.Dir(), "src", "internal")).
		Sources(self.Glob("src/**/*.cpp"))
})
