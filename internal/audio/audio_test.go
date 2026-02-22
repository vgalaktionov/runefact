package audio

import (
	"math"
	"testing"
)

func TestSineOsc(t *testing.T) {
	osc := SineOsc{}
	// Phase 0 -> 0, phase 0.25 -> 1, phase 0.5 -> 0, phase 0.75 -> -1
	if v := osc.Sample(0); math.Abs(v) > 1e-10 {
		t.Errorf("sine(0) = %f, want ~0", v)
	}
	if v := osc.Sample(0.25); math.Abs(v-1) > 1e-10 {
		t.Errorf("sine(0.25) = %f, want ~1", v)
	}
	if v := osc.Sample(0.75); math.Abs(v+1) > 1e-10 {
		t.Errorf("sine(0.75) = %f, want ~-1", v)
	}
}

func TestSquareOsc(t *testing.T) {
	osc := SquareOsc{}
	if v := osc.Sample(0.25); v != 1 {
		t.Errorf("square(0.25) = %f, want 1", v)
	}
	if v := osc.Sample(0.75); v != -1 {
		t.Errorf("square(0.75) = %f, want -1", v)
	}
}

func TestTriangleOsc(t *testing.T) {
	osc := TriangleOsc{}
	// 4*|0-0.5|-1 = 4*0.5-1 = 1 at phase 0
	if v := osc.Sample(0.0); math.Abs(v-1) > 1e-10 {
		t.Errorf("triangle(0) = %f, want 1", v)
	}
	// 4*|0.25-0.5|-1 = 4*0.25-1 = 0 at phase 0.25
	if v := osc.Sample(0.25); math.Abs(v) > 1e-10 {
		t.Errorf("triangle(0.25) = %f, want 0", v)
	}
	// 4*|0.5-0.5|-1 = -1 at phase 0.5
	if v := osc.Sample(0.5); math.Abs(v+1) > 1e-10 {
		t.Errorf("triangle(0.5) = %f, want -1", v)
	}
}

func TestSawtoothOsc(t *testing.T) {
	osc := SawtoothOsc{}
	if v := osc.Sample(0); math.Abs(v+1) > 1e-10 {
		t.Errorf("saw(0) = %f, want -1", v)
	}
	if v := osc.Sample(0.5); math.Abs(v) > 1e-10 {
		t.Errorf("saw(0.5) = %f, want 0", v)
	}
}

func TestNoiseOsc(t *testing.T) {
	osc := NoiseOsc{}
	// Just check it's in range [-1, 1].
	for i := 0; i < 100; i++ {
		v := osc.Sample(0)
		if v < -1 || v > 1 {
			t.Errorf("noise sample %f out of range", v)
		}
	}
}

func TestPulseOsc(t *testing.T) {
	osc := PulseOsc{DutyCycle: 0.25}
	if v := osc.Sample(0.1); v != 1 {
		t.Errorf("pulse(0.1, dc=0.25) = %f, want 1", v)
	}
	if v := osc.Sample(0.5); v != -1 {
		t.Errorf("pulse(0.5, dc=0.25) = %f, want -1", v)
	}
}

func TestNewOscillator(t *testing.T) {
	for _, name := range []string{"sine", "square", "triangle", "sawtooth", "noise", "pulse"} {
		osc := NewOscillator(name, 0.5)
		if osc == nil {
			t.Errorf("NewOscillator(%q) returned nil", name)
		}
	}
	// Unknown defaults to sine.
	osc := NewOscillator("unknown", 0)
	if _, ok := osc.(SineOsc); !ok {
		t.Error("unknown waveform should default to sine")
	}
}

func TestADSR_Attack(t *testing.T) {
	env := ADSR{Attack: 0.1, Decay: 0, Sustain: 1, Release: 0}
	if v := env.Level(0.05, 1.0); math.Abs(v-0.5) > 1e-10 {
		t.Errorf("mid-attack = %f, want 0.5", v)
	}
}

func TestADSR_Decay(t *testing.T) {
	env := ADSR{Attack: 0, Decay: 0.1, Sustain: 0.5, Release: 0}
	if v := env.Level(0.05, 1.0); math.Abs(v-0.75) > 1e-10 {
		t.Errorf("mid-decay = %f, want 0.75", v)
	}
}

