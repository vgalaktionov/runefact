package preview

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/vgalaktionov/runefact/internal/palette"
	"github.com/vgalaktionov/runefact/internal/sfx"
	"github.com/vgalaktionov/runefact/internal/sprite"
	"github.com/vgalaktionov/runefact/internal/tilemap"
	"github.com/vgalaktionov/runefact/internal/track"
	"github.com/vgalaktionov/runefact/internal/watcher"
)

// BackgroundType enumerates preview background options.
type BackgroundType int

const (
	BackgroundDark BackgroundType = iota
	BackgroundLight
	BackgroundCheckerboard
)

// PreviewMode identifies what kind of asset is being previewed.
type PreviewMode int

const (
	ModeSpritePreview PreviewMode = iota
	ModeMapPreview
	ModeSFXPreview
	ModeMusicPreview
)

// RenderedSprite holds an ebitengine image for each frame of a sprite.
type RenderedSprite struct {
	Name       string
	Frames     []*ebiten.Image
	FrameW     int
	FrameH     int
	FPS        int
	FrameCount int
}

// Previewer implements ebiten.Game for live asset preview.
type Previewer struct {
	mode       PreviewMode
	background BackgroundType
	errorMsg   string
	winW, winH int
	filePath   string
	assetsDir  string
	sampleRate int

	// Sprite mode state.
	sprites   []*RenderedSprite
	zoom      int
	paused    bool
	frameTime float64
	selected  int // -1 = grid view, >= 0 = isolated sprite
	showGrid  bool

	// Map mode state.
	mapState *MapPreviewState

	// SFX mode state.
	sfxState *SFXPreviewState

	// Music mode state.
	musicState *MusicPreviewState

	// File watching.
	watcher     *watcher.Watcher
	reloadMu    sync.Mutex
	pendingLoad []*RenderedSprite
	pendingErr  string
}

// NewPreviewer creates a previewer for the given file.
func NewPreviewer(filePath, assetsDir string, winW, winH, sampleRate int) *Previewer {
	ext := filepath.Ext(filePath)
	mode := ModeSpritePreview
	switch ext {
	case ".sprite":
		mode = ModeSpritePreview
	case ".map":
		mode = ModeMapPreview
	case ".sfx":
		mode = ModeSFXPreview
	case ".track":
		mode = ModeMusicPreview
	}

	return &Previewer{
		mode:       mode,
		zoom:       2,
		selected:   -1,
		winW:       winW,
		winH:       winH,
		filePath:   filePath,
		assetsDir:  assetsDir,
		sampleRate: sampleRate,
	}
}

// Run starts the ebitengine window and event loop.
func (p *Previewer) Run() error {
	// Load state.
	if st := loadState(); st != nil {
		if st.Zoom > 0 {
			p.zoom = st.Zoom
		}
		p.background = BackgroundType(st.Background)
		p.showGrid = st.ShowGrid
	}

	// Initial load based on mode.
	if err := p.loadAsset(); err != nil {
		p.errorMsg = err.Error()
	}

	// Start file watcher.
	p.startWatcher()

	ebiten.SetWindowSize(p.winW, p.winH)
	ebiten.SetWindowTitle("Runefact Preview")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	defer p.stopWatcher()
	defer p.saveState()

	return ebiten.RunGame(p)
}

func (p *Previewer) loadAsset() error {
	switch p.mode {
	case ModeSpritePreview:
		sprites, err := p.loadSprites()
		if err != nil {
			return err
		}
		p.sprites = sprites
	case ModeMapPreview:
		mf, _, err := tilemap.LoadMapFile(p.filePath)
		if err != nil {
			return err
		}
		p.initMapState(mf)
	case ModeSFXPreview:
		s, err := sfx.LoadSFX(p.filePath)
		if err != nil {
			return err
		}
		sr := p.sampleRate
		if sr == 0 {
			sr = 44100
		}
		p.initSFXState(s, sr)
	case ModeMusicPreview:
		tr, err := track.LoadTrack(p.filePath)
		if err != nil {
			return err
		}
		p.initMusicState(tr)
	}
	return nil
}

