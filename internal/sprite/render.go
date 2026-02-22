package sprite

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// SpriteInfo holds metadata about a sprite's position in the sheet.
type SpriteInfo struct {
	Sheet  string `json:"sheet"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	H      int    `json:"h"`
	Frames int    `json:"frames"`
	FPS    int    `json:"fps"`
}

// SpriteSheetMeta contains metadata for all sprites in a sheet.
type SpriteSheetMeta struct {
	SheetPath string
	Sprites   map[string]SpriteInfo
}

// RenderSpriteSheet renders resolved sprites into a sprite sheet image.
// Layout: frames horizontal per sprite, sprites stacked vertically.
func RenderSpriteSheet(sprites []ResolvedSprite) (*image.RGBA, SpriteSheetMeta, error) {
	if len(sprites) == 0 {
		return nil, SpriteSheetMeta{}, fmt.Errorf("no sprites to render")
	}

	// Calculate sheet dimensions.
	maxWidth := 0
	totalHeight := 0
	for _, s := range sprites {
		frameWidth := s.Grid.W * len(s.Frames)
		if frameWidth > maxWidth {
			maxWidth = frameWidth
		}
		totalHeight += s.Grid.H
	}

	img := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	meta := SpriteSheetMeta{Sprites: make(map[string]SpriteInfo)}

	y := 0
	for _, s := range sprites {
		for frameIdx, frame := range s.Frames {
			xOff := frameIdx * s.Grid.W
			for py, row := range frame.Pixels {
				for px, c := range row {
					img.Set(xOff+px, y+py, c.ToRGBA())
				}
			}
		}
		meta.Sprites[s.Name] = SpriteInfo{
			X:      0,
			Y:      y,
			W:      s.Grid.W,
			H:      s.Grid.H,
			Frames: len(s.Frames),
			FPS:    s.Framerate,
		}
		y += s.Grid.H
	}

	return img, meta, nil
}

// WritePNG encodes an image as PNG and writes it to path, creating directories as needed.
func WritePNG(img image.Image, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("creating PNG file: %w", err)
	}
	defer f.Close()

	if err := png.Encode(f, img); err != nil {
		return fmt.Errorf("encoding PNG: %w", err)
	}
	return nil
}
