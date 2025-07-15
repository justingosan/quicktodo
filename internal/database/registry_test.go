package database

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewProjectRegistry(t *testing.T) {
	registry := NewProjectRegistry()
	
	if registry.Projects == nil {
		t.Error("Expected Projects map to be initialized")
	}
	
	if registry.PathToProject == nil {
		t.Error("Expected PathToProject map to be initialized")
	}
	
	if len(registry.Projects) != 0 {
		t.Errorf("Expected empty Projects map, got %d items", len(registry.Projects))
	}
	
	if len(registry.PathToProject) != 0 {
		t.Errorf("Expected empty PathToProject map, got %d items", len(registry.PathToProject))
	}
}

func TestProjectRegistryRegisterProject(t *testing.T) {
	registry := NewProjectRegistry()
	
	err := registry.RegisterProject("test-project", "/path/to/project")
	if err != nil {
		t.Errorf("RegisterProject failed: %v", err)
	}
	
	// Check that project was added to Projects map
	project, exists := registry.Projects["test-project"]
	if !exists {
		t.Error("Project not found in Projects map")
	}
	
	if project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", project.Name)
	}
	
	if project.Path != "/path/to/project" {
		t.Errorf("Expected project path '/path/to/project', got '%s'", project.Path)
	}
	
	// Check that project was added to PathToProject map
	name, exists := registry.PathToProject["/path/to/project"]
	if !exists {
		t.Error("Project not found in PathToProject map")
	}
	
	if name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", name)
	}
}

func TestProjectRegistryRegisterDuplicateName(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register first project
	err := registry.RegisterProject("test-project", "/path/to/project1")
	if err != nil {
		t.Errorf("First RegisterProject failed: %v", err)
	}
	
	// Try to register project with same name but different path
	err = registry.RegisterProject("test-project", "/path/to/project2")
	if err == nil {
		t.Error("Expected error when registering duplicate project name")
	}
}

func TestProjectRegistryRegisterDuplicatePath(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register first project
	err := registry.RegisterProject("project1", "/path/to/project")
	if err != nil {
		t.Errorf("First RegisterProject failed: %v", err)
	}
	
	// Try to register project with same path but different name
	err = registry.RegisterProject("project2", "/path/to/project")
	if err == nil {
		t.Error("Expected error when registering duplicate project path")
	}
}

func TestProjectRegistryGetProjectByName(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register a project
	registry.RegisterProject("test-project", "/path/to/project")
	
	// Test getting existing project
	project, exists := registry.GetProjectByName("test-project")
	if !exists {
		t.Error("Project should exist")
	}
	
	if project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", project.Name)
	}
	
	// Test getting non-existent project
	_, exists = registry.GetProjectByName("non-existent")
	if exists {
		t.Error("Non-existent project should not exist")
	}
}

func TestProjectRegistryGetProjectByPath(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register a project
	registry.RegisterProject("test-project", "/path/to/project")
	
	// Test getting existing project
	project, exists := registry.GetProjectByPath("/path/to/project")
	if !exists {
		t.Error("Project should exist")
	}
	
	if project.Name != "test-project" {
		t.Errorf("Expected project name 'test-project', got '%s'", project.Name)
	}
	
	// Test getting non-existent project
	_, exists = registry.GetProjectByPath("/non/existent/path")
	if exists {
		t.Error("Non-existent project should not exist")
	}
}

func TestProjectRegistryUpdateLastAccessed(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register a project
	registry.RegisterProject("test-project", "/path/to/project")
	
	// Get initial last accessed time
	project, _ := registry.GetProjectByName("test-project")
	initialTime := project.LastAccessed
	
	// Wait a bit to ensure time difference
	time.Sleep(time.Millisecond * 10)
	
	// Update last accessed
	err := registry.UpdateLastAccessed("test-project")
	if err != nil {
		t.Errorf("UpdateLastAccessed failed: %v", err)
	}
	
	// Check that time was updated
	updatedProject, _ := registry.GetProjectByName("test-project")
	if !updatedProject.LastAccessed.After(initialTime) {
		t.Error("LastAccessed should be updated")
	}
	
	// Test updating non-existent project
	err = registry.UpdateLastAccessed("non-existent")
	if err == nil {
		t.Error("Expected error when updating non-existent project")
	}
}

func TestProjectRegistryRemoveProject(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Register a project
	registry.RegisterProject("test-project", "/path/to/project")
	
	// Verify project exists
	_, exists := registry.GetProjectByName("test-project")
	if !exists {
		t.Error("Project should exist before removal")
	}
	
	// Remove project
	err := registry.RemoveProject("test-project")
	if err != nil {
		t.Errorf("RemoveProject failed: %v", err)
	}
	
	// Verify project is removed from both maps
	_, exists = registry.GetProjectByName("test-project")
	if exists {
		t.Error("Project should not exist after removal")
	}
	
	_, exists = registry.PathToProject["/path/to/project"]
	if exists {
		t.Error("Project path should not exist after removal")
	}
	
	// Test removing non-existent project
	err = registry.RemoveProject("non-existent")
	if err == nil {
		t.Error("Expected error when removing non-existent project")
	}
}

