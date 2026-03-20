package target

import (
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
)

type mockConfig struct {
	includes []string
	defines  []string
}

func emptyMock() mockConfig { return mockConfig{} }

func (c mockConfig) Merge(other mockConfig) mockConfig {
	return mockConfig{
		includes: append(append([]string{}, c.includes...), other.includes...),
		defines:  append(append([]string{}, c.defines...), other.defines...),
	}
}

func (c mockConfig) WithInclude(dir string) mockConfig {
	c.includes = append(append([]string{}, c.includes...), dir)
	return c
}

func (c mockConfig) WithDefine(d string) mockConfig {
	c.defines = append(append([]string{}, c.defines...), d)
	return c
}

type mockBuilder struct{}

func (m *mockBuilder) Compile(name string, cfg mockConfig, sources artifact.ArtifactSet) artifact.ArtifactSet {
	return sources
}
func (m *mockBuilder) StaticLib(name string, cfg mockConfig, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return inputs
}
func (m *mockBuilder) SharedLib(name string, cfg mockConfig, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return inputs
}
func (m *mockBuilder) Exe(name string, cfg mockConfig, inputs artifact.ArtifactSet) artifact.ArtifactSet {
	return inputs
}
func (m *mockBuilder) EmptyConfig() mockConfig { return emptyMock() }

func TestNewTarget(t *testing.T) {
	tgt := NewStaticLib("mathcore", emptyMock(), &mockBuilder{})
	if tgt.Name() != "mathcore" {
		t.Fatal("name")
	}
	if tgt.Kind() != StaticLibrary {
		t.Fatal("kind")
	}
}

func TestPublicPrivateConfig(t *testing.T) {
	tgt := NewStaticLib("mc", emptyMock(), &mockBuilder{}).
		PublicConfig(emptyMock().WithInclude("/pub")).
		PrivateConfig(emptyMock().WithInclude("/priv"))
	if len(tgt.PublicCfg().includes) != 1 {
		t.Fatal("public")
	}
	if len(tgt.PrivateCfg().includes) != 1 {
		t.Fatal("private")
	}
}

func TestTransitivePublicConfig(t *testing.T) {
	mb := &mockBuilder{}
	mathcore := NewStaticLib("mathcore", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/mathcore/include")).
		PrivateConfig(emptyMock().WithInclude("/mathcore/internal"))
	physics := NewStaticLib("physics", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/physics/include")).
		LinkPublic(mathcore)
	simulation := NewSharedLib("simulation", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/simulation/include")).
		LinkPrivate(physics)

	// mathcore: just own public
	mcPub := mathcore.TransitivePublicConfig()
	assertHas(t, mcPub.includes, "/mathcore/include")
	assertNotHas(t, mcPub.includes, "/mathcore/internal")

	// physics: own + mathcore (PUBLIC dep)
	phPub := physics.TransitivePublicConfig()
	assertHas(t, phPub.includes, "/physics/include")
	assertHas(t, phPub.includes, "/mathcore/include")

	// simulation: only own (physics was PRIVATE)
	simPub := simulation.TransitivePublicConfig()
	assertHas(t, simPub.includes, "/simulation/include")
	assertNotHas(t, simPub.includes, "/physics/include")
	assertNotHas(t, simPub.includes, "/mathcore/include")
}

func TestResolvedConfig(t *testing.T) {
	mb := &mockBuilder{}
	mathcore := NewStaticLib("mathcore", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/mathcore/include")).
		PrivateConfig(emptyMock().WithInclude("/mathcore/internal"))
	physics := NewStaticLib("physics", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/physics/include")).
		PrivateConfig(emptyMock().WithInclude("/physics/internal")).
		LinkPublic(mathcore)

	resolved := physics.ResolvedConfig()
	assertHas(t, resolved.includes, "/physics/include")
	assertHas(t, resolved.includes, "/physics/internal")
	assertHas(t, resolved.includes, "/mathcore/include")
	assertNotHas(t, resolved.includes, "/mathcore/internal")
}

func TestPrivateChainDoesNotLeak(t *testing.T) {
	mb := &mockBuilder{}
	mathcore := NewStaticLib("mathcore", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/mathcore/include"))
	physics := NewStaticLib("physics", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/physics/include")).
		LinkPublic(mathcore)
	simulation := NewSharedLib("simulation", emptyMock(), mb).
		PublicConfig(emptyMock().WithInclude("/simulation/include")).
		LinkPrivate(physics)
	app := NewExe("app", emptyMock(), mb).
		LinkPrivate(simulation)

	resolved := app.ResolvedConfig()
	assertHas(t, resolved.includes, "/simulation/include")
	assertNotHas(t, resolved.includes, "/physics/include")
	assertNotHas(t, resolved.includes, "/mathcore/include")
}

func TestArtifactSetNotNil(t *testing.T) {
	tgt := NewStaticLib("mc", emptyMock(), &mockBuilder{}).
		Sources(artifact.NewSliceSet(nil))
	if tgt.ArtifactSet() == nil {
		t.Fatal("nil")
	}
}

func TestCollectDeduplicatesGraph(t *testing.T) {
	mb := &mockBuilder{}
	mathcore := NewStaticLib("mathcore", emptyMock(), mb)
	physics := NewStaticLib("physics", emptyMock(), mb).
		LinkPublic(mathcore)
	renderer := NewSharedLib("renderer", emptyMock(), mb).
		LinkPublic(mathcore)
	simulation := NewSharedLib("simulation", emptyMock(), mb).
		LinkPrivate(physics).
		LinkPrivate(renderer)
	app := NewExe("app", emptyMock(), mb).
		LinkPublic(mathcore).
		LinkPrivate(simulation)

	collected := Collect(app)
	if len(collected) != 5 {
		t.Fatalf("expected 5 unique targets, got %d", len(collected))
	}
	assertTargetNames(t, collected, []string{"app", "mathcore", "simulation", "physics", "renderer"})
}

func assertTargetNames(t *testing.T, targets []*Target[mockConfig], want []string) {
	t.Helper()
	if len(targets) != len(want) {
		t.Fatalf("expected %d targets, got %d", len(want), len(targets))
	}
	for i, target := range targets {
		if target.Name() != want[i] {
			t.Fatalf("target[%d] = %q, want %q", i, target.Name(), want[i])
		}
	}
}

func assertHas(t *testing.T, s []string, item string) {
	t.Helper()
	for _, v := range s {
		if v == item {
			return
		}
	}
	t.Fatalf("expected %q in %v", item, s)
}

func assertNotHas(t *testing.T, s []string, item string) {
	t.Helper()
	for _, v := range s {
		if v == item {
			t.Fatalf("unexpected %q in %v", item, s)
		}
	}
}
