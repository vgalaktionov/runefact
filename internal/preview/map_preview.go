package preview

import (
	"fmt"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

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
}

func (p *Previewer) initMapState(mf *tilemap.MapFile) {
	tileLayerCount := 0
	for _, l := range mf.Layers {
		if l.Type == "tile" {
			tileLayerCount++
		}
	}
	p.mapState = &MapPreviewState{
		mapFile:    mf,
		mapZoom:    2.0,
		layerCount: tileLayerCount,
	}
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
	ebitenutil.DebugPrintAt(screen, label, 10, 10)
}

func (p *Previewer) drawTileLayer(screen *ebiten.Image, layer tilemap.Layer, ts int, z, camX, camY float64) {
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

			// Draw a colored rectangle for each tile type.
			sx := float64(x*ts)*z - camX
			sy := float64(y*ts)*z - camY
			w := float64(ts) * z
			h := float64(ts) * z

			// Color based on tile ID.
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

func (p *Previewer) drawEntityLayer(screen *ebiten.Image, layer tilemap.Layer, ts int, z, camX, camY float64) {
	entityColors := map[string]color.RGBA{
		"spawn":   {R: 0x00, G: 0xff, B: 0x00, A: 0xcc},
		"trigger": {R: 0xff, G: 0xff, B: 0x00, A: 0xcc},
		"enemy":   {R: 0xff, G: 0x00, B: 0x00, A: 0xcc},
	}
	defaultColor := color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xcc}

	for _, e := range layer.Entities {
		sx := float64(e.X*ts)*z - camX
		sy := float64(e.Y*ts)*z - camY
		size := float64(ts) * z * 0.6

		c, ok := entityColors[e.Type]
		if !ok {
			c = defaultColor
		}

		// Draw marker rectangle.
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

		// Label.
		ebitenutil.DebugPrintAt(screen, e.Type, int(sx), int(sy)-12)
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
