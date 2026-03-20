package main

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func AppTarget(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, mathcore, simulation *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "app")
	return cpp.NewExe("app", builder, baseCfg).
		LinkPublic(mathcore).
		LinkPrivate(simulation).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
