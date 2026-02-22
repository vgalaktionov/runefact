package main

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/vgalaktionov/runefact/internal/build"
	"github.com/vgalaktionov/runefact/internal/watcher"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and rebuild automatically",
	Long: `Watch monitors rune files for changes and triggers incremental rebuilds.
Runs until interrupted with Ctrl+C.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}

		assetsDir := filepath.Join(root, "assets")
		if _, err := os.Stat(assetsDir); err != nil {
			return fmt.Errorf("assets directory not found: %s", assetsDir)
		}

		// Initial build.
		if !flagQuiet {
			fmt.Println("Running initial build...")
		}
		result := build.Build(build.Options{}, cfg, root)
		if len(result.Errors) > 0 {
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "error: %v\n", e)
			}
		} else if !flagQuiet {
			fmt.Printf("Built %d artifact(s)\n", len(result.Artifacts))
		}

		w, err := watcher.New(100*time.Millisecond, func(changed []string) error {
			r := build.Build(build.Options{}, cfg, root)
			for _, e := range r.Errors {
				fmt.Fprintf(os.Stderr, "error: %v\n", e)
			}
			if len(r.Errors) == 0 && !flagQuiet {
				fmt.Printf("Rebuilt %d artifact(s)\n", len(r.Artifacts))
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("creating watcher: %w", err)
		}

		if err := w.WatchDir(assetsDir); err != nil {
			return fmt.Errorf("watching %s: %w", assetsDir, err)
		}

		if !flagQuiet {
			fmt.Println("Watching for changes... (Ctrl+C to stop)")
		}

		// Handle Ctrl+C.
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

		go w.Start()

		<-sig
		if !flagQuiet {
			fmt.Println("\nStopping watcher.")
		}
		return w.Stop()
	},
}
