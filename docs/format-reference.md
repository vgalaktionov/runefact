# Format Reference

Complete specifications for all six Runefact file formats. Every format is TOML-based with optional multiline string blocks (`"""..."""`) for grid data.

## Common Conventions

- `_` is always transparent/empty
- Single-char palette keys for readability; multi-char keys use `[xx]` bracket syntax in grids
- Grid rows must have consistent width (no ragged rows)
- Validation is lenient: warns over errors when ambiguous

---

## Palette (.palette)

Color definitions shared across sprites and maps.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | — | Palette name (referenced by sprites/maps) |
| `[colors]` | map | yes | — | Key-to-color mappings |

**Color values:** `"transparent"`, `"#RGB"`, `"#RRGGBB"`, or `"#RRGGBBAA"`

### Minimal Example

```toml
name = "mono"

[colors]
_ = "transparent"
x = "#000000"
```

### Full Example

```toml
name = "fantasy"

[colors]
_ = "transparent"
k = "#1a1c2c"
p = "#b13e53"
o = "#ef7d57"
y = "#ffcd75"
g = "#38b764"
c = "#41a6f6"
b = "#29366f"
w = "#f4f4f4"
s = "#94b0c280"
```

### Common Mistakes

- Missing `name` field — required for palette resolution
- Invalid hex format — must be `#` followed by 3, 6, or 8 hex digits
- Duplicate keys — second definition shadows the first (warning)

---

## Sprite (.sprite)

Sprite sheet definitions with optional animation frames.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `palette` | string | yes | — | Name of `.palette` file to use |
| `grid` | int or "WxH" | no | — | Default sprite dimensions |
| `[palette_extend]` | map | no | — | Additional/override palette colors |
| `[sprite.NAME]` | table | yes (1+) | — | Sprite definitions |

**Per-sprite fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `grid` | int or "WxH" | no | file default | Override dimensions |
| `framerate` | int | no | 0 (static) | Animation FPS |
| `pixels` | multiline | if no frames | — | Single-frame pixel data |
| `[[sprite.NAME.frame]]` | array | if animated | — | Animation frames |

### Grid Syntax

```
Single char:   a b c _ k
Bracket key:   [sk] [bg] [hi]
Transparent:   _
```

### Minimal Example

```toml
palette = "default"

[sprite.dot]
grid = 2
pixels = """
_x
x_
"""
```

### Full Example

```toml
palette = "fantasy"
grid = 16

[palette_extend]
skin = "#ffaa77"

[sprite.hero_idle]
pixels = """
____kkkk____
___ksssskk__
__kskin k__
__ksksksk__
___kskin k__
____kkkk____
___k[skin]k___
__kk[skin]kk__
_k__kkkk__k_
____k__k____
____kk_kk___
"""

[sprite.hero_walk]
framerate = 8

[[sprite.hero_walk.frame]]
pixels = """
____kkkk____
___ksssskk__
__kskin k__
___kskin k__
____kkkk____
___k[skin]k___
__kk__kk____
_k____k_____
___kk__kk___
"""

[[sprite.hero_walk.frame]]
pixels = """
____kkkk____
___ksssskk__
__kskin k__
___kskin k__
____kkkk____
___k[skin]k___
____kk__kk__
_____k____k_
___kk__kk___
"""
```

### Common Mistakes

- Ragged rows — all rows within a frame must have the same width
- Frame dimension mismatch — all frames in one sprite must be identical size
- Unknown palette key — check palette file and `palette_extend`
- Missing palette reference — `palette` field is required

---

## Map (.map)

Tilemaps with multiple layers and entity placement.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `tile_size` | int | yes | — | Pixel size of each tile (must be > 0) |
| `[tileset]` | map | yes | — | Char key → sprite reference mapping |
| `[layer.NAME]` | table | yes (1+) | — | Layer definitions |

**Tileset references:** `"sprite_file:sprite_name"` format.

**Tile layer fields:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `scroll_x` | float | no | 0.0 | Parallax scroll factor (horizontal) |
| `scroll_y` | float | no | 0.0 | Parallax scroll factor (vertical) |
| `pixels` | multiline | yes | — | Grid of tileset keys |

