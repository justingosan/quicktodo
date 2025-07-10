package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the global configuration
type Config struct {
	DataDir        string `json:"data_dir"`
	LockTimeout    int    `json:"lock_timeout"`    // in seconds
	StaleTimeout   int    `json:"stale_timeout"`   // in minutes
	DefaultPriority string `json:"default_priority"`
	CreateBackups  bool   `json:"create_backups"`
	MaxBackups     int    `json:"max_backups"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}
	
	return &Config{
		DataDir:        filepath.Join(homeDir, ".config", "quicktodo"),
		LockTimeout:    30,
		StaleTimeout:   5,
		DefaultPriority: "medium",
		CreateBackups:  true,
		MaxBackups:     5,
	}
}

// GetConfigPath returns the path to the configuration file
func GetConfigPath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(homeDir, ".config", "quicktodo", "config.json")
}

// Load loads the configuration from file or creates default if not exists
func Load() (*Config, error) {
	configPath := GetConfigPath()
	
	// If config file doesn't exist, create it with defaults
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return config, nil
	}
	
	// Read existing config
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}
	
	// Validate and set defaults for missing fields
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := GetConfigPath()
	
	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	// Marshal config to JSON
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}
	
	return nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.DataDir == "" {
		return fmt.Errorf("data_dir cannot be empty")
	}
	
	if c.LockTimeout <= 0 {
		c.LockTimeout = 30
	}
	
	if c.StaleTimeout <= 0 {
		c.StaleTimeout = 5
	}
	
	if c.DefaultPriority == "" {
		c.DefaultPriority = "medium"
	}
	
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"high":   true,
	}
	
	if !validPriorities[c.DefaultPriority] {
		return fmt.Errorf("invalid default_priority: %s (must be low, medium, or high)", c.DefaultPriority)
	}
	
	if c.MaxBackups < 0 {
		c.MaxBackups = 5
	}
	
	return nil
}

// EnsureDataDir ensures the data directory exists
func (c *Config) EnsureDataDir() error {
	return os.MkdirAll(c.DataDir, 0755)
}

// GetProjectsPath returns the path to the projects registry file
func (c *Config) GetProjectsPath() string {
	return filepath.Join(c.DataDir, "projects.json")
}

// GetProjectDatabasePath returns the path to a project's database file
func (c *Config) GetProjectDatabasePath(projectName string) string {
	return filepath.Join(c.DataDir, "projects", projectName+".json")
}

// GetProjectLockPath returns the path to a project's lock file
func (c *Config) GetProjectLockPath(projectName string) string {
	return filepath.Join(c.DataDir, "locks", projectName+".lock")
}

// EnsureAllDirectories ensures all required directories exist
func (c *Config) EnsureAllDirectories() error {
	dirs := []string{
		c.DataDir,
		filepath.Join(c.DataDir, "projects"),
		filepath.Join(c.DataDir, "locks"),
	}
	
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	
	return nil
}