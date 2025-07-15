package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	if config.DataDir == "" {
		t.Error("Expected DataDir to be set")
	}
	
	if config.LockTimeout <= 0 {
		t.Errorf("Expected positive LockTimeout, got %d", config.LockTimeout)
	}
	
	if config.StaleTimeout <= 0 {
		t.Errorf("Expected positive StaleTimeout, got %d", config.StaleTimeout)
	}
	
	if config.DefaultPriority == "" {
		t.Error("Expected DefaultPriority to be set")
	}
	
	if config.MaxBackups <= 0 {
		t.Errorf("Expected positive MaxBackups, got %d", config.MaxBackups)
	}
	
	// Test that CreateBackups is set to a boolean value (should be true by default)
	if !config.CreateBackups {
		t.Error("Expected CreateBackups to be true by default")
	}
}

func TestConfigGetProjectsPath(t *testing.T) {
	config := DefaultConfig()
	
	projectsPath := config.GetProjectsPath()
	if projectsPath == "" {
		t.Error("Expected GetProjectsPath to return non-empty string")
	}
	
	// Should be within the data directory
	if !filepath.IsAbs(projectsPath) {
		t.Error("Expected absolute path from GetProjectsPath")
	}
	
	expectedPath := filepath.Join(config.DataDir, "projects.json")
	if projectsPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, projectsPath)
	}
}

func TestConfigGetProjectDatabasePath(t *testing.T) {
	config := DefaultConfig()
	
	projectName := "test-project"
	dbPath := config.GetProjectDatabasePath(projectName)
	
	if dbPath == "" {
		t.Error("Expected GetProjectDatabasePath to return non-empty string")
	}
	
	if !filepath.IsAbs(dbPath) {
		t.Error("Expected absolute path from GetProjectDatabasePath")
	}
	
	expectedPath := filepath.Join(config.DataDir, "projects", projectName+".json")
	if dbPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, dbPath)
	}
}

func TestConfigEnsureAllDirectories(t *testing.T) {
	// Create a config with a temporary data directory
	tempDir, err := os.MkdirTemp("", "quicktodo-config-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	config := &Config{
		DataDir:         tempDir,
		LockTimeout:     30,
		StaleTimeout:    300,
		DefaultPriority: "medium",
		CreateBackups:   true,
		MaxBackups:      5,
	}
	
	err = config.EnsureAllDirectories()
	if err != nil {
		t.Errorf("EnsureAllDirectories failed: %v", err)
	}
	
	// Check that directories were created
	projectsDir := filepath.Join(tempDir, "projects")
	if _, err := os.Stat(projectsDir); os.IsNotExist(err) {
		t.Error("Projects directory was not created")
	}
	
	locksDir := filepath.Join(tempDir, "locks")
	if _, err := os.Stat(locksDir); os.IsNotExist(err) {
		t.Error("Locks directory was not created")
	}
}

func TestConfigLoad(t *testing.T) {
	// Test loading config (should create default if not exists)
	config, err := Load()
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}
	
	if config == nil {
		t.Error("Expected config to be non-nil")
	}
	
	// Should have default values
	defaultConfig := DefaultConfig()
	if config.LockTimeout != defaultConfig.LockTimeout {
		t.Errorf("Expected default LockTimeout %d, got %d", defaultConfig.LockTimeout, config.LockTimeout)
	}
}

func TestConfigJSON(t *testing.T) {
	config := DefaultConfig()
	
	// Test JSON marshaling
	jsonData, err := json.Marshal(config)
	if err != nil {
		t.Errorf("JSON marshal failed: %v", err)
	}
	
	// Test JSON unmarshaling
	var newConfig Config
	err = json.Unmarshal(jsonData, &newConfig)
	if err != nil {
		t.Errorf("JSON unmarshal failed: %v", err)
	}
	
	// Verify fields match
	if newConfig.DataDir != config.DataDir {
		t.Errorf("Expected DataDir '%s', got '%s'", config.DataDir, newConfig.DataDir)
	}
	
	if newConfig.LockTimeout != config.LockTimeout {
		t.Errorf("Expected LockTimeout %d, got %d", config.LockTimeout, newConfig.LockTimeout)
	}
	
	if newConfig.DefaultPriority != config.DefaultPriority {
		t.Errorf("Expected DefaultPriority '%s', got '%s'", config.DefaultPriority, newConfig.DefaultPriority)
	}
}

func TestGetConfigPath(t *testing.T) {
	configPath := GetConfigPath()
	
	if configPath == "" {
		t.Error("Expected getConfigPath to return non-empty string")
	}
	
	if !filepath.IsAbs(configPath) {
		t.Error("Expected absolute path from getConfigPath")
	}
	
	// Should end with config.json
	if filepath.Base(configPath) != "config.json" {
		t.Errorf("Expected config path to end with 'config.json', got '%s'", configPath)
	}
}