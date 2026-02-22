# Runefact: Runes Become Artifacts

## Problem

Building 2D pixel art games without art skills. AI coding agents (Claude Code, etc.) are excellent at generating and manipulating structured text but terrible at working with binary asset formats. We need a text-first asset pipeline where every game asset -- sprites, tilemaps, sounds, music -- is defined in human-readable, LLM-friendly text files, then compiled to formats ebitengine can consume.

## Goal

A Go CLI tool (`runefact`) that:

1. Reads text-based asset definitions ("runes") from a project directory
2. Compiles them to PNG sprite sheets, JSON tilemaps, and WAV audio files ("artifacts")
3. Generates a Go asset manifest package for type-safe loading in ebitengine
4. Includes a live-reloading previewer for rapid iteration

A VS Code extension (`runefact-vscode`) that:

5. Provides syntax highlighting for all Runefact file formats, including color-coded pixel grids and tracker patterns
6. Runs inline validation with diagnostics (errors/warnings in the editor)
7. Integrates with the CLI for build, validate, and preview commands

A local MCP server (`runefact mcp`) that:

8. Exposes Runefact operations as MCP tools so Claude Code, Claude Desktop, and other MCP-capable agents can build, validate, inspect, and manipulate assets directly
9. Serves format specifications and project state as MCP resources for in-context reference

Comprehensive usage documentation (`docs/`) that:

10. Covers installation, quickstart, format references, ebitengine integration, and AI-assisted workflows
11. Serves as both human docs and agent context (Markdown, readable by Claude when included as project files)

## Non-Goals

