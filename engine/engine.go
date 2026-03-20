package engine

import (
	"context"
	"fmt"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/fingerprint"
	"github.com/bradphelan/nuke-engine/rule"
)

type Engine struct {
	cache *Cache
}

func New(cachePath string) (*Engine, error) {
	c, err := NewCache(cachePath)
	if err != nil {
		return nil, err
	}
	return &Engine{cache: c}, nil
}

func (e *Engine) Close() error { return e.cache.Close() }

func (e *Engine) Resolve(ctx context.Context, set artifact.ArtifactSet) ([]artifact.Artifact, error) {
	ts, ok := set.(*rule.TransformSet)
	if !ok {
		return set.Resolve(ctx)
	}

	// 1. Resolve inputs recursively
	inputs, err := e.Resolve(ctx, ts.Inputs())
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
		return e.entryToArtifacts(entry), nil
	}

	// 5. Cache miss — execute rule
	outputs, discovered, err := ts.Rule().Apply(ctx, inputs)
	if err != nil {
		return nil, err
	}

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
		a := artifact.NewFileArtifact(uri)
		if a.Fingerprint() != entry.DiscoveredFPs[i] {
			return false
		}
	}
	return true
}

func (e *Engine) entryToArtifacts(entry *CacheEntry) []artifact.Artifact {
	arts := make([]artifact.Artifact, len(entry.OutputURIs))
	for i, uri := range entry.OutputURIs {
		arts[i] = artifact.NewFileArtifact(uri)
	}
	return arts
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
