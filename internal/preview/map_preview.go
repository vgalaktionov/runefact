package preview

import (
	"fmt"
	"image"
	"image/color"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/vgalaktionov/runefact/internal/palette"
	"github.com/vgalaktionov/runefact/internal/sprite"
	"github.com/vgalaktionov/runefact/internal/tilemap"
)

// LayerVisibility tracks which layers are visible.
type LayerVisibility int

const (
	LayerVisAll    LayerVisibility = iota
	LayerVisEntity                         // entity layers only
	// Additional values represent individual tile layer indices (starting at 2).
)

// MapPreviewState holds map preview state inside the Previewer.
type MapPreviewState struct {
	mapFile    *tilemap.MapFile
	camX, camY float64
	mapZoom    float64
	layerVis   LayerVisibility
	gridVis    bool
	layerCount int

	// tileImages maps tile ID (1+) to a rendered ebiten.Image of the tile sprite.
	tileImages map[int]*ebiten.Image

	// entityImages maps "file:sprite" ref to a rendered ebiten.Image.
	entityImages map[string]*ebiten.Image
}

func (p *Previewer) initMapState(mf *tilemap.MapFile) {
	tileLayerCount := 0
	mapW, mapH := 0, 0
	for _, l := range mf.Layers {
		if l.Type == "tile" {
			tileLayerCount++
			if len(l.Data) > mapH {
				mapH = len(l.Data)
			}
			if len(l.Data) > 0 && len(l.Data[0]) > mapW {
				mapW = len(l.Data[0])
			}
		}
	}

	// Auto-zoom to fill window, leaving room for the label.
	zoom := 2.0
	if mapW > 0 && mapH > 0 {
		pixW := float64(mapW * mf.TileSize)
		pixH := float64(mapH * mf.TileSize)
		labelMargin := float64(scaledCharH() + 20)
		zx := float64(p.winW) * 0.9 / pixW
		zy := (float64(p.winH) - labelMargin) * 0.9 / pixH
		zoom = min(zx, zy)
		zoom = float64(max(1, int(zoom)))
	}

	// Center the map by offsetting camera.
	camX := -(float64(p.winW) - float64(mapW*mf.TileSize)*zoom) / 2
	camY := -(float64(p.winH) - float64(mapH*mf.TileSize)*zoom) / 2
	if camY > 0 {
		camY = 0
	}

	// Load tile sprite images.
	tileImages := p.loadTileImages(mf)

	// Load entity sprite images.
	entityImages := p.loadEntityImages(mf)

	p.mapState = &MapPreviewState{
		mapFile:      mf,
		camX:         camX,
		camY:         camY,
		mapZoom:      zoom,
		layerCount:   tileLayerCount,
		tileImages:   tileImages,
		entityImages: entityImages,
	}
}

