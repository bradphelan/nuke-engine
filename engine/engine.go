package engine

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/fingerprint"
	"github.com/bradphelan/nuke-engine/rule"
)

type Engine struct {
	cache *Cache
	stats BuildStats
}

// BuildStats reports cache and execution behavior for a resolve run.
type BuildStats struct {
	CacheHits     int
	CacheMisses   int
	RulesExecuted int
}

func New(cachePath string) (*Engine, error) {
	c, err := NewCache(cachePath)
	if err != nil {
		return nil, err
	}
	return &Engine{cache: c}, nil
}

func (e *Engine) Close() error { return e.cache.Close() }

// Stats returns the metrics captured during the most recent Resolve call.
func (e *Engine) Stats() BuildStats { return e.stats }

func (e *Engine) Resolve(ctx context.Context, set artifact.ArtifactSet) ([]artifact.Artifact, error) {
	e.stats = BuildStats{}
	return e.resolve(ctx, set)
}

func (e *Engine) resolve(ctx context.Context, set artifact.ArtifactSet) ([]artifact.Artifact, error) {
	if u, ok := set.(*artifact.UnionSet); ok {
		l, err := e.resolve(ctx, u.Left())
		if err != nil {
			return nil, err
		}
		r, err := e.resolve(ctx, u.Right())
		if err != nil {
			return nil, err
		}
		seen := make(map[string]bool)
		out := make([]artifact.Artifact, 0, len(l)+len(r))
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

	if d, ok := set.(*artifact.DiffSet); ok {
		l, err := e.resolve(ctx, d.Left())
		if err != nil {
			return nil, err
		}
		r, err := e.resolve(ctx, d.Right())
		if err != nil {
			return nil, err
		}
		exclude := make(map[string]bool)
		for _, a := range r {
			exclude[a.URI()] = true
		}
		out := make([]artifact.Artifact, 0, len(l))
		for _, a := range l {
			if !exclude[a.URI()] {
				out = append(out, a)
			}
		}
		return out, nil
	}

	if f, ok := set.(*artifact.FilterSet); ok {
		arts, err := e.resolve(ctx, f.Source())
		if err != nil {
			return nil, err
		}
		out := make([]artifact.Artifact, 0, len(arts))
		for _, a := range arts {
			matched, _ := doublestar.Match(f.Pattern(), a.URI())
			if matched {
				out = append(out, a)
			}
		}
		return out, nil
	}

	ts, ok := set.(*rule.TransformSet)
	if !ok {
		return set.Resolve(ctx)
	}

	// 1. Resolve inputs recursively
	inputs, err := e.resolve(ctx, ts.Inputs())
	if err != nil {
		return nil, err
	}

	// 2. Fingerprint inputs
	fps := make([]string, len(inputs))
	for i, a := range inputs {
		fps[i] = a.Fingerprint()
	}
	inputFP := fingerprint.MerkleRoot(fps)

	// 3. Build cache key
	cacheKey := fingerprint.HashBytes([]byte(inputFP + "|" + ts.Rule().ID()))

	// 4. Check cache
	entry, err := e.cache.Lookup(cacheKey)
	if err != nil {
		return nil, fmt.Errorf("cache lookup: %w", err)
	}
	if entry != nil && e.validateEntry(entry) {
		e.stats.CacheHits++
		return e.entryToArtifacts(entry), nil
	}
	e.stats.CacheMisses++

	// 5. Cache miss — execute rule
	outputs, discovered, err := ts.Rule().Apply(ctx, inputs)
	if err != nil {
		return nil, err
	}
	e.stats.RulesExecuted++

	// 6. Store in cache
	newEntry := &CacheEntry{
		OutputURIs: uris(outputs),
		OutputFPs:  fpsFrom(outputs),
	}
	if len(discovered) > 0 {
		newEntry.DiscoveredURIs = uris(discovered)
		newEntry.DiscoveredFPs = fpsFrom(discovered)
	}
	if err := e.cache.Store(cacheKey, newEntry); err != nil {
		return nil, fmt.Errorf("cache store: %w", err)
	}

	return outputs, nil
}

func (e *Engine) validateEntry(entry *CacheEntry) bool {
	for i, uri := range entry.DiscoveredURIs {
		a := artifact.NewFileArtifact(uriToPath(uri))
		if a.Fingerprint() != entry.DiscoveredFPs[i] {
			return false
		}
	}
	return true
}

func (e *Engine) entryToArtifacts(entry *CacheEntry) []artifact.Artifact {
	arts := make([]artifact.Artifact, len(entry.OutputURIs))
	for i, uri := range entry.OutputURIs {
		arts[i] = artifact.NewFileArtifact(uriToPath(uri))
	}
	return arts
}

func uriToPath(uri string) string {
	path := strings.TrimPrefix(uri, "file://")
	return filepath.FromSlash(path)
}

func uris(arts []artifact.Artifact) []string {
	out := make([]string, len(arts))
	for i, a := range arts {
		out[i] = a.URI()
	}
	return out
}

func fpsFrom(arts []artifact.Artifact) []string {
	out := make([]string, len(arts))
	for i, a := range arts {
		out[i] = a.Fingerprint()
	}
	return out
}
