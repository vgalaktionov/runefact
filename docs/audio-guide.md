# Audio Authoring Guide

Runefact generates audio procedurally — no sample files needed. Sound effects (`.sfx`) use layered oscillators with envelopes and sweeps. Music (`.track`) uses a tracker pattern format with instrument definitions (`.inst`).

## Audio Safety

All audio output is processed through:
- **Brickwall limiter** at -1 dBFS (0ms attack, 50ms release)
- **DC offset removal** via 10Hz high-pass filter
- **NaN/Inf protection** — replaced with silence + warning

Audio is never auto-played in the previewer — always requires explicit user action.

## Waveform Selection

| Waveform | Character | Use Cases |
|----------|-----------|-----------|
| `sine` | Pure, smooth | Bass tones, gentle beeps, sub-bass |
| `square` | Hollow, buzzy | Retro leads, chip-tune, NES-style |
| `triangle` | Soft, warm | Bass, flutes, softer leads |
| `sawtooth` | Bright, rich | Brass, strings, aggressive leads |
| `noise` | Noise burst | Drums, explosions, wind, static |
| `pulse` | Variable | Thin-to-thick via `duty_cycle` |

## ADSR Envelopes

The amplitude envelope shapes how a sound evolves over time:

```
    ^
    |   /\
Amp |  /  \___________
    | /    |          |\
    |/     |          | \
    +------+----------+--+-->
    Attack Decay Sustain Release
```

| Parameter | Effect | Typical Ranges |
|-----------|--------|----------------|
| `attack` | Rise time to peak | 0.0–0.5s |
| `decay` | Fall time to sustain | 0.0–0.5s |
| `sustain` | Held level (0–1) | 0.0–1.0 |
| `release` | Fade after note-off | 0.05–2.0s |

**Quick presets:**

| Sound Type | Attack | Decay | Sustain | Release |
|------------|--------|-------|---------|---------|
| Pluck | 0.001 | 0.1 | 0.0 | 0.1 |
| Pad | 0.3 | 0.2 | 0.7 | 0.5 |
| Percussion | 0.0 | 0.05 | 0.0 | 0.05 |
| Organ | 0.01 | 0.0 | 1.0 | 0.05 |
| Swell | 0.5 | 0.0 | 1.0 | 1.0 |

## Sound Effects (.sfx)

### SFX Recipe Book

**UI Click:**
```toml
duration = 0.05

[[voice]]
waveform = "sine"

[voice.pitch]
start = 1200
end = 800
curve = "exponential"

[voice.envelope]
attack = 0.001
decay = 0.04
sustain = 0.0
release = 0.01
```

**Jump:**
```toml
duration = 0.15

[[voice]]
waveform = "square"

[voice.pitch]
start = 200
end = 600
curve = "exponential"

[voice.envelope]
attack = 0.01
decay = 0.05
sustain = 0.3
release = 0.05
```

**Coin Collect:**
```toml
duration = 0.2

[[voice]]
waveform = "square"

[voice.pitch]
start = 800

[voice.envelope]
attack = 0.001
decay = 0.05
sustain = 0.3
release = 0.1

[[voice]]
waveform = "square"

[voice.pitch]
start = 1200

[voice.envelope]
attack = 0.05
decay = 0.05
sustain = 0.2
release = 0.1
```

**Explosion:**
```toml
duration = 0.6
volume = 0.8

[[voice]]
waveform = "noise"

[voice.envelope]
attack = 0.0
decay = 0.1
sustain = 0.3
release = 0.3

[voice.filter]
type = "lowpass"
cutoff_start = 3000
cutoff_end = 200
curve = "exponential"
resonance = 0.2

[[voice]]
waveform = "sine"

[voice.pitch]
start = 80
end = 30
curve = "linear"

[voice.envelope]
attack = 0.0
decay = 0.2
sustain = 0.1
release = 0.3
```

