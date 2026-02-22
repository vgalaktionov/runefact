package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vgalaktionov/runefact/internal/config"
	"github.com/spf13/cobra"
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

	if !flagQuiet {
		fmt.Printf("Initialized Runefact project %q\n", flagInitName)
		fmt.Println("Created:")
		for _, f := range files {
			fmt.Printf("  %s\n", f.path)
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
window_width = 800
window_height = 600
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
			path: "assets/sprites/demo.sprite",
			content: `palette = "default"
grid = 8

[sprite.heart]
pixels = """
_rr__rr_
rrrrrrr_
rrrrrrrr
rrrrrrrr
_rrrrrr_
__rrrr__
___rr___
________
"""

[sprite.star]
framerate = 4

[[sprite.star.frame]]
pixels = """
____y___
___yyy__
__yyyyy_
___yyy__
____y___
________
________
________
"""

[[sprite.star.frame]]
pixels = """
________
___y_y__
____y___
__yyyyy_
____y___
___y_y__
________
________
"""
`,
		},
		{
			path: "assets/maps/demo.map",
			content: `tile_size = 8

[tileset]
H = "demo:heart"
S = "demo:star"
_ = ""

[layer.main]
pixels = """
________
_H___H__
________
_S_S_S__
________
________
________
________
"""
`,
		},
		{
			path: "assets/instruments/demo.inst",
			content: `name = "demo"

[oscillator]
waveform = "square"

[envelope]
attack = 0.01
decay = 0.1
sustain = 0.6
release = 0.15
`,
		},
		{
			path: "assets/sfx/demo.sfx",
			content: `duration = 0.15
volume = 0.8

[[voice]]
waveform = "square"
duty_cycle = 0.5

[voice.envelope]
attack = 0.0
decay = 0.05
sustain = 0.3
release = 0.1

[voice.pitch]
start = 200
end = 600
curve = "exponential"
`,
		},
		{
			path: "assets/tracks/demo.track",
			content: `tempo = 120
ticks_per_beat = 4
loop = true
loop_start = 0

[[channel]]
name = "melody"
instrument = "demo"
volume = 0.8

[pattern.main]
ticks = 8
data = """
melody
C4
---
E4
---
G4
---
E4
---
"""

[song]
sequence = ["main", "main"]
`,
		},
	}
}
