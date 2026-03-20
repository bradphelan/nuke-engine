package rule

import (
	"context"

	"github.com/bradphelan/nuke-engine/artifact"
)

// Rule represents a pure transformation from input artifacts to output artifacts.
// Apply returns outputs, optionally discovered sidecar dependencies (e.g., C++ headers),
// and any error encountered during the transformation.
type Rule interface {
	ID() string
	Apply(ctx context.Context, inputs []artifact.Artifact) (outputs []artifact.Artifact, discovered []artifact.Artifact, err error)
}
