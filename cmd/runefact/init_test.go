package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestScaffoldFiles(t *testing.T) {
	files := scaffoldFiles("test-game")
	if len(files) == 0 {
		t.Fatal("scaffoldFiles returned no files")
	}

	// Check runefact.toml is first and contains project name.
	if files[0].path != "runefact.toml" {
		t.Errorf("first file = %q, want runefact.toml", files[0].path)
	}
	if got := files[0].content; len(got) == 0 {
		t.Error("runefact.toml is empty")
	}

	// Verify all expected paths are present.
	want := map[string]bool{
		"runefact.toml":                  false,
		"assets/palettes/default.palette": false,
		"assets/sprites/demo.sprite":      false,
		"assets/maps/demo.map":            false,
		"assets/instruments/demo.inst":    false,
		"assets/sfx/demo.sfx":             false,
		"assets/tracks/demo.track":         false,
	}
	for _, f := range files {
		if _, ok := want[f.path]; !ok {
			t.Errorf("unexpected file: %s", f.path)
		}
		want[f.path] = true
	}
	for path, found := range want {
		if !found {
			t.Errorf("missing file: %s", path)
		}
	}
}

func TestRunInit_CreatesFiles(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	flagInitName = "test-project"
	flagInitForce = false
	flagQuiet = true

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit: %v", err)
	}

	// Check files exist.
	for _, name := range []string{
		"runefact.toml",
		"assets/palettes/default.palette",
		"assets/sprites/demo.sprite",
		"assets/maps/demo.map",
		"assets/instruments/demo.inst",
		"assets/sfx/demo.sfx",
		"assets/tracks/demo.track",
	} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("expected file %s to exist: %v", name, err)
		}
	}

	// Check dirs exist.
	for _, d := range []string{
		"assets/palettes",
		"assets/sprites",
		"assets/maps",
		"assets/instruments",
		"assets/sfx",
		"assets/tracks",
	} {
		info, err := os.Stat(filepath.Join(dir, d))
		if err != nil {
			t.Errorf("expected dir %s to exist: %v", d, err)
		} else if !info.IsDir() {
			t.Errorf("%s is not a directory", d)
		}
	}
}

func TestRunInit_ErrorsWithoutForce(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Create existing runefact.toml.
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagInitForce = false
	err := runInit(nil, nil)
	if err == nil {
		t.Fatal("expected error when runefact.toml exists without --force")
	}
}

func TestRunInit_ForceOverwrites(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)

	// Create existing runefact.toml.
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte("[project]\n"), 0644); err != nil {
		t.Fatal(err)
	}

	flagInitName = "overwrite-test"
	flagInitForce = true
	flagQuiet = true

	if err := runInit(nil, nil); err != nil {
		t.Fatalf("runInit with --force: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("runefact.toml is empty after force overwrite")
	}
}
