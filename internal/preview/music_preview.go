package preview

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/vgalaktionov/runefact/internal/track"
)

// MusicPreviewState holds music tracker preview data.
type MusicPreviewState struct {
	track      *track.Track
	currentRow int
	currentPat int
	playing    bool
	elapsed    float64
}

func (p *Previewer) initMusicState(tr *track.Track) {
	p.musicState = &MusicPreviewState{
		track: tr,
	}
}

func (p *Previewer) drawMusic(screen *ebiten.Image) {
	ms := p.musicState
	if ms == nil || ms.track == nil {
		return
	}

	tr := ms.track
	colW := 100
	rowH := 16
	headerH := 40
	offsetX := 20

	// Header.
	info := fmt.Sprintf("Track  tempo:%d  channels:%d  patterns:%d",
		tr.Tempo, len(tr.Channels), len(tr.Patterns))
	ebitenutil.DebugPrintAt(screen, info, 10, 10)

	// Channel headers.
	for i, ch := range tr.Channels {
		x := offsetX + i*colW
		ebitenutil.DebugPrintAt(screen, ch.Name, x, headerH)
	}

	// Draw divider.
	for x := offsetX; x < offsetX+len(tr.Channels)*colW; x++ {
		screen.Set(x, headerH+14, color.RGBA{R: 0x60, G: 0x60, B: 0x60, A: 0xff})
	}

	// Draw pattern data.
	if len(tr.Sequence) == 0 {
		ebitenutil.DebugPrintAt(screen, "No patterns in sequence", 10, headerH+30)
		return
	}

	patName := tr.Sequence[ms.currentPat%len(tr.Sequence)]
	pat := tr.Patterns[patName]
	if pat == nil {
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pattern %q not found", patName), 10, headerH+30)
		return
	}

	// Position info.
	posLabel := fmt.Sprintf("Pattern: %s  Row: %02d/%02d", patName, ms.currentRow, pat.Ticks)
	ebitenutil.DebugPrintAt(screen, posLabel, 10, 26)

	startY := headerH + 20
	visibleRows := (p.winH - startY) / rowH

	// Center current row.
	scrollOffset := ms.currentRow - visibleRows/2
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	for rowIdx := 0; rowIdx < visibleRows && scrollOffset+rowIdx < len(pat.Rows); rowIdx++ {
		absRow := scrollOffset + rowIdx
		y := startY + rowIdx*rowH

		// Highlight current row.
		if absRow == ms.currentRow {
			for x := offsetX - 2; x < offsetX+len(tr.Channels)*colW+2; x++ {
				screen.Set(x, y, color.RGBA{R: 0x33, G: 0x44, B: 0x66, A: 0xff})
				screen.Set(x, y+rowH-1, color.RGBA{R: 0x33, G: 0x44, B: 0x66, A: 0xff})
			}
		}

		// Row number.
		ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%02X", absRow), offsetX-16, y)

		row := pat.Rows[absRow]
		for ch, note := range row {
			x := offsetX + ch*colW
			text := formatNote(note)
			ebitenutil.DebugPrintAt(screen, text, x, y)
		}
	}

	ebitenutil.DebugPrintAt(screen, "Press Enter to play (audio not connected)", 10, p.winH-16)
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
