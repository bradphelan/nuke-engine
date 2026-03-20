package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func Target(ctx *cpp.Context) *target.Target[compiler.Config] {
	return ctx.Once("mathcore", func() *target.Target[compiler.Config] {
		dir := filepath.Join(ctx.RootDir, "mathcore")
		return cpp.NewStaticLib("mathcore", ctx.Builder, ctx.Config).
			PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
			PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "src", "internal"))).
			Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
	})
}