func TestADSR_Sustain(t *testing.T) {
	env := ADSR{Attack: 0.01, Decay: 0.01, Sustain: 0.6, Release: 0.1}
	if v := env.Level(0.5, 1.0); math.Abs(v-0.6) > 1e-10 {
		t.Errorf("sustain = %f, want 0.6", v)
	}
}

func TestADSR_Release(t *testing.T) {
	env := ADSR{Attack: 0, Decay: 0, Sustain: 0.8, Release: 0.1}
	// At noteOnDuration, release starts from sustain level.
	v := env.Level(0.55, 0.5)
	if v < 0 || v > 0.8 {
		t.Errorf("release = %f, expected between 0 and 0.8", v)
	}
	// After full release, should be 0.
	v = env.Level(0.7, 0.5)
	if math.Abs(v) > 1e-10 {
		t.Errorf("post-release = %f, want 0", v)
	}
}

func TestInterpolate_Linear(t *testing.T) {
	v := Interpolate(100, 200, 0.5, CurveLinear)
	if math.Abs(v-150) > 1e-10 {
		t.Errorf("linear(0.5) = %f, want 150", v)
	}
}

func TestInterpolate_Exponential(t *testing.T) {
	v := Interpolate(100, 400, 0.5, CurveExponential)
	if math.Abs(v-200) > 1 {
		t.Errorf("exponential(0.5) = %f, want ~200", v)
	}
}

func TestInterpolate_Boundaries(t *testing.T) {
	if v := Interpolate(100, 200, 0, CurveLinear); v != 100 {
		t.Errorf("t=0: %f, want 100", v)
	}
	if v := Interpolate(100, 200, 1, CurveLinear); v != 200 {
		t.Errorf("t=1: %f, want 200", v)
	}
}

func TestMIDIToFreq(t *testing.T) {
	// A4 = 440 Hz
	f := MIDIToFreq(69)
	if math.Abs(f-440) > 0.01 {
		t.Errorf("MIDI 69 = %f Hz, want 440", f)
	}
	// C4 = ~261.63 Hz
	f = MIDIToFreq(60)
	if math.Abs(f-261.63) > 0.1 {
		t.Errorf("MIDI 60 = %f Hz, want ~261.63", f)
	}
}

func TestRenderVoice_ProducesSamples(t *testing.T) {
	v := &Voice{
		Osc:       SineOsc{},
		Env:       ADSR{Attack: 0.01, Decay: 0.01, Sustain: 0.8, Release: 0.05},
		Frequency: 440,
	}
	samples := RenderVoice(v, 0.1, 44100)
	if len(samples) != 4410 {
		t.Errorf("got %d samples, want 4410", len(samples))
	}
	// Check samples are in valid range.
	for i, s := range samples {
		if s < -1.1 || s > 1.1 {
			t.Errorf("sample %d = %f, out of range", i, s)
			break
		}
	}
}

func TestRenderVoice_PitchSweep(t *testing.T) {
	v := &Voice{
		Osc:        SineOsc{},
		Env:        ADSR{Attack: 0, Decay: 0, Sustain: 1, Release: 0},
		PitchStart: 200,
		PitchEnd:   400,
		PitchCurve: CurveLinear,
	}
	samples := RenderVoice(v, 0.1, 44100)
	if len(samples) == 0 {
		t.Fatal("no samples rendered")
	}
}

func TestBiquadFilter_Lowpass(t *testing.T) {
	f := NewBiquadFilter(FilterLowpass, 1000, 0.3, 44100)
	// Process some samples — just check it doesn't panic or produce NaN.
	for i := 0; i < 100; i++ {
		phase := float64(i) / 100
		sample := math.Sin(2 * math.Pi * phase * 10000 / 44100)
		out := f.Process(sample)
		if math.IsNaN(out) || math.IsInf(out, 0) {
			t.Fatalf("filter produced NaN/Inf at sample %d", i)
		}
	}
}

func TestBiquadFilter_SetCutoff(t *testing.T) {
	f := NewBiquadFilter(FilterLowpass, 1000, 0, 44100)
	f.SetCutoff(500)
	out := f.Process(1.0)
	if math.IsNaN(out) {
		t.Fatal("NaN after SetCutoff")
	}
}

func TestBiquadFilter_Reset(t *testing.T) {
	f := NewBiquadFilter(FilterLowpass, 1000, 0, 44100)
	f.Process(1.0)
	f.Reset()
	if f.x1 != 0 || f.y1 != 0 {
		t.Error("state not cleared after Reset")
	}
}
