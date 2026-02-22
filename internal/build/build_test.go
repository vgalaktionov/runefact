package build

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/config"
)

// setupDemoProject creates a minimal project for testing.
func setupDemoProject(t *testing.T) (string, *config.ProjectConfig) {
	t.Helper()
	dir := t.TempDir()

	// Create project config.
	cfgContent := `[project]
name = "test"
output = "build/assets"
package = "assets"

[defaults]
sprite_size = 8
sample_rate = 44100
bit_depth = 16
`
	os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte(cfgContent), 0644)

	// Create asset directories.
	for _, d := range []string{
		"assets/palettes", "assets/sprites", "assets/maps",
		"assets/instruments", "assets/sfx", "assets/tracks",
	} {
		os.MkdirAll(filepath.Join(dir, d), 0755)
	}

	// Palette.
	os.WriteFile(filepath.Join(dir, "assets/palettes/default.palette"), []byte(`name = "default"
[colors]
_ = "transparent"
r = "#ff0000"
g = "#00ff00"
b = "#0000ff"
k = "#000000"
`), 0644)

	// Sprite.
	os.WriteFile(filepath.Join(dir, "assets/sprites/demo.sprite"), []byte(`palette = "default"
grid = 2

[sprite.dot]
pixels = """
r_
_r
"""
`), 0644)

	// Map.
	os.WriteFile(filepath.Join(dir, "assets/maps/demo.map"), []byte(`tile_size = 2
[tileset]
D = "demo:dot"
_ = ""
[layer.main]
pixels = """
D_
_D
"""
`), 0644)

	// Instrument.
	os.WriteFile(filepath.Join(dir, "assets/instruments/demo.inst"), []byte(`name = "demo"
[oscillator]
waveform = "sine"
[envelope]
attack = 0
decay = 0
sustain = 1
release = 0.01
`), 0644)

	// SFX.
	os.WriteFile(filepath.Join(dir, "assets/sfx/demo.sfx"), []byte(`duration = 0.05
volume = 0.5
[[voice]]
waveform = "sine"
[voice.envelope]
sustain = 0.5
release = 0.01
[voice.pitch]
start = 440
end = 440
`), 0644)

	// Track.
	os.WriteFile(filepath.Join(dir, "assets/tracks/demo.track"), []byte(`tempo = 120
ticks_per_beat = 4
[[channel]]
name = "m"
instrument = "demo"
volume = 0.5
[pattern.p]
ticks = 2
data = """
m
C4
---
"""
[song]
sequence = ["p"]
`), 0644)

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	return dir, cfg
}

func TestBuild_FullProject(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	result := Build(Options{}, cfg, dir)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatalf("build had %d error(s)", len(result.Errors))
	}

	// Check artifacts were created.
	expectedFiles := []string{
		"build/assets/sprites/demo.png",
		"build/assets/maps/demo.json",
		"build/assets/audio/demo.wav", // sfx
		"build/assets/manifest.go",
	}
	for _, f := range expectedFiles {
		path := filepath.Join(dir, f)
		if _, err := os.Stat(path); err != nil {
			t.Errorf("expected artifact %s: %v", f, err)
		}
	}

	if result.ManifestPath == "" {
		t.Error("manifest path not set")
	}
}

func TestBuild_SpritesOnly(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	result := Build(Options{Scope: ScopeSprites}, cfg, dir)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
	}

	// Sprite should exist.
	if _, err := os.Stat(filepath.Join(dir, "build/assets/sprites/demo.png")); err != nil {
		t.Error("expected sprite PNG")
	}
	// Map should NOT exist.
	if _, err := os.Stat(filepath.Join(dir, "build/assets/maps/demo.json")); err == nil {
		t.Error("map should not be built with --sprites scope")
	}
}

func TestBuild_MapsOnly(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	result := Build(Options{Scope: ScopeMaps}, cfg, dir)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
	}

	if _, err := os.Stat(filepath.Join(dir, "build/assets/maps/demo.json")); err != nil {
		t.Error("expected map JSON")
	}
}

func TestValidate_Valid(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	result := Validate(Options{}, cfg, dir)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatal("validation should pass for valid project")
	}
}

func TestValidate_InvalidSprite(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	// Overwrite sprite with invalid content.
	os.WriteFile(filepath.Join(dir, "assets/sprites/demo.sprite"), []byte(`palette = "default"
grid = 2
[sprite.bad]
pixels = """
abc
de
"""
`), 0644)

	result := Validate(Options{}, cfg, dir)

	if len(result.Errors) == 0 {
		t.Fatal("expected validation errors for ragged sprite")
	}
}

func TestValidate_NoOutputCreated(t *testing.T) {
	dir, cfg := setupDemoProject(t)

	Validate(Options{}, cfg, dir)

	// Ensure no build output was created.
	if _, err := os.Stat(filepath.Join(dir, "build")); err == nil {
		t.Error("validate should not create build directory")
	}
}
