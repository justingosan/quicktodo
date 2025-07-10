# QuickTodo

A simple todo list CLI built specifically for AI coding workflows. 

## Why QuickTodo?

- **Built for AI**: Designed from the ground up for AI-assisted development
- **Lightweight alternative**: Simple alternative to taskmaster.dev which can feel bloated
- **Zero setup**: Just run `init` and start coding with your AI

## Installation

```bash
# Install from source
git clone https://github.com/yourusername/quicktodo.git
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
This creates a `QUICKTODO.md` file with AI usage instructions.

### 2. Configure your AI
Add this to your `.cursorrules` or `CLAUDE.md`:
```
Refer to ./QUICKTODO.md for task management instructions.
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

## Why not taskmaster.dev?

TaskMaster is powerful but can be overwhelming for simple AI workflows. QuickTodo focuses on:
- **Simplicity**: One command to get started
- **Speed**: Minimal overhead, maximum coding time
- **AI-first**: Every feature designed for AI interaction

Perfect for when you want to let AI loose on your codebase without complex project management overhead.