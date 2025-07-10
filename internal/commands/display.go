package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// displayTaskCmd represents the display-task command
var displayTaskCmd = &cobra.Command{
	Use:     "display-task <id>",
	Aliases: []string{"get-task"},
	Short:   "Show detailed task information",
	Long: `Display detailed information about a specific task by ID.

The command will auto-detect the current project from the working directory
and show comprehensive task details including metadata, timestamps, and status.

Examples:
  quicktodo display-task 1
  quicktodo get-task 5 --json
  quicktodo display-task 3 --verbose`,
	Args: cobra.ExactArgs(1),
	Run:  runDisplayTask,
}

func runDisplayTask(cmd *cobra.Command, args []string) {
	// Parse task ID
	taskID, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: invalid task ID '%s'. Task ID must be a number.\n", args[0])
		os.Exit(1)
	}

	if taskID <= 0 {
		fmt.Fprintf(os.Stderr, "Error: task ID must be positive\n")
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

	// Save updated registry (for last accessed time)
	if err := registry.Save(registryPath); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save registry: %v\n", err)
	}

	// Output result
	if jsonOutput {
		outputTaskDetailJSON(task, projectInfo)
	} else {
		outputTaskDetailHuman(task, projectInfo)
	}
}

func outputTaskDetailJSON(task *models.Task, projectInfo *database.ProjectInfo) {
	output := map[string]interface{}{
		"success": true,
		"project": map[string]interface{}{
			"name": projectInfo.Name,
			"path": projectInfo.Path,
		},
		"task": task,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func outputTaskDetailHuman(task *models.Task, projectInfo *database.ProjectInfo) {
	// Header
	statusIcon := getStatusIcon(task.Status)
	priorityColor := getPriorityIndicator(task.Priority)

	fmt.Printf("%s Task #%d\n", statusIcon, task.ID)
	fmt.Printf("Title: %s%s\n", priorityColor, task.Title)

	if task.Description != "" {
		fmt.Printf("Description: %s\n", task.Description)
	}

	fmt.Printf("Status: %s\n", task.Status)
	fmt.Printf("Priority: %s\n", task.Priority)

	// Timestamps
	fmt.Printf("Created: %s (%s)\n",
		task.CreatedAt.Format("2006-01-02 15:04:05"),
		task.GetAge())

	if !task.UpdatedAt.Equal(task.CreatedAt) {
		fmt.Printf("Updated: %s (%s)\n",
			task.UpdatedAt.Format("2006-01-02 15:04:05"),
			formatTimeAgo(task.UpdatedAt))
	}

	// Assignment and locking
	if task.AssignedTo != "" {
		fmt.Printf("Assigned to: %s\n", task.AssignedTo)
	}

	if task.IsLocked() {
		fmt.Printf("Locked by: %s\n", task.LockedBy)
		fmt.Printf("Locked at: %s (%s)\n",
			task.LockedAt.Format("2006-01-02 15:04:05"),
			formatTimeAgo(task.LockedAt))

		if task.IsStale() {
			fmt.Printf("🟠 Warning: Lock appears to be stale\n")
		}
	}

	// Project info
	fmt.Printf("\nProject: %s\n", projectInfo.Name)
	if verbose {
		fmt.Printf("Project path: %s\n", projectInfo.Path)
	}

	// Additional metadata in verbose mode
	if verbose {
		fmt.Printf("\nMetadata:\n")
		fmt.Printf("  Task ID: %d\n", task.ID)
		fmt.Printf("  Created timestamp: %s\n", task.CreatedAt.Format(time.RFC3339))
		fmt.Printf("  Updated timestamp: %s\n", task.UpdatedAt.Format(time.RFC3339))

		if !task.LockedAt.IsZero() {
			fmt.Printf("  Locked timestamp: %s\n", task.LockedAt.Format(time.RFC3339))
		}

		fmt.Printf("  Status transitions: created -> %s\n", task.Status)
		if task.IsComplete() {
			duration := task.UpdatedAt.Sub(task.CreatedAt)
			fmt.Printf("  Completion time: %s\n", formatDuration(duration))
		}
	}
}

func formatTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "just now"
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	days := int(duration.Hours() / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "less than a minute"
	}

	if d < time.Hour {
		minutes := int(d.Minutes())
		if minutes == 1 {
			return "1 minute"
		}
		return fmt.Sprintf("%d minutes", minutes)
	}

	if d < 24*time.Hour {
		hours := int(d.Hours())
		minutes := int(d.Minutes()) % 60
		if hours == 1 && minutes == 0 {
			return "1 hour"
		} else if hours == 1 {
			return fmt.Sprintf("1 hour %d minutes", minutes)
		} else if minutes == 0 {
			return fmt.Sprintf("%d hours", hours)
		}
		return fmt.Sprintf("%d hours %d minutes", hours, minutes)
	}

	days := int(d.Hours() / 24)
	hours := int(d.Hours()) % 24
	if days == 1 && hours == 0 {
		return "1 day"
	} else if days == 1 {
		return fmt.Sprintf("1 day %d hours", hours)
	} else if hours == 0 {
		return fmt.Sprintf("%d days", days)
	}
	return fmt.Sprintf("%d days %d hours", days, hours)
}

func init() {
	RootCmd.AddCommand(displayTaskCmd)
}
