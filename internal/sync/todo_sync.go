package sync

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"quicktodo/internal/models"
	"time"
)

// TodoSyncConfig represents configuration for TODO synchronization
type TodoSyncConfig struct {
	Enabled        bool   `json:"enabled"`
	TodoFilePath   string `json:"todo_file_path"`
	AutoSync       bool   `json:"auto_sync"`
	SyncOnStatus   bool   `json:"sync_on_status"`
	SyncOnEdit     bool   `json:"sync_on_edit"`
	SyncOnCreate   bool   `json:"sync_on_create"`
	SyncOnDelete   bool   `json:"sync_on_delete"`
	AgentID        string `json:"agent_id"`
	LastSyncTime   time.Time `json:"last_sync_time"`
}

// TodoItem represents a simplified TODO item for AI tracking
type TodoItem struct {
	ID          string    `json:"id"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ProjectName string    `json:"project_name"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TodoSyncManager manages synchronization between QuickTodo and AI TODO lists
type TodoSyncManager struct {
	config    *TodoSyncConfig
	todoItems map[string]*TodoItem
	enabled   bool
}

// NewTodoSyncManager creates a new TODO synchronization manager
func NewTodoSyncManager(configPath string) (*TodoSyncManager, error) {
	config, err := loadSyncConfig(configPath)
	if err != nil {
		// Create default config if not found
		config = defaultSyncConfig()
		if err := saveSyncConfig(config, configPath); err != nil {
			return nil, fmt.Errorf("failed to save default sync config: %w", err)
		}
	}

	manager := &TodoSyncManager{
		config:    config,
		todoItems: make(map[string]*TodoItem),
		enabled:   config.Enabled,
	}

	if config.Enabled {
		if err := manager.loadTodoItems(); err != nil {
			return nil, fmt.Errorf("failed to load TODO items: %w", err)
		}
	}

	return manager, nil
}

// OnTaskCreated handles task creation synchronization
func (m *TodoSyncManager) OnTaskCreated(task *models.Task, projectName string) error {
	if !m.enabled || !m.config.SyncOnCreate {
		return nil
	}

	todoItem := &TodoItem{
		ID:          fmt.Sprintf("%s-%d", projectName, task.ID),
		Content:     fmt.Sprintf("#%d %s", task.ID, task.Title),
		Status:      mapTaskStatusToTodoStatus(task.Status),
		Priority:    mapTaskPriorityToTodoPriority(task.Priority),
		ProjectName: projectName,
		UpdatedAt:   task.UpdatedAt,
	}

	if task.Description != "" {
		todoItem.Content += " - " + task.Description
	}

	m.todoItems[todoItem.ID] = todoItem
	return m.saveTodoItems()
}

// OnTaskUpdated handles task update synchronization
func (m *TodoSyncManager) OnTaskUpdated(task *models.Task, projectName string, changeType string) error {
	if !m.enabled {
		return nil
	}

	// Check if we should sync based on change type
	shouldSync := false
	switch changeType {
	case "status":
		shouldSync = m.config.SyncOnStatus
	case "edit":
		shouldSync = m.config.SyncOnEdit
	default:
		shouldSync = m.config.AutoSync
	}

	if !shouldSync {
		return nil
	}

	todoID := fmt.Sprintf("%s-%d", projectName, task.ID)
	
	// Update existing item or create new one
	todoItem := m.todoItems[todoID]
	if todoItem == nil {
		return m.OnTaskCreated(task, projectName)
	}

	// Update the TODO item
	todoItem.Content = fmt.Sprintf("#%d %s", task.ID, task.Title)
	if task.Description != "" {
		todoItem.Content += " - " + task.Description
	}
	todoItem.Status = mapTaskStatusToTodoStatus(task.Status)
	todoItem.Priority = mapTaskPriorityToTodoPriority(task.Priority)
	todoItem.UpdatedAt = task.UpdatedAt

	return m.saveTodoItems()
}

// OnTaskDeleted handles task deletion synchronization
func (m *TodoSyncManager) OnTaskDeleted(taskID int, projectName string) error {
	if !m.enabled || !m.config.SyncOnDelete {
		return nil
	}

	todoID := fmt.Sprintf("%s-%d", projectName, taskID)
	delete(m.todoItems, todoID)
	return m.saveTodoItems()
}

// SyncFromQuickTodo performs a full sync from QuickTodo database to TODO list
func (m *TodoSyncManager) SyncFromQuickTodo(tasks []*models.Task, projectName string) error {
	if !m.enabled {
		return nil
	}

	// Clear existing items for this project
	for id := range m.todoItems {
		if todoItem := m.todoItems[id]; todoItem.ProjectName == projectName {
			delete(m.todoItems, id)
		}
	}

	// Add all current tasks
	for _, task := range tasks {
		if err := m.OnTaskCreated(task, projectName); err != nil {
			return fmt.Errorf("failed to sync task %d: %w", task.ID, err)
		}
	}

	m.config.LastSyncTime = time.Now()
	return m.saveSyncConfig()
}

// GetTodoItems returns all TODO items
func (m *TodoSyncManager) GetTodoItems() map[string]*TodoItem {
	return m.todoItems
}

// GetTodoItemsAsJSON returns TODO items formatted for Claude's TodoWrite tool
func (m *TodoSyncManager) GetTodoItemsAsJSON() ([]byte, error) {
	var todos []map[string]interface{}
	
	for _, item := range m.todoItems {
		todo := map[string]interface{}{
			"id":       item.ID,
			"content":  item.Content,
			"status":   item.Status,
			"priority": item.Priority,
		}
		todos = append(todos, todo)
	}

	return json.MarshalIndent(map[string]interface{}{
		"todos": todos,
		"last_sync": m.config.LastSyncTime,
		"project_count": len(m.getProjectCounts()),
	}, "", "  ")
}

// Enable enables TODO synchronization
func (m *TodoSyncManager) Enable() error {
	m.enabled = true
	m.config.Enabled = true
	return m.saveSyncConfig()
}

// Disable disables TODO synchronization
func (m *TodoSyncManager) Disable() error {
	m.enabled = false
	m.config.Enabled = false
	return m.saveSyncConfig()
}

// Helper functions

func mapTaskStatusToTodoStatus(status models.Status) string {
	switch status {
	case models.StatusPending:
		return "pending"
	case models.StatusInProgress:
		return "in_progress"
	case models.StatusDone:
		return "completed"
	default:
		return "pending"
	}
}

func mapTaskPriorityToTodoPriority(priority models.Priority) string {
	switch priority {
	case models.PriorityHigh:
		return "high"
	case models.PriorityMedium:
		return "medium"
	case models.PriorityLow:
		return "low"
	default:
		return "medium"
	}
}

func (m *TodoSyncManager) getProjectCounts() map[string]int {
	counts := make(map[string]int)
	for _, item := range m.todoItems {
		counts[item.ProjectName]++
	}
	return counts
}

func (m *TodoSyncManager) loadTodoItems() error {
	if m.config.TodoFilePath == "" {
		return nil
	}

	data, err := os.ReadFile(m.config.TodoFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // File doesn't exist yet, that's OK
		}
		return fmt.Errorf("failed to read TODO file: %w", err)
	}

	var items map[string]*TodoItem
	if err := json.Unmarshal(data, &items); err != nil {
		return fmt.Errorf("failed to parse TODO file: %w", err)
	}

	m.todoItems = items
	return nil
}

