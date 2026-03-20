package renderer

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"

	"physics_engine/mathcore"
)

func Build(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string) *target.Target[compiler.Config] {
	mc := mathcore.Build(builder, baseCfg, rootDir)
	return Target(builder, baseCfg, rootDir, mc)
}

func Target(builder *cpp.CppBuilder, baseCfg compiler.Config, rootDir string, mc *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	dir := filepath.Join(rootDir, "renderer")
	return cpp.NewSharedLib("renderer", builder, baseCfg).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include")).WithDefine("RENDERER_EXPORTS")).
		LinkPublic(mc).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
}
