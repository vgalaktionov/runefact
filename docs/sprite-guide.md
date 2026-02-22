# Sprite Authoring Guide

## Palette Design

Start with a small, purposeful palette. 8–16 colors is typical for retro-style games.

**Naming conventions:**
- Single-char keys for frequently-used colors: `k` (black), `w` (white), `r` (red)
- `_` for transparent (always)
- Multi-char bracket keys `[skin]` for semantic colors used sparingly

```toml
name = "character"

[colors]
_ = "transparent"
k = "#1a1c2c"
w = "#f4f4f4"
r = "#b13e53"
b = "#29366f"
s = "#ffaa77"
h = "#7e3517"
```

## Grid Size Selection

| Grid Size | Use Case |
|-----------|----------|
| 8x8 | Tiny icons, particles, simple tiles |
| 16x16 | Standard retro sprites, most common |
| 16x24 | Taller characters with more detail |
| 32x32 | Detailed characters, large objects |
| Non-square | Use `"WxH"` format: `grid = "16x24"` |

## Pixel Grid Basics

Each character in the grid corresponds to one pixel, mapped to a palette color:

```toml
[sprite.heart]
grid = 8
pixels = """
_rr_rr_
rrrrrrr
rrrrrrr
_rrrrr_
__rrr__
___r___
"""
```

For multi-char keys, use bracket syntax:

```toml
[palette_extend]
skin = "#ffaa77"
hair = "#7e3517"

[sprite.face]
grid = 8
pixels = """
_[hair][hair][hair][hair]_
[hair][skin][skin][skin][skin][hair]
[skin][skin]k[skin]k[skin]
[skin][skin][skin][skin][skin][skin]
[skin]_[skin][skin]_[skin]
_[skin][skin][skin][skin]_
"""
```

## Animation

### Frame-based animation

Use `[[sprite.NAME.frame]]` arrays with a `framerate`:

```toml
[sprite.coin]
grid = 8
framerate = 6

[[sprite.coin.frame]]
pixels = """
_yyyy_
yyyyyy
yyyyyy
yyyyyy
yyyyyy
_yyyy_
"""

[[sprite.coin.frame]]
pixels = """
__yy__
_yyyy_
_yyyy_
_yyyy_
_yyyy_
__yy__
"""

[[sprite.coin.frame]]
pixels = """
___y__
__yy__
__yy__
__yy__
__yy__
___y__
"""
```

**Guidelines:**
- All frames must have identical dimensions
- `framerate` is frames per second (6–12 for most animations)
- Fewer frames with good key poses beats many similar frames
- Loop-friendly: last frame should transition smoothly back to first

### Common animation patterns

| Pattern | Frames | FPS | Notes |
|---------|--------|-----|-------|
| Idle breathing | 2–4 | 2–4 | Subtle vertical shift |
| Walk cycle | 4–6 | 8–10 | Key poses: contact, passing, contact, passing |
| Attack | 3–5 | 12 | Wind-up, strike, recovery |
| Coin spin | 3–4 | 6–8 | Front, narrow, side, narrow |
| Explosion | 4–6 | 12 | Expand outward, fade colors |

## Sprite Sheet Organization

Group related sprites in one `.sprite` file:

```toml
palette = "character"
grid = 16

[sprite.player_idle]
pixels = """..."""

[sprite.player_walk]
framerate = 8
[[sprite.player_walk.frame]]
pixels = """..."""
[[sprite.player_walk.frame]]
pixels = """..."""

[sprite.player_jump]
pixels = """..."""
```

The renderer packs all sprites from one file into a single PNG sheet. Keep logically related sprites together — one file per character or tileset.

## Palette Extend

Override or add colors without modifying the shared palette:

```toml
palette = "default"

[palette_extend]
glow = "#ffff0080"
shadow = "#00000040"

[sprite.effect]
grid = 8
pixels = """
_[glow][glow]_
[glow][glow][glow][glow]
[glow][glow][glow][glow]
_[glow][glow]_
"""
```

`palette_extend` can also be set per-sprite for sprite-specific colors.

## Troubleshooting

**"Ragged row"** — rows have inconsistent widths. Count characters carefully; bracket keys `[xx]` count as one pixel.

**"Unknown palette key"** — the character isn't in your palette or `palette_extend`. Check for typos.

**"Frame dimension mismatch"** — animation frames aren't the same size. Every frame must match the sprite's `grid`.

**Sprite not appearing in sheet** — ensure the sprite section header is `[sprite.name]`, not just `[name]`.
