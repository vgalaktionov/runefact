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
runefact.toml           # project configuration
assets/
  palettes/             # color palette definitions
  sprites/              # sprite sheet definitions
  maps/                 # tilemap definitions
  instruments/          # synthesizer instrument definitions
  sfx/                  # sound effect definitions
  tracks/               # tracker-style music definitions
```

### 2. Create a palette

Create `assets/palettes/default.palette`:

```toml
name = "default"

[colors]
_ = "transparent"
k = "#1a1c2c"
p = "#b13e53"
o = "#ef7d57"
y = "#ffcd75"
g = "#38b764"
b = "#29366f"
w = "#f4f4f4"
```

### 3. Create a sprite

Create `assets/sprites/player.sprite`:

```toml
palette = "default"
grid = 8

[sprite.idle]
pixels = """
__pp__
_pppp_
_kppk_
pppppp
_pppp_
__pp__
_p__p_
_k__k_
"""
```

### 4. Build

```bash
runefact build
```

Output goes to `build/assets/` by default:

- `sprites/player.png` — sprite sheet PNG
- `manifest.go` — type-safe Go asset loader

### 5. Preview

```bash
runefact preview assets/sprites/player.sprite
```

Opens a live-reloading window. Edit the `.sprite` file and watch changes appear instantly.

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
width = 640               # preview window width
height = 480              # preview window height
scale = 2                 # pixel scaling factor
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
