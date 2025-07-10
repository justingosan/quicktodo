package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectDatabase represents a project's task database
type ProjectDatabase struct {
	ProjectName  string      `json:"project_name"`
	ProjectPath  string      `json:"project_path"`
	Tasks        []TaskEntry `json:"tasks"`
	NextID       int         `json:"next_id"`
	LastModified time.Time   `json:"last_modified"`
	Version      int         `json:"version"`
}

// TaskEntry represents a task in the database
type TaskEntry struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AssignedTo  string    `json:"assigned_to"`
	LockedBy    string    `json:"locked_by"`
	LockedAt    time.Time `json:"locked_at"`
}

// NewProjectDatabase creates a new project database
func NewProjectDatabase(projectName, projectPath string) *ProjectDatabase {
	return &ProjectDatabase{
		ProjectName:  projectName,
		ProjectPath:  projectPath,
		Tasks:        make([]TaskEntry, 0),
		NextID:       1,
		LastModified: time.Now(),
		Version:      1,
	}
}

// LoadProjectDatabase loads a project database from file
func LoadProjectDatabase(filePath string) (*ProjectDatabase, error) {
	// If file doesn't exist, return error (should be created explicitly)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("project database file does not exist: %s", filePath)
	}

	// Read existing database
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read project database: %w", err)
	}

	var db ProjectDatabase
	if err := json.Unmarshal(data, &db); err != nil {
		return nil, fmt.Errorf("failed to parse project database: %w", err)
	}

	// Validate database structure
	if err := db.Validate(); err != nil {
		return nil, fmt.Errorf("invalid project database: %w", err)
	}

	return &db, nil
}

// Save saves the project database to file
func (db *ProjectDatabase) Save(filePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Update last modified time and increment version
	db.LastModified = time.Now()
	db.Version++

	// Marshal database to JSON
	data, err := json.MarshalIndent(db, "", "  ")
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

// AddTask adds a new task to the database
func (db *ProjectDatabase) AddTask(title, description, priority string) (*TaskEntry, error) {
	// Validate inputs
	if title == "" {
		return nil, fmt.Errorf("task title cannot be empty")
	}

	if priority == "" {
		priority = "medium"
	}

	// Create new task
	task := TaskEntry{
		ID:          db.NextID,
		Title:       title,
		Description: description,
		Status:      "pending",
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AssignedTo:  "",
		LockedBy:    "",
		LockedAt:    time.Time{},
	}

	// Add to tasks list
	db.Tasks = append(db.Tasks, task)
	db.NextID++

	return &task, nil
}

// GetTask retrieves a task by ID
func (db *ProjectDatabase) GetTask(id int) (*TaskEntry, error) {
	for i, task := range db.Tasks {
		if task.ID == id {
			return &db.Tasks[i], nil
		}
	}
	return nil, fmt.Errorf("task with ID %d not found", id)
}

// UpdateTask updates a task in the database
func (db *ProjectDatabase) UpdateTask(id int, updates map[string]interface{}) error {
	taskIndex := -1
	for i, task := range db.Tasks {
		if task.ID == id {
			taskIndex = i
			break
		}
	}

	if taskIndex == -1 {
		return fmt.Errorf("task with ID %d not found", id)
	}

	// Update fields
	task := &db.Tasks[taskIndex]

	if title, ok := updates["title"]; ok {
		if titleStr, ok := title.(string); ok {
			task.Title = titleStr
		}
	}

	if description, ok := updates["description"]; ok {
		if descStr, ok := description.(string); ok {
			task.Description = descStr
		}
	}

	if status, ok := updates["status"]; ok {
		if statusStr, ok := status.(string); ok {
			task.Status = statusStr
		}
	}

	if priority, ok := updates["priority"]; ok {
		if priorityStr, ok := priority.(string); ok {
			task.Priority = priorityStr
		}
	}

	if assignedTo, ok := updates["assigned_to"]; ok {
		if assignedStr, ok := assignedTo.(string); ok {
			task.AssignedTo = assignedStr
		}
	}

	// Always update the updated_at timestamp
	task.UpdatedAt = time.Now()

	return nil
}

// DeleteTask removes a task from the database
func (db *ProjectDatabase) DeleteTask(id int) error {
	for i, task := range db.Tasks {
		if task.ID == id {
			// Remove task from slice
			db.Tasks = append(db.Tasks[:i], db.Tasks[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("task with ID %d not found", id)
}

// ListTasks returns all tasks, optionally filtered by status
func (db *ProjectDatabase) ListTasks(statusFilter string) []TaskEntry {
	if statusFilter == "" {
		// Return all tasks
		tasks := make([]TaskEntry, len(db.Tasks))
		copy(tasks, db.Tasks)
		return tasks
	}

	// Filter by status
	var filteredTasks []TaskEntry
	for _, task := range db.Tasks {
		if task.Status == statusFilter {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks
}

// SearchTasks searches for tasks by title or description
func (db *ProjectDatabase) SearchTasks(query string) []TaskEntry {
	var matchedTasks []TaskEntry

	for _, task := range db.Tasks {
		if contains(task.Title, query) || contains(task.Description, query) {
			matchedTasks = append(matchedTasks, task)
		}
	}

	return matchedTasks
}

// Validate validates the database structure
func (db *ProjectDatabase) Validate() error {
	if db.ProjectName == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if db.ProjectPath == "" {
		return fmt.Errorf("project path cannot be empty")
	}

	if db.NextID < 1 {
		return fmt.Errorf("next_id must be at least 1")
	}

	if db.Version < 1 {
		return fmt.Errorf("version must be at least 1")
	}

	// Validate tasks
	for _, task := range db.Tasks {
		if err := db.validateTask(&task); err != nil {
			return fmt.Errorf("invalid task %d: %w", task.ID, err)
		}
	}

	return nil
}

// validateTask validates a single task
func (db *ProjectDatabase) validateTask(task *TaskEntry) error {
	if task.ID <= 0 {
		return fmt.Errorf("task ID must be positive")
	}

	if task.Title == "" {
		return fmt.Errorf("task title cannot be empty")
	}

	validStatuses := map[string]bool{
		"pending":     true,
		"in_progress": true,
		"done":        true,
	}

	if !validStatuses[task.Status] {
		return fmt.Errorf("invalid status: %s", task.Status)
	}

	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}

	if !validPriorities[task.Priority] {
		return fmt.Errorf("invalid priority: %s", task.Priority)
	}

	if task.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}

	if task.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at cannot be zero")
	}

	return nil
}

// Helper function to check if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				anyMatch(s, substr)))
}

// Helper function for case-insensitive substring matching
func anyMatch(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if caseInsensitiveMatch(s[i:i+len(substr)], substr) {
			return true
		}
	}
	return false
}

// Helper function for case-insensitive comparison
func caseInsensitiveMatch(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}

	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		if c1 >= 'A' && c1 <= 'Z' {
			c1 = c1 + ('a' - 'A')
		}
		if c2 >= 'A' && c2 <= 'Z' {
			c2 = c2 + ('a' - 'A')
		}
		if c1 != c2 {
			return false
		}
	}

	return true
}
