package track

import (
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/audio"
	"github.com/vgalaktionov/runefact/internal/instrument"
)

func TestParseNote_NoteOn(t *testing.T) {
	tests := []struct {
		input string
		name  string
		oct   int
	}{
		{"C4", "C", 4},
		{"C#5", "C#", 5},
		{"D#3", "D#", 3},
		{"A0", "A", 0},
		{"B9", "B", 9},
		{"F#5", "F#", 5},
	}
	for _, tt := range tests {
		n, err := parseNote(tt.input)
		if err != nil {
			t.Errorf("parseNote(%q): %v", tt.input, err)
			continue
		}
		if n.Type != NoteOn {
			t.Errorf("parseNote(%q) type = %d, want NoteOn", tt.input, n.Type)
		}
		if n.Name != tt.name || n.Octave != tt.oct {
			t.Errorf("parseNote(%q) = %s%d, want %s%d", tt.input, n.Name, n.Octave, tt.name, tt.oct)
		}
	}
}

func TestParseNote_Special(t *testing.T) {
	n, _ := parseNote("---")
	if n.Type != Sustain {
		t.Error("--- should be Sustain")
	}
	n, _ = parseNote("...")
	if n.Type != Silence {
		t.Error("... should be Silence")
	}
	n, _ = parseNote("^^^")
	if n.Type != NoteOff {
		t.Error("^^^ should be NoteOff")
	}
}

func TestParseNote_Effects(t *testing.T) {
	n, err := parseNote("C4 v08")
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Effects) != 1 {
		t.Fatalf("got %d effects, want 1", len(n.Effects))
	}
	if n.Effects[0].Type != 'v' || n.Effects[0].Value != 8 {
		t.Errorf("effect = %c%02x, want v08", n.Effects[0].Type, n.Effects[0].Value)
	}
}

func TestParseNote_MultipleEffects(t *testing.T) {
	n, err := parseNote("C4 v0F >02")
	if err != nil {
		t.Fatal(err)
	}
	if len(n.Effects) != 2 {
		t.Fatalf("got %d effects, want 2", len(n.Effects))
	}
}

func TestNoteFreq(t *testing.T) {
	n := Note{Type: NoteOn, Name: "A", Octave: 4}
	f := n.Freq()
	if math.Abs(f-440) > 0.1 {
		t.Errorf("A4 freq = %f, want 440", f)
	}
}

func TestParseTrack_Basic(t *testing.T) {
	input := []byte(`
tempo = 120
ticks_per_beat = 4
loop = true
loop_start = 0

[[channel]]
name = "melody"
instrument = "demo"
volume = 0.8

[pattern.main]
ticks = 4
data = """
melody
C4
---
E4
---
"""

[song]
sequence = ["main"]
`)
	tr, err := ParseTrack(input, "test.track")
	if err != nil {
		t.Fatal(err)
	}
	if tr.Tempo != 120 {
		t.Errorf("tempo = %d, want 120", tr.Tempo)
	}
	if tr.TicksPerBeat != 4 {
		t.Errorf("ticks_per_beat = %d, want 4", tr.TicksPerBeat)
	}
	if !tr.Loop {
		t.Error("loop should be true")
	}
	if len(tr.Channels) != 1 {
		t.Fatalf("got %d channels, want 1", len(tr.Channels))
	}
	if len(tr.Patterns) != 1 {
		t.Fatalf("got %d patterns, want 1", len(tr.Patterns))
	}
	p := tr.Patterns["main"]
	if len(p.Rows) != 4 {
		t.Errorf("got %d rows, want 4", len(p.Rows))
	}
}

func TestParseTrack_MultiChannel(t *testing.T) {
	input := []byte(`
tempo = 120
ticks_per_beat = 4

[[channel]]
name = "melody"
instrument = "lead"
volume = 0.8

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.7

[pattern.intro]
ticks = 4
data = """
melody   | bass
C4       | C2
...      | ...
E4       | C2
...      | ...
"""

[song]
sequence = ["intro"]
`)
	tr, err := ParseTrack(input, "test.track")
	if err != nil {
		t.Fatal(err)
	}
	if len(tr.Channels) != 2 {
		t.Fatalf("got %d channels, want 2", len(tr.Channels))
	}
	p := tr.Patterns["intro"]
	if len(p.Rows[0]) != 2 {
		t.Errorf("row 0 has %d cols, want 2", len(p.Rows[0]))
	}
}

func TestParseTrack_ColumnMismatch(t *testing.T) {
	input := []byte(`
tempo = 120

[[channel]]
name = "a"
instrument = "x"
volume = 1

[[channel]]
name = "b"
instrument = "x"
volume = 1

[pattern.bad]
ticks = 2
data = """
a | b
C4
C4
"""

[song]
sequence = ["bad"]
`)
	_, err := ParseTrack(input, "test.track")
	if err == nil {
		t.Fatal("expected column count mismatch error")
	}
}

func TestParseTrack_UnknownPatternInSequence(t *testing.T) {
	input := []byte(`
tempo = 120

[[channel]]
name = "a"
instrument = "x"
volume = 1

[pattern.main]
ticks = 1
data = """
a
C4
"""

[song]
sequence = ["main", "missing"]
`)
	_, err := ParseTrack(input, "test.track")
	if err == nil {
		t.Fatal("expected unknown pattern error")
	}
}

func TestParseTrack_InvalidTempo(t *testing.T) {
	input := []byte(`tempo = 0`)
	_, err := ParseTrack(input, "test.track")
	if err == nil {
		t.Fatal("expected error for tempo=0")
	}
}

func TestTrack_Render(t *testing.T) {
	inst := &instrument.Instrument{
		Name:       "demo",
		Oscillator: instrument.OscillatorDef{Waveform: "sine"},
		Envelope:   audio.ADSR{Attack: 0, Decay: 0, Sustain: 1, Release: 0.01},
	}

	tr := &Track{
		Tempo:        120,
		TicksPerBeat: 4,
		Channels: []Channel{
			{Name: "melody", Instrument: "demo", Volume: 0.8},
		},
		Patterns: map[string]*Pattern{
			"main": {
				Name:  "main",
				Ticks: 4,
				Rows: [][]Note{
					{{Type: NoteOn, Name: "C", Octave: 4}},
					{{Type: Sustain}},
					{{Type: NoteOff}},
					{{Type: Silence}},
				},
			},
		},
		Sequence: []string{"main"},
	}

	instruments := map[string]*instrument.Instrument{"demo": inst}
	samples, err := tr.Render(instruments, 44100)
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) == 0 {
		t.Fatal("no samples rendered")
	}
	// Check no NaN/Inf.
	for i, s := range samples {
		if math.IsNaN(s) || math.IsInf(s, 0) {
			t.Errorf("sample %d is NaN/Inf", i)
			break
		}
	}
}

func TestLoadTrack(t *testing.T) {
	dir := t.TempDir()
	content := `tempo = 120
ticks_per_beat = 4
[[channel]]
name = "m"
instrument = "demo"
volume = 1
[pattern.p]
ticks = 1
data = """
m
C4
"""
[song]
sequence = ["p"]
`
	path := filepath.Join(dir, "test.track")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	tr, err := LoadTrack(path)
	if err != nil {
		t.Fatal(err)
	}
	if tr.Tempo != 120 {
		t.Errorf("tempo = %d, want 120", tr.Tempo)
	}
}