// Update handles input and animation.
func (p *Previewer) Update() error {
	// Check for pending reload (sprite mode).
	p.reloadMu.Lock()
	if p.pendingLoad != nil {
		p.sprites = p.pendingLoad
		p.errorMsg = ""
		p.pendingLoad = nil
	}
	if p.pendingErr != "" {
		p.errorMsg = p.pendingErr
		p.pendingErr = ""
	}
	p.reloadMu.Unlock()

	// B: cycle background (all modes).
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		p.background = (p.background + 1) % 3
	}

	switch p.mode {
	case ModeSpritePreview:
		p.updateSprite()
	case ModeMapPreview:
		p.updateMap()
	case ModeSFXPreview:
		p.updateSFX()
	case ModeMusicPreview:
		p.updateMusic()
	}

	return nil
}

func (p *Previewer) updateSprite() {
	// Zoom: mouse wheel.
	_, dy := ebiten.Wheel()
	if dy > 0 {
		p.zoom = min(32, p.zoom*2)
	} else if dy < 0 {
		p.zoom = max(1, p.zoom/2)
	}

	// Space: pause/resume.
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		p.paused = !p.paused
	}

	// Frame stepping when paused.
	if p.paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			for _, s := range p.sprites {
				if s.FrameCount > 1 && s.FPS > 0 {
					p.frameTime += 1.0 / float64(s.FPS)
				}
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			for _, s := range p.sprites {
				if s.FrameCount > 1 && s.FPS > 0 {
					p.frameTime -= 1.0 / float64(s.FPS)
					if p.frameTime < 0 {
						p.frameTime = 0
					}
				}
			}
		}
	}

	// G: toggle grid.
	if inpututil.IsKeyJustPressed(ebiten.KeyG) {
		p.showGrid = !p.showGrid
	}

	// Escape: back to grid from isolation.
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		p.selected = -1
	}

	// Click: isolate/deselect sprite.
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && p.selected == -1 {
		mx, my := ebiten.CursorPosition()
		if idx := p.hitTestSprite(mx, my); idx >= 0 {
			p.selected = idx
		}
	} else if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && p.selected >= 0 {
		p.selected = -1
	}

	// Advance animation.
	if !p.paused {
		p.frameTime += 1.0 / float64(ebiten.TPS())
	}
}

// Draw renders the current frame.
func (p *Previewer) Draw(screen *ebiten.Image) {
	p.drawBackground(screen)

	switch p.mode {
	case ModeSpritePreview:
		p.drawSpriteMode(screen)
	case ModeMapPreview:
		p.drawMap(screen)
	case ModeSFXPreview:
		p.drawSFX(screen)
	case ModeMusicPreview:
		p.drawMusic(screen)
	}

	// Error overlay.
	if p.errorMsg != "" {
		p.drawErrorOverlay(screen)
	}
}

func (p *Previewer) drawSpriteMode(screen *ebiten.Image) {
	if len(p.sprites) == 0 && p.errorMsg == "" {
		drawText(screen, "No sprites loaded", 10, 10)
		return
	}

	if p.selected >= 0 && p.selected < len(p.sprites) {
		p.drawIsolated(screen, p.sprites[p.selected])
	} else {
		p.drawSpriteGrid(screen)
	}

	if p.showGrid && len(p.sprites) > 0 {
		p.drawPixelGrid(screen)
	}

	if p.paused {
		drawText(screen, "PAUSED", 10, p.winH-scaledCharH()-6)
	}
}

// Layout returns the internal resolution.
func (p *Previewer) Layout(outsideWidth, outsideHeight int) (int, int) {
	p.winW = outsideWidth
	p.winH = outsideHeight
	return outsideWidth, outsideHeight
}

