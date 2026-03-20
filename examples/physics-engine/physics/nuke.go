package physics

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
)

func Target(ctx *cpp.Context) *target.Target[compiler.Config] {
	return ctx.Once("physics", func() *target.Target[compiler.Config] {
		dir := filepath.Join(ctx.RootDir, "physics")
		return cpp.NewStaticLib("physics", ctx.Builder, ctx.Config).
			PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
			LinkPublic(mathcore.Target(ctx)).
			Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
	})
}
