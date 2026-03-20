package compiler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/fingerprint"
	"github.com/bradphelan/nuke-engine/rule"
)

// URIToPath strips the "file://" prefix and converts slashes to native path separators.
func URIToPath(uri string) string {
	path := strings.TrimPrefix(uri, "file://")
	return filepath.FromSlash(path)
}

// CompileRule compiles each input source file to an object file using a Backend.
type CompileRule struct {
	backend  Backend
	config   Config
	name     string
	buildDir string
	id       string
}

func NewCompileRule(b Backend, cfg Config, name, buildDir string) *CompileRule {
	raw := fmt.Sprintf("%s:compile:%s:%v", b.ID(), name, cfg)
	return &CompileRule{
		backend:  b,
		config:   cfg,
		name:     name,
		buildDir: buildDir,
		id:       fingerprint.HashBytes([]byte(raw)),
	}
}

func (r *CompileRule) ID() string   { return r.id }
func (r *CompileRule) Name() string { return "compile:" + r.name }
func (r *CompileRule) PredictedOutputURIs() []string {
	return []string{"file://" + filepath.ToSlash(filepath.Join(r.buildDir, "obj", r.name, "*.obj"))}
}
func (r *CompileRule) PredictedOutputURIsForInputs(inputs []artifact.Artifact) []string {
	out := make([]string, 0, len(inputs))
	for _, input := range inputs {
		srcPath := URIToPath(input.URI())
		stem := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		objPath := filepath.Join(r.buildDir, "obj", r.name, stem+".obj")
		out = append(out, "file://"+filepath.ToSlash(objPath))
	}
	return out
}

func (r *CompileRule) CompileCommands(inputs []artifact.Artifact) []rule.CompileCommand {
	out := make([]rule.CompileCommand, 0, len(inputs))
	for _, input := range inputs {
		srcPath := URIToPath(input.URI())
		stem := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		objPath := filepath.Join(r.buildDir, "obj", r.name, stem+".obj")

		cmd, args, _ := r.backend.CompileArgs(r.config, srcPath, objPath)
		argv := append([]string{cmd}, args...)
		out = append(out, rule.CompileCommand{
			Directory: filepath.ToSlash(filepath.Dir(srcPath)),
			Command:   joinCommand(argv),
			File:      filepath.ToSlash(srcPath),
		})
	}
	return out
}

func (r *CompileRule) Apply(ctx context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	objDir := filepath.Join(r.buildDir, "obj", r.name)
	os.MkdirAll(objDir, 0755)

	var outputs []artifact.Artifact
	allDiscovered := make(map[string]bool)

	for _, input := range inputs {
		srcPath := URIToPath(input.URI())
		stem := strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
		objPath := filepath.Join(objDir, stem+".obj")

		cmd, args, env := r.backend.CompileArgs(r.config, srcPath, objPath)
		c := exec.CommandContext(ctx, cmd, args...)
		c.Env = env

		out, err := c.CombinedOutput()
		if err != nil {
			return nil, nil, fmt.Errorf("compile failed for %s:\n%s\n%w", srcPath, string(out), err)
		}

		outputs = append(outputs, artifact.NewFileArtifact(objPath))
		for _, dep := range r.backend.ParseDiscoveredDeps(string(out)) {
			allDiscovered[dep] = true
		}
	}

	var discovered []artifact.Artifact
	for dep := range allDiscovered {
		discovered = append(discovered, artifact.NewFileArtifact(dep))
	}
	return outputs, discovered, nil
}

func joinCommand(argv []string) string {
	if len(argv) == 0 {
		return ""
	}
	parts := make([]string, len(argv))
	for i, a := range argv {
		parts[i] = shellQuote(a)
	}
	return strings.Join(parts, " ")
}

func shellQuote(s string) string {
	if s == "" {
		return "\"\""
	}
	if !strings.ContainsAny(s, " \t\n\r\"") {
		return s
	}
	return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
}

// StaticLibRule archives input object files into a static library using a Backend.
type StaticLibRule struct {
	backend  Backend
	config   Config
	name     string
	buildDir string
	id       string
}

func NewStaticLibRule(b Backend, cfg Config, name, buildDir string) *StaticLibRule {
	raw := fmt.Sprintf("%s:staticlib:%s", b.ID(), name)
	return &StaticLibRule{
		backend:  b,
		config:   cfg,
		name:     name,
		buildDir: buildDir,
		id:       fingerprint.HashBytes([]byte(raw)),
	}
}

func (r *StaticLibRule) ID() string   { return r.id }
func (r *StaticLibRule) Name() string { return "staticlib:" + r.name }
func (r *StaticLibRule) PredictedOutputURIs() []string {
	return []string{"file://" + filepath.ToSlash(filepath.Join(r.buildDir, "lib", r.name+".lib"))}
}

