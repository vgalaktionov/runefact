package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/cobra"
	"github.com/vgalaktionov/runefact/internal/config"
)

var (
	flagInitName  string
	flagInitForce bool
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a new Runefact project",
	Long: `Init scaffolds a new Runefact project with directory structure,
default runefact.toml, and demo asset files.

Examples:
  runefact init                     # create project in current directory
  runefact init --name my-game      # set project name`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().StringVar(&flagInitName, "name", "my-game", "project name")
	initCmd.Flags().BoolVar(&flagInitForce, "force", false, "overwrite existing files")
}

func runInit(cmd *cobra.Command, args []string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	configPath := config.GetConfigPath(wd)
	if !flagInitForce {
		if _, err := os.Stat(configPath); err == nil {
			return fmt.Errorf("runefact.toml already exists (use --force to overwrite)")
		}
	}

	dirs := []string{
		"assets/palettes",
		"assets/sprites",
		"assets/maps",
		"assets/instruments",
		"assets/sfx",
		"assets/tracks",
		".claude",
	}
	for _, d := range dirs {
		if err := os.MkdirAll(filepath.Join(wd, d), 0755); err != nil {
			return fmt.Errorf("creating %s: %w", d, err)
		}
	}

	files := scaffoldFiles(flagInitName)
	for _, f := range files {
		path := filepath.Join(wd, f.path)
		if err := os.WriteFile(path, []byte(f.content), 0644); err != nil {
			return fmt.Errorf("writing %s: %w", f.path, err)
		}
	}

	// Set up MCP config (merges into existing files).
	mcpFiles := setupMCPConfig(wd)

	if !flagQuiet {
		fmt.Printf("Initialized Runefact project %q\n", flagInitName)
		fmt.Println("Created:")
		for _, f := range files {
			fmt.Printf("  %s\n", f.path)
		}
		for _, f := range mcpFiles {
			fmt.Printf("  %s\n", f)
		}
		fmt.Println("\nRun 'runefact build' to compile assets.")
	}

	return nil
}

type scaffoldFile struct {
	path    string
	content string
}

