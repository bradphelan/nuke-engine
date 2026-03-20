package artifact

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/bradphelan/nuke-engine/fingerprint"
)

func TestFileArtifact_URI(t *testing.T) {
	a := NewFileArtifact("/src/main.cpp")
	want := "file:///src/main.cpp"
	if got := a.URI(); got != want {
		t.Fatalf("URI() = %q, want %q", got, want)
	}
}

func TestFileArtifact_URI_BackslashNormalised(t *testing.T) {
	a := NewFileArtifact(`C:\src\main.cpp`)
	want := "file://C:/src/main.cpp"
	if got := a.URI(); got != want {
		t.Fatalf("URI() = %q, want %q", got, want)
	}
}

func TestFileArtifact_Fingerprint_Stable(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "hello.txt")
	if err := os.WriteFile(p, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}
	a := NewFileArtifact(p)
	h1 := a.Fingerprint()
	h2 := a.Fingerprint()
	if h1 != h2 {
		t.Fatalf("fingerprint not stable: %s vs %s", h1, h2)
	}
	want := fingerprint.HashBytes([]byte("hello"))
	if h1 != want {
		t.Fatalf("fingerprint = %s, want %s", h1, want)
	}
}

func TestFileArtifact_Fingerprint_Lazy(t *testing.T) {
	// For a missing file the fingerprint falls back to hashing the path.
	a := NewFileArtifact("/no/such/file.txt")
	got := a.Fingerprint()
	want := fingerprint.HashBytes([]byte("/no/such/file.txt"))
	if got != want {
		t.Fatalf("fallback fingerprint = %s, want %s", got, want)
	}
}

func TestFileArtifact_Metadata_NonNil(t *testing.T) {
	a := NewFileArtifact("/a.txt")
	m := a.Metadata()
	if m == nil {
		t.Fatal("Metadata() returned nil")
	}
}

func TestFileArtifact_Metadata_CopyOnRead(t *testing.T) {
	a := NewFileArtifact("/a.txt")
	m := a.Metadata()
	m["injected"] = "value"
	if _, ok := a.Metadata()["injected"]; ok {
		t.Fatal("external mutation leaked into artifact metadata")
	}
}

func TestFileArtifact_WithMetadata_Immutability(t *testing.T) {
	a := NewFileArtifact("/a.txt")
	b := a.WithMetadata("lang", "cpp")

	if _, ok := a.Metadata()["lang"]; ok {
		t.Fatal("WithMetadata mutated the original artifact")
	}
	if v := b.Metadata()["lang"]; v != "cpp" {
		t.Fatalf("new artifact metadata[lang] = %q, want %q", v, "cpp")
	}
}

// Verify FileArtifact satisfies the Artifact interface at compile time.
var _ Artifact = (*FileArtifact)(nil)
