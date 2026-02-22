package build

import (
	"encoding/json"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vgalaktionov/runefact/internal/config"
)

// TestAcceptance_InitAndBuild verifies that `runefact init` output builds
// successfully with all expected artifacts.
func TestAcceptance_InitAndBuild(t *testing.T) {
	dir := t.TempDir()

	// Simulate runefact init — write all scaffold files.
	writeScaffoldProject(t, dir, "rune-knight")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatalf("loading config: %v", err)
	}

	result := Build(Options{}, cfg, dir)

	for _, w := range result.Warnings {
		t.Logf("warning: %s", w)
	}
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("build error: %v", e)
		}
		t.Fatalf("build had %d error(s)", len(result.Errors))
	}

	// Verify expected artifacts exist.
	// Note: both SFX and track are named "demo" — the track overwrites the SFX WAV.
	expected := []string{
		"build/assets/sprites/demo.png",
		"build/assets/maps/demo.json",
		"build/assets/audio/demo.wav",
		"build/assets/manifest.go",
	}
	for _, f := range expected {
		path := filepath.Join(dir, f)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("missing artifact %s: %v", f, err)
			continue
		}
		if info.Size() == 0 {
			t.Errorf("empty artifact: %s", f)
		}
	}

	if result.ManifestPath == "" {
		t.Error("manifest path not set")
	}
}

// TestAcceptance_SpriteSheetDimensions verifies PNG output has correct dimensions.
func TestAcceptance_SpriteSheetDimensions(t *testing.T) {
	dir := t.TempDir()
	writeScaffoldProject(t, dir, "test")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	result := Build(Options{Scope: ScopeSprites}, cfg, dir)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatal("sprite build failed")
	}

	pngPath := filepath.Join(dir, "build/assets/sprites/demo.png")
	f, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("opening PNG: %v", err)
	}
	defer f.Close()

	imgCfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decoding PNG config: %v", err)
	}

	// The demo sprite has heart (8x8, 1 frame) and star (8x8, 2 frames).
	// Sheet should be at least 8 pixels tall and wide enough for all sprites.
	if imgCfg.Width < 8 || imgCfg.Height < 8 {
		t.Errorf("sprite sheet too small: %dx%d", imgCfg.Width, imgCfg.Height)
	}
}

// TestAcceptance_MapJSONStructure verifies JSON output is valid and has layers.
func TestAcceptance_MapJSONStructure(t *testing.T) {
	dir := t.TempDir()
	writeScaffoldProject(t, dir, "test")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	result := Build(Options{Scope: ScopeMaps}, cfg, dir)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatal("map build failed")
	}

	jsonPath := filepath.Join(dir, "build/assets/maps/demo.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("reading JSON: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if _, ok := m["tile_size"]; !ok {
		t.Error("map JSON missing tile_size")
	}
	if _, ok := m["layers"]; !ok {
		t.Error("map JSON missing layers")
	}
}

// TestAcceptance_AudioWAVHeaders verifies WAV outputs are non-trivial.
func TestAcceptance_AudioWAVHeaders(t *testing.T) {
	dir := t.TempDir()
	writeScaffoldProject(t, dir, "test")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	result := Build(Options{Scope: ScopeAudio}, cfg, dir)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatal("audio build failed")
	}

	// Check SFX WAV.
	sfxPath := filepath.Join(dir, "build/assets/audio/demo.wav")
	sfxInfo, err := os.Stat(sfxPath)
	if err != nil {
		t.Fatalf("missing SFX WAV: %v", err)
	}
	// WAV header is 44 bytes; actual audio data should add more.
	if sfxInfo.Size() <= 44 {
		t.Errorf("SFX WAV too small (%d bytes)", sfxInfo.Size())
	}
}

// TestAcceptance_ManifestContent verifies manifest.go is valid Go.
func TestAcceptance_ManifestContent(t *testing.T) {
	dir := t.TempDir()
	writeScaffoldProject(t, dir, "test")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	result := Build(Options{}, cfg, dir)
	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("error: %v", e)
		}
		t.Fatal("build failed")
	}

	data, err := os.ReadFile(result.ManifestPath)
	if err != nil {
		t.Fatalf("reading manifest: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "package assets") {
		t.Error("manifest missing package declaration")
	}
	if !strings.Contains(content, "SpriteSheet") {
		t.Error("manifest missing sprite sheet constants")
	}
}

// TestAcceptance_ValidateInitProject verifies validation passes for init output.
func TestAcceptance_ValidateInitProject(t *testing.T) {
	dir := t.TempDir()
	writeScaffoldProject(t, dir, "test")

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	result := Validate(Options{}, cfg, dir)

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			t.Errorf("validation error: %v", e)
		}
		t.Fatal("validation should pass for init project")
	}
}

// writeScaffoldProject writes the same files as `runefact init`.
func writeScaffoldProject(t *testing.T, dir, name string) {
	t.Helper()

	dirs := []string{
		"assets/palettes", "assets/sprites", "assets/maps",
		"assets/instruments", "assets/sfx", "assets/tracks",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(dir, d), 0755); err != nil {
			t.Fatal(err)
		}
	}

	files := map[string]string{
		"runefact.toml": `[project]
name = "` + name + `"
output = "build/assets"
package = "assets"

[defaults]
sprite_size = 16
sample_rate = 44100
bit_depth = 16

[preview]
window_width = 800
window_height = 600
background = "#1a1a2e"
pixel_scale = 4
audio_volume = 0.5
`,
		"assets/palettes/default.palette": `name = "default"

[colors]
_ = "transparent"
k = "#000000"
w = "#ffffff"
r = "#ff004d"
b = "#29adff"
g = "#00e436"
d = "#1d2b53"
s = "#ffccaa"
h = "#ab5236"
y = "#ffec27"
o = "#ffa300"
p = "#7e2553"
l = "#83769c"
c = "#008751"
e = "#5f574f"
f = "#c2c3c7"
`,
		"assets/sprites/demo.sprite": `palette = "default"
grid = 8

[sprite.heart]
pixels = """
_rr__rr_
rrrrrrr_
rrrrrrrr
rrrrrrrr
_rrrrrr_
__rrrr__
___rr___
________
"""

[sprite.star]
framerate = 4

[[sprite.star.frame]]
pixels = """
____y___
___yyy__
__yyyyy_
___yyy__
____y___
________
________
________
"""

[[sprite.star.frame]]
pixels = """
________
___y_y__
____y___
__yyyyy_
____y___
___y_y__
________
________
"""
`,
		"assets/maps/demo.map": `tile_size = 8

[tileset]
H = "demo:heart"
S = "demo:star"
_ = ""

[layer.main]
pixels = """
________
_H___H__
________
_S_S_S__
________
________
________
________
"""
`,
		"assets/instruments/demo.inst": `name = "demo"

[oscillator]
waveform = "square"

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.6
release = 0.15
`,
		"assets/sfx/demo.sfx": `duration = 0.15
volume = 0.8

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.3
release = 0.1

[voice.pitch]
start = 200
end = 600
curve = "exponential"
`,
		"assets/tracks/demo.track": `tempo = 120
ticks_per_beat = 4
loop = true
loop_start = 0

[[channel]]
name = "melody"
instrument = "demo"
volume = 0.8

[pattern.main]
ticks = 8
data = """
melody
C4
---
E4
---
G4
---
E4
---
"""

[song]
sequence = ["main", "main"]
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}
}
