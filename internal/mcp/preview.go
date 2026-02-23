package mcp

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"

	"github.com/vgalaktionov/runefact/internal/palette"
	"github.com/vgalaktionov/runefact/internal/sprite"
	"github.com/vgalaktionov/runefact/internal/tilemap"
)

// handlePreviewMap renders a map file to a PNG and returns it as inline image content.
func (ctx *ServerContext) handlePreviewMap(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	scale := int(req.GetFloat("scale", 2))
	if scale < 1 {
		scale = 1
	}
	if scale > 8 {
		scale = 8
	}

	assetsDir := filepath.Join(ctx.ProjectRoot, "assets")

	// Load map.
	mapPath := filepath.Join(assetsDir, "maps", file)
	mf, _, err := tilemap.LoadMapFile(mapPath)
	if err != nil {
		return errorResult(fmt.Sprintf("loading %s: %v", file, err))
	}

	// Find map dimensions from tile layers.
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

	if mapW == 0 || mapH == 0 {
		return errorResult("map has no tile data")
	}

	ts := mf.TileSize

	// Load tile and entity sprites.
	spriteLoader := newSpriteLoader(assetsDir)
	tileImages := spriteLoader.loadTileSprites(mf)
	entityImages := spriteLoader.loadEntitySprites(mf)

	// Create output image.
	imgW := mapW * ts * scale
	imgH := mapH * ts * scale
	img := image.NewRGBA(image.Rect(0, 0, imgW, imgH))

	// Render tile layers bottom-up.
	for _, layer := range mf.Layers {
		if layer.Type != "tile" {
			continue
		}
		for y, row := range layer.Data {
			for x, tileID := range row {
				if tileID == 0 {
					continue
				}
				tileImg, ok := tileImages[tileID]
				if !ok {
					continue
				}
				drawScaled(img, tileImg, x*ts*scale, y*ts*scale, ts, scale)
			}
		}
	}

	// Render entity layers.
	for _, layer := range mf.Layers {
		if layer.Type != "entity" {
			continue
		}
		for _, e := range layer.Entities {
			ref, _ := e.Properties["sprite"].(string)
			if eImg, ok := entityImages[ref]; ok {
				drawScaled(img, eImg, e.X*ts*scale, e.Y*ts*scale, ts, scale)
			} else {
				// Fallback: colored marker.
				drawEntityMarker(img, e.Type, e.X*ts*scale, e.Y*ts*scale, ts*scale)
			}
		}
	}

	return imageResult(img)
}

// handlePreviewSprite renders a sprite file to a PNG grid and returns it inline.
func (ctx *ServerContext) handlePreviewSprite(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	file, err := req.RequireString("file")
	if err != nil {
		return errorResult("file parameter required")
	}

	scale := int(req.GetFloat("scale", 4))
	if scale < 1 {
		scale = 1
	}
	if scale > 16 {
		scale = 16
	}

	assetsDir := filepath.Join(ctx.ProjectRoot, "assets")

	// Load and resolve sprite file.
	spritePath := filepath.Join(assetsDir, "sprites", file)
	sf, err := sprite.LoadSpriteFile(spritePath)
	if err != nil {
		return errorResult(fmt.Sprintf("loading %s: %v", file, err))
	}

	var pal *palette.Palette
	if sf.PaletteRef != "" {
		palPath := filepath.Join(assetsDir, "palettes", sf.PaletteRef+".palette")
		pal, _ = palette.LoadPalette(palPath)
	}
	if pal == nil {
		pal = &palette.Palette{Colors: map[string]palette.Color{}}
	}

	resolved, err := sf.Resolve(pal)
	if err != nil {
		return errorResult(fmt.Sprintf("resolving %s: %v", file, err))
	}

	if len(resolved) == 0 {
		return errorResult("sprite file has no sprites")
	}

	// Layout: each sprite on its own row, frames laid out horizontally.
	// 1px gap between frames, 1px gap between sprite rows.
	gap := 1
	maxWidth := 0
	totalHeight := 0
	for _, rs := range resolved {
		rowW := len(rs.Frames)*(rs.Grid.W*scale+gap) - gap
		if rowW > maxWidth {
			maxWidth = rowW
		}
		totalHeight += rs.Grid.H*scale + gap
	}
	totalHeight -= gap

	// Checkerboard background for transparency.
	img := image.NewRGBA(image.Rect(0, 0, maxWidth, totalHeight))
	drawCheckerboard(img)

	// Render each sprite's frames.
	curY := 0
	for _, rs := range resolved {
		curX := 0
		for _, frame := range rs.Frames {
			frameImg := renderFrame(frame.Pixels, rs.Grid.W, rs.Grid.H)
			drawScaledDirect(img, frameImg, curX, curY, scale)
			curX += rs.Grid.W*scale + gap
		}
		curY += rs.Grid.H*scale + gap
	}

	return imageResult(img)
}

