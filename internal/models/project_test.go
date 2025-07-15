package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewProject(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	
	if project.Name != "test-project" {
		t.Errorf("Expected name 'test-project', got '%s'", project.Name)
	}
	
	if project.Path == "" {
		t.Error("Expected path to be set")
	}
	
	if project.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	
	if project.LastAccessed.IsZero() {
		t.Error("Expected LastAccessed to be set")
	}
	
	if project.TaskCount != 0 {
		t.Errorf("Expected TaskCount 0, got %d", project.TaskCount)
	}
}

func TestNewProjectDatabase(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	if db.Project != project {
		t.Error("Expected project to be set in database")
	}
	
	if db.NextID != 1 {
		t.Errorf("Expected NextID 1, got %d", db.NextID)
	}
	
	if len(db.Tasks) != 0 {
		t.Errorf("Expected empty tasks slice, got %d tasks", len(db.Tasks))
	}
	
	if db.Version != 1 {
		t.Errorf("Expected Version 1, got %d", db.Version)
	}
	
	if db.LastModified.IsZero() {
		t.Error("Expected LastModified to be set")
	}
}

func TestProjectDatabaseAddTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task := NewTask(1, "Test Task")
	
	err := db.AddTask(task)
	if err != nil {
		t.Errorf("AddTask failed: %v", err)
	}
	
	if len(db.Tasks) != 1 {
		t.Errorf("Expected 1 task, got %d", len(db.Tasks))
	}
	
	if db.Tasks[0] != task {
		t.Error("Task not added correctly")
	}
	
	if db.NextID != 2 {
		t.Errorf("Expected NextID 2 after adding task, got %d", db.NextID)
	}
}

func TestProjectDatabaseAddTaskNil(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	// Try to add nil task
	err := db.AddTask(nil)
	if err == nil {
		t.Error("Expected error when adding nil task")
	}
}

func TestProjectDatabaseGetTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task1 := NewTask(1, "Task 1")
	task2 := NewTask(2, "Task 2")
	
	db.AddTask(task1)
	db.AddTask(task2)
	
	// Test getting existing task
	found, err := db.GetTask(1)
	if err != nil {
		t.Errorf("GetTask failed: %v", err)
	}
	if found.ID != 1 {
		t.Errorf("Expected task ID 1, got %d", found.ID)
	}
	
	// Test getting non-existent task
	_, err = db.GetTask(999)
	if err == nil {
		t.Error("Expected error for non-existent task")
	}
}

func TestProjectDatabaseUpdateTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task := NewTask(1, "Original Task")
	db.AddTask(task)
	
	// Update the task
	task.Title = "Updated Task"
	task.UpdatedAt = time.Now()
	
	err := db.UpdateTask(task)
	if err != nil {
		t.Errorf("UpdateTask failed: %v", err)
	}
	
	// Verify update
	found, _ := db.GetTask(1)
	if found.Title != "Updated Task" {
		t.Errorf("Expected title 'Updated Task', got '%s'", found.Title)
	}
}

func TestProjectDatabaseUpdateNonExistentTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task := NewTask(999, "Non-existent Task")
	
	err := db.UpdateTask(task)
	if err == nil {
		t.Error("Expected error when updating non-existent task")
	}
}

func TestProjectDatabaseDeleteTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task1 := NewTask(1, "Task 1")
	task2 := NewTask(2, "Task 2")
	
	db.AddTask(task1)
	db.AddTask(task2)
	
	// Delete task 1
	err := db.DeleteTask(1)
	if err != nil {
		t.Errorf("DeleteTask failed: %v", err)
	}
	
	if len(db.Tasks) != 1 {
		t.Errorf("Expected 1 task after deletion, got %d", len(db.Tasks))
	}
	
	// Verify correct task was deleted
	if db.Tasks[0].ID != 2 {
		t.Errorf("Expected remaining task to have ID 2, got %d", db.Tasks[0].ID)
	}
	
	// Try to get deleted task
	_, err = db.GetTask(1)
	if err == nil {
		t.Error("Expected error when getting deleted task")
	}
}

