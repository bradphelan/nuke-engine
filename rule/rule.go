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

// Namer is an optional interface a Rule may implement to provide a short
// human-readable label for display (e.g. dependency tree output).
type Namer interface {
	Name() string
}

// Predictor is an optional interface a Rule may implement to advertise its
// expected output artifact URIs without executing. Used by the tree printer.
type Predictor interface {
	PredictedOutputURIs() []string
}

// InputPredictor is an optional interface a Rule may implement to advertise
// expected output artifact URIs using resolved input artifacts.
// This lets rules (like compile) return concrete outputs per input file.
type InputPredictor interface {
	PredictedOutputURIsForInputs(inputs []artifact.Artifact) []string
}

// CompileCommand is one entry in compile_commands.json.
type CompileCommand struct {
	Directory string `json:"directory"`
	Command   string `json:"command"`
	File      string `json:"file"`
}

// CompileCommandsProvider is an optional interface a Rule may implement to
// provide compile_commands.json entries for a given resolved input set.
type CompileCommandsProvider interface {
	CompileCommands(inputs []artifact.Artifact) []CompileCommand
}
