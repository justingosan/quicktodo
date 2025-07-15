package models

import (
	"testing"
	"time"
)

func TestNewTask(t *testing.T) {
	task := NewTask(1, "Test Task")
	
	if task.ID != 1 {
		t.Errorf("Expected ID 1, got %d", task.ID)
	}
	
	if task.Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", task.Title)
	}
	
	if task.Status != StatusPending {
		t.Errorf("Expected status pending, got %s", task.Status)
	}
	
	if task.Priority != PriorityMedium {
		t.Errorf("Expected priority medium, got %s", task.Priority)
	}
	
	if task.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	
	if task.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestNewTaskWithDetails(t *testing.T) {
	task := NewTaskWithDetails(2, "Detailed Task", "Task description", PriorityHigh)
	
	if task.ID != 2 {
		t.Errorf("Expected ID 2, got %d", task.ID)
	}
	
	if task.Title != "Detailed Task" {
		t.Errorf("Expected title 'Detailed Task', got '%s'", task.Title)
	}
	
	if task.Description != "Task description" {
		t.Errorf("Expected description 'Task description', got '%s'", task.Description)
	}
	
	if task.Priority != PriorityHigh {
		t.Errorf("Expected priority high, got %s", task.Priority)
	}
}

func TestTaskValidation(t *testing.T) {
	tests := []struct {
		name    string
		task    *Task
		wantErr bool
	}{
		{
			name: "valid task",
			task: &Task{
				ID:        1,
				Title:     "Valid Task",
				Status:    StatusPending,
				Priority:  PriorityMedium,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "empty title",
			task: &Task{
				ID:       1,
				Title:    "",
				Status:   StatusPending,
				Priority: PriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			task: &Task{
				ID:       1,
				Title:    "Test Task",
				Status:   "invalid",
				Priority: PriorityMedium,
			},
			wantErr: true,
		},
		{
			name: "invalid priority",
			task: &Task{
				ID:       1,
				Title:    "Test Task",
				Status:   StatusPending,
				Priority: "invalid",
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Task.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestIsValidStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{"pending", true},
		{"in_progress", true},
		{"done", true},
		{"invalid", false},
		{"", false},
		{"PENDING", false}, // case sensitive
	}
	
	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			if got := IsValidStatus(tt.status); got != tt.want {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestIsValidPriority(t *testing.T) {
	tests := []struct {
		priority string
		want     bool
	}{
		{"low", true},
		{"medium", true},
		{"high", true},
		{"invalid", false},
		{"", false},
		{"LOW", false}, // case sensitive
	}
	
	for _, tt := range tests {
		t.Run(tt.priority, func(t *testing.T) {
			if got := IsValidPriority(tt.priority); got != tt.want {
				t.Errorf("IsValidPriority(%q) = %v, want %v", tt.priority, got, tt.want)
			}
		})
	}
}

func TestTaskStatusUpdates(t *testing.T) {
	task := NewTask(1, "Test Task")
	originalTime := task.UpdatedAt
	
	// Wait a small amount to ensure time difference
	time.Sleep(time.Millisecond)
	
	err := task.UpdateStatus(StatusInProgress)
	if err != nil {
		t.Errorf("UpdateStatus failed: %v", err)
	}
	
	if task.Status != StatusInProgress {
		t.Errorf("Expected status in_progress, got %s", task.Status)
	}
	
	if !task.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated when status changes")
	}
	
	// Test invalid status
	err = task.UpdateStatus("invalid")
	if err == nil {
		t.Error("Expected error for invalid status")
	}
}

func TestTaskPriorityUpdates(t *testing.T) {
	task := NewTask(1, "Test Task")
	originalTime := task.UpdatedAt
	
	time.Sleep(time.Millisecond)
	
	err := task.UpdatePriority(PriorityHigh)
	if err != nil {
		t.Errorf("UpdatePriority failed: %v", err)
	}
	
	if task.Priority != PriorityHigh {
		t.Errorf("Expected priority high, got %s", task.Priority)
	}
	
	if !task.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated when priority changes")
	}
	
	// Test invalid priority
	err = task.UpdatePriority("invalid")
	if err == nil {
		t.Error("Expected error for invalid priority")
	}
}

func TestTaskTitleUpdates(t *testing.T) {
	task := NewTask(1, "Original Title")
	originalTime := task.UpdatedAt
	
	time.Sleep(time.Millisecond)
	
	err := task.UpdateTitle("New Title")
	if err != nil {
		t.Errorf("UpdateTitle failed: %v", err)
	}
	
	if task.Title != "New Title" {
		t.Errorf("Expected title 'New Title', got '%s'", task.Title)
	}
	
	if !task.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated when title changes")
	}
	
	// Test empty title
	err = task.UpdateTitle("")
	if err == nil {
		t.Error("Expected error for empty title")
	}
}

func TestTaskDescriptionUpdates(t *testing.T) {
	task := NewTask(1, "Test Task")
	originalTime := task.UpdatedAt
	
	time.Sleep(time.Millisecond)
	
	task.UpdateDescription("New description")
	
	if task.Description != "New description" {
		t.Errorf("Expected description 'New description', got '%s'", task.Description)
	}
	
	if !task.UpdatedAt.After(originalTime) {
		t.Error("UpdatedAt should be updated when description changes")
	}
}

func TestTaskAssignment(t *testing.T) {
	task := NewTask(1, "Test Task")
	
	task.AssignTo("user123")
	
	if task.AssignedTo != "user123" {
		t.Errorf("Expected assigned to 'user123', got '%s'", task.AssignedTo)
	}
}

func TestTaskLocking(t *testing.T) {
	task := NewTask(1, "Test Task")
	
	// Initially unlocked
	if task.IsLocked() {
		t.Error("Task should not be locked initially")
	}
	
	// Lock the task
	task.Lock("process123")
	
	if !task.IsLocked() {
		t.Error("Task should be locked after Lock()")
	}
	
	if task.LockedBy != "process123" {
		t.Errorf("Expected locked by 'process123', got '%s'", task.LockedBy)
	}
	
	if !task.IsLockedBy("process123") {
		t.Error("Task should be locked by 'process123'")
	}
	
	if task.IsLockedBy("other") {
		t.Error("Task should not be locked by 'other'")
	}
	
	// Unlock the task
	task.Unlock()
	
	if task.IsLocked() {
		t.Error("Task should not be locked after Unlock()")
	}
	
	if task.LockedBy != "" {
		t.Errorf("Expected empty LockedBy after unlock, got '%s'", task.LockedBy)
	}
}

func TestTaskCompletion(t *testing.T) {
	task := NewTask(1, "Test Task")
	
	// Initially not complete
	if task.IsComplete() {
		t.Error("Task should not be complete initially")
	}
	
	// Mark as done
	task.UpdateStatus(StatusDone)
	
	if !task.IsComplete() {
		t.Error("Task should be complete when status is done")
	}
}

func TestTaskClone(t *testing.T) {
	original := NewTaskWithDetails(1, "Original Task", "Description", PriorityHigh)
	original.AssignTo("user123")
	original.Lock("process123")
	
	clone := original.Clone()
	
	// Check that all fields are copied
	if clone.ID != original.ID {
		t.Error("Clone ID mismatch")
	}
	if clone.Title != original.Title {
		t.Error("Clone Title mismatch")
	}
	if clone.Description != original.Description {
		t.Error("Clone Description mismatch")
	}
	if clone.Status != original.Status {
		t.Error("Clone Status mismatch")
	}
	if clone.Priority != original.Priority {
		t.Error("Clone Priority mismatch")
	}
	if clone.AssignedTo != original.AssignedTo {
		t.Error("Clone AssignedTo mismatch")
	}
	if clone.LockedBy != original.LockedBy {
		t.Error("Clone LockedBy mismatch")
	}
	
	// Check that modifying clone doesn't affect original
	clone.Title = "Modified Title"
	if original.Title == "Modified Title" {
		t.Error("Modifying clone affected original")
	}
}

func TestTaskGetAge(t *testing.T) {
	task := NewTask(1, "Test Task")
	
	age := task.GetAge()
	if age == "" {
		t.Error("GetAge should return non-empty string")
	}
	
	// Should contain time information
	if len(age) < 3 {
		t.Errorf("Expected meaningful age string, got '%s'", age)
	}
}