# AI Workflows

Runefact is designed for LLM authoring. The MCP server exposes all Runefact operations as tools, enabling AI agents to create, validate, and iterate on game assets conversationally.

## MCP Server Setup

### Claude Code

Add to your project's `.mcp.json`:

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

The server runs on stdio and exposes tools and resources automatically.

### Claude Desktop

Add to your Claude Desktop MCP configuration:

```json
{
  "mcpServers": {
    "runefact": {
      "command": "/path/to/runefact",
      "args": ["mcp"],
      "cwd": "/path/to/your/game/project"
    }
  }
}
```

Set `cwd` to your project root (where `runefact.toml` lives).

## Available MCP Tools

| Tool | Description |
|------|-------------|
| `runefact_build` | Compile assets (all, sprites, maps, or audio) |
| `runefact_validate` | Check files for errors without building |
| `runefact_inspect_sprite` | Get sprite metadata (names, dimensions, frames) |
| `runefact_inspect_map` | Get map metadata (layers, dimensions, entities) |
| `runefact_inspect_audio` | Get audio metadata (duration, voices, instruments) |
| `runefact_list_assets` | List all asset files, optionally by type |
| `runefact_palette_colors` | Get resolved colors for a palette |
| `runefact_format_help` | Get format documentation |
| `runefact_preview_map` | Render a map as an inline PNG image |
| `runefact_preview_sprite` | Render a sprite sheet as an inline PNG image |

## Available MCP Resources

| Resource | Description |
|----------|-------------|
| `runefact://project/status` | Project configuration and build status |
| `runefact://manifest` | Current build manifest |

## Workflow Patterns

### Creating a New Sprite

```
User: Create a 16x16 player character sprite with idle animation

Agent workflow:
1. runefact_list_assets(type: "palette")        → find available palettes
2. runefact_palette_colors(file: "default")     → see color options
3. Write assets/sprites/player.sprite
4. runefact_validate(files: ["player.sprite"])   → check for errors
5. runefact_build(scope: "sprites")              → generate PNG
6. runefact_preview_sprite(file: "player.sprite") → visually verify the result
```

### Adding Sound Effects

```
User: Make a jump sound effect

Agent workflow:
1. runefact_format_help(format: "sfx")       → review format
2. Write assets/sfx/jump.sfx
3. runefact_validate(files: ["jump.sfx"])
4. runefact_build(scope: "audio")
5. runefact_inspect_audio(file: "jump")      → check duration/voices
```

### Building a Level

```
User: Create a platformer level using the terrain tileset

Agent workflow:
1. runefact_inspect_sprite(file: "tiles.sprite")  → see available tiles
2. runefact_format_help(format: "map")             → review format
3. Write assets/maps/level1.map
4. runefact_validate(files: ["level1.map"])
5. runefact_build(scope: "maps")
6. runefact_preview_map(file: "level1.map")        → visually verify the level
```

### Iterating on Assets

```
User: The explosion sound is too long, make it punchier

Agent workflow:
1. runefact_inspect_audio(file: "explosion") → see current params
2. Read assets/sfx/explosion.sfx
3. Edit: reduce duration, shorten decay
4. runefact_validate(files: ["explosion.sfx"])
5. runefact_build(scope: "audio")
6. runefact_inspect_audio(file: "explosion") → verify changes
```

### Full Project Build

```
User: Build everything and check for errors

Agent workflow:
1. runefact_validate()                        → check all files
2. runefact_build(scope: "all")              → compile everything
3. Read runefact://project/status            → verify build status
4. Read runefact://manifest                  → check generated manifest
```

## Agent Best Practices

1. **Inspect before creating** — use `inspect` and `palette_colors` to understand existing assets before adding new ones. This ensures consistency.

2. **Validate after every change** — catch errors immediately rather than discovering them at build time.

3. **Preview to verify** — use `preview_sprite` and `preview_map` to visually verify your work. These return inline PNG images so you can see the actual rendered result without leaving the conversation.

4. **Use `format_help` when unsure** — the tool returns complete format documentation, no need to guess syntax.

5. **Iterate incrementally** — create a minimal version first, validate it, then add detail. This is especially important for pixel art and audio.

6. **Read the palette** — always check `palette_colors` before writing pixel grids. Using the wrong key produces validation warnings.

7. **Build scoped** — use `scope: "sprites"` instead of `scope: "all"` when only sprite files changed. It's faster.

8. **Check project status** — the `runefact://project/status` resource shows current config and defaults, useful for knowing sprite sizes and audio sample rates.
