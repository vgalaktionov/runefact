package build

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBuildCache_NeedsRebuild(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("hello"), 0644)

	cache := LoadCache(dir)

	// First check: always needs rebuild (no cache).
	if !cache.NeedsRebuild(file) {
		t.Error("first check should need rebuild")
	}

	// Second check: same content, no rebuild.
	if cache.NeedsRebuild(file) {
		t.Error("unchanged file should not need rebuild")
	}

	// Modify file: needs rebuild again.
	os.WriteFile(file, []byte("world"), 0644)
	if !cache.NeedsRebuild(file) {
		t.Error("modified file should need rebuild")
	}
}

func TestBuildCache_Persistence(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "test.txt")
	os.WriteFile(file, []byte("data"), 0644)

	// Populate cache and save.
	cache1 := LoadCache(dir)
	cache1.NeedsRebuild(file)
	if err := cache1.Save(); err != nil {
		t.Fatalf("saving cache: %v", err)
	}

	// Load fresh cache — file should not need rebuild.
	cache2 := LoadCache(dir)
	if cache2.NeedsRebuild(file) {
		t.Error("cached file should not need rebuild after reload")
	}
}

func TestBuildCache_MissingFile(t *testing.T) {
	dir := t.TempDir()
	cache := LoadCache(dir)

	if !cache.NeedsRebuild(filepath.Join(dir, "nonexistent.txt")) {
		t.Error("missing file should need rebuild")
	}
}
