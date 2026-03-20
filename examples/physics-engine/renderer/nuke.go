package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"

	"physics_engine/mathcore"
)

var Def = cpp.Define(func(ctx *cpp.Context, self cpp.Def) *cpp.Target {
	dir := self.Dir(ctx)
	return self.SharedLib(ctx).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mathcore.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
