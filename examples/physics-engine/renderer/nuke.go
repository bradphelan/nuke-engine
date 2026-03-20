package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
)

var Def = target.Define(func(ctx *cpp.Context, self target.Def[*cpp.Context, compiler.Config]) *target.Target[compiler.Config] {
	dir := self.Dir(ctx.RootDir)
	return cpp.NewSharedLib(self.Name(), ctx.Builder, ctx.Config).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mathcore.Def.Resolve(ctx)).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