// spriteLoader caches loaded sprite files for reuse across tile and entity loading.
type spriteLoader struct {
	assetsDir string
	cache     map[string][]sprite.ResolvedSprite
}

func newSpriteLoader(assetsDir string) *spriteLoader {
	return &spriteLoader{
		assetsDir: assetsDir,
		cache:     make(map[string][]sprite.ResolvedSprite),
	}
}

func (sl *spriteLoader) resolve(fileName string) []sprite.ResolvedSprite {
	if resolved, ok := sl.cache[fileName]; ok {
		return resolved
	}

	spritePath := filepath.Join(sl.assetsDir, "sprites", fileName+".sprite")
	sf, err := sprite.LoadSpriteFile(spritePath)
	if err != nil {
		sl.cache[fileName] = nil
		return nil
	}

	var pal *palette.Palette
	if sf.PaletteRef != "" {
		palPath := filepath.Join(sl.assetsDir, "palettes", sf.PaletteRef+".palette")
		pal, _ = palette.LoadPalette(palPath)
	}
	if pal == nil {
		pal = &palette.Palette{Colors: map[string]palette.Color{}}
	}

	resolved, err := sf.Resolve(pal)
	if err != nil {
		sl.cache[fileName] = nil
		return nil
	}
	sl.cache[fileName] = resolved
	return resolved
}

func (sl *spriteLoader) findSprite(ref string) *image.RGBA {
	parts := strings.SplitN(ref, ":", 2)
	if len(parts) != 2 {
		return nil
	}
	fileName, spriteName := parts[0], parts[1]

	resolved := sl.resolve(fileName)
	for _, rs := range resolved {
		if rs.Name == spriteName && len(rs.Frames) > 0 {
			return renderFrame(rs.Frames[0].Pixels, rs.Grid.W, rs.Grid.H)
		}
	}
	return nil
}

func (sl *spriteLoader) loadTileSprites(mf *tilemap.MapFile) map[int]*image.RGBA {
	images := make(map[int]*image.RGBA)

	// Rebuild tile index.
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

	for key, ref := range mf.Tileset {
		if ref == "" {
			continue
		}
		id := tileIndex[key]
		if img := sl.findSprite(ref); img != nil {
			images[id] = img
		}
	}

	return images
}

func (sl *spriteLoader) loadEntitySprites(mf *tilemap.MapFile) map[string]*image.RGBA {
	images := make(map[string]*image.RGBA)

	for _, layer := range mf.Layers {
		if layer.Type != "entity" {
			continue
		}
		for _, e := range layer.Entities {
			ref, ok := e.Properties["sprite"].(string)
			if !ok || ref == "" {
				continue
			}
			if _, loaded := images[ref]; loaded {
				continue
			}
			if img := sl.findSprite(ref); img != nil {
				images[ref] = img
			}
		}
	}

	return images
}

