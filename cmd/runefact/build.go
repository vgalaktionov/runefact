package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/vgalaktionov/runefact/internal/build"
	"github.com/vgalaktionov/runefact/internal/config"
	"github.com/spf13/cobra"
)

var (
	flagSprites bool
	flagMaps    bool
	flagAudio   bool
	flagNoCache bool
)

var buildCmd = &cobra.Command{
	Use:   "build [files...]",
	Short: "Compile rune files into game-ready artifacts",
	Long: `Build compiles .palette, .sprite, .map, .inst, .sfx, and .track files
into PNG sprite sheets, JSON tilemaps, WAV audio, and a Go manifest package.

Examples:
  runefact build                    # build everything
  runefact build --sprites          # build only sprites
  runefact build player.sprite      # build specific file`,
	RunE: runBuild,
}

func init() {
	buildCmd.Flags().BoolVar(&flagSprites, "sprites", false, "build only sprites")
	buildCmd.Flags().BoolVar(&flagMaps, "maps", false, "build only maps")
	buildCmd.Flags().BoolVar(&flagAudio, "audio", false, "build only audio")
	buildCmd.Flags().BoolVar(&flagNoCache, "no-cache", false, "force full rebuild, ignore cache")
}

func runBuild(cmd *cobra.Command, args []string) error {
	root, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	scope := build.ScopeAll
	if flagSprites {
		scope = build.ScopeSprites
	} else if flagMaps {
		scope = build.ScopeMaps
	} else if flagAudio {
		scope = build.ScopeAudio
	}

	opts := build.Options{
		Scope: scope,
		Files: args,
	}

	result := build.Build(opts, cfg, root)

	for _, w := range result.Warnings {
		if !flagQuiet {
			fmt.Fprintf(os.Stderr, "warning: %s\n", w)
		}
	}

	if len(result.Errors) > 0 {
		for _, e := range result.Errors {
			fmt.Fprintf(os.Stderr, "error: %v\n", e)
		}
		return fmt.Errorf("build failed with %d error(s)", len(result.Errors))
	}

	if !flagQuiet {
		fmt.Printf("Built %d artifact(s)\n", len(result.Artifacts))
		if flagVerbose {
			for _, a := range result.Artifacts {
				fmt.Printf("  %s\n", a)
			}
		}
	}

	return nil
}

func loadProjectConfig() (string, *config.ProjectConfig, error) {
	var root string
	var err error

	if flagConfig != "" {
		root = filepath.Dir(flagConfig)
	} else {
		root, err = config.FindProjectRoot()
		if err != nil {
			return "", nil, err
		}
	}

	cfgPath := config.GetConfigPath(root)
	cfg, err := config.LoadConfig(cfgPath)
	if err != nil {
		return "", nil, err
	}

	return root, cfg, nil
}
