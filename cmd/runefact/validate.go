package main

import (
	"fmt"
	"os"

	"github.com/vgalaktionov/runefact/internal/build"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate [files...]",
	Short: "Check rune files for errors without building",
	Long: `Validate parses rune files and reports errors without producing output artifacts.
Returns exit code 0 if all files are valid, 1 if errors are found.

Examples:
  runefact validate                 # validate everything
  runefact validate player.sprite   # validate specific file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		root, cfg, err := loadProjectConfig()
		if err != nil {
			return err
		}

		opts := build.Options{
			Scope: build.ScopeAll,
			Files: args,
		}

		result := build.Validate(opts, cfg, root)

		for _, w := range result.Warnings {
			if !flagQuiet {
				fmt.Fprintf(os.Stderr, "warning: %s\n", w)
			}
		}

		if len(result.Errors) > 0 {
			for _, e := range result.Errors {
				fmt.Fprintf(os.Stderr, "error: %v\n", e)
			}
			return fmt.Errorf("validation failed with %d error(s)", len(result.Errors))
		}

		if !flagQuiet {
			fmt.Println("All files valid.")
		}
		return nil
	},
}
