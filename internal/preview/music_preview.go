package preview

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image/color"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/vgalaktionov/runefact/internal/instrument"
	"github.com/vgalaktionov/runefact/internal/track"
)

// MusicPreviewState holds music tracker preview data.
type MusicPreviewState struct {
	track      *track.Track
	currentRow int
	currentPat int
	playing    bool
	elapsed    float64

	// Audio playback.
	samples    []float64
	sampleRate int
	audioCtx   *audio.Context
	player     *audio.Player
	audioErr   string

	// Playback position tracking.
	samplesPerTick float64
	totalTicks     int
}

func (p *Previewer) initMusicState(tr *track.Track) {
	sr := p.sampleRate
	if sr == 0 {
		sr = 44100
	}

	// Load instruments from assets dir.
	instruments := loadInstruments(p.assetsDir)

	// Render to samples.
	samples, _ := tr.Render(instruments, sr)

	// Compute timing for playback cursor.
	samplesPerTick := float64(sr) * 60.0 / float64(tr.Tempo) / float64(tr.TicksPerBeat)
	totalTicks := 0
	for _, pname := range tr.Sequence {
		if pat := tr.Patterns[pname]; pat != nil {
			totalTicks += len(pat.Rows)
		}
	}

	p.musicState = &MusicPreviewState{
		track:          tr,
		samples:        samples,
		sampleRate:     sr,
		samplesPerTick: samplesPerTick,
		totalTicks:     totalTicks,
	}
}

func loadInstruments(assetsDir string) map[string]*instrument.Instrument {
	instruments := map[string]*instrument.Instrument{}
	instDir := filepath.Join(assetsDir, "instruments")
	entries, err := os.ReadDir(instDir)
	if err != nil {
		return instruments
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".inst") {
			continue
		}
		inst, err := instrument.LoadInstrument(filepath.Join(instDir, e.Name()))
		if err != nil {
			continue
		}
		instruments[inst.Name] = inst
	}
	return instruments
}

func (p *Previewer) updateMusic() {
	if p.musicState == nil {
		return
	}
	ms := p.musicState

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if ms.playing {
			// Stop.
			if ms.player != nil {
				ms.player.Pause()
			}
			ms.playing = false
			ms.currentRow = 0
			ms.currentPat = 0
			ms.elapsed = 0
		} else {
			ms.play()
		}
	}

	// Advance playback cursor.
	if ms.playing && ms.player != nil && ms.player.IsPlaying() {
		ms.elapsed += 1.0 / float64(ebiten.TPS())
		tickIdx := int(ms.elapsed * float64(ms.sampleRate) / ms.samplesPerTick)

		// Map tick index to pattern + row.
		remaining := tickIdx
		ms.currentPat = 0
		ms.currentRow = 0
		for i, pname := range ms.track.Sequence {
			pat := ms.track.Patterns[pname]
			if pat == nil {
				continue
			}
			if remaining < len(pat.Rows) {
				ms.currentPat = i
				ms.currentRow = remaining
				break
			}
			remaining -= len(pat.Rows)
		}
	} else if ms.playing {
		// Playback finished.
		ms.playing = false
	}
}

func (ms *MusicPreviewState) ensureAudio() {
	if ms.audioCtx != nil || ms.audioErr != "" {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			ms.audioErr = fmt.Sprintf("%v", r)
		}
	}()
	ms.audioCtx = audio.NewContext(ms.sampleRate)
}

func (ms *MusicPreviewState) play() {
	ms.ensureAudio()
	if ms.audioErr != "" || ms.audioCtx == nil {
		return
	}

	// Convert float64 samples to 16-bit stereo PCM.
	buf := &bytes.Buffer{}
	for _, s := range ms.samples {
		if s > 1.0 {
			s = 1.0
		} else if s < -1.0 {
			s = -1.0
		}
		v := int16(s * 32767)
		binary.Write(buf, binary.LittleEndian, v) // left
		binary.Write(buf, binary.LittleEndian, v) // right
	}

	defer func() {
		if r := recover(); r != nil {
			ms.audioErr = fmt.Sprintf("%v", r)
		}
	}()
	player := ms.audioCtx.NewPlayerFromBytes(buf.Bytes())
	player.Play()
	ms.player = player
	ms.playing = true
	ms.elapsed = 0
	ms.currentRow = 0
	ms.currentPat = 0
}

