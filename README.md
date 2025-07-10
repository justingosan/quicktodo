# QuickTodo

A simple CLI todo management tool designed for AI-assisted development workflows. QuickTodo provides a lightweight, fast, and AI-friendly interface for managing tasks across multiple projects.

## Features

- **Multi-project support**: Manage tasks across different projects
- **File-based storage**: Simple JSON file storage with no external dependencies
- **Concurrent access**: Safe concurrent access with file locking
- **AI-friendly**: JSON output and descriptive commands for AI agents
- **Fast and lightweight**: Single binary with minimal dependencies

## Installation

```bash
go install quicktodo
```

## Quick Start

```bash
# Initialize a project
quicktodo initialize-project

# Create a new task
quicktodo create-task "Implement user authentication"

# List all tasks
quicktodo list-tasks

# Mark a task as completed
quicktodo mark-completed 1
```

## Commands

### Project Management
- `quicktodo initialize-project [name]` - Register current directory as project
- `quicktodo list-projects` - Show all registered projects
- `quicktodo project-status` - Show current project information

### Task Management
- `quicktodo create-task "title"` - Add new task to current project
- `quicktodo list-tasks` - Show all tasks with optional filters
- `quicktodo display-task <id>` - Show detailed task information
- `quicktodo edit-task <id>` - Modify task details
- `quicktodo delete-task <id>` - Remove task from project

### Status Operations
- `quicktodo mark-completed <id>` - Mark task as done
- `quicktodo mark-in-progress <id>` - Start working on task
- `quicktodo update-status <id> <status>` - Change task status

### Advanced Features
- `quicktodo search-tasks <query>` - Find tasks by text search
- `quicktodo export-data` - Export tasks to file
- `quicktodo import-data <file>` - Import tasks from file

## Documentation

For detailed usage and AI integration, see the [AGENTS.md](AGENTS.md) file.

## License

MIT License