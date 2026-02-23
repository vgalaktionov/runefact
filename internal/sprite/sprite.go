package sprite

import (
	"fmt"
	"os"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/vgalaktionov/runefact/internal/palette"
)

// Grid represents sprite dimensions (width x height).
type Grid struct {
	W, H int
}

// Frame holds a parsed pixel grid as 2D palette keys.
type Frame struct {
	Pixels [][]string
}

// Sprite is a single named sprite with optional animation frames.
type Sprite struct {
	Name      string
	Grid      Grid
	Framerate int
	Frames    []Frame
}

// SpriteFile represents a parsed .sprite file.
type SpriteFile struct {
	PaletteRef    string
	PaletteExtend map[string]string
	DefaultGrid   Grid
	Sprites       []Sprite
}

// ResolvedSprite has palette keys replaced with actual colors.
type ResolvedSprite struct {
	Name      string
	Grid      Grid
	Framerate int
	Frames    []ResolvedFrame
}

// ResolvedFrame contains color-resolved pixel data.
type ResolvedFrame struct {
	Pixels [][]palette.Color
}

// ParsePixelGrid parses a pixel grid string into a 2D array of palette keys.
func ParsePixelGrid(raw string) ([][]string, error) {
	lines := strings.Split(strings.TrimSpace(raw), "\n")
	var grid [][]string
	var expectedWidth int

	for lineNum, line := range lines {
		line = strings.TrimRight(line, " \t\r")
		if line == "" {
			continue
		}
		row, err := parseGridRow(line)
		if err != nil {
			return nil, fmt.Errorf("line %d: %w", lineNum+1, err)
		}
		if len(grid) == 0 {
			expectedWidth = len(row)
		} else if len(row) != expectedWidth {
			return nil, fmt.Errorf("line %d: ragged row, expected width %d, got %d", lineNum+1, expectedWidth, len(row))
		}
		grid = append(grid, row)
	}
	return grid, nil
}

func parseGridRow(line string) ([]string, error) {
	var row []string
	i := 0
	for i < len(line) {
		if line[i] == '[' {
			end := strings.Index(line[i:], "]")
			if end == -1 {
				return nil, fmt.Errorf("unclosed bracket at position %d", i)
			}
			key := line[i+1 : i+end]
			if key == "" {
				return nil, fmt.Errorf("empty bracket key at position %d", i)
			}
			row = append(row, key)
			i += end + 1
		} else {
			row = append(row, string(line[i]))
			i++
		}
	}
	return row, nil
}

// rawSpriteFile is the TOML-level structure for deserializing .sprite files.
type rawSpriteFile struct {
	Palette       string            `toml:"palette"`
	Grid          interface{}       `toml:"grid"` // int or string "WxH"
	PaletteExtend map[string]string `toml:"palette_extend"`
	Sprite        map[string]rawSprite
}

type rawSprite struct {
	Grid          interface{}       `toml:"grid"`
	Framerate     int               `toml:"framerate"`
	Pixels        string            `toml:"pixels"`
	PaletteExtend map[string]string `toml:"palette_extend"`
	Frame         []rawFrame        `toml:"frame"`
}

type rawFrame struct {
	Pixels string `toml:"pixels"`
}

// ParseSpriteFile parses .sprite file content.
func ParseSpriteFile(data []byte, filename string) (*SpriteFile, error) {
	var raw rawSpriteFile
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	defaultGrid, err := parseGrid(raw.Grid)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	sf := &SpriteFile{
		PaletteRef:    raw.Palette,
		PaletteExtend: raw.PaletteExtend,
		DefaultGrid:   defaultGrid,
	}

	for name, rs := range raw.Sprite {
		sprite, err := parseSprite(name, rs, defaultGrid, filename)
		if err != nil {
			return nil, err
		}
		sf.Sprites = append(sf.Sprites, *sprite)
	}

	return sf, nil
}

func parseSprite(name string, raw rawSprite, defaultGrid Grid, filename string) (*Sprite, error) {
	grid, err := parseGrid(raw.Grid)
	if err != nil {
		return nil, fmt.Errorf("%s: sprite %q: %w", filename, name, err)
	}
	if grid.W == 0 && grid.H == 0 {
		grid = defaultGrid
	}

	s := &Sprite{
		Name:      name,
		Grid:      grid,
		Framerate: raw.Framerate,
	}

	if raw.Pixels != "" {
		// Static sprite: single frame.
		pixels, err := ParsePixelGrid(raw.Pixels)
		if err != nil {
			return nil, fmt.Errorf("%s: sprite %q: %w", filename, name, err)
		}
		s.Frames = []Frame{{Pixels: pixels}}
	} else if len(raw.Frame) > 0 {
		// Animated sprite: multiple frames.
		for i, f := range raw.Frame {
			pixels, err := ParsePixelGrid(f.Pixels)
			if err != nil {
				return nil, fmt.Errorf("%s: sprite %q frame %d: %w", filename, name, i+1, err)
			}
			s.Frames = append(s.Frames, Frame{Pixels: pixels})
		}
	}

	// Validate frame dimensions.
	if err := validateFrames(s, filename); err != nil {
		return nil, err
	}

	return s, nil
}

