package cpp

import (
	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/target"
)

// Context is a thin C++ adapter over the generic target.Context lifecycle.
// It adds C++-specific runtime state (builder + config) used when materializing
// target definitions.
type Context struct {
	Builder *CppBuilder
	Config  compiler.Config
	RootDir string
	graph   *target.Context[compiler.Config]
}

// NewContext creates a new Context.
func NewContext(builder *CppBuilder, cfg compiler.Config, rootDir string) *Context {
	return &Context{
		Builder: builder,
		Config:  cfg,
		RootDir: rootDir,
		graph:   target.NewContext[compiler.Config](rootDir),
	}
}

// Once returns the named target, constructing it via fn on the first call and
// returning the cached result on every subsequent call.
func (c *Context) Once(name string, fn func() *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	return c.graph.Once(name, fn)
}
