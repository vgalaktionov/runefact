# Runefact

[![Go Reference](https://pkg.go.dev/badge/github.com/vgalaktionov/runefact.svg)](https://pkg.go.dev/github.com/vgalaktionov/runefact)
[![CI](https://github.com/vgalaktionov/runefact/actions/workflows/ci.yml/badge.svg)](https://github.com/vgalaktionov/runefact/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/vgalaktionov/runefact)](https://goreportcard.com/report/github.com/vgalaktionov/runefact)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A text-based, AI-native asset pipeline for [Ebitengine](https://ebitengine.org). Define sprites, tilemaps, sound effects, and music as human-readable TOML files — compile them to PNGs, JSON, and WAVs with a single command.

Designed from the ground up for authoring by LLMs and coding agents via the built-in [MCP](https://modelcontextprotocol.io) server.

## Features

- **Text-first assets** — Every format is TOML with embedded pixel/note grids. Version control, diff, and review game assets like code.
- **Procedural audio** — Synthesize sound effects and tracker-style music entirely from text definitions. No sample files needed.
- **AI-native tooling** — Built-in MCP server exposes build, validate, and inspect operations to AI agents. Ship a VS Code extension with syntax highlighting, inline diagnostics, and color decorators.
- **Ebitengine integration** — Generates a type-safe `manifest.go` for zero-config asset loading.
- **Live preview** — Hot-reloading previewer for sprites, maps, SFX, and music.
- **Watch mode** — Auto-rebuild on file changes with dependency tracking.

## Quick Start

```bash
# Install
go install github.com/vgalaktionov/runefact/cmd/runefact@latest

# Create a project
mkdir my-game && cd my-game
runefact init --name my-game

# Build assets
runefact build

# Preview a sprite
runefact preview assets/sprites/demo.sprite
```

The `init` command scaffolds a project with example assets for every format — palette, sprite, map, instrument, SFX, and track.

## What It Looks Like

A sprite definition:

```toml
palette = "default"
grid = 8

[sprite.heart]
pixels = """
_rr_rr_
rrrrrrr
rrrrrrr
_rrrrr_
__rrr__
___r___
"""
```

A sound effect:

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

A tracker pattern:

```toml
tempo = 120

[[channel]]
instrument = "lead"

[pattern.intro]
data = """
C4  E4
D4  ---
E4  G4
G4  ---
"""

[song]
sequence = ["intro"]
```

## CLI

| Command | Description |
|---------|-------------|
| `runefact build` | Compile all assets (or `--sprites`, `--maps`, `--audio`) |
| `runefact validate` | Check for errors without building |
| `runefact preview <file>` | Live-reloading asset previewer |
| `runefact watch` | Auto-rebuild on file changes |
| `runefact init` | Scaffold a new project |
| `runefact mcp` | Start MCP server for AI agent integration |

## Asset Formats

| Extension | Produces | Description |
|-----------|----------|-------------|
| `.palette` | — | Shared color definitions |
| `.sprite` | PNG | Sprite sheets with optional animation frames |
| `.map` | JSON | Tilemaps with layers, parallax, and entities |
| `.inst` | — | Synthesizer instrument definitions |
| `.sfx` | WAV | Procedural sound effects |
| `.track` | WAV | Tracker-style music |

## MCP Server

Runefact ships an [MCP](https://modelcontextprotocol.io) server for AI agent integration. Add to your `.mcp.json`:

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

Exposes 8 tools (`build`, `validate`, `inspect_sprite`, `inspect_map`, `inspect_audio`, `list_assets`, `palette_colors`, `format_help`) and 2 resources (`project/status`, `manifest`).

## VS Code Extension

The `vscode/runefact-vscode/` directory contains a VS Code extension with:

- Syntax highlighting for all 6 file types
- Inline color decorators and pixel grid colorization
- Real-time validation diagnostics
- Chromatic note coloring in tracker data
- Build/validate/preview commands
- Code snippets for scaffolding new files

## Project Structure

```
runefact.toml                  # Project configuration
assets/
  palettes/*.palette           # Color palettes
  sprites/*.sprite             # Sprite definitions
  maps/*.map                   # Tilemap definitions
  instruments/*.inst           # Instrument definitions
  sfx/*.sfx                    # Sound effect definitions
  tracks/*.track               # Music definitions
build/assets/                  # Compiled output
  manifest.go                  # Type-safe Go asset loader
  sprites/*.png                # Sprite sheets
  maps/*.json                  # Tilemap data
  audio/*.wav                  # Audio files
```

## Documentation

- [Getting Started](docs/getting-started.md) — Installation, quickstart, project layout
- [Format Reference](docs/format-reference.md) — Complete schemas for all file types
- [Sprite Guide](docs/sprite-guide.md) — Palette design, pixel grids, animation
- [Map Guide](docs/map-guide.md) — Tilesets, layers, parallax, entities
- [Audio Guide](docs/audio-guide.md) — Waveforms, envelopes, SFX recipes, tracker patterns
- [Ebitengine Integration](docs/ebitengine-integration.md) — Loading artifacts in your game
- [AI Workflows](docs/ai-workflows.md) — Using Runefact with Claude and MCP
- [MCP Reference](docs/mcp-reference.md) — Tool and resource documentation

## Requirements

- Go 1.21+
- Linux/macOS/Windows
- X11 dev headers for the previewer on Linux (`libx11-dev libgl-dev libxrandr-dev libxxf86vm-dev libxi-dev libxcursor-dev libxinerama-dev`)

## Contributing

This is a solo project. Code contributions (PRs) are **not accepted**, but bug reports and feature requests are welcome — please open an issue.

### Building from source

```bash
make build     # Build binary
make test      # Run all tests
make vet       # Go vet
```

## License

[MIT](LICENSE)
