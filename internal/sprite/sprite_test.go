package sprite

import (
	"strings"
	"testing"

	"github.com/vgalaktionov/runefact/internal/palette"
)

func TestParsePixelGrid_Simple(t *testing.T) {
	grid, err := ParsePixelGrid("ab\ncd")
	if err != nil {
		t.Fatal(err)
	}
	if len(grid) != 2 || len(grid[0]) != 2 {
		t.Fatalf("got %dx%d, want 2x2", len(grid[0]), len(grid))
	}
	if grid[0][0] != "a" || grid[0][1] != "b" || grid[1][0] != "c" || grid[1][1] != "d" {
		t.Errorf("unexpected values: %v", grid)
	}
}

func TestParsePixelGrid_BracketKeys(t *testing.T) {
	grid, err := ParsePixelGrid("a[sk]b\nc[rb]d")
	if err != nil {
		t.Fatal(err)
	}
	if len(grid[0]) != 3 {
		t.Fatalf("row width = %d, want 3", len(grid[0]))
	}
	if grid[0][1] != "sk" {
		t.Errorf("grid[0][1] = %q, want %q", grid[0][1], "sk")
	}
	if grid[1][1] != "rb" {
		t.Errorf("grid[1][1] = %q, want %q", grid[1][1], "rb")
	}
}

func TestParsePixelGrid_Ragged(t *testing.T) {
	_, err := ParsePixelGrid("abc\nde")
	if err == nil {
		t.Fatal("expected ragged row error")
	}
	if !strings.Contains(err.Error(), "ragged") {
		t.Errorf("error should mention ragged: %v", err)
	}
}

func TestParsePixelGrid_UnclosedBracket(t *testing.T) {
	_, err := ParsePixelGrid("a[skb")
	if err == nil {
		t.Fatal("expected unclosed bracket error")
	}
}

func TestParseSpriteFile_Static(t *testing.T) {
	input := []byte(`
palette = "default"
grid = 4

[sprite.test]
pixels = """
abcd
efgh
ijkl
mnop
"""
`)
	sf, err := ParseSpriteFile(input, "test.sprite")
	if err != nil {
		t.Fatal(err)
	}
	if sf.PaletteRef != "default" {
		t.Errorf("palette = %q, want default", sf.PaletteRef)
	}
	if sf.DefaultGrid.W != 4 || sf.DefaultGrid.H != 4 {
		t.Errorf("grid = %v, want 4x4", sf.DefaultGrid)
	}
	if len(sf.Sprites) != 1 {
		t.Fatalf("got %d sprites, want 1", len(sf.Sprites))
	}
	s := sf.Sprites[0]
	if s.Name != "test" {
		t.Errorf("name = %q, want test", s.Name)
	}
	if len(s.Frames) != 1 {
		t.Errorf("got %d frames, want 1", len(s.Frames))
	}
}

func TestParseSpriteFile_Animated(t *testing.T) {
	input := []byte(`
palette = "default"
grid = 2

[sprite.blink]
framerate = 4

[[sprite.blink.frame]]
pixels = """
ab
cd
"""

[[sprite.blink.frame]]
pixels = """
ef
gh
"""
`)
	sf, err := ParseSpriteFile(input, "test.sprite")
	if err != nil {
		t.Fatal(err)
	}
	if len(sf.Sprites) != 1 {
		t.Fatalf("got %d sprites, want 1", len(sf.Sprites))
	}
	s := sf.Sprites[0]
	if s.Framerate != 4 {
		t.Errorf("framerate = %d, want 4", s.Framerate)
	}
	if len(s.Frames) != 2 {
		t.Errorf("got %d frames, want 2", len(s.Frames))
	}
}

func TestParseSpriteFile_GridMismatch(t *testing.T) {
	input := []byte(`
palette = "default"
grid = 4

[sprite.bad]
pixels = """
abc
def
ghi
"""
`)
	_, err := ParseSpriteFile(input, "test.sprite")
	if err == nil {
		t.Fatal("expected grid mismatch error")
	}
}

func TestParseSpriteFile_FrameDimensionMismatch(t *testing.T) {
	input := []byte(`
palette = "default"
grid = 2

[sprite.bad]
framerate = 4

[[sprite.bad.frame]]
pixels = """
ab
cd
"""

[[sprite.bad.frame]]
pixels = """
abc
def
ghi
"""
`)
	_, err := ParseSpriteFile(input, "test.sprite")
	if err == nil {
		t.Fatal("expected frame dimension mismatch error")
	}
}

func TestParseSpriteFile_NonSquareGrid(t *testing.T) {
	input := []byte(`
palette = "default"
grid = "4x2"

[sprite.wide]
pixels = """
abcd
efgh
"""
`)
	sf, err := ParseSpriteFile(input, "test.sprite")
	if err != nil {
		t.Fatal(err)
	}
	s := sf.Sprites[0]
	if s.Grid.W != 4 || s.Grid.H != 2 {
		t.Errorf("grid = %v, want 4x2", s.Grid)
	}
}

func TestSpriteFile_Resolve(t *testing.T) {
	pal := &palette.Palette{
		Name: "test",
		Colors: map[string]palette.Color{
			"r": {R: 255, A: 255},
			"_": {A: 0},
		},
	}

	sf := &SpriteFile{
		PaletteRef: "test",
		DefaultGrid: Grid{W: 2, H: 2},
		Sprites: []Sprite{
			{
				Name: "dot",
				Grid: Grid{W: 2, H: 2},
				Frames: []Frame{
					{Pixels: [][]string{{"_", "r"}, {"r", "_"}}},
				},
			},
		},
	}

	resolved, err := sf.Resolve(pal)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolved) != 1 {
		t.Fatalf("got %d resolved, want 1", len(resolved))
	}
	rs := resolved[0]
	if !rs.Frames[0].Pixels[0][0].IsTransparent() {
		t.Error("expected transparent at 0,0")
	}
	if rs.Frames[0].Pixels[0][1].R != 255 {
		t.Error("expected red at 0,1")
	}
}

func TestSpriteFile_Resolve_UnknownKey(t *testing.T) {
	pal := &palette.Palette{
		Name:   "test",
		Colors: map[string]palette.Color{"r": {R: 255, A: 255}},
	}
	sf := &SpriteFile{
		Sprites: []Sprite{
			{
				Name: "bad",
				Grid: Grid{W: 2, H: 1},
				Frames: []Frame{
					{Pixels: [][]string{{"r", "x"}}},
				},
			},
		},
	}
	_, err := sf.Resolve(pal)
	if err == nil {
		t.Fatal("expected unknown key error")
	}
	if !strings.Contains(err.Error(), "unknown palette keys") {
		t.Errorf("error should mention unknown keys: %v", err)
	}
}

func TestSpriteFile_Resolve_PaletteExtend(t *testing.T) {
	pal := &palette.Palette{
		Name:   "test",
		Colors: map[string]palette.Color{"_": {A: 0}},
	}
	sf := &SpriteFile{
		PaletteExtend: map[string]string{"x": "#ff0000"},
		Sprites: []Sprite{
			{
				Name: "ext",
				Grid: Grid{W: 2, H: 1},
				Frames: []Frame{
					{Pixels: [][]string{{"_", "x"}}},
				},
			},
		},
	}
	resolved, err := sf.Resolve(pal)
	if err != nil {
		t.Fatal(err)
	}
	if resolved[0].Frames[0].Pixels[0][1].R != 255 {
		t.Error("palette_extend key 'x' not resolved to red")
	}
}
