package simulation

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, physics, renderer *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "simulation")
	return cpp.NewSharedLib("simulation", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("SIMULATION_EXPORTS")).
		LinkPrivate(physics).
		LinkPrivate(renderer).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
