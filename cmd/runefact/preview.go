package main

import (
	"fmt"
	"path/filepath"

	"github.com/vgalaktionov/runefact/internal/preview"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview [file]",
	Short: "Open live-reloading asset previewer",
	Long: `Preview opens an ebitengine window to display sprites, maps, or audio.

Examples:
  runefact preview player.sprite    # preview sprite file
  runefact preview world.map        # preview map file
  runefact preview laser.sfx        # preview sound effect
  runefact preview bgm.track        # preview music track`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}

		if len(args) == 0 {
			return fmt.Errorf("please specify a file to preview (e.g., runefact preview player.sprite)")
		}

		file := args[0]
		ext := filepath.Ext(file)

		// Resolve to full path based on file type.
		var fullPath string
		switch ext {
		case ".sprite":
			fullPath = filepath.Join(root, "assets", "sprites", file)
		case ".map":
			fullPath = filepath.Join(root, "assets", "maps", file)
		case ".sfx":
			fullPath = filepath.Join(root, "assets", "sfx", file)
		case ".track":
			fullPath = filepath.Join(root, "assets", "tracks", file)
		case ".palette":
			fullPath = filepath.Join(root, "assets", "palettes", file)
		case ".inst":
			fullPath = filepath.Join(root, "assets", "instruments", file)
		default:
			fullPath = filepath.Join(root, "assets", file)
		}

		// If the arg is already an absolute path, use it directly.
		if filepath.IsAbs(file) {
			fullPath = file
		}

		assetsDir := filepath.Join(root, "assets")
		p := preview.NewPreviewer(
			fullPath,
			assetsDir,
			cfg.Preview.WindowWidth,
			cfg.Preview.WindowHeight,
			cfg.Defaults.SampleRate,
		)
		return p.Run()
	},
}
