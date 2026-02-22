package sprite

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/vgalaktionov/runefact/internal/palette"
)

func TestRenderSpriteSheet_SingleStatic(t *testing.T) {
	red := palette.Color{R: 255, A: 255}
	trans := palette.Color{A: 0}

	sprites := []ResolvedSprite{
		{
			Name: "dot",
			Grid: Grid{W: 2, H: 2},
			Frames: []ResolvedFrame{
				{Pixels: [][]palette.Color{
					{red, trans},
					{trans, red},
				}},
			},
		},
	}

	img, meta, err := RenderSpriteSheet(sprites)
	if err != nil {
		t.Fatal(err)
	}

	// Check dimensions.
	bounds := img.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Errorf("sheet size = %dx%d, want 2x2", bounds.Dx(), bounds.Dy())
	}

	// Check pixel values.
	r, _, _, a := img.At(0, 0).RGBA()
	if r>>8 != 255 || a>>8 != 255 {
		t.Errorf("pixel 0,0: want red, got r=%d a=%d", r>>8, a>>8)
	}
	_, _, _, a = img.At(1, 0).RGBA()
	if a != 0 {
		t.Errorf("pixel 1,0: want transparent, got a=%d", a)
	}

	// Check metadata.
	info, ok := meta.Sprites["dot"]
	if !ok {
		t.Fatal("missing metadata for 'dot'")
	}
	if info.X != 0 || info.Y != 0 || info.W != 2 || info.H != 2 || info.Frames != 1 {
		t.Errorf("meta = %+v", info)
	}
}

func TestRenderSpriteSheet_Animated(t *testing.T) {
	red := palette.Color{R: 255, A: 255}
	blue := palette.Color{B: 255, A: 255}

	sprites := []ResolvedSprite{
		{
			Name:      "blink",
			Grid:      Grid{W: 2, H: 2},
			Framerate: 4,
			Frames: []ResolvedFrame{
				{Pixels: [][]palette.Color{{red, red}, {red, red}}},
				{Pixels: [][]palette.Color{{blue, blue}, {blue, blue}}},
			},
		},
	}

	img, meta, err := RenderSpriteSheet(sprites)
	if err != nil {
		t.Fatal(err)
	}

	// Width should be 2 frames * 2 pixels = 4.
	bounds := img.Bounds()
	if bounds.Dx() != 4 || bounds.Dy() != 2 {
		t.Errorf("sheet size = %dx%d, want 4x2", bounds.Dx(), bounds.Dy())
	}

	// Frame 1 at x=0, frame 2 at x=2.
	r, _, _, _ := img.At(0, 0).RGBA()
	if r>>8 != 255 {
		t.Error("frame 1 should be red")
	}
	_, _, b, _ := img.At(2, 0).RGBA()
	if b>>8 != 255 {
		t.Error("frame 2 should be blue")
	}

	info := meta.Sprites["blink"]
	if info.Frames != 2 || info.FPS != 4 {
		t.Errorf("meta = %+v", info)
	}
}

func TestRenderSpriteSheet_MultipleSprites(t *testing.T) {
	red := palette.Color{R: 255, A: 255}
	green := palette.Color{G: 255, A: 255}

	sprites := []ResolvedSprite{
		{
			Name: "top",
			Grid: Grid{W: 2, H: 2},
			Frames: []ResolvedFrame{
				{Pixels: [][]palette.Color{{red, red}, {red, red}}},
			},
		},
		{
			Name: "bottom",
			Grid: Grid{W: 3, H: 1},
			Frames: []ResolvedFrame{
				{Pixels: [][]palette.Color{{green, green, green}}},
			},
		},
	}

	img, meta, err := RenderSpriteSheet(sprites)
	if err != nil {
		t.Fatal(err)
	}

	// Width = max(2, 3) = 3, Height = 2 + 1 = 3.
	bounds := img.Bounds()
	if bounds.Dx() != 3 || bounds.Dy() != 3 {
		t.Errorf("sheet size = %dx%d, want 3x3", bounds.Dx(), bounds.Dy())
	}

	// top at y=0, bottom at y=2.
	if meta.Sprites["top"].Y != 0 {
		t.Error("top should be at y=0")
	}
	if meta.Sprites["bottom"].Y != 2 {
		t.Error("bottom should be at y=2")
	}
}

func TestRenderSpriteSheet_Empty(t *testing.T) {
	_, _, err := RenderSpriteSheet(nil)
	if err == nil {
		t.Fatal("expected error for empty sprites")
	}
}

func TestWritePNG(t *testing.T) {
	red := palette.Color{R: 255, A: 255}
	sprites := []ResolvedSprite{
		{
			Name: "test",
			Grid: Grid{W: 1, H: 1},
			Frames: []ResolvedFrame{
				{Pixels: [][]palette.Color{{red}}},
			},
		},
	}

	img, _, err := RenderSpriteSheet(sprites)
	if err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "sprites", "test.png")
	if err := WritePNG(img, path); err != nil {
		t.Fatal(err)
	}

	// Verify file exists and is valid PNG.
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	decoded, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	if decoded.Bounds().Dx() != 1 || decoded.Bounds().Dy() != 1 {
		t.Errorf("decoded size = %v", decoded.Bounds())
	}
}
