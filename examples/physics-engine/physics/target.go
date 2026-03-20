package physics

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, mathcore *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "physics")
	return cpp.NewStaticLib("physics", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
		LinkPublic(mathcore).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