func parseGrid(v interface{}) (Grid, error) {
	if v == nil {
		return Grid{}, nil
	}
	switch val := v.(type) {
	case int64:
		return Grid{W: int(val), H: int(val)}, nil
	case float64:
		return Grid{W: int(val), H: int(val)}, nil
	case string:
		var w, h int
		if _, err := fmt.Sscanf(val, "%dx%d", &w, &h); err != nil {
			return Grid{}, fmt.Errorf("invalid grid size %q (expected int or WxH)", val)
		}
		return Grid{W: w, H: h}, nil
	default:
		return Grid{}, fmt.Errorf("invalid grid type %T", v)
	}
}

func validateFrames(s *Sprite, filename string) error {
	if len(s.Frames) == 0 {
		return nil
	}

	firstH := len(s.Frames[0].Pixels)
	firstW := 0
	if firstH > 0 {
		firstW = len(s.Frames[0].Pixels[0])
	}

	for i, f := range s.Frames {
		h := len(f.Pixels)
		w := 0
		if h > 0 {
			w = len(f.Pixels[0])
		}
		if w != firstW || h != firstH {
			return fmt.Errorf("%s: sprite %q: frame %d dimensions %dx%d differ from frame 1 (%dx%d)",
				filename, s.Name, i+1, w, h, firstW, firstH)
		}
	}

	// Validate grid matches actual dimensions if grid is set.
	if s.Grid.W > 0 && s.Grid.H > 0 {
		if firstW != s.Grid.W || firstH != s.Grid.H {
			return fmt.Errorf("%s: sprite %q: pixel dimensions %dx%d don't match grid %dx%d",
				filename, s.Name, firstW, firstH, s.Grid.W, s.Grid.H)
		}
	}

	return nil
}

// LoadSpriteFile reads and parses a .sprite file from disk.
func LoadSpriteFile(path string) (*SpriteFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading sprite file: %w", err)
	}
	return ParseSpriteFile(data, path)
}

// Resolve resolves palette keys to actual colors for all sprites.
func (sf *SpriteFile) Resolve(pal *palette.Palette) ([]ResolvedSprite, error) {
	// Merge palette_extend into a combined color map.
	colors := make(map[string]palette.Color, len(pal.Colors))
	for k, v := range pal.Colors {
		colors[k] = v
	}
	for k, hex := range sf.PaletteExtend {
		c, err := palette.ParseHexColor(hex)
		if err != nil {
			return nil, fmt.Errorf("palette_extend key %q: %w", k, err)
		}
		colors[k] = c
	}

	// "_" is always transparent, even if not defined in the palette.
	colors["_"] = palette.Color{A: 0}

	var resolved []ResolvedSprite
	for _, s := range sf.Sprites {
		rs, err := resolveSprite(s, colors)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, *rs)
	}
	return resolved, nil
}

func resolveSprite(s Sprite, colors map[string]palette.Color) (*ResolvedSprite, error) {
	rs := &ResolvedSprite{
		Name:      s.Name,
		Grid:      s.Grid,
		Framerate: s.Framerate,
	}

	var unknownKeys []string
	for _, f := range s.Frames {
		rf := ResolvedFrame{Pixels: make([][]palette.Color, len(f.Pixels))}
		for y, row := range f.Pixels {
			rf.Pixels[y] = make([]palette.Color, len(row))
			for x, key := range row {
				c, ok := colors[key]
				if !ok {
					unknownKeys = append(unknownKeys, key)
					continue
				}
				rf.Pixels[y][x] = c
			}
		}
		rs.Frames = append(rs.Frames, rf)
	}

	if len(unknownKeys) > 0 {
		// Deduplicate.
		seen := map[string]bool{}
		var unique []string
		for _, k := range unknownKeys {
			if !seen[k] {
				seen[k] = true
				unique = append(unique, k)
			}
		}

		available := make([]string, 0, len(colors))
		for k := range colors {
			available = append(available, k)
		}

		msg := fmt.Sprintf("sprite %q: unknown palette keys: %v", s.Name, unique)
		for _, k := range unique {
			if suggestion := palette.SuggestSimilarKey(k, available); suggestion != "" {
				msg += fmt.Sprintf(" (did you mean %q for %q?)", suggestion, k)
			}
		}
		return nil, fmt.Errorf("%s", msg)
	}

	return rs, nil
}
