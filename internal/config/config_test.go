package config

import (
	"testing"
)

func TestParseConfig_Valid(t *testing.T) {
	input := []byte(`
[project]
name = "my-game"
output = "build/assets"
package = "assets"

[defaults]
sprite_size = 16
sample_rate = 44100
bit_depth = 16

[preview]
window_width = 800
window_height = 600
background = "#1a1a2e"
pixel_scale = 4
audio_volume = 0.5
`)
	cfg, err := ParseConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Name != "my-game" {
		t.Errorf("project.name = %q, want %q", cfg.Project.Name, "my-game")
	}
	if cfg.Project.Output != "build/assets" {
		t.Errorf("project.output = %q, want %q", cfg.Project.Output, "build/assets")
	}
	if cfg.Defaults.SpriteSize != 16 {
		t.Errorf("defaults.sprite_size = %d, want 16", cfg.Defaults.SpriteSize)
	}
	if cfg.Preview.AudioVolume != 0.5 {
		t.Errorf("preview.audio_volume = %f, want 0.5", cfg.Preview.AudioVolume)
	}
}

func TestParseConfig_Defaults(t *testing.T) {
	input := []byte(`
[project]
name = "minimal"
`)
	cfg, err := ParseConfig(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Project.Output != "build/assets" {
		t.Errorf("default output = %q, want %q", cfg.Project.Output, "build/assets")
	}
	if cfg.Project.Package != "assets" {
		t.Errorf("default package = %q, want %q", cfg.Project.Package, "assets")
	}
	if cfg.Defaults.SpriteSize != 16 {
		t.Errorf("default sprite_size = %d, want 16", cfg.Defaults.SpriteSize)
	}
	if cfg.Defaults.SampleRate != 44100 {
		t.Errorf("default sample_rate = %d, want 44100", cfg.Defaults.SampleRate)
	}
	if cfg.Defaults.BitDepth != 16 {
		t.Errorf("default bit_depth = %d, want 16", cfg.Defaults.BitDepth)
	}
	if cfg.Preview.WindowWidth != 800 {
		t.Errorf("default window_width = %d, want 800", cfg.Preview.WindowWidth)
	}
	if cfg.Preview.PixelScale != 4 {
		t.Errorf("default pixel_scale = %d, want 4", cfg.Preview.PixelScale)
	}
	if cfg.Preview.AudioVolume != 0.5 {
		t.Errorf("default audio_volume = %f, want 0.5", cfg.Preview.AudioVolume)
	}
}

func TestParseConfig_InvalidBitDepth(t *testing.T) {
	input := []byte(`
[defaults]
bit_depth = 32
`)
	_, err := ParseConfig(input)
	if err == nil {
		t.Fatal("expected validation error for bit_depth=32")
	}
}

func TestParseConfig_InvalidAudioVolume(t *testing.T) {
	input := []byte(`
[preview]
audio_volume = 1.5
`)
	_, err := ParseConfig(input)
	if err == nil {
		t.Fatal("expected validation error for audio_volume=1.5")
	}
}

func TestParseConfig_MalformedTOML(t *testing.T) {
	input := []byte(`[project
broken toml`)
	_, err := ParseConfig(input)
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
}

func TestLoadConfig_MissingFile(t *testing.T) {
	_, err := LoadConfig("/nonexistent/runefact.toml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
