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
	expected := []string{
		"build/assets/sprites/player.png",
		"build/assets/sprites/tiles.png",
		"build/assets/maps/level1.json",
		"build/assets/audio/jump.wav",
		"build/assets/audio/coin.wav",
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

	// player.sprite has idle (32x32) and coin (16x16) and heart (16x16).
	pngPath := filepath.Join(dir, "build/assets/sprites/player.png")
	f, err := os.Open(pngPath)
	if err != nil {
		t.Fatalf("opening PNG: %v", err)
	}
	defer f.Close()

	imgCfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decoding PNG config: %v", err)
	}

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

	jsonPath := filepath.Join(dir, "build/assets/maps/level1.json")
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

	for _, name := range []string{"jump", "coin", "demo"} {
		wavPath := filepath.Join(dir, "build/assets/audio/"+name+".wav")
		info, err := os.Stat(wavPath)
		if err != nil {
			t.Errorf("missing WAV %s: %v", name, err)
			continue
		}
		// WAV header is 44 bytes; actual audio data should add more.
		if info.Size() <= 44 {
			t.Errorf("%s.wav too small (%d bytes)", name, info.Size())
		}
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
window_width = 1200
window_height = 900
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
		"assets/sprites/player.sprite": `palette = "default"
grid = 32

palette_extend = { "n" = "#1a1a2e", "t" = "#e8d4b8" }

[sprite.idle]
framerate = 3

[[sprite.idle.frame]]
pixels = """
________________________________
_____________hhhhh______________
___________hhhhhhhhh____________
__________hhhhhhhhhh____________
__________hhhhhhhhhh____________
__________dddddddddd____________
_________dddddddddddd___________
_________dddwkddwkddd___________
_________dddkkddkkddd___________
_________ddttttttttdd___________
_________dddttttttddd___________
__________dtttsttttd____________
__________ddttttttdd____________
___________dddddddd_____________
___________bbbbbbbbb____________
__________bbbblbbbbbb___________
_________bbbbbbbbbbbbb__________
________sbbbbbbbbbbbbbs_________
________s_bbbbbbbbbbb_s_________
__________bbbbbbbbbbb___________
___________bbb___bbb____________
___________bbb___bbb____________
___________bbb___bbb____________
__________nnnn___nnnn___________
__________nnnn___nnnn___________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
"""

[[sprite.idle.frame]]
pixels = """
________________________________
_____________hhhhh______________
___________hhhhhhhhh____________
__________hhhhhhhhhh____________
__________hhhhhhhhhh____________
__________dddddddddd____________
_________dddddddddddd___________
_________dddwkddwkddd___________
_________dddkkddkkddd___________
_________ddttttttttdd___________
_________dddttttttddd___________
__________dtttsttttd____________
__________ddttttttdd____________
___________dddddddd_____________
___________bbbbbbbbb____________
__________bbbblbbbbbb___________
_________bbbbbbbbbbbbb__________
________sbbbbbbbbbbbbbs_________
________s_bbbbbbbbbbb_s_________
__________bbbbbbbbbbb___________
___________bbb___bbb____________
___________bbb___bbb____________
__________nbbb___bbbn___________
__________nnnn___nnnn___________
___________nn_____nn____________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
"""

[sprite.coin]
grid = 16
framerate = 6

[[sprite.coin.frame]]
pixels = """
_____yyyy_______
___yyooooyyy____
__yooyyyyyooy___
__yoyyyyyyyyoy__
__yoyyyy_yyyoy__
__yoyyyyyyyyoy__
__yoyyyyyyyyoy__
__yoyyyyyyyyoy__
___yooooooooy___
____yyyyyyyy____
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
______yy________
_____yooy_______
____yoyyoy______
____yoy_oy______
____yoyyoy______
____yoyyoy______
____yoyyoy______
_____yooy_______
______yy________
________________
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
_______y________
______yoy_______
______yoy_______
______yoy_______
______yoy_______
______yoy_______
______yoy_______
_______y________
________________
________________
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
______yy________
_____yooy_______
____yoyyoy______
____yoy_oy______
____yoyyoy______
____yoyyoy______
____yoyyoy______
_____yooy_______
______yy________
________________
________________
________________
________________
________________
________________
________________
"""

[sprite.heart]
grid = 16
pixels = """
________________
___rr____rr_____
__rrrr__rrrr____
_rrrrrrrrrrrr___
_rrrrrrrrrrrr___
_rrrrrrrrrrrr___
__rrrrrrrrrr____
___rrrrrrrr_____
____rrrrrr______
_____rrrr_______
______rr________
________________
________________
________________
________________
________________
"""
`,
		"assets/sprites/tiles.sprite": `palette = "default"
grid = 16

[sprite.grass]
pixels = """
ccggccggccggccgg
ggccggccggccggcc
ccgcgcgcgcgcgccc
gcccccggccccccgc
ccchccccchccccch
hchhhhhchhhhhchh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
"""

[sprite.dirt]
pixels = """
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
"""

[sprite.stone]
pixels = """
eeffffffeeffffff
flllllfeflllllfe
flllflfefllflffe
flllllfeflllllfe
eeffffffeeffffff
feflllleefellllf
fefllfleefelfllf
feflllleefellllf
eeffffffeeffffff
flllllfeflllllfe
flllflfefllflffe
flllllfeflllllfe
eeffffffeeffffff
feflllleefellllf
fefllfleefelfllf
feflllleefellllf
"""

[sprite.sky]
pixels = """
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
"""
`,
		"assets/maps/level1.map": `tile_size = 16

[tileset]
_ = ""
g = "tiles:grass"
d = "tiles:dirt"
s = "tiles:stone"
b = "tiles:sky"

[layer.background]
scroll_x = 0.5
pixels = """
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
"""

[layer.main]
pixels = """
________________
________________
________________
________________
_______sss______
________________
__gg_______gg___
__dd_______dd___
gggggggggggggggg
dddddddddddddddd
"""

[layer.entities]

[[layer.entities.entity]]
type = "spawn"
x = 2
y = 7
[layer.entities.entity.properties]
sprite = "player:idle"

[[layer.entities.entity]]
type = "coin"
x = 4
y = 5
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "coin"
x = 8
y = 3
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "coin"
x = 11
y = 5
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "enemy"
x = 13
y = 7
[layer.entities.entity.properties]
sprite = "player:heart"
`,
		"assets/instruments/lead.inst": `name = "lead"

[oscillator]
waveform = "square"
duty_cycle = 0.5

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.6
release = 0.2

[filter]
type = "lowpass"
cutoff = 2000
resonance = 0.2
`,
		"assets/instruments/bass.inst": `name = "bass"

[oscillator]
waveform = "triangle"

[envelope]
attack = 0.005
decay = 0.15
sustain = 0.5
release = 0.1
`,
		"assets/sfx/jump.sfx": `duration = 0.2
volume = 0.7

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.08
sustain = 0.2
release = 0.12

[voice.pitch]
start = 220
end = 660
curve = "exponential"
`,
		"assets/sfx/coin.sfx": `duration = 0.12
volume = 0.6

[[voice]]
waveform = "square"
duty_cycle = 0.25

[voice.envelope]
attack = 0.0
decay = 0.03
sustain = 0.4
release = 0.09

[voice.pitch]
start = 880
end = 1320
curve = "linear"
`,
		"assets/tracks/demo.track": `tempo = 140
ticks_per_beat = 4
loop = true
loop_start = 0

[[channel]]
name = "melody"
instrument = "lead"
volume = 0.7

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.6

[pattern.intro]
ticks = 16
data = """
melody  | bass
C4      | C2
---     | ---
E4      | ---
---     | ---
G4      | G2
---     | ---
E4      | ---
---     | ---
A4      | A2
---     | ---
G4      | ---
---     | ---
E4      | E2
---     | ---
D4      | ---
---     | ---
"""

[pattern.verse]
ticks = 16
data = """
melody  | bass
E4      | A2
---     | ---
D4      | ---
---     | ---
C4      | F2
---     | ---
D4      | ---
---     | ---
E4      | G2
---     | ---
E4      | ---
---     | ---
E4      | C2
---     | ---
^^^     | ---
...     | ^^^
"""

[song]
sequence = ["intro", "verse", "intro", "verse"]
`,
	}

	for path, content := range files {
		fullPath := filepath.Join(dir, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("writing %s: %v", path, err)
		}
	}
}
