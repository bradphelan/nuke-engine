package engine

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/rule"
)

type countingRule struct {
	callCount int
}

func (r *countingRule) ID() string { return "counting" }
func (r *countingRule) Apply(_ context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	r.callCount++
	var out []artifact.Artifact
	for _, a := range inputs {
		out = append(out, &stubArt{uri: a.URI() + ".out", fp: a.Fingerprint()})
	}
	return out, nil, nil
}

type stubArt struct {
	uri string
	fp  string
}

func (s *stubArt) URI() string            { return s.uri }
func (s *stubArt) Fingerprint() string    { return s.fp }
func (s *stubArt) Metadata() map[string]string { return nil }

func TestEngineResolveCacheHit(t *testing.T) {
	dir := t.TempDir()
	eng, err := New(filepath.Join(dir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer eng.Close()

	r := &countingRule{}
	input := artifact.NewSliceSet([]artifact.Artifact{
		&stubArt{uri: "a", fp: "fp-a"},
	})
	ts := rule.NewTransformSet(input, r)

	ctx := context.Background()

	// First resolve: cache miss
	result1, err := eng.Resolve(ctx, ts)
	if err != nil {
		t.Fatal(err)
	}
	if r.callCount != 1 {
		t.Fatalf("expected 1 call, got %d", r.callCount)
	}
	if len(result1) != 1 {
		t.Fatalf("expected 1 output, got %d", len(result1))
	}

	// Second resolve: cache hit
	result2, err := eng.Resolve(ctx, ts)
	if err != nil {
		t.Fatal(err)
	}
	if r.callCount != 1 {
		t.Fatalf("expected still 1 call (cache hit), got %d", r.callCount)
	}
	if len(result2) != 1 {
		t.Fatalf("expected 1 output from cache, got %d", len(result2))
	}
}
