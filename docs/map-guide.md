# Map Authoring Guide

## Tileset Design

A tileset maps single characters to sprite references. Design tiles to be modular and reusable.

```toml
tile_size = 16

[tileset]
g = "terrain:grass"
d = "terrain:dirt"
w = "terrain:water"
t = "terrain:tree"
s = "terrain:stone"
_ = ""
```

**Reference format:** `"sprite_file:sprite_name"` — the sprite file without extension, colon, then the sprite name defined in that file.

**Tips:**
- Choose memorable single-char keys: `g` for grass, `w` for water
- `_` (empty string) for empty tiles
- Keep one tileset per map; reuse sprite files across maps

## Layer Organization

Maps support multiple layers rendered back-to-front:

```toml
[layer.sky]          # furthest back
scroll_x = 0.2
pixels = """..."""

[layer.ground]       # main gameplay layer
pixels = """..."""

[layer.foreground]   # rendered on top
pixels = """..."""

[layer.objects]      # entity layer
[[layer.objects.entity]]
type = "spawn"
x = 32
y = 48
```

**Recommended layer stack:**
1. **background** — sky, distant scenery (with parallax)
2. **ground** — main terrain the player walks on
3. **details** — decorative overlay (flowers, cracks)
4. **foreground** — elements rendered in front of the player
5. **objects** — entity layer for spawns, items, triggers

## Parallax Scrolling

Set `scroll_x` and `scroll_y` to control parallax. Values are multipliers relative to camera movement:

| Value | Effect |
|-------|--------|
| 0.0 | Fixed (doesn't scroll) |
| 0.5 | Scrolls at half speed (distant background) |
| 1.0 | Normal scroll (default, gameplay layer) |
| 1.5 | Scrolls faster than camera (close foreground) |

```toml
[layer.clouds]
scroll_x = 0.3
scroll_y = 0.1
pixels = """
________c___c_______
____cc____cc________
_________ccc________
"""
```

## Entity Placement

Entity layers place objects at pixel coordinates with arbitrary properties:

```toml
[layer.gameplay]

[[layer.gameplay.entity]]
type = "player_spawn"
x = 64
y = 192

[[layer.gameplay.entity]]
type = "enemy"
x = 256
y = 192
properties = { enemy_type = "slime", patrol_range = 128 }

[[layer.gameplay.entity]]
type = "chest"
x = 384
y = 160
properties = { locked = true, contents = "sword", rarity = "rare" }

[[layer.gameplay.entity]]
type = "trigger"
x = 512
y = 0
properties = { width = 32, height = 224, on_enter = "boss_fight" }
```

**Properties** are arbitrary key-value pairs — use them to encode game logic. The JSON output preserves the types (string, number, boolean).

## Building a Platformer Level

```toml
tile_size = 16

[tileset]
_ = ""
g = "terrain:grass"
d = "terrain:dirt"
s = "terrain:stone"
b = "terrain:brick"
l = "terrain:lava"
c = "terrain:coin"

[layer.background]
scroll_x = 0.3
pixels = """
________________
________________
________________
________________
________________
________________
"""

[layer.terrain]
pixels = """
________________
________________
____bb____bb____
________________
__gg__gg__gg__gg
dddddddddddddd
ssssssssssssssss
"""

[layer.items]

[[layer.items.entity]]
type = "coin"
x = 80
y = 48

[[layer.items.entity]]
type = "coin"
x = 112
y = 48

[[layer.items.entity]]
type = "player_spawn"
x = 16
y = 64

[[layer.items.entity]]
type = "goal"
x = 224
y = 64
```

## Building an RPG Overworld

```toml
tile_size = 16

[tileset]
g = "overworld:grass"
f = "overworld:forest"
m = "overworld:mountain"
w = "overworld:water"
p = "overworld:path"
b = "overworld:bridge"
s = "overworld:sand"
_ = ""

[layer.terrain]
pixels = """
mmmmffggggggwwww
mmmfggggggggwwww
mmfgggppgggggsss
mfggggpggggggsss
fggggppppggggggs
ggggppppppgggggg
gggppbbppppggggg
ggppwwwwpppggggg
gppwwwwwwpgggggg
gpwwwwwwwpgggggg
"""

[layer.locations]

[[layer.locations.entity]]
type = "town"
x = 80
y = 32
properties = { name = "Startville", population = 50 }

[[layer.locations.entity]]
type = "dungeon"
x = 16
y = 16
properties = { name = "Dark Cave", level = 5 }
```

## Troubleshooting

**"Ragged row"** — tile grid rows have different lengths. Every row must have the same number of characters.

**"Unknown tileset key"** — a character in the pixel grid isn't in the `[tileset]` section.

**"Missing tile_size"** — `tile_size` is required and must be a positive integer.

**"Tileset reference should be file:sprite format"** — use `"filename:spritename"`, not just a filename.
