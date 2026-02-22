package audio

import "math"

// FilterType defines the type of biquad filter.
type FilterType string

const (
	FilterLowpass  FilterType = "lowpass"
	FilterHighpass FilterType = "highpass"
	FilterBandpass FilterType = "bandpass"
)

// BiquadFilter implements a biquad digital filter.
type BiquadFilter struct {
	filterType FilterType
	cutoff     float64
	resonance  float64
	sampleRate float64

	// Coefficients.
	b0, b1, b2 float64
	a1, a2     float64

	// State.
	x1, x2 float64
	y1, y2 float64
}

// NewBiquadFilter creates a new biquad filter.
func NewBiquadFilter(filterType FilterType, cutoff, resonance float64, sampleRate int) *BiquadFilter {
	f := &BiquadFilter{
		filterType: filterType,
		cutoff:     cutoff,
		resonance:  resonance,
		sampleRate: float64(sampleRate),
	}
	f.recalculate()
	return f
}

// SetCutoff updates the filter cutoff frequency.
func (f *BiquadFilter) SetCutoff(cutoff float64) {
	f.cutoff = cutoff
	f.recalculate()
}

func (f *BiquadFilter) recalculate() {
	w0 := 2 * math.Pi * f.cutoff / f.sampleRate
	sinW0 := math.Sin(w0)
	cosW0 := math.Cos(w0)
	q := 0.707 + f.resonance*10 // map 0-1 resonance to Q range

	alpha := sinW0 / (2 * q)

	var b0, b1, b2, a0, a1, a2 float64

	switch f.filterType {
	case FilterLowpass:
		b0 = (1 - cosW0) / 2
		b1 = 1 - cosW0
		b2 = (1 - cosW0) / 2
		a0 = 1 + alpha
		a1 = -2 * cosW0
		a2 = 1 - alpha
	case FilterHighpass:
		b0 = (1 + cosW0) / 2
		b1 = -(1 + cosW0)
		b2 = (1 + cosW0) / 2
		a0 = 1 + alpha
		a1 = -2 * cosW0
		a2 = 1 - alpha
	case FilterBandpass:
		b0 = alpha
		b1 = 0
		b2 = -alpha
		a0 = 1 + alpha
		a1 = -2 * cosW0
		a2 = 1 - alpha
	}

	// Normalize.
	f.b0 = b0 / a0
	f.b1 = b1 / a0
	f.b2 = b2 / a0
	f.a1 = a1 / a0
	f.a2 = a2 / a0
}

// Process runs one sample through the filter.
func (f *BiquadFilter) Process(x float64) float64 {
	y := f.b0*x + f.b1*f.x1 + f.b2*f.x2 - f.a1*f.y1 - f.a2*f.y2

	f.x2 = f.x1
	f.x1 = x
	f.y2 = f.y1
	f.y1 = y

	return y
}

// Reset clears the filter state.
func (f *BiquadFilter) Reset() {
	f.x1, f.x2 = 0, 0
	f.y1, f.y2 = 0, 0
}
