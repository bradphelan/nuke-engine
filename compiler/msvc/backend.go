package msvc

import (
	"strings"

	"github.com/bradphelan/nuke-engine/compiler"
)

// Backend implements compiler.Backend for MSVC.
type Backend struct {
	tc *Toolchain
}

// NewBackend creates a Backend from a discovered Toolchain.
func NewBackend(tc *Toolchain) *Backend { return &Backend{tc: tc} }

// ID returns a stable identifier for cache keys.
func (b *Backend) ID() string { return "msvc:" + b.tc.Version }

// CompileArgs returns cl.exe invocation for a single translation unit.
func (b *Backend) CompileArgs(cfg compiler.Config, srcPath, objPath string) (string, []string, []string) {
	args := []string{"/nologo", "/c", "/EHsc", "/showIncludes"}

	// Language standard.
	switch cfg.Standard() {
	case compiler.C17:
		args = append(args, "/std:c17")
	case compiler.Cpp17:
		args = append(args, "/std:c++17")
	case compiler.Cpp20:
		args = append(args, "/std:c++20")
	case compiler.Cpp23:
		args = append(args, "/std:c++latest")
	}

	// Build type flags.
	switch cfg.BuildType() {
	case compiler.Debug:
		args = append(args, "/Od", "/Zi", "/MDd", "/RTC1")
	case compiler.Release:
		args = append(args, "/O2", "/DNDEBUG", "/MD")
	case compiler.RelWithDebInfo:
		args = append(args, "/O2", "/Zi", "/DNDEBUG", "/MD")
	}

	// Defines.
	for _, d := range cfg.Defines() {
		args = append(args, "/D"+d)
	}

	// User include dirs.
	for _, dir := range cfg.IncludeDirs() {
		args = append(args, "/I"+dir)
	}

	// Toolchain system include dirs.
	for _, dir := range b.tc.IncludeDirs {
		args = append(args, "/I"+dir)
	}

	// Raw CFlags pass through.
	args = append(args, cfg.RawCFlags()...)

	// Output and source.
	args = append(args, "/Fo"+objPath, srcPath)

	return b.tc.CL, args, b.tc.Environ()
}

// StaticLibArgs returns lib.exe invocation to create a static library.
func (b *Backend) StaticLibArgs(cfg compiler.Config, outputPath string, objPaths []string) (string, []string, []string) {
	args := []string{"/NOLOGO", "/OUT:" + outputPath}
	args = append(args, objPaths...)
	return b.tc.Lib, args, b.tc.Environ()
}

// SharedLibArgs returns link.exe invocation to create a DLL.
func (b *Backend) SharedLibArgs(cfg compiler.Config, outputPath string, objPaths, libPaths []string) (string, []string, []string) {
	args := []string{"/NOLOGO", "/DLL", "/OUT:" + outputPath}

	if cfg.BuildType() == compiler.Debug || cfg.BuildType() == compiler.RelWithDebInfo {
		args = append(args, "/DEBUG")
	}

	for _, dir := range b.tc.LibDirs {
		args = append(args, "/LIBPATH:"+dir)
	}

	args = append(args, cfg.RawLDFlags()...)
	args = append(args, objPaths...)
	args = append(args, libPaths...)

	return b.tc.Link, args, b.tc.Environ()
}

// ExeArgs returns link.exe invocation to create an executable.
func (b *Backend) ExeArgs(cfg compiler.Config, outputPath string, objPaths, libPaths []string) (string, []string, []string) {
	args := []string{"/NOLOGO", "/OUT:" + outputPath}

	if cfg.BuildType() == compiler.Debug || cfg.BuildType() == compiler.RelWithDebInfo {
		args = append(args, "/DEBUG")
	}

	for _, dir := range b.tc.LibDirs {
		args = append(args, "/LIBPATH:"+dir)
	}

	args = append(args, cfg.RawLDFlags()...)
	args = append(args, objPaths...)
	args = append(args, libPaths...)

	return b.tc.Link, args, b.tc.Environ()
}

// ParseDiscoveredDeps extracts header paths from cl.exe /showIncludes output.
// Lines have the form: "Note: including file:   C:\path\to\header.h"
func (b *Backend) ParseDiscoveredDeps(output string) []string {
	const prefix = "Note: including file:"
	var deps []string
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.HasPrefix(line, prefix) {
			dep := strings.TrimSpace(line[len(prefix):])
			if dep != "" {
				deps = append(deps, dep)
			}
		}
	}
	return deps
}

// Verify at compile time that Backend satisfies compiler.Backend.
var _ compiler.Backend = (*Backend)(nil)