**Laser:**
```toml
duration = 0.25

[[voice]]
waveform = "sawtooth"

[voice.pitch]
start = 1500
end = 300
curve = "exponential"

[voice.envelope]
attack = 0.01
decay = 0.05
sustain = 0.4
release = 0.1

[voice.filter]
type = "lowpass"
cutoff_start = 5000
cutoff_end = 500
resonance = 0.6
```

**Damage:**
```toml
duration = 0.3

[[voice]]
waveform = "square"

[voice.pitch]
start = 300
end = 100
curve = "linear"

[voice.envelope]
attack = 0.0
decay = 0.1
sustain = 0.2
release = 0.1

[[voice]]
waveform = "noise"

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.0
release = 0.0
```

### Layering Voices

Combine multiple voices for richer effects:
- **Body + transient**: sine/square for tone + noise for attack texture
- **Octave layers**: same waveform at different pitches for fullness
- **Tonal + noise**: pitched voice for character + noise for texture

## Instruments (.inst)

Instruments define the sound for each tracker channel.

**Lead synth:**
```toml
name = "lead"

[oscillator]
waveform = "square"

[envelope]
attack = 0.02
decay = 0.1
sustain = 0.6
release = 0.15

[effects]
vibrato_depth = 1.5
vibrato_rate = 5.0
```

**Bass:**
```toml
name = "bass"

[oscillator]
waveform = "sawtooth"

[envelope]
attack = 0.01
decay = 0.05
sustain = 0.8
release = 0.1

[filter]
type = "lowpass"
cutoff = 600
resonance = 0.5
```

**Pad:**
```toml
name = "pad"

[oscillator]
waveform = "triangle"

[envelope]
attack = 0.3
decay = 0.2
sustain = 0.7
release = 0.5
```

## Tracker Music (.track)

### Pattern Basics

Patterns are grids where each row is a tick and each column is a channel:

```
C4  C2      ← tick 0: melody plays C4, bass plays C2
---  ---     ← tick 1: both sustain
E4  ---     ← tick 2: melody changes to E4, bass sustains
---  C2      ← tick 3: melody sustains, bass plays C2
```

**Note syntax:**
- `C4`, `D#5`, `A#3` — note on (name + octave)
- `---` — sustain (hold previous note)
- `...` — silence (no sound)
- `^^^` — note off (release envelope)

### Effects

Add effects after the note:

```
C4 vC       ← velocity 0xC (loud)
D4 v4       ← velocity 0x4 (soft)
E4 >02      ← slide up
F4 <03      ← slide down
G4 ~04      ← vibrato
```

### Composing a Song

```toml
tempo = 120
ticks_per_beat = 4
loop = true

[[channel]]
name = "melody"
instrument = "lead"
volume = 0.8

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.7

[[channel]]
name = "drums"
instrument = "perc"
volume = 0.6

[pattern.intro]
data = """
...    C2     ...
...    ---    ...
...    ---    ...
...    C2     ...
C4     ---    ...
---    ---    ...
E4     E2     ...
---    ---    ...
"""

[pattern.verse]
data = """
C4 vC  C2     C5
---    ---    ...
D4     ---    ...
---    C2     C5
E4     D2     ...
---    ---    C5
---    ---    ...
---    C2     C5
"""

[pattern.chorus]
data = """
G4 vF  G2     C5
A4     ---    C5
B4     G2     ...
^^^    ---    C5
C5     A2     C5
---    ---    ...
^^^    ---    C5
---    A2     ...
"""

[song]
sequence = ["intro", "verse", "verse", "chorus", "verse", "chorus", "chorus"]
```

### Mixing Tips

- Keep melody channel at 0.7–0.9 volume
- Bass at 0.5–0.7
- Percussion at 0.4–0.6
- Use velocity effects (`v`) for dynamic variation within patterns
- Start patterns with strong notes, end with sustains or releases for smooth transitions
