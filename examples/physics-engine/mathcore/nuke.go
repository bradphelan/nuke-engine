package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
	"github.com/bradphelan/nuke-engine/target"
)

var Def = target.Define(func(ctx *cpp.Context, self target.Def[*cpp.Context, compiler.Config]) *target.Target[compiler.Config] {
	dir := self.Dir(ctx.RootDir)
	return cpp.NewStaticLib(self.Name(), ctx.Builder, ctx.Config).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "src", "internal"))).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
