package palette

import (
	"fmt"
	"image/color"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	toml "github.com/pelletier/go-toml/v2"
)

// Color represents an RGBA color value.
type Color struct {
	R, G, B, A uint8
}

// ToRGBA converts to Go's image/color.RGBA.
func (c Color) ToRGBA() color.RGBA {
	return color.RGBA{R: c.R, G: c.G, B: c.B, A: c.A}
}

// IsTransparent reports whether the color is fully transparent.
func (c Color) IsTransparent() bool {
	return c.A == 0
}

// Palette represents a parsed .palette file.
type Palette struct {
	Name   string
	Colors map[string]Color
}

// rawPalette is the TOML-level structure.
type rawPalette struct {
	Name   string            `toml:"name"`
	Colors map[string]string `toml:"colors"`
}

// ParsePalette parses .palette file content.
func ParsePalette(data []byte, filename string) (*Palette, error) {
	var raw rawPalette
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("%s: %w", filename, err)
	}

	p := &Palette{
		Name:   raw.Name,
		Colors: make(map[string]Color, len(raw.Colors)),
	}

	for key, value := range raw.Colors {
		if value == "transparent" {
			p.Colors[key] = Color{A: 0}
			continue
		}
		c, err := ParseHexColor(value)
		if err != nil {
			return nil, fmt.Errorf("%s: invalid color for key %q: %w", filename, key, err)
		}
		p.Colors[key] = c
	}

	return p, nil
}

// LoadPalette reads and parses a .palette file from disk.
func LoadPalette(path string) (*Palette, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("loading palette: %w", err)
	}
	return ParsePalette(data, filepath.Base(path))
}

// ResolvePalette searches for a palette by name in the given directories.
func ResolvePalette(name string, searchPaths []string) (*Palette, error) {
	for _, dir := range searchPaths {
		path := filepath.Join(dir, name+".palette")
		if _, err := os.Stat(path); err == nil {
			return LoadPalette(path)
		}
	}
	return nil, fmt.Errorf("palette %q not found in search paths: %v", name, searchPaths)
}

// ParseHexColor parses hex color strings: #RGB, #RRGGBB, #RRGGBBAA.
func ParseHexColor(hex string) (Color, error) {
	if !strings.HasPrefix(hex, "#") {
		return Color{}, fmt.Errorf("color must start with #, got %q", hex)
	}
	hex = hex[1:]

	switch len(hex) {
	case 3: // #RGB -> #RRGGBB
		r, err := strconv.ParseUint(string(hex[0])+string(hex[0]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		g, err := strconv.ParseUint(string(hex[1])+string(hex[1]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		b, err := strconv.ParseUint(string(hex[2])+string(hex[2]), 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		return Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil

	case 6: // #RRGGBB
		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		return Color{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}, nil

	case 8: // #RRGGBBAA
		r, err := strconv.ParseUint(hex[0:2], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		g, err := strconv.ParseUint(hex[2:4], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		b, err := strconv.ParseUint(hex[4:6], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		a, err := strconv.ParseUint(hex[6:8], 16, 8)
		if err != nil {
			return Color{}, fmt.Errorf("invalid hex color #%s: %w", hex, err)
		}
		return Color{R: uint8(r), G: uint8(g), B: uint8(b), A: uint8(a)}, nil

	default:
		return Color{}, fmt.Errorf("invalid hex color length: #%s (expected 3, 6, or 8 hex digits)", hex)
	}
}

// SuggestSimilarKey returns the closest key from available using Levenshtein distance.
// Returns empty string if no reasonable match found (distance > 3).
func SuggestSimilarKey(unknown string, available []string) string {
	best := ""
	bestDist := 4 // max acceptable distance
	for _, key := range available {
		d := levenshtein(unknown, key)
		if d < bestDist {
			bestDist = d
			best = key
		}
	}
	return best
}

func levenshtein(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(curr[j-1]+1, min(prev[j]+1, prev[j-1]+cost))
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}
