package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
	"strings"

	"github.com/spf13/cobra"
)

// initProjectCmd represents the init command
var initProjectCmd = &cobra.Command{
	Use:   "init [project_name]",
	Short: "Initialize current directory as a QuickTodo project",
	Long: `Initialize the current directory as a QuickTodo project. 

This command:
- Creates a new project entry in the registry
- Initializes the project database

If no project name is provided, it will use the current directory name.
Use 'quicktodo context' to see AI usage instructions.

Examples:
  quicktodo init myproject
  quicktodo init
  quicktodo init "My Amazing Project"`,
	Args: cobra.MaximumNArgs(1),
	Run:  runInitProject,
}

func runInitProject(cmd *cobra.Command, args []string) {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Ensure all directories exist
	if err := cfg.EnsureAllDirectories(); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directories: %v\n", err)
		os.Exit(1)
	}

	// Get current directory
	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)
		os.Exit(1)
	}

	// Determine project name
	var projectName string
	if len(args) > 0 {
		projectName = strings.TrimSpace(args[0])
	} else {
		projectName = filepath.Base(currentDir)
	}

	if projectName == "" {
		fmt.Fprintf(os.Stderr, "Error: project name cannot be empty\n")
		os.Exit(1)
	}

	// Validate project name
	if err := validateProjectName(projectName); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Load project registry
	registryPath := cfg.GetProjectsPath()
	registry, err := database.LoadProjectRegistry(registryPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading project registry: %v\n", err)
		os.Exit(1)
	}

	// Check if project already exists
	if _, exists := registry.GetProjectByName(projectName); exists {
		fmt.Fprintf(os.Stderr, "Error: project '%s' already exists\n", projectName)
		os.Exit(1)
	}

	// Check if current directory is already registered
	if existingProject, exists := registry.GetProjectByPath(currentDir); exists {
		fmt.Fprintf(os.Stderr, "Error: directory '%s' is already registered as project '%s'\n",
			currentDir, existingProject.Name)
		os.Exit(1)
	}

	// Register project
	if err := registry.RegisterProject(projectName, currentDir); err != nil {
		fmt.Fprintf(os.Stderr, "Error registering project: %v\n", err)
		os.Exit(1)
	}

	// Save updated registry
	if err := registry.Save(registryPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving project registry: %v\n", err)
		os.Exit(1)
	}

	// Create project database
	project := models.NewProject(projectName, currentDir)
	projectDB := models.NewProjectDatabase(project)

	// Save project database
	dbPath := cfg.GetProjectDatabasePath(projectName)
	if err := saveProjectDatabase(projectDB, dbPath); err != nil {
		// Try to rollback registry change
		registry.RemoveProject(projectName)
		registry.Save(registryPath)
		fmt.Fprintf(os.Stderr, "Error creating project database: %v\n", err)
		os.Exit(1)
	}

	// Output success message
	fmt.Printf("Successfully initialized project '%s' in directory '%s'\n", projectName, currentDir)
	fmt.Printf("Run 'quicktodo context' to see AI usage instructions\n")
	if verbose {
		fmt.Printf("Project database: %s\n", dbPath)
		fmt.Printf("Registry: %s\n", registryPath)
	}
}

func validateProjectName(name string) error {
	if name == "" {
		return fmt.Errorf("project name cannot be empty")
	}

	if len(name) > 100 {
		return fmt.Errorf("project name cannot be longer than 100 characters")
	}

	// Check for invalid characters
	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", "\n", "\r", "\t"}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("project name cannot contain '%s'", char)
		}
	}

	return nil
}

func saveProjectDatabase(db *models.ProjectDatabase, filePath string) error {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Convert to JSON
	data, err := db.ToJSON()
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


func init() {
	RootCmd.AddCommand(initProjectCmd)
}
