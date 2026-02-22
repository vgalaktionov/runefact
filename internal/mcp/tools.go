package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/vgalaktionov/runefact/internal/build"
	"github.com/vgalaktionov/runefact/internal/palette"
	"github.com/vgalaktionov/runefact/internal/sfx"
	"github.com/vgalaktionov/runefact/internal/sprite"
	"github.com/vgalaktionov/runefact/internal/tilemap"
	"github.com/vgalaktionov/runefact/internal/track"
)

func (ctx *ServerContext) handleBuild(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	ctx.BuildMu.Lock()
	defer ctx.BuildMu.Unlock()

	scope := req.GetString("scope", "all")
	files := req.GetStringSlice("files", nil)

	opts := build.Options{
		Scope: build.Scope(scope),
		Files: files,
	}

	result := build.Build(opts, ctx.Config, ctx.ProjectRoot)

	resp := map[string]any{
		"success":   len(result.Errors) == 0,
		"artifacts": result.Artifacts,
		"warnings":  result.Warnings,
	}

	if len(result.Errors) > 0 {
		errs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			errs[i] = e.Error()
		}
		resp["errors"] = errs
	}
	if result.ManifestPath != "" {
		resp["manifest_path"] = result.ManifestPath
	}

	return jsonResult(resp)
}

func (ctx *ServerContext) handleValidate(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	files := req.GetStringSlice("files", nil)

	opts := build.Options{
		Scope: build.ScopeAll,
		Files: files,
	}

	result := build.Validate(opts, ctx.Config, ctx.ProjectRoot)

	resp := map[string]any{
		"valid":    len(result.Errors) == 0,
		"warnings": result.Warnings,
	}

	if len(result.Errors) > 0 {
		errs := make([]string, len(result.Errors))
		for i, e := range result.Errors {
			errs[i] = e.Error()
		}
		resp["errors"] = errs
	}

	return jsonResult(resp)
}

func (ctx *ServerContext) handleInspectSprite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	path := filepath.Join(ctx.ProjectRoot, "assets", "sprites", file)
	sf, err := sprite.LoadSpriteFile(path)
	if err != nil {
		return errorResult(fmt.Sprintf("loading %s: %v", file, err))
	}

	sprites := make([]map[string]any, len(sf.Sprites))
	for i, s := range sf.Sprites {
		sprites[i] = map[string]any{
			"name":      s.Name,
			"width":     s.Grid.W,
			"height":    s.Grid.H,
			"frames":    len(s.Frames),
			"framerate": s.Framerate,
		}
	}

	return jsonResult(map[string]any{
		"file":           file,
		"palette":        sf.PaletteRef,
		"palette_extend": sf.PaletteExtend,
		"default_grid":   fmt.Sprintf("%dx%d", sf.DefaultGrid.W, sf.DefaultGrid.H),
		"sprites":        sprites,
	})
}

func (ctx *ServerContext) handleInspectMap(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	path := filepath.Join(ctx.ProjectRoot, "assets", "maps", file)
	mf, warnings, err := tilemap.LoadMapFile(path)
	if err != nil {
		return errorResult(fmt.Sprintf("loading %s: %v", file, err))
	}

	layers := make([]map[string]any, len(mf.Layers))
	for i, l := range mf.Layers {
		layer := map[string]any{
			"name": l.Name,
			"type": l.Type,
		}
		if l.Type == "tile" && len(l.Data) > 0 {
			layer["rows"] = len(l.Data)
			layer["cols"] = len(l.Data[0])
		}
		if l.Type == "entity" {
			layer["entity_count"] = len(l.Entities)
		}
		layers[i] = layer
	}

	warnStrs := make([]string, len(warnings))
	for i, w := range warnings {
		warnStrs[i] = w.Message
	}

	return jsonResult(map[string]any{
		"file":      file,
		"tile_size": mf.TileSize,
		"layers":    layers,
		"warnings":  warnStrs,
	})
}

func (ctx *ServerContext) handleInspectAudio(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	ext := filepath.Ext(file)
	switch ext {
	case ".sfx":
		path := filepath.Join(ctx.ProjectRoot, "assets", "sfx", file)
		s, err := sfx.LoadSFX(path)
		if err != nil {
			return errorResult(fmt.Sprintf("loading %s: %v", file, err))
		}
		return jsonResult(map[string]any{
			"file":     file,
			"type":     "sfx",
			"duration": s.Duration,
			"volume":   s.Volume,
			"voices":   len(s.Voices),
		})

	case ".track":
		path := filepath.Join(ctx.ProjectRoot, "assets", "tracks", file)
		tr, err := track.LoadTrack(path)
		if err != nil {
			return errorResult(fmt.Sprintf("loading %s: %v", file, err))
		}
		return jsonResult(map[string]any{
			"file":     file,
			"type":     "track",
			"tempo":    tr.Tempo,
			"channels": len(tr.Channels),
			"patterns": len(tr.Patterns),
			"sequence": tr.Sequence,
			"loop":     tr.Loop,
		})

	default:
		return errorResult(fmt.Sprintf("unsupported audio type: %s", ext))
	}
}