**Entity layer fields:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `[[layer.NAME.entity]]` | array | yes (1+) | Entity definitions |
| `entity.type` | string | yes | Entity type identifier |
| `entity.x` | int | yes | X position (pixels) |
| `entity.y` | int | yes | Y position (pixels) |
| `entity.properties` | map | no | Arbitrary key-value data |

### Minimal Example

```toml
tile_size = 16

[tileset]
g = "terrain:grass"

[layer.ground]
pixels = """
gggg
gggg
"""
```

### Full Example

```toml
tile_size = 16

[tileset]
g = "terrain:grass"
d = "terrain:dirt"
w = "terrain:water"
t = "terrain:tree"
_ = ""

[layer.background]
scroll_x = 0.5
scroll_y = 0.5
pixels = """
gggggggggggg
gggggggggggg
gggggggggggg
gggggggggggg
"""

[layer.ground]
pixels = """
ggggddddgggg
ggtgddddgtgg
ggggddddgggg
wwwwddddwwww
"""

[layer.objects]

[[layer.objects.entity]]
type = "spawn"
x = 64
y = 32

[[layer.objects.entity]]
type = "chest"
x = 128
y = 96
properties = { locked = true, contents = "key" }
```

### Common Mistakes

- Missing `tile_size` — required, must be positive
- Tileset reference format — must be `"file:sprite"`, not just a filename
- Unknown tileset key in grid — char must be defined in `[tileset]`
- Ragged rows — all rows in a tile layer must have the same width

---

## Instrument (.inst)

Synthesizer instrument definitions for tracker music.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | — | Instrument name (referenced by tracks) |
| `[oscillator]` | table | yes | — | Sound source |
| `[envelope]` | table | yes | — | ADSR amplitude envelope |
| `[filter]` | table | no | — | Biquad frequency filter |
| `[effects]` | table | no | — | Modulation effects |

**Oscillator:**

| Field | Type | Default | Values |
|-------|------|---------|--------|
| `waveform` | string | "sine" | `sine`, `square`, `triangle`, `sawtooth`, `noise`, `pulse` |
| `duty_cycle` | float | 0.5 | 0.0–1.0 (for `pulse` waveform) |

**Envelope (ADSR):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `attack` | float | 0.0 | Time to peak (seconds) |
| `decay` | float | 0.0 | Time to sustain level (seconds) |
| `sustain` | float | 0.0 | Sustain level (0.0–1.0) |
| `release` | float | 0.0 | Fade-out time (seconds) |

**Filter:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | — | `lowpass`, `highpass`, `bandpass` |
| `cutoff` | float | — | Frequency in Hz |
| `resonance` | float | 0.0 | 0.0–1.0 |

**Effects:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `vibrato_depth` | float | 0.0 | Semitones |
| `vibrato_rate` | float | 0.0 | Hz |

### Minimal Example

```toml
name = "beep"

[oscillator]
waveform = "sine"

[envelope]
attack = 0.01
decay = 0.0
sustain = 1.0
release = 0.1
```

### Full Example

```toml
name = "bass"

[oscillator]
waveform = "sawtooth"

[envelope]
attack = 0.01
decay = 0.05
sustain = 0.8
release = 0.3

[filter]
type = "lowpass"
cutoff = 800
resonance = 0.7

[effects]
vibrato_depth = 1.0
vibrato_rate = 6.0
```

---

## SFX (.sfx)

Procedural sound effects with layered voices.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `duration` | float | yes | — | Effect length in seconds (must be > 0) |
| `volume` | float | no | 1.0 | Master volume |
| `[[voice]]` | array | yes (1+) | — | Voice layers |

**Per-voice fields:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `waveform` | string | "sine" | `sine`, `square`, `triangle`, `sawtooth`, `noise`, `pulse` |
| `duty_cycle` | float | 0.5 | For `pulse` waveform |
| `[voice.envelope]` | table | — | ADSR (same fields as instrument) |
| `[voice.pitch]` | table | — | Frequency sweep |
| `[voice.filter]` | table | — | Filter with optional sweep |
| `[voice.effects]` | table | — | Vibrato modulation |

**Pitch sweep:**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `start` | float | 440 | Start frequency (Hz) |
| `end` | float | start | End frequency (Hz) |
| `curve` | string | "linear" | `linear`, `exponential`, `logarithmic` |

