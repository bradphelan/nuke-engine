package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"demo/physics"
	"demo/renderer"
)

func Build(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string) *target.Target[compiler.Config] {
	ph := physics.Build(builder, baseCfg, rootDir)
	rn := renderer.Build(builder, baseCfg, rootDir)
	return Target(builder, baseCfg, rootDir, ph, rn)
}

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, ph, rn *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "simulation")
	return cpp.NewSharedLib("simulation", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("SIMULATION_EXPORTS")).
		LinkPrivate(ph).
		LinkPrivate(rn).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
