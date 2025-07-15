package commands

import (
	"fmt"
	"os"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"quicktodo/internal/notify"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	editTitle       string
	editDescription string
	editPriority    string
)

// editTaskCmd represents the edit-task command
var editTaskCmd = &cobra.Command{
	Use:     "edit-task <id>",
	Aliases: []string{"edit"},
	Short:   "Edit an existing task",
	Long: `Edit an existing task's title, description, or priority.

You can specify which fields to update using the flags. If no flags are provided,
the command will show the current task details.

Examples:
  quicktodo edit-task 1 --title "Updated task title"
  quicktodo edit 2 --description "New description"
  quicktodo edit-task 3 --priority high
  quicktodo edit 4 --title "New title" --description "New description" --priority medium`,
	Args: cobra.ExactArgs(1),
	Run:  runEditTask,
}

func runEditTask(cmd *cobra.Command, args []string) {
	// Parse task ID
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s'\n", args[0])
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

	// Find task
	task, err := projectDB.GetTask(taskID)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: task #%d not found\n", taskID)
		os.Exit(1)
	}

	// Check if any edit flags were provided
	hasUpdates := editTitle != "" || editDescription != "" || editPriority != ""
	if !hasUpdates {
		// No updates requested, just show current task details
		if jsonOutput {
			outputTaskJSON(task)
		} else {
			fmt.Printf("Task #%d: %s\n", task.ID, task.Title)
			if task.Description != "" {
				fmt.Printf("Description: %s\n", task.Description)
			}
			fmt.Printf("Priority: %s\n", task.Priority)
			fmt.Printf("Status: %s\n", task.Status)
		}
		return
	}

	// Update task fields
	updated := false

	if editTitle != "" {
		task.Title = strings.TrimSpace(editTitle)
		updated = true
	}

	if editDescription != "" {
		task.Description = strings.TrimSpace(editDescription)
		updated = true
	}

	if editPriority != "" {
		priority := models.Priority(strings.ToLower(editPriority))
		if !models.IsValidPriority(string(priority)) {
			fmt.Fprintf(os.Stderr, "Error: invalid priority '%s'. Valid priorities: low, medium, high\n", editPriority)
			os.Exit(1)
		}
		task.Priority = priority
		updated = true
	}

	if updated {
		// Save project database
		if err := saveProjectDatabase(projectDB, dbPath); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving project database: %v\n", err)
			os.Exit(1)
		}

		// Save updated registry
		if err := registry.Save(registryPath); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to save registry: %v\n", err)
		}

		// Sync to TODO list if enabled
		syncToTodoList(task, projectInfo.Name, "edit", cfg)

		// Notify web server of task update
		if err := notify.NotifyTaskUpdated(cfg, task, projectInfo.Name); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to notify web server: %v\n", err)
		}

		// Output result
		if jsonOutput {
			outputTaskJSON(task)
		} else {
			fmt.Printf("Updated task #%d: %s\n", task.ID, task.Title)
			if verbose {
				fmt.Printf("Project: %s\n", projectInfo.Name)
				fmt.Printf("Priority: %s\n", task.Priority)
				fmt.Printf("Status: %s\n", task.Status)
				if task.Description != "" {
					fmt.Printf("Description: %s\n", task.Description)
				}
			}
		}
	}
}

func init() {
	editTaskCmd.Flags().StringVarP(&editTitle, "title", "t", "", "New task title")
	editTaskCmd.Flags().StringVarP(&editDescription, "description", "d", "", "New task description")
	editTaskCmd.Flags().StringVarP(&editPriority, "priority", "p", "", "New task priority (low, medium, high)")

	RootCmd.AddCommand(editTaskCmd)
}