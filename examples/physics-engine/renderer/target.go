package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, mathcore *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "renderer")
	return cpp.NewSharedLib("renderer", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mathcore).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