func TestProjectDatabaseDeleteNonExistentTask(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	err := db.DeleteTask(999)
	if err == nil {
		t.Error("Expected error when deleting non-existent task")
	}
}

func TestProjectDatabaseListTasks(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task1 := NewTaskWithDetails(1, "Task 1", "", PriorityHigh)
	task2 := NewTaskWithDetails(2, "Task 2", "", PriorityLow)
	task3 := NewTaskWithDetails(3, "Task 3", "", PriorityMedium)
	
	task1.UpdateStatus(StatusDone)
	task2.UpdateStatus(StatusInProgress)
	// task3 remains pending
	
	db.AddTask(task1)
	db.AddTask(task2)
	db.AddTask(task3)
	
	// Test listing all tasks
	allTasks := db.ListTasks(nil)
	if len(allTasks) != 3 {
		t.Errorf("Expected 3 tasks, got %d", len(allTasks))
	}
	
	// Test filtering by status
	pendingStatus := StatusPending
	pendingFilter := &TaskFilter{Status: &pendingStatus}
	pendingTasks := db.ListTasks(pendingFilter)
	if len(pendingTasks) != 1 {
		t.Errorf("Expected 1 pending task, got %d", len(pendingTasks))
	}
	if pendingTasks[0].ID != 3 {
		t.Errorf("Expected pending task ID 3, got %d", pendingTasks[0].ID)
	}
	
	// Test filtering by priority
	highPriority := PriorityHigh
	highPriorityFilter := &TaskFilter{Priority: &highPriority}
	highTasks := db.ListTasks(highPriorityFilter)
	if len(highTasks) != 1 {
		t.Errorf("Expected 1 high priority task, got %d", len(highTasks))
	}
	if highTasks[0].ID != 1 {
		t.Errorf("Expected high priority task ID 1, got %d", highTasks[0].ID)
	}
	
	// Test filtering by assigned user
	task1.AssignTo("user123")
	db.UpdateTask(task1)
	
	assignedFilter := &TaskFilter{AssignedTo: stringPtr("user123")}
	assignedTasks := db.ListTasks(assignedFilter)
	if len(assignedTasks) != 1 {
		t.Errorf("Expected 1 assigned task, got %d", len(assignedTasks))
	}
}

func TestProjectDatabaseSearchTasks(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task1 := NewTaskWithDetails(1, "Login Bug Fix", "Fix authentication issue", PriorityHigh)
	task2 := NewTaskWithDetails(2, "User Interface", "Update login page", PriorityMedium)
	task3 := NewTaskWithDetails(3, "Database Migration", "Update schema", PriorityLow)
	
	db.AddTask(task1)
	db.AddTask(task2)
	db.AddTask(task3)
	
	// Search for "login"
	results := db.SearchTasks("login")
	if len(results) != 2 {
		t.Errorf("Expected 2 results for 'login', got %d", len(results))
	}
	
	// Search for "database"
	results = db.SearchTasks("database")
	if len(results) != 1 {
		t.Errorf("Expected 1 result for 'database', got %d", len(results))
	}
	if results[0].ID != 3 {
		t.Errorf("Expected database task ID 3, got %d", results[0].ID)
	}
	
	// Search for non-existent term
	results = db.SearchTasks("nonexistent")
	if len(results) != 0 {
		t.Errorf("Expected 0 results for 'nonexistent', got %d", len(results))
	}
}

