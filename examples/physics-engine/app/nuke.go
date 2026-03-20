package app

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
	"physics_engine/physics"
	"physics_engine/renderer"
	"physics_engine/simulation"
)

func Target(ctx *cpp.Context) *target.Target[compiler.Config] {
	return ctx.Once("app", func() *target.Target[compiler.Config] {
		dir := filepath.Join(ctx.RootDir, "app")
		return cpp.NewExe("app", ctx.Builder, ctx.Config).
			LinkPublic(mathcore.Target(ctx)).
			LinkPrivate(simulation.Target(ctx)).
			Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
	})
}

// Build returns all targets for TUI enumeration. The DAG self-assembles via
// ctx.Once so each target is constructed exactly once regardless of call order.
func Build(ctx *cpp.Context) []*target.Target[compiler.Config] {
	return []*target.Target[compiler.Config]{
		Target(ctx),
		mathcore.Target(ctx),
		physics.Target(ctx),
		renderer.Target(ctx),
		simulation.Target(ctx),
	}
}
