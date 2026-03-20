package rule

import (
	"context"
	"strings"
	"testing"

	"github.com/bradphelan/nuke-engine/artifact"
)

type upperRule struct{}

func (r *upperRule) ID() string { return "upper" }
func (r *upperRule) Apply(_ context.Context, inputs []artifact.Artifact) ([]artifact.Artifact, []artifact.Artifact, error) {
	var out []artifact.Artifact
	for _, a := range inputs {
		out = append(out, &fakeArtifact{
			uri: strings.ToUpper(a.URI()),
			fp:  a.Fingerprint(),
		})
	}
	return out, nil, nil
}

type fakeArtifact struct {
	uri string
	fp  string
}

func (f *fakeArtifact) URI() string                { return f.uri }
func (f *fakeArtifact) Fingerprint() string        { return f.fp }
func (f *fakeArtifact) Metadata() map[string]string { return nil }

func TestTransformSetResolve(t *testing.T) {
	input := artifact.NewSliceSet([]artifact.Artifact{
		&fakeArtifact{uri: "hello", fp: "aaa"},
	})
	ts := NewTransformSet(input, &upperRule{})
	result, err := ts.Resolve(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 1 || result[0].URI() != "HELLO" {
		t.Fatalf("expected HELLO, got %v", result)
	}
}
