# MCP Reference

Complete reference for the Runefact MCP server tools and resources.

## Starting the Server

```bash
runefact mcp
```

The server uses stdio transport (JSON-RPC over stdin/stdout). It reads `runefact.toml` from the current directory to resolve project paths.

---

## Tools

### runefact_build

Compile rune asset files into game-ready artifacts.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `scope` | string | no | `"all"`, `"sprites"`, `"maps"`, or `"audio"` (default: `"all"`) |
| `files` | string[] | no | Specific files to build (empty = all matching scope) |

**Example:**
```json
{
  "name": "runefact_build",
  "arguments": {
    "scope": "sprites",
    "files": ["player.sprite"]
  }
}
```

**Returns:** JSON with build results (files built, errors if any).

---

### runefact_validate

Check rune files for errors without producing output.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `files` | string[] | no | Specific files to validate (empty = all) |

**Example:**
```json
{
  "name": "runefact_validate",
  "arguments": {
    "files": ["player.sprite", "world.map"]
  }
}
```

**Returns:** JSON with validation results (errors and warnings per file).

---

### runefact_inspect_sprite

Get sprite sheet metadata.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Sprite file name (e.g., `"player.sprite"`) |

**Example:**
```json
{
  "name": "runefact_inspect_sprite",
  "arguments": { "file": "player.sprite" }
}
```

**Returns:** JSON with sprite names, dimensions, frame counts, and framerate for each sprite in the file.

---

### runefact_inspect_map

Get map metadata.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Map file name (e.g., `"world.map"`) |

**Example:**
```json
{
  "name": "runefact_inspect_map",
  "arguments": { "file": "world.map" }
}
```

**Returns:** JSON with tile size, layer information (names, types, dimensions), tileset keys, and entity counts.

---

### runefact_inspect_audio

Get audio file metadata.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Audio file name (e.g., `"laser.sfx"` or `"bgm.track"`) |

**Example:**
```json
{
  "name": "runefact_inspect_audio",
  "arguments": { "file": "laser.sfx" }
}
```

**Returns:** JSON with duration, voice count (SFX) or channel/pattern info (track).

---

### runefact_list_assets

List all rune asset files in the project.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `type` | string | no | Filter: `"palette"`, `"sprite"`, `"map"`, `"instrument"`, `"sfx"`, or `"track"` |

**Example:**
```json
{
  "name": "runefact_list_assets",
  "arguments": { "type": "sprite" }
}
```

**Returns:** JSON array of file paths matching the filter.

---

### runefact_palette_colors

Get resolved palette colors.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Palette file name (e.g., `"default.palette"`) |

**Example:**
```json
{
  "name": "runefact_palette_colors",
  "arguments": { "file": "default.palette" }
}
```

**Returns:** JSON map of key → hex color for all colors in the palette.

---

### runefact_format_help

Get documentation for a rune asset format.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `format` | string | yes | `"palette"`, `"sprite"`, `"map"`, `"instrument"`, `"sfx"`, or `"track"` |

**Example:**
```json
{
  "name": "runefact_format_help",
  "arguments": { "format": "sprite" }
}
```

**Returns:** Format documentation including schema, fields, examples, and conventions.

---

### runefact_preview_map

Render a map file as a PNG image and return it inline. Use this to visually inspect map layouts, tile art, and entity placement.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Map file name (e.g., `"level1.map"`) |
| `scale` | integer | no | Pixel scale factor (default: 2, max: 8) |

**Example:**
```json
{
  "name": "runefact_preview_map",
  "arguments": { "file": "level1.map", "scale": 3 }
}
```

**Returns:** Inline PNG image with tile sprites rendered at the given scale. Entities with `sprite` properties are rendered using their referenced sprite; others show a colored diamond marker.

---

### runefact_preview_sprite

Render a sprite file as a PNG image and return it inline. Shows all sprites with all animation frames laid out in a grid.

**Parameters:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `file` | string | yes | Sprite file name (e.g., `"player.sprite"`) |
| `scale` | integer | no | Pixel scale factor (default: 4, max: 16) |

**Example:**
```json
{
  "name": "runefact_preview_sprite",
  "arguments": { "file": "player.sprite", "scale": 6 }
}
```

**Returns:** Inline PNG image with each sprite on its own row and frames laid out horizontally. Transparent areas show a checkerboard pattern.

---

## Resources

### runefact://project/status

Current project configuration and build status.

**MIME type:** `application/json`

**Returns:**
```json
{
  "name": "my-game",
  "root": "/path/to/project",
  "output": "build/assets",
  "package": "assets",
  "defaults": {
    "sprite_size": 16,
    "sample_rate": 44100,
    "bit_depth": 16
  },
  "preview": {
    "width": 640,
    "height": 480,
    "scale": 2
  }
}
```

---

### runefact://manifest

Current build manifest (generated Go source).

**MIME type:** `text/x-go`

**Returns:** The content of the generated `manifest.go` file, or an error message if no build has been performed yet.

---

## Error Handling

All tools return errors as JSON with an `error` field:

```json
{
  "error": "file not found: player.sprite"
}
```

Validation errors include line/column information:

```json
{
  "errors": [
    {
      "file": "player.sprite",
      "line": 12,
      "column": 0,
      "message": "Ragged row: expected width 16, got 14",
      "severity": "error"
    }
  ]
}
```
