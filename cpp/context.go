package cpp

import (
	"sync"

	"github.com/bradphelan/nuke-engine/compiler"
	"github.com/bradphelan/nuke-engine/target"
)

// Context holds the build configuration shared across all modules and
// memoizes target construction so each named target is built exactly once
// regardless of how many modules depend on it.
type Context struct {
	Builder *CppBuilder
	Config  compiler.Config
	RootDir string
	mu      sync.Mutex
	cache   map[string]*contextEntry
}

type contextEntry struct {
	once sync.Once
	t    *target.Target[compiler.Config]
}

// NewContext creates a new Context.
func NewContext(builder *CppBuilder, cfg compiler.Config, rootDir string) *Context {
	return &Context{
		Builder: builder,
		Config:  cfg,
		RootDir: rootDir,
		cache:   make(map[string]*contextEntry),
	}
}

// Once returns the named target, constructing it via fn on the first call and
// returning the cached result on every subsequent call.
func (c *Context) Once(name string, fn func() *target.Target[compiler.Config]) *target.Target[compiler.Config] {
	c.mu.Lock()
	entry, ok := c.cache[name]
	if !ok {
		entry = &contextEntry{}
		c.cache[name] = entry
	}
	c.mu.Unlock()

	entry.once.Do(func() {
		entry.t = fn()
	})
	return entry.t
}
