# Ebitengine Integration

How to load and use Runefact-built artifacts in an [ebitengine](https://ebitengine.org) game.

## Build Output

After `runefact build`, the output directory (default `build/assets/`) contains:

```
build/assets/
  manifest.go           # type-safe Go asset loader
  sprites/
    player.png          # packed sprite sheets
    terrain.png
  maps/
    world.json          # tilemap data
  audio/
    jump.wav            # sound effects
    bgm.wav             # music
```

## Manifest

`manifest.go` provides constants and metadata for all built assets:

```go
package assets

// Sprite sheet paths
const SpriteSheetPlayer = "sprites/player.png"
const SpriteSheetTerrain = "sprites/terrain.png"

// Map paths
const MapWorld = "maps/world.json"

// Audio paths
const AudioJump = "audio/jump.wav"
const AudioBGM = "audio/bgm.wav"

// SpriteInfo contains metadata for each sprite
type SpriteInfo struct {
    Sheet    string
    X, Y     int
    W, H     int
    Frames   int
    Framerate int
}

// Sprites maps "file:name" to sprite metadata
var Sprites = map[string]SpriteInfo{
    "player:idle": {Sheet: SpriteSheetPlayer, X: 0, Y: 0, W: 16, H: 16, Frames: 1},
    "player:walk": {Sheet: SpriteSheetPlayer, X: 16, Y: 0, W: 16, H: 16, Frames: 4, Framerate: 8},
    // ...
}
```

## Loading Sprites

```go
package main

import (
    "image"
    _ "image/png"
    "os"

    "github.com/hajimehoshi/ebiten/v2"
    "your-game/build/assets"
)

func loadSpriteSheet(path string) *ebiten.Image {
    f, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    defer f.Close()

    img, _, err := image.Decode(f)
    if err != nil {
        panic(err)
    }
    return ebiten.NewImageFromImage(img)
}

// Get a sub-image for a specific sprite
func spriteImage(sheet *ebiten.Image, info assets.SpriteInfo, frame int) *ebiten.Image {
    x := info.X + frame*info.W
    rect := image.Rect(x, info.Y, x+info.W, info.Y+info.H)
    return sheet.SubImage(rect).(*ebiten.Image)
}
```

## Animation

```go
type AnimatedSprite struct {
    sheet    *ebiten.Image
    info     assets.SpriteInfo
    elapsed  float64
    frame    int
}

func (a *AnimatedSprite) Update(dt float64) {
    if a.info.Frames <= 1 || a.info.Framerate <= 0 {
        return
    }
    a.elapsed += dt
    frameDuration := 1.0 / float64(a.info.Framerate)
    for a.elapsed >= frameDuration {
        a.elapsed -= frameDuration
        a.frame = (a.frame + 1) % a.info.Frames
    }
}

func (a *AnimatedSprite) Draw(screen *ebiten.Image, x, y float64) {
    img := spriteImage(a.sheet, a.info, a.frame)
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(x, y)
    screen.DrawImage(img, op)
}
```

## Loading Maps

Map JSON output structure:

```json
{
    "tile_size": 16,
    "layers": [
        {
            "name": "ground",
            "type": "tile",
            "scroll_x": 0,
            "scroll_y": 0,
            "data": [[0, 1, 1, 0], [1, 1, 1, 1]],
            "width": 4,
            "height": 2
        },
        {
            "name": "objects",
            "type": "entity",
            "entities": [
                {"type": "spawn", "x": 32, "y": 48},
                {"type": "chest", "x": 96, "y": 48, "properties": {"locked": true}}
            ]
        }
    ],
    "tileset": {
        "0": "",
        "1": "terrain:grass"
    }
}
```

```go
import (
    "encoding/json"
    "os"
)

type TileMap struct {
    TileSize int              `json:"tile_size"`
    Layers   []MapLayer       `json:"layers"`
    Tileset  map[string]string `json:"tileset"`
}

type MapLayer struct {
    Name     string     `json:"name"`
    Type     string     `json:"type"`
    ScrollX  float64    `json:"scroll_x"`
    ScrollY  float64    `json:"scroll_y"`
    Data     [][]int    `json:"data"`
    Width    int        `json:"width"`
    Height   int        `json:"height"`
    Entities []Entity   `json:"entities"`
}

type Entity struct {
    Type       string                 `json:"type"`
    X          int                    `json:"x"`
    Y          int                    `json:"y"`
    Properties map[string]interface{} `json:"properties"`
}

func loadMap(path string) *TileMap {
    data, err := os.ReadFile(path)
    if err != nil {
        panic(err)
    }
    var m TileMap
    if err := json.Unmarshal(data, &m); err != nil {
        panic(err)
    }
    return &m
}
```

## Rendering Tile Layers

```go
func drawTileLayer(screen *ebiten.Image, layer *MapLayer, tileSize int,
    tileImages map[string]*ebiten.Image, cameraX, cameraY float64) {

    scrollX := cameraX * layer.ScrollX
    scrollY := cameraY * layer.ScrollY

    for row := 0; row < layer.Height; row++ {
        for col := 0; col < layer.Width; col++ {
            tileIdx := layer.Data[row][col]
            // Look up sprite reference from tileset
            // Draw at (col*tileSize - scrollX, row*tileSize - scrollY)
            _ = tileIdx
            op := &ebiten.DrawImageOptions{}
            x := float64(col*tileSize) - scrollX
            y := float64(row*tileSize) - scrollY
            op.GeoM.Translate(x, y)
            // screen.DrawImage(tileImage, op)
        }
    }
}
```

## Loading Audio

```go
import (
    "os"

    "github.com/hajimehoshi/ebiten/v2/audio"
    "github.com/hajimehoshi/ebiten/v2/audio/wav"
)

var audioContext = audio.NewContext(44100)

func loadSound(path string) *audio.Player {
    f, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    stream, err := wav.DecodeWithSampleRate(44100, f)
    if err != nil {
        panic(err)
    }
    player, err := audioContext.NewPlayer(stream)
    if err != nil {
        panic(err)
    }
    return player
}

// Play a sound effect (fire-and-forget)
func playSFX(player *audio.Player) {
    player.Rewind()
    player.Play()
}

// Play background music with looping
func playMusic(path string) *audio.Player {
    f, err := os.Open(path)
    if err != nil {
        panic(err)
    }
    stream, err := wav.DecodeWithSampleRate(44100, f)
    if err != nil {
        panic(err)
    }
    loop := audio.NewInfiniteLoop(stream, stream.Length())
    player, err := audioContext.NewPlayer(loop)
    if err != nil {
        panic(err)
    }
    player.Play()
    return player
}
```

## Complete Example Game

```go
package main

import (
    "fmt"
    "image"
    _ "image/png"
    "os"

    "github.com/hajimehoshi/ebiten/v2"
    "github.com/hajimehoshi/ebiten/v2/audio"
    "github.com/hajimehoshi/ebiten/v2/audio/wav"
    "github.com/hajimehoshi/ebiten/v2/ebitenutil"
)

const (
    screenW = 320
    screenH = 240
    scale   = 2
)

type Game struct {
    playerX, playerY float64
    playerSheet      *ebiten.Image
    frame            int
    elapsed          float64
    audioCtx         *audio.Context
    jumpSound        *audio.Player
}

func NewGame() *Game {
    g := &Game{
        playerX: 100,
        playerY: 180,
        audioCtx: audio.NewContext(44100),
    }

    // Load sprite sheet
    f, _ := os.Open("build/assets/sprites/player.png")
    img, _, _ := image.Decode(f)
    f.Close()
    g.playerSheet = ebiten.NewImageFromImage(img)

    // Load jump sound
    sf, _ := os.Open("build/assets/audio/jump.wav")
    stream, _ := wav.DecodeWithSampleRate(44100, sf)
    g.jumpSound, _ = g.audioCtx.NewPlayer(stream)

    return g
}

func (g *Game) Update() error {
    // Movement
    if ebiten.IsKeyPressed(ebiten.KeyLeft) {
        g.playerX -= 2
    }
    if ebiten.IsKeyPressed(ebiten.KeyRight) {
        g.playerX += 2
    }

    // Jump with sound
    if ebiten.IsKeyPressed(ebiten.KeySpace) {
        g.jumpSound.Rewind()
        g.jumpSound.Play()
    }

    // Animation
    g.elapsed += 1.0 / 60.0
    if g.elapsed >= 1.0/8.0 { // 8 FPS animation
        g.elapsed = 0
        g.frame = (g.frame + 1) % 4
    }

    return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
    // Draw animated sprite (16x16, 4 frames horizontal)
    sx := g.frame * 16
    rect := image.Rect(sx, 0, sx+16, 16)
    sprite := g.playerSheet.SubImage(rect).(*ebiten.Image)

    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(g.playerX, g.playerY)
    screen.DrawImage(sprite, op)

    ebitenutil.DebugPrint(screen, fmt.Sprintf("Arrow keys: move, Space: jump"))
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return screenW, screenH
}

func main() {
    ebiten.SetWindowSize(screenW*scale, screenH*scale)
    ebiten.SetWindowTitle("Runefact Demo")
    if err := ebiten.RunGame(NewGame()); err != nil {
        panic(err)
    }
}
```

## Build Integration

### Makefile

```makefile
.PHONY: assets game

assets:
	runefact build

game: assets
	go build -o bin/game ./cmd/game

watch:
	runefact watch &
	go run ./cmd/game

clean:
	rm -rf build/assets bin/
```

### go generate

Add to your main package:

```go
//go:generate runefact build
```

Then run `go generate ./...` before building.

## Development Workflow

1. Edit `.sprite`, `.map`, `.sfx`, `.track` files
2. Run `runefact watch` in one terminal (auto-rebuilds on change)
3. Run your game in another terminal
4. Use `runefact preview <file>` to inspect individual assets
