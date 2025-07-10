# QuickTodo Technical Requirements Document

## Project Overview
QuickTodo is a simple CLI todo management tool designed for AI-assisted development workflows. It provides a lightweight, fast, and AI-friendly interface for managing tasks across multiple projects.

## Core Requirements

### Technology Stack
- **Language**: Go (for single binary compilation)
- **Data Storage**: JSON file-based storage
- **Configuration**: File-based configuration in user home directory
- **Dependencies**: Minimal external dependencies

### File Structure
```
~/.config/quicktodo/
├── config.json          # Global configuration
├── projects.json        # Project registry (path -> project mapping)
├── projects/            # Individual project databases
│   ├── project1.json    # Tasks for project1
│   ├── project2.json    # Tasks for project2
│   └── ...
└── locks/               # Lock files for concurrent access
    ├── project1.lock
    ├── project2.lock
    └── ...
```

### Database Schema

**projects.json** (Project Registry)
```json
{
  "projects": {
    "project_name": {
      "path": "/absolute/path/to/project",
      "name": "project_name",
      "created_at": "2025-01-01T00:00:00Z",
      "last_accessed": "2025-01-01T00:00:00Z"
    }
  },
  "path_to_project": {
    "/absolute/path/to/project": "project_name"
  }
}
```

**projects/project_name.json** (Individual Project Tasks)
```json
{
  "project_name": "project_name",
  "project_path": "/absolute/path/to/project",
  "tasks": [
    {
      "id": 1,
      "title": "Task title",
      "description": "Task description",
      "status": "pending|in_progress|done",
      "priority": "low|medium|high",
      "created_at": "2025-01-01T00:00:00Z",
      "updated_at": "2025-01-01T00:00:00Z",
      "assigned_to": "ai_agent_id|user", 
      "locked_by": "process_id",
      "locked_at": "2025-01-01T00:00:00Z"
    }
  ],
  "next_id": 2,
  "last_modified": "2025-01-01T00:00:00Z",
  "version": 1
}
```

## AI Development Tasks

### Phase 1: Core Infrastructure

**Task 1.1: Project Setup** ✅ COMPLETED
```bash
# Create a new Go project for QuickTodo CLI tool
# - Initialize Go module: go mod init quicktodo
# - Set up basic project structure with cmd/, internal/, pkg/ directories
# - Create main.go with basic CLI entry point
# - Add .gitignore for Go projects
# - Create basic README.md with project description
```

**Task 1.2: CLI Framework Setup** ✅ COMPLETED
```bash
# Set up Cobra CLI framework
# - Add github.com/spf13/cobra dependency
# - Create cmd/root.go with basic command structure
# - Add version command and global flags
# - Set up proper command help and usage text
# - Add basic logging configuration
```

**Task 1.3: Configuration Management** ✅ COMPLETED
```bash
# Implement configuration system
# - Create internal/config/config.go
# - Add ~/.config/quicktodo directory creation
# - Implement config.json reading/writing with defaults
# - Add config validation and error handling
# - Create config initialization on first run
```

**Task 1.4: Database Foundation** ✅ COMPLETED
```bash
# Create database layer foundation
# - Create internal/database/registry.go for projects.json management
# - Create internal/database/project.go for individual project databases
# - Add internal/database/locks.go for file locking mechanism
# - Implement atomic file operations with temp files
# - Add database directory initialization
```

### Phase 2: Basic Task Management

**Task 2.1: Data Models** ✅ COMPLETED
```bash
# Create data models and structures
# - Define Task struct in internal/models/task.go with all fields
# - Define Project struct in internal/models/project.go
# - Add JSON serialization tags and validation
# - Create model helper functions (New, Validate, etc.)
# - Add time handling utilities
```

**Task 2.2: Project Registry** ✅ COMPLETED
```bash
# Implement project registry system
# - Create projects.json management functions
# - Add project registration and lookup by path
# - Implement project auto-detection from current directory
# - Add project cleanup and maintenance functions
# - Handle project registry corruption gracefully
```

**Task 2.3: File Locking System** ✅ COMPLETED
```bash
# Implement concurrent access protection
# - Add github.com/gofrs/flock dependency
# - Create lock acquisition and release functions
# - Implement lock timeout with configurable duration
# - Add stale lock cleanup mechanism
# - Create process ID tracking in lock files
```