// loadSprites parses and resolves sprites from the file path.
func (p *Previewer) loadSprites() ([]*RenderedSprite, error) {
	sf, err := sprite.LoadSpriteFile(p.filePath)
	if err != nil {
		return nil, err
	}

	// Resolve palette.
	var pal *palette.Palette
	if sf.PaletteRef != "" {
		palPath := filepath.Join(p.assetsDir, "palettes", sf.PaletteRef+".palette")
		pal, err = palette.LoadPalette(palPath)
		if err != nil {
			return nil, fmt.Errorf("loading palette %q: %w", sf.PaletteRef, err)
		}
	}
	if pal == nil {
		pal = &palette.Palette{Colors: map[string]palette.Color{}}
	}

	resolved, err := sf.Resolve(pal)
	if err != nil {
		return nil, err
	}

	var result []*RenderedSprite
	for _, rs := range resolved {
		rendered := &RenderedSprite{
			Name:       rs.Name,
			FrameW:     rs.Grid.W,
			FrameH:     rs.Grid.H,
			FPS:        rs.Framerate,
			FrameCount: len(rs.Frames),
		}

		for _, frame := range rs.Frames {
			img := ebiten.NewImageFromImage(frameToImage(frame, rs.Grid))
			rendered.Frames = append(rendered.Frames, img)
		}
		result = append(result, rendered)
	}
	return result, nil
}

func frameToImage(frame sprite.ResolvedFrame, grid sprite.Grid) *image.RGBA {
	img := image.NewRGBA(image.Rect(0, 0, grid.W, grid.H))
	for y, row := range frame.Pixels {
		for x, c := range row {
			img.Set(x, y, c.ToRGBA())
		}
	}
	return img
}

// drawBackground fills the screen with the selected background.
func (p *Previewer) drawBackground(screen *ebiten.Image) {
	switch p.background {
	case BackgroundDark:
		screen.Fill(color.RGBA{R: 0x1a, G: 0x1a, B: 0x1a, A: 0xff})
	case BackgroundLight:
		screen.Fill(color.RGBA{R: 0xe0, G: 0xe0, B: 0xe0, A: 0xff})
	case BackgroundCheckerboard:
		screen.Fill(color.RGBA{R: 0x40, G: 0x40, B: 0x40, A: 0xff})
		w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
		light := color.RGBA{R: 0x50, G: 0x50, B: 0x50, A: 0xff}
		for cy := 0; cy < h; cy += 8 {
			for cx := 0; cx < w; cx += 8 {
				if ((cx/8)+(cy/8))%2 == 0 {
					for dy := 0; dy < 8 && cy+dy < h; dy++ {
						for dx := 0; dx < 8 && cx+dx < w; dx++ {
							screen.Set(cx+dx, cy+dy, light)
						}
					}
				}
			}
		}
	}
}

// spriteGridLayout computes the auto-zoom grid layout for the current sprites/window.
func (p *Previewer) spriteGridLayout() (z float64, cols, padding, cellW, cellH, offsetX, offsetY int) {
	labelH := scaledCharH() + 6
	padding = 24

	maxW, maxH := 0, 0
	maxLabelW := 0
	for _, s := range p.sprites {
		if s.FrameW > maxW {
			maxW = s.FrameW
		}
		if s.FrameH > maxH {
			maxH = s.FrameH
		}
		// Estimate label width for this sprite.
		lbl := fmt.Sprintf("%s %dx%d", s.Name, s.FrameW, s.FrameH)
		if s.FrameCount > 1 {
			lbl += fmt.Sprintf(" f:%d/%d @%dfps", s.FrameCount, s.FrameCount, s.FPS)
		}
		if w := len(lbl) * scaledCharW(); w > maxLabelW {
			maxLabelW = w
		}
	}

	// Auto-calculate zoom. For each column count, find the max zoom
	// where both the scaled sprites AND labels fit without overlap.
	n := len(p.sprites)
	bestZoom := 1.0
	bestCols := 1
	for c := 1; c <= n; c++ {
		rows := (n + c - 1) / c
		// Available width per cell = window / cols, minus padding.
		availW := float64(p.winW)/float64(c) - float64(padding*2)
		availH := float64(p.winH)/float64(rows) - float64(padding) - float64(labelH)

		zx := availW / float64(maxW)
		zy := availH / float64(maxH)
		zFit := min(zx, zy)

		// Cell must also be wide enough for the label text.
		minCellW := maxLabelW + padding*2
		maxZoomForLabel := (float64(p.winW)/float64(c) - float64(minCellW-int(float64(maxW)))) / float64(maxW)
		if maxZoomForLabel < zFit {
			zFit = maxZoomForLabel
		}

		if zFit > bestZoom {
			bestZoom = zFit
			bestCols = c
		}
	}
	// Snap to integer zoom for pixel-perfect rendering.
	z = float64(max(1, int(bestZoom)))
	cols = bestCols

	spriteW := int(float64(maxW) * z)
	cellW = max(spriteW, maxLabelW) + padding*2
	cellH = int(float64(maxH)*z) + padding + labelH

	usedCols := min(cols, n)
	totalW := usedCols * cellW
	rows := (n + cols - 1) / cols
	totalH := rows * cellH
	offsetX = (p.winW - totalW) / 2
	offsetY = (p.winH - totalH) / 2
	if offsetY < padding {
		offsetY = padding
	}
	return
}

