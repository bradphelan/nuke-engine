package msvc

import (
	"os"
	"strings"
	"testing"
)

func hasMSVC() bool {
	_, err := os.Stat(`C:\Program Files (x86)\Microsoft Visual Studio\Installer\vswhere.exe`)
	return err == nil
}

func TestDiscover(t *testing.T) {
	if !hasMSVC() {
		t.Skip("MSVC not installed")
	}

	tc, err := Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	// All tool paths must exist.
	for _, p := range []string{tc.CL, tc.Lib, tc.Link} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("tool path does not exist: %s", p)
		}
	}

	if tc.Version == "" {
		t.Error("Version is empty")
	}

	if len(tc.IncludeDirs) == 0 {
		t.Error("IncludeDirs is empty")
	}

	if len(tc.LibDirs) == 0 {
		t.Error("LibDirs is empty")
	}
}

func TestEnviron(t *testing.T) {
	if !hasMSVC() {
		t.Skip("MSVC not installed")
	}

	tc, err := Discover()
	if err != nil {
		t.Fatalf("Discover() error: %v", err)
	}

	env := tc.Environ()
	hasInclude := false
	for _, e := range env {
		if strings.HasPrefix(e, "INCLUDE=") {
			hasInclude = true
		}
	}
	if !hasInclude {
		t.Error("Environ() missing INCLUDE= entry")
	}
}
