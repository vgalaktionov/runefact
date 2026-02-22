package preview

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/vgalaktionov/runefact/internal/sfx"
)

// SFXPreviewState holds SFX preview data.
type SFXPreviewState struct {
	sfxDef     *sfx.SFX
	waveform   []float64 // downsampled waveform for display
	sampleRate int
}

func (p *Previewer) initSFXState(s *sfx.SFX, sampleRate int) {
	samples, _ := s.Render(sampleRate)

	// Downsample for display.
	displayWidth := 700
	waveform := downsampleWaveform(samples, displayWidth)

	p.sfxState = &SFXPreviewState{
		sfxDef:     s,
		waveform:   waveform,
		sampleRate: sampleRate,
	}
}

func downsampleWaveform(samples []float64, width int) []float64 {
	if len(samples) == 0 || width <= 0 {
		return nil
	}
	result := make([]float64, width)
	samplesPerPixel := float64(len(samples)) / float64(width)
	for i := range result {
		start := int(float64(i) * samplesPerPixel)
		end := int(float64(i+1) * samplesPerPixel)
		if end > len(samples) {
			end = len(samples)
		}
		maxAbs := 0.0
		for j := start; j < end; j++ {
			abs := math.Abs(samples[j])
			if abs > maxAbs {
				maxAbs = abs
			}
		}
		// Keep sign of the sample at midpoint.
		mid := (start + end) / 2
		if mid < len(samples) && samples[mid] < 0 {
			maxAbs = -maxAbs
		}
		result[i] = maxAbs
	}
	return result
}

func (p *Previewer) drawSFX(screen *ebiten.Image) {
	ss := p.sfxState
	if ss == nil {
		return
	}

	// Waveform area: top 60%.
	waveH := int(float64(p.winH) * 0.6)
	midY := waveH / 2
	offsetX := 50

	// Draw zero line.
	for x := offsetX; x < p.winW-20; x++ {
		screen.Set(x, midY, color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff})
	}

	// Draw waveform.
	waveColor := color.RGBA{R: 0x00, G: 0xcc, B: 0xcc, A: 0xff}
	drawWidth := p.winW - offsetX - 20
	for i, v := range ss.waveform {
		if i >= drawWidth {
			break
		}
		x := offsetX + i
		h := int(v * float64(midY-10))
		if h > 0 {
			for dy := 0; dy < h; dy++ {
				screen.Set(x, midY-dy, waveColor)
			}
		} else {
			for dy := 0; dy > h; dy-- {
				screen.Set(x, midY-dy, waveColor)
			}
		}
	}

	// Envelope graph: bottom 40%, left half.
	graphY := waveH + 10
	graphH := p.winH - graphY - 30
	envColor := color.RGBA{R: 0x00, G: 0xcc, B: 0x00, A: 0xff}

	if len(ss.sfxDef.Voices) > 0 {
		v := ss.sfxDef.Voices[0]
		duration := ss.sfxDef.Duration
		halfW := (p.winW - offsetX - 20) / 2

		// ADSR envelope shape.
		for x := 0; x < halfW; x++ {
			t := float64(x) / float64(halfW) * duration
			level := adsrLevel(v.Envelope, duration, t)
			h := int(level * float64(graphH))
			py := graphY + graphH - h
			screen.Set(offsetX+x, py, envColor)
		}
		ebitenutil.DebugPrintAt(screen, "Envelope", offsetX, graphY-12)

		// Pitch curve: bottom 40%, right half.
		pitchColor := color.RGBA{R: 0xff, G: 0x99, B: 0x00, A: 0xff}
		pitchX := offsetX + halfW + 20
		if v.Pitch.Start > 0 || v.Pitch.End > 0 {
			maxFreq := math.Max(v.Pitch.Start, v.Pitch.End)
			if maxFreq == 0 {
				maxFreq = 440
			}
			for x := 0; x < halfW-20; x++ {
				t := float64(x) / float64(halfW-20)
				freq := v.Pitch.Start + (v.Pitch.End-v.Pitch.Start)*t
				h := int((freq / maxFreq) * float64(graphH))
				py := graphY + graphH - h
				screen.Set(pitchX+x, py, pitchColor)
			}
			ebitenutil.DebugPrintAt(screen, "Pitch", pitchX, graphY-12)
		}
	}

	// Info.
	info := fmt.Sprintf("SFX  dur:%.2fs  voices:%d  rate:%dHz",
		ss.sfxDef.Duration, len(ss.sfxDef.Voices), ss.sampleRate)
	ebitenutil.DebugPrintAt(screen, info, 10, 10)
	ebitenutil.DebugPrintAt(screen, "Press Enter to play (not implemented in headless)", 10, p.winH-16)
}

// adsrLevel computes ADSR amplitude at time t.
func adsrLevel(env sfx.EnvelopeDef, duration, t float64) float64 {
	a, d, s, r := env.Attack, env.Decay, env.Sustain, env.Release
	noteOff := duration - r

	if t < a {
		if a == 0 {
			return 1.0
		}
		return t / a
	}
	if t < a+d {
		return 1.0 - (1.0-s)*((t-a)/d)
	}
	if t < noteOff {
		return s
	}
	if r == 0 {
		return 0
	}
	releaseT := (t - noteOff) / r
	if releaseT > 1 {
		return 0
	}
	return s * (1 - releaseT)
}
