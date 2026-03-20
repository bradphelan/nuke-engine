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

// Build returns all top-level targets so that nuke-build can present them in
// the interactive TUI and build them individually or as a group.
func Build(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string) []*target.Target[compiler.Config] {
	mc := mathcore.Build(builder, baseCfg, rootDir)
	ph := physics.Build(builder, baseCfg, rootDir)
	rn := renderer.Build(builder, baseCfg, rootDir)
	sm := simulation.Build(builder, baseCfg, rootDir)
	app := AppTarget(builder, baseCfg, rootDir, mc, sm)
	return []*target.Target[compiler.Config]{app, mc, ph, rn, sm}
}

func AppTarget(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, mc, sm *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "app")
	return cpp.NewExe("app", builder, baseCfg).
		LinkPublic(mc).
		LinkPrivate(sm).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
