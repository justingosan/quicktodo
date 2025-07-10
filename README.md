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
For task management, run: quicktodo context
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
quicktodo mark-completed 1               # Mark task done
```

All commands support `--json` for AI consumption.
