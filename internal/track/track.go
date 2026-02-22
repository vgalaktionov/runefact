package track

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/vgalaktionov/runefact/internal/audio"
	"github.com/vgalaktionov/runefact/internal/instrument"
)

// Track represents a parsed .track file.
type Track struct {
	Tempo        int
	TicksPerBeat int
	Loop         bool
	LoopStart    int
	Channels     []Channel
	Patterns     map[string]*Pattern
	Sequence     []string
}

// Channel defines a named channel with an instrument reference and volume.
type Channel struct {
	Name       string `toml:"name"`
	Instrument string `toml:"instrument"`
	Volume     float64 `toml:"volume"`
}

// Pattern holds rows of notes, one per tick.
type Pattern struct {
	Name  string
	Ticks int
	Rows  [][]Note // [tick][channel]
}

// NoteType classifies what a cell in a pattern represents.
type NoteType int

const (
	NoteOn    NoteType = iota // e.g. C4
	Sustain                   // ---
	Silence                   // ...
	NoteOff                   // ^^^
)

// Note represents a single cell in a pattern.
type Note struct {
	Type    NoteType
	Name    string  // e.g. "C", "C#", "D"
	Octave  int
	Effects []Effect
}

// Effect is a per-note effect modifier.
type Effect struct {
	Type  byte   // 'v' velocity, '>' slide up, '<' slide down, '~' vibrato, 'a' arpeggio
	Value int
}

// Freq returns the frequency in Hz for a NoteOn.
func (n Note) Freq() float64 {
	if n.Type != NoteOn {
		return 0
	}
	midi := noteToMIDI(n.Name, n.Octave)
	return audio.MIDIToFreq(midi)
}

// noteToMIDI converts a note name and octave to MIDI number.
func noteToMIDI(name string, octave int) int {
	semitones := map[string]int{
		"C": 0, "C#": 1, "D": 2, "D#": 3, "E": 4, "F": 5,
		"F#": 6, "G": 7, "G#": 8, "A": 9, "A#": 10, "B": 11,
	}
	s, ok := semitones[name]
	if !ok {
		return 60 // default to C4
	}
	return (octave+1)*12 + s
}

// rawTrack is the TOML-level structure.
type rawTrack struct {
	Tempo        int        `toml:"tempo"`
	TicksPerBeat int        `toml:"ticks_per_beat"`
	Loop         bool       `toml:"loop"`
	LoopStart    int        `toml:"loop_start"`
	Channel      []Channel  `toml:"channel"`
	Pattern      map[string]rawPattern
	Song         rawSong    `toml:"song"`
}

type rawPattern struct {
	Ticks int    `toml:"ticks"`
	Data  string `toml:"data"`
}

type rawSong struct {
	Sequence []string `toml:"sequence"`
}

// ParseTrack parses .track file content.
func ParseTrack(data []byte, filename string) (*Track, error) {
	var raw rawTrack
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	if raw.Tempo <= 0 {
		return nil, fmt.Errorf("%s: tempo must be positive", filename)
	}
	if raw.TicksPerBeat <= 0 {
		raw.TicksPerBeat = 4
	}

	t := &Track{
		Tempo:        raw.Tempo,
		TicksPerBeat: raw.TicksPerBeat,
		Loop:         raw.Loop,
		LoopStart:    raw.LoopStart,
		Channels:     raw.Channel,
		Patterns:     make(map[string]*Pattern),
		Sequence:     raw.Song.Sequence,
	}

	numChannels := len(t.Channels)

	for name, rp := range raw.Pattern {
		pattern, err := parsePattern(name, rp, numChannels, filename)
		if err != nil {
			return nil, err
		}
		t.Patterns[name] = pattern
	}

	// Validate sequence references.
	for _, pname := range t.Sequence {
		if _, ok := t.Patterns[pname]; !ok {
			return nil, fmt.Errorf("%s: unknown pattern %q in sequence", filename, pname)
		}
	}

	return t, nil
}

func parsePattern(name string, raw rawPattern, numChannels int, filename string) (*Pattern, error) {
	lines := strings.Split(strings.TrimSpace(raw.Data), "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("%s: pattern %q: empty data", filename, name)
	}

	// First line is channel header — skip it.
	dataLines := lines[1:]

	ticks := raw.Ticks
	if ticks <= 0 {
		ticks = len(dataLines)
	}

	p := &Pattern{
		Name:  name,
		Ticks: ticks,
		Rows:  make([][]Note, len(dataLines)),
	}

	for i, line := range dataLines {
		cols := splitColumns(line)
		if numChannels > 0 && len(cols) != numChannels {
			return nil, fmt.Errorf("%s: pattern %q row %d: got %d columns, expected %d",
				filename, name, i+1, len(cols), numChannels)
		}

		row := make([]Note, len(cols))
		for j, cell := range cols {
			note, err := parseNote(strings.TrimSpace(cell))
			if err != nil {
				return nil, fmt.Errorf("%s: pattern %q row %d col %d: %w",
					filename, name, i+1, j+1, err)
			}
			row[j] = note
		}
		p.Rows[i] = row
	}

	return p, nil
}