// renderFrame converts resolved pixel data to an image.RGBA.
func renderFrame(pixels [][]palette.Color, w, h int) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y, row := range pixels {
		for x, c := range row {
			img.Set(x, y, c.ToRGBA())
		}
	}
	return img
}

// drawScaled draws src scaled to fit a tile_size*scale cell at (dx, dy) using nearest-neighbor.
func drawScaled(dst *image.RGBA, src *image.RGBA, dx, dy, tileSize, scale int) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()
	dstSize := tileSize * scale

	for sy := 0; sy < srcH; sy++ {
		for sx := 0; sx < srcW; sx++ {
			c := src.RGBAAt(sx, sy)
			if c.A == 0 {
				continue
			}
			// Scale pixel position to destination.
			x0 := dx + sx*dstSize/srcW
			y0 := dy + sy*dstSize/srcH
			x1 := dx + (sx+1)*dstSize/srcW
			y1 := dy + (sy+1)*dstSize/srcH
			for py := y0; py < y1; py++ {
				for px := x0; px < x1; px++ {
					if px >= 0 && px < dst.Bounds().Dx() && py >= 0 && py < dst.Bounds().Dy() {
						dst.SetRGBA(px, py, c)
					}
				}
			}
		}
	}
}

// drawScaledDirect draws src with a simple integer scale factor at (dx, dy).
func drawScaledDirect(dst *image.RGBA, src *image.RGBA, dx, dy, scale int) {
	srcW := src.Bounds().Dx()
	srcH := src.Bounds().Dy()

	for sy := 0; sy < srcH; sy++ {
		for sx := 0; sx < srcW; sx++ {
			c := src.RGBAAt(sx, sy)
			if c.A == 0 {
				continue
			}
			for py := 0; py < scale; py++ {
				for px := 0; px < scale; px++ {
					x := dx + sx*scale + px
					y := dy + sy*scale + py
					if x >= 0 && x < dst.Bounds().Dx() && y >= 0 && y < dst.Bounds().Dy() {
						dst.SetRGBA(x, y, c)
					}
				}
			}
		}
	}
}

// drawEntityMarker draws a small colored diamond for entities without sprites.
func drawEntityMarker(img *image.RGBA, entityType string, dx, dy, size int) {
	var c color.RGBA
	switch entityType {
	case "spawn":
		c = color.RGBA{R: 0x00, G: 0xff, B: 0x00, A: 0xcc}
	case "enemy":
		c = color.RGBA{R: 0xff, G: 0x00, B: 0x00, A: 0xcc}
	default:
		c = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xcc}
	}

	// Draw filled diamond.
	half := size / 2
	for py := 0; py < size; py++ {
		for px := 0; px < size; px++ {
			dist := abs(px-half) + abs(py-half)
			if dist <= half {
				x, y := dx+px, dy+py
				if x >= 0 && x < img.Bounds().Dx() && y >= 0 && y < img.Bounds().Dy() {
					img.SetRGBA(x, y, c)
				}
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// drawCheckerboard fills an image with a transparency checkerboard pattern.
func drawCheckerboard(img *image.RGBA) {
	light := color.RGBA{R: 0xcc, G: 0xcc, B: 0xcc, A: 0xff}
	dark := color.RGBA{R: 0x99, G: 0x99, B: 0x99, A: 0xff}
	checkSize := 8

	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			if ((x/checkSize)+(y/checkSize))%2 == 0 {
				img.SetRGBA(x, y, light)
			} else {
				img.SetRGBA(x, y, dark)
			}
		}
	}
}

// imageResult encodes an image as PNG and returns it as MCP ImageContent.
func imageResult(img image.Image) (*mcp.CallToolResult, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return errorResult(fmt.Sprintf("encoding PNG: %v", err))
	}

	b64 := base64.StdEncoding.EncodeToString(buf.Bytes())

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.ImageContent{
				Type:     "image",
				Data:     b64,
				MIMEType: "image/png",
			},
		},
	}, nil
}
