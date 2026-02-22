package main

import (
	"github.com/spf13/cobra"
)

var (
	flagConfig  string
	flagVerbose bool
	flagQuiet   bool
)

var rootCmd = &cobra.Command{
	Use:   "runefact",
	Short: "Runes become artifacts",
	Long:  "Runefact compiles text-based asset definitions into game-ready artifacts for ebitengine.",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&flagConfig, "config", "", "path to runefact.toml (default: auto-detect)")
	rootCmd.PersistentFlags().BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	rootCmd.PersistentFlags().BoolVarP(&flagQuiet, "quiet", "q", false, "suppress non-error output")

	rootCmd.AddCommand(buildCmd)
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(previewCmd)
	rootCmd.AddCommand(watchCmd)
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(mcpCmd)
	rootCmd.AddCommand(docsCmd)
	rootCmd.AddCommand(versionCmd)
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}