func (ctx *ServerContext) handleListAssets(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	filterType := req.GetString("type", "")
	assetsDir := filepath.Join(ctx.ProjectRoot, "assets")

	type assetEntry struct {
		File string `json:"file"`
		Type string `json:"type"`
		Dir  string `json:"dir"`
	}

	dirs := map[string]struct {
		ext      string
		typeName string
	}{
		"palettes":    {".palette", "palette"},
		"sprites":     {".sprite", "sprite"},
		"maps":        {".map", "map"},
		"instruments": {".inst", "instrument"},
		"sfx":         {".sfx", "sfx"},
		"tracks":      {".track", "track"},
	}

	var assets []assetEntry
	for dir, info := range dirs {
		if filterType != "" && filterType != info.typeName {
			continue
		}
		dirPath := filepath.Join(assetsDir, dir)
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), info.ext) {
				assets = append(assets, assetEntry{
					File: e.Name(),
					Type: info.typeName,
					Dir:  dir,
				})
			}
		}
	}

	return jsonResult(map[string]any{
		"assets": assets,
		"count":  len(assets),
	})
}

func (ctx *ServerContext) handlePaletteColors(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	path := filepath.Join(ctx.ProjectRoot, "assets", "palettes", file)
	pal, err := palette.LoadPalette(path)
	if err != nil {
		return errorResult(fmt.Sprintf("loading %s: %v", file, err))
	}

	colors := make(map[string]string, len(pal.Colors))
	for key, c := range pal.Colors {
		if c.A == 255 {
			colors[key] = fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
		} else {
			colors[key] = fmt.Sprintf("#%02x%02x%02x%02x", c.R, c.G, c.B, c.A)
		}
	}

	return jsonResult(map[string]any{
		"file":   file,
		"name":   pal.Name,
		"colors": colors,
	})
}

func (ctx *ServerContext) handleFormatHelp(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	format, err := req.RequireString("format")
	if err != nil {
		return errorResult("format parameter required")
	}

	doc, ok := formatDocs[format]
	if !ok {
		return errorResult(fmt.Sprintf("unknown format: %s", format))
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: doc,
			},
		},
	}, nil
}

// formatDocs maps format name to documentation.
var formatDocs = map[string]string{
	"palette": `# .palette Format

TOML file defining a named color palette with single-char or multi-char keys.

` + "```toml" + `
name = "default"

[colors]
r = "#ff0000"
g = "#00ff00"
b = "#0000ff"
_ = "#00000000"      # transparent
sk = "#ffcc99"       # multi-char key
` + "```" + `

Colors use hex notation: #RGB, #RRGGBB, or #RRGGBBAA.
Key "_" is always transparent. Keys can be 1+ characters.
`,

	"sprite": `# .sprite Format

TOML file defining one or more sprites with pixel grids.

` + "```toml" + `
palette = "default"
grid = "16x16"

[palette_extend]
x = "#ff00ff"

[sprite.player]
pixels = """
____rrrr____
__rrrrrrrr__
__rr__rr__rr
"""

[sprite.coin]
framerate = 8
frame_count = 4
pixels = """
__yy__
_yyyy_
_yyyy_
__yy__
--
_yyyy_
yyyyyy
yyyyyy
_yyyy_
"""
` + "```" + `

palette: references a .palette file by name (without extension).
grid: "WxH" default dimensions. Sprites auto-detect if omitted.
Frames separated by "--" on its own line. [xx] bracket syntax for multi-char palette keys.
`,

	"map": `# .map Format

TOML file defining a tilemap with tile and entity layers.

` + "```toml" + `
tile_size = 16

[tileset]
g = "terrain:grass"
w = "terrain:wall"
_ = ""

[layer.background]
pixels = """
gggggggg
gggggggg
"""

[layer.entities]
[[layer.entities.entity]]
type = "spawn"
x = 2
y = 3
` + "```" + `

tileset maps single chars to "sprite_file:sprite_name" references.
Layers can be tile (with pixels) or entity (with entity list).
"_" or empty string = empty/transparent tile.
`,

	"instrument": `# .inst Format

TOML file defining a synthesizer instrument for use in .track files.

` + "```toml" + `
name = "bass"
waveform = "square"
duty_cycle = 0.25

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.6
release = 0.15

[filter]
type = "lowpass"
cutoff = 800
resonance = 2.0
` + "```" + `

Waveforms: sine, square, triangle, sawtooth, noise, pulse.
Envelope: ADSR in seconds. Filter: lowpass, highpass, bandpass.
`,

	"sfx": `# .sfx Format

TOML file defining a procedural sound effect with one or more voices.

` + "```toml" + `
duration = 0.3
volume = 0.8

[[voice]]
waveform = "square"

[voice.envelope]
attack = 0.01
decay = 0.05
sustain = 0.3
release = 0.1

[voice.pitch]
start = 880
end = 220
curve = "exponential"
` + "```" + `

duration: total length in seconds. volume: master volume 0.0-1.0.
Each voice has waveform, envelope (ADSR), pitch (start/end/curve), optional filter and effects.
Pitch curves: linear, exponential, logarithmic.
`,

	"track": `# .track Format

TOML file defining tracker-style music with patterns and channels.

` + "```toml" + `
tempo = 120
ticks_per_beat = 4

[[channel]]
name = "lead"
instrument = "synth"
volume = 0.8

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.6

[pattern.intro]
ticks = 16
data = """
C4  C3
D4  C3
E4  E3
G4  G3
"""

[song]
sequence = ["intro", "verse", "chorus"]
loop = true
` + "```" + `

Notes: C4, C#5, D3, etc. Special: --- (sustain), ... (silence), ^^^ (note off).
Effects after note: v80 (velocity), >4 (slide up), <4 (slide down), ~3 (vibrato).
`,
}

func jsonResult(data any) (*mcp.CallToolResult, error) {
	b, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return nil, err
	}
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(b),
			},
		},
	}, nil
}

func errorResult(msg string) (*mcp.CallToolResult, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf(`{"error": %q}`, msg),
			},
		},
		IsError: true,
	}, nil
}