**Task 2.4: Project Database Operations** ✅ COMPLETED
```bash
# Implement individual project database operations
# - Create project database file creation/initialization
# - Add task CRUD operations with locking
# - Implement atomic updates with version checking
# - Add database backup and restore functionality
# - Handle database corruption and recovery
```

### Phase 3: Core Commands

**Task 3.1: Initialize Project Command** ✅ COMPLETED
```bash
# Create project initialization command
# - Add 'quicktodo init [project_name]' command (simplified from initialize-project)
# - Register current directory as project in registry
# - Create project database file
# - Generate QUICKTODO.md with AI usage instructions
# - Add validation for existing projects
# - Support custom project names or auto-generate from directory
```

**Task 3.2: Create Task Command** ✅ COMPLETED
```bash
# Create task creation command
# - Add 'quicktodo create-task "title" [--description] [--priority]' command
# - Also support alias 'quicktodo new-task' for convenience
# - Auto-detect current project from working directory
# - Validate input parameters
# - Generate unique task IDs
# - Add task to project database with proper locking
```

**Task 3.3: List Tasks Command** ✅ COMPLETED
```bash
# Create task listing command
# - Add 'quicktodo list-tasks [--status] [--priority] [--json]' command
# - Also support alias 'quicktodo show-tasks' for clarity
# - Display tasks in human-readable format
# - Add filtering by status, priority
# - Support JSON output for AI consumption
# - Show project information and task counts
```

**Task 3.4: Display Task Command** ✅ COMPLETED
```bash
# Create individual task display command
# - Add 'quicktodo display-task <id>' command
# - Also support alias 'quicktodo get-task' for API-style naming
# - Display detailed task information
# - Support JSON output format
# - Add task not found error handling
# - Show task metadata (created, updated, etc.)
```

### Phase 4: Task Operations

**Task 4.1: Status Management Commands**
```bash
# Create task status update commands
# - Add 'quicktodo mark-completed <id>' command (also alias 'mark-done')
# - Add 'quicktodo mark-in-progress <id>' command (also alias 'start-task')
# - Add 'quicktodo update-status <id> <status>' command
# - Add 'quicktodo mark-pending <id>' command to reset to pending
# - Validate status transitions
# - Update task timestamps on status changes
```

**Task 4.2: Task Editing Commands**
```bash
# Create task modification commands
# - Add 'quicktodo edit-task <id> [--title] [--description] [--priority]' command
# - Also support alias 'quicktodo update-task' for API-style naming
# - Support partial updates (only specified fields)
# - Add input validation
# - Preserve task history and timestamps
# - Handle concurrent edit conflicts
```

**Task 4.3: Task Deletion Commands**
```bash
# Create task deletion command
# - Add 'quicktodo delete-task <id>' command
# - Also support alias 'quicktodo remove-task' for clarity
# - Add confirmation prompt (unless --force)
# - Soft delete option for recovery
# - Update task counters
# - Handle deletion of non-existent tasks
```

**Task 4.4: Project Management Commands**
```bash
# Create project management commands
# - Add 'quicktodo list-projects' command to list all projects
# - Add 'quicktodo cleanup-projects' command to remove orphaned projects
# - Add 'quicktodo project-status' command to show current project info
# - Add 'quicktodo project-summary' command for detailed project overview
# - Add project path resolution and validation
# - Handle missing or moved project directories
```

### Phase 5: Advanced Features

**Task 5.1: Search and Query Commands**
```bash
# Create search functionality
# - Add 'quicktodo search-tasks <query>' command
# - Also support alias 'quicktodo find-tasks' for clarity
# - Search in task titles and descriptions
# - Support multiple search terms
# - Add case-insensitive searching
# - Support JSON output for search results
```

**Task 5.2: Batch Operations Commands**
```bash
# Create batch operation support
# - Add 'quicktodo execute-batch <command_file>' command
# - Also support alias 'quicktodo batch-operations' for clarity
# - Add 'quicktodo bulk-update-status <status> <id1,id2,id3>' command
# - Support multiple operations in single invocation
# - Implement transaction-like behavior
# - Add progress reporting for large batches
```

