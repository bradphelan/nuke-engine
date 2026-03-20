package cpp

import (
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/compiler"
)

func TestCppBuilderEmptyConfig(t *testing.T) {
	b := NewBuilder(nil, "/build")
	cfg := b.EmptyConfig()
	if len(cfg.IncludeDirs()) != 0 {
		t.Fatal("should be empty includes")
	}
	if len(cfg.Defines()) != 0 {
		t.Fatal("should be empty defines")
	}
}

func TestConvenienceConstructors(t *testing.T) {
	b := NewBuilder(nil, "/build")
	cfg := compiler.New()

	sl := NewStaticLib("foo", b, cfg)
	if sl.Name() != "foo" {
		t.Fatal("name")
	}

	dl := NewSharedLib("bar", b, cfg)
	if dl.Name() != "bar" {
		t.Fatal("name")
	}

	exe := NewExe("app", b, cfg)
	if exe.Name() != "app" {
		t.Fatal("name")
	}
}

func TestCppTargetWithConfig(t *testing.T) {
	b := NewBuilder(nil, "/build")
	cfg := compiler.New().WithStandard(compiler.Cpp20)

	tgt := NewStaticLib("mc", b, cfg).
		PublicConfig(compiler.New().WithIncludeDir("/mc/include")).
		PrivateConfig(compiler.New().WithIncludeDir("/mc/internal")).
		Sources(artifact.NewSliceSet(nil))

	resolved := tgt.ResolvedConfig()
	incs := resolved.IncludeDirs()

	found := false
	for _, inc := range incs {
		if inc == "/mc/include" {
			found = true
		}
	}
	if !found {
		t.Fatal("public include missing from resolved config")
	}

	found = false
	for _, inc := range incs {
		if inc == "/mc/internal" {
			found = true
		}
	}
	if !found {
		t.Fatal("private include missing from resolved config")
	}

	if resolved.Standard() != compiler.Cpp20 {
		t.Fatal("base config standard lost")
	}
}
