package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "mathcore")
	return cpp.NewStaticLib("mathcore", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "src", "internal"))).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
