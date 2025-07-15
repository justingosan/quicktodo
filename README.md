# QuickTodo

A simple todo list CLI built specifically for AI coding workflows.

## Why QuickTodo?

- **Built for AI**: Designed from the ground up for AI-assisted development
- **Lightweight alternative**: Simple alternative to taskmaster.dev which can feel bloated
- **Zero setup**: Just run `init` and start coding with your AI

## Installation

```bash
# Install from source
git clone https://github.com/justingosan/quicktodo.git
cd quicktodo
make install

# Or download binary from releases
```

## How to Use

### 1. Initialize in your project
```bash
cd your-project
quicktodo init
```

### 2. Configure your AI
Add this to your `.cursorrules` or `CLAUDE.md`:
```
Before starting any task, use the quicktodo task management tool. If you don't know how to use it, run: quicktodo context
```

**For Claude Code users**: Set up a PreToolUse hook to automatically run `quicktodo context` before any tool use. This ensures instructions are always available throughout long conversations:

```json
// ~/.claude/settings.json
{
  "hooks": {
    "PreToolUse": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "quicktodo context",
            "timeout": 30
          }
        ]
      }
    ]
  }
}
```

### 3. Code with AI in YOLO mode
Run your AI coder with maximum autonomy - QuickTodo handles the task tracking:

```bash
# AI creates tasks
quicktodo create-task "Fix user login bug" --priority high

# AI lists current work
quicktodo list-tasks --json

# AI marks tasks complete
quicktodo mark-completed 1
```

## Basic Commands

```bash
quicktodo init                           # Initialize project
quicktodo create-task "Task title"       # Create new task
quicktodo list-tasks                     # List all tasks
quicktodo display-task 1                 # Show task details
quicktodo edit-task 1 --title "New title" --description "New description"
quicktodo mark-completed 1               # Mark task done
quicktodo serve                          # Start web kanban board
```

All commands support `--json` for AI consumption.

## Web Interface

QuickTodo now includes a web-based kanban board for visual task management:

```bash
# Start the web server (default port 8080)
quicktodo serve

# Custom port and auto-open browser
quicktodo serve --port 9000 --open
```

### Features:
- **Drag & Drop**: Move tasks between Pending, In Progress, and Done columns
- **Real-time Updates**: See changes instantly when AI modifies tasks via CLI
- **Enhanced Task Cards**: Display task ID, creation date, and last updated time
- **Smart Project Detection**: Automatically detects and loads current project
- **Task Management**: Click tasks to edit, delete, or view details
- **Copy Tasks**: Hover over tasks and click 📋 to copy task details with configurable prefix
- **Multi-Project**: Switch between different quicktodo projects
- **Settings**: Configure copy prefix and format for AI workflows
- **WebSocket Integration**: Live synchronization between CLI and web interface
- **Responsive**: Works on desktop, tablet, and mobile devices

#### Real-time AI Collaboration
When you run `quicktodo serve`, any CLI commands executed by AI assistants will instantly appear in the web interface:
- Task creation shows up immediately in the appropriate column
- Status changes move tasks between columns in real-time
- Task edits update titles, descriptions, and metadata instantly
- Visual notifications keep you informed of all AI activity

#### Smart Project Management
- **Auto-detection**: Running `quicktodo serve` in a project directory automatically loads that project
- **Helpful guidance**: If no project is found, shows initialization instructions and lists available projects
- **Project switching**: Easily navigate between multiple projects from the web interface

The web interface is perfect for:
- **AI-Human collaboration**: Monitor AI progress in real-time while maintaining visual oversight
- **Visual project overview**: See the big picture of your project's task flow
- **Quick status updates**: Use drag-and-drop for instant task status changes
- **Task details**: Enhanced cards show when tasks were created and last modified
- **Copying tasks**: Share formatted task information with AI assistants
- **Multi-project workflows**: Manage multiple projects from one central interface
