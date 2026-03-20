package artifact

import (
	"context"
	"testing"
)

func makeSlice(uris ...string) ArtifactSet {
	arts := make([]Artifact, len(uris))
	for i, u := range uris {
		arts[i] = &stubArtifact{uri: u, fp: u}
	}
	return NewSliceSet(arts)
}

type stubArtifact struct {
	uri string
	fp  string
}

func (s *stubArtifact) URI() string                { return s.uri }
func (s *stubArtifact) Fingerprint() string        { return s.fp }
func (s *stubArtifact) Metadata() map[string]string { return nil }

func uris(arts []Artifact) []string {
	out := make([]string, len(arts))
	for i, a := range arts {
		out[i] = a.URI()
	}
	return out
}

func TestSliceSetResolve(t *testing.T) {
	s := makeSlice("a", "b", "c")
	arts, err := s.Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(arts) != 3 {
		t.Fatalf("expected 3, got %d", len(arts))
	}
}

func TestUnion(t *testing.T) {
	a := makeSlice("x", "y")
	b := makeSlice("y", "z")
	result, _ := a.Add(b).Resolve(context.Background())
	if len(result) != 3 {
		t.Fatalf("union should have 3 unique, got %d", len(result))
	}
}

func TestDifference(t *testing.T) {
	a := makeSlice("x", "y", "z")
	b := makeSlice("y")
	result, _ := a.Sub(b).Resolve(context.Background())
	got := uris(result)
	if len(got) != 2 {
		t.Fatalf("diff should have 2, got %v", got)
	}
}

func TestFilter(t *testing.T) {
	s := makeSlice("src/core/a.cpp", "src/core/b.cpp", "src/test/c.cpp")
	result, _ := s.Filter("src/core/**").Resolve(context.Background())
	if len(result) != 2 {
		t.Fatalf("filter should match 2, got %d", len(result))
	}
}

func TestImmutability(t *testing.T) {
	a := makeSlice("x", "y")
	b := makeSlice("z")
	_ = a.Add(b)
	arts, _ := a.Resolve(context.Background())
	if len(arts) != 2 {
		t.Fatal("Add must not mutate the original set")
	}
}