// loadTileImages resolves tileset references to actual sprite images.
// It builds the same tile index as the parser to map tile IDs to sprite refs.
func (p *Previewer) loadTileImages(mf *tilemap.MapFile) map[int]*ebiten.Image {
	images := make(map[int]*ebiten.Image)

	// Rebuild the tile index (same logic as tilemap.buildTileIndex).
	tileIndex := make(map[string]int)
	nextID := 0
	for key, ref := range mf.Tileset {
		if ref == "" {
			tileIndex[key] = 0
		} else {
			nextID++
			tileIndex[key] = nextID
		}
	}

	// Cache loaded sprite files to avoid reloading the same file.
	type spriteCache struct {
		sprites []sprite.ResolvedSprite
	}
	cache := make(map[string]*spriteCache)

	for key, ref := range mf.Tileset {
		if ref == "" {
			continue
		}
		id := tileIndex[key]

		// Parse "file:sprite" reference.
		parts := strings.SplitN(ref, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fileName, spriteName := parts[0], parts[1]

		// Load and resolve sprites from file (cached).
		resolved, ok := cache[fileName]
		if !ok {
			resolved = &spriteCache{}
			spritePath := filepath.Join(p.assetsDir, "sprites", fileName+".sprite")
			sf, err := sprite.LoadSpriteFile(spritePath)
			if err == nil {
				var pal *palette.Palette
				if sf.PaletteRef != "" {
					palPath := filepath.Join(p.assetsDir, "palettes", sf.PaletteRef+".palette")
					pal, _ = palette.LoadPalette(palPath)
				}
				if pal == nil {
					pal = &palette.Palette{Colors: map[string]palette.Color{}}
				}
				rs, err := sf.Resolve(pal)
				if err == nil {
					resolved.sprites = rs
				}
			}
			cache[fileName] = resolved
		}

		// Find the named sprite and render first frame.
		for _, rs := range resolved.sprites {
			if rs.Name == spriteName && len(rs.Frames) > 0 {
				img := image.NewRGBA(image.Rect(0, 0, rs.Grid.W, rs.Grid.H))
				for y, row := range rs.Frames[0].Pixels {
					for x, c := range row {
						img.Set(x, y, c.ToRGBA())
					}
				}
				images[id] = ebiten.NewImageFromImage(img)
				break
			}
		}
	}

	return images
}

func (p *Previewer) updateMap() {
	ms := p.mapState
	if ms == nil {
		return
	}

	speed := 4.0
	if ebiten.IsKeyPressed(ebiten.KeyW) || ebiten.IsKeyPressed(ebiten.KeyArrowUp) {
		ms.camY -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyS) || ebiten.IsKeyPressed(ebiten.KeyArrowDown) {
		ms.camY += speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyA) || ebiten.IsKeyPressed(ebiten.KeyArrowLeft) {
		ms.camX -= speed
	}
	if ebiten.IsKeyPressed(ebiten.KeyD) || ebiten.IsKeyPressed(ebiten.KeyArrowRight) {
		ms.camX += speed
	}

	// Zoom.
	_, dy := ebiten.Wheel()
	if dy > 0 && ms.mapZoom < 16 {
		ms.mapZoom *= 1.5
	} else if dy < 0 && ms.mapZoom > 0.25 {
		ms.mapZoom /= 1.5
	}

	// Tab: cycle layer visibility.
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		maxVis := LayerVisibility(2 + ms.layerCount) // all, entity, then each tile layer
		ms.layerVis = (ms.layerVis + 1) % maxVis
	}

	// G: toggle grid.
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		ms.gridVis = !ms.gridVis
	}
}

func (p *Previewer) drawMap(screen *ebiten.Image) {
	ms := p.mapState
	if ms == nil || ms.mapFile == nil {
		return
	}

	mf := ms.mapFile
	ts := mf.TileSize
	z := ms.mapZoom

	// Draw tile layers.
	tileLayerIdx := 0
	for _, layer := range mf.Layers {
		if layer.Type == "tile" {
			visible := ms.layerVis == LayerVisAll ||
				(ms.layerVis >= 2 && int(ms.layerVis)-2 == tileLayerIdx)
			if visible {
				p.drawTileLayer(screen, layer, ts, z, ms.camX, ms.camY)
			}
			tileLayerIdx++
		}
	}

	// Draw entity layers.
	for _, layer := range mf.Layers {
		if layer.Type == "entity" {
			visible := ms.layerVis == LayerVisAll || ms.layerVis == LayerVisEntity
			if visible {
				p.drawEntityLayer(screen, layer, ts, z, ms.camX, ms.camY)
			}
		}
	}

	// Grid overlay.
	if ms.gridVis {
		p.drawMapGrid(screen, mf, ts, z, ms.camX, ms.camY)
	}

	// Mode label.
	label := "Map"
	switch {
	case ms.layerVis == LayerVisAll:
		label += " [All layers]"
	case ms.layerVis == LayerVisEntity:
		label += " [Entities]"
	default:
		label += fmt.Sprintf(" [Layer %d/%d]", int(ms.layerVis)-1, ms.layerCount)
	}
	drawText(screen, label, 10, 10)
}

