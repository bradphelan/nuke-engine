package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/physics"
	"physics_engine/renderer"
)

func Target(ctx *cpp.Context) *target.Target[compiler.Config] {
	return ctx.Once("simulation", func() *target.Target[compiler.Config] {
		dir := filepath.Join(ctx.RootDir, "simulation")
		return cpp.NewSharedLib("simulation", ctx.Builder, ctx.Config).
			PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("SIMULATION_EXPORTS")).
			LinkPrivate(physics.Target(ctx)).
			LinkPrivate(renderer.Target(ctx)).
			Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
	})
}
