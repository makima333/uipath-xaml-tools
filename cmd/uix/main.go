package main

import (
	"fmt"
	"os"

	"github.com/makima333/uipath-xaml-tools/internal/cli"
)

// version is set by the release workflow using -ldflags.
var version = "0.1.0-dev"

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout, os.Stderr, version); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
