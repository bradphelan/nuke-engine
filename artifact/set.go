package artifact

import (
	"context"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradphelan/nuke-engine/fingerprint"
)

type ArtifactSet interface {
	Resolve(ctx context.Context) ([]Artifact, error)
	Fingerprint(ctx context.Context) (string, error)
	Add(other ArtifactSet) ArtifactSet
	Sub(other ArtifactSet) ArtifactSet
	Filter(pattern string) ArtifactSet
}

// --- SliceSet: in-memory set of artifacts ---
type SliceSet struct {
	items []Artifact
}

func NewSliceSet(items []Artifact) *SliceSet {
	return &SliceSet{items: items}
}

func (s *SliceSet) Resolve(_ context.Context) ([]Artifact, error) { return s.items, nil }
func (s *SliceSet) Fingerprint(ctx context.Context) (string, error) {
	return setFingerprint(ctx, s)
}
func (s *SliceSet) Add(other ArtifactSet) ArtifactSet   { return &UnionSet{left: s, right: other} }
func (s *SliceSet) Sub(other ArtifactSet) ArtifactSet   { return &DiffSet{left: s, right: other} }
func (s *SliceSet) Filter(pattern string) ArtifactSet   { return &FilterSet{source: s, pattern: pattern} }

// --- UnionSet ---
type UnionSet struct{ left, right ArtifactSet }

func (u *UnionSet) Resolve(ctx context.Context) ([]Artifact, error) {
	l, err := u.left.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	r, err := u.right.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool)
	var out []Artifact
	for _, a := range l {
		if !seen[a.URI()] {
			seen[a.URI()] = true
			out = append(out, a)
		}
	}
	for _, a := range r {
		if !seen[a.URI()] {
			seen[a.URI()] = true
			out = append(out, a)
		}
	}
	return out, nil
}
func (u *UnionSet) Fingerprint(ctx context.Context) (string, error) { return setFingerprint(ctx, u) }
func (u *UnionSet) Add(other ArtifactSet) ArtifactSet   { return &UnionSet{left: u, right: other} }
func (u *UnionSet) Sub(other ArtifactSet) ArtifactSet   { return &DiffSet{left: u, right: other} }
func (u *UnionSet) Filter(pattern string) ArtifactSet   { return &FilterSet{source: u, pattern: pattern} }

// --- DiffSet ---
type DiffSet struct{ left, right ArtifactSet }

func (d *DiffSet) Resolve(ctx context.Context) ([]Artifact, error) {
	l, err := d.left.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	r, err := d.right.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	exclude := make(map[string]bool)
	for _, a := range r {
		exclude[a.URI()] = true
	}
	var out []Artifact
	for _, a := range l {
		if !exclude[a.URI()] {
			out = append(out, a)
		}
	}
	return out, nil
}
func (d *DiffSet) Fingerprint(ctx context.Context) (string, error) { return setFingerprint(ctx, d) }
func (d *DiffSet) Add(other ArtifactSet) ArtifactSet   { return &UnionSet{left: d, right: other} }
func (d *DiffSet) Sub(other ArtifactSet) ArtifactSet   { return &DiffSet{left: d, right: other} }
func (d *DiffSet) Filter(pattern string) ArtifactSet   { return &FilterSet{source: d, pattern: pattern} }

// --- FilterSet ---
type FilterSet struct {
	source  ArtifactSet
	pattern string
}

func (f *FilterSet) Resolve(ctx context.Context) ([]Artifact, error) {
	arts, err := f.source.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	var out []Artifact
	for _, a := range arts {
		matched, _ := doublestar.Match(f.pattern, a.URI())
		if matched {
			out = append(out, a)
		}
	}
	return out, nil
}
func (f *FilterSet) Fingerprint(ctx context.Context) (string, error) { return setFingerprint(ctx, f) }
func (f *FilterSet) Add(other ArtifactSet) ArtifactSet   { return &UnionSet{left: f, right: other} }
func (f *FilterSet) Sub(other ArtifactSet) ArtifactSet   { return &DiffSet{left: f, right: other} }
func (f *FilterSet) Filter(pattern string) ArtifactSet   { return &FilterSet{source: f, pattern: pattern} }

// shared helper
func setFingerprint(ctx context.Context, s ArtifactSet) (string, error) {
	arts, err := s.Resolve(ctx)
	if err != nil {
		return "", err
	}
	fps := make([]string, len(arts))
	for i, a := range arts {
		fps[i] = a.Fingerprint()
	}
	return fingerprint.MerkleRoot(fps), nil
}
