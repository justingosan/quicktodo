package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version information for QuickTodo CLI tool.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("QuickTodo CLI v%s\n", RootCmd.Version)
		fmt.Println("A simple CLI todo management tool for AI-assisted development workflows")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}