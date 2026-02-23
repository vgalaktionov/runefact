# Getting Started with Runefact

Runefact compiles text-based asset definitions ("runes") into game-ready artifacts for [ebitengine](https://ebitengine.org): PNG sprite sheets, JSON tilemaps, and WAV audio files. Every format is human-readable TOML designed for hand-authoring or LLM generation.

## Installation

### From source (requires Go 1.21+)

```bash
go install github.com/vgalaktionov/runefact/cmd/runefact@latest
```

### Verify

```bash
runefact version
```

## Quickstart

### 1. Create a project

```bash
mkdir my-game && cd my-game
runefact init --name my-game
```

This creates:

```
runefact.toml                      # project configuration
.mcp.json                          # MCP server config (for Claude Code / Claude Desktop)
.claude/settings.local.json        # Claude Code permissions for runefact MCP
assets/
  palettes/default.palette         # PICO-8-style 16-color palette
  sprites/player.sprite            # player character with idle animation + coin + heart
  sprites/tiles.sprite             # terrain tiles (grass, dirt, stone, sky)
  maps/level1.map                  # platformer level with background, tiles, and entities
  instruments/lead.inst            # square-wave lead synth
  instruments/bass.inst            # triangle-wave bass synth
  sfx/jump.sfx                     # jump sound effect
  sfx/coin.sfx                     # coin pickup sound effect
  tracks/demo.track                # two-channel music track (intro + verse)
```

The scaffold includes complete example assets so you can immediately build and preview.

### 2. Build

```bash
runefact build
```

Output goes to `build/assets/` by default:

- `sprites/player.png` — sprite sheet PNG (all frames packed)
- `sprites/tiles.png` — terrain tile sprites
- `maps/level1.json` — tilemap JSON
- `audio/jump.wav`, `audio/coin.wav` — sound effects
- `audio/demo.wav` — music track
- `manifest.go` — type-safe Go asset loader

### 3. Preview

```bash
runefact preview assets/sprites/player.sprite
runefact preview assets/maps/level1.map
runefact preview assets/sfx/jump.sfx
runefact preview assets/tracks/demo.track
```

Opens a live-reloading window. Edit the rune file and watch changes appear instantly.

- **Sprites**: auto-zoom grid, click to isolate, arrow keys to navigate frames
- **Maps**: renders actual tile sprites, shows entities, mouse drag to pan, scroll to zoom
- **SFX**: waveform + envelope + pitch graphs, press Enter to play
- **Music**: tracker-style note display with waveform, press Enter to play

### 6. Validate

```bash
runefact validate
```

Checks all rune files for errors without producing output.

### 7. Watch mode

```bash
runefact watch
```

Monitors your assets directory and rebuilds automatically when files change.

## Project Configuration

`runefact.toml` controls project settings:

```toml
[project]
name = "my-game"
package = "assets"        # Go package name for manifest

[defaults]
sprite_size = 16          # default sprite grid size
sample_rate = 44100       # audio sample rate
bit_depth = 16            # audio bit depth

[preview]
window_width = 1200       # preview window width
window_height = 900       # preview window height
background = "#1a1a2e"    # preview background color
pixel_scale = 4           # pixel scaling factor
audio_volume = 0.5        # preview audio volume (0.0-1.0)
```

## CLI Reference

| Command | Description |
|---------|-------------|
| `runefact build [files...]` | Compile rune files into artifacts |
| `runefact validate [files...]` | Check for errors without building |
| `runefact preview [file]` | Live-reloading asset previewer |
| `runefact watch` | Auto-rebuild on file changes |
| `runefact init` | Initialize a new project |
| `runefact mcp` | Start MCP server for AI integration |
| `runefact version` | Print version |

### Build flags

```bash
runefact build --sprites    # build only sprites
runefact build --maps       # build only maps
runefact build --audio      # build only audio
```

### Global flags

```
--config <path>   Path to runefact.toml
-v, --verbose     Verbose output
-q, --quiet       Suppress non-error output
```

## VS Code Extension

Install the **Runefact** extension for:

- Syntax highlighting for all 6 file types
- Inline color decorators in palette files
- Pixel grid colorization in sprite/map editors
- Chromatic note coloring in tracker data
- Real-time validation diagnostics
- Build/validate/preview commands (Ctrl+Shift+B to build)
- Code snippets for scaffolding new files

## Next Steps

- [Format Reference](format-reference.md) — complete schemas for all file types
- [Sprite Authoring Guide](sprite-guide.md) — designing sprites and animations
- [Map Authoring Guide](map-guide.md) — building tilemaps and levels
- [Audio Authoring Guide](audio-guide.md) — creating SFX and music
- [Ebitengine Integration](ebitengine-integration.md) — loading artifacts in your game
- [AI Workflows](ai-workflows.md) — using Runefact with Claude and MCP
- [MCP Reference](mcp-reference.md) — MCP tool and resource documentation
