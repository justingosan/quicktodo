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
	statusFilter   string
	priorityFilter string
	assignedFilter string
)

// listTasksCmd represents the list-tasks command
var listTasksCmd = &cobra.Command{
	Use:     "list-tasks",
	Aliases: []string{"show-tasks"},
	Short:   "Show all tasks with optional filters",
	Long: `List all tasks in the current project with optional filtering by status, priority, or assignee.

The command will auto-detect the current project from the working directory.
Use filters to narrow down the results to specific task types.

Examples:
  quicktodo list-tasks
  quicktodo show-tasks --status pending
  quicktodo list-tasks --priority high --json
  quicktodo list-tasks --assigned-to ai-agent-1
  quicktodo list-tasks --status in_progress --priority high`,
	Run: runListTasks,
}

func runListTasks(cmd *cobra.Command, args []string) {
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

	// Create filter
	filter := createTaskFilter()

	// Get filtered tasks
	tasks := projectDB.ListTasks(filter)

	// Save updated registry (for last accessed time)
	if err := registry.Save(registryPath); err != nil && verbose {
		fmt.Fprintf(os.Stderr, "Warning: failed to save registry: %v\n", err)
	}

	// Output results
	if jsonOutput {
		outputTasksJSON(tasks, projectInfo)
	} else {
		outputTasksHuman(tasks, projectInfo)
	}
}

func createTaskFilter() *models.TaskFilter {
	filter := &models.TaskFilter{}

	if statusFilter != "" {
		status := models.Status(strings.ToLower(statusFilter))
		if !models.IsValidStatus(string(status)) {
			fmt.Fprintf(os.Stderr, "Error: invalid status '%s'. Valid statuses: pending, in_progress, done\n", statusFilter)
			os.Exit(1)
		}
		filter.Status = &status
	}

	if priorityFilter != "" {
		priority := models.Priority(strings.ToLower(priorityFilter))
		if !models.IsValidPriority(string(priority)) {
			fmt.Fprintf(os.Stderr, "Error: invalid priority '%s'. Valid priorities: low, medium, high\n", priorityFilter)
			os.Exit(1)
		}
		filter.Priority = &priority
	}

	if assignedFilter != "" {
		filter.AssignedTo = &assignedFilter
	}

	return filter
}

func outputTasksJSON(tasks []*models.Task, projectInfo *database.ProjectInfo) {
	output := map[string]interface{}{
		"success": true,
		"project": map[string]interface{}{
			"name": projectInfo.Name,
			"path": projectInfo.Path,
		},
		"task_count": len(tasks),
		"tasks":      tasks,
	}

	data, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error formatting JSON output: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(data))
}

func outputTasksHuman(tasks []*models.Task, projectInfo *database.ProjectInfo) {
	// Project header
	fmt.Printf("Project: %s (%s)\n", projectInfo.Name, projectInfo.Path)

	if len(tasks) == 0 {
		fmt.Println("No tasks found")
		if statusFilter != "" || priorityFilter != "" || assignedFilter != "" {
			fmt.Println("Try removing filters to see all tasks")
		}
		return
	}

	fmt.Printf("Found %d task(s):\n\n", len(tasks))

	// Sort tasks by ID
	sorter := &models.TaskSorter{Field: "id", Desc: false}
	sorter.Sort(tasks)

	// Display tasks
	for _, task := range tasks {
		displayTask(task)
		fmt.Println()
	}

	// Show summary if verbose
	if verbose {
		showTaskSummary(tasks)
	}
}

func displayTask(task *models.Task) {
	// Status indicator
	statusIcon := getStatusIcon(task.Status)
	priorityColor := getPriorityIndicator(task.Priority)

	fmt.Printf("%s #%-3d %s%s\n", statusIcon, task.ID, priorityColor, task.Title)

	if task.Description != "" {
		fmt.Printf("     %s\n", task.Description)
	}

	// Show metadata in verbose mode or if assigned
	if verbose || task.AssignedTo != "" {
		var metadata []string

		metadata = append(metadata, fmt.Sprintf("Priority: %s", task.Priority))
		metadata = append(metadata, fmt.Sprintf("Created: %s", task.GetAge()))

		if task.AssignedTo != "" {
			metadata = append(metadata, fmt.Sprintf("Assigned: %s", task.AssignedTo))
		}

		if task.IsLocked() {
			metadata = append(metadata, fmt.Sprintf("Locked by: %s", task.LockedBy))
		}

		fmt.Printf("     %s\n", strings.Join(metadata, " | "))
	}
}

func getStatusIcon(status models.Status) string {
	switch status {
	case models.StatusPending:
		return "⏳"
	case models.StatusInProgress:
		return "🏃"
	case models.StatusDone:
		return "✅"
	default:
		return "❓"
	}
}

func getPriorityIndicator(priority models.Priority) string {
	switch priority {
	case models.PriorityHigh:
		return "🔴 "
	case models.PriorityMedium:
		return "🟡 "
	case models.PriorityLow:
		return "🟢 "
	default:
		return ""
	}
}

func showTaskSummary(tasks []*models.Task) {
	statusCounts := make(map[models.Status]int)
	priorityCounts := make(map[models.Priority]int)

	for _, task := range tasks {
		statusCounts[task.Status]++
		priorityCounts[task.Priority]++
	}

	fmt.Println("Summary:")
	fmt.Printf("  Status: %d pending, %d in progress, %d done\n",
		statusCounts[models.StatusPending],
		statusCounts[models.StatusInProgress],
		statusCounts[models.StatusDone])

	fmt.Printf("  Priority: %d high, %d medium, %d low\n",
		priorityCounts[models.PriorityHigh],
		priorityCounts[models.PriorityMedium],
		priorityCounts[models.PriorityLow])
}

func init() {
	listTasksCmd.Flags().StringVarP(&statusFilter, "status", "s", "", "Filter by status (pending, in_progress, done)")
	listTasksCmd.Flags().StringVarP(&priorityFilter, "priority", "p", "", "Filter by priority (low, medium, high)")
	listTasksCmd.Flags().StringVarP(&assignedFilter, "assigned-to", "a", "", "Filter by assignee")

	RootCmd.AddCommand(listTasksCmd)
}
