package target

import "sync"

type contextEntry[C interface{ Merge(C) C }] struct {
	once sync.Once
	t    *Target[C]
}

// Context owns generic target-graph lifecycle concerns (identity and
// memoization) and is language-agnostic.
type Context[C interface{ Merge(C) C }] struct {
	RootDir string
	mu      sync.Mutex
	cache   map[string]*contextEntry[C]
}

// NewContext creates a new generic target context rooted at rootDir.
func NewContext[C interface{ Merge(C) C }](rootDir string) *Context[C] {
	return &Context[C]{
		RootDir: rootDir,
		cache:   make(map[string]*contextEntry[C]),
	}
}

// Once returns the named target, constructing it via fn on the first call and
// returning the cached result on every subsequent call.
func (c *Context[C]) Once(name string, fn func() *Target[C]) *Target[C] {
	c.mu.Lock()
	entry, ok := c.cache[name]
	if !ok {
		entry = &contextEntry[C]{}
		c.cache[name] = entry
	}
	c.mu.Unlock()

	entry.once.Do(func() {
		entry.t = fn()
	})
	return entry.t
}