// drawSpriteGrid draws all sprites in a grid layout.
func (p *Previewer) drawSpriteGrid(screen *ebiten.Image) {
	if len(p.sprites) == 0 {
		return
	}

	z, cols, padding, cellW, cellH, offsetX, offsetY := p.spriteGridLayout()

	maxW, maxH := 0, 0
	for _, s := range p.sprites {
		if s.FrameW > maxW {
			maxW = s.FrameW
		}
		if s.FrameH > maxH {
			maxH = s.FrameH
		}
	}

	for i, s := range p.sprites {
		col := i % cols
		row := i / cols
		cx := offsetX + col*cellW + padding
		cy := offsetY + row*cellH

		frame := p.currentFrame(s)
		if frame < len(s.Frames) {
			// Center sprite within the cell.
			spriteW := int(float64(s.FrameW) * z)
			spriteH := int(float64(s.FrameH) * z)
			sx := cx + (cellW-padding*2-spriteW)/2
			sy := cy + (int(float64(maxH)*z)-spriteH)/2

			op := &ebiten.DrawImageOptions{}
			op.GeoM.Scale(z, z)
			op.GeoM.Translate(float64(sx), float64(sy))
			op.Filter = ebiten.FilterNearest
			screen.DrawImage(s.Frames[frame], op)
		}

		label := fmt.Sprintf("%s %dx%d", s.Name, s.FrameW, s.FrameH)
		if s.FrameCount > 1 {
			label += fmt.Sprintf(" f:%d/%d @%dfps", frame+1, s.FrameCount, s.FPS)
		}
		// Center label under sprite.
		labelW := len(label) * scaledCharW()
		labelX := cx + (cellW-padding*2-labelW)/2
		drawText(screen, label, labelX, cy+int(float64(maxH)*z)+4)
	}
}

// drawIsolated draws a single sprite centered in the window.
func (p *Previewer) drawIsolated(screen *ebiten.Image, s *RenderedSprite) {
	z := float64(p.zoom)
	frame := p.currentFrame(s)
	if frame >= len(s.Frames) {
		return
	}

	sw := float64(s.FrameW) * z
	sh := float64(s.FrameH) * z
	cx := (float64(p.winW) - sw) / 2
	cy := (float64(p.winH) - sh) / 2

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(z, z)
	op.GeoM.Translate(cx, cy)
	op.Filter = ebiten.FilterNearest
	screen.DrawImage(s.Frames[frame], op)

	label := fmt.Sprintf("%s %dx%d", s.Name, s.FrameW, s.FrameH)
	if s.FrameCount > 1 {
		label += fmt.Sprintf(" f:%d/%d @%dfps", frame+1, s.FrameCount, s.FPS)
	}
	drawText(screen, label, 10, 10)
}

// drawPixelGrid overlays 1px grid lines at pixel boundaries.
func (p *Previewer) drawPixelGrid(screen *ebiten.Image) {
	if p.zoom < 4 {
		return
	}
	gridColor := color.RGBA{R: 0x80, G: 0x80, B: 0x80, A: 0x40}
	z := float64(p.zoom)

	for x := 0.0; x < float64(p.winW); x += z {
		for y := 0; y < p.winH; y++ {
			screen.Set(int(x), y, gridColor)
		}
	}
	for y := 0.0; y < float64(p.winH); y += z {
		for x := 0; x < p.winW; x++ {
			screen.Set(x, int(y), gridColor)
		}
	}
}

