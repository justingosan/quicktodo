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

## Essential Commands

### Initialize project (run once)
quicktodo init

### Task Management
quicktodo create-task "Task title" --description "Optional description" --priority high|medium|low
quicktodo list-tasks
quicktodo list-tasks --status pending|in_progress|done
quicktodo list-tasks --priority high|medium|low
quicktodo display-task <id>

### Status Updates  
quicktodo set-task-status <id> <status>
quicktodo mark-completed <id>
quicktodo mark-in-progress <id>
quicktodo mark-pending <id>

### JSON Output (AI-friendly)
Add --json flag to any command for machine-readable output:
quicktodo list-tasks --json
quicktodo display-task <id> --json
quicktodo create-task "title" --json

## Task Status Values
- pending: Task not started (default)
- in_progress: Task currently being worked on  
- done: Task completed

## Priority Values
- high: Urgent/important tasks
- medium: Normal priority (default)
- low: Nice-to-have tasks

## JSON Response Formats

### Task Object
{
  "id": 1,
  "title": "Task title",
  "description": "Task description",
  "status": "pending",
  "priority": "medium", 
  "created_at": "2025-01-01T00:00:00Z",
  "updated_at": "2025-01-01T00:00:00Z",
  "assigned_to": "",
  "locked_by": "",
  "locked_at": "0001-01-01T00:00:00Z"
}

### List Response
{
  "success": true,
  "project": {
    "name": "project-name",
    "path": "/path/to/project"
  },
  "task_count": 5,
  "tasks": [...]
}

### Status Change Response
{
  "success": true,
  "task": {...},
  "old_status": "pending",
  "new_status": "done",
  "changed_at": "2025-01-01T00:00:00Z"
}

## Best Practices for AI Agents

1. **Always use --json flag** for programmatic access
2. **Check exit codes** - 0 = success, 1 = error
3. **Handle file locking** - retry if lock acquisition fails
4. **Use descriptive titles** for better task management
5. **Set appropriate priority** based on task importance
6. **Include --agent-id** to track AI assignments

## Error Handling

- Commands return exit code 0 on success, 1 on error
- Error messages are sent to stderr
- JSON responses include "success": true/false field
- File locking prevents concurrent access conflicts

## Workflow Examples

### Create and track task
quicktodo create-task "Fix login bug" --priority high --json
quicktodo set-task-status 1 in_progress --json
quicktodo mark-completed 1 --json

### Query current work
quicktodo list-tasks --status pending --json
quicktodo list-tasks --priority high --json

### Get task details
quicktodo display-task 1 --json

## Troubleshooting

- **"not a registered project"** - Run 'quicktodo init' first
- **Lock timeout errors** - Another process is using the project, wait and retry
- **Task not found** - Check task ID with 'quicktodo list-tasks'
- **Permission errors** - Ensure write access to ~/.config/quicktodo/

## File Locations

- Project database: ~/.config/quicktodo/projects/{project-name}.json
- Registry: ~/.config/quicktodo/projects.json
- Configuration: ~/.config/quicktodo/config.json
- Lock files: ~/.config/quicktodo/locks/
`)
}

func init() {
	RootCmd.AddCommand(contextCmd)
}