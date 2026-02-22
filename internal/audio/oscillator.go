package audio

import (
	"math"
	"math/rand"
)

// Oscillator generates a sample value for a given phase [0, 1).
type Oscillator interface {
	Sample(phase float64) float64
}

// SineOsc generates a sine waveform.
type SineOsc struct{}

func (SineOsc) Sample(phase float64) float64 {
	return math.Sin(2 * math.Pi * phase)
}

// SquareOsc generates a square waveform.
type SquareOsc struct{}

func (SquareOsc) Sample(phase float64) float64 {
	if phase < 0.5 {
		return 1
	}
	return -1
}

// TriangleOsc generates a triangle waveform.
type TriangleOsc struct{}

func (TriangleOsc) Sample(phase float64) float64 {
	return 4*math.Abs(phase-0.5) - 1
}

// SawtoothOsc generates a sawtooth waveform.
type SawtoothOsc struct{}

func (SawtoothOsc) Sample(phase float64) float64 {
	return 2*phase - 1
}

// NoiseOsc generates white noise (ignores phase).
type NoiseOsc struct{}

func (NoiseOsc) Sample(_ float64) float64 {
	return rand.Float64()*2 - 1
}

// PulseOsc generates a pulse waveform with adjustable duty cycle.
type PulseOsc struct {
	DutyCycle float64 // 0.0 to 1.0, default 0.5
}

func (p PulseOsc) Sample(phase float64) float64 {
	dc := p.DutyCycle
	if dc <= 0 || dc >= 1 {
		dc = 0.5
	}
	if phase < dc {
		return 1
	}
	return -1
}

// NewOscillator creates an oscillator by waveform name.
func NewOscillator(waveform string, dutyCycle float64) Oscillator {
	switch waveform {
	case "sine":
		return SineOsc{}
	case "square":
		return SquareOsc{}
	case "triangle":
		return TriangleOsc{}
	case "sawtooth":
		return SawtoothOsc{}
	case "noise":
		return NoiseOsc{}
	case "pulse":
		return PulseOsc{DutyCycle: dutyCycle}
	default:
		return SineOsc{}
	}
}
