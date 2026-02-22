package main

import (
	"fmt"
	"os"
)

// version is set at build time via -ldflags.
var version = "dev"

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
