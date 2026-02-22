package audio

import "math"

// CurveType defines pitch/filter sweep interpolation.
type CurveType string

const (
	CurveLinear      CurveType = "linear"
	CurveExponential CurveType = "exponential"
	CurveLogarithmic CurveType = "logarithmic"
)

// Voice combines an oscillator, envelope, and modulation parameters.
type Voice struct {
	Osc        Oscillator
	Env        ADSR
	Frequency  float64 // base frequency Hz
	PitchStart float64 // Hz (0 = use Frequency)
	PitchEnd   float64 // Hz (0 = use Frequency)
	PitchCurve CurveType
	Vibrato    struct {
		Depth float64 // semitones
		Rate  float64 // Hz
	}
	Filter *BiquadFilter
}

// Interpolate returns a value between start and end based on t [0,1] and curve type.
func Interpolate(start, end, t float64, curve CurveType) float64 {
	if t <= 0 {
		return start
	}
	if t >= 1 {
		return end
	}
	switch curve {
	case CurveExponential:
		// Exponential: start * (end/start)^t
		if start <= 0 {
			return end * t
		}
		return start * math.Pow(end/start, t)
	case CurveLogarithmic:
		// Logarithmic: fast start, slow finish.
		lt := math.Log1p(t*9) / math.Log(10) // map [0,1] to [0,1] log curve
		return start + (end-start)*lt
	default: // linear
		return start + (end-start)*t
	}
}

// RenderVoice generates audio samples for a voice at the given duration and sample rate.
func RenderVoice(v *Voice, duration float64, sampleRate int) []float64 {
	numSamples := int(duration * float64(sampleRate))
	samples := make([]float64, numSamples)

	noteOnDur := duration
	if v.Env.Release > 0 && v.Env.Release < duration {
		noteOnDur = duration - v.Env.Release
	}

	pitchStart := v.PitchStart
	if pitchStart == 0 {
		pitchStart = v.Frequency
	}
	pitchEnd := v.PitchEnd
	if pitchEnd == 0 {
		pitchEnd = v.Frequency
	}

	var phase float64
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		progress := t / duration

		// Calculate current frequency with pitch sweep.
		freq := Interpolate(pitchStart, pitchEnd, progress, v.PitchCurve)

		// Apply vibrato.
		if v.Vibrato.Depth > 0 && v.Vibrato.Rate > 0 {
			vibrato := math.Sin(2*math.Pi*v.Vibrato.Rate*t) * v.Vibrato.Depth
			freq *= math.Pow(2, vibrato/12)
		}

		// Generate sample.
		sample := v.Osc.Sample(phase)

		// Apply filter.
		if v.Filter != nil {
			sample = v.Filter.Process(sample)
		}

		// Apply envelope.
		sample *= v.Env.Level(t, noteOnDur)

		samples[i] = sample

		// Accumulate phase.
		phase += freq / float64(sampleRate)
		phase -= math.Floor(phase) // keep in [0, 1)
	}

	return samples
}

// MIDIToFreq converts a MIDI note number to frequency in Hz.
// Note 69 = A4 = 440 Hz.
func MIDIToFreq(note int) float64 {
	return 440 * math.Pow(2, float64(note-69)/12)
}
