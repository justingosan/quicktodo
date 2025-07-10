package models

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"time"
)

// Project represents a project in the system
type Project struct {
	Name         string    `json:"name"`
	Path         string    `json:"path"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
	TaskCount    int       `json:"task_count"`
	Description  string    `json:"description"`
}

// ProjectDatabase represents the complete project database structure
type ProjectDatabase struct {
	Project      *Project `json:"project"`
	Tasks        []*Task  `json:"tasks"`
	NextID       int      `json:"next_id"`
	LastModified time.Time `json:"last_modified"`
	Version      int      `json:"version"`
}

// ProjectSummary provides a summary of project statistics
type ProjectSummary struct {
	Project       *Project          `json:"project"`
	TaskCount     int               `json:"task_count"`
	StatusCounts  map[Status]int    `json:"status_counts"`
	PriorityCounts map[Priority]int  `json:"priority_counts"`
	CompletedTasks int              `json:"completed_tasks"`
	PendingTasks  int               `json:"pending_tasks"`
	InProgressTasks int             `json:"in_progress_tasks"`
	LastTaskUpdate time.Time         `json:"last_task_update"`
}

// NewProject creates a new project with default values
func NewProject(name, path string) *Project {
	absPath, err := filepath.Abs(path)
	if err != nil {
		absPath = path
	}
	
	return &Project{
		Name:         name,
		Path:         absPath,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		TaskCount:    0,
		Description:  "",
	}
}

// NewProjectWithDescription creates a new project with a description
func NewProjectWithDescription(name, path, description string) *Project {
	project := NewProject(name, path)
	project.Description = description
	return project
}

// Validate validates the project fields
func (p *Project) Validate() error {
	if p.Name == "" {
		return fmt.Errorf("project name cannot be empty")
	}
	
	if p.Path == "" {
		return fmt.Errorf("project path cannot be empty")
	}
	
	if p.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}
	
	if p.LastAccessed.IsZero() {
		return fmt.Errorf("last_accessed cannot be zero")
	}
	
	if p.LastAccessed.Before(p.CreatedAt) {
		return fmt.Errorf("last_accessed cannot be before created_at")
	}
	
	if p.TaskCount < 0 {
		return fmt.Errorf("task_count cannot be negative")
	}
	
	return nil
}

// UpdateLastAccessed updates the last accessed timestamp
func (p *Project) UpdateLastAccessed() {
	p.LastAccessed = time.Now()
}

// UpdateTaskCount updates the task count
func (p *Project) UpdateTaskCount(count int) {
	if count >= 0 {
		p.TaskCount = count
	}
}

// UpdateDescription updates the project description
func (p *Project) UpdateDescription(description string) {
	p.Description = description
}

// GetAge returns the age of the project in a human-readable format
func (p *Project) GetAge() string {
	duration := time.Since(p.CreatedAt)
	
	if duration < time.Hour {
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	}
	
	if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	}
	
	return fmt.Sprintf("%d days", int(duration.Hours()/24))
}

// GetLastAccessedAge returns how long ago the project was last accessed
func (p *Project) GetLastAccessedAge() string {
	duration := time.Since(p.LastAccessed)
	
	if duration < time.Hour {
		return fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	}
	
	if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours ago", int(duration.Hours()))
	}
	
	return fmt.Sprintf("%d days ago", int(duration.Hours()/24))
}

// Clone creates a copy of the project
func (p *Project) Clone() *Project {
	return &Project{
		Name:         p.Name,
		Path:         p.Path,
		CreatedAt:    p.CreatedAt,
		LastAccessed: p.LastAccessed,
		TaskCount:    p.TaskCount,
		Description:  p.Description,
	}
}

// ToJSON converts the project to JSON
func (p *Project) ToJSON() ([]byte, error) {
	return json.MarshalIndent(p, "", "  ")
}

// String returns a string representation of the project
func (p *Project) String() string {
	return fmt.Sprintf("Project[%s]: %s (%d tasks)", p.Name, p.Path, p.TaskCount)
}

// NewProjectDatabase creates a new project database
func NewProjectDatabase(project *Project) *ProjectDatabase {
	return &ProjectDatabase{
		Project:      project,
		Tasks:        make([]*Task, 0),
		NextID:       1,
		LastModified: time.Now(),
		Version:      1,
	}
}

// Validate validates the project database structure
func (db *ProjectDatabase) Validate() error {
	if db.Project == nil {
		return fmt.Errorf("project cannot be nil")
	}
	
	if err := db.Project.Validate(); err != nil {
		return fmt.Errorf("invalid project: %w", err)
	}
	
	if db.NextID < 1 {
		return fmt.Errorf("next_id must be at least 1")
	}
	
	if db.Version < 1 {
		return fmt.Errorf("version must be at least 1")
	}
	
	if db.LastModified.IsZero() {
		return fmt.Errorf("last_modified cannot be zero")
	}
	
	// Validate all tasks
	for i, task := range db.Tasks {
		if task == nil {
			return fmt.Errorf("task at index %d is nil", i)
		}
		
		if err := task.Validate(); err != nil {
			return fmt.Errorf("invalid task at index %d: %w", i, err)
		}
	}
	
	return nil
}

// AddTask adds a new task to the database
func (db *ProjectDatabase) AddTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}
	
	// Set the task ID
	task.ID = db.NextID
	db.NextID++
	
	// Add to tasks
	db.Tasks = append(db.Tasks, task)
	
	// Update metadata
	db.LastModified = time.Now()
	db.Version++
	db.Project.UpdateTaskCount(len(db.Tasks))
	
	return nil
}

// GetTask retrieves a task by ID
func (db *ProjectDatabase) GetTask(id int) (*Task, error) {
	for _, task := range db.Tasks {
		if task.ID == id {
			return task, nil
		}
	}
	
	return nil, fmt.Errorf("task with ID %d not found", id)
}

// UpdateTask updates a task in the database
func (db *ProjectDatabase) UpdateTask(task *Task) error {
	if task == nil {
		return fmt.Errorf("task cannot be nil")
	}
	
	if err := task.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}
	
	// Find and update the task
	for i, existingTask := range db.Tasks {
		if existingTask.ID == task.ID {
			db.Tasks[i] = task
			db.LastModified = time.Now()
			db.Version++
			return nil
		}
	}
	
	return fmt.Errorf("task with ID %d not found", task.ID)
}

// DeleteTask removes a task from the database
func (db *ProjectDatabase) DeleteTask(id int) error {
	for i, task := range db.Tasks {
		if task.ID == id {
			// Remove task from slice
			db.Tasks = append(db.Tasks[:i], db.Tasks[i+1:]...)
			
			// Update metadata
			db.LastModified = time.Now()
			db.Version++
			db.Project.UpdateTaskCount(len(db.Tasks))
			
			return nil
		}
	}
	
	return fmt.Errorf("task with ID %d not found", id)
}

// ListTasks returns all tasks, optionally filtered
func (db *ProjectDatabase) ListTasks(filter *TaskFilter) []*Task {
	if filter == nil {
		// Return all tasks
		tasks := make([]*Task, len(db.Tasks))
		for i, task := range db.Tasks {
			tasks[i] = task.Clone()
		}
		return tasks
	}
	
	// Filter tasks
	var filteredTasks []*Task
	for _, task := range db.Tasks {
		if filter.Matches(task) {
			filteredTasks = append(filteredTasks, task.Clone())
		}
	}
	
	return filteredTasks
}

// SearchTasks searches for tasks by title or description
func (db *ProjectDatabase) SearchTasks(query string) []*Task {
	var matchedTasks []*Task
	
	for _, task := range db.Tasks {
		if containsIgnoreCase(task.Title, query) || containsIgnoreCase(task.Description, query) {
			matchedTasks = append(matchedTasks, task.Clone())
		}
	}
	
	return matchedTasks
}

// GetSummary returns a summary of the project
func (db *ProjectDatabase) GetSummary() *ProjectSummary {
	summary := &ProjectSummary{
		Project:         db.Project.Clone(),
		TaskCount:       len(db.Tasks),
		StatusCounts:    make(map[Status]int),
		PriorityCounts:  make(map[Priority]int),
		CompletedTasks:  0,
		PendingTasks:    0,
		InProgressTasks: 0,
		LastTaskUpdate:  time.Time{},
	}
	
	// Calculate statistics
	for _, task := range db.Tasks {
		// Count by status
		summary.StatusCounts[task.Status]++
		
		// Count by priority
		summary.PriorityCounts[task.Priority]++
		
		// Count by specific statuses
		switch task.Status {
		case StatusDone:
			summary.CompletedTasks++
		case StatusPending:
			summary.PendingTasks++
		case StatusInProgress:
			summary.InProgressTasks++
		}
		
		// Track latest update
		if task.UpdatedAt.After(summary.LastTaskUpdate) {
			summary.LastTaskUpdate = task.UpdatedAt
		}
	}
	
	return summary
}

// ToJSON converts the project database to JSON
func (db *ProjectDatabase) ToJSON() ([]byte, error) {
	return json.MarshalIndent(db, "", "  ")
}

// Helper function for case-insensitive substring matching
func containsIgnoreCase(s, substr string) bool {
	if len(substr) == 0 {
		return true
	}
	
	if len(s) < len(substr) {
		return false
	}
	
	// Simple case-insensitive search
	for i := 0; i <= len(s)-len(substr); i++ {
		if caseInsensitiveEqual(s[i:i+len(substr)], substr) {
			return true
		}
	}
	
	return false
}

// Helper function for case-insensitive string comparison
func caseInsensitiveEqual(s1, s2 string) bool {
	if len(s1) != len(s2) {
		return false
	}
	
	for i := 0; i < len(s1); i++ {
		c1, c2 := s1[i], s2[i]
		
		// Convert to lowercase
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