- Runtime asset loading/management (that's your game code)
- 3D anything
- Asset editing GUI (the previewer is read-only, your editor is your text editor + AI agent)
- Networking/multiplayer asset sync
- Compression/optimization beyond basic PNG

---

## Project Structure

**Game project (user's repo):**

```
assets/
  palettes/
    default.palette
  sprites/
    player.sprite
    enemies.sprite
    tiles.sprite
  maps/
    level1.map
    overworld.map
  sounds/
    jump.sfx
    explosion.sfx
  music/
    theme.track
    battle.track
  instruments/
    bass.inst
    lead.inst
    drums.inst
runefact.toml             # project config
```

**Runefact repo structure:**

```
cmd/
  runefact/               # CLI entry point (build, validate, preview, watch, init, mcp)
internal/
  palette/                # .palette parser
  sprite/                 # .sprite parser + PNG renderer
  tilemap/                # .map parser + JSON output
  instrument/             # .inst parser
  sfx/                    # .sfx parser + WAV renderer
  track/                  # .track parser + WAV renderer
  audio/                  # shared audio: synthesis, limiter, WAV writer
  manifest/               # manifest.go generator
  preview/                # ebitengine previewer
  watcher/                # fsnotify file watcher
  mcp/                    # MCP server implementation
    server.go             # MCP protocol handler (stdio transport)
    tools.go              # tool definitions and handlers
    resources.go          # resource definitions and handlers
    inspect.go            # asset inspection logic
vscode/
  runefact-vscode/        # VS Code extension
    package.json
    syntaxes/             # TextMate grammar JSON files
      runefact-palette.tmLanguage.json
      runefact-sprite.tmLanguage.json
      runefact-map.tmLanguage.json
      runefact-instrument.tmLanguage.json
      runefact-sfx.tmLanguage.json
      runefact-track.tmLanguage.json
    src/
      extension.ts        # extension entry point
      diagnostics.ts      # inline validation logic
      colorProvider.ts    # color decorators
      pixelGrid.ts        # pixel grid background coloring
      trackerColors.ts    # tracker note coloring
    snippets/
      runefact.code-snippets
docs/
  getting-started.md      # installation + quickstart
  format-reference.md     # complete spec for all file formats
  sprite-guide.md         # detailed sprite authoring guide
  map-guide.md            # detailed map authoring guide
  audio-guide.md          # SFX + music authoring guide
  ebitengine-integration.md  # loading artifacts in your game
  ai-workflows.md         # working with Claude Code + MCP
  mcp-reference.md        # MCP server tool/resource reference
  CLAUDE.md               # agent context file (see Usage Docs section)
```

Output:

```
build/assets/
  sprites/
    player.png            # sprite sheet per .sprite file
    enemies.png
    tiles.png
  maps/
    level1.json
    overworld.json
  audio/
    jump.wav
    explosion.wav
    theme.wav
    battle.wav
  manifest.go             # generated Go package with constants/loaders
```

---

## Project Config: `runefact.toml`

```toml
[project]
name = "my-game"
output = "build/assets"
package = "assets"        # Go package name for manifest

[defaults]
sprite_size = 16          # default tile/sprite size if not specified per file
sample_rate = 44100
bit_depth = 16

[preview]
window_width = 800
window_height = 600
background = "#1a1a2e"    # previewer background color
pixel_scale = 4           # default zoom for sprite/map preview
audio_volume = 0.5        # default preview volume (0.0-1.0)
```

---

## Format Specifications

### 1. Palette Files (`.palette`)

Shared color palettes that sprites reference by name. Keeps color definitions DRY and ensures visual consistency.

```toml
# palettes/default.palette
name = "default"

[colors]
_  = "transparent"
k  = "#000000"       # black
w  = "#ffffff"       # white
r  = "#ff004d"       # red
b  = "#29adff"       # blue
g  = "#00e436"       # green
d  = "#1d2b53"       # dark blue
s  = "#ffccaa"       # skin
h  = "#ab5236"       # hair/brown
y  = "#ffec27"       # yellow
o  = "#ffa300"       # orange
p  = "#7e2553"       # purple
l  = "#83769c"       # lavender
c  = "#008751"       # dark green
e  = "#5f574f"       # dark grey
f  = "#c2c3c7"       # light grey
```

Design notes:
- Single-char keys optimize for grid readability (the whole point is that LLMs generate grids of these)
- Multi-char keys allowed for extended palettes: `sk = "#ffccaa"` -- grid cells use `[sk]` bracket syntax
- Palette files are optional; sprites can inline their palette
- Max recommended palette size: 32 colors (keeps grids scannable)

### 2. Sprite Files (`.sprite`)

Each `.sprite` file defines one or more named sprites/animations. Compiles to a single PNG sprite sheet.

```toml
# sprites/player.sprite
palette = "default"      # reference to .palette file
grid = 16                # pixel dimensions per frame (square; or "16x24" for non-square)

# --- Static sprite ---
[sprite.heart]
grid = 8                 # override per-sprite
pixels = """
___rr___
__rrrr__
_rrrrrr_
rrrrrrrr
rrrrrrrr
_rrrrrr_
__rrrr__
___rr___
"""

# --- Animated sprite ---
[sprite.idle]
framerate = 8            # fps for this animation

[[sprite.idle.frame]]
pixels = """
____kk____kk____
___ksskk__kk____
___ksssk__kk____
____kssk________
_____kbbbbk_____
____kbbbbbbk____
___kbbbbbbbbk___
___kbkbbbbkbk___
___kbbbbbbbbk___
___kbbbbbbbbk___
____kbbbbbbk____
_____kbbbbk_____
______kkkk______
_____kk__kk_____
____kk____kk____
___kk______kk___
"""

[[sprite.idle.frame]]
pixels = """
____kk____kk____
___ksskk__kk____
___ksssk__kk____
____kssk________
_____kbbbbk_____
____kbbbbbbk____
___kbbbbbbbbk___
___kbkbbbbkbk___
___kbbbbbbbbk___
___kbbbbbbbbk___
____kbbbbbbk____
_____kbbbbk_____
_____kkkkk______
____kk__kk______
____kk__kk______
____kk__kk______
"""

# --- Sprite with multi-char palette keys ---
[sprite.wizard]
palette_extend = { sk = "#ffccaa", rb = "#4400ff", ht = "#ffdd00" }
grid = 16

[[sprite.wizard.frame]]
pixels = """
____[ht][ht][ht][ht][ht][ht]________
___[ht][ht][ht][ht][ht][ht][ht]_____
__[rb][rb][rb][rb][rb][rb][rb][rb]___
__[rb][sk][sk][rb][rb][sk][sk][rb]___
__[rb][sk]k_[rb][rb][sk]k_[rb]______
__[rb][sk][sk][rb][rb][sk][sk][rb]___
___[rb][sk][sk][sk][sk][sk][sk]______
____[rb][rb][rb][rb][rb][rb]_________
_____[rb][rb][rb][rb][rb]____________
____[rb][rb]____[rb][rb]_____________
___[rb][rb]______[rb][rb]____________
__kk________________kk______________
"""
```

**Sprite sheet layout rules:**
- All frames for a sprite are laid out horizontally in the sheet
- Different sprites within the same file stack vertically
- Sheet is auto-sized to fit; no manual sizing needed
- Manifest records each sprite's position, frame count, and framerate

**Validation rules:**
- All frames in an animation must have the same dimensions
- All pixel rows must have the same width
- All palette keys must resolve (via palette file or inline extend)
- Grid size must match actual pixel dimensions

### 3. Map Files (`.map`)

Tilemaps that reference sprites as tile sources.

```toml
# maps/level1.map
tile_size = 16

[tileset]
# Map single chars to sprite references: "file:sprite_name"
# Static sprites only (first frame used for animated sprites)
G = "tiles:grass"
D = "tiles:dirt"
W = "tiles:water"
S = "tiles:stone"
T = "tiles:tree_top"
t = "tiles:tree_trunk"
_ = ""                   # empty/air

# --- Layers render bottom to top ---

[layer.background]
scroll_x = 0.5           # parallax factor (optional, informational -- your game interprets this)
pixels = """
________________
________________
________________
________________
________________
________________
GGGGGGGGGGGGGGGG
DDDDDDDDDDDDDDDD
"""

[layer.main]
pixels = """
________________
________T_______
________t_______
____SS__t_______
___SSSS_________
__SSSSSS________
GGGGGGGGGGGGGGGG
DDDDDDDDDDDDDDDD
"""

[layer.objects]
# For entity/object placement, use coordinate lists instead of grids
[[layer.objects.entity]]
type = "spawn"
x = 2
y = 5

[[layer.objects.entity]]
type = "coin"
x = 7
y = 4

[[layer.objects.entity]]
type = "enemy:slime"
x = 10
y = 5
properties = { patrol_range = 3, direction = "left" }
```

**Output format (JSON):**

```json
{
  "tile_size": 16,
  "width": 16,
  "height": 8,
  "tileset": {
    "G": { "source": "tiles.png", "sprite": "grass", "x": 0, "y": 0 },
    "D": { "source": "tiles.png", "sprite": "dirt", "x": 16, "y": 0 }
  },
  "layers": [
    {
      "name": "background",
      "type": "tile",
      "scroll_x": 0.5,
      "data": [[0,0,0,...], [1,1,1,...]]
    },
    {
      "name": "objects",
      "type": "entity",
      "entities": [
        { "type": "spawn", "x": 2, "y": 5 },
        { "type": "coin", "x": 7, "y": 4 }
      ]
    }
  ]
}
```

### 4. Instrument Files (`.inst`)

Define synthesizer instruments that sound effects and music reference.

```toml
# instruments/bass.inst
name = "bass"

[oscillator]
waveform = "square"      # sine | square | triangle | sawtooth | noise | pulse
duty_cycle = 0.25        # pulse width (only for "pulse" waveform, 0.0-1.0)

[envelope]
attack = 0.01            # seconds
decay = 0.1
sustain = 0.6            # level 0.0-1.0
release = 0.15

[filter]                 # optional low-pass filter
type = "lowpass"         # lowpass | highpass | bandpass
cutoff = 800             # Hz
resonance = 0.3          # 0.0-1.0

[effects]                # all optional
vibrato_depth = 0.0      # semitones
vibrato_rate = 0.0       # Hz
pitch_sweep = 0.0        # semitones per second (positive = up, negative = down)
distortion = 0.0         # 0.0-1.0
```

### 5. Sound Effect Files (`.sfx`)

Procedural sound effect definitions. No samples, no binary -- just synthesis parameters.

```toml
# sounds/jump.sfx
duration = 0.15          # total duration in seconds
volume = 0.8             # 0.0-1.0

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.3
release = 0.1

[voice.pitch]
start = 200              # Hz
end = 600                # Hz
curve = "exponential"    # linear | exponential | logarithmic

[voice.effects]
vibrato_depth = 0.0
vibrato_rate = 0.0

# --- Multi-voice SFX (e.g., explosion) ---
# sounds/explosion.sfx

# [[voice]]
# waveform = "noise"
# [voice.envelope]
# attack = 0.0
# decay = 0.3
# sustain = 0.0
# release = 0.2
# [voice.filter]
# type = "lowpass"
# cutoff_start = 4000
# cutoff_end = 200
# curve = "exponential"
#
# [[voice]]
# waveform = "sine"
# [voice.envelope]
# attack = 0.0
# decay = 0.1
# sustain = 0.0
# release = 0.15
# [voice.pitch]
# start = 120
# end = 40
# curve = "linear"
```

**Design rationale for SFX:**
- Voices are layered additively (mixed together)
- Each voice has independent waveform, envelope, pitch, and filter
- Filter cutoff can sweep independently from pitch
- This covers the vast majority of retro game SFX: jumps, hits, coins, explosions, laser shots, powerups
- An LLM can iterate on these parameters conversationally: "make the explosion bassier" = lower the cutoff_end, increase decay

### 6. Music Track Files (`.track`)

Tracker-style music format. Pattern-based, multi-channel, references instruments.

```toml
# music/theme.track
tempo = 120              # BPM
ticks_per_beat = 4       # resolution (4 = 16th notes)
loop = true
loop_start = 0           # pattern index to loop back to

# --- Channel definitions ---
[[channel]]
name = "melody"
instrument = "lead"      # references instruments/lead.inst
volume = 0.8

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.7

[[channel]]
name = "kick"
instrument = "drums"
volume = 0.9

[[channel]]
name = "hihat"
instrument = "drums"
volume = 0.5

# --- Pattern definitions ---
# Each pattern is N ticks long (rows)
# Note format: NOTE OCTAVE [VOLUME] [EFFECT]
#   Note: C, C#, D, D#, E, F, F#, G, G#, A, A#, B
#   ---  = continue previous note
#   ... = silence
#   ^^^ = note off (trigger release)

[pattern.intro]
ticks = 16
data = """
melody   | bass    | kick   | hihat
C4       | C2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
E4       | C2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
G4       | E2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
E4       | E2      | C2     | ...
...      | ...     | ...    | F#5
^^^      | ^^^     | ...    | ...
...      | ...     | ...    | F#5
"""

[pattern.verse]
ticks = 16
data = """
melody   | bass    | kick   | hihat
C4       | C2      | C2     | F#5
---      | ---     | ...    | ...
D4       | ...     | ...    | F#5
---      | ...     | ...    | ...
E4       | G2      | C2     | F#5
---      | ---     | ...    | ...
D4       | ...     | ...    | F#5
---      | ...     | ...    | ...
C4       | A1      | C2     | F#5
---      | ---     | ...    | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
G3       | F2      | C2     | F#5
---      | ---     | ...    | ...
^^^      | ^^^     | ...    | F#5
...      | ...     | ...    | ...
"""

# --- Song arrangement: ordered list of pattern names ---
[song]
sequence = [
  "intro",
  "intro",
  "verse",
  "verse",
]
```

**Design rationale for music:**
- Tracker format is the most text-dense, LLM-friendly music representation that exists
- Column alignment makes patterns visually scannable
- Note names are universal (no MIDI numbers to decode)
- `---` (sustain), `...` (silence), `^^^` (note-off) are visually distinct
- Instruments are separate files so they're reusable and independently iterable
- Pattern + arrangement separation means you can say "repeat the verse but swap the bass pattern" easily

**Per-note effects (optional, column suffix):**

```
C4 v08     # volume 08 (hex, 00-0F)
C4 >02     # pitch slide up 2
C4 <02     # pitch slide down 2
C4 ~04     # vibrato depth 4
C4 a08     # arpeggio (semitones: +0, +8)
```

---

## CLI Interface

```
runefact build                # compile all assets
runefact build --sprites      # compile sprites only
runefact build --maps         # compile maps only
runefact build --audio        # compile audio only (sfx + music)
runefact watch                # watch + auto-rebuild (headless, no preview window)
runefact preview              # open live-reloading previewer
runefact preview player.sprite          # open previewer focused on a specific file
runefact preview level1.map             # preview a specific map
runefact preview jump.sfx               # preview a specific sound
runefact init                 # scaffold project structure
runefact validate             # check all files for errors without building
runefact mcp                  # start MCP server (stdio transport)
runefact docs                 # print path to docs directory
```

---

## Previewer

The previewer is a live-reloading ebitengine window that renders compiled artifacts and watches source rune files for changes. It is strictly a viewer, not an editor. The AI agent or human edits the text files; the previewer shows the result.

### Architecture

- Built on ebitengine (same rendering stack as the target game)
- File watcher (fsnotify) triggers incremental rebuild on source changes
- Rebuild errors display inline in the preview window (red overlay with error text + file/line reference) rather than crashing
- Asset type auto-detected from file extension; previewer mode adapts accordingly

### Preview Modes

**Sprite Preview** (`runefact preview player.sprite`)
- Renders all sprites in the file on a grid layout
- Animated sprites play their animation loop
- Displays: sprite name, dimensions, frame count, framerate as overlay text
- Controls:
  - Mouse wheel: zoom in/out (1x to 32x, pixel-perfect nearest-neighbor scaling)
  - Click sprite: isolate it (show only that sprite, centered)
  - Space: pause/resume all animations
  - Left/Right arrows: step through frames when paused
  - G: toggle pixel grid overlay
  - B: cycle background color (dark, light, checkerboard) for transparency visibility

**Map Preview** (`runefact preview level1.map`)
- Renders all tile layers composited
- Entity layer shown as colored markers with type labels
- Controls:
  - WASD/Arrow keys: pan
  - Mouse wheel: zoom
  - Tab: cycle layer visibility (all, individual layers, entity overlay toggle)
  - G: toggle tile grid overlay

**SFX Preview** (`runefact preview jump.sfx`)
- Displays waveform visualization (rendered from the compiled WAV data, not real-time)
- Displays envelope shape, pitch curve, and filter sweep as simple graphs
- Controls:
  - Enter: play the sound effect once
  - Audio is NEVER auto-played on load or on file change
  - Volume knob (mouse drag or +/- keys), defaults to `preview.audio_volume` from config

**Music Preview** (`runefact preview theme.track`)
- Displays tracker-style pattern view: scrolling rows with channel columns, current row highlighted
- Shows song position: which pattern, which row, total elapsed time
- Controls:
  - Enter: start playback from beginning
  - Space: pause/resume
  - Escape: stop and reset to beginning
  - Audio is NEVER auto-played on load or on file change
  - Left/Right: skip to previous/next pattern
  - Volume: +/- keys, defaults to `preview.audio_volume` from config

**Auto-detect mode** (`runefact preview` with no argument)
- Opens previewer showing an asset browser: tree view of all rune files in the project
- Click any file to preview it
- File watcher active on entire project; any changed file re-renders if currently viewed

### Audio Safety

Audio in the previewer follows strict safety-first principles. A buggy `.sfx` or `.track` definition could produce ear-destroying output (DC offset, full-scale noise, resonance blowup). The previewer must protect against this.

**Hard rules:**
- Audio is NEVER auto-played. Not on file load, not on file change, not on previewer startup. Playback is always explicit user action (Enter/Space).
- All audio output runs through a brickwall limiter before reaching the audio device. Threshold: -1 dBFS. Attack: 0ms (lookahead). Release: 50ms. This is non-optional and cannot be disabled.
- A peak meter is always visible during playback showing current output level.
- If a compiled audio file exceeds 10 seconds of consecutive clipping (samples at max amplitude), the previewer displays a warning: "This audio is likely broken -- sustained clipping detected."
- Volume defaults to 50% (configurable in `runefact.toml`). The previewer remembers volume across sessions (stored in `~/.config/runefact/preview.json`).
- On file change during playback: playback stops immediately. User must explicitly restart. No hot-swap of playing audio.

**Soft protections:**
- DC offset removal: high-pass filter at 10Hz on all audio output
- If compiled WAV has NaN or Inf samples, replace with silence and display a warning
- If any single SFX voice produces output >6 dB above the mix (likely a misconfigured envelope), display a per-voice level warning

### Previewer State Persistence

The previewer stores minimal UI state in `~/.config/runefact/preview.json`:
- Window size and position
- Last viewed file
- Zoom level per file type
- Volume level
- Background color preference

This state is convenience only. The previewer works fine if this file is missing or deleted.

---

## Build Output

On successful build:
- All compiled assets in `build/assets/` (or configured output dir)
- `manifest.go` generated with:
  - Sprite sheet metadata (positions, frame counts, framerates)
  - Map metadata (dimensions, layer names)
  - Audio file references
  - Constants for all named sprites, maps, sounds, tracks

### `manifest.go` Example

```go
// Code generated by runefact. DO NOT EDIT.
package assets

// Sprite sheets
const (
    SpriteSheetPlayer  = "sprites/player.png"
    SpriteSheetEnemies = "sprites/enemies.png"
    SpriteSheetTiles   = "sprites/tiles.png"
)

// Sprite metadata
type SpriteInfo struct {
    Sheet    string
    X, Y    int    // position in sheet
    W, H    int    // frame dimensions
    Frames  int    // frame count (1 for static)
    FPS     int    // 0 for static
}

var Sprites = map[string]SpriteInfo{
    "player:idle":  {SpriteSheetPlayer, 0, 0, 16, 16, 2, 8},
    "player:heart": {SpriteSheetPlayer, 0, 16, 8, 8, 1, 0},
    "tiles:grass":  {SpriteSheetTiles, 0, 0, 16, 16, 1, 0},
    // ...
}

// Maps
const (
    MapLevel1    = "maps/level1.json"
    MapOverworld = "maps/overworld.json"
)

// Audio
const (
    SFXJump      = "audio/jump.wav"
    SFXExplosion = "audio/explosion.wav"
    TrackTheme   = "audio/theme.wav"
    TrackBattle  = "audio/battle.wav"
)
```

---

## VS Code Extension: `runefact-vscode`

A VS Code extension that provides syntax highlighting, inline validation, and quality-of-life features for all Runefact file formats. Published to the VS Code Marketplace as `runefact`.

### Why This Matters

The primary user of these files is an AI agent, but the human still needs to read, review, and occasionally hand-edit them. Syntax highlighting makes pixel grids dramatically easier to scan visually, and inline errors catch problems before you even run `runefact build`. This also benefits the AI workflow: when Claude Code or Cursor opens a `.sprite` file, VS Code's language features provide context that helps the agent understand the file structure.

### File Association

The extension registers these file extensions with custom language IDs:

| Extension | Language ID | Icon |
|-----------|------------|------|
| `.palette` | `runefact-palette` | Color swatch |
| `.sprite` | `runefact-sprite` | Grid/pixel icon |
| `.map` | `runefact-map` | Map pin icon |
| `.inst` | `runefact-instrument` | Waveform icon |
| `.sfx` | `runefact-sfx` | Speaker icon |
| `.track` | `runefact-track` | Music note icon |
| `runefact.toml` | `toml` (built-in) | Uses existing TOML support |

### Syntax Highlighting

All formats are TOML-based, so the extension layers Runefact-specific highlighting on top of TOML's base grammar via TextMate injection grammars.

**Palette files (`.palette`)**
- TOML keys/values highlighted normally
- Hex color values (`#ff004d`) highlighted in the actual color they represent (inline color decorator)
- Single-char palette keys highlighted distinctly from values

**Sprite files (`.sprite`)**
- TOML structure (headers, keys) highlighted normally
- Inside `pixels = """..."""` blocks:
  - Each unique palette character gets a distinct background color matching its palette definition
  - `_` (transparent) rendered with a subtle dotted/checkerboard background
  - Bracket sequences `[sk]` highlighted as a unit with the corresponding palette color
  - This is the killer feature: a pixel grid in your editor is color-coded to approximate what the sprite actually looks like
- `palette_extend` color values get inline color decorators

**Map files (`.map`)**
- TOML structure highlighted normally
- Inside layer `pixels = """..."""` blocks:
  - Each tileset key character gets a distinct background color (auto-assigned from a preset hue rotation, since tiles don't map directly to colors)
  - `_` (empty/air) rendered with subtle dotted background
- Entity definitions: `type` values highlighted as string constants
- Sprite references (`"tiles:grass"`) highlighted as cross-file links

**Instrument files (`.inst`)**
- TOML structure highlighted normally
- Waveform type values (`sine`, `square`, etc.) highlighted as enum/keyword
- Numeric parameter values: range-appropriate coloring (e.g., values outside 0.0-1.0 for `sustain` flagged)

**SFX files (`.sfx`)**
- TOML structure highlighted normally
- Waveform types highlighted as enum/keyword
- Curve types (`linear`, `exponential`, `logarithmic`) highlighted as enum/keyword
- Filter types highlighted as enum/keyword

**Track files (`.track`)**
- TOML structure highlighted normally
- Inside `data = """..."""` pattern blocks:
  - Column headers (channel names) highlighted as labels
  - `|` separators dimmed/subtle
  - Note values (`C4`, `F#5`, `A#3`) highlighted with pitch-class coloring (C=red, D=orange, E=yellow, etc. -- chromatic rainbow)
  - `---` (sustain) highlighted as a continuation marker (dimmed/gray)
  - `...` (silence) highlighted as a rest (very dim/faded)
  - `^^^` (note-off) highlighted distinctly (e.g., italic or different color)
  - Effect suffixes (`v08`, `>02`, `~04`) highlighted as modifiers
- Pattern names in `[song] sequence` highlighted as references
- Instrument references in `[[channel]]` highlighted as cross-file links

### Inline Color Decorators

For `.palette` and `.sprite` files, the extension uses VS Code's `DocumentColorProvider` API to:
- Show colored squares next to hex color definitions
- Allow clicking the color square to open the VS Code color picker (edits the hex value in place)
- In pixel grids, optionally show a thin colored underline per character matching its palette color (togglable -- can be noisy at high density)

### Inline Diagnostics

The extension runs lightweight validation and reports errors/warnings as VS Code diagnostics (squiggly underlines + Problems panel). This does NOT require the `runefact` CLI to be installed -- the validation logic is self-contained in the extension's TypeScript.

**Palette validation:**
- Duplicate key names -- error
- Invalid hex color format -- error
- Multi-char keys longer than 4 characters -- warning (gets unwieldy in grids)

**Sprite validation:**
- Ragged rows (inconsistent width within a `pixels` block) -- error on the short/long row
- Unknown palette key (if `.palette` file is resolvable in the workspace) -- warning
- Frame dimension mismatch across `[[sprite.X.frame]]` entries -- error
- `pixels` block dimensions don't match `grid` value -- error

**Map validation:**
- Unknown tileset key in layer grid -- warning
- Ragged rows -- error
- Sprite reference format check (`"file:name"` pattern) -- error on malformed refs

**Track validation:**
- Column count in `data` block doesn't match `[[channel]]` count -- error
- Unknown note format (not matching `[A-G]#?[0-9]` or `---`/`...`/`^^^`) -- error
- Pattern name in `sequence` not found in `[pattern.*]` definitions -- error
- Instrument reference not resolvable in workspace -- warning

**Validation is best-effort and intentionally lenient.** It catches structural/syntactic issues, not semantic ones (e.g., it won't tell you your music sounds bad). When in doubt, warn rather than error. False positives that block the agent's workflow are worse than missed errors.

### Snippets

The extension includes snippets to scaffold common structures. These are useful both for humans and for AI agents that can trigger VS Code snippets.

| Prefix | Description | File Type |
|--------|------------|-----------|
| `rfpal` | New palette file skeleton | `.palette` |
| `rfsprite` | New static sprite definition | `.sprite` |
| `rfanim` | New animated sprite with 2 frames | `.sprite` |
| `rfframe` | Additional animation frame | `.sprite` |
| `rfmap` | New map file with tileset and one layer | `.map` |
| `rfinst` | New instrument definition | `.inst` |
| `rfsfx` | New sound effect with one voice | `.sfx` |
| `rftrack` | New track with 2 channels and 1 pattern | `.track` |
| `rfpat` | Additional pattern block | `.track` |

### Commands

The extension registers these commands in the VS Code command palette:

- **Runefact: Build All** -- runs `runefact build` in the integrated terminal
- **Runefact: Build Current File Type** -- runs `runefact build --sprites` (or `--maps`, `--audio`) based on the active file
- **Runefact: Validate** -- runs `runefact validate` in the integrated terminal
- **Runefact: Preview Current File** -- runs `runefact preview <current-file>` 
- **Runefact: Open Previewer** -- runs `runefact preview` (asset browser mode)
- **Runefact: Initialize Project** -- runs `runefact init` in workspace root

Keybindings (default, user-overridable):
- `Ctrl+Shift+B` (in a Runefact file) -- Build All
- `Ctrl+Shift+P` (in a Runefact file) -- Preview Current File

### Extension Settings

```json
{
  "runefact.cliPath": "runefact",           // path to runefact binary
  "runefact.autoValidateOnSave": true,       // run inline validation on save
  "runefact.pixelGridColorMode": "palette",  // "palette" | "hue-rotate" | "off"
  "runefact.colorDecorators": true,          // show inline color swatches
  "runefact.tracker.chromaticColors": true   // color-code notes by pitch class in .track files
}
```

### Implementation Notes

- **Built with TypeScript** as a standard VS Code extension
- **TextMate grammars** (JSON) for syntax highlighting -- injected on top of a TOML base grammar
- **No Language Server Protocol (LSP) needed** for v1. The inline validation is simple enough to run as a `DocumentDiagnosticProvider` in the extension host. If validation grows complex enough to warrant it (e.g., cross-file reference resolution across large projects), an LSP can be introduced later.
- **Pixel grid coloring** is the hardest part. It requires parsing the `pixels` block, resolving the palette (potentially from another file), and applying `DocumentSemanticTokensProvider` or `DecorationProvider` with computed background colors. This is doable but performance-sensitive on large sprite files -- debounce to 200ms after last keystroke.
- The extension should bundle a copy of the format specs so it can validate without the CLI installed. The CLI and extension must agree on the format, so the spec is the source of truth for both.

### Distribution

- Published to VS Code Marketplace under publisher `runefact`
- Also works in Cursor, Windsurf, and other VS Code forks out of the box
- Extension `.vsix` also available as a GitHub release artifact for offline install
- Zero dependencies on the `runefact` CLI for syntax highlighting and validation -- the CLI is only needed for the build/preview commands

- Zero dependencies on the `runefact` CLI for syntax highlighting and validation -- the CLI is only needed for the build/preview commands

---

## MCP Server: `runefact mcp`

A local MCP (Model Context Protocol) server that exposes Runefact operations to AI agents. This is the integration layer that makes Claude Code, Claude Desktop, Cursor, and any MCP-capable client a first-class Runefact user. The agent doesn't shell out to the CLI and parse stdout -- it calls structured tools with typed inputs/outputs.

### Transport

Stdio transport only. Launched as a subprocess by the MCP client:

```
runefact mcp
```

The server reads the project root from the nearest `runefact.toml` (walking up from CWD), identical to how the CLI finds the project. All paths in tool inputs/outputs are relative to the project root.

### Client Configuration

**Claude Desktop (`claude_desktop_config.json`):**

```json
{
  "mcpServers": {
    "runefact": {
      "command": "runefact",
      "args": ["mcp"],
      "cwd": "/path/to/your/game"
    }
  }
}
```

**Claude Code (`.claude/mcp.json` in project root):**

```json
{
  "mcpServers": {
    "runefact": {
      "command": "runefact",
      "args": ["mcp"]
    }
  }
}
```

**Cursor / other MCP clients:** equivalent configuration pointing at `runefact mcp` as a stdio command.

### Tools

The MCP server exposes these tools. Each tool returns structured JSON, not human-readable text, so the agent can act on results programmatically.

---

**`runefact_build`** -- Compile assets

```
Input:
  scope: "all" | "sprites" | "maps" | "audio"    (optional, default "all")
  files: string[]                                  (optional, specific files to build)

Output:
  success: boolean
  artifacts: [                                     (list of produced files)
    { path: "sprites/player.png", type: "spritesheet", size_bytes: 2048 },
    { path: "audio/jump.wav", type: "wav", duration_ms: 150 },
    ...
  ]
  errors: [                                        (empty on success)
    { file: "sprites/player.sprite", line: 14, column: 8, message: "ragged row: expected 16 chars, got 14", severity: "error" },
    ...
  ]
  warnings: [
    { file: "sounds/explosion.sfx", message: "peak output at -0.2 dBFS, close to clipping", severity: "warning" },
    ...
  ]
  manifest_path: "build/assets/manifest.go"
```

---

**`runefact_validate`** -- Validate rune files without building

```
Input:
  files: string[]                                  (optional, default all files)

Output:
  valid: boolean
  errors: [...]                                    (same format as build)
  warnings: [...]
```

---

**`runefact_inspect_sprite`** -- Get metadata about a compiled sprite sheet

```
Input:
  file: "sprites/player.sprite"                    (source rune file)

Output:
  sheet: "build/assets/sprites/player.png"
  sheet_width: 64
  sheet_height: 48
  sprites: [
    { name: "idle", x: 0, y: 0, width: 16, height: 16, frames: 2, fps: 4 },
    { name: "walk", x: 0, y: 16, width: 16, height: 16, frames: 4, fps: 8 },
    { name: "jump", x: 0, y: 32, width: 16, height: 16, frames: 1, fps: 0 },
    { name: "heart", x: 64, y: 0, width: 8, height: 8, frames: 1, fps: 0 },
    { name: "coin", x: 64, y: 8, width: 8, height: 8, frames: 4, fps: 6 },
  ]
  palette: { "_": "transparent", "k": "#000000", "s": "#ffccaa", ... }
```

---

**`runefact_inspect_map`** -- Get metadata about a compiled map

```
Input:
  file: "maps/level1.map"

Output:
  json_path: "build/assets/maps/level1.json"
  tile_size: 16
  width: 32
  height: 18
  layers: [
    { name: "sky", type: "tile", scroll_x: 0.2 },
    { name: "terrain", type: "tile" },
    { name: "hazards", type: "tile" },
    { name: "objects", type: "entity", entity_count: 9 },
  ]
  tileset_keys: ["S", "C", "G", "D", "R", "W", "_"]
  entity_types: ["spawn", "coin", "heart", "enemy:slime", "exit"]
```

---

**`runefact_inspect_audio`** -- Get metadata about a compiled audio file

```
Input:
  file: "sounds/jump.sfx"                         (or "music/theme.track")

Output:
  wav_path: "build/assets/audio/jump.wav"
  duration_ms: 150
  sample_rate: 44100
  peak_db: -3.2
  clipping: false
  voices: 1                                        (for .sfx)
  channels: null                                   (for .sfx, populated for .track)
  patterns: null                                   (for .sfx, populated for .track)
  loop: false

  # For .track files additionally:
  # tempo: 140
  # channels: ["melody", "bass", "kick", "hihat"]
  # patterns: ["intro", "main"]
  # total_patterns: 4
  # loop: true
  # loop_start: 0
```

---

**`runefact_list_assets`** -- List all rune files in the project

```
Input:
  type: "all" | "palette" | "sprite" | "map" | "instrument" | "sfx" | "track"  (optional)

Output:
  assets: [
    { path: "palettes/default.palette", type: "palette" },
    { path: "sprites/player.sprite", type: "sprite" },
    { path: "sprites/tiles.sprite", type: "sprite" },
    { path: "sprites/enemies.sprite", type: "sprite" },
    { path: "maps/level1.map", type: "map" },
    { path: "sounds/jump.sfx", type: "sfx" },
    { path: "sounds/coin.sfx", type: "sfx" },
    { path: "music/theme.track", type: "track" },
    ...
  ]
  build_state: "stale" | "current" | "never_built"
  last_build: "2026-02-22T15:30:00Z"              (null if never built)
```

---

**`runefact_palette_colors`** -- Get resolved palette for a file

```
Input:
  file: "sprites/enemies.sprite"                   (any file that uses a palette)

Output:
  base_palette: "default"
  colors: {
    "_": { "hex": "transparent", "r": 0, "g": 0, "b": 0, "a": 0 },
    "k": { "hex": "#000000", "r": 0, "g": 0, "b": 0, "a": 255 },
    "sg": { "hex": "#44cc44", "r": 68, "g": 204, "b": 68, "a": 255, "source": "palette_extend" },
    ...
  }
  available_keys: ["_", "k", "w", "r", "b", "g", "d", "s", "h", "y", "o", "p", "l", "c", "e", "f", "n", "sg", "sd", "se"]
```

This is critical for the agent: when it needs to draw a sprite, it can query available palette keys and their colors so it picks the right ones. Without this, the agent guesses at key names and gets validation errors.

---

**`runefact_scaffold`** -- Generate skeleton rune files

```
Input:
  type: "sprite" | "map" | "sfx" | "track" | "instrument" | "palette"
  name: "goblin"                                   (asset name)
  options: {                                       (type-specific, all optional)
    grid: 16,                                      (sprite: pixel dimensions)
    palette: "default",                            (sprite: which palette)
    animated: true,                                (sprite: include animation frames)
    frame_count: 4,                                (sprite: number of frames)
    tile_size: 16,                                 (map: tile dimensions)
    width: 20,                                     (map: map width in tiles)
    height: 12,                                    (map: map height in tiles)
    waveform: "square",                            (instrument/sfx: oscillator type)
    tempo: 120,                                    (track: BPM)
    channels: ["melody", "bass", "drums"],         (track: channel names)
  }

Output:
  path: "sprites/goblin.sprite"                    (where the file was created)
  content: "..."                                   (full file content that was written)
```

The scaffold tool writes a valid, buildable file with placeholder content (e.g., a sprite filled with a single color, a silent SFX, an empty pattern). This gives the agent a structurally correct starting point it can then edit, avoiding format errors on the first attempt.

---

**`runefact_format_help`** -- Get format documentation for a file type

```
Input:
  type: "sprite" | "map" | "sfx" | "track" | "instrument" | "palette"

Output:
  format_spec: "..."                               (Markdown documentation for the format)
  example: "..."                                   (complete example file content)
  fields: [                                        (structured field reference)
    { name: "palette", type: "string", required: false, description: "Reference to .palette file" },
    { name: "grid", type: "int | string", required: false, description: "Pixel dimensions per frame. Integer for square, 'WxH' for non-square." },
    ...
  ]
```

### Resources

The MCP server also exposes resources (read-only context the agent can pull in).

**`runefact://project/status`** -- Current project state

```json
{
  "project_name": "rune-knight",
  "root": "/home/vadim/games/rune-knight",
  "asset_count": { "palette": 1, "sprite": 3, "map": 1, "instrument": 5, "sfx": 6, "track": 2 },
  "build_state": "current",
  "last_build": "2026-02-22T15:30:00Z",
  "errors": []
}
```

**`runefact://formats/{type}`** -- Format specification for a file type

Returns the complete Markdown documentation for the requested format type (`palette`, `sprite`, `map`, `instrument`, `sfx`, `track`). This is the same content served by `runefact_format_help` but available as a resource the agent can subscribe to.

**`runefact://palette/{name}`** -- Resolved palette colors

Returns the full color map for a named palette, including all single-char and multi-char keys. Agents can reference this to pick correct keys when generating pixel grids.

**`runefact://manifest`** -- Current manifest data

Returns the same metadata that goes into `manifest.go` but as JSON. The agent can use this to know what sprites/maps/audio already exist without reading the generated Go file.

### Design Principles

- **Tools for actions, resources for context.** If the agent needs to do something (build, validate, scaffold), it calls a tool. If it needs to know something (what palette keys exist, what the format spec is), it reads a resource.
- **Structured output always.** Every tool returns JSON the agent can parse, never human-readable prose. Error messages are in structured arrays with file/line/message, not sentences.
- **Idempotent where possible.** `runefact_build` with the same inputs produces the same outputs. `runefact_scaffold` overwrites if file exists (with a `overwrite: false` option to prevent this).
- **Fail fast, fail clearly.** If `runefact.toml` isn't found, every tool returns a clear error with fix instructions. No silent failures.
- **No side effects from reads.** Inspect and list tools never modify files. Only `build` and `scaffold` write to disk.

---

## Usage Documentation

### Documentation Structure

All docs live in `docs/` in the Runefact repo and are installed alongside the binary. They are plain Markdown, designed to be both human-readable on GitHub/website and directly includable as AI agent context.

### `docs/getting-started.md`

Covers:

**Installation:**
- `go install github.com/runefact/runefact@latest`
- Binary releases for macOS (arm64, amd64), Linux (amd64), Windows (amd64)
- Homebrew tap: `brew install runefact/tap/runefact`
- Verification: `runefact --version`

**Quickstart (5-minute tutorial):**
1. `mkdir my-game && cd my-game`
2. `runefact init` -- scaffolds project structure with demo assets
3. `runefact build` -- compiles everything, produces `build/assets/`
4. `runefact preview` -- opens previewer, browse the demo assets
5. Edit `assets/sprites/player.sprite` in your editor -- see live update in previewer
6. Overview of what was generated: sprite sheets, JSON maps, WAV files, `manifest.go`

**Project layout explanation** with annotated tree view.

### `docs/format-reference.md`

The canonical specification for all file formats. This is the source of truth that both the CLI parser and the VS Code extension validate against. Structured as:

For each format (`.palette`, `.sprite`, `.map`, `.inst`, `.sfx`, `.track`):
- Purpose and when to use
- Complete TOML schema with every field, its type, whether it's required, default value, and allowed values
- Minimal valid example
- Full-featured example exercising all options
- Common mistakes and how to fix them

This file is intentionally verbose and redundant. It's optimized for an AI agent doing a `cat docs/format-reference.md` before writing a rune file.

### `docs/sprite-guide.md`

In-depth guide to sprite authoring:

- Designing a palette: color theory basics for pixel art, recommended palette sizes, how to structure palette files for a cohesive game look
- Anatomy of a pixel grid: how to read and write them, the single-char vs bracket `[xx]` syntax, whitespace rules
- Static vs animated sprites: when to use each, framerate guidelines (4 FPS for breathing, 8 for walking, 12 for fast actions)
- Sprite sheet layout: how Runefact packs them, how to predict what the output PNG looks like
- Working with the previewer: workflow for iterating on sprites
- Tips for AI agents: how to describe sprites to Claude so it generates good pixel grids (e.g., "draw a 16x16 knight facing right with a blue tunic and silver helmet, using palette 'default'")

### `docs/map-guide.md`

In-depth guide to map authoring:

- Tileset design: how to plan tiles that connect properly (edges, corners)
- Layer strategy: when to use tile layers vs entity layers, parallax guidelines
- Entity placement: types, properties, how your game code interprets them
- Large maps: performance considerations, how the previewer handles them
- Tips for AI agents: how to describe levels to Claude (e.g., "create a 40x12 underground cave level with platforms, a lava pit in the middle, and 5 coin pickups along the jumping path")

### `docs/audio-guide.md`

In-depth guide to sound and music authoring:

- Waveform primer: what sine, square, triangle, sawtooth, noise, and pulse sound like, with parameter guidance
- ADSR envelopes explained: attack for plucks vs pads, decay for punch, sustain for held notes, release for tails
- Designing SFX by category: jumps (upward pitch sweep), coins (two-tone ding), explosions (noise + sub-bass), hits (short + harsh), powerups (ascending sweep), UI sounds (short sine blips)
- Filter basics: lowpass for warmth/distance, highpass for brightness/hats, bandpass for telephony/radio
- Tracker format explained: how patterns work, column alignment, note format, sustain/silence/note-off
- Song structure: intro patterns, verse/chorus, using the sequence array for arrangement
- Instrument design: matching instruments to roles (lead, bass, drums, pads)
- Tips for AI agents: how to describe sounds to Claude (e.g., "make the jump sound snappier -- shorter attack, higher end frequency, exponential curve"), how to iterate on music ("the verse feels empty -- add a pad channel playing sustained chords under the melody")

### `docs/ebitengine-integration.md`

How to use compiled artifacts in an ebitengine game:

- Loading the manifest: `import "your-game/build/assets"`
- Loading sprite sheets: `ebiten.NewImageFromFile(assets.SpriteSheetPlayer)`
- Drawing sprites: using `SpriteInfo` to calculate source rectangles
- Animating sprites: frame counter + `SpriteInfo.FPS` to advance frames
- Loading maps: `json.Unmarshal` the JSON tilemap, rendering tile layers
- Playing audio: `audio.NewPlayer` with WAV files, managing SFX vs music
- Complete minimal game example: a character that walks, jumps, collects coins, with all assets loaded from Runefact output
- The example should be a buildable `main.go` that imports the generated manifest

### `docs/ai-workflows.md`

The playbook for using Runefact with AI agents:

**Setup:**
1. Install Runefact
2. Configure MCP server in your AI client (Claude Code, Claude Desktop, Cursor)
3. Open the previewer in a side window: `runefact preview`
4. Start chatting with the agent

**Workflow patterns:**

*"I need a new enemy sprite"*
1. Agent calls `runefact_palette_colors` to see available colors
2. Agent calls `runefact_scaffold` with `type: "sprite", name: "bat", options: { grid: 16, animated: true, frame_count: 4 }`
3. Agent edits the scaffold file, filling in pixel grids
4. Agent calls `runefact_build` with `files: ["sprites/bat.sprite"]`
5. If errors, agent reads them and fixes the file. If success, you see it in the previewer.
6. You say "make the wings wider" -- agent edits the pixel grid, rebuilds

*"Make a new level"*
1. Agent calls `runefact_list_assets` with `type: "sprite"` to see available tilesets
2. Agent calls `runefact_inspect_sprite` on the tileset to see available tile names
3. Agent calls `runefact_scaffold` with `type: "map", name: "level2", options: { width: 40, height: 15 }`
4. Agent fills in the tileset references and layer grids
5. Agent calls `runefact_build`, you see the level in the previewer
6. You say "add more platforms on the right side" -- agent edits the grid, rebuilds

*"The jump sound is too weak"*
1. Agent calls `runefact_inspect_audio` on `jump.sfx` to see current parameters
2. Agent reads the `.sfx` file, adjusts pitch end frequency and decay
3. Agent calls `runefact_build` with `files: ["sounds/jump.sfx"]`
4. You press Enter in the previewer to hear it
5. Iterate until it sounds right

*"Add background music"*
1. Agent calls `runefact_format_help` with `type: "track"` to refresh its memory on the format
2. Agent calls `runefact_list_assets` with `type: "instrument"` to see available instruments
3. Agent creates the `.track` file from scratch (or scaffolds first)
4. Agent calls `runefact_build`, you listen in the previewer
5. "Make the chorus more energetic" -- agent adds notes, increases velocity, rebuilds

**Key principles for agents:**
- Always call `runefact_palette_colors` before drawing sprites -- don't guess at key names
- Always call `runefact_validate` after editing to catch errors before building
- Use `runefact_scaffold` for new files rather than writing from scratch -- avoids format errors
- Read `runefact_format_help` if you haven't worked with a format recently -- the format spec is your reference
- Build incrementally: `files: ["the-file-you-just-edited"]` is faster than building everything

### `docs/mcp-reference.md`

Complete reference for the MCP server:

- Every tool with full input schema, output schema, and example call/response
- Every resource with URI pattern and response format
- Error codes and what they mean
- Rate limiting / performance notes (build is synchronous, inspect is fast)

### `docs/CLAUDE.md` -- Agent Context File

This is the most important doc for AI integration. It's a single file designed to be included in Claude Code's context (via `CLAUDE.md` convention) or manually fed to any agent. It contains:

```markdown
# Runefact - AI Agent Context

You are working on a project that uses Runefact, a text-based asset engine
for ebitengine games. Assets are defined as human-readable text files ("runes")
and compiled to PNG sprite sheets, JSON tilemaps, and WAV audio files ("artifacts").

## Quick Reference

### Available MCP tools (if runefact MCP server is configured):
- `runefact_build` -- compile assets
- `runefact_validate` -- check for errors
- `runefact_inspect_sprite` -- get sprite sheet metadata
- `runefact_inspect_map` -- get map metadata
- `runefact_inspect_audio` -- get audio metadata
- `runefact_list_assets` -- list all asset files
- `runefact_palette_colors` -- get available palette colors
- `runefact_scaffold` -- create skeleton asset files
- `runefact_format_help` -- get format documentation

### File formats:
- `.palette` -- color palette definitions (TOML)
- `.sprite` -- pixel art sprites and animations (TOML + pixel grids)
- `.map` -- tilemaps with layers (TOML + tile grids)
- `.inst` -- synthesizer instrument definitions (TOML)
- `.sfx` -- procedural sound effects (TOML)
- `.track` -- tracker-style music (TOML + pattern grids)

### Key rules:
1. Always check available palette keys before writing pixel grids
2. Pixel grid rows must all have the same width
3. All animation frames must have the same dimensions
4. Use `_` for transparent pixels
5. Multi-char palette keys use bracket syntax: `[sk]`
6. Tracker patterns: `---` = sustain, `...` = silence, `^^^` = note off
7. Audio is never auto-played in the previewer -- user must press Enter

### Workflow:
1. Use `runefact_scaffold` to create new files (avoids format errors)
2. Edit the file content
3. Use `runefact_validate` to check for errors
4. Use `runefact_build` to compile
5. User checks result in the previewer

For full format specs, call `runefact_format_help` with the file type.
```

This file is deliberately concise -- it's the cheat sheet, not the manual. The agent can call `runefact_format_help` for the full spec when it needs it.

---

## Implementation Plan

### Phase 1: Sprites (MVP)
- TOML parser for `.palette` and `.sprite` files
- Grid parser (single-char and bracket `[xx]` multi-char)
- PNG sprite sheet renderer using Go `image` stdlib
- Manifest generator (sprites section)
- `runefact build --sprites` and `runefact validate`

**DoD:** Given a `.palette` and `.sprite` file, produces a correct PNG sprite sheet and valid `manifest.go`. All validation errors surface with file/line references.

### Phase 2: Previewer (Sprites)
- Ebitengine-based window with sprite rendering
- fsnotify file watcher triggering rebuild on `.sprite` / `.palette` changes
- Sprite grid layout, animation playback, zoom, grid overlay, background toggle
- Inline error display on build failure
- State persistence

**DoD:** Editing a `.sprite` file in a text editor causes the previewer to live-update within 500ms. Animations play. Zoom works pixel-perfectly. Build errors show inline without crashing.

### Phase 3: Maps
- TOML parser for `.map` files
- Tileset resolution (validate sprite references exist)
- JSON tilemap output
- Entity layer parsing
- Manifest generator (maps section)
- Map preview mode in previewer (pan, zoom, layer toggle)

**DoD:** Given `.map` files referencing existing sprites, produces JSON tilemaps with correct tile indices and entity data. Map previewer renders composited layers with pan/zoom.

### Phase 4: Audio - SFX
- TOML parser for `.inst` and `.sfx` files
- Waveform generators: sine, square, triangle, sawtooth, noise, pulse
- ADSR envelope
- Pitch sweep (linear, exponential, logarithmic curves)
- Low-pass / high-pass / band-pass filter
- Multi-voice mixing
- WAV output (PCM, 16-bit, 44100Hz default)
- SFX preview mode: waveform display, manual playback, brickwall limiter, peak meter

**DoD:** Given `.inst` and `.sfx` files, produces playable WAV files. A `jump.sfx` sounds recognizably like a jump. Previewer plays SFX on Enter with limiter active and no auto-play.

### Phase 5: Audio - Music
- TOML parser for `.track` files
- Pattern parser (column-aligned tracker format)
- Note-to-frequency conversion
- Per-tick sequencer
- Multi-channel mixing with volume
- Per-note effects (volume, pitch slide, vibrato, arpeggio)
- Loop support
- WAV output
- Music preview mode: tracker view, transport controls, pattern navigation

**DoD:** Given `.track` and `.inst` files, produces a playable WAV that sounds like intended music. Loops cleanly. Previewer shows scrolling tracker view with transport controls. Audio safety: limiter, no auto-play, stop-on-change.

### Phase 6: VS Code Extension - Syntax Highlighting
- TextMate injection grammars for all 6 file types on top of TOML base
- File icon theme entries
- Inline color decorators for `.palette` and `.sprite` hex values
- Pixel grid background coloring in `.sprite` files (palette-resolved)
- Tracker note coloring in `.track` pattern blocks (chromatic pitch-class colors)
- Snippets for all file types
- Command palette integration (build, validate, preview)

**DoD:** Opening a `.sprite` file in VS Code shows color-coded pixel grids that approximate the actual sprite. `.track` pattern blocks show chromatically colored notes. All TOML structure is properly highlighted. Snippets scaffold valid file skeletons.

### Phase 7: VS Code Extension - Inline Validation
- `DocumentDiagnosticProvider` for structural validation of all file types
- Cross-file palette resolution (read `.palette` files from workspace)
- Cross-file reference validation (sprite refs in `.map`, instrument refs in `.track`)
- Squiggly underlines + Problems panel integration
- Settings for toggling features on/off

**DoD:** Saving a `.sprite` file with a ragged row shows a red underline on the bad row. Unknown palette keys get warnings. Unknown pattern names in `.track` sequence arrays get errors. All diagnostics include actionable messages. Zero false positives on valid files from `runefact init`.

### Phase 8: Usage Documentation
- `docs/getting-started.md` -- install + quickstart tutorial
- `docs/format-reference.md` -- canonical format spec for all file types
- `docs/sprite-guide.md` -- sprite authoring deep dive
- `docs/map-guide.md` -- map authoring deep dive
- `docs/audio-guide.md` -- SFX and music authoring deep dive
- `docs/ebitengine-integration.md` -- loading artifacts in your game, with buildable example `main.go`
- `runefact docs` command that prints the path to the docs directory

**DoD:** A developer who has never seen Runefact can go from `runefact init` to a running ebitengine game with custom sprites, a level, and sound effects by following the docs alone. `docs/format-reference.md` is complete and accurate enough that the VS Code extension and CLI parser can both be verified against it.

### Phase 9: MCP Server
- `runefact mcp` subcommand launching stdio MCP server
- All 8 tools implemented: build, validate, inspect_sprite, inspect_map, inspect_audio, list_assets, palette_colors, scaffold, format_help
- All 4 resources implemented: project/status, formats/{type}, palette/{name}, manifest
- `docs/ai-workflows.md` -- agent playbook with workflow patterns
- `docs/mcp-reference.md` -- complete MCP tool/resource reference
- `docs/CLAUDE.md` -- agent context file

**DoD:** Configure the MCP server in Claude Code, start a new conversation, say "create a goblin enemy sprite using the default palette." Claude calls palette_colors, scaffold, edits the file, calls build. The sprite appears in the previewer. No manual intervention beyond approving tool calls. Repeat for "add a forest level" and "create a coin pickup sound."

### Phase 10: DX Polish
- `runefact watch` (headless file watcher, rebuild only, no preview window -- for CI or background use)
- `runefact init` (project scaffolding with example files for each asset type)
- Better error messages with line numbers, column numbers, and fix suggestions
- Performance: parallel compilation, file-level caching (skip unchanged files)
- Asset browser mode in previewer (tree view of all project files)

---

## Edge Cases and Failure Modes

**Sprites:**
- Ragged rows (different widths) -- error with line number
- Unknown palette key -- error naming the key and suggesting similar keys (Levenshtein)
- Frame size mismatch in animation -- error showing expected vs actual dimensions
- Empty sprite (all transparent) -- warning, still generate

**Maps:**
- Reference to nonexistent sprite -- error with suggestion
- Inconsistent row widths -- error with line number
- Layer with no data -- warning

**Audio:**
- Frequency outside audible range (< 20Hz or > 20kHz) -- clamp and warn
- Envelope longer than duration -- clamp release, warn
- Missing instrument reference -- error
- Pattern column count doesn't match channel count -- error with both counts
- Note parsing failure -- error with pattern name, line number, column
- NaN/Inf samples in output -- replace with silence, error with voice identification
- DC offset in output -- auto-corrected by previewer high-pass, warn during build
- Resonance blowup (filter self-oscillation) -- detected by peak analysis, warn

**Previewer:**
- Build fails on file change -- show error overlay, keep displaying last good state
- File deleted while being previewed -- show "file removed" message, fall back to asset browser
- Audio device unavailable -- previewer still works for visual assets, audio controls grayed out with message
- Very large map (>1000x1000 tiles) -- viewport culling, only render visible tiles

**General:**
- Missing `runefact.toml` -- error with `runefact init` suggestion
- Circular palette references -- error
- File encoding issues -- require UTF-8, error otherwise
- Output directory not writable -- error early, before doing any work
- Concurrent file writes (agent writing multiple files) -- debounce watcher by 100ms, batch rebuilds

**VS Code Extension:**
- Large pixel grids (64x64+) with palette coloring -- debounce decoration updates by 200ms, skip coloring if grid exceeds 128x128
- Palette file not found in workspace -- degrade gracefully, skip palette-dependent coloring, show info message
- Multiple `runefact.toml` files in workspace (monorepo) -- use nearest parent `runefact.toml` for each file
- Extension loaded but `runefact` CLI not installed -- syntax highlighting and validation still work, build/preview commands show install prompt
- VS Code fork compatibility (Cursor, Windsurf) -- no APIs beyond standard Extension API, no webview panels

**MCP Server:**
- `runefact.toml` not found -- all tools return clear error: `{ "error": "no_project", "message": "No runefact.toml found. Run 'runefact init' to create a project." }`
- Build already in progress (concurrent tool calls) -- serialize builds with a mutex, second call waits. Return error if wait exceeds 30s.
- Agent calls `runefact_inspect_*` before ever building -- return `{ "error": "not_built", "message": "No build artifacts found. Call runefact_build first." }`
- Agent scaffolds a file that already exists -- `overwrite` defaults to false, returns error. Agent must pass `overwrite: true` explicitly.
- Agent sends invalid tool input (wrong types, missing required fields) -- return JSON Schema validation error with the specific field that failed
- MCP client disconnects mid-build -- build completes, results are discarded (no partial state)
- Very large project (hundreds of asset files) -- `runefact_list_assets` is always fast (just file enumeration). `runefact_build` with `scope: "all"` may be slow; prefer targeted builds with `files` parameter.
- Agent calls `runefact_format_help` -- always works, even without a valid project. This is pure documentation, no file I/O.

---

## Dependencies

**Core (no CGo):**
- `image`, `image/png`, `image/color` -- sprite rendering
- `encoding/json` -- map output
- `os`, `path/filepath` -- file I/O
- `math` -- audio synthesis
- `encoding/binary` -- WAV writing
- `github.com/BurntSushi/toml` (or `github.com/pelletier/go-toml/v2`) -- TOML parsing

**Previewer (CGo OK):**
- `github.com/hajimehoshi/ebiten/v2` -- preview window rendering
- `github.com/hajimehoshi/oto/v2` (via ebiten audio) -- audio playback in previewer
- `github.com/fsnotify/fsnotify` -- file watching

**MCP Server:**
- `github.com/mark3labs/mcp-go` (or equivalent Go MCP SDK) -- MCP protocol handling, stdio transport
- No additional dependencies beyond core; the MCP server reuses all internal packages

**VS Code Extension (TypeScript):**
- VS Code Extension API (`@types/vscode`)
- `vscode-languageclient` (only if LSP is introduced later)
- TextMate grammar JSON files (no runtime deps)
- Published via `vsce` CLI to VS Code Marketplace

**Build tool itself writes WAV files directly (PCM encoding is trivial). No audio playback libraries needed for `runefact build`.**

---

---

## Demo Assets

A complete set of demo rune files shipped with `runefact init` and used as the acceptance test suite for each implementation phase. These represent a tiny but complete platformer called "Rune Knight" -- enough to exercise every format feature and validate the full pipeline.

All files below go into `assets/` in the scaffolded project.

### Phase 1 Demo: Palettes + Sprites

**`palettes/default.palette`**

```toml
name = "default"

[colors]
_  = "transparent"
k  = "#000000"       # black / outlines
w  = "#ffffff"       # white / highlights
r  = "#ff004d"       # red
b  = "#29adff"       # blue
g  = "#00e436"       # green
d  = "#1d2b53"       # dark blue
s  = "#ffccaa"       # skin
h  = "#ab5236"       # hair / brown
y  = "#ffec27"       # yellow
o  = "#ffa300"       # orange
p  = "#7e2553"       # purple
l  = "#83769c"       # lavender
c  = "#008751"       # dark green
e  = "#5f574f"       # dark grey
f  = "#c2c3c7"       # light grey
n  = "#1d2b53"       # navy
```

**`sprites/player.sprite`** -- tests static sprites, animated sprites, per-sprite grid override

```toml
palette = "default"
grid = 16

# --- Idle animation: 2 frames, breathing motion ---
[sprite.idle]
framerate = 4

[[sprite.idle.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
____kbbbbbbk____
_____kbbbbk_____
_____kk__kk_____
____kdk__kdk____
____kkk__kkk____
"""

[[sprite.idle.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
_____kbbbbk_____
_____kbbbbk_____
_____kk__kk_____
____kdk__kdk____
____kkk__kkk____
"""

# --- Walk animation: 4 frames ---
[sprite.walk]
framerate = 8

[[sprite.walk.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
____kbbbbbbk____
_____kb__bbk____
____kdk___kk____
____kkk__kdk____
_________kkk____
"""

[[sprite.walk.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
____kbbbbbbk____
_____kbbbbk_____
_____kk__kk_____
____kdk__kdk____
____kkk__kkk____
"""

[[sprite.walk.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
____kbbbbbbk____
____kbb__bk_____
____kk___kdk____
____kdk__kkk____
____kkk_________
"""

[[sprite.walk.frame]]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
_____kbbbbk_____
____kbbbbbbk____
____kbkbbkbk____
____kbbbbbbk____
____kbbbbbbk____
_____kbbbbk_____
_____kk__kk_____
____kdk__kdk____
____kkk__kkk____
"""

# --- Jump: single static frame ---
[sprite.jump]
pixels = """
______kkkk______
_____kssssk_____
____kssssssk____
____kskkskksk___
____kssssssk____
_____kssssk_____
______kbbk______
____kkbbbbkk____
___kbbbbbbbbk___
___kbkbbbbkbk___
____kbbbbbbk____
_____kbbbbk_____
____kk_kk_kk____
___kdk_kk_kdk___
___kkk____kkk___
________________
"""

# --- Small heart pickup: tests per-sprite grid override ---
[sprite.heart]
grid = 8
pixels = """
_kk__kk_
krrkrrrk
krrrrrrrk
krrrrrrrk
_krrrrk_
__krrrk_
___krk__
____k___
"""

# --- Coin: small animated sprite ---
[sprite.coin]
grid = 8
framerate = 6

[[sprite.coin.frame]]
pixels = """
__kkkk__
_kyyyyk_
kyyyyyyk
kyyyyyyk
kyyyyyyk
kyyyyyyk
_kyyyyk_
__kkkk__
"""

[[sprite.coin.frame]]
pixels = """
___kk___
__kyyk__
__kyyk__
__kyyk__
__kyyk__
__kyyk__
__kyyk__
___kk___
"""

[[sprite.coin.frame]]
pixels = """
____k___
____k___
____k___
____k___
____k___
____k___
____k___
____k___
"""

[[sprite.coin.frame]]
pixels = """
___kk___
__kyyk__
__kyyk__
__kyyk__
__kyyk__
__kyyk__
__kyyk__
___kk___
"""
```

**`sprites/tiles.sprite`** -- tests multiple static sprites in one file

```toml
palette = "default"
grid = 16

[sprite.grass]
pixels = """
___cg_cg__gc_gc_
__cgcgcgcgcgcgc_
_gcgcgcgcgcgcgcg
cgcgcgcgcgcgcgcg
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
"""

[sprite.dirt]
pixels = """
hhhhhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
hhhhehhhhhhhhhhh
hhhhhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhhhehhhhhhh
hhhhhhhhhhhhhhhh
hhehhhhhhhhhhhhh
hhhhhhhhhhhhhehh
hhhhhhhhhhhhhhhh
hhhhhehhhhhhhhhe
hhhhhhhhhhhhhhhh
hhhhhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhehhhhhhhhhhhhh
hhhhhhhhhhhhhhhh
"""

[sprite.stone]
pixels = """
keeeeeeeeeeeeefk
keeeeeeeeeeeeeek
eeeeeeeeeeeeeefe
eefeeeeeeeeeeeee
eeeeeeeeeefeeeee
eeeeeeeeeeeeeeee
eeeeeefeeeeeeeee
eeeeeeeeeeeeefee
eeeeeeeeeeeeeeee
eeeeeeeeeeeeeefe
eeefeeeeeeeeeeee
eeeeeeeeeeeeeeee
eeeeeeeeefeeeeee
feeeeeeeeeeeeeee
eeeeeeeeeeeeeeee
keeeeeeeeeeeeefk
"""

[sprite.sky]
pixels = """
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
dddddddddddddddd
"""

[sprite.cloud]
pixels = """
________________
____wwww________
__wwwwwwww______
_wwwwwwwwwww____
wwwwwwwwwwwww___
_wwwwwwwwwwwww__
__wwwwwwwwwwww__
________________
________________
________________
________________
________________
________________
________________
________________
________________
"""

[sprite.water]
framerate = 3

[[sprite.water.frame]]
pixels = """
bbbbbbbbbbbbbbbb
bbbbbbbbbbbblbbb
bbblbbbbbbbbbbbb
bbbbbbbbblbbbbbb
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
"""

[[sprite.water.frame]]
pixels = """
bbbbbbbbbbbbbbbb
bbbblbbbbbbbbbbb
bbbbbbbbblbbbbbb
blbbbbbbbbbbblbb
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
nnnnnnnnnnnnnnnn
"""
```

**`sprites/enemies.sprite`** -- tests multi-char palette keys

```toml
palette = "default"
grid = 16

[sprite.slime]
palette_extend = { sg = "#44cc44", sd = "#228822", se = "#116611" }
framerate = 3

[[sprite.slime.frame]]
pixels = """
________________
________________
________________
________________
________________
________________
____[sg][sg][sg][sg][sg][sg][sg][sg]____
__[sg][sg][sg][sg][sg][sg][sg][sg][sg][sg]__
_[sg][sg][sd][sd][sg][sg][sd][sd][sg][sg]_
[sg][sg][sd]kk[sd][sg][sd]kk[sd][sg][sg]
[sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg]
[sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg]
_[se][sg][sg][sg][sg][sg][sg][sg][sg][sg][se]_
__[se][se][se][se][se][se][se][se][se][se]__
________________
________________
"""

[[sprite.slime.frame]]
pixels = """
________________
________________
________________
________________
________________
________________
________________
____[sg][sg][sg][sg][sg][sg][sg][sg]____
__[sg][sg][sg][sg][sg][sg][sg][sg][sg][sg]__
_[sg][sg][sd][sd][sg][sg][sd][sd][sg][sg]_
[sg][sg][sd]kk[sd][sg][sd]kk[sd][sg][sg]
[sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg]
[se][sg][sg][sg][sg][sg][sg][sg][sg][sg][sg][se]
[se][se][se][se][se][se][se][se][se][se][se][se]
________________
________________
"""
```

**Phase 1 acceptance tests:**
- `runefact build --sprites` produces `player.png`, `tiles.png`, `enemies.png` sprite sheets
- `player.png` contains: idle (2 frames), walk (4 frames), jump (1 frame), heart (8x8), coin (4 frames at 8x8) -- all laid out correctly
- `tiles.png` contains: grass, dirt, stone, sky, cloud (static), water (2 frames)
- `enemies.png` correctly resolves multi-char palette keys `[sg]`, `[sd]`, `[se]`
- `manifest.go` has correct entries for all sprites with positions, dimensions, frame counts, FPS

### Phase 2 Demo: Previewer (Sprites)

Uses the same sprite files from Phase 1. Manual test checklist:

- Open `runefact preview player.sprite` -- all sprites visible in grid layout
- Idle and walk animations are playing
- Mouse wheel zoom from 1x to 32x, pixels stay sharp
- Click the coin sprite -- isolates it, centered, animation playing
- Press Space -- all animations pause
- Left/Right arrows step coin through its 4 frames
- Press G -- pixel grid overlay toggles
- Press B -- background cycles dark / light / checkerboard
- Edit `player.sprite` in another window (change a pixel) -- previewer updates within 500ms
- Introduce a syntax error -- red error overlay appears, last good render stays visible
- Fix the error -- preview recovers automatically

### Phase 3 Demo: Maps

**`maps/level1.map`** -- tests tile layers, parallax, entity placement

```toml
tile_size = 16

[tileset]
S = "tiles:sky"
C = "tiles:cloud"
G = "tiles:grass"
D = "tiles:dirt"
R = "tiles:stone"
W = "tiles:water"
_ = ""

[layer.sky]
scroll_x = 0.2
pixels = """
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSCSSSSSSSSSSSSSSSSCSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSCSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
SSSSSSSSSSSSSSSSSSSSSSSSSSSSSSSS
"""

[layer.terrain]
pixels = """
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
____________RRRR________________
__________RR____RR________RRR___
________RR________RR____RR______
________________________________
________________________________
GGGGGGGGGGGG____GGGGGGGGGGGGGGGG
DDDDDDDDDDDD____DDDDDDDDDDDDDDD
DDDDDDDDDDDD____DDDDDDDDDDDDDDD
DDDDDDDDDDDD____DDDDDDDDDDDDDDD
"""

[layer.hazards]
pixels = """
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
____________WWWW________________
____________WWWW________________
____________WWWW________________
____________WWWW________________
"""

[layer.objects]
[[layer.objects.entity]]
type = "spawn"
x = 2
y = 13

[[layer.objects.entity]]
type = "coin"
x = 10
y = 8

[[layer.objects.entity]]
type = "coin"
x = 11
y = 8

[[layer.objects.entity]]
type = "coin"
x = 12
y = 8

[[layer.objects.entity]]
type = "coin"
x = 13
y = 8

[[layer.objects.entity]]
type = "heart"
x = 27
y = 9

[[layer.objects.entity]]
type = "enemy:slime"
x = 20
y = 13
properties = { patrol_range = 4, direction = "left" }

[[layer.objects.entity]]
type = "enemy:slime"
x = 28
y = 9
properties = { patrol_range = 2, direction = "right" }

[[layer.objects.entity]]
type = "exit"
x = 31
y = 13
```

**Phase 3 acceptance tests:**
- `runefact build --maps` produces `level1.json`
- JSON contains 4 layers: sky, terrain, hazards, objects
- Tile indices in `data` arrays correctly reference tileset entries
- Entity layer has all 9 entities with correct types, positions, properties
- `sky` layer has `scroll_x: 0.2`
- Previewer renders composited layers, water tiles animate, entity markers visible
- Tab cycles through individual layers
- Pan/zoom works, tile grid toggleable

### Phase 4 Demo: Instruments + Sound Effects

**`instruments/sfx_basic.inst`** -- general purpose instrument for sound effects

```toml
name = "sfx_basic"

[oscillator]
waveform = "square"
duty_cycle = 0.5

[envelope]
attack = 0.0
decay = 0.1
sustain = 0.5
release = 0.1
```

**`instruments/sfx_noise.inst`** -- noise source for explosions, impacts

```toml
name = "sfx_noise"

[oscillator]
waveform = "noise"

[envelope]
attack = 0.0
decay = 0.2
sustain = 0.0
release = 0.15

[filter]
type = "lowpass"
cutoff = 3000
resonance = 0.1
```

**`sounds/jump.sfx`** -- classic platformer jump: quick upward pitch sweep

```toml
duration = 0.15
volume = 0.7

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.4
release = 0.1

[voice.pitch]
start = 220
end = 660
curve = "exponential"
```

**`sounds/coin.sfx`** -- two-tone coin ding

```toml
duration = 0.2
volume = 0.5

[[voice]]
waveform = "square"
duty_cycle = 0.25

[voice.envelope]
attack = 0.0
decay = 0.08
sustain = 0.3
release = 0.12

[voice.pitch]
start = 880
end = 1320
curve = "linear"

[[voice]]
waveform = "square"
duty_cycle = 0.25

[voice.envelope]
attack = 0.05
decay = 0.05
sustain = 0.2
release = 0.1

[voice.pitch]
start = 1320
end = 1320
curve = "linear"
```

**`sounds/hurt.sfx`** -- quick descending buzz for taking damage

```toml
duration = 0.25
volume = 0.6

[[voice]]
waveform = "sawtooth"

[voice.envelope]
attack = 0.0
decay = 0.15
sustain = 0.0
release = 0.1

[voice.pitch]
start = 440
end = 110
curve = "exponential"

[voice.effects]
vibrato_depth = 0.5
vibrato_rate = 30.0
```

**`sounds/explosion.sfx`** -- multi-voice: noise body + sine sub-bass thump

```toml
duration = 0.5
volume = 0.8

[[voice]]
waveform = "noise"

[voice.envelope]
attack = 0.0
decay = 0.3
sustain = 0.0
release = 0.2

[voice.filter]
type = "lowpass"
cutoff_start = 4000
cutoff_end = 200
curve = "exponential"

[[voice]]
waveform = "sine"

[voice.envelope]
attack = 0.0
decay = 0.15
sustain = 0.0
release = 0.1

[voice.pitch]
start = 100
end = 30
curve = "linear"
```

**`sounds/powerup.sfx`** -- ascending arpeggio-like sweep

```toml
duration = 0.4
volume = 0.5

[[voice]]
waveform = "triangle"

[voice.envelope]
attack = 0.01
decay = 0.1
sustain = 0.5
release = 0.2

[voice.pitch]
start = 330
end = 1320
curve = "logarithmic"

[voice.effects]
vibrato_depth = 0.2
vibrato_rate = 8.0
```

**`sounds/death.sfx`** -- long descending noise + tone for game over feel

```toml
duration = 0.8
volume = 0.7

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.6
sustain = 0.0
release = 0.2

[voice.pitch]
start = 440
end = 55
curve = "exponential"

[voice.effects]
vibrato_depth = 1.0
vibrato_rate = 6.0

[[voice]]
waveform = "noise"

[voice.envelope]
attack = 0.1
decay = 0.4
sustain = 0.1
release = 0.2

[voice.filter]
type = "lowpass"
cutoff_start = 2000
cutoff_end = 400
curve = "linear"
```

**Phase 4 acceptance tests:**
- `runefact build --audio` produces WAV files for all 6 SFX
- `jump.wav` is ~0.15s, audibly sweeps upward
- `coin.wav` has two distinct tones (two voices)
- `explosion.wav` has both noise and bass content
- `hurt.wav` has audible vibrato
- `powerup.wav` sweeps upward with a logarithmic curve (fast start, slow top)
- `death.wav` is ~0.8s with vibrato on the tone voice
- No WAV file contains NaN/Inf/DC offset
- All WAV files peak below 0 dBFS (no clipping)
- Previewer: Enter plays the sound once, no auto-play, limiter visible, peak meter active

### Phase 5 Demo: Music

**`instruments/lead.inst`** -- bright melody synth

```toml
name = "lead"

[oscillator]
waveform = "pulse"
duty_cycle = 0.25

[envelope]
attack = 0.01
decay = 0.15
sustain = 0.6
release = 0.2

[effects]
vibrato_depth = 0.15
vibrato_rate = 5.0
```

**`instruments/bass.inst`** -- thick bass

```toml
name = "bass"

[oscillator]
waveform = "square"
duty_cycle = 0.25

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.7
release = 0.1

[filter]
type = "lowpass"
cutoff = 600
resonance = 0.2
```

**`instruments/kick.inst`** -- kick drum via fast pitch drop

```toml
name = "kick"

[oscillator]
waveform = "sine"

[envelope]
attack = 0.0
decay = 0.15
sustain = 0.0
release = 0.05

[effects]
pitch_sweep = -80.0
```

**`instruments/hihat.inst`** -- hi-hat via short noise burst

```toml
name = "hihat"

[oscillator]
waveform = "noise"

[envelope]
attack = 0.0
decay = 0.04
sustain = 0.0
release = 0.02

[filter]
type = "highpass"
cutoff = 8000
resonance = 0.1
```

**`instruments/pad.inst`** -- soft pad for chords

```toml
name = "pad"

[oscillator]
waveform = "triangle"

[envelope]
attack = 0.2
decay = 0.3
sustain = 0.5
release = 0.4

[effects]
vibrato_depth = 0.1
vibrato_rate = 3.0
```

**`music/theme.track`** -- 8-bar loop, 4 channels, 2 patterns. Tests the full tracker pipeline.

```toml
tempo = 140
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
volume = 0.8

[[channel]]
name = "kick"
instrument = "kick"
volume = 0.9

[[channel]]
name = "hihat"
instrument = "hihat"
volume = 0.4

[pattern.intro]
ticks = 32
data = """
melody   | bass    | kick   | hihat
...      | C2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
...      | C2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
...      | E2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
...      | E2      | C2     | ...
...      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
C4       | G2      | C2     | ...
---      | ...     | ...    | F#5
D4       | ...     | ...    | ...
---      | ...     | ...    | F#5
E4       | G2      | C2     | ...
---      | ...     | ...    | F#5
D4       | ...     | ...    | ...
---      | ...     | ...    | F#5
C4       | A2      | C2     | ...
---      | ...     | ...    | F#5
...      | ...     | ...    | ...
...      | ...     | ...    | F#5
G3       | A2      | C2     | ...
---      | ...     | ...    | F#5
^^^      | ^^^     | ...    | ...
...      | ...     | ...    | F#5
"""

[pattern.main]
ticks = 32
data = """
melody   | bass    | kick   | hihat
C4       | C2      | C2     | ...
---      | ...     | ...    | F#5
E4       | ...     | ...    | ...
---      | ...     | ...    | F#5
G4       | C2      | C2     | F#5
---      | ...     | ...    | ...
E4       | ...     | ...    | F#5
---      | ...     | ...    | ...
A4       | F2      | C2     | ...
---      | ...     | ...    | F#5
G4       | ...     | ...    | ...
---      | ...     | ...    | F#5
F4       | F2      | C2     | F#5
---      | ...     | ...    | ...
E4       | ...     | ...    | F#5
---      | ...     | ...    | ...
D4       | G2      | C2     | ...
---      | ...     | ...    | F#5
E4       | ...     | ...    | ...
---      | ...     | ...    | F#5
F4       | G2      | C2     | F#5
---      | ...     | ...    | ...
E4       | ...     | ...    | F#5
---      | ...     | ...    | ...
C4       | A2      | C2     | ...
---      | ...     | ...    | F#5
D4       | ...     | ...    | ...
---      | ...     | ...    | F#5
C4       | A2      | C2     | F#5
---      | ...     | ...    | ...
^^^      | ^^^     | ...    | F#5
...      | ...     | ...    | ...
"""

[song]
sequence = [
  "intro",
  "main",
  "main",
  "intro",
]
```

**`music/gameover.track`** -- short non-looping jingle. Tests loop=false, single pattern, different tempo, per-note volume effects.

```toml
tempo = 80
ticks_per_beat = 4
loop = false

[[channel]]
name = "melody"
instrument = "pad"
volume = 0.8

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.6

[pattern.ending]
ticks = 16
data = """
melody      | bass
E4 v0F      | E2
---         | ...
D4 v0C      | ...
---         | ...
C4 v0A      | C2
---         | ...
B3 v08      | ...
---         | ...
A3 v06      | A1
---         | ...
---         | ...
---         | ...
--- v04     | ...
--- v02     | ...
--- v01     | ...
^^^         | ^^^
"""

[song]
sequence = [
  "ending",
]
```

**Phase 5 acceptance tests:**
- `runefact build --audio` produces `theme.wav` and `gameover.wav`
- `theme.wav` is a looping track (~4 patterns at 140 BPM). Audible: melody, bass line, kick, hi-hat
- `gameover.wav` is a short descending melody that does NOT loop, ends with silence
- `gameover.wav` demonstrates per-note volume effects (notes get progressively quieter)
- Bass line has low-pass filtered square character
- Kick sounds like a kick (fast pitch drop sine)
- Hi-hat sounds like a hi-hat (short filtered noise)
- `theme.wav` loops seamlessly (last sample connects to loop_start position)
- Previewer: tracker view scrolls, transport controls work, pattern skip works, no auto-play

### Phase 6+7 Demo: VS Code Extension

Uses all files above. Manual test checklist:

**Syntax highlighting:**
- Open `default.palette` -- hex colors show inline color swatches
- Open `player.sprite` -- pixel grids are color-coded matching palette (k=black, s=skin, b=blue, etc.)
- Open `enemies.sprite` -- bracket sequences `[sg]`, `[sd]`, `[se]` render with their extended palette colors
- Open `level1.map` -- tileset chars in layer grids have distinct hue-rotated backgrounds
- Open `theme.track` -- notes in pattern blocks have chromatic coloring, `---`/`...`/`^^^` are visually distinct, `|` separators are dimmed
- Open `gameover.track` -- `v0F`, `v0C` etc. highlighted as effect modifiers

**Validation:**
- In `player.sprite`, change one row to be shorter -- red squiggly on that row
- In `level1.map`, change a tileset ref to `"tiles:nonexistent"` -- warning on the ref
- In `theme.track`, add `"bogus"` to the sequence array -- error: pattern not found
- In `theme.track`, remove a `|` separator in a pattern row -- error: column count mismatch
- In `gameover.track`, change `E4` to `X4` -- error: invalid note

**Snippets:**
- In new `.sprite` file, type `rfsprite` + Tab -- scaffolds a valid static sprite
- In new `.track` file, type `rftrack` + Tab -- scaffolds a valid 2-channel track with 1 pattern

### `runefact.toml` for Demo Project

```toml
[project]
name = "rune-knight"
output = "build/assets"
package = "assets"

[defaults]
sprite_size = 16
sample_rate = 44100
bit_depth = 16

[preview]
window_width = 960
window_height = 720
background = "#1a1a2e"
pixel_scale = 4
audio_volume = 0.5
```

### Phase 8 Demo: Usage Documentation

Acceptance tests are human walkthroughs. Each doc must pass its own scenario:

**`getting-started.md` test:**
1. Fresh machine with only Go installed. No prior knowledge of Runefact.
2. Follow the doc from top to bottom.
3. End result: `runefact preview` shows the demo project with all asset types.
4. Time to complete: under 5 minutes.

**`format-reference.md` test:**
1. For each format section, copy-paste the "minimal valid example" into a new file.
2. Run `runefact validate` -- zero errors.
3. Copy-paste the "full-featured example" -- zero errors.
4. Every field listed in the schema is present in at least one example.

**`sprite-guide.md` test:**
1. Follow the "create a new sprite from scratch" section.
2. Result: a new `.sprite` file that builds successfully and looks reasonable in the previewer.
3. The guide correctly explains every option shown in the demo `player.sprite`.

**`map-guide.md` test:**
1. Follow the "create a new level" section.
2. Result: a new `.map` file that builds successfully.
3. The guide explains parallax, entity properties, and layer ordering.

**`audio-guide.md` test:**
1. Follow the "create a coin pickup sound" section.
2. Result: a `.sfx` file that produces a recognizable coin sound.
3. Follow the "create a simple loop" section.
4. Result: a `.track` file that produces audible music with correct timing.

**`ebitengine-integration.md` test:**
1. Follow the doc using the Rune Knight demo assets.
2. Copy the example `main.go` into a Go project alongside the `build/assets/` output.
3. `go run main.go` -- a window opens showing the player sprite, animated, on a tiled background. Arrow keys move. Space plays the jump sound.
4. This is the most important doc test. If this doesn't work, the entire pipeline is broken.

**`CLAUDE.md` test:**
1. Place the file in a project root.
2. Start Claude Code in that directory.
3. Ask: "What file formats does this project use?"
4. Claude should answer correctly from the CLAUDE.md context without using any tools.
5. Ask: "Create a new sprite for a mushroom enemy."
6. If MCP is configured, Claude should use the MCP tools. If not, it should create a valid `.sprite` file based on the CLAUDE.md instructions.

### Phase 9 Demo: MCP Server

The MCP demo is a scripted conversation with Claude Code (or Claude Desktop) that exercises every tool. This serves as both the acceptance test and the showcase demo.

**Setup:**
1. Rune Knight demo project fully built (Phases 1-5 complete)
2. MCP server configured in `.claude/mcp.json`:
```json
{
  "mcpServers": {
    "runefact": {
      "command": "runefact",
      "args": ["mcp"]
    }
  }
}
```
3. Previewer running in a side window: `runefact preview`

**Test script (human prompts in quotes, expected agent behavior described):**

*"What assets does this project have?"*
- Agent calls `runefact_list_assets` with no filter
- Returns all rune files grouped by type
- Agent summarizes: 1 palette, 3 sprite files, 1 map, 5 instruments, 6 SFX, 2 tracks

*"Show me what colors are available in the default palette."*
- Agent calls `runefact_palette_colors` with file referencing the default palette
- Lists all 16+ colors with their hex values
- Agent presents them readably

*"Create a new sprite for a treasure chest. It should be 16x16 with two frames: closed and open."*
- Agent calls `runefact_palette_colors` to check available colors
- Agent calls `runefact_scaffold` with `type: "sprite", name: "chest", options: { grid: 16, animated: true, frame_count: 2, palette: "default" }`
- Agent reads the scaffold, edits both frames to look like a treasure chest (brown body, yellow/gold lid)
- Agent calls `runefact_validate` on the new file
- Agent calls `runefact_build` with `files: ["sprites/chest.sprite"]`
- Build succeeds. Chest appears in previewer.

*"The chest looks too flat. Add some shading -- darker brown on the bottom half."*
- Agent reads the current `.sprite` file
- Agent edits the pixel grid, replacing `h` (brown) with `e` (dark grey) or similar in the bottom rows
- Agent calls `runefact_build`
- Updated chest appears in previewer

*"Create a 'chest open' sound effect -- a short creak then a sparkle."*
- Agent calls `runefact_format_help` with `type: "sfx"`
- Agent creates `sounds/chest_open.sfx` with two voices: a short noise creak + a high sine sparkle
- Agent calls `runefact_build` with `files: ["sounds/chest_open.sfx"]`
- Agent calls `runefact_inspect_audio` to verify duration and peak level
- Build succeeds. User presses Enter in previewer to hear it.

*"Add some treasure chests to level 1."*
- Agent calls `runefact_inspect_map` on `level1.map` to see current structure
- Agent reads the map file, adds `chest_open` and `chest_closed` entity entries at reasonable positions (on platforms)
- Agent calls `runefact_build` with `files: ["maps/level1.map"]`
- Map updates in previewer with new entity markers

*"What does the game over music sound like? Describe it."*
- Agent calls `runefact_inspect_audio` on `music/gameover.track`
- Agent reads the `.track` file
- Agent describes: "It's a short non-looping descending melody at 80 BPM, using the pad instrument. The notes E4, D4, C4, B3, A3 descend with decreasing volume. About 3 seconds long."

*"Make the theme music faster -- 160 BPM instead of 140."*
- Agent reads `music/theme.track`, changes `tempo = 140` to `tempo = 160`
- Agent calls `runefact_build` with `files: ["music/theme.track"]`
- Build succeeds. User presses Enter in previewer to hear the faster version.

**Phase 9 acceptance criteria:**
- All 8 MCP tools called successfully during the test script
- Agent never produces a malformed rune file (thanks to scaffold + format_help)
- Agent recovers from validation errors by reading the error output and fixing the file
- Agent uses `runefact_palette_colors` before generating pixel grids (never guesses at key names)
- Total test script completes in under 15 minutes including human review time
- Every artifact the agent creates builds successfully and is visible/audible in the previewer

---

## What This Enables for AI-Assisted Game Dev

The entire game asset pipeline becomes text manipulation:

- "Add a walking animation to the player sprite" -- agent edits `.sprite` file, adds frames
- "Make the background tiles darker" -- agent edits `.palette` file
- "Add a new level" -- agent creates `.map` file referencing existing tilesets
- "The jump sound should be punchier" -- agent tweaks `.sfx` parameters
- "Speed up the battle music" -- agent changes `tempo` in `.track`
- "Add a bass drop before the chorus" -- agent adds a pattern and inserts it in the sequence

All diffs are human-readable. All changes are git-friendly. No binary files to merge.

The full integration stack:

1. **CLAUDE.md** gives the agent instant context about Runefact formats and workflow
2. **MCP server** lets the agent call structured tools: query palettes, scaffold files, build, validate, inspect results -- no shell parsing, no guessing
3. **VS Code extension** provides syntax highlighting and inline validation as the agent writes
4. **Previewer** gives the human instant visual/audio feedback without leaving the editor
5. **Usage docs** serve double duty: human reference and agent context via `runefact_format_help`

The end state: you describe what you want in plain language, the agent creates the asset files, builds them, and you see/hear the result in the previewer. The feedback loop is measured in seconds, not minutes.