func scaffoldFiles(name string) []scaffoldFile {
	return []scaffoldFile{
		{
			path: "runefact.toml",
			content: fmt.Sprintf(`[project]
name = %q
output = "build/assets"
package = "assets"

[defaults]
sprite_size = 16
sample_rate = 44100
bit_depth = 16

[preview]
window_width = 1200
window_height = 900
background = "#1a1a2e"
pixel_scale = 4
audio_volume = 0.5
`, name),
		},
		{
			path: "assets/palettes/default.palette",
			content: `name = "default"

[colors]
_ = "transparent"
k = "#000000"
w = "#ffffff"
r = "#ff004d"
b = "#29adff"
g = "#00e436"
d = "#1d2b53"
s = "#ffccaa"
h = "#ab5236"
y = "#ffec27"
o = "#ffa300"
p = "#7e2553"
l = "#83769c"
c = "#008751"
e = "#5f574f"
f = "#c2c3c7"
`,
		},
		{
			path: "assets/sprites/player.sprite",
			content: `palette = "default"
grid = 32

palette_extend = { "n" = "#1a1a2e", "t" = "#e8d4b8" }

[sprite.idle]
framerate = 3

[[sprite.idle.frame]]
pixels = """
________________________________
_____________hhhhh______________
___________hhhhhhhhh____________
__________hhhhhhhhhh____________
__________hhhhhhhhhh____________
__________dddddddddd____________
_________dddddddddddd___________
_________dddwkddwkddd___________
_________dddkkddkkddd___________
_________ddttttttttdd___________
_________dddttttttddd___________
__________dtttsttttd____________
__________ddttttttdd____________
___________dddddddd_____________
___________bbbbbbbbb____________
__________bbbblbbbbbb___________
_________bbbbbbbbbbbbb__________
________sbbbbbbbbbbbbbs_________
________s_bbbbbbbbbbb_s_________
__________bbbbbbbbbbb___________
___________bbb___bbb____________
___________bbb___bbb____________
___________bbb___bbb____________
__________nnnn___nnnn___________
__________nnnn___nnnn___________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
"""

[[sprite.idle.frame]]
pixels = """
________________________________
_____________hhhhh______________
___________hhhhhhhhh____________
__________hhhhhhhhhh____________
__________hhhhhhhhhh____________
__________dddddddddd____________
_________dddddddddddd___________
_________dddwkddwkddd___________
_________dddkkddkkddd___________
_________ddttttttttdd___________
_________dddttttttddd___________
__________dtttsttttd____________
__________ddttttttdd____________
___________dddddddd_____________
___________bbbbbbbbb____________
__________bbbblbbbbbb___________
_________bbbbbbbbbbbbb__________
________sbbbbbbbbbbbbbs_________
________s_bbbbbbbbbbb_s_________
__________bbbbbbbbbbb___________
___________bbb___bbb____________
___________bbb___bbb____________
__________nbbb___bbbn___________
__________nnnn___nnnn___________
___________nn_____nn____________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
________________________________
"""

[sprite.coin]
grid = 16
framerate = 6

[[sprite.coin.frame]]
pixels = """
_____yyyy_______
___yyooooyyy____
__yooyyyyyooy___
__yoyyyyyyyyoy__
__yoyyyy_yyyoy__
__yoyyyyyyyyoy__
__yoyyyyyyyyoy__
__yoyyyyyyyyoy__
___yooooooooy___
____yyyyyyyy____
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
______yy________
_____yooy_______
____yoyyoy______
____yoy_oy______
____yoyyoy______
____yoyyoy______
____yoyyoy______
_____yooy_______
______yy________
________________
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
_______y________
______yoy_______
______yoy_______
______yoy_______
______yoy_______
______yoy_______
______yoy_______
_______y________
________________
________________
________________
________________
________________
________________
________________
________________
"""

[[sprite.coin.frame]]
pixels = """
______yy________
_____yooy_______
____yoyyoy______
____yoy_oy______
____yoyyoy______
____yoyyoy______
____yoyyoy______
_____yooy_______
______yy________
________________
________________
________________
________________
________________
________________
________________
"""

[sprite.heart]
grid = 16
pixels = """
________________
___rr____rr_____
__rrrr__rrrr____
_rrrrrrrrrrrr___
_rrrrrrrrrrrr___
_rrrrrrrrrrrr___
__rrrrrrrrrr____
___rrrrrrrr_____
____rrrrrr______
_____rrrr_______
______rr________
________________
________________
________________
________________
________________
"""
`,
		},
		{
			path: "assets/sprites/tiles.sprite",
			content: `palette = "default"
grid = 16

[sprite.grass]
pixels = """
ccggccggccggccgg
ggccggccggccggcc
ccgcgcgcgcgcgccc
gcccccggccccccgc
ccchccccchccccch
hchhhhhchhhhhchh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
"""

[sprite.dirt]
pixels = """
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhehhhhhhhehhhhh
hhhhhhhhhhhhhhhh
hhhhhhehhhhhhhhh
hhhhhhhhhhhhhhhh
hhehhhhhehhhhehh
"""

[sprite.stone]
pixels = """
eeffffffeeffffff
flllllfeflllllfe
flllflfefllflffe
flllllfeflllllfe
eeffffffeeffffff
feflllleefellllf
fefllfleefelfllf
feflllleefellllf
eeffffffeeffffff
flllllfeflllllfe
flllflfefllflffe
flllllfeflllllfe
eeffffffeeffffff
feflllleefellllf
fefllfleefelfllf
feflllleefellllf
"""

[sprite.sky]
pixels = """
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
"""
`,
		},
		{
			path: "assets/maps/level1.map",
			content: `tile_size = 16

[tileset]
_ = ""
g = "tiles:grass"
d = "tiles:dirt"
s = "tiles:stone"
b = "tiles:sky"

[layer.background]
scroll_x = 0.5
pixels = """
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
bbbbbbbbbbbbbbbb
"""

[layer.main]
pixels = """
________________
________________
________________
________________
_______sss______
________________
__gg_______gg___
__dd_______dd___
gggggggggggggggg
dddddddddddddddd
"""

[layer.entities]

[[layer.entities.entity]]
type = "spawn"
x = 2
y = 7
[layer.entities.entity.properties]
sprite = "player:idle"

[[layer.entities.entity]]
type = "coin"
x = 4
y = 5
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "coin"
x = 8
y = 3
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "coin"
x = 11
y = 5
[layer.entities.entity.properties]
sprite = "player:coin"

[[layer.entities.entity]]
type = "enemy"
x = 13
y = 7
[layer.entities.entity.properties]
sprite = "player:heart"
`,
		},
		{
			path: "assets/instruments/lead.inst",
			content: `name = "lead"

[oscillator]
waveform = "square"
duty_cycle = 0.5

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.6
release = 0.2

[filter]
type = "lowpass"
cutoff = 2000
resonance = 0.2
`,
		},
		{
			path: "assets/instruments/bass.inst",
			content: `name = "bass"

[oscillator]
waveform = "triangle"

[envelope]
attack = 0.005
decay = 0.15
sustain = 0.5
release = 0.1
`,
		},
		{
			path: "assets/sfx/jump.sfx",
			content: `duration = 0.2
volume = 0.7

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.08
sustain = 0.2
release = 0.12

[voice.pitch]
start = 220
end = 660
curve = "exponential"
`,
		},
		{
			path: "assets/sfx/coin.sfx",
			content: `duration = 0.12
volume = 0.6

[[voice]]
waveform = "square"
duty_cycle = 0.25

[voice.envelope]
attack = 0.0
decay = 0.03
sustain = 0.4
release = 0.09

[voice.pitch]
start = 880
end = 1320
curve = "linear"
`,
		},
		{
			path: "assets/tracks/demo.track",
			content: `tempo = 140
ticks_per_beat = 4
loop = true
loop_start = 0

[[channel]]
name = "melody"
instrument = "lead"
volume = 0.7

[[channel]]
name = "bass"
instrument = "bass"
volume = 0.6

[pattern.intro]
ticks = 16
data = """
melody  | bass
C4      | C2
---     | ---
E4      | ---
---     | ---
G4      | G2
---     | ---
E4      | ---
---     | ---
A4      | A2
---     | ---
G4      | ---
---     | ---
E4      | E2
---     | ---
D4      | ---
---     | ---
"""

[pattern.verse]
ticks = 16
data = """
melody  | bass
E4      | A2
---     | ---
D4      | ---
---     | ---
C4      | F2
---     | ---
D4      | ---
---     | ---
E4      | G2
---     | ---
E4      | ---
---     | ---
E4      | C2
---     | ---
^^^     | ---
...     | ^^^
"""

[song]
sequence = ["intro", "verse", "intro", "verse"]
`,
		},
	}
}

