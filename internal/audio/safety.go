package audio

import (
	"fmt"
	"math"
)

// Warning represents a non-fatal audio processing issue.
type Warning struct {
	Message string
}

// BrickwallLimit applies a brickwall limiter at -1 dBFS with 50ms release.
func BrickwallLimit(samples []float64, sampleRate int) []float64 {
	threshold := math.Pow(10, -1.0/20.0) // -1 dBFS ≈ 0.891
	releaseSamples := sampleRate * 50 / 1000 // 50ms release

	out := make([]float64, len(samples))
	gain := 1.0
	releaseCoeff := 1.0 / float64(releaseSamples)

	for i, s := range samples {
		abs := math.Abs(s)
		if abs*gain > threshold {
			gain = threshold / abs
		} else if gain < 1.0 {
			gain += releaseCoeff
			if gain > 1.0 {
				gain = 1.0
			}
		}
		out[i] = s * gain
	}
	return out
}

// RemoveDCOffset applies a high-pass filter at 10 Hz to remove DC offset.
func RemoveDCOffset(samples []float64, sampleRate int) []float64 {
	if len(samples) == 0 {
		return samples
	}

	// Simple one-pole high-pass: y[n] = x[n] - x[n-1] + R * y[n-1]
	// R = 1 - (2*pi*fc/sr)
	fc := 10.0
	R := 1 - (2*math.Pi*fc/float64(sampleRate))

	out := make([]float64, len(samples))
	out[0] = samples[0]
	for i := 1; i < len(samples); i++ {
		out[i] = samples[i] - samples[i-1] + R*out[i-1]
	}
	return out
}

// SanitizeSamples replaces NaN and Inf values with silence.
// Returns the number of replaced samples.
func SanitizeSamples(samples []float64) int {
	replaced := 0
	for i, s := range samples {
		if math.IsNaN(s) || math.IsInf(s, 0) {
			samples[i] = 0
			replaced++
		}
	}
	return replaced
}

// PeakLevel returns the peak amplitude in dBFS.
func PeakLevel(samples []float64) float64 {
	peak := 0.0
	for _, s := range samples {
		abs := math.Abs(s)
		if abs > peak {
			peak = abs
		}
	}
	if peak == 0 {
		return -math.Inf(1)
	}
	return 20 * math.Log10(peak)
}

// ProcessSafety applies the full audio safety chain:
// DC offset removal, NaN/Inf sanitization, and brickwall limiting.
func ProcessSafety(samples []float64, sampleRate int) ([]float64, []Warning) {
	var warnings []Warning

	// Sanitize NaN/Inf.
	if n := SanitizeSamples(samples); n > 0 {
		warnings = append(warnings, Warning{
			Message: fmt.Sprintf("%d NaN/Inf samples replaced with silence", n),
		})
	}

	// Remove DC offset.
	samples = RemoveDCOffset(samples, sampleRate)

	// Apply limiter.
	samples = BrickwallLimit(samples, sampleRate)

	return samples, warnings
}
