package mathcore

import (
	"path/filepath"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/cpp"
)

var Def = cpp.Define(func(ctx *cpp.Context, self cpp.Def) *cpp.Target {
	dir := self.Dir(ctx)
	return self.StaticLib(ctx).
		PublicConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "include"))).
		PrivateConfig(compiler.New().WithIncludeDir(filepath.Join(dir, "src", "internal"))).
		Sources(artifact.Glob(filepath.Join(dir, "src"), "**/*.cpp"))
})
