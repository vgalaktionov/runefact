package palette

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseHexColor_RGB(t *testing.T) {
	c, err := ParseHexColor("#f00")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 0 || c.B != 0 || c.A != 255 {
		t.Errorf("got %+v, want red", c)
	}
}

func TestParseHexColor_RRGGBB(t *testing.T) {
	c, err := ParseHexColor("#ff004d")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 0 || c.B != 77 || c.A != 255 {
		t.Errorf("got %+v, want {255,0,77,255}", c)
	}
}

func TestParseHexColor_RRGGBBAA(t *testing.T) {
	c, err := ParseHexColor("#ff004d80")
	if err != nil {
		t.Fatal(err)
	}
	if c.R != 255 || c.G != 0 || c.B != 77 || c.A != 128 {
		t.Errorf("got %+v, want {255,0,77,128}", c)
	}
}

func TestParseHexColor_Invalid(t *testing.T) {
	cases := []string{"#xyz", "#12345", "ff0000", "#1", "#12", "#1234", "#12345", "#1234567", "#123456789"}
	for _, s := range cases {
		if _, err := ParseHexColor(s); err == nil {
			t.Errorf("expected error for %q", s)
		}
	}
}

func TestParsePalette_Default(t *testing.T) {
	input := []byte(`name = "default"

[colors]
_ = "transparent"
k = "#000000"
w = "#ffffff"
r = "#ff004d"
b = "#29adff"
`)
	p, err := ParsePalette(input, "default.palette")
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "default" {
		t.Errorf("name = %q, want default", p.Name)
	}
	if len(p.Colors) != 5 {
		t.Errorf("got %d colors, want 5", len(p.Colors))
	}
	if !p.Colors["_"].IsTransparent() {
		t.Error("_ should be transparent")
	}
	if c := p.Colors["k"]; c.R != 0 || c.G != 0 || c.B != 0 {
		t.Errorf("k = %+v, want black", c)
	}
	if c := p.Colors["w"]; c.R != 255 || c.G != 255 || c.B != 255 {
		t.Errorf("w = %+v, want white", c)
	}
}

func TestParsePalette_MultiCharKeys(t *testing.T) {
	input := []byte(`name = "extended"

[colors]
sk = "#ffccaa"
rb = "#4400ff"
`)
	p, err := ParsePalette(input, "test.palette")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := p.Colors["sk"]; !ok {
		t.Error("missing multi-char key 'sk'")
	}
	if _, ok := p.Colors["rb"]; !ok {
		t.Error("missing multi-char key 'rb'")
	}
}

func TestParsePalette_InvalidColor(t *testing.T) {
	input := []byte(`name = "bad"

[colors]
x = "#xyz"
`)
	_, err := ParsePalette(input, "bad.palette")
	if err == nil {
		t.Fatal("expected error for invalid color")
	}
}

func TestParsePalette_MalformedTOML(t *testing.T) {
	input := []byte(`[broken`)
	_, err := ParsePalette(input, "broken.palette")
	if err == nil {
		t.Fatal("expected error for malformed TOML")
	}
}

func TestLoadPalette(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.palette")
	content := `name = "test"

[colors]
r = "#ff0000"
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	p, err := LoadPalette(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "test" {
		t.Errorf("name = %q, want test", p.Name)
	}
}

func TestResolvePalette(t *testing.T) {
	dir := t.TempDir()
	content := `name = "default"

[colors]
k = "#000000"
`
	if err := os.WriteFile(filepath.Join(dir, "default.palette"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p, err := ResolvePalette("default", []string{dir})
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "default" {
		t.Errorf("name = %q, want default", p.Name)
	}
}

func TestResolvePalette_NotFound(t *testing.T) {
	_, err := ResolvePalette("missing", []string{t.TempDir()})
	if err == nil {
		t.Fatal("expected error for missing palette")
	}
}

func TestSuggestSimilarKey(t *testing.T) {
	available := []string{"sk", "rb", "ht", "k", "w"}

	tests := []struct {
		input string
		want  string
	}{
		{"skn", "sk"},
		{"rbb", "rb"},
		{"K", "k"},
		{"xyzabc", ""},
	}
	for _, tt := range tests {
		got := SuggestSimilarKey(tt.input, available)
		if got != tt.want {
			t.Errorf("SuggestSimilarKey(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestColor_ToRGBA(t *testing.T) {
	c := Color{R: 255, G: 128, B: 64, A: 200}
	rgba := c.ToRGBA()
	if rgba.R != 255 || rgba.G != 128 || rgba.B != 64 || rgba.A != 200 {
		t.Errorf("ToRGBA = %+v, want {255,128,64,200}", rgba)
	}
}
