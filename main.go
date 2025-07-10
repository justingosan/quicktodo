package main

import (
	"fmt"
	"os"
	"quicktodo/cmd"
	"quicktodo/internal/commands"
)

// Build-time variables
var (
	Version   = "dev"
	BuildTime = "unknown"
	GoVersion = "unknown"
)

func main() {
	// Set version information in the root command
	commands.RootCmd.Version = fmt.Sprintf("%s (built %s with %s)", Version, BuildTime, GoVersion)

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
