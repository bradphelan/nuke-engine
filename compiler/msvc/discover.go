package msvc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// Toolchain holds paths and metadata for an installed MSVC toolchain.
type Toolchain struct {
	CL, Lib, Link string
	IncludeDirs   []string
	LibDirs       []string
	Version       string
}

// Discover locates the latest MSVC toolchain using vswhere.exe and the Windows SDK.
func Discover() (*Toolchain, error) {
	vsPath, err := findVSInstallPath()
	if err != nil {
		return nil, fmt.Errorf("msvc discover: %w", err)
	}

	msvcRoot := filepath.Join(vsPath, "VC", "Tools", "MSVC")
	msvcVer, err := latestSubdir(msvcRoot)
	if err != nil {
		return nil, fmt.Errorf("msvc discover: finding MSVC version: %w", err)
	}

	msvcBase := filepath.Join(msvcRoot, msvcVer)
	binDir := filepath.Join(msvcBase, "bin", "Hostx64", "x64")

	tc := &Toolchain{
		CL:      filepath.Join(binDir, "cl.exe"),
		Lib:     filepath.Join(binDir, "lib.exe"),
		Link:    filepath.Join(binDir, "link.exe"),
		Version: msvcVer,
	}

	// MSVC include and lib dirs.
	tc.IncludeDirs = append(tc.IncludeDirs, filepath.Join(msvcBase, "include"))
	tc.LibDirs = append(tc.LibDirs, filepath.Join(msvcBase, "lib", "x64"))

	// Windows SDK dirs.
	sdkRoot := `C:\Program Files (x86)\Windows Kits\10`
	sdkIncRoot := filepath.Join(sdkRoot, "Include")
	sdkVer, err := latestSubdir(sdkIncRoot)
	if err != nil {
		return nil, fmt.Errorf("msvc discover: finding Windows SDK version: %w", err)
	}

	sdkInc := filepath.Join(sdkIncRoot, sdkVer)
	tc.IncludeDirs = append(tc.IncludeDirs,
		filepath.Join(sdkInc, "ucrt"),
		filepath.Join(sdkInc, "um"),
		filepath.Join(sdkInc, "shared"),
	)

	sdkLibBase := filepath.Join(sdkRoot, "Lib", sdkVer)
	tc.LibDirs = append(tc.LibDirs,
		filepath.Join(sdkLibBase, "ucrt", "x64"),
		filepath.Join(sdkLibBase, "um", "x64"),
	)

	return tc, nil
}

// Environ returns environment variables that configure the MSVC toolchain:
// INCLUDE, LIB, and PATH (with the tool bin dir prepended).
func (tc *Toolchain) Environ() []string {
	binDir := filepath.Dir(tc.CL)

	// Start with the current process environment so critical variables
	// (SystemRoot, TMP, TEMP, etc.) are preserved.
	overrides := map[string]string{
		"INCLUDE": strings.Join(tc.IncludeDirs, ";"),
		"LIB":     strings.Join(tc.LibDirs, ";"),
		"PATH":    binDir + ";" + os.Getenv("PATH"),
	}

	var env []string
	seen := make(map[string]bool)
	for _, e := range os.Environ() {
		key, _, _ := strings.Cut(e, "=")
		upper := strings.ToUpper(key)
		if v, ok := overrides[upper]; ok {
			env = append(env, upper+"="+v)
			seen[upper] = true
		} else {
			env = append(env, e)
		}
	}
	for k, v := range overrides {
		if !seen[k] {
			env = append(env, k+"="+v)
		}
	}
	return env
}

// findVSInstallPath runs vswhere.exe to locate the latest Visual Studio installation.
func findVSInstallPath() (string, error) {
	vswhere := `C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`
	out, err := exec.Command(vswhere, "-latest", "-property", "installationPath").Output()
	if err != nil {
		return "", fmt.Errorf("running vswhere: %w", err)
	}
	p := strings.TrimSpace(string(out))
	if p == "" {
		return "", fmt.Errorf("vswhere returned empty path")
	}
	return p, nil
}

// latestSubdir returns the name of the lexicographically last entry in dir.
func latestSubdir(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", err
	}
	var dirs []string
	for _, e := range entries {
		if e.IsDir() {
			dirs = append(dirs, e.Name())
		}
	}
	if len(dirs) == 0 {
		return "", fmt.Errorf("no subdirectories in %s", dir)
	}
	sort.Strings(dirs)
	return dirs[len(dirs)-1], nil
}
