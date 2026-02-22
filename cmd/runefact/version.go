package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print runefact version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("runefact %s\n", version)
	},
}