func splitColumns(line string) []string {
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

func parseNote(cell string) (Note, error) {
	if cell == "---" {
		return Note{Type: Sustain}, nil
	}
	if cell == "..." {
		return Note{Type: Silence}, nil
	}
	if cell == "^^^" {
		return Note{Type: NoteOff}, nil
	}

	// Parse note: C4, C#5, D#3, possibly with effects suffix.
	parts := strings.Fields(cell)
	if len(parts) == 0 {
		return Note{Type: Silence}, nil
	}

	noteStr := parts[0]
	n := Note{Type: NoteOn}

	// Parse note name and octave.
	i := 0
	if i < len(noteStr) {
		n.Name = string(noteStr[i])
		i++
	}
	if i < len(noteStr) && noteStr[i] == '#' {
		n.Name += "#"
		i++
	}
	if i < len(noteStr) {
		octave, err := strconv.Atoi(noteStr[i:])
		if err != nil {
			return Note{}, fmt.Errorf("invalid note %q: cannot parse octave", noteStr)
		}
		n.Octave = octave
	}

	// Parse effects from remaining parts.
	for _, eff := range parts[1:] {
		if len(eff) < 2 {
			continue
		}
		e := Effect{Type: eff[0]}
		val, err := strconv.ParseInt(eff[1:], 16, 32)
		if err != nil {
			return Note{}, fmt.Errorf("invalid effect %q", eff)
		}
		e.Value = int(val)
		n.Effects = append(n.Effects, e)
	}

	return n, nil
}

// LoadTrack reads and parses a .track file from disk.
func LoadTrack(path string) (*Track, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading track: %w", err)
	}
	return ParseTrack(data, filepath.Base(path))
}

// channelState tracks the current note state for a single channel.
type channelState struct {
	active    bool
	voice     *audio.Voice
	noteOnTime float64
	releasing bool
	releaseTime float64
}

// Render generates audio samples for the track.
func (t *Track) Render(instruments map[string]*instrument.Instrument, sampleRate int) ([]float64, error) {
	samplesPerTick := float64(sampleRate) * 60.0 / float64(t.Tempo) / float64(t.TicksPerBeat)
	intSamplesPerTick := int(math.Round(samplesPerTick))

	// Calculate total ticks.
	totalTicks := 0
	for _, pname := range t.Sequence {
		p := t.Patterns[pname]
		totalTicks += len(p.Rows)
	}

	totalSamples := totalTicks * intSamplesPerTick
	mixed := make([]float64, totalSamples)

	states := make([]channelState, len(t.Channels))
	sampleOffset := 0
	globalTime := 0.0

	for _, pname := range t.Sequence {
		pattern := t.Patterns[pname]
		for _, row := range pattern.Rows {
			for chIdx := 0; chIdx < len(row) && chIdx < len(t.Channels); chIdx++ {
				note := row[chIdx]
				ch := t.Channels[chIdx]
				state := &states[chIdx]

				switch note.Type {
				case NoteOn:
					inst, ok := instruments[ch.Instrument]
					if !ok {
						continue
					}
					freq := note.Freq()

					// Apply velocity effect.
					volume := ch.Volume
					for _, eff := range note.Effects {
						if eff.Type == 'v' {
							volume = ch.Volume * float64(eff.Value) / 15.0
						}
					}

					state.voice = inst.CreateVoice(freq, sampleRate)
					state.active = true
					state.noteOnTime = globalTime
					state.releasing = false

					// Render this tick.
					for s := 0; s < intSamplesPerTick; s++ {
						idx := sampleOffset + s
						if idx >= len(mixed) {
							break
						}
						t := float64(s) / float64(sampleRate)
						sample := renderVoiceSample(state.voice, t, 10.0) // long noteOn for sustain
						mixed[idx] += sample * volume
					}

				case Sustain:
					if state.active && state.voice != nil && !state.releasing {
						elapsed := globalTime - state.noteOnTime
						for s := 0; s < intSamplesPerTick; s++ {
							idx := sampleOffset + s
							if idx >= len(mixed) {
								break
							}
							t := elapsed + float64(s)/float64(sampleRate)
							sample := renderVoiceSample(state.voice, t, 10.0)
							mixed[idx] += sample * ch.Volume
						}
					}

				case NoteOff:
					state.releasing = true
					state.active = false

				case Silence:
					// Do nothing.
				}
			}

			sampleOffset += intSamplesPerTick
			globalTime += float64(intSamplesPerTick) / float64(sampleRate)
		}
	}

	// Apply safety.
	mixed, _ = audio.ProcessSafety(mixed, sampleRate)

	return mixed, nil
}

func renderVoiceSample(v *audio.Voice, t, noteOnDur float64) float64 {
	if v == nil {
		return 0
	}
	freq := v.Frequency
	phase := math.Mod(t*freq, 1.0)
	sample := v.Osc.Sample(phase)
	sample *= v.Env.Level(t, noteOnDur)
	return sample
}