**Task 5.3: Output Format Commands**
```bash
# Add comprehensive JSON output support
# - Add 'quicktodo export-json [--filter]' command for full JSON export
# - Add --json flag to all existing commands
# - Add 'quicktodo validate-data' command to check data integrity
# - Standardize JSON response format
# - Include metadata in JSON responses
# - Add machine-readable error responses
```

**Task 5.4: Import/Export Commands**
```bash
# Create data import/export functionality
# - Add 'quicktodo export-data [--format json|csv]' command
# - Add 'quicktodo import-data <file>' command
# - Add 'quicktodo backup-project [project_name]' command
# - Add 'quicktodo restore-project <backup_file>' command
# - Support multiple file formats
# - Add data validation on import
# - Create backup before import operations
```

### Phase 6: Polish and Documentation

**Task 6.1: Error Handling**
```bash
# Improve error handling and user experience
# - Standardize error messages and exit codes
# - Add helpful suggestions for common errors
# - Implement graceful handling of corrupted data
# - Add verbose mode for debugging
# - Create error recovery mechanisms
```

**Task 6.2: Testing Suite**
```bash
# Create comprehensive test suite
# - Add unit tests for all core functions
# - Create integration tests for command workflows
# - Add concurrent access tests
# - Test data corruption scenarios
# - Add performance benchmarks
```

**Task 6.3: Build and Distribution**
```bash
# Set up build and distribution
# - Create cross-platform build scripts
# - Add version embedding in binary
# - Create installation instructions
# - Add shell completion support
# - Test on multiple platforms
```

**Task 6.4: AGENTS.md Documentation**
```bash
# Create comprehensive AI agent documentation
# - Document all commands with examples
# - Add JSON output examples for each command
# - Create workflow patterns for common tasks
# - Add error handling examples
# - Include best practices for AI usage
```

## Task Assignment Guidelines

### For AI Agents:
- Each task should be completed independently
- Always include comprehensive error handling
- Add unit tests for new functionality
- Follow Go best practices and conventions
- Document all public functions and types

### Prerequisites:
- Task 1.1 must be completed before any other tasks
- Phase 1 must be completed before Phase 2
- Core commands (Phase 3) should be completed before advanced features
- Testing should be added incrementally with each feature

### Validation Criteria:
- All code must compile without warnings
- Unit tests must pass
- Commands must work from any directory
- Concurrent access must be safe
- Error messages must be helpful and actionable

## Command Reference Summary

### Primary Commands (verbose, descriptive)
- `quicktodo init [name]` - Initialize current directory as QuickTodo project
- `quicktodo create-task "title"` - Add new task to current project
- `quicktodo list-tasks [--status] [--priority]` - Show all tasks with filters
- `quicktodo display-task <id>` - Show detailed task information
- `quicktodo mark-completed <id>` - Mark task as done
- `quicktodo mark-in-progress <id>` - Start working on task
- `quicktodo update-status <id> <status>` - Change task status
- `quicktodo edit-task <id> [--title] [--description]` - Modify task details
- `quicktodo delete-task <id>` - Remove task from project
- `quicktodo search-tasks <query>` - Find tasks by text search
- `quicktodo list-projects` - Show all registered projects
- `quicktodo project-status` - Show current project information
- `quicktodo export-data [--format json|csv]` - Export tasks to file
- `quicktodo import-data <file>` - Import tasks from file

### Convenient Aliases (shorter alternatives)
- `quicktodo new-task` → `create-task`
- `quicktodo show-tasks` → `list-tasks`
- `quicktodo get-task` → `display-task`
- `quicktodo mark-done` → `mark-completed`
- `quicktodo start-task` → `mark-in-progress`
- `quicktodo update-task` → `edit-task`
- `quicktodo remove-task` → `delete-task`
- `quicktodo find-tasks` → `search-tasks`

### Advanced Commands
- `quicktodo execute-batch <file>` - Run multiple commands from file
- `quicktodo bulk-update-status <status> <id1,id2,id3>` - Update multiple tasks
- `quicktodo cleanup-projects` - Remove orphaned/invalid projects
- `quicktodo project-summary` - Detailed project overview with statistics
- `quicktodo backup-project [name]` - Create project backup
- `quicktodo restore-project <backup>` - Restore from backup
- `quicktodo validate-data` - Check data integrity
- `quicktodo export-json [--filter]` - Export with JSON formatting

