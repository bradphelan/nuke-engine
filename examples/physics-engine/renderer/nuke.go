package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
)

func Target(ctx *cpp.Context) *target.Target[compiler.Config] {
	return ctx.Once("renderer", func() *target.Target[compiler.Config] {
		dir := filepath.Join(ctx.RootDir, "renderer")
		return cpp.NewSharedLib("renderer", ctx.Builder, ctx.Config).
			PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("RENDERER_EXPORTS")).
			LinkPublic(mathcore.Target(ctx)).
			Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
	})
}
