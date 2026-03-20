package compiler

import "slices"

type Standard int

const (
	C17 Standard = iota
	Cpp17
	Cpp20
	Cpp23
)

type BuildType int

const (
	Debug BuildType = iota
	Release
	RelWithDebInfo
)

// Config holds abstract, compiler-agnostic build properties.
// It is a value type — every With* method returns a new independent copy.
type Config struct {
	standard      Standard
	buildType     BuildType
	defines       []string
	includeDirs   []string
	rawCFlags     []string
	rawLDFlags    []string
	exportSymbols bool
}

func New() Config {
	return Config{standard: Cpp17, buildType: Debug}
}

// Getters
func (c Config) Standard() Standard      { return c.standard }
func (c Config) BuildType() BuildType    { return c.buildType }
func (c Config) Defines() []string       { return slices.Clone(c.defines) }
func (c Config) IncludeDirs() []string   { return slices.Clone(c.includeDirs) }
func (c Config) RawCFlags() []string     { return slices.Clone(c.rawCFlags) }
func (c Config) RawLDFlags() []string    { return slices.Clone(c.rawLDFlags) }
func (c Config) ExportsAllSymbols() bool { return c.exportSymbols }

// Fluent builders
func (c Config) WithStandard(s Standard) Config   { c.standard = s; return c }
func (c Config) WithBuildType(bt BuildType) Config { c.buildType = bt; return c }

func (c Config) WithDefine(d string) Config {
	c.defines = append(slices.Clone(c.defines), d)
	return c
}
func (c Config) WithIncludeDir(dir string) Config {
	c.includeDirs = append(slices.Clone(c.includeDirs), dir)
	return c
}
func (c Config) WithCFlags(flags ...string) Config {
	c.rawCFlags = append(slices.Clone(c.rawCFlags), flags...)
	return c
}
func (c Config) WithLDFlags(flags ...string) Config {
	c.rawLDFlags = append(slices.Clone(c.rawLDFlags), flags...)
	return c
}
func (c Config) ExportAllSymbols(b bool) Config {
	c.exportSymbols = b
	return c
}
