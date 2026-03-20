package engine

import (
	"path/filepath"
	"testing"
	"time"
)

func TestCacheStoreAndLookup(t *testing.T) {
	dir := t.TempDir()
	c, err := NewCache(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer c.Close()

	entry := &CacheEntry{
		OutputURIs:     []string{"file:///out/a.obj"},
		OutputFPs:      []string{"abc123"},
		DiscoveredURIs: []string{"file:///inc/h.h"},
		DiscoveredFPs:  []string{"def456"},
		Timestamp:      time.Now(),
	}
	if err := c.Store("key1", entry); err != nil {
		t.Fatal(err)
	}
	got, err := c.Lookup("key1")
	if err != nil {
		t.Fatal(err)
	}
	if got == nil {
		t.Fatal("expected cache hit")
	}
	if got.OutputURIs[0] != "file:///out/a.obj" {
		t.Fatalf("unexpected: %v", got)
	}
}

func TestCacheMiss(t *testing.T) {
	dir := t.TempDir()
	c, _ := NewCache(filepath.Join(dir, "test.db"))
	defer c.Close()
	got, err := c.Lookup("nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Fatal("expected nil for cache miss")
	}
}
