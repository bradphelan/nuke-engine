package compiler

import (
	"testing"
)

func TestConfigImmutability(t *testing.T) {
	base := New().WithStandard(Cpp20)
	debug := base.WithBuildType(Debug).WithDefine("DEBUG")
	release := base.WithBuildType(Release).WithDefine("NDEBUG")

	if len(base.Defines()) != 0 {
		t.Fatal("base should have no defines")
	}
	if len(debug.Defines()) != 1 || debug.Defines()[0] != "DEBUG" {
		t.Fatalf("debug defines: %v", debug.Defines())
	}
	if len(release.Defines()) != 1 || release.Defines()[0] != "NDEBUG" {
		t.Fatalf("release defines: %v", release.Defines())
	}
}

func TestConfigFluent(t *testing.T) {
	c := New().
		WithStandard(Cpp20).
		WithBuildType(Release).
		WithDefine("WIN32").
		WithDefine("_UNICODE").
		WithIncludeDir("C:/myinc").
		ExportAllSymbols(true)

	if c.Standard() != Cpp20 {
		t.Fatal("standard")
	}
	if c.BuildType() != Release {
		t.Fatal("build type")
	}
	if len(c.Defines()) != 2 {
		t.Fatal("defines")
	}
	if !c.ExportsAllSymbols() {
		t.Fatal("export symbols")
	}
}

func TestConfigDerivation(t *testing.T) {
	base := New().WithStandard(Cpp20).WithDefine("SHARED")
	debug := base.WithBuildType(Debug)
	release := base.WithBuildType(Release)

	debug2 := debug.WithDefine("EXTRA")
	if len(release.Defines()) != 1 {
		t.Fatal("release should still have 1 define")
	}
	if len(debug2.Defines()) != 2 {
		t.Fatal("debug2 should have 2 defines")
	}
}
