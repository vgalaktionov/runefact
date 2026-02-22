package instrument

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/vgalaktionov/runefact/internal/audio"
)

// Instrument represents a parsed .inst file.
type Instrument struct {
	Name       string
	Oscillator OscillatorDef
	Envelope   audio.ADSR
	Filter     *FilterDef
	Effects    EffectsDef
}

// OscillatorDef defines oscillator parameters.
type OscillatorDef struct {
	Waveform  string  `toml:"waveform"`
	DutyCycle float64 `toml:"duty_cycle"`
}

// FilterDef defines optional filter parameters.
type FilterDef struct {
	Type      string  `toml:"type"`
	Cutoff    float64 `toml:"cutoff"`
	Resonance float64 `toml:"resonance"`
}

// EffectsDef defines optional effects parameters.
type EffectsDef struct {
	VibratoDepth float64 `toml:"vibrato_depth"`
	VibratoRate  float64 `toml:"vibrato_rate"`
	PitchSweep   float64 `toml:"pitch_sweep"`
	Distortion   float64 `toml:"distortion"`
}

// rawInstrument is the TOML-level structure.
type rawInstrument struct {
	Name       string        `toml:"name"`
	Oscillator OscillatorDef `toml:"oscillator"`
	Envelope   rawEnvelope   `toml:"envelope"`
	Filter     *FilterDef    `toml:"filter"`
	Effects    EffectsDef    `toml:"effects"`
}

type rawEnvelope struct {
	Attack  float64 `toml:"attack"`
	Decay   float64 `toml:"decay"`
	Sustain float64 `toml:"sustain"`
	Release float64 `toml:"release"`
}

// ParseInstrument parses .inst file content.
func ParseInstrument(data []byte, filename string) (*Instrument, error) {
	var raw rawInstrument
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	if raw.Oscillator.Waveform == "" {
		raw.Oscillator.Waveform = "sine"
	}

	inst := &Instrument{
		Name:       raw.Name,
		Oscillator: raw.Oscillator,
		Envelope: audio.ADSR{
			Attack:  raw.Envelope.Attack,
			Decay:   raw.Envelope.Decay,
			Sustain: raw.Envelope.Sustain,
			Release: raw.Envelope.Release,
		},
		Filter:  raw.Filter,
		Effects: raw.Effects,
	}

	return inst, nil
}

// LoadInstrument reads and parses an .inst file from disk.
func LoadInstrument(path string) (*Instrument, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading instrument: %w", err)
	}
	return ParseInstrument(data, filepath.Base(path))
}

// ResolveInstrument searches for an instrument by name in the given directories.
func ResolveInstrument(name string, searchPaths []string) (*Instrument, error) {
	for _, dir := range searchPaths {
		path := filepath.Join(dir, name+".inst")
		if _, err := os.Stat(path); err == nil {
			return LoadInstrument(path)
		}
	}
	return nil, fmt.Errorf("instrument %q not found in search paths: %v", name, searchPaths)
}

// CreateVoice creates an audio.Voice from this instrument at the given frequency.
func (inst *Instrument) CreateVoice(frequency float64, sampleRate int) *audio.Voice {
	v := &audio.Voice{
		Osc:       audio.NewOscillator(inst.Oscillator.Waveform, inst.Oscillator.DutyCycle),
		Env:       inst.Envelope,
		Frequency: frequency,
	}

	v.Vibrato.Depth = inst.Effects.VibratoDepth
	v.Vibrato.Rate = inst.Effects.VibratoRate

	if inst.Filter != nil {
		v.Filter = audio.NewBiquadFilter(
			audio.FilterType(inst.Filter.Type),
			inst.Filter.Cutoff,
			inst.Filter.Resonance,
			sampleRate,
		)
	}

	return v
}
