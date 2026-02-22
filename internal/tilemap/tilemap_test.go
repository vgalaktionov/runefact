package tilemap

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseMapFile_Basic(t *testing.T) {
	input := []byte(`
tile_size = 16

[tileset]
G = "tiles:grass"
D = "tiles:dirt"
_ = ""

[layer.background]
scroll_x = 0.5
pixels = """
GGGG
DDDD
"""

[layer.objects]
[[layer.objects.entity]]
type = "spawn"
x = 2
y = 5

[[layer.objects.entity]]
type = "coin"
x = 7
y = 4
properties = { value = 10 }
`)
	mf, warnings, err := ParseMapFile(input, "test.map")
	if err != nil {
		t.Fatal(err)
	}

	if mf.TileSize != 16 {
		t.Errorf("tile_size = %d, want 16", mf.TileSize)
	}
	if len(mf.Tileset) != 3 {
		t.Errorf("tileset has %d entries, want 3", len(mf.Tileset))
	}
	if len(mf.Layers) != 2 {
		t.Fatalf("got %d layers, want 2", len(mf.Layers))
	}
	if len(warnings) != 0 {
		t.Errorf("unexpected warnings: %v", warnings)
	}

	// Find tile layer.
	var tileLyr, entityLyr *Layer
	for i := range mf.Layers {
		switch mf.Layers[i].Type {
		case "tile":
			tileLyr = &mf.Layers[i]
		case "entity":
			entityLyr = &mf.Layers[i]
		}
	}

	if tileLyr == nil {
		t.Fatal("no tile layer found")
	}
	if tileLyr.ScrollX != 0.5 {
		t.Errorf("scroll_x = %f, want 0.5", tileLyr.ScrollX)
	}
	if len(tileLyr.Data) != 2 || len(tileLyr.Data[0]) != 4 {
		t.Errorf("tile data size = %dx%d, want 4x2", len(tileLyr.Data[0]), len(tileLyr.Data))
	}

	if entityLyr == nil {
		t.Fatal("no entity layer found")
	}
	if len(entityLyr.Entities) != 2 {
		t.Errorf("got %d entities, want 2", len(entityLyr.Entities))
	}
	if entityLyr.Entities[0].Type != "spawn" {
		t.Errorf("entity 0 type = %q, want spawn", entityLyr.Entities[0].Type)
	}
}

func TestParseMapFile_UnknownTilesetKey(t *testing.T) {
	input := []byte(`
tile_size = 8

[tileset]
G = "tiles:grass"
_ = ""

[layer.main]
pixels = """
GX
GG
"""
`)
	_, warnings, err := ParseMapFile(input, "test.map")
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) == 0 {
		t.Fatal("expected warning for unknown key 'X'")
	}
	if !strings.Contains(warnings[0].Message, "X") {
		t.Errorf("warning should mention key X: %s", warnings[0].Message)
	}
}

func TestParseMapFile_RaggedRows(t *testing.T) {
	input := []byte(`
tile_size = 8

[tileset]
G = "tiles:grass"

[layer.main]
pixels = """
GGG
GG
"""
`)
	_, _, err := ParseMapFile(input, "test.map")
	if err == nil {
		t.Fatal("expected error for ragged rows")
	}
}

func TestParseMapFile_InvalidTileSize(t *testing.T) {
	input := []byte(`
tile_size = 0

[tileset]
G = "tiles:grass"
`)
	_, _, err := ParseMapFile(input, "test.map")
	if err == nil {
		t.Fatal("expected error for tile_size=0")
	}
}

func TestToJSON(t *testing.T) {
	mf := &MapFile{
		TileSize: 16,
		Tileset:  map[string]string{"G": "tiles:grass", "_": ""},
		Layers: []Layer{
			{
				Name:    "bg",
				Type:    "tile",
				ScrollX: 0.5,
				Data:    [][]int{{1, 0}, {1, 1}},
			},
			{
				Name: "objects",
				Type: "entity",
				Entities: []Entity{
					{Type: "spawn", X: 2, Y: 5},
				},
			},
		},
	}

	j := mf.ToJSON()
	if j.TileSize != 16 {
		t.Errorf("tile_size = %d, want 16", j.TileSize)
	}
	if j.Width != 2 || j.Height != 2 {
		t.Errorf("dimensions = %dx%d, want 2x2", j.Width, j.Height)
	}
	if len(j.Tileset) != 1 { // _ is empty, excluded
		t.Errorf("tileset has %d entries, want 1", len(j.Tileset))
	}
	ref, ok := j.Tileset["G"]
	if !ok {
		t.Fatal("missing tileset entry G")
	}
	if ref.Source != "tiles.png" || ref.Sprite != "grass" {
		t.Errorf("tileset G = %+v", ref)
	}
	if len(j.Layers) != 2 {
		t.Errorf("got %d layers, want 2", len(j.Layers))
	}
}

func TestWriteJSON(t *testing.T) {
	tm := &JSONTilemap{
		TileSize: 8,
		Width:    2,
		Height:   2,
		Tileset:  map[string]JSONTileRef{},
		Layers:   []JSONLayer{},
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "maps", "test.json")
	if err := WriteJSON(tm, path); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded JSONTilemap
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.TileSize != 8 {
		t.Errorf("decoded tile_size = %d, want 8", decoded.TileSize)
	}
}

func TestParseSpriteRef(t *testing.T) {
	file, spriteName := parseSpriteRef("tiles:grass")
	if file != "tiles" || spriteName != "grass" {
		t.Errorf("got %q, %q, want tiles, grass", file, spriteName)
	}

	file, spriteName = parseSpriteRef("nocolon")
	if file != "nocolon" || spriteName != "" {
		t.Errorf("got %q, %q, want nocolon, empty", file, spriteName)
	}
}
