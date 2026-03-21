package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
)

var Def = cpp.Define(func(self cpp.Self) *cpp.Target {
	return self.SharedLib().
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(self.Dir(), "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mathcore.Def).
		Sources(self.Glob("src/**/*.cpp"))
})
