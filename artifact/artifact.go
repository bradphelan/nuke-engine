package artifact

import (
	"path/filepath"
	"sync"

	"github.com/bradphelan/nuke-engine/fingerprint"
)

// Artifact represents a build input or output with identity and content hash.
type Artifact interface {
	URI() string
	Fingerprint() string
	Metadata() map[string]string
}

// FileArtifact is an Artifact backed by a file on disk.
type FileArtifact struct {
	path     string
	metadata map[string]string

	fpOnce sync.Once
	fpVal  string
}

// NewFileArtifact creates a FileArtifact with the given path normalised to
// forward slashes.
func NewFileArtifact(path string) *FileArtifact {
	return &FileArtifact{
		path:     filepath.ToSlash(path),
		metadata: make(map[string]string),
	}
}

// URI returns the file URI for this artifact.
func (f *FileArtifact) URI() string {
	return "file://" + f.path
}

// Fingerprint returns the SHA-256 hex digest of the file contents. The value
// is computed lazily (sync.Once). If reading the file fails, the path itself
// is hashed as a fallback.
func (f *FileArtifact) Fingerprint() string {
	f.fpOnce.Do(func() {
		h, err := fingerprint.HashFile(f.path)
		if err != nil {
			h = fingerprint.HashBytes([]byte(f.path))
		}
		f.fpVal = h
	})
	return f.fpVal
}

// Metadata returns a copy of the metadata map (copy-on-read for immutability).
func (f *FileArtifact) Metadata() map[string]string {
	out := make(map[string]string, len(f.metadata))
	for k, v := range f.metadata {
		out[k] = v
	}
	return out
}

// WithMetadata returns a new FileArtifact with the given key-value pair added
// to the metadata. The original FileArtifact is not modified.
func (f *FileArtifact) WithMetadata(key, value string) *FileArtifact {
	newMeta := make(map[string]string, len(f.metadata)+1)
	for k, v := range f.metadata {
		newMeta[k] = v
	}
	newMeta[key] = value
	return &FileArtifact{
		path:     f.path,
		metadata: newMeta,
	}
}
