package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRootFrom_AtRoot(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	root, err := FindProjectRootFrom(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Resolve symlinks for comparison since TempDir may involve symlinks.
	want, _ := filepath.EvalSymlinks(dir)
	if root != want {
		t.Errorf("root = %q, want %q", root, want)
	}
}

func TestFindProjectRootFrom_NestedDir(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(dir, "a", "b", "c")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}

	root, err := FindProjectRootFrom(nested)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want, _ := filepath.EvalSymlinks(dir)
	if root != want {
		t.Errorf("root = %q, want %q", root, want)
	}
}

func TestFindProjectRootFrom_NotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := FindProjectRootFrom(dir)
	if !errors.Is(err, ErrProjectNotFound) {
		t.Errorf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestIsProjectRoot(t *testing.T) {
	dir := t.TempDir()
	if IsProjectRoot(dir) {
		t.Error("empty dir should not be project root")
	}
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte{}, 0644); err != nil {
		t.Fatal(err)
	}
	if !IsProjectRoot(dir) {
		t.Error("dir with runefact.toml should be project root")
	}
}

func TestGetConfigPath(t *testing.T) {
	got := GetConfigPath("/foo/bar")
	want := filepath.Join("/foo/bar", "runefact.toml")
	if got != want {
		t.Errorf("GetConfigPath = %q, want %q", got, want)
	}
}
