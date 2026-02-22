package tilemap

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml/v2"

	"github.com/vgalaktionov/runefact/internal/sprite"
)

// MapFile represents a parsed .map file.
type MapFile struct {
	TileSize int
	Tileset  map[string]string // char -> "file:sprite" or ""
	Layers   []Layer
}

// Layer is either a tile layer (with grid data) or an entity layer.
type Layer struct {
	Name     string
	Type     string // "tile" or "entity"
	ScrollX  float64
	ScrollY  float64
	Data     [][]int  // tile indices for tile layers
	Entities []Entity // entities for entity layers
}

// Entity represents a placed object in an entity layer.
type Entity struct {
	Type       string                 `json:"type"`
	X          int                    `json:"x"`
	Y          int                    `json:"y"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// Warning represents a non-fatal issue found during parsing.
type Warning struct {
	Message string
}

// rawMap is the TOML-level structure.
type rawMap struct {
	TileSize int               `toml:"tile_size"`
	Tileset  map[string]string `toml:"tileset"`
	Layer    map[string]rawLayer
}

type rawLayer struct {
	ScrollX float64     `toml:"scroll_x"`
	ScrollY float64     `toml:"scroll_y"`
	Pixels  string      `toml:"pixels"`
	Entity  []rawEntity `toml:"entity"`
}

type rawEntity struct {
	Type       string                 `toml:"type"`
	X          int                    `toml:"x"`
	Y          int                    `toml:"y"`
	Properties map[string]interface{} `toml:"properties"`
}

// ParseMapFile parses .map file content.
func ParseMapFile(data []byte, filename string) (*MapFile, []Warning, error) {
	var raw rawMap
	if err := toml.Unmarshal(data, &raw); err != nil {
		return nil, nil, fmt.Errorf("%s: %w", filename, err)
	}

	if raw.TileSize <= 0 {
		return nil, nil, fmt.Errorf("%s: tile_size must be positive", filename)
	}

	mf := &MapFile{
		TileSize: raw.TileSize,
		Tileset:  raw.Tileset,
	}

	var warnings []Warning

	// Build tileset index: assign each tileset key a numeric index.
	tileIndex := buildTileIndex(raw.Tileset)

	for name, rl := range raw.Layer {
		layer, layerWarnings, err := parseLayer(name, rl, tileIndex, filename)
		if err != nil {
			return nil, nil, err
		}
		warnings = append(warnings, layerWarnings...)
		mf.Layers = append(mf.Layers, *layer)
	}

	return mf, warnings, nil
}

func buildTileIndex(tileset map[string]string) map[string]int {
	idx := make(map[string]int, len(tileset))
	nextID := 0
	for key, ref := range tileset {
		if ref == "" {
			idx[key] = 0 // empty tile
		} else {
			nextID++
			idx[key] = nextID
		}
	}
	return idx
}

func parseLayer(name string, raw rawLayer, tileIndex map[string]int, filename string) (*Layer, []Warning, error) {
	if len(raw.Entity) > 0 {
		return parseEntityLayer(name, raw)
	}
	return parseTileLayer(name, raw, tileIndex, filename)
}

func parseTileLayer(name string, raw rawLayer, tileIndex map[string]int, filename string) (*Layer, []Warning, error) {
	var warnings []Warning

	grid, err := sprite.ParsePixelGrid(raw.Pixels)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: layer %q: %w", filename, name, err)
	}

	data := make([][]int, len(grid))
	for y, row := range grid {
		data[y] = make([]int, len(row))
		for x, key := range row {
			idx, ok := tileIndex[key]
			if !ok {
				warnings = append(warnings, Warning{
					Message: fmt.Sprintf("%s: layer %q: unknown tileset key %q at row %d, col %d", filename, name, key, y+1, x+1),
				})
				data[y][x] = 0
			} else {
				data[y][x] = idx
			}
		}
	}

	return &Layer{
		Name:    name,
		Type:    "tile",
		ScrollX: raw.ScrollX,
		ScrollY: raw.ScrollY,
		Data:    data,
	}, warnings, nil
}

func parseEntityLayer(name string, raw rawLayer) (*Layer, []Warning, error) {
	entities := make([]Entity, len(raw.Entity))
	for i, re := range raw.Entity {
		entities[i] = Entity{
			Type:       re.Type,
			X:          re.X,
			Y:          re.Y,
			Properties: re.Properties,
		}
	}
	return &Layer{
		Name:     name,
		Type:     "entity",
		Entities: entities,
	}, nil, nil
}

// LoadMapFile reads and parses a .map file from disk.
func LoadMapFile(path string) (*MapFile, []Warning, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("loading map file: %w", err)
	}
	return ParseMapFile(data, path)
}

// JSONTilemap is the output JSON format.
type JSONTilemap struct {
	TileSize int                    `json:"tile_size"`
	Width    int                    `json:"width"`
	Height   int                    `json:"height"`
	Tileset  map[string]JSONTileRef `json:"tileset"`
	Layers   []JSONLayer            `json:"layers"`
}

// JSONTileRef describes a tile's source sprite.
type JSONTileRef struct {
	Source string `json:"source"`
	Sprite string `json:"sprite"`
	Index  int    `json:"index"`
}

// JSONLayer is a layer in the output JSON.
type JSONLayer struct {
	Name     string       `json:"name"`
	Type     string       `json:"type"`
	ScrollX  float64      `json:"scroll_x,omitempty"`
	ScrollY  float64      `json:"scroll_y,omitempty"`
	Data     [][]int      `json:"data,omitempty"`
	Entities []JSONEntity `json:"entities,omitempty"`
}

// JSONEntity is an entity in the output JSON.
type JSONEntity struct {
	Type       string                 `json:"type"`
	X          int                    `json:"x"`
	Y          int                    `json:"y"`
	Properties map[string]interface{} `json:"properties,omitempty"`
}

// ToJSON converts a parsed map to the output JSON structure.
func (mf *MapFile) ToJSON() *JSONTilemap {
	// Calculate map dimensions from tile layers.
	width, height := 0, 0
	for _, l := range mf.Layers {
		if l.Type == "tile" && len(l.Data) > 0 {
			if len(l.Data) > height {
				height = len(l.Data)
			}
			if len(l.Data[0]) > width {
				width = len(l.Data[0])
			}
		}
	}

	// Build tileset refs.
	tileIndex := buildTileIndex(mf.Tileset)
	tilesetJSON := make(map[string]JSONTileRef, len(mf.Tileset))
	for key, ref := range mf.Tileset {
		if ref == "" {
			continue
		}
		source, spriteName := parseSpriteRef(ref)
		tilesetJSON[key] = JSONTileRef{
			Source: source + ".png",
			Sprite: spriteName,
			Index:  tileIndex[key],
		}
	}

	// Build layers.
	layers := make([]JSONLayer, len(mf.Layers))
	for i, l := range mf.Layers {
		jl := JSONLayer{
			Name:    l.Name,
			Type:    l.Type,
			ScrollX: l.ScrollX,
			ScrollY: l.ScrollY,
		}
		if l.Type == "tile" {
			jl.Data = l.Data
		} else {
			entities := make([]JSONEntity, len(l.Entities))
			for j, e := range l.Entities {
				entities[j] = JSONEntity{
					Type:       e.Type,
					X:          e.X,
					Y:          e.Y,
					Properties: e.Properties,
				}
			}
			jl.Entities = entities
		}
		layers[i] = jl
	}

	return &JSONTilemap{
		TileSize: mf.TileSize,
		Width:    width,
		Height:   height,
		Tileset:  tilesetJSON,
		Layers:   layers,
	}
}

// parseSpriteRef splits "file:sprite" into components.
func parseSpriteRef(ref string) (file, spriteName string) {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return ref, ""
}

// WriteJSON writes the tilemap JSON to a file.
func WriteJSON(tm *JSONTilemap, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}
	data, err := json.MarshalIndent(tm, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	data = append(data, '\n')
	return os.WriteFile(path, data, 0644)
}