func (r *StaticLibRule) Apply(ctx context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	libDir := filepath.Join(r.buildDir, "lib")
	os.MkdirAll(libDir, 0755)
	libPath := filepath.Join(libDir, r.name+".lib")

	var objPaths []string
	for _, obj := range inputs {
		objPaths = append(objPaths, URIToPath(obj.URI()))
	}

	cmd, args, env := r.backend.StaticLibArgs(r.config, libPath, objPaths)
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = env

	out, err := c.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("static lib failed:\n%s\n%w", string(out), err)
	}

	return []artifact.Artifact{artifact.NewFileArtifact(libPath)}, nil, nil
}

// SharedLibRule links input objects and libraries into a shared library (DLL) using a Backend.
type SharedLibRule struct {
	backend  Backend
	config   Config
	name     string
	buildDir string
	id       string
}

func NewSharedLibRule(b Backend, cfg Config, name, buildDir string) *SharedLibRule {
	raw := fmt.Sprintf("%s:sharedlib:%s", b.ID(), name)
	return &SharedLibRule{
		backend:  b,
		config:   cfg,
		name:     name,
		buildDir: buildDir,
		id:       fingerprint.HashBytes([]byte(raw)),
	}
}

func (r *SharedLibRule) ID() string   { return r.id }
func (r *SharedLibRule) Name() string { return "sharedlib:" + r.name }
func (r *SharedLibRule) PredictedOutputURIs() []string {
	// Downstream link steps depend on the import library. The DLL is a side effect.
	return []string{"file://" + filepath.ToSlash(filepath.Join(r.buildDir, "lib", r.name+".lib"))}
}

func (r *SharedLibRule) Apply(ctx context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	binDir := filepath.Join(r.buildDir, "bin")
	libDir := filepath.Join(r.buildDir, "lib")
	os.MkdirAll(binDir, 0755)
	os.MkdirAll(libDir, 0755)
	dllPath := filepath.Join(binDir, r.name+".dll")
	impLibPath := filepath.Join(libDir, r.name+".lib")

	var objPaths, libPaths []string
	for _, a := range inputs {
		p := URIToPath(a.URI())
		if strings.HasSuffix(strings.ToLower(p), ".lib") {
			libPaths = append(libPaths, p)
		} else {
			objPaths = append(objPaths, p)
		}
	}

	cmd, args, env := r.backend.SharedLibArgs(r.config, dllPath, objPaths, libPaths)
	// Tell the linker to place the import library in the lib directory.
	args = append(args, "/IMPLIB:"+impLibPath)
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = env

	out, err := c.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("shared lib failed:\n%s\n%w", string(out), err)
	}

	// Return the import library as the primary artifact so downstream
	// targets can link against it. The DLL is a side-effect on disk.
	return []artifact.Artifact{artifact.NewFileArtifact(impLibPath)}, nil, nil
}

// ExeRule links input objects and libraries into an executable using a Backend.
type ExeRule struct {
	backend  Backend
	config   Config
	name     string
	buildDir string
	id       string
}

func NewExeRule(b Backend, cfg Config, name, buildDir string) *ExeRule {
	raw := fmt.Sprintf("%s:exe:%s", b.ID(), name)
	return &ExeRule{
		backend:  b,
		config:   cfg,
		name:     name,
		buildDir: buildDir,
		id:       fingerprint.HashBytes([]byte(raw)),
	}
}

func (r *ExeRule) ID() string   { return r.id }
func (r *ExeRule) Name() string { return "exe:" + r.name }
func (r *ExeRule) PredictedOutputURIs() []string {
	return []string{"file://" + filepath.ToSlash(filepath.Join(r.buildDir, "bin", r.name+".exe"))}
}

func (r *ExeRule) Apply(ctx context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	binDir := filepath.Join(r.buildDir, "bin")
	os.MkdirAll(binDir, 0755)
	exePath := filepath.Join(binDir, r.name+".exe")

	var objPaths, libPaths []string
	for _, a := range inputs {
		p := URIToPath(a.URI())
		if strings.HasSuffix(strings.ToLower(p), ".lib") {
			libPaths = append(libPaths, p)
		} else {
			objPaths = append(objPaths, p)
		}
	}

	cmd, args, env := r.backend.ExeArgs(r.config, exePath, objPaths, libPaths)
	c := exec.CommandContext(ctx, cmd, args...)
	c.Env = env

	out, err := c.CombinedOutput()
	if err != nil {
		return nil, nil, fmt.Errorf("link failed:\n%s\n%w", string(out), err)
	}

	return []artifact.Artifact{artifact.NewFileArtifact(exePath)}, nil, nil
}
