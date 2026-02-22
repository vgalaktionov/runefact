package mcp

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/vgalaktionov/runefact/internal/config"
)

func setupTestProject(t *testing.T) (*ServerContext, string) {
	t.Helper()
	dir := t.TempDir()

	// Create project structure.
	for _, sub := range []string{"assets/palettes", "assets/sprites", "assets/maps", "assets/sfx", "assets/tracks", "assets/instruments"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0755); err != nil {
			t.Fatal(err)
		}
	}

	// Config.
	configData := `
[project]
name = "test"
output = "build/assets"
package = "assets"
`
	if err := os.WriteFile(filepath.Join(dir, "runefact.toml"), []byte(configData), 0644); err != nil {
		t.Fatal(err)
	}

	// Palette.
	paletteData := `name = "default"
[colors]
r = "#ff0000"
g = "#00ff00"
b = "#0000ff"
_ = "#00000000"
`
	if err := os.WriteFile(filepath.Join(dir, "assets/palettes/default.palette"), []byte(paletteData), 0644); err != nil {
		t.Fatal(err)
	}

	// Sprite.
	spriteData := `palette = "default"
grid = "2x2"
[sprite.test]
pixels = """
rg
br
"""
`
	if err := os.WriteFile(filepath.Join(dir, "assets/sprites/demo.sprite"), []byte(spriteData), 0644); err != nil {
		t.Fatal(err)
	}

	// Map.
	mapData := `tile_size = 16
[tileset]
g = "demo:test"
_ = ""

[layer.bg]
pixels = """
g_
_g
"""
`
	if err := os.WriteFile(filepath.Join(dir, "assets/maps/demo.map"), []byte(mapData), 0644); err != nil {
		t.Fatal(err)
	}

	// SFX.
	sfxData := `duration = 0.1
volume = 0.8
[[voice]]
waveform = "sine"
[voice.pitch]
start = 440
end = 220
[voice.envelope]
attack = 0.01
decay = 0.02
sustain = 0.5
release = 0.03
`
	if err := os.WriteFile(filepath.Join(dir, "assets/sfx/test.sfx"), []byte(sfxData), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.LoadConfig(filepath.Join(dir, "runefact.toml"))
	if err != nil {
		t.Fatal(err)
	}

	return &ServerContext{
		Config:      cfg,
		ProjectRoot: dir,
	}, dir
}

func TestHandleBuild(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"scope": "all"}

	result, err := ctx.handleBuild(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["success"] != true {
		t.Errorf("expected success, got: %v", data)
	}
}

func TestHandleValidate(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := ctx.handleValidate(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["valid"] != true {
		t.Errorf("expected valid, got: %v", data)
	}
}

func TestHandleInspectSprite(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"file": "demo.sprite"}

	result, err := ctx.handleInspectSprite(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["palette"] != "default" {
		t.Errorf("expected palette 'default', got: %v", data["palette"])
	}
	sprites := data["sprites"].([]any)
	if len(sprites) != 1 {
		t.Errorf("expected 1 sprite, got %d", len(sprites))
	}
}

func TestHandleInspectMap(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"file": "demo.map"}

	result, err := ctx.handleInspectMap(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["tile_size"].(float64) != 16 {
		t.Errorf("expected tile_size 16, got: %v", data["tile_size"])
	}
}

func TestHandleInspectAudio(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"file": "test.sfx"}

	result, err := ctx.handleInspectAudio(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["type"] != "sfx" {
		t.Errorf("expected type 'sfx', got: %v", data["type"])
	}
	if data["duration"].(float64) != 0.1 {
		t.Errorf("expected duration 0.1, got: %v", data["duration"])
	}
}

func TestHandleListAssets(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{}

	result, err := ctx.handleListAssets(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	count := data["count"].(float64)
	if count < 3 {
		t.Errorf("expected at least 3 assets, got %v", count)
	}
}

func TestHandlePaletteColors(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.CallToolRequest{}
	req.Params.Arguments = map[string]any{"file": "default.palette"}

	result, err := ctx.handlePaletteColors(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := result.Content[0].(mcp.TextContent).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	colors := data["colors"].(map[string]any)
	if colors["r"] != "#ff0000" {
		t.Errorf("expected red=#ff0000, got: %v", colors["r"])
	}
}

func TestHandleFormatHelp(t *testing.T) {
	ctx, _ := setupTestProject(t)

	for _, format := range []string{"palette", "sprite", "map", "instrument", "sfx", "track"} {
		req := mcp.CallToolRequest{}
		req.Params.Arguments = map[string]any{"format": format}

		result, err := ctx.handleFormatHelp(context.Background(), req)
		if err != nil {
			t.Fatalf("format_help(%s) error: %v", format, err)
		}

		text := result.Content[0].(mcp.TextContent).Text
		if len(text) < 50 {
			t.Errorf("format_help(%s) returned too little text: %d chars", format, len(text))
		}
	}
}

func TestHandleProjectStatus(t *testing.T) {
	ctx, _ := setupTestProject(t)

	req := mcp.ReadResourceRequest{}

	contents, err := ctx.handleProjectStatus(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}

	text := contents[0].(mcp.TextResourceContents).Text
	var data map[string]any
	if err := json.Unmarshal([]byte(text), &data); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	if data["project_name"] != "test" {
		t.Errorf("expected project_name 'test', got: %v", data["project_name"])
	}
}

func TestErrorResult(t *testing.T) {
	result, err := errorResult("something went wrong")
	if err != nil {
		t.Fatal(err)
	}
	if !result.IsError {
		t.Error("expected IsError=true")
	}
}
