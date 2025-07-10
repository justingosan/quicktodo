package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// Task represents a task in the system
type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      Status    `json:"status"`
	Priority    Priority  `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	AssignedTo  string    `json:"assigned_to"`
	LockedBy    string    `json:"locked_by"`
	LockedAt    time.Time `json:"locked_at"`
}

// Status represents task status
type Status string

// Task statuses
const (
	StatusPending    Status = "pending"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

// Priority represents task priority
type Priority string

// Task priorities
const (
	PriorityLow    Priority = "low"
	PriorityMedium Priority = "medium"
	PriorityHigh   Priority = "high"
)

// ValidStatuses returns a slice of all valid statuses
func ValidStatuses() []Status {
	return []Status{StatusPending, StatusInProgress, StatusDone}
}

// ValidPriorities returns a slice of all valid priorities
func ValidPriorities() []Priority {
	return []Priority{PriorityLow, PriorityMedium, PriorityHigh}
}

// IsValidStatus checks if a status is valid
func IsValidStatus(status string) bool {
	switch Status(status) {
	case StatusPending, StatusInProgress, StatusDone:
		return true
	default:
		return false
	}
}

// IsValidPriority checks if a priority is valid
func IsValidPriority(priority string) bool {
	switch Priority(priority) {
	case PriorityLow, PriorityMedium, PriorityHigh:
		return true
	default:
		return false
	}
}

// NewTask creates a new task with default values
func NewTask(id int, title string) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: "",
		Status:      StatusPending,
		Priority:    PriorityMedium,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AssignedTo:  "",
		LockedBy:    "",
		LockedAt:    time.Time{},
	}
}

// NewTaskWithDetails creates a new task with specified details
func NewTaskWithDetails(id int, title, description string, priority Priority) *Task {
	return &Task{
		ID:          id,
		Title:       title,
		Description: description,
		Status:      StatusPending,
		Priority:    priority,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		AssignedTo:  "",
		LockedBy:    "",
		LockedAt:    time.Time{},
	}
}

// Validate validates the task fields
func (t *Task) Validate() error {
	if t.ID <= 0 {
		return fmt.Errorf("task ID must be positive")
	}

	if t.Title == "" {
		return fmt.Errorf("task title cannot be empty")
	}

	if !IsValidStatus(string(t.Status)) {
		return fmt.Errorf("invalid status: %s", t.Status)
	}

	if !IsValidPriority(string(t.Priority)) {
		return fmt.Errorf("invalid priority: %s", t.Priority)
	}

	if t.CreatedAt.IsZero() {
		return fmt.Errorf("created_at cannot be zero")
	}

	if t.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at cannot be zero")
	}

	if t.UpdatedAt.Before(t.CreatedAt) {
		return fmt.Errorf("updated_at cannot be before created_at")
	}

	return nil
}

// UpdateStatus updates the task status and timestamp
func (t *Task) UpdateStatus(status Status) error {
	if !IsValidStatus(string(status)) {
		return fmt.Errorf("invalid status: %s", status)
	}

	t.Status = status
	t.UpdatedAt = time.Now()

	return nil
}

// UpdatePriority updates the task priority and timestamp
func (t *Task) UpdatePriority(priority Priority) error {
	if !IsValidPriority(string(priority)) {
		return fmt.Errorf("invalid priority: %s", priority)
	}

	t.Priority = priority
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateTitle updates the task title and timestamp
func (t *Task) UpdateTitle(title string) error {
	if title == "" {
		return fmt.Errorf("task title cannot be empty")
	}

	t.Title = title
	t.UpdatedAt = time.Now()

	return nil
}

// UpdateDescription updates the task description and timestamp
func (t *Task) UpdateDescription(description string) {
	t.Description = description
	t.UpdatedAt = time.Now()
}

// AssignTo assigns the task to an agent or user
func (t *Task) AssignTo(assignee string) {
	t.AssignedTo = assignee
	t.UpdatedAt = time.Now()
}

// Lock locks the task for exclusive access
func (t *Task) Lock(processID string) {
	t.LockedBy = processID
	t.LockedAt = time.Now()
}

// Unlock unlocks the task
func (t *Task) Unlock() {
	t.LockedBy = ""
	t.LockedAt = time.Time{}
}

// IsLocked checks if the task is currently locked
func (t *Task) IsLocked() bool {
	return t.LockedBy != ""
}

// IsLockedBy checks if the task is locked by a specific process
func (t *Task) IsLockedBy(processID string) bool {
	return t.LockedBy == processID
}

// IsStale checks if the task lock is stale (older than 5 minutes)
func (t *Task) IsStale() bool {
	if !t.IsLocked() {
		return false
	}

	return time.Since(t.LockedAt) > 5*time.Minute
}

// Clone creates a copy of the task
func (t *Task) Clone() *Task {
	return &Task{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Status:      t.Status,
		Priority:    t.Priority,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
		AssignedTo:  t.AssignedTo,
		LockedBy:    t.LockedBy,
		LockedAt:    t.LockedAt,
	}
}

// ToJSON converts the task to JSON
func (t *Task) ToJSON() ([]byte, error) {
	return json.MarshalIndent(t, "", "  ")
}

// FromJSON creates a task from JSON
func FromJSON(data []byte) (*Task, error) {
	var task Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal task: %w", err)
	}

	if err := task.Validate(); err != nil {
		return nil, fmt.Errorf("invalid task data: %w", err)
	}

	return &task, nil
}

// String returns a string representation of the task
func (t *Task) String() string {
	return fmt.Sprintf("Task[%d]: %s (%s, %s)", t.ID, t.Title, t.Status, t.Priority)
}

// IsComplete checks if the task is marked as done
func (t *Task) IsComplete() bool {
	return t.Status == StatusDone
}

// IsPending checks if the task is pending
func (t *Task) IsPending() bool {
	return t.Status == StatusPending
}

// IsInProgress checks if the task is in progress
func (t *Task) IsInProgress() bool {
	return t.Status == StatusInProgress
}

// GetDuration returns the time elapsed since task creation
func (t *Task) GetDuration() time.Duration {
	return time.Since(t.CreatedAt)
}

// GetAge returns the age of the task in a human-readable format
func (t *Task) GetAge() string {
	duration := t.GetDuration()

	if duration < time.Hour {
		return fmt.Sprintf("%d minutes", int(duration.Minutes()))
	}

	if duration < 24*time.Hour {
		return fmt.Sprintf("%d hours", int(duration.Hours()))
	}

	return fmt.Sprintf("%d days", int(duration.Hours()/24))
}

// TaskFilter represents filter criteria for tasks
type TaskFilter struct {
	Status     *Status
	Priority   *Priority
	AssignedTo *string
	LockedBy   *string
}

// Matches checks if a task matches the filter criteria
func (f *TaskFilter) Matches(task *Task) bool {
	if f.Status != nil && task.Status != *f.Status {
		return false
	}

	if f.Priority != nil && task.Priority != *f.Priority {
		return false
	}

	if f.AssignedTo != nil && task.AssignedTo != *f.AssignedTo {
		return false
	}

	if f.LockedBy != nil && task.LockedBy != *f.LockedBy {
		return false
	}

	return true
}

// TaskSorter defines how tasks should be sorted
type TaskSorter struct {
	Field string // "id", "title", "status", "priority", "created_at", "updated_at"
	Desc  bool   // true for descending order
}

// Sort sorts a slice of tasks according to the sorter criteria
func (s *TaskSorter) Sort(tasks []*Task) {
	if len(tasks) <= 1 {
		return
	}

	// Simple bubble sort for now - can be optimized later
	for i := 0; i < len(tasks)-1; i++ {
		for j := 0; j < len(tasks)-i-1; j++ {
			if s.shouldSwap(tasks[j], tasks[j+1]) {
				tasks[j], tasks[j+1] = tasks[j+1], tasks[j]
			}
		}
	}
}

// shouldSwap determines if two tasks should be swapped based on sort criteria
func (s *TaskSorter) shouldSwap(t1, t2 *Task) bool {
	var result bool

	switch s.Field {
	case "id":
		result = t1.ID > t2.ID
	case "title":
		result = t1.Title > t2.Title
	case "status":
		result = string(t1.Status) > string(t2.Status)
	case "priority":
		// Priority sorting: high > medium > low
		p1 := priorityWeight(t1.Priority)
		p2 := priorityWeight(t2.Priority)
		result = p1 > p2
	case "created_at":
		result = t1.CreatedAt.After(t2.CreatedAt)
	case "updated_at":
		result = t1.UpdatedAt.After(t2.UpdatedAt)
	default:
		result = t1.ID > t2.ID // Default to ID sorting
	}

	if s.Desc {
		return result
	}

	return !result
}

// priorityWeight returns a numeric weight for priority comparison
func priorityWeight(priority Priority) int {
	switch priority {
	case PriorityHigh:
		return 3
	case PriorityMedium:
		return 2
	case PriorityLow:
		return 1
	default:
		return 0
	}
}
