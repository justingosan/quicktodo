package cmd

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
	agentID string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "quicktodo",
	Short: "A simple CLI todo management tool for AI-assisted development workflows",
	Long: `QuickTodo is a lightweight, fast, and AI-friendly CLI tool for managing tasks across multiple projects.

It provides file-based storage with concurrent access protection and comprehensive JSON output
for seamless integration with AI agents and development workflows.`,
	Version: "1.0.0",
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVar(&agentID, "agent-id", "", "Agent identifier for AI coordination")
	
	// Add version flag
	rootCmd.Flags().BoolP("version", "V", false, "Show version information")
}