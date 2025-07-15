package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

// contextCmd represents the context command
var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Show AI usage instructions for this project",
	Long: `Display comprehensive usage instructions for AI agents working with QuickTodo.

This command outputs detailed information about available commands, JSON formats,
best practices, and troubleshooting for AI-assisted development workflows.`,
	Run: runContext,
}

func runContext(cmd *cobra.Command, args []string) {
	fmt.Print(`# QuickTodo AI Usage Instructions

## Core Commands
quicktodo init                                    # Initialize project (run once)
quicktodo create-task "Title" --priority high    # Create task
quicktodo list-tasks --json                      # List all tasks
quicktodo set-task-status <id> <status>          # Change status
quicktodo mark-completed <id>                    # Mark done
quicktodo edit-task <id> --title "New title"     # Edit task

## Status Values: pending | in_progress | done
## Priority Values: low | medium | high

## Essential Usage
- **Always use --json flag** for programmatic access
- Commands return exit code 0=success, 1=error
- Use descriptive titles and appropriate priority

## Quick Examples
quicktodo create-task "Fix login bug" --priority high --json
quicktodo list-tasks --status pending --json
quicktodo mark-completed 1 --json

## Common Errors
- "not a registered project" → Run 'quicktodo init' first
- "task not found" → Check ID with 'quicktodo list-tasks'
`)
}

func init() {
	RootCmd.AddCommand(contextCmd)
}