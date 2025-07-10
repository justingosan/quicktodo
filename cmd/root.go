package cmd

import (
	"quicktodo/internal/commands"
)

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return commands.RootCmd.Execute()
}