package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Show documentation location",
	Long:  `Print the path to the Runefact documentation directory.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Try to find docs relative to the executable.
		exe, err := os.Executable()
		if err == nil {
			docsDir := filepath.Join(filepath.Dir(exe), "..", "docs")
			if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
				fmt.Println(docsDir)
				return nil
			}
		}

		// Try GOPATH.
		gopath := os.Getenv("GOPATH")
		if gopath == "" {
			home, _ := os.UserHomeDir()
			gopath = filepath.Join(home, "go")
		}
		pkgDir := filepath.Join(gopath, "pkg", "mod", "github.com", "vgalaktionov", "runefact@*")
		matches, _ := filepath.Glob(pkgDir)
		for _, m := range matches {
			docsDir := filepath.Join(m, "docs")
			if info, err := os.Stat(docsDir); err == nil && info.IsDir() {
				fmt.Println(docsDir)
				return nil
			}
		}

		// Fallback: print online URL.
		_ = runtime.GOOS
		fmt.Println("Documentation: https://github.com/vgalaktionov/runefact/tree/main/docs")
		return nil
	},
}
