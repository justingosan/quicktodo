package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"strings"

	"github.com/spf13/cobra"
)

// initProjectCmd represents the init command
var initProjectCmd = &cobra.Command{
	Use:   "init [project_name]",
	Short: "Initialize current directory as a QuickTodo project",
	Long: `Initialize the current directory as a QuickTodo project. 

This command:
- Creates a new project entry in the registry
- Initializes the project database
- Generates QUICKTODO.md with AI usage instructions

If no project name is provided, it will use the current directory name.

Examples:
  quicktodo init myproject
  quicktodo init
  quicktodo init "My Amazing Project"`,
	Args: cobra.MaximumNArgs(1),
	Run:  runInitProject,
}

func runInitProject(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure all directories exist
	if err := cfg.EnsureAllDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Determine project name
	var projectName string
	if len(args) > 0 {
		projectName = strings.TrimSpace(args[0])
	} else {
		projectName = filepath.Base(currentDir)
	}

	if projectName == "" {
		fmt.Fprintf(os.Stderr, "Error: project name cannot be empty\n")
		os.Exit(1)
	}

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load project registry
	registryPath := cfg.GetProjectsPath()
	registry, err := database.LoadProjectRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project registry: %v\n", err)
		os.Exit(1)
	}

	// Check if project already exists
	if _, exists := registry.GetProjectByName(projectName); exists {
		fmt.Fprintf(os.Stderr, "Error: project '%s' already exists\n", projectName)
		os.Exit(1)
	}

	// Check if current directory is already registered
	if existingProject, exists := registry.GetProjectByPath(currentDir); exists {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' is already registered as project '%s'\n", 
			currentDir, existingProject.Name)
		os.Exit(1)
	}

	// Register project
	if err := registry.RegisterProject(projectName, currentDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering project: %v\n", err)
		os.Exit(1)
	}

	// Save updated registry
	if err := registry.Save(registryPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving project registry: %v\n", err)
		os.Exit(1)
	}

	// Create project database
	project := models.NewProject(projectName, currentDir)
	projectDB := models.NewProjectDatabase(project)

	// Save project database
	dbPath := cfg.GetProjectDatabasePath(projectName)
	if err := saveProjectDatabase(projectDB, dbPath); err != nil {
		// Try to rollback registry change
		registry.RemoveProject(projectName)
		registry.Save(registryPath)
		fmt.Fprintf(os.Stderr, "Error creating project database: %v\n", err)
		os.Exit(1)
	}

	// Copy QUICKTODO.md file to current directory
	quicktodoPath := filepath.Join(currentDir, "QUICKTODO.md")
	if err := copyQuickTodoMD(quicktodoPath); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to create QUICKTODO.md: %v\n", err)
	}

	// Output success message
	fmt.Printf("Successfully initialized project '%s' in directory '%s'\n", projectName, currentDir)
	fmt.Printf("Generated QUICKTODO.md with AI usage instructions\n")
	if verbose {
		fmt.Printf("Project database: %s\n", dbPath)
		fmt.Printf("Registry: %s\n", registryPath)
		fmt.Printf("Documentation: %s\n", quicktodoPath)
	}
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("project name cannot be longer than 100 characters")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("project name cannot contain '%s'", char)
		}
	}

	return nil
}

func saveProjectDatabase(db *models.ProjectDatabase, filePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert to JSON
	data, err := db.ToJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal database: %w", err)
	}

	// Write to temporary file first, then rename for atomicity
	tempPath := filePath + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write temporary file: %w", err)
	}

	if err := os.Rename(tempPath, filePath); err != nil {
		os.Remove(tempPath) // Clean up temp file
		return fmt.Errorf("failed to rename temporary file: %w", err)
	}

	return nil
}

func copyQuickTodoMD(targetPath string) error {
	// Read the existing QUICKTODO.md from the source
	content, err := os.ReadFile("QUICKTODO.md")
	if err != nil {
		// If we can't find the source file, create a basic version
		content = []byte(`# QuickTodo Usage Guide

This project uses QuickTodo for task management. This document provides concise instructions for AI agents.

## Quick Reference

### Essential Commands
` + "```bash" + `
# Initialize project
quicktodo init

# Create a new task
quicktodo create-task "Task title" --description "Optional description" --priority high|medium|low

# List all tasks
quicktodo list-tasks

# List tasks with filters
quicktodo list-tasks --status pending|in_progress|done
quicktodo list-tasks --priority high|medium|low

# Show detailed task information
quicktodo display-task <id>

# Get JSON output (AI-friendly)
quicktodo list-tasks --json
quicktodo display-task <id> --json
` + "```" + `

### Task Status Values
- **pending** - Task not started (default)
- **in_progress** - Task currently being worked on
- **done** - Task completed

### Priority Values
- **high** - Urgent/important tasks
- **medium** - Normal priority (default)
- **low** - Nice-to-have tasks

## Best Practices for AI Agents

1. **Always use --json flag** for programmatic access
2. **Check exit codes** to detect errors
3. **Use descriptive titles** for better task management
4. **Set appropriate priority** based on task importance
5. **Include --agent-id** to track AI assignments

## Troubleshooting

- **"not a registered project"** - Run ` + "`quicktodo init`" + ` first
- **Lock timeout errors** - Another process is using the project, wait and retry
- **Task not found** - Check task ID with ` + "`quicktodo list-tasks`" + `

---
*Generated by QuickTodo v1.0.0*`)
	}
	
	return os.WriteFile(targetPath, content, 0644)
}

func init() {
	RootCmd.AddCommand(initProjectCmd)
}