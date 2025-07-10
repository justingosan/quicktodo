package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"strings"

	"github.com/spf13/cobra"
)

var (
	taskDescription string
	taskPriority    string
)

// createTaskCmd represents the create-task command
var createTaskCmd = &cobra.Command{
	Use:     "create-task <title>",
	Aliases: []string{"new-task"},
	Short:   "Add new task to current project",
	Long: `Create a new task in the current project with the specified title.

The command will auto-detect the current project from the working directory.
You can optionally specify a description and priority for the task.

Examples:
  quicktodo create-task "Implement user authentication"
  quicktodo new-task "Fix login bug" --description "Users can't log in with email" --priority high
  quicktodo create-task "Write documentation" --priority low`,
	Args: cobra.ExactArgs(1),
	Run:  runCreateTask,
}

func runCreateTask(cmd *cobra.Command, args []string) {
	title := strings.TrimSpace(args[0])
	if title == "" {
		fmt.Fprintf(os.Stderr, "Error: task title cannot be empty\n")
		os.Exit(1)
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Load project registry
	registryPath := cfg.GetProjectsPath()
	registry, err := database.LoadProjectRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project registry: %v\n", err)
		os.Exit(1)
	}

	// Find project for current directory
	projectInfo, exists := registry.GetProjectByPath(currentDir)
	if !exists {
		fmt.Fprintf(os.Stderr, "Error: current directory is not a registered project\n")
		fmt.Fprintf(os.Stderr, "Run 'quicktodo initialize-project' first\n")
		os.Exit(1)
	}

	// Update last accessed time
	if err := registry.UpdateLastAccessed(projectInfo.Name); err != nil {
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to update last accessed time: %v\n", err)
		}
	}

	// Validate priority
	priority := models.Priority(strings.ToLower(taskPriority))
	if taskPriority != "" && !models.IsValidPriority(string(priority)) {
		fmt.Fprintf(os.Stderr, "Error: invalid priority '%s'. Valid priorities: low, medium, high\n", taskPriority)
		os.Exit(1)
	}

	if taskPriority == "" {
		priority = models.Priority(cfg.DefaultPriority)
	}

	// Create lock manager
	lockManager := database.NewLockManager(cfg.DataDir+"/locks", cfg.LockTimeout)

	// Acquire lock for project
	lockInfo, err := lockManager.AcquireLock(projectInfo.Name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error acquiring project lock: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if err := lockManager.ReleaseLock(lockInfo); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to release lock: %v\n", err)
		}
	}()

	// Load project database
	dbPath := cfg.GetProjectDatabasePath(projectInfo.Name)
	projectDB, err := loadProjectDatabase(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project database: %v\n", err)
		os.Exit(1)
	}

	// Create new task
	task := models.NewTaskWithDetails(projectDB.NextID, title, taskDescription, priority)

	// Assign to agent if specified
	if agentID != "" {
		task.AssignTo(agentID)
	}

	// Add task to database
	if err := projectDB.AddTask(task); err != nil {
		fmt.Fprintf(os.Stderr, "Error adding task: %v\n", err)
		os.Exit(1)
	}

	// Save project database
	if err := saveProjectDatabase(projectDB, dbPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving project database: %v\n", err)
		os.Exit(1)
	}

	// Save updated registry
	if err := registry.Save(registryPath); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save registry: %v\n", err)
	}

	// Output result
	if jsonOutput {
		outputTaskJSON(task)
	} else {
		fmt.Printf("Created task #%d: %s\n", task.ID, task.Title)
		if verbose {
			fmt.Printf("Project: %s\n", projectInfo.Name)
			fmt.Printf("Priority: %s\n", task.Priority)
			fmt.Printf("Status: %s\n", task.Status)
			if task.Description != "" {
				fmt.Printf("Description: %s\n", task.Description)
			}
			if task.AssignedTo != "" {
				fmt.Printf("Assigned to: %s\n", task.AssignedTo)
			}
		}
	}
}

func loadProjectDatabase(filePath string) (*models.ProjectDatabase, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project database file does not exist: %s", filePath)
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project database: %w", err)
	}

	// Unmarshal JSON
	var db models.ProjectDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to parse project database: %w", err)
	}

	// Validate database
	if err := db.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project database: %w", err)
	}

	return &db, nil
}

func outputTaskJSON(task *models.Task) {
	output := map[string]interface{}{
		"success": true,
		"task":    task,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func init() {
	createTaskCmd.Flags().StringVarP(&taskDescription, "description", "d", "", "Task description")
	createTaskCmd.Flags().StringVarP(&taskPriority, "priority", "p", "", "Task priority (low, medium, high)")

	RootCmd.AddCommand(createTaskCmd)
}
