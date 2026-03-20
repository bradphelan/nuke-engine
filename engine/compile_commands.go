package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/bradphelan/nuke-engine/artifact"
	"github.com/bradphelan/nuke-engine/rule"
)

// WriteCompileCommandsJSON writes compile_commands.json in clang/cmake format
// without executing compilation.
func WriteCompileCommandsJSON(outPath string, set artifact.ArtifactSet) error {
	ctx := context.Background()
	collector := &compileCommandsCollector{seen: map[string]struct{}{}}
	if _, err := collectCompileCommands(ctx, set, collector); err != nil {
		return err
	}

	sort.Slice(collector.items, func(i, j int) bool {
		if collector.items[i].File != collector.items[j].File {
			return collector.items[i].File < collector.items[j].File
		}
		return collector.items[i].Command < collector.items[j].Command
	})

	data, err := json.MarshalIndent(collector.items, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(outPath, append(data, '\n'), 0644)
}

type compileCommandsCollector struct {
	items []rule.CompileCommand
	seen  map[string]struct{}
}

func (c *compileCommandsCollector) add(items []rule.CompileCommand) {
	for _, it := range items {
		key := it.File + "\n" + it.Command
		if _, ok := c.seen[key]; ok {
			continue
		}
		c.seen[key] = struct{}{}
		c.items = append(c.items, it)
	}
}

// collectCompileCommands predicts outputs recursively and collects compile
// command entries from rules that provide them.
func collectCompileCommands(ctx context.Context, set artifact.ArtifactSet, c *compileCommandsCollector) ([]artifact.Artifact, error) {
	switch s := set.(type) {
	case *artifact.UnionSet:
		l, err := collectCompileCommands(ctx, s.Left(), c)
		if err != nil {
			return nil, err
		}
		r, err := collectCompileCommands(ctx, s.Right(), c)
		if err != nil {
			return nil, err
		}
		return dedupeArtifacts(append(l, r...)), nil

	case *artifact.DiffSet:
		l, err := collectCompileCommands(ctx, s.Left(), c)
		if err != nil {
			return nil, err
		}
		r, err := collectCompileCommands(ctx, s.Right(), c)
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
		src, err := collectCompileCommands(ctx, s.Source(), c)
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
		inputs, err := collectCompileCommands(ctx, s.Inputs(), c)
		if err != nil {
			return nil, err
		}
		if p, ok := s.Rule().(rule.CompileCommandsProvider); ok {
			c.add(p.CompileCommands(inputs))
		}
		return predictOutputsForTransform(s, inputs), nil

	default:
		return set.Resolve(ctx)
	}
}
