package preview

import (
	"image"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
	// debugCharW and debugCharH are the dimensions of ebitenutil.DebugPrint's font.
	debugCharW = 6
	debugCharH = 16
	// textScale is the factor to enlarge the debug font by.
	textScale = 2.5
)

// textBuf is a reusable offscreen buffer for scaled text rendering.
var textBuf *ebiten.Image

func ensureTextBuf(w, h int) {
	if textBuf == nil || textBuf.Bounds().Dx() < w || textBuf.Bounds().Dy() < h {
		newW := max(w, 1024)
		newH := max(h, debugCharH+4)
		textBuf = ebiten.NewImage(newW, newH)
	}
}

// drawText renders text at (x, y) scaled up for readability.
func drawText(screen *ebiten.Image, str string, x, y int) {
	if str == "" {
		return
	}

	// Handle multiline.
	lines := strings.Split(str, "\n")
	maxLen := 0
	for _, l := range lines {
		if len(l) > maxLen {
			maxLen = len(l)
		}
	}

	w := maxLen*debugCharW + 2
	h := len(lines)*debugCharH + 2
	ensureTextBuf(w, h)
	textBuf.Clear()

	ebitenutil.DebugPrintAt(textBuf, str, 0, 0)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(textScale, textScale)
	op.GeoM.Translate(float64(x), float64(y))
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(textBuf.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image), op)
}

// scaledCharW returns the width of a single character at the text scale.
func scaledCharW() int {
	return int(float64(debugCharW) * textScale)
}

// scaledCharH returns the height of a single text line at the text scale.
func scaledCharH() int {
	return int(float64(debugCharH) * textScale)
}
