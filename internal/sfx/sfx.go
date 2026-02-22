package sfx

import (
	"fmt"
	"os"
	"path/filepath"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/vgalaktionov/runefact/internal/audio"
)

// SFX represents a parsed .sfx file.
type SFX struct {
	Duration float64
	Volume   float64
	Voices   []VoiceDef
}

// VoiceDef defines a single voice in an SFX.
type VoiceDef struct {
	Waveform  string  `toml:"waveform"`
	DutyCycle float64 `toml:"duty_cycle"`
	Envelope  EnvelopeDef
	Pitch     PitchDef
	Filter    *FilterDef
	Effects   EffectsDef
}

// EnvelopeDef is the TOML envelope section.
type EnvelopeDef struct {
	Attack  float64 `toml:"attack"`
	Decay   float64 `toml:"decay"`
	Sustain float64 `toml:"sustain"`
	Release float64 `toml:"release"`
}

// PitchDef is the TOML pitch section.
type PitchDef struct {
	Start float64 `toml:"start"`
	End   float64 `toml:"end"`
	Curve string  `toml:"curve"`
}

// FilterDef is the TOML filter section for voice-level filters.
type FilterDef struct {
	Type        string  `toml:"type"`
	CutoffStart float64 `toml:"cutoff_start"`
	CutoffEnd   float64 `toml:"cutoff_end"`
	Cutoff      float64 `toml:"cutoff"`
	Resonance   float64 `toml:"resonance"`
	Curve       string  `toml:"curve"`
}

// EffectsDef is the TOML effects section.
type EffectsDef struct {
	VibratoDepth float64 `toml:"vibrato_depth"`
	VibratoRate  float64 `toml:"vibrato_rate"`
}

// rawSFX is the TOML-level structure.
type rawSFX struct {
	Duration float64       `toml:"duration"`
	Volume   float64       `toml:"volume"`
	Voice    []rawVoiceDef `toml:"voice"`
}

type rawVoiceDef struct {
	Waveform  string      `toml:"waveform"`
	DutyCycle float64     `toml:"duty_cycle"`
	Envelope  EnvelopeDef `toml:"envelope"`
	Pitch     PitchDef    `toml:"pitch"`
	Filter    *FilterDef  `toml:"filter"`
	Effects   EffectsDef  `toml:"effects"`
}

// ParseSFX parses .sfx file content.
func ParseSFX(data []byte, filename string) (*SFX, error) {
	var raw rawSFX
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	if raw.Duration <= 0 {
		return nil, fmt.Errorf("%s: duration must be positive", filename)
	}
	if raw.Volume <= 0 {
		raw.Volume = 1.0
	}

	s := &SFX{
		Duration: raw.Duration,
		Volume:   raw.Volume,
	}

	for _, rv := range raw.Voice {
		if rv.Waveform == "" {
			rv.Waveform = "sine"
		}
		s.Voices = append(s.Voices, VoiceDef{
			Waveform:  rv.Waveform,
			DutyCycle: rv.DutyCycle,
			Envelope:  rv.Envelope,
			Pitch:     rv.Pitch,
			Filter:    rv.Filter,
			Effects:   rv.Effects,
		})
	}

	return s, nil
}

// LoadSFX reads and parses an .sfx file from disk.
func LoadSFX(path string) (*SFX, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading sfx: %w", err)
	}
	return ParseSFX(data, filepath.Base(path))
}

// Render generates audio samples for the SFX.
func (s *SFX) Render(sampleRate int) ([]float64, []audio.Warning) {
	numSamples := int(s.Duration * float64(sampleRate))
	mixed := make([]float64, numSamples)
	var warnings []audio.Warning

	for _, vd := range s.Voices {
		voice := buildVoice(vd, sampleRate)
		voiceSamples := audio.RenderVoice(voice, s.Duration, sampleRate)

		for i := 0; i < len(mixed) && i < len(voiceSamples); i++ {
			mixed[i] += voiceSamples[i]
		}
	}

	// Apply volume.
	for i := range mixed {
		mixed[i] *= s.Volume
	}

	// Apply audio safety.
	mixed, safetyWarnings := audio.ProcessSafety(mixed, sampleRate)
	warnings = append(warnings, safetyWarnings...)

	return mixed, warnings
}

func buildVoice(vd VoiceDef, sampleRate int) *audio.Voice {
	v := &audio.Voice{
		Osc: audio.NewOscillator(vd.Waveform, vd.DutyCycle),
		Env: audio.ADSR{
			Attack:  vd.Envelope.Attack,
			Decay:   vd.Envelope.Decay,
			Sustain: vd.Envelope.Sustain,
			Release: vd.Envelope.Release,
		},
		PitchStart: vd.Pitch.Start,
		PitchEnd:   vd.Pitch.End,
		PitchCurve: audio.CurveType(vd.Pitch.Curve),
	}

	if v.PitchStart == 0 {
		v.PitchStart = 440
	}
	if v.PitchEnd == 0 {
		v.PitchEnd = v.PitchStart
	}

	v.Vibrato.Depth = vd.Effects.VibratoDepth
	v.Vibrato.Rate = vd.Effects.VibratoRate

	if vd.Filter != nil {
		cutoff := vd.Filter.Cutoff
		if cutoff == 0 {
			cutoff = vd.Filter.CutoffStart
		}
		if cutoff == 0 {
			cutoff = 1000
		}
		v.Filter = audio.NewBiquadFilter(
			audio.FilterType(vd.Filter.Type),
			cutoff,
			vd.Filter.Resonance,
			sampleRate,
		)
	}

	return v
}
