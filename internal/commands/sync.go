package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"quicktodo/internal/sync"

	"github.com/spf13/cobra"
)

var (
	enableSync  bool
	disableSync bool
	fullSync    bool
	showStatus  bool
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Manage TODO synchronization between QuickTodo and AI TODO lists",
	Long: `Manage synchronization between QuickTodo database and AI TODO tracking systems.
	
This command allows you to:
- Enable/disable automatic synchronization
- Perform full synchronization of current tasks
- Check synchronization status
- View synchronized TODO items

When enabled, changes to QuickTodo tasks will automatically update the AI's TODO list.`,
	RunE: runSync,
}

func init() {
	syncCmd.Flags().BoolVar(&enableSync, "enable", false, "Enable TODO synchronization")
	syncCmd.Flags().BoolVar(&disableSync, "disable", false, "Disable TODO synchronization")
	syncCmd.Flags().BoolVar(&fullSync, "full-sync", false, "Perform full synchronization of current project")
	syncCmd.Flags().BoolVar(&showStatus, "status", false, "Show synchronization status")
	
	// Make flags mutually exclusive
	syncCmd.MarkFlagsMutuallyExclusive("enable", "disable", "full-sync", "status")
	
	RootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Initialize sync manager
	syncConfigPath := filepath.Join(cfg.DataDir, "sync_config.json")
	syncManager, err := sync.NewTodoSyncManager(syncConfigPath)
	if err != nil {
		return fmt.Errorf("failed to initialize sync manager: %w", err)
	}

	switch {
	case enableSync:
		return handleEnableSync(syncManager)
	case disableSync:
		return handleDisableSync(syncManager)
	case showStatus:
		return handleShowStatus(syncManager, cfg)
	case fullSync:
		return handleFullSync(syncManager, cfg)
	default:
		// Default behavior: show status
		return handleShowStatus(syncManager, cfg)
	}
}

func handleEnableSync(syncManager *sync.TodoSyncManager) error {
	if err := syncManager.Enable(); err != nil {
		return fmt.Errorf("failed to enable sync: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"success": true,
			"message": "TODO synchronization enabled",
			"status":  "enabled",
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("✅ TODO synchronization enabled")
		fmt.Println("QuickTodo database changes will now automatically sync to AI TODO lists")
		fmt.Println("Use 'quicktodo sync --full-sync' to synchronize existing tasks")
	}

	return nil
}

func handleDisableSync(syncManager *sync.TodoSyncManager) error {
	if err := syncManager.Disable(); err != nil {
		return fmt.Errorf("failed to disable sync: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"success": true,
			"message": "TODO synchronization disabled",
			"status":  "disabled",
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Println("❌ TODO synchronization disabled")
		fmt.Println("QuickTodo database changes will no longer sync to AI TODO lists")
	}

	return nil
}

func handleShowStatus(syncManager *sync.TodoSyncManager, cfg *config.Config) error {
	todoItems := syncManager.GetTodoItems()
	
	if jsonOutput {
		data, err := syncManager.GetTodoItemsAsJSON()
		if err != nil {
			return fmt.Errorf("failed to get TODO items as JSON: %w", err)
		}
		fmt.Println(string(data))
	} else {
		fmt.Println("TODO Synchronization Status")
		fmt.Println("==========================")
		
		if len(todoItems) == 0 {
			fmt.Println("No synchronized TODO items")
			fmt.Println("Run 'quicktodo sync --enable' to enable synchronization")
			fmt.Println("Run 'quicktodo sync --full-sync' to sync existing tasks")
		} else {
			fmt.Printf("Total TODO items: %d\n\n", len(todoItems))
			
			// Group by project
			projectGroups := make(map[string][]*sync.TodoItem)
			for _, item := range todoItems {
				projectGroups[item.ProjectName] = append(projectGroups[item.ProjectName], item)
			}
			
			for projectName, items := range projectGroups {
				fmt.Printf("Project: %s (%d items)\n", projectName, len(items))
				for _, item := range items {
					statusIcon := getTodoStatusIcon(item.Status)
					priorityIcon := getPriorityIcon(item.Priority)
					fmt.Printf("  %s %s %s\n", statusIcon, priorityIcon, item.Content)
				}
				fmt.Println()
			}
		}
		
		fmt.Println("Commands:")
		fmt.Println("  quicktodo sync --enable     Enable automatic synchronization")
		fmt.Println("  quicktodo sync --disable    Disable automatic synchronization")
		fmt.Println("  quicktodo sync --full-sync  Sync all tasks from current project")
	}

	return nil
}

func handleFullSync(syncManager *sync.TodoSyncManager, cfg *config.Config) error {
	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	// Load project registry
	registryPath := cfg.GetProjectsPath()
	registry, err := database.LoadProjectRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load project registry: %w", err)
	}

	// Find project for current directory
	projectInfo, exists := registry.GetProjectByPath(currentDir)
	if !exists {
		return fmt.Errorf("current directory is not a registered project")
	}

	// Load project database
	dbPath := cfg.GetProjectDatabasePath(projectInfo.Name)
	projectDB, err := loadProjectDatabase(dbPath)
	if err != nil {
		return fmt.Errorf("failed to load project database: %w", err)
	}

	// Get all tasks
	tasks := projectDB.ListTasks(nil)

	// Perform full sync
	if err := syncManager.SyncFromQuickTodo(tasks, projectInfo.Name); err != nil {
		return fmt.Errorf("failed to perform full sync: %w", err)
	}

	if jsonOutput {
		output := map[string]interface{}{
			"success":     true,
			"message":     "Full synchronization completed",
			"project":     projectInfo.Name,
			"task_count":  len(tasks),
			"synced_items": len(syncManager.GetTodoItems()),
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Printf("✅ Full synchronization completed for project '%s'\n", projectInfo.Name)
		fmt.Printf("Synchronized %d tasks to TODO list\n", len(tasks))
		
		if len(tasks) > 0 {
			fmt.Println("\nSynchronized tasks:")
			for _, task := range tasks {
				statusIcon := getTaskStatusIcon(task.Status)
				priorityIcon := getTaskPriorityIcon(task.Priority)
				fmt.Printf("  %s %s #%d %s\n", statusIcon, priorityIcon, task.ID, task.Title)
			}
		}
	}

	return nil
}

func getTodoStatusIcon(status string) string {
	switch status {
	case "pending":
		return "⏳"
	case "in_progress":
		return "🏃"
	case "completed":
		return "✅"
	default:
		return "❓"
	}
}

func getPriorityIcon(priority string) string {
	switch priority {
	case "high":
		return "🔴"
	case "medium":
		return "🟡"
	case "low":
		return "🟢"
	default:
		return "⚪"
	}
}

func getTaskStatusIcon(status models.Status) string {
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

func getTaskPriorityIcon(priority models.Priority) string {
	switch priority {
	case models.PriorityHigh:
		return "🔴"
	case models.PriorityMedium:
		return "🟡"
	case models.PriorityLow:
		return "🟢"
	default:
		return "⚪"
	}
}