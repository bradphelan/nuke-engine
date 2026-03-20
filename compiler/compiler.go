package compiler

// Backend translates abstract Config into concrete tool invocations.
type Backend interface {
	// CompileArgs returns the command + args + env to compile a single source file.
	CompileArgs(cfg Config, srcPath, objPath string) (cmd string, args []string, env []string)

	// StaticLibArgs returns the command + args + env to archive objects into a static lib.
	StaticLibArgs(cfg Config, outputPath string, objPaths []string) (cmd string, args []string, env []string)

	// SharedLibArgs returns the command + args + env to link a shared library/DLL.
	SharedLibArgs(cfg Config, outputPath string, objPaths, libPaths []string) (cmd string, args []string, env []string)

	// ExeArgs returns the command + args + env to link an executable.
	ExeArgs(cfg Config, outputPath string, objPaths, libPaths []string) (cmd string, args []string, env []string)

	// ParseDiscoveredDeps extracts header dependencies from compiler output.
	ParseDiscoveredDeps(output string) []string

	// ID returns a stable identifier for this backend+version (used in cache keys).
	ID() string
}
