# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Runefact is a Go CLI tool that compiles text-based asset definitions ("runes") into game-ready artifacts for ebitengine: PNG sprite sheets, JSON tilemaps, and WAV audio files. Every asset format is human-readable TOML designed for LLM authoring.

The repo also contains a VS Code extension (`vscode/runefact-vscode/`) and a local MCP server (`runefact mcp`).

## Development Commands

```bash
make build                           # build binary to bin/runefact
make run ARGS="build"                # go run with subcommand (build, validate, preview, etc.)
make test                            # all tests
make vet                             # go vet

# Targeted testing
go test ./internal/sprite/...                       # single package
go test -run TestParsePixels ./internal/sprite/...  # single test
```

## Architecture

```
cmd/runefact/          CLI entry point (build, validate, preview, watch, init, mcp subcommands)
internal/
  palette/             .palette parser — shared color definitions, single-char keys
  sprite/              .sprite parser + PNG sprite sheet renderer
  tilemap/             .map parser + JSON output
  instrument/          .inst parser — synthesizer instrument definitions
  sfx/                 .sfx parser + WAV renderer — procedural sound effects
  track/               .track parser + WAV renderer — tracker-style music
  audio/               shared audio: synthesis engine, brickwall limiter, WAV writer
  manifest/            manifest.go code generator (type-safe ebitengine asset loading)
  preview/             ebitengine live-reloading previewer
  watcher/             fsnotify file watcher for watch/preview modes
  mcp/                 MCP server (stdio transport): tools, resources, inspect handlers
vscode/runefact-vscode/  VS Code extension (TypeScript): TextMate grammars, diagnostics, color decorators
docs/                  Usage documentation (Markdown, doubles as agent context)
```

### Key data flow

1. Parser reads TOML rune file, validates structure, returns typed AST
2. Renderer takes AST + resolved palette, produces artifact (PNG/JSON/WAV)
3. Manifest generator collects all sprite/map/audio metadata, emits `manifest.go`
4. Previewer uses the same parsers/renderers with fsnotify hot-reload

All parsers live in their own `internal/` package and share no global state. Palette resolution is the one cross-cutting concern: sprites and maps reference palettes by name, so palette files must be parsed first.

### Asset format conventions

- All formats are TOML-based with `pixels = """..."""` or `data = """..."""` text blocks for grid data
- Single-char palette keys for readability; multi-char keys use `[xx]` bracket syntax in grids
- `_` is always transparent/empty
- Validation is lenient: warn over error when ambiguous, never block the agent workflow on false positives

### Audio safety (non-negotiable)

- All audio output goes through a brickwall limiter (-1 dBFS, 0ms attack, 50ms release)
- Audio is never auto-played in the previewer — always requires explicit user action
- DC offset removal via 10Hz high-pass on all output
- NaN/Inf samples replaced with silence + warning

## Commit Conventions

Use [Conventional Commits](https://www.conventionalcommits.org/). Every commit must have:

1. A short headline: `type(scope): summary` (imperative mood, lowercase, no period, max 72 chars)
2. A blank line followed by a detailed body explaining *what* changed and *why*
3. A task reference when the work relates to a Task Master task

**Types:** `feat`, `fix`, `refactor`, `test`, `docs`, `chore`, `build`, `ci`, `perf`

**Scopes:** use the relevant `internal/` package name (`palette`, `sprite`, `tilemap`, `sfx`, `track`, `audio`, `manifest`, `preview`, `mcp`, `watcher`) or `cli`, `vscode`, `docs`

**Task references:** include `Task: <id>` (e.g. `Task: 3.2`) on its own line in the body when the commit implements or advances a Task Master task or subtask.

Example:
```
feat(sprite): parse multi-char palette keys in pixel grids

Add bracket-syntax parsing for multi-char keys like [sk] inside
pixels blocks. The parser now resolves these against both the
referenced .palette file and inline palette_extend definitions.

Task: 4.3
```

## Project Config

User projects have a `runefact.toml` at root with `[project]`, `[defaults]`, and `[preview]` sections. Output goes to `build/assets/` by default.

## Task Master AI Instructions
**Import Task Master's development workflow commands and guidelines, treat as if import is in the main CLAUDE.md file.**
@./.taskmaster/CLAUDE.md
