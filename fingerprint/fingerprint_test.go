package fingerprint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashBytes_Deterministic(t *testing.T) {
	data := []byte("hello world")
	h1 := HashBytes(data)
	h2 := HashBytes(data)
	if h1 != h2 {
		t.Fatalf("same input produced different hashes: %s vs %s", h1, h2)
	}
}

func TestHashBytes_DifferentInput(t *testing.T) {
	h1 := HashBytes([]byte("hello"))
	h2 := HashBytes([]byte("world"))
	if h1 == h2 {
		t.Fatal("different inputs produced the same hash")
	}
}

func TestHashBytes_HexLength(t *testing.T) {
	h := HashBytes([]byte("test"))
	if len(h) != 64 {
		t.Fatalf("expected 64-char hex string, got %d chars", len(h))
	}
}

func TestHashFile_MatchesHashBytes(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "test.txt")
	content := []byte("file content for hashing")
	if err := os.WriteFile(p, content, 0644); err != nil {
		t.Fatal(err)
	}
	got, err := HashFile(p)
	if err != nil {
		t.Fatal(err)
	}
	want := HashBytes(content)
	if got != want {
		t.Fatalf("HashFile = %s, want %s", got, want)
	}
}

func TestHashFile_ErrorOnMissing(t *testing.T) {
	_, err := HashFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestMerkleRoot_OrderIndependent(t *testing.T) {
	a := []string{"aaa", "bbb", "ccc"}
	b := []string{"ccc", "aaa", "bbb"}
	if MerkleRoot(a) != MerkleRoot(b) {
		t.Fatal("MerkleRoot is not order-independent")
	}
}

func TestMerkleRoot_DifferentInputs(t *testing.T) {
	a := MerkleRoot([]string{"aaa", "bbb"})
	b := MerkleRoot([]string{"aaa", "ccc"})
	if a == b {
		t.Fatal("different inputs produced the same MerkleRoot")
	}
}

func TestMerkleRoot_Empty(t *testing.T) {
	got := MerkleRoot([]string{})
	want := HashBytes([]byte(""))
	if got != want {
		t.Fatalf("MerkleRoot([]) = %s, want %s", got, want)
	}
}
