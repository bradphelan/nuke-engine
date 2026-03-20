package engine

import (
	"context"
	"fmt"
	"os"

	"github.com/bradphelan/nuke-engine/artifact"
)

// CleanArtifactSet removes all predicted build outputs for the given artifact
// set. Only nodes that are produced by a build rule (i.e., that have at least
// one incoming edge in the dependency graph) are deleted. Source files, which
// are leaves in the graph with no incoming edges, are never touched.
func CleanArtifactSet(set artifact.ArtifactSet) {
	ctx := context.Background()
	d := &artifactDAG{
		nodes:   map[string]struct{}{},
		edges:   map[string]struct{}{},
		reverse: map[string]map[string]struct{}{},
		forward: map[string]map[string]struct{}{},
	}

	if _, err := collectArtifactDAG(ctx, set, d); err != nil {
		fmt.Fprintf(os.Stderr, "clean: %v\n", err)
		return
	}

	for uri := range d.nodes {
		// Skip source files: they have no incoming edges (nothing produces them).
		if len(d.reverse[uri]) == 0 {
			continue
		}
		p := uriToPath(uri)
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "clean: remove %s: %v\n", p, err)
		}
	}
}
