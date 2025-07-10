package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// ProjectRegistry manages the projects.json file
type ProjectRegistry struct {
	Projects      map[string]*ProjectInfo `json:"projects"`
	PathToProject map[string]string       `json:"path_to_project"`
}

// ProjectInfo contains information about a registered project
type ProjectInfo struct {
	Path         string    `json:"path"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	LastAccessed time.Time `json:"last_accessed"`
}

// NewProjectRegistry creates a new empty project registry
func NewProjectRegistry() *ProjectRegistry {
	return &ProjectRegistry{
		Projects:      make(map[string]*ProjectInfo),
		PathToProject: make(map[string]string),
	}
}

// LoadProjectRegistry loads the project registry from file
func LoadProjectRegistry(filePath string) (*ProjectRegistry, error) {
	// If file doesn't exist, create empty registry
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		registry := NewProjectRegistry()
		if err := registry.Save(filePath); err != nil {
			return nil, fmt.Errorf("failed to create empty registry: %w", err)
		}
		return registry, nil
	}

	// Read existing registry
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read registry file: %w", err)
	}

	var registry ProjectRegistry
	if err := json.Unmarshal(data, &registry); err != nil {
		return nil, fmt.Errorf("failed to parse registry file: %w", err)
	}

	// Initialize maps if they are nil
	if registry.Projects == nil {
		registry.Projects = make(map[string]*ProjectInfo)
	}
	if registry.PathToProject == nil {
		registry.PathToProject = make(map[string]string)
	}

	return &registry, nil
}

// Save saves the project registry to file
func (r *ProjectRegistry) Save(filePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal registry to JSON
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal registry: %w", err)
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

// RegisterProject registers a new project in the registry
func (r *ProjectRegistry) RegisterProject(name, path string) error {
	// Convert path to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Check if project already exists
	if _, exists := r.Projects[name]; exists {
		return fmt.Errorf("project %s already exists", name)
	}

	// Check if path is already registered
	if existingName, exists := r.PathToProject[absPath]; exists {
		return fmt.Errorf("path %s is already registered as project %s", absPath, existingName)
	}

	// Create project info
	projectInfo := &ProjectInfo{
		Path:         absPath,
		Name:         name,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
	}

	// Add to registry
	r.Projects[name] = projectInfo
	r.PathToProject[absPath] = name

	return nil
}

// GetProjectByName returns project info by name
func (r *ProjectRegistry) GetProjectByName(name string) (*ProjectInfo, bool) {
	project, exists := r.Projects[name]
	return project, exists
}

// GetProjectByPath returns project info by path
func (r *ProjectRegistry) GetProjectByPath(path string) (*ProjectInfo, bool) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, false
	}

	if projectName, exists := r.PathToProject[absPath]; exists {
		return r.Projects[projectName], true
	}

	return nil, false
}

// UpdateLastAccessed updates the last accessed time for a project
func (r *ProjectRegistry) UpdateLastAccessed(name string) error {
	if project, exists := r.Projects[name]; exists {
		project.LastAccessed = time.Now()
		return nil
	}
	return fmt.Errorf("project %s not found", name)
}

// RemoveProject removes a project from the registry
func (r *ProjectRegistry) RemoveProject(name string) error {
	project, exists := r.Projects[name]
	if !exists {
		return fmt.Errorf("project %s not found", name)
	}

	// Remove from both maps
	delete(r.Projects, name)
	delete(r.PathToProject, project.Path)

	return nil
}

// ListProjects returns all registered projects
func (r *ProjectRegistry) ListProjects() map[string]*ProjectInfo {
	// Return a copy to prevent external modification
	projects := make(map[string]*ProjectInfo)
	for name, info := range r.Projects {
		projects[name] = info
	}
	return projects
}

// Validate validates the registry structure
func (r *ProjectRegistry) Validate() error {
	// Check for consistency between the two maps
	for name, info := range r.Projects {
		if mappedName, exists := r.PathToProject[info.Path]; !exists || mappedName != name {
			return fmt.Errorf("inconsistent registry: project %s path %s", name, info.Path)
		}
	}

	for path, name := range r.PathToProject {
		if info, exists := r.Projects[name]; !exists || info.Path != path {
			return fmt.Errorf("inconsistent registry: path %s project %s", path, name)
		}
	}

	return nil
}

// Cleanup removes projects that point to non-existent directories
func (r *ProjectRegistry) Cleanup() ([]string, error) {
	var removed []string

	for name, info := range r.Projects {
		if _, err := os.Stat(info.Path); os.IsNotExist(err) {
			delete(r.Projects, name)
			delete(r.PathToProject, info.Path)
			removed = append(removed, name)
		}
	}

	return removed, nil
}