func TestProjectRegistryListProjects(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Initially empty
	projects := registry.ListProjects()
	if len(projects) != 0 {
		t.Errorf("Expected 0 projects, got %d", len(projects))
	}
	
	// Add some projects
	registry.RegisterProject("project1", "/path/to/project1")
	registry.RegisterProject("project2", "/path/to/project2")
	
	projects = registry.ListProjects()
	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
	
	// Check that both projects are in the list
	_, exists1 := projects["project1"]
	_, exists2 := projects["project2"]
	
	if !exists1 {
		t.Error("project1 should be in the list")
	}
	if !exists2 {
		t.Error("project2 should be in the list")
	}
}

func TestProjectRegistryValidation(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Valid empty registry should pass
	err := registry.Validate()
	if err != nil {
		t.Errorf("Valid empty registry failed validation: %v", err)
	}
	
	// Add valid project
	registry.RegisterProject("test-project", "/path/to/project")
	
	err = registry.Validate()
	if err != nil {
		t.Errorf("Valid registry with project failed validation: %v", err)
	}
	
	// Create inconsistent state (this should not happen in normal usage)
	registry.Projects["orphan"] = &ProjectInfo{
		Name:         "orphan",
		Path:         "/orphan/path",
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}
	
	err = registry.Validate()
	if err == nil {
		t.Error("Expected validation error for inconsistent state")
	}
}

func TestProjectRegistryCleanup(t *testing.T) {
	registry := NewProjectRegistry()
	
	// Add some projects with paths that don't exist
	registry.RegisterProject("project1", "/non/existent/path1")
	registry.RegisterProject("project2", "/non/existent/path2")
	
	// Create a temporary directory for project3
	tempDir, err := os.MkdirTemp("", "quicktodo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	registry.RegisterProject("project3", tempDir)
	
	// Run cleanup
	removed, err := registry.Cleanup()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	}
	
	// Should have removed 2 projects (the ones with non-existent paths)
	if len(removed) != 2 {
		t.Errorf("Expected 2 removed projects, got %d", len(removed))
	}
	
	// project3 should still exist
	_, exists := registry.GetProjectByName("project3")
	if !exists {
		t.Error("project3 should still exist after cleanup")
	}
	
	// Other projects should be removed
	_, exists = registry.GetProjectByName("project1")
	if exists {
		t.Error("project1 should be removed after cleanup")
	}
	
	_, exists = registry.GetProjectByName("project2")
	if exists {
		t.Error("project2 should be removed after cleanup")
	}
}

func TestProjectRegistrySaveAndLoad(t *testing.T) {
	// Create a temporary file for testing
	tempFile := filepath.Join(os.TempDir(), "test-registry.json")
	defer os.Remove(tempFile)
	
	// Create registry and add some projects
	registry := NewProjectRegistry()
	registry.RegisterProject("project1", "/path/to/project1")
	registry.RegisterProject("project2", "/path/to/project2")
	
	// Save registry
	err := registry.Save(tempFile)
	if err != nil {
		t.Errorf("Save failed: %v", err)
	}
	
	// Verify file exists
	if _, err := os.Stat(tempFile); os.IsNotExist(err) {
		t.Error("Registry file was not created")
	}
	
	// Load registry
	loadedRegistry, err := LoadProjectRegistry(tempFile)
	if err != nil {
		t.Errorf("LoadProjectRegistry failed: %v", err)
	}
	
	// Verify loaded registry has same data
	if len(loadedRegistry.Projects) != 2 {
		t.Errorf("Expected 2 projects in loaded registry, got %d", len(loadedRegistry.Projects))
	}
	
	project1, exists := loadedRegistry.GetProjectByName("project1")
	if !exists {
		t.Error("project1 should exist in loaded registry")
	}
	if project1.Path != "/path/to/project1" {
		t.Errorf("Expected project1 path '/path/to/project1', got '%s'", project1.Path)
	}
	
	project2, exists := loadedRegistry.GetProjectByName("project2")
	if !exists {
		t.Error("project2 should exist in loaded registry")
	}
	if project2.Path != "/path/to/project2" {
		t.Errorf("Expected project2 path '/path/to/project2', got '%s'", project2.Path)
	}
}

func TestLoadProjectRegistryNonExistentFile(t *testing.T) {
	// Try to load non-existent file
	_, err := LoadProjectRegistry("/non/existent/file.json")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
}

func TestLoadProjectRegistryInvalidJSON(t *testing.T) {
	// Create a temporary file with invalid JSON
	tempFile := filepath.Join(os.TempDir(), "invalid-registry.json")
	defer os.Remove(tempFile)
	
	err := os.WriteFile(tempFile, []byte("invalid json"), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	
	_, err = LoadProjectRegistry(tempFile)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}

func TestProjectInfoJSON(t *testing.T) {
	info := &ProjectInfo{
		Name:         "test-project",
		Path:         "/path/to/project",
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(info)
	if err != nil {
		t.Errorf("JSON marshal failed: %v", err)
	}
	
	// Test JSON unmarshaling
	var newInfo ProjectInfo
	err = json.Unmarshal(jsonData, &newInfo)
	if err != nil {
		t.Errorf("JSON unmarshal failed: %v", err)
	}
	
	if newInfo.Name != info.Name {
		t.Errorf("Expected name '%s', got '%s'", info.Name, newInfo.Name)
	}
	
	if newInfo.Path != info.Path {
		t.Errorf("Expected path '%s', got '%s'", info.Path, newInfo.Path)
	}
}