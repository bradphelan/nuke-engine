package engine

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/rule"
)

// PrintTree prints the actual artifact dependency DAG as an ASCII tree.
// Labels are artifact URIs. The tree is rendered from final outputs down to
// their input dependencies.
func PrintTree(w io.Writer, set artifact.ArtifactSet) {
	ctx := context.Background()
	d := &artifactDAG{
		nodes:   map[string]struct{}{},
		edges:   map[string]struct{}{},
		reverse: map[string]map[string]struct{}{},
		forward: map[string]map[string]struct{}{},
	}

	outputs, err := collectArtifactDAG(ctx, set, d)
	if err != nil {
		fmt.Fprintf(w, "failed to build DAG: %v\n", err)
		return
	}

	roots := make([]string, 0, len(outputs))
	seen := map[string]bool{}
	for _, a := range outputs {
		uri := a.URI()
		if seen[uri] {
			continue
		}
		seen[uri] = true
		roots = append(roots, uri)
	}
	sort.Strings(roots)

	for i, root := range roots {
		last := i == len(roots)-1
		printASCIINode(w, d, root, "", last)
	}
}

// PrintTreeGraphviz prints the actual artifact dependency DAG as Graphviz DOT.
func PrintTreeGraphviz(w io.Writer, set artifact.ArtifactSet) {
	ctx := context.Background()
	d := &artifactDAG{
		nodes:   map[string]struct{}{},
		edges:   map[string]struct{}{},
		reverse: map[string]map[string]struct{}{},
		forward: map[string]map[string]struct{}{},
	}

	outputs, err := collectArtifactDAG(ctx, set, d)
	if err != nil {
		fmt.Fprintf(w, "// failed to build DAG: %v\n", err)
		return
	}
	for _, a := range outputs {
		d.nodes[a.URI()] = struct{}{}
	}

	writeDOT(w, d)
}

type artifactDAG struct {
	nodes   map[string]struct{}
	edges   map[string]struct{}            // key: from + "\n" + to
	reverse map[string]map[string]struct{} // to -> set(from)
	forward map[string]map[string]struct{} // from -> set(to)
}

func collectArtifactDAG(ctx context.Context, set artifact.ArtifactSet, d *artifactDAG) ([]artifact.Artifact, error) {
	switch s := set.(type) {
	case *artifact.UnionSet:
		l, err := collectArtifactDAG(ctx, s.Left(), d)
		if err != nil {
			return nil, err
		}
		r, err := collectArtifactDAG(ctx, s.Right(), d)
		if err != nil {
			return nil, err
		}
		return dedupeArtifacts(append(l, r...)), nil

	case *artifact.DiffSet:
		l, err := collectArtifactDAG(ctx, s.Left(), d)
		if err != nil {
			return nil, err
		}
		r, err := collectArtifactDAG(ctx, s.Right(), d)
		if err != nil {
			return nil, err
		}
		exclude := make(map[string]bool, len(r))
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

	case *artifact.FilterSet:
		src, err := collectArtifactDAG(ctx, s.Source(), d)
		if err != nil {
			return nil, err
		}
		out := make([]artifact.Artifact, 0, len(src))
		for _, a := range src {
			matched, _ := doublestar.Match(s.Pattern(), a.URI())
			if matched {
				out = append(out, a)
			}
		}
		return out, nil

	case *rule.TransformSet:
		inputs, err := collectArtifactDAG(ctx, s.Inputs(), d)
		if err != nil {
			return nil, err
		}

		outputs := predictOutputsForTransform(s, inputs)
		for _, in := range inputs {
			d.nodes[in.URI()] = struct{}{}
		}
		for _, out := range outputs {
			d.nodes[out.URI()] = struct{}{}
		}

		// For input-aware rules (e.g. compile), outputs map 1:1 with inputs.
		if _, ok := s.Rule().(rule.InputPredictor); ok && len(inputs) == len(outputs) {
			for i := range inputs {
				d.addEdge(inputs[i].URI(), outputs[i].URI())
			}
			return outputs, nil
		}

		// For aggregate rules (e.g. staticlib/sharedlib/exe), dependencies are
		// input -> each produced output artifact.
		for _, in := range inputs {
			for _, out := range outputs {
				d.addEdge(in.URI(), out.URI())
			}
		}
		return outputs, nil

	default:
		arts, err := set.Resolve(ctx)
		if err != nil {
			return nil, err
		}
		for _, a := range arts {
			d.nodes[a.URI()] = struct{}{}
		}
		return arts, nil
	}
}

func (d *artifactDAG) addEdge(from, to string) {
	d.edges[from+"\n"+to] = struct{}{}
	if _, ok := d.reverse[to]; !ok {
		d.reverse[to] = map[string]struct{}{}
	}
	d.reverse[to][from] = struct{}{}
	if _, ok := d.forward[from]; !ok {
		d.forward[from] = map[string]struct{}{}
	}
	d.forward[from][to] = struct{}{}
}

func predictOutputsForTransform(ts *rule.TransformSet, inputs []artifact.Artifact) []artifact.Artifact {
	if p, ok := ts.Rule().(rule.InputPredictor); ok {
		return urisToArtifacts(p.PredictedOutputURIsForInputs(inputs))
	}
	if p, ok := ts.Rule().(rule.Predictor); ok {
		return urisToArtifacts(p.PredictedOutputURIs())
	}
	return nil
}

func urisToArtifacts(uris []string) []artifact.Artifact {
	out := make([]artifact.Artifact, 0, len(uris))
	for _, uri := range uris {
		out = append(out, artifact.NewFileArtifact(uriToPath(uri)))
	}
	return out
}

func dedupeArtifacts(in []artifact.Artifact) []artifact.Artifact {
	seen := map[string]bool{}
	out := make([]artifact.Artifact, 0, len(in))
	for _, a := range in {
		if seen[a.URI()] {
			continue
		}
		seen[a.URI()] = true
		out = append(out, a)
	}
	return out
}

func writeDOT(w io.Writer, d *artifactDAG) {
	nodes := make([]string, 0, len(d.nodes))
	for uri := range d.nodes {
		nodes = append(nodes, displayPath(uri))
	}
	sort.Strings(nodes)

	edges := make([]string, 0, len(d.edges))
	for key := range d.edges {
		edges = append(edges, key)
	}
	sort.Strings(edges)

	fmt.Fprintln(w, "digraph ArtifactDAG {")
	fmt.Fprintln(w, "  rankdir=LR;")
	for _, path := range nodes {
		q := dotQuote(path)
		fmt.Fprintf(w, "  %s [label=%s];\n", q, q)
	}
	for _, key := range edges {
		parts := strings.SplitN(key, "\n", 2)
		if len(parts) != 2 {
			continue
		}
		fmt.Fprintf(w, "  %s -> %s;\n", dotQuote(displayPath(parts[0])), dotQuote(displayPath(parts[1])))
	}
	fmt.Fprintln(w, "}")
}

func printASCIINode(w io.Writer, d *artifactDAG, uri, prefix string, last bool) {
	g := treeGlyphs()
	branch := g.tee
	nextPrefix := prefix + g.vert
	if last {
		branch = g.last
		nextPrefix = prefix + g.space
	}
	fmt.Fprintf(w, "%s%s%s\n", prefix, branch, displayLabel(uri, d))

	parentsSet := d.reverse[uri]
	if len(parentsSet) == 0 {
		return
	}

	parents := make([]string, 0, len(parentsSet))
	for p := range parentsSet {
		parents = append(parents, p)
	}
	sort.Strings(parents)
	for i, p := range parents {
		printASCIINode(w, d, p, nextPrefix, i == len(parents)-1)
	}
}

func dotQuote(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\\\"") + "\""
}

func displayPath(uri string) string {
	return filepath.ToSlash(uriToPath(uri))
}

func displayLabel(uri string, d *artifactDAG) string {
	p := displayPath(uri)
	t := targetNameFromPath(p)
	if t == "" {
		return p
	}
	return p + " [target:" + t + "]"
}

func targetNameFromPath(p string) string {
	if i := strings.Index(p, "/build/obj/"); i >= 0 {
		rest := p[i+len("/build/obj/"):]
		parts := strings.Split(rest, "/")
		if len(parts) > 0 && parts[0] != "" {
			return parts[0]
		}
	}

	if i := strings.Index(p, "/build/bin/"); i >= 0 {
		base := path.Base(p)
		ext := path.Ext(base)
		return strings.TrimSuffix(base, ext)
	}

	if i := strings.Index(p, "/build/lib/"); i >= 0 {
		base := path.Base(p)
		ext := path.Ext(base)
		name := strings.TrimSuffix(base, ext)
		if (ext == ".a" || ext == ".so" || ext == ".dylib") && strings.HasPrefix(name, "lib") {
			name = strings.TrimPrefix(name, "lib")
		}
		return name
	}

	return ""
}

type glyphSet struct {
	tee   string
	last  string
	vert  string
	space string
}

func treeGlyphs() glyphSet {
	// Force ASCII when requested or in very basic terminals.
	if os.Getenv("NO_UNICODE_TREE") != "" || os.Getenv("TERM") == "dumb" {
		return glyphSet{tee: "|-- ", last: "\\-- ", vert: "|   ", space: "    "}
	}
	return glyphSet{tee: "├── ", last: "└── ", vert: "│   ", space: "    "}
}
