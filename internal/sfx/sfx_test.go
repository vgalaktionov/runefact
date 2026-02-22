package sfx

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/audio"
)

func TestParseSFX_Basic(t *testing.T) {
	input := []byte(`
duration = 0.15
volume = 0.8

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.3
release = 0.1

[voice.pitch]
start = 200
end = 600
curve = "exponential"
`)
	s, err := ParseSFX(input, "jump.sfx")
	if err != nil {
		t.Fatal(err)
	}
	if s.Duration != 0.15 {
		t.Errorf("duration = %f, want 0.15", s.Duration)
	}
	if s.Volume != 0.8 {
		t.Errorf("volume = %f, want 0.8", s.Volume)
	}
	if len(s.Voices) != 1 {
		t.Fatalf("got %d voices, want 1", len(s.Voices))
	}
	v := s.Voices[0]
	if v.Waveform != "square" {
		t.Errorf("waveform = %q, want square", v.Waveform)
	}
	if v.Pitch.Start != 200 || v.Pitch.End != 600 {
		t.Errorf("pitch = %f -> %f, want 200 -> 600", v.Pitch.Start, v.Pitch.End)
	}
}

func TestParseSFX_MultiVoice(t *testing.T) {
	input := []byte(`
duration = 0.5
volume = 1.0

[[voice]]
waveform = "noise"
[voice.envelope]
decay = 0.3

[[voice]]
waveform = "sine"
[voice.envelope]
decay = 0.1
[voice.pitch]
start = 120
end = 40
`)
	s, err := ParseSFX(input, "explosion.sfx")
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Voices) != 2 {
		t.Fatalf("got %d voices, want 2", len(s.Voices))
	}
}

func TestParseSFX_InvalidDuration(t *testing.T) {
	input := []byte(`duration = 0`)
	_, err := ParseSFX(input, "bad.sfx")
	if err == nil {
		t.Fatal("expected error for duration=0")
	}
}

func TestParseSFX_DefaultVolume(t *testing.T) {
	input := []byte(`
duration = 0.1

[[voice]]
waveform = "sine"
[voice.envelope]
sustain = 1
[voice.pitch]
start = 440
`)
	s, err := ParseSFX(input, "test.sfx")
	if err != nil {
		t.Fatal(err)
	}
	if s.Volume != 1.0 {
		t.Errorf("default volume = %f, want 1.0", s.Volume)
	}
}

func TestSFX_Render(t *testing.T) {
	s := &SFX{
		Duration: 0.1,
		Volume:   0.8,
		Voices: []VoiceDef{
			{
				Waveform: "square",
				Envelope: EnvelopeDef{Sustain: 0.5, Release: 0.02},
				Pitch:    PitchDef{Start: 440, End: 440},
			},
		},
	}
	samples, warnings := s.Render(44100)
	if len(samples) != 4410 {
		t.Errorf("got %d samples, want 4410", len(samples))
	}
	// Check no NaN/Inf warnings.
	for _, w := range warnings {
		if w.Message != "" {
			t.Logf("warning: %s", w.Message)
		}
	}
	// Check samples are in valid range after safety processing.
	for i, sample := range samples {
		if math.IsNaN(sample) || math.IsInf(sample, 0) {
			t.Errorf("sample %d is NaN/Inf", i)
			break
		}
	}
}

func TestLoadSFX(t *testing.T) {
	dir := t.TempDir()
	content := `duration = 0.1
volume = 1.0
[[voice]]
waveform = "sine"
[voice.envelope]
sustain = 1
[voice.pitch]
start = 440
`
	path := filepath.Join(dir, "test.sfx")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadSFX(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Duration != 0.1 {
		t.Errorf("duration = %f, want 0.1", s.Duration)
	}
}

func TestWriteWAV(t *testing.T) {
	// Generate simple samples.
	samples := make([]float64, 4410) // 0.1s at 44100
	for i := range samples {
		samples[i] = math.Sin(2 * math.Pi * 440 * float64(i) / 44100)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "test.wav")
	if err := audio.WriteWAV(path, samples, 44100, 16); err != nil {
		t.Fatal(err)
	}

	// Verify file exists and has correct header.
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	// 44 bytes header + 4410 samples * 2 bytes = 8864 bytes
	expectedSize := int64(44 + 4410*2)
	if info.Size() != expectedSize {
		t.Errorf("file size = %d, want %d", info.Size(), expectedSize)
	}
}