func (m *TodoSyncManager) saveTodoItems() error {
	if m.config.TodoFilePath == "" {
		return nil
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.config.TodoFilePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(m.todoItems, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal TODO items: %w", err)
	}

	return os.WriteFile(m.config.TodoFilePath, data, 0644)
}

func (m *TodoSyncManager) saveSyncConfig() error {
	configDir := filepath.Dir(m.config.TodoFilePath)
	configPath := filepath.Join(configDir, "sync_config.json")
	return saveSyncConfig(m.config, configPath)
}

func loadSyncConfig(configPath string) (*TodoSyncConfig, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config TodoSyncConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse sync config: %w", err)
	}

	return &config, nil
}

func saveSyncConfig(config *TodoSyncConfig, configPath string) error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal sync config: %w", err)
	}

	return os.WriteFile(configPath, data, 0644)
}

func defaultSyncConfig() *TodoSyncConfig {
	homeDir, _ := os.UserHomeDir()
	return &TodoSyncConfig{
		Enabled:      false, // Disabled by default, user must opt-in
		TodoFilePath: filepath.Join(homeDir, ".config", "quicktodo", "ai_todos.json"),
		AutoSync:     true,
		SyncOnStatus: true,
		SyncOnEdit:   true,
		SyncOnCreate: true,
		SyncOnDelete: true,
		AgentID:      "",
		LastSyncTime: time.Now(),
	}
}