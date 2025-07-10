package commands

import (
	"github.com/spf13/cobra"
)

var (
	verbose    bool
	agentID    string
	jsonOutput bool
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "quicktodo",
	Short: "A simple CLI todo management tool for AI-assisted development workflows",
	Long: `QuickTodo is a lightweight, fast, and AI-friendly CLI tool for managing tasks across multiple projects.

It provides file-based storage with concurrent access protection and comprehensive JSON output
for seamless integration with AI agents and development workflows.`,
	Version: "1.0.0",
}

func init() {
	// Global flags
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	RootCmd.PersistentFlags().StringVar(&agentID, "agent-id", "", "Agent identifier for AI coordination")
	RootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "Output in JSON format")
	
	// Disable completion command
	RootCmd.CompletionOptions.DisableDefaultCmd = true
}