**Filter (with sweep):**

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `type` | string | — | `lowpass`, `highpass`, `bandpass` |
| `cutoff` | float | — | Static cutoff (Hz) |
| `cutoff_start` | float | — | Sweep start (Hz) |
| `cutoff_end` | float | — | Sweep end (Hz) |
| `resonance` | float | 0.0 | 0.0–1.0 |
| `curve` | string | "linear" | Sweep curve |

### Minimal Example

```toml
duration = 0.2

[[voice]]
waveform = "sine"

[voice.pitch]
start = 440
```

### Full Example

```toml
duration = 0.3
volume = 0.8

[[voice]]
waveform = "square"

[voice.envelope]
attack = 0.01
decay = 0.05
sustain = 0.3
release = 0.15

[voice.pitch]
start = 2000
end = 500
curve = "exponential"

[voice.filter]
type = "lowpass"
cutoff_start = 4000
cutoff_end = 1000
curve = "exponential"
resonance = 0.5

[[voice]]
waveform = "noise"

[voice.envelope]
attack = 0.0
decay = 0.1
sustain = 0.0
release = 0.0

[voice.pitch]
start = 800
```

### Common Mistakes

- Missing `duration` — required, must be positive
- No voices — at least one `[[voice]]` is required
- Specifying both `cutoff` and `cutoff_start/cutoff_end` — `cutoff` takes precedence

---

## Track (.track)

Tracker-style music with patterns and sequences.

### Schema

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `tempo` | int | yes | — | BPM (must be > 0) |
| `ticks_per_beat` | int | no | 4 | Subdivisions per beat |
| `loop` | bool | no | false | Enable looping |
| `loop_start` | int | no | 0 | Pattern index to loop back to |
| `[[channel]]` | array | yes (1+) | — | Channel definitions |
| `[pattern.NAME]` | table | yes (1+) | — | Pattern definitions |
| `[song]` | table | yes | — | Playback sequence |

**Channel:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | yes | — | Channel identifier |
| `instrument` | string | yes | — | References `.inst` file |
| `volume` | float | no | 1.0 | Channel volume |

**Pattern:**

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `ticks` | int | no | auto | Row count (auto-detected from data) |
| `data` | multiline | yes | — | Note grid (see below) |

**Song:**

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `sequence` | string[] | yes | Pattern names in playback order |

### Note Syntax

| Syntax | Type | Meaning |
|--------|------|---------|
| `C4` | Note on | Note name + octave (C, C#, D, D#, E, F, F#, G, G#, A, A#, B) |
| `---` | Sustain | Hold current note |
| `...` | Silence | Empty/rest |
| `^^^` | Note off | Release current note |

### Effects

Effects follow the note, separated by a space:

| Effect | Syntax | Description |
|--------|--------|-------------|
| Velocity | `v00`–`vFF` | Volume (hex) |
| Slide up | `>00`–`>FF` | Pitch slide up |
| Slide down | `<00`–`<FF` | Pitch slide down |
| Vibrato | `~00`–`~FF` | Vibrato depth |
| Arpeggio | `a00`–`aFF` | Chord arpeggio |

Example: `C4 vC` (note C4 at velocity 0xC), `D#5 >03 ~02` (D#5 with slide up and vibrato).

### Minimal Example

```toml
tempo = 120

[[channel]]
name = "lead"
instrument = "beep"

[pattern.main]
data = """
C4
E4
G4
C5
"""

[song]
sequence = ["main"]
```

### Full Example

```toml
tempo = 140
ticks_per_beat = 4
loop = true

[[channel]]
name = "melody"
instrument = "synth"
volume = 0.9

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.7

[pattern.verse]
data = """
C4 vC  C2
---    ---
D4 vC  ---
---    C2
E4 vC  D2
---    ---
---    ---
---    C2
"""

[pattern.chorus]
data = """
G4 vF  G2
A4 vF  ---
B4 vF  G2
^^^    ---
C5 vF  A2
---    ---
^^^    ---
---    A2
"""

[song]
sequence = ["verse", "verse", "chorus", "verse", "chorus"]
```

### Common Mistakes

- Column count mismatch — each row must have one column per channel
- Unknown pattern in sequence — all names in `sequence` must match a `[pattern.NAME]`
- Invalid note format — must be note name (A-G, optional #) + octave digit
- Tempo zero — must be positive
