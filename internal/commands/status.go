package commands

import (
	"encoding/json"
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

// setTaskStatusCmd represents the set-task-status command
var setTaskStatusCmd = &cobra.Command{
	Use:   "set-task-status <id> <status>",
	Short: "Update task status",
	Long: `Update the status of a task by ID.

Valid statuses: pending, in_progress, done

Examples:
  quicktodo set-task-status 1 in_progress
  quicktodo set-task-status 5 done
  quicktodo set-task-status 3 pending`,
	Args: cobra.ExactArgs(2),
	Run:  runSetTaskStatus,
}

// markCompletedCmd represents the mark-completed command
var markCompletedCmd = &cobra.Command{
	Use:     "mark-completed <id>",
	Aliases: []string{"mark-done"},
	Short:   "Mark task as completed",
	Long: `Mark a task as done/completed.

Examples:
  quicktodo mark-completed 1
  quicktodo mark-done 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSetTaskStatusWithValue(args[0], "done")
	},
}

// markInProgressCmd represents the mark-in-progress command
var markInProgressCmd = &cobra.Command{
	Use:   "mark-in-progress <id>",
	Short: "Mark task as in progress",
	Long: `Mark a task as in progress.

Examples:
  quicktodo mark-in-progress 1
  quicktodo mark-in-progress 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSetTaskStatusWithValue(args[0], "in_progress")
	},
}

// markPendingCmd represents the mark-pending command
var markPendingCmd = &cobra.Command{
	Use:   "mark-pending <id>",
	Short: "Mark task as pending",
	Long: `Mark a task as pending.

Examples:
  quicktodo mark-pending 1
  quicktodo mark-pending 5`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		runSetTaskStatusWithValue(args[0], "pending")
	},
}

func runSetTaskStatus(cmd *cobra.Command, args []string) {
	taskIDStr := args[0]
	newStatus := strings.ToLower(args[1])
	
	runSetTaskStatusWithValue(taskIDStr, newStatus)
}

func runSetTaskStatusWithValue(taskIDStr, newStatus string) {
	// Parse task ID
	taskID, err := strconv.Atoi(taskIDStr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s'. Task ID must be a number.\n", taskIDStr)
		os.Exit(1)
	}

	if taskID <= 0 {
		fmt.Fprintf(os.Stderr, "Error: task ID must be positive\n")
		os.Exit(1)
	}

	// Validate status
	status := models.Status(newStatus)
	if !models.IsValidStatus(string(status)) {
		fmt.Fprintf(os.Stderr, "Error: invalid status '%s'. Valid statuses: pending, in_progress, done\n", newStatus)
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
		fmt.Fprintf(os.Stderr, "Run 'quicktodo init' first\n")
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

	// Store old status for output
	oldStatus := task.Status

	// Update task status
	if err := task.UpdateStatus(status); err != nil {
		fmt.Fprintf(os.Stderr, "Error updating task status: %v\n", err)
		os.Exit(1)
	}

	// Update task in database
	if err := projectDB.UpdateTask(task); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving task: %v\n", err)
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

	// Sync to TODO list if enabled
	syncToTodoList(task, projectInfo.Name, "status", cfg)

	// Notify web server of task update
	if err := notify.NotifyTaskUpdated(cfg, task, projectInfo.Name); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to notify web server: %v\n", err)
	}

	// Output result
	if jsonOutput {
		outputStatusChangeJSON(task, string(oldStatus), projectInfo)
	} else {
		outputStatusChangeHuman(task, string(oldStatus), projectInfo)
	}
}

func outputStatusChangeJSON(task *models.Task, oldStatus string, projectInfo *database.ProjectInfo) {
	output := map[string]interface{}{
		"success": true,
		"project": map[string]interface{}{
			"name": projectInfo.Name,
			"path": projectInfo.Path,
		},
		"task":        task,
		"old_status":  oldStatus,
		"new_status":  task.Status,
		"changed_at":  task.UpdatedAt,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func outputStatusChangeHuman(task *models.Task, oldStatus string, projectInfo *database.ProjectInfo) {
	statusIcon := getStatusIcon(task.Status)
	
	fmt.Printf("%s Task #%d status changed: %s → %s\n", 
		statusIcon, task.ID, oldStatus, task.Status)
	fmt.Printf("Title: %s\n", task.Title)
	
	if verbose {
		fmt.Printf("Project: %s\n", projectInfo.Name)
		fmt.Printf("Updated: %s\n", task.UpdatedAt.Format("2006-01-02 15:04:05"))
	}
}

func init() {
	RootCmd.AddCommand(setTaskStatusCmd)
	RootCmd.AddCommand(markCompletedCmd)
	RootCmd.AddCommand(markInProgressCmd)
	RootCmd.AddCommand(markPendingCmd)
}