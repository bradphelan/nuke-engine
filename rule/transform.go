package rule

import (
	"context"

	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/fingerprint"
)

// TransformSet is a lazy ArtifactSet node that wraps a Rule. It resolves its
// input set and applies the rule only when Resolve is called.
type TransformSet struct {
	inputs         artifact.ArtifactSet
	rule           Rule
	lastDiscovered []artifact.Artifact
}

// NewTransformSet creates a TransformSet that applies r to inputs when resolved.
func NewTransformSet(inputs artifact.ArtifactSet, r Rule) *TransformSet {
	return &TransformSet{inputs: inputs, rule: r}
}

// Inputs returns the upstream ArtifactSet this rule reads from.
func (t *TransformSet) Inputs() artifact.ArtifactSet { return t.inputs }

// RuleID returns the identifier of the rule applied by this node.
func (t *TransformSet) RuleID() string { return t.rule.ID() }

// Rule returns the Rule applied by this TransformSet.
func (t *TransformSet) Rule() Rule { return t.rule }

// Resolve resolves the input set, applies the rule, stores discovered dependencies,
// and returns the output artifacts.
func (t *TransformSet) Resolve(ctx context.Context) ([]artifact.Artifact, error) {
	resolved, err := t.inputs.Resolve(ctx)
	if err != nil {
		return nil, err
	}
	outputs, discovered, err := t.rule.Apply(ctx, resolved)
	if err != nil {
		return nil, err
	}
	t.lastDiscovered = discovered
	return outputs, nil
}

// Fingerprint resolves the set and returns a MerkleRoot of the output fingerprints.
func (t *TransformSet) Fingerprint(ctx context.Context) (string, error) {
	arts, err := t.Resolve(ctx)
	if err != nil {
		return "", err
	}
	fps := make([]string, len(arts))
	for i, a := range arts {
		fps[i] = a.Fingerprint()
	}
	return fingerprint.MerkleRoot(fps), nil
}

// Discovered returns the sidecar dependency artifacts found during the last Resolve call.
func (t *TransformSet) Discovered() []artifact.Artifact { return t.lastDiscovered }

// Add returns a UnionSet of this TransformSet and other.
func (t *TransformSet) Add(other artifact.ArtifactSet) artifact.ArtifactSet {
	return artifact.NewSliceSet(nil).Add(t).Add(other)
}

// Sub returns a DiffSet of this TransformSet minus other.
func (t *TransformSet) Sub(other artifact.ArtifactSet) artifact.ArtifactSet {
	return artifact.NewSliceSet(nil).Add(t).Sub(other)
}

// Filter returns a FilterSet over this TransformSet matching pattern.
func (t *TransformSet) Filter(pattern string) artifact.ArtifactSet {
	return artifact.NewSliceSet(nil).Add(t).Filter(pattern)
}
