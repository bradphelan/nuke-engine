package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.StaticLib().
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "include"))).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "src", "internal"))).
		Sources(self.Glob("src/**/*.cpp"))
})
