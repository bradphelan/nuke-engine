package artifact

import (
	"context"
	"os"
	"path/filepath"

	"github.com/bmatcuk/doublestar/v4"
)

type GlobSet struct {
	baseDir string
	pattern string
}

func NewGlobSet(baseDir, pattern string) *GlobSet {
	return &GlobSet{baseDir: baseDir, pattern: pattern}
}

func (g *GlobSet) Resolve(_ context.Context) ([]Artifact, error) {
	var arts []Artifact
	err := filepath.WalkDir(g.baseDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(g.baseDir, path)
		rel = filepath.ToSlash(rel)
		matched, _ := doublestar.Match(g.pattern, rel)
		if matched {
			arts = append(arts, NewFileArtifact(path))
		}
		return nil
	})
	return arts, err
}

func (g *GlobSet) Fingerprint(ctx context.Context) (string, error) { return setFingerprint(ctx, g) }
func (g *GlobSet) Add(other ArtifactSet) ArtifactSet               { return &UnionSet{left: g, right: other} }
func (g *GlobSet) Sub(other ArtifactSet) ArtifactSet               { return &DiffSet{left: g, right: other} }
func (g *GlobSet) Filter(pattern string) ArtifactSet               { return &FilterSet{source: g, pattern: pattern} }

// Glob is a convenience constructor
func Glob(baseDir, pattern string) ArtifactSet {
	return NewGlobSet(baseDir, pattern)
}

// File returns a single-file ArtifactSet
func File(path string) ArtifactSet {
	return NewSliceSet([]Artifact{NewFileArtifact(path)})
}