func (p *Previewer) drawMusic(screen *ebiten.Image) {
	ms := p.musicState
	if ms == nil || ms.track == nil {
		return
	}

	tr := ms.track
	lineH := scaledCharH()
	charW := scaledCharW()
	colW := charW * 8 // 8 chars per column (e.g. "C#4  ---")
	rowH := lineH + 4
	headerH := lineH*2 + 10
	offsetX := charW * 4

	// Header.
	dur := 0.0
	if ms.samplesPerTick > 0 {
		dur = float64(ms.totalTicks) * ms.samplesPerTick / float64(ms.sampleRate)
	}
	info := fmt.Sprintf("Track  tempo:%d  channels:%d  patterns:%d  dur:%.1fs",
		tr.Tempo, len(tr.Channels), len(tr.Patterns), dur)
	drawText(screen, info, 10, 10)

	// Channel headers.
	for i, ch := range tr.Channels {
		x := offsetX + i*colW
		drawText(screen, ch.Name, x, headerH)
	}

	// Draw divider.
	divY := headerH + lineH + 4
	for x := offsetX; x < offsetX+len(tr.Channels)*colW; x++ {
		screen.Set(x, divY, color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff})
	}

	// Draw pattern data.
	if len(tr.Sequence) == 0 {
		drawText(screen, "No patterns in sequence", 10, divY+10)
		return
	}

	patIdx := ms.currentPat % len(tr.Sequence)
	patName := tr.Sequence[patIdx]
	pat := tr.Patterns[patName]
	if pat == nil {
		drawText(screen, fmt.Sprintf("Pattern %q not found", patName), 10, divY+10)
		return
	}

	// Position info.
	posLabel := fmt.Sprintf("Pattern: %s (%d/%d)  Row: %02d/%02d",
		patName, patIdx+1, len(tr.Sequence), ms.currentRow, len(pat.Rows))
	drawText(screen, posLabel, 10, lineH+14)

	startY := divY + 6
	waveArea := 80 // reserve space for waveform + status
	visibleRows := (p.winH - startY - waveArea) / rowH

	// Center current row.
	scrollOffset := ms.currentRow - visibleRows/2
	if scrollOffset < 0 {
		scrollOffset = 0
	}
	maxScroll := len(pat.Rows) - visibleRows
	if maxScroll > 0 && scrollOffset > maxScroll {
		scrollOffset = maxScroll
	}

	for rowIdx := 0; rowIdx < visibleRows && scrollOffset+rowIdx < len(pat.Rows); rowIdx++ {
		absRow := scrollOffset + rowIdx
		y := startY + rowIdx*rowH

		// Highlight current row.
		if absRow == ms.currentRow && ms.playing {
			for hx := offsetX - 2; hx < offsetX+len(tr.Channels)*colW+2; hx++ {
				for hy := 0; hy < rowH; hy++ {
					screen.Set(hx, y+hy, color.RGBA{R: 0x22, G: 0x33, B: 0x55, A: 0xff})
				}
			}
		}

		// Row number.
		drawText(screen, fmt.Sprintf("%02X", absRow), 4, y)

		row := pat.Rows[absRow]
		for ch, note := range row {
			x := offsetX + ch*colW
			text := formatNote(note)
			drawText(screen, text, x, y)
		}
	}

	// Waveform mini-display at bottom.
	waveY := p.winH - 70
	waveH := 30
	waveColor := color.RGBA{R: 0x00, G: 0xcc, B: 0xcc, A: 0xff}
	drawW := p.winW - offsetX*2
	if len(ms.samples) > 0 && drawW > 0 {
		samplesPerPx := float64(len(ms.samples)) / float64(drawW)
		for x := 0; x < drawW; x++ {
			start := int(float64(x) * samplesPerPx)
			end := int(float64(x+1) * samplesPerPx)
			if end > len(ms.samples) {
				end = len(ms.samples)
			}
			maxAbs := 0.0
			for j := start; j < end; j++ {
				a := math.Abs(ms.samples[j])
				if a > maxAbs {
					maxAbs = a
				}
			}
			h := int(maxAbs * float64(waveH/2))
			mid := waveY + waveH/2
			for dy := -h; dy <= h; dy++ {
				screen.Set(offsetX+x, mid+dy, waveColor)
			}
		}

		// Playback position indicator.
		if ms.playing && len(ms.samples) > 0 {
			pos := ms.elapsed * float64(ms.sampleRate)
			px := int(pos / float64(len(ms.samples)) * float64(drawW))
			if px >= 0 && px < drawW {
				for dy := 0; dy < waveH; dy++ {
					screen.Set(offsetX+px, waveY+dy, color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff})
				}
			}
		}
	}

	// Status line.
	statusY := p.winH - lineH - 6
	if ms.audioErr != "" {
		drawText(screen, "No audio device available", 10, statusY)
	} else if ms.playing {
		drawText(screen, "Playing - Enter to stop", 10, statusY)
	} else {
		drawText(screen, "Press Enter to play", 10, statusY)
	}
}

func formatNote(n track.Note) string {
	switch n.Type {
	case track.NoteOn:
		return fmt.Sprintf("%s%d", n.Name, n.Octave)
	case track.Sustain:
		return "---"
	case track.Silence:
		return "..."
	case track.NoteOff:
		return "^^^"
	default:
		return "???"
	}
}
