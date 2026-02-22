package instrument

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/audio"
)

func TestParseInstrument_Basic(t *testing.T) {
	input := []byte(`
name = "bass"

[oscillator]
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
resonance = 0.3

[effects]
vibrato_depth = 1.0
vibrato_rate = 5.0
`)
	inst, err := ParseInstrument(input, "bass.inst")
	if err != nil {
		t.Fatal(err)
	}
	if inst.Name != "bass" {
		t.Errorf("name = %q, want bass", inst.Name)
	}
	if inst.Oscillator.Waveform != "square" {
		t.Errorf("waveform = %q, want square", inst.Oscillator.Waveform)
	}
	if inst.Envelope.Attack != 0.01 {
		t.Errorf("attack = %f, want 0.01", inst.Envelope.Attack)
	}
	if inst.Filter == nil {
		t.Fatal("filter is nil")
	}
	if inst.Filter.Cutoff != 800 {
		t.Errorf("cutoff = %f, want 800", inst.Filter.Cutoff)
	}
	if inst.Effects.VibratoDepth != 1.0 {
		t.Errorf("vibrato_depth = %f, want 1.0", inst.Effects.VibratoDepth)
	}
}

func TestParseInstrument_Minimal(t *testing.T) {
	input := []byte(`
name = "simple"

[oscillator]
waveform = "sine"

[envelope]
attack = 0
decay = 0
sustain = 1
release = 0
`)
	inst, err := ParseInstrument(input, "simple.inst")
	if err != nil {
		t.Fatal(err)
	}
	if inst.Filter != nil {
		t.Error("filter should be nil for minimal instrument")
	}
}

func TestParseInstrument_DefaultWaveform(t *testing.T) {
	input := []byte(`
name = "default"

[oscillator]

[envelope]
sustain = 1
`)
	inst, err := ParseInstrument(input, "default.inst")
	if err != nil {
		t.Fatal(err)
	}
	if inst.Oscillator.Waveform != "sine" {
		t.Errorf("default waveform = %q, want sine", inst.Oscillator.Waveform)
	}
}

func TestCreateVoice(t *testing.T) {
	inst := &Instrument{
		Oscillator: OscillatorDef{Waveform: "square", DutyCycle: 0.25},
		Envelope:   audio.ADSR{Attack: 0.01, Decay: 0.1, Sustain: 0.6, Release: 0.15},
	}
	v := inst.CreateVoice(440, 44100)
	if v == nil {
		t.Fatal("CreateVoice returned nil")
	}
	if v.Frequency != 440 {
		t.Errorf("frequency = %f, want 440", v.Frequency)
	}
}

func TestLoadInstrument(t *testing.T) {
	dir := t.TempDir()
	content := `name = "test"
[oscillator]
waveform = "triangle"
[envelope]
sustain = 0.8
`
	path := filepath.Join(dir, "test.inst")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	inst, err := LoadInstrument(path)
	if err != nil {
		t.Fatal(err)
	}
	if inst.Name != "test" {
		t.Errorf("name = %q, want test", inst.Name)
	}
}

func TestResolveInstrument(t *testing.T) {
	dir := t.TempDir()
	content := `name = "demo"
[oscillator]
waveform = "sine"
[envelope]
sustain = 1
`
	if err := os.WriteFile(filepath.Join(dir, "demo.inst"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	inst, err := ResolveInstrument("demo", []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if inst.Name != "demo" {
		t.Errorf("name = %q, want demo", inst.Name)
	}
}

func TestResolveInstrument_NotFound(t *testing.T) {
	_, err := ResolveInstrument("missing", []string{t.TempDir()})
	if err == nil {
		t.Fatal("expected error for missing instrument")
	}
}
