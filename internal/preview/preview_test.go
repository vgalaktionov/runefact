package preview

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/sfx"
	"github.com/vgalaktionov/runefact/internal/track"
)

func TestCurrentFrame(t *testing.T) {
	p := &Previewer{}

	t.Run("static sprite returns 0", func(t *testing.T) {
		s := &RenderedSprite{FrameCount: 1, FPS: 0}
		p.frameTime = 5.0
		if got := p.currentFrame(s); got != 0 {
			t.Errorf("currentFrame = %d, want 0", got)
		}
	})

	t.Run("animated sprite cycles frames", func(t *testing.T) {
		s := &RenderedSprite{FrameCount: 4, FPS: 8}

		p.frameTime = 0.0
		if got := p.currentFrame(s); got != 0 {
			t.Errorf("currentFrame at t=0 = %d, want 0", got)
		}

		p.frameTime = 0.125 // 1/8 sec = 1 frame at 8fps
		if got := p.currentFrame(s); got != 1 {
			t.Errorf("currentFrame at t=0.125 = %d, want 1", got)
		}

		p.frameTime = 0.5 // 4 frames at 8fps = wraps to 0
		if got := p.currentFrame(s); got != 0 {
			t.Errorf("currentFrame at t=0.5 = %d, want 0 (wrapped)", got)
		}
	})
}

func TestStatePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "preview.json")

	st := previewState{
		Zoom:       8,
		Background: 2,
		ShowGrid:   true,
		LastFile:   "/tmp/test.sprite",
	}

	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Read back.
	readData, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var loaded previewState
	if err := json.Unmarshal(readData, &loaded); err != nil {
		t.Fatal(err)
	}

	if loaded.Zoom != 8 {
		t.Errorf("zoom = %d, want 8", loaded.Zoom)
	}
	if loaded.Background != 2 {
		t.Errorf("background = %d, want 2", loaded.Background)
	}
	if !loaded.ShowGrid {
		t.Error("show_grid should be true")
	}
	if loaded.LastFile != "/tmp/test.sprite" {
		t.Errorf("last_file = %q, want /tmp/test.sprite", loaded.LastFile)
	}
}

func TestBackgroundTypeCycling(t *testing.T) {
	bg := BackgroundDark
	bg = (bg + 1) % 3
	if bg != BackgroundLight {
		t.Errorf("expected Light, got %d", bg)
	}
	bg = (bg + 1) % 3
	if bg != BackgroundCheckerboard {
		t.Errorf("expected Checkerboard, got %d", bg)
	}
	bg = (bg + 1) % 3
	if bg != BackgroundDark {
		t.Errorf("expected Dark (wrapped), got %d", bg)
	}
}

func TestNewPreviewer(t *testing.T) {
	p := NewPreviewer("/tmp/test.sprite", "/tmp/assets", 800, 600, 44100)
	if p.zoom != 4 {
		t.Errorf("default zoom = %d, want 4", p.zoom)
	}
	if p.selected != -1 {
		t.Errorf("selected = %d, want -1", p.selected)
	}
	if p.winW != 800 || p.winH != 600 {
		t.Errorf("window size = %dx%d, want 800x600", p.winW, p.winH)
	}
	if p.filePath != "/tmp/test.sprite" {
		t.Errorf("filePath = %q", p.filePath)
	}
}

func TestHitTestSprite_Empty(t *testing.T) {
	p := &Previewer{zoom: 4, winW: 800, winH: 600}
	if got := p.hitTestSprite(100, 100); got != -1 {
		t.Errorf("hitTestSprite on empty = %d, want -1", got)
	}
}

func TestModeDetection(t *testing.T) {
	tests := []struct {
		file string
		want PreviewMode
	}{
		{"player.sprite", ModeSpritePreview},
		{"world.map", ModeMapPreview},
		{"laser.sfx", ModeSFXPreview},
		{"bgm.track", ModeMusicPreview},
	}
	for _, tt := range tests {
		p := NewPreviewer("/tmp/"+tt.file, "/tmp/assets", 800, 600, 44100)
		if p.mode != tt.want {
			t.Errorf("mode for %q = %d, want %d", tt.file, p.mode, tt.want)
		}
	}
}

func TestDownsampleWaveform(t *testing.T) {
	// 100 samples → 10 points
	samples := make([]float64, 100)
	for i := range samples {
		samples[i] = float64(i) / 100.0
	}

	result := downsampleWaveform(samples, 10)
	if len(result) != 10 {
		t.Fatalf("expected 10 points, got %d", len(result))
	}

	// Each bucket covers 10 samples; the max of [0..9]/100 is 0.09.
	if result[0] < 0 || result[0] > 0.1 {
		t.Errorf("first bucket = %f, expected small positive", result[0])
	}
}

func TestAdsrLevel(t *testing.T) {
	env := sfx.EnvelopeDef{Attack: 0.1, Decay: 0.1, Sustain: 0.5, Release: 0.2}
	duration := 1.0

	// During attack (t=0.05, attack=0.1) → 50%.
	if level := adsrLevel(env, duration, 0.05); level < 0.49 || level > 0.51 {
		t.Errorf("attack level = %f, want ~0.5", level)
	}

	// Peak of attack (t=0.1) → 1.0.
	if level := adsrLevel(env, duration, 0.1); level < 0.99 {
		t.Errorf("attack peak = %f, want ~1.0", level)
	}

	// Sustain phase (t=0.5) → 0.5.
	if level := adsrLevel(env, duration, 0.5); level < 0.49 || level > 0.51 {
		t.Errorf("sustain level = %f, want ~0.5", level)
	}

	// After release (t=1.0) → 0.
	if level := adsrLevel(env, duration, 1.0); level > 0.01 {
		t.Errorf("after release = %f, want ~0", level)
	}
}

func TestFormatNote(t *testing.T) {
	tests := []struct {
		note track.Note
		want string
	}{
		{track.Note{Type: track.NoteOn, Name: "C", Octave: 4}, "C4"},
		{track.Note{Type: track.Sustain}, "---"},
		{track.Note{Type: track.Silence}, "..."},
		{track.Note{Type: track.NoteOff}, "^^^"},
	}
	for _, tt := range tests {
		if got := formatNote(tt.note); got != tt.want {
			t.Errorf("formatNote(%v) = %q, want %q", tt.note, got, tt.want)
		}
	}
}