func TestProjectDatabaseGetSummary(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task1 := NewTaskWithDetails(1, "Task 1", "", PriorityHigh)
	task2 := NewTaskWithDetails(2, "Task 2", "", PriorityMedium)
	task3 := NewTaskWithDetails(3, "Task 3", "", PriorityLow)
	
	task1.UpdateStatus(StatusDone)
	task2.UpdateStatus(StatusInProgress)
	// task3 remains pending
	
	db.AddTask(task1)
	db.AddTask(task2)
	db.AddTask(task3)
	
	summary := db.GetSummary()
	
	if summary.TaskCount != 3 {
		t.Errorf("Expected TaskCount 3, got %d", summary.TaskCount)
	}
	
	if summary.CompletedTasks != 1 {
		t.Errorf("Expected CompletedTasks 1, got %d", summary.CompletedTasks)
	}
	
	if summary.PendingTasks != 1 {
		t.Errorf("Expected PendingTasks 1, got %d", summary.PendingTasks)
	}
	
	if summary.InProgressTasks != 1 {
		t.Errorf("Expected InProgressTasks 1, got %d", summary.InProgressTasks)
	}
	
	// Check status counts
	if summary.StatusCounts[StatusDone] != 1 {
		t.Errorf("Expected 1 done task in status counts, got %d", summary.StatusCounts[StatusDone])
	}
	
	if summary.StatusCounts[StatusInProgress] != 1 {
		t.Errorf("Expected 1 in_progress task in status counts, got %d", summary.StatusCounts[StatusInProgress])
	}
	
	if summary.StatusCounts[StatusPending] != 1 {
		t.Errorf("Expected 1 pending task in status counts, got %d", summary.StatusCounts[StatusPending])
	}
	
	// Check priority counts
	if summary.PriorityCounts[PriorityHigh] != 1 {
		t.Errorf("Expected 1 high priority task, got %d", summary.PriorityCounts[PriorityHigh])
	}
	
	if summary.PriorityCounts[PriorityMedium] != 1 {
		t.Errorf("Expected 1 medium priority task, got %d", summary.PriorityCounts[PriorityMedium])
	}
	
	if summary.PriorityCounts[PriorityLow] != 1 {
		t.Errorf("Expected 1 low priority task, got %d", summary.PriorityCounts[PriorityLow])
	}
}

func TestProjectDatabaseValidation(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	// Valid database should pass validation
	err := db.Validate()
	if err != nil {
		t.Errorf("Valid database failed validation: %v", err)
	}
	
	// Database with nil project should fail
	db.Project = nil
	err = db.Validate()
	if err == nil {
		t.Error("Expected validation error for nil project")
	}
	
	// Reset project
	db.Project = project
	
	// Add invalid task
	invalidTask := &Task{
		ID:       1,
		Title:    "", // Empty title
		Status:   StatusPending,
		Priority: PriorityMedium,
	}
	db.Tasks = []*Task{invalidTask}
	
	err = db.Validate()
	if err == nil {
		t.Error("Expected validation error for invalid task")
	}
}

func TestProjectDatabaseJSON(t *testing.T) {
	project := NewProject("test-project", "/path/to/project")
	db := NewProjectDatabase(project)
	
	task := NewTask(1, "Test Task")
	db.AddTask(task)
	
	// Test ToJSON
	jsonData, err := db.ToJSON()
	if err != nil {
		t.Errorf("ToJSON failed: %v", err)
	}
	
	if len(jsonData) == 0 {
		t.Error("Expected non-empty JSON data")
	}
	
	// Test that JSON is valid
	var temp interface{}
	err = json.Unmarshal(jsonData, &temp)
	if err != nil {
		t.Errorf("Invalid JSON produced: %v", err)
	}
	
	// Test that we can unmarshal back to ProjectDatabase
	var newDB ProjectDatabase
	err = json.Unmarshal(jsonData, &newDB)
	if err != nil {
		t.Errorf("Failed to unmarshal JSON back to ProjectDatabase: %v", err)
	}
	
	if newDB.Project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", newDB.Project.Name)
	}
	
	if len(newDB.Tasks) != 1 {
		t.Errorf("Expected 1 task in unmarshaled database, got %d", len(newDB.Tasks))
	}
}

// Helper function for tests
func stringPtr(s string) *string {
	return &s
}