// setupMCPConfig merges runefact MCP server config into .mcp.json and
// .claude/settings.local.json without clobbering existing entries.
// Returns the list of files that were written.
func setupMCPConfig(wd string) []string {
	runefactBin, _ := os.Executable()
	if runefactBin == "" {
		runefactBin = "runefact"
	}

	var written []string

	// --- .mcp.json: add runefact server entry ---
	mcpPath := filepath.Join(wd, ".mcp.json")
	var mcpDoc map[string]any
	if data, err := os.ReadFile(mcpPath); err == nil {
		_ = json.Unmarshal(data, &mcpDoc)
	}
	if mcpDoc == nil {
		mcpDoc = map[string]any{}
	}
	servers, _ := mcpDoc["mcpServers"].(map[string]any)
	if servers == nil {
		servers = map[string]any{}
	}
	if _, exists := servers["runefact"]; !exists {
		servers["runefact"] = map[string]any{
			"type":    "stdio",
			"command": runefactBin,
			"args":    []string{"mcp"},
		}
		mcpDoc["mcpServers"] = servers
		if data, err := json.MarshalIndent(mcpDoc, "", "\t"); err == nil {
			data = append(data, '\n')
			_ = os.WriteFile(mcpPath, data, 0644)
			written = append(written, ".mcp.json")
		}
	}

	// --- .claude/settings.local.json: add runefact to enabled servers and permissions ---
	claudePath := filepath.Join(wd, ".claude", "settings.local.json")
	var claudeDoc map[string]any
	if data, err := os.ReadFile(claudePath); err == nil {
		_ = json.Unmarshal(data, &claudeDoc)
	}
	if claudeDoc == nil {
		claudeDoc = map[string]any{}
	}

	// enableAllProjectMcpServers
	if _, ok := claudeDoc["enableAllProjectMcpServers"]; !ok {
		claudeDoc["enableAllProjectMcpServers"] = true
	}

	// enabledMcpjsonServers: add "runefact" if missing.
	enabledRaw, _ := claudeDoc["enabledMcpjsonServers"].([]any)
	enabled := make([]string, 0, len(enabledRaw)+1)
	for _, v := range enabledRaw {
		if s, ok := v.(string); ok {
			enabled = append(enabled, s)
		}
	}
	if !slices.Contains(enabled, "runefact") {
		enabled = append(enabled, "runefact")
	}
	claudeDoc["enabledMcpjsonServers"] = enabled

	// permissions.allow: add "mcp__runefact__*" if missing.
	permsRaw, _ := claudeDoc["permissions"].(map[string]any)
	if permsRaw == nil {
		permsRaw = map[string]any{}
	}
	allowRaw, _ := permsRaw["allow"].([]any)
	allow := make([]string, 0, len(allowRaw)+1)
	for _, v := range allowRaw {
		if s, ok := v.(string); ok {
			allow = append(allow, s)
		}
	}
	if !slices.Contains(allow, "mcp__runefact__*") {
		allow = append(allow, "mcp__runefact__*")
	}
	permsRaw["allow"] = allow
	claudeDoc["permissions"] = permsRaw

	if data, err := json.MarshalIndent(claudeDoc, "", "  "); err == nil {
		data = append(data, '\n')
		_ = os.WriteFile(claudePath, data, 0644)
		written = append(written, ".claude/settings.local.json")
	}

	return written
}