// drawErrorOverlay renders a semi-transparent red box with error text.
func (p *Previewer) drawErrorOverlay(screen *ebiten.Image) {
	boxH := scaledCharH() + 16
	for y := 0; y < boxH; y++ {
		for x := 0; x < p.winW; x++ {
			screen.Set(x, y, color.RGBA{R: 0xcc, G: 0x22, B: 0x22, A: 0xdd})
		}
	}
	msg := p.errorMsg
	if len(msg) > 100 {
		msg = msg[:100] + "..."
	}
	drawText(screen, "ERROR: "+msg, 10, 8)
}

// currentFrame computes the current animation frame index.
func (p *Previewer) currentFrame(s *RenderedSprite) int {
	if s.FrameCount <= 1 || s.FPS <= 0 {
		return 0
	}
	frame := int(p.frameTime*float64(s.FPS)) % s.FrameCount
	if frame < 0 {
		frame += s.FrameCount
	}
	return frame
}

// hitTestSprite returns the sprite index at screen position, or -1.
func (p *Previewer) hitTestSprite(mx, my int) int {
	if len(p.sprites) == 0 {
		return -1
	}

	z, cols, padding, cellW, cellH, offsetX, offsetY := p.spriteGridLayout()

	for i, s := range p.sprites {
		col := i % cols
		row := i / cols
		cx := offsetX + col*cellW + padding
		cy := offsetY + row*cellH
		sw := int(float64(s.FrameW) * z)
		sh := int(float64(s.FrameH) * z)

		// Center sprite within cell.
		sx := cx + (cellW-padding*2-sw)/2

		if mx >= sx && mx < sx+sw && my >= cy && my < cy+sh {
			return i
		}
	}
	return -1
}

// startWatcher starts file watching for live reload.
func (p *Previewer) startWatcher() {
	dir := filepath.Dir(p.filePath)
	w, err := watcher.New(100*time.Millisecond, func(changed []string) error {
		for _, f := range changed {
			if f == p.filePath || strings.HasSuffix(f, ".palette") {
				if p.mode == ModeSpritePreview {
					sprites, err := p.loadSprites()
					p.reloadMu.Lock()
					if err != nil {
						p.pendingErr = err.Error()
					} else {
						p.pendingLoad = sprites
						p.pendingErr = ""
					}
					p.reloadMu.Unlock()
				} else {
					err := p.loadAsset()
					p.reloadMu.Lock()
					if err != nil {
						p.pendingErr = err.Error()
					} else {
						p.pendingErr = ""
					}
					p.reloadMu.Unlock()
				}
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return
	}

	_ = w.WatchDir(dir)

	if p.assetsDir != "" {
		palDir := filepath.Join(p.assetsDir, "palettes")
		if info, err := os.Stat(palDir); err == nil && info.IsDir() {
			_ = w.WatchDir(palDir)
		}
	}

	p.watcher = w
	go w.Start()
}

func (p *Previewer) stopWatcher() {
	if p.watcher != nil {
		_ = p.watcher.Stop()
	}
}

// State persistence.

type previewState struct {
	Zoom       int    `json:"zoom"`
	Background int    `json:"background"`
	ShowGrid   bool   `json:"show_grid"`
	LastFile   string `json:"last_file"`
}

func stateFilePath() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "runefact", "preview.json")
}

func loadState() *previewState {
	path := stateFilePath()
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var st previewState
	if json.Unmarshal(data, &st) != nil {
		return nil
	}
	return &st
}

func (p *Previewer) saveState() {
	path := stateFilePath()
	if path == "" {
		return
	}
	st := previewState{
		Zoom:       p.zoom,
		Background: int(p.background),
		ShowGrid:   p.showGrid,
		LastFile:   p.filePath,
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	_ = os.WriteFile(path, data, 0644)
}
