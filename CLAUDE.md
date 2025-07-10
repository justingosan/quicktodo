Before starting any task, use the quicktodo task management tool. If you don't know how to use it, run: quicktodo context

# important-instruction-reminders
Do what has been asked; nothing more, nothing less.
NEVER create files unless they're absolutely necessary for achieving your goal.
ALWAYS prefer editing an existing file to creating a new one.
NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.

# hooks-setup
Claude Code is configured with PreToolUse hooks that automatically run `quicktodo context` before any tool use to ensure these instructions are always available. This is more reliable than depending on Claude to remember CLAUDE.md contents throughout long conversations.