func (p *Previewer) drawTileLayer(screen *ebiten.Image, layer tilemap.Layer, ts int, z, camX, camY float64) {
	ms := p.mapState

	// Viewport culling.
	startCol := max(0, int(camX/z)/ts-1)
	startRow := max(0, int(camY/z)/ts-1)
	endCol := int((camX+float64(p.winW))/z)/ts + 2
	endRow := int((camY+float64(p.winH))/z)/ts + 2

	for y, row := range layer.Data {
		if y < startRow || y > endRow {
			continue
		}
		for x, tileID := range row {
			if x < startCol || x > endCol {
				continue
			}
			if tileID == 0 {
				continue // empty tile
			}

			sx := float64(x*ts)*z - camX
			sy := float64(y*ts)*z - camY

			// Draw actual tile sprite if available.
			if img, ok := ms.tileImages[tileID]; ok {
				op := &ebiten.DrawImageOptions{}
				// Scale sprite to tile size * zoom.
				imgW := float64(img.Bounds().Dx())
				imgH := float64(img.Bounds().Dy())
				op.GeoM.Scale(float64(ts)/imgW*z, float64(ts)/imgH*z)
				op.GeoM.Translate(sx, sy)
				op.Filter = ebiten.FilterNearest
				screen.DrawImage(img, op)
			} else {
				// Fallback: colored rectangle.
				w := float64(ts) * z
				h := float64(ts) * z
				c := tileColor(tileID)
				for dy := 0; dy < int(h) && int(sy)+dy >= 0; dy++ {
					for dx := 0; dx < int(w) && int(sx)+dx >= 0; dx++ {
						px, py := int(sx)+dx, int(sy)+dy
						if px >= 0 && px < p.winW && py >= 0 && py < p.winH {
							screen.Set(px, py, c)
						}
					}
				}
			}
		}
	}
}

func tileColor(id int) color.RGBA {
	colors := []color.RGBA{
		{R: 0x4a, G: 0x9e, B: 0x4a, A: 0xff}, // green
		{R: 0x8b, G: 0x6e, B: 0x4b, A: 0xff}, // brown
		{R: 0x5b, G: 0x7b, B: 0xb5, A: 0xff}, // blue
		{R: 0xc4, G: 0x9a, B: 0x3c, A: 0xff}, // gold
		{R: 0x7b, G: 0x5b, B: 0x9e, A: 0xff}, // purple
		{R: 0xb5, G: 0x5b, B: 0x5b, A: 0xff}, // red
	}
	return colors[(id-1)%len(colors)]
}

// loadEntityImages collects unique sprite refs from entity properties and loads them.
func (p *Previewer) loadEntityImages(mf *tilemap.MapFile) map[string]*ebiten.Image {
	images := make(map[string]*ebiten.Image)

	// Collect unique sprite refs.
	refs := make(map[string]bool)
	for _, layer := range mf.Layers {
		if layer.Type != "entity" {
			continue
		}
		for _, e := range layer.Entities {
			if ref, ok := e.Properties["sprite"]; ok {
				if s, ok := ref.(string); ok && s != "" {
					refs[s] = true
				}
			}
		}
	}

	// Cache loaded sprite files.
	type spriteCache struct {
		sprites []sprite.ResolvedSprite
	}
	cache := make(map[string]*spriteCache)

	for ref := range refs {
		parts := strings.SplitN(ref, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fileName, spriteName := parts[0], parts[1]

		resolved, ok := cache[fileName]
		if !ok {
			resolved = &spriteCache{}
			spritePath := filepath.Join(p.assetsDir, "sprites", fileName+".sprite")
			sf, err := sprite.LoadSpriteFile(spritePath)
			if err == nil {
				var pal *palette.Palette
				if sf.PaletteRef != "" {
					palPath := filepath.Join(p.assetsDir, "palettes", sf.PaletteRef+".palette")
					pal, _ = palette.LoadPalette(palPath)
				}
				if pal == nil {
					pal = &palette.Palette{Colors: map[string]palette.Color{}}
				}
				rs, err := sf.Resolve(pal)
				if err == nil {
					resolved.sprites = rs
				}
			}
			cache[fileName] = resolved
		}

		for _, rs := range resolved.sprites {
			if rs.Name == spriteName && len(rs.Frames) > 0 {
				img := image.NewRGBA(image.Rect(0, 0, rs.Grid.W, rs.Grid.H))
				for y, row := range rs.Frames[0].Pixels {
					for x, c := range row {
						img.Set(x, y, c.ToRGBA())
					}
				}
				images[ref] = ebiten.NewImageFromImage(img)
				break
			}
		}
	}

	return images
}

func (p *Previewer) drawEntityLayer(screen *ebiten.Image, layer tilemap.Layer, ts int, z, camX, camY float64) {
	ms := p.mapState

	entityColors := map[string]color.RGBA{
		"spawn":   {R: 0x00, G: 0xff, B: 0x00, A: 0xcc},
		"trigger": {R: 0xff, G: 0xff, B: 0x00, A: 0xcc},
		"enemy":   {R: 0xff, G: 0x00, B: 0x00, A: 0xcc},
	}
	defaultColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xcc}

	for _, e := range layer.Entities {
		sx := float64(e.X*ts)*z - camX
		sy := float64(e.Y*ts)*z - camY

		// Try to render sprite from properties.
		drawn := false
		if ref, ok := e.Properties["sprite"]; ok {
			if s, ok := ref.(string); ok {
				if img, ok := ms.entityImages[s]; ok {
					op := &ebiten.DrawImageOptions{}
					imgW := float64(img.Bounds().Dx())
					imgH := float64(img.Bounds().Dy())
					op.GeoM.Scale(float64(ts)/imgW*z, float64(ts)/imgH*z)
					op.GeoM.Translate(sx, sy)
					op.Filter = ebiten.FilterNearest
					screen.DrawImage(img, op)
					drawn = true
				}
			}
		}

		if !drawn {
			// Fallback: colored rectangle.
			size := float64(ts) * z * 0.6
			c, ok := entityColors[e.Type]
			if !ok {
				c = defaultColor
			}
			ox := sx + (float64(ts)*z-size)/2
			oy := sy + (float64(ts)*z-size)/2
			for dy := 0; dy < int(size); dy++ {
				for dx := 0; dx < int(size); dx++ {
					px, py := int(ox)+dx, int(oy)+dy
					if px >= 0 && px < p.winW && py >= 0 && py < p.winH {
						screen.Set(px, py, c)
					}
				}
			}
		}

		// Label.
		drawText(screen, e.Type, int(sx), int(sy)-scaledCharH()-2)
	}
}

func (p *Previewer) drawMapGrid(screen *ebiten.Image, mf *tilemap.MapFile, ts int, z, camX, camY float64) {
	gc := color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0x40}
	cellSize := float64(ts) * z

	// Find map dimensions.
	mapW, mapH := 0, 0
	for _, l := range mf.Layers {
		if l.Type == "tile" && len(l.Data) > 0 {
			if len(l.Data) > mapH {
				mapH = len(l.Data)
			}
			if len(l.Data[0]) > mapW {
				mapW = len(l.Data[0])
			}
		}
	}

	// Vertical lines.
	for col := 0; col <= mapW; col++ {
		x := int(float64(col)*cellSize - camX)
		if x >= 0 && x < p.winW {
			for y := 0; y < p.winH; y++ {
				screen.Set(x, y, gc)
			}
		}
	}
	// Horizontal lines.
	for row := 0; row <= mapH; row++ {
		y := int(float64(row)*cellSize - camY)
		if y >= 0 && y < p.winH {
			for x := 0; x < p.winW; x++ {
				screen.Set(x, y, gc)
			}
		}
	}
}