### File Locking Mechanism
- **Lock Files**: Create `.lock` files for each project during operations
- **Lock Timeout**: 30-second timeout for acquiring locks
- **Stale Lock Cleanup**: Automatic cleanup of locks older than 5 minutes
- **Retry Logic**: Exponential backoff for lock acquisition
- **Process ID Tracking**: Store process ID in lock files for debugging

### Database Versioning
- **Version Field**: Each project database has a version number
- **Optimistic Locking**: Check version before writing, increment on success
- **Conflict Resolution**: Fail fast with clear error messages on conflicts
- **Atomic Operations**: Use temporary files + atomic rename for consistency

### AI Agent Coordination
- **Agent Identification**: Optional `--agent-id` flag for AI agents
- **Task Assignment**: Track which agent is working on which tasks
- **Status Visibility**: Show who is working on what in list commands
- **Graceful Failures**: Clear error messages when locks can't be acquired

### Command Behavior
- **Auto-Retry**: Automatically retry failed operations due to locks
- **Timeout Handling**: Graceful timeout with helpful error messages
- **Status Reporting**: Show lock status in verbose mode
- **Emergency Override**: `--force` flag to break locks (with warnings)

## Technical Implementation Details

### Go Project Structure
```
quicktodo/
├── cmd/
│   └── root.go              # Main CLI setup
├── internal/
│   ├── config/
│   │   └── config.go        # Configuration management
│   ├── database/
│   │   ├── registry.go      # Project registry operations
│   │   ├── project.go       # Individual project database
│   │   └── locks.go         # File locking mechanism
│   ├── models/
│   │   ├── task.go          # Task model
│   │   └── project.go       # Project model
│   ├── commands/
│   │   ├── add.go           # Add command
│   │   ├── list.go          # List command
│   │   ├── init.go          # Project initialization
│   │   └── ...              # Other commands
│   └── utils/
│       ├── helpers.go       # Utility functions
│       └── path.go          # Path resolution utilities
├── pkg/
│   └── quicktodo/
│       └── client.go        # Public API (future)
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### Key Dependencies
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- Standard library for JSON operations
- `github.com/fatih/color` - Colored output (optional)
- `github.com/gofrs/flock` - File locking for concurrent access

### Error Handling Strategy
- Consistent error message format
- Proper exit codes for different error types
- Graceful handling of corrupted data
- Lock acquisition timeout handling
- Concurrent access conflict resolution
- User-friendly error messages with suggestions

### Performance Considerations
- Lazy loading of project databases
- Efficient JSON parsing for large task lists
- Minimal lock duration to reduce contention
- Fast project auto-detection
- Optimized search operations across projects

## Testing Strategy

### Unit Tests
- Test all core functions
- Database operations testing
- Command parsing validation
- Error handling verification

### Integration Tests
- End-to-end command testing
- File system operations
- Multi-project scenarios
- Concurrent access testing with multiple processes
- Lock acquisition and release testing
- Data consistency under concurrent load
- Project auto-detection scenarios

### Performance Tests
- Large dataset handling across multiple projects
- Memory usage profiling
- Startup time optimization
- Concurrent access performance
- Lock contention measurement
- Project auto-detection speed

## Deployment and Distribution

### Build Process
- Cross-platform compilation (Linux, macOS, Windows)
- Single binary output
- Version embedding
- Build automation

### Installation Methods
- Direct binary download
- Package managers (homebrew, apt, etc.)
- Go install support
- Container image (optional)

## Documentation Requirements

### User Documentation
- README with installation instructions
- Command reference documentation
- Usage examples and tutorials
- Troubleshooting guide

### AI Integration Documentation (AGENTS.md)
- Complete command reference
- JSON output examples
- Common workflow patterns
- Best practices for AI agents
- Error handling for automated usage

## Security Considerations
- Safe file operations (atomic writes)
- Input validation and sanitization
- No sensitive data in plain text
- Secure temporary file handling

## Future Extensibility
- Plugin architecture consideration
- API endpoint potential
- Integration with external tools
- Export format extensibility

## Success Criteria
- Single binary under 10MB
- Startup time under 100ms
- Handles 10,000+ tasks efficiently
- 100% CLI coverage for AI usage
- Comprehensive error handling
- Cross-platform compatibility
