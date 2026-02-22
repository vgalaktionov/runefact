package manifest

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/vgalaktionov/runefact/internal/sprite"
)

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input, want string
	}{
		{"player", "Player"},
		{"game-over", "GameOver"},
		{"my_game", "MyGame"},
		{"level1", "Level1"},
		{"dark.theme", "DarkTheme"},
		{"a-b_c.d", "ABCD"},
	}
	for _, tt := range tests {
		got := ToPascalCase(tt.input)
		if got != tt.want {
			t.Errorf("ToPascalCase(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestManifestData_AddSpriteSheet(t *testing.T) {
	md := &ManifestData{Package: "assets"}
	meta := sprite.SpriteSheetMeta{
		Sprites: map[string]sprite.SpriteInfo{
			"idle": {X: 0, Y: 0, W: 16, H: 16, Frames: 2, FPS: 8},
			"jump": {X: 0, Y: 16, W: 16, H: 16, Frames: 1, FPS: 0},
		},
	}

	md.AddSpriteSheet("player.sprite", "sprites/player.png", meta)

	if len(md.SpriteSheets) != 1 {
		t.Fatalf("got %d sheets, want 1", len(md.SpriteSheets))
	}
	if md.SpriteSheets[0].Const != "SpriteSheetPlayer" {
		t.Errorf("const = %q, want SpriteSheetPlayer", md.SpriteSheets[0].Const)
	}
	if len(md.Sprites) != 2 {
		t.Fatalf("got %d sprites, want 2", len(md.Sprites))
	}
}

func TestManifestData_AddMap(t *testing.T) {
	md := &ManifestData{}
	md.AddMap("level1.map", "maps/level1.json")
	if md.Maps[0].Const != "MapLevel1" {
		t.Errorf("const = %q, want MapLevel1", md.Maps[0].Const)
	}
}

func TestManifestData_AddAudio(t *testing.T) {
	md := &ManifestData{}
	md.AddAudio("jump.sfx", "audio/jump.wav")
	md.AddAudio("theme.track", "audio/theme.wav")
	if md.Audio[0].Const != "SFXJump" {
		t.Errorf("sfx const = %q, want SFXJump", md.Audio[0].Const)
	}
	if md.Audio[1].Const != "TrackTheme" {
		t.Errorf("track const = %q, want TrackTheme", md.Audio[1].Const)
	}
}

func TestGenerate_ValidGo(t *testing.T) {
	md := &ManifestData{
		Package: "assets",
		SpriteSheets: []SheetEntry{
			{Const: "SpriteSheetPlayer", Path: "sprites/player.png"},
		},
		Sprites: []SpriteEntry{
			{Key: "player:idle", Sheet: "SpriteSheetPlayer", X: 0, Y: 0, W: 16, H: 16, Frames: 2, FPS: 8},
		},
		Maps: []AssetEntry{
			{Const: "MapLevel1", Path: "maps/level1.json"},
		},
		Audio: []AssetEntry{
			{Const: "SFXJump", Path: "audio/jump.wav"},
		},
	}

	dir := t.TempDir()
	outputPath := filepath.Join(dir, "manifest.go")
	if err := Generate(md, outputPath); err != nil {
		t.Fatal(err)
	}

	// Read and check content.
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)

	if !strings.Contains(content, "DO NOT EDIT") {
		t.Error("missing DO NOT EDIT header")
	}
	if !strings.Contains(content, "package assets") {
		t.Error("missing package declaration")
	}
	if !strings.Contains(content, "SpriteSheetPlayer") {
		t.Error("missing SpriteSheetPlayer constant")
	}
	if !strings.Contains(content, `"player:idle"`) {
		t.Error("missing sprite entry")
	}
	if !strings.Contains(content, "MapLevel1") {
		t.Error("missing MapLevel1 constant")
	}
	if !strings.Contains(content, "SFXJump") {
		t.Error("missing SFXJump constant")
	}

	// Verify it's valid Go by running go vet.
	// Write a go.mod so `go vet` can parse the file.
	goMod := filepath.Join(dir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module test\ngo 1.23\n"), 0644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("go", "vet", "./...")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Errorf("generated Go is invalid: %v\n%s", err, out)
	}
}
