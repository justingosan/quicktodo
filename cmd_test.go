package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestCLIIntegration tests the main CLI commands end-to-end
func TestCLIIntegration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "quicktodo-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Change to temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(originalDir)
	
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}
	
	// Build the binary
	binaryPath := filepath.Join(tempDir, "quicktodo")
	buildCmd := exec.Command("go", "build", "-o", binaryPath)
	buildCmd.Dir = originalDir
	buildOutput, err := buildCmd.CombinedOutput()
	if err != nil {
		t.Fatalf("Failed to build binary: %v, output: %s", err, buildOutput)
	}
	
	// Test init command
	t.Run("init", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "init", "test-project")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Init command failed: %v, output: %s", err, output)
		}
		
		if !strings.Contains(string(output), "Successfully initialized") {
			t.Errorf("Expected success message in output: %s", output)
		}
	})
	
	// Test create-task command
	t.Run("create-task", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "create-task", "Test Task", "--description", "Test description", "--priority", "high", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Create task command failed: %v, output: %s", err, output)
		}
		
		// Parse JSON response
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		success, ok := response["success"].(bool)
		if !ok || !success {
			t.Errorf("Expected success=true in response: %s", output)
		}
		
		task, ok := response["task"].(map[string]interface{})
		if !ok {
			t.Errorf("Expected task object in response: %s", output)
		}
		
		// Task created successfully
		
		if task["title"].(string) != "Test Task" {
			t.Errorf("Expected task title 'Test Task', got '%s'", task["title"])
		}
		
		if task["priority"].(string) != "high" {
			t.Errorf("Expected task priority 'high', got '%s'", task["priority"])
		}
	})
	
	// Test list-tasks command
	t.Run("list-tasks", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "list-tasks", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("List tasks command failed: %v, output: %s", err, output)
		}
		
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		taskCount := int(response["task_count"].(float64))
		if taskCount != 1 {
			t.Errorf("Expected 1 task, got %d", taskCount)
		}
	})
	
	// Test display-task command
	t.Run("display-task", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "display-task", "1", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Display task command failed: %v, output: %s", err, output)
		}
		
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		task := response["task"].(map[string]interface{})
		if task["title"].(string) != "Test Task" {
			t.Errorf("Expected task title 'Test Task', got '%s'", task["title"])
		}
	})
	
	// Test set-task-status command
	t.Run("set-task-status", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "set-task-status", "1", "in_progress", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Set task status command failed: %v, output: %s", err, output)
		}
		
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		if response["new_status"].(string) != "in_progress" {
			t.Errorf("Expected new status 'in_progress', got '%s'", response["new_status"])
		}
	})
	
	// Test mark-completed command
	t.Run("mark-completed", func(t *testing.T) {
		cmd := exec.Command(binaryPath, "mark-completed", "1", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Mark completed command failed: %v, output: %s", err, output)
		}
		
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		if response["new_status"].(string) != "done" {
			t.Errorf("Expected new status 'done', got '%s'", response["new_status"])
		}
	})
	
	// Test filtering
	t.Run("list-tasks-with-filter", func(t *testing.T) {
		// Create another task first
		cmd := exec.Command(binaryPath, "create-task", "Another Task", "--priority", "low")
		_, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("Failed to create second task: %v", err)
		}
		
		// Filter by status
		cmd = exec.Command(binaryPath, "list-tasks", "--status", "done", "--json")
		output, err := cmd.CombinedOutput()
		if err != nil {
			t.Errorf("List tasks with filter failed: %v, output: %s", err, output)
		}
		
		var response map[string]interface{}
		err = json.Unmarshal(output, &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response: %v", err)
		}
		
		taskCount := int(response["task_count"].(float64))
		if taskCount != 1 {
			t.Errorf("Expected 1 done task, got %d", taskCount)
		}
	})
}

// TestCLIHelp tests that help commands work
func TestCLIHelp(t *testing.T) {
	// Build the binary
	binaryPath := "./quicktodo"
	
	tests := []struct {
		name string
		args []string
	}{
		{"root help", []string{"--help"}},
		{"context help", []string{"context", "--help"}},
		{"create help", []string{"create-task", "--help"}},
		{"list help", []string{"list-tasks", "--help"}},
		{"serve help", []string{"serve", "--help"}},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			output, err := cmd.CombinedOutput()
			
			// Help commands should exit with code 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() != 0 {
					t.Errorf("Help command should exit with code 0, got: %v", err)
				}
			}
			
			if len(output) == 0 {
				t.Errorf("Expected help output, got empty")
			}
			
			if !strings.Contains(string(output), "Usage:") {
				t.Errorf("Expected 'Usage:' in help output: %s", output)
			}
		})
	}
}

// TestCLIVersion tests the version command
func TestCLIVersion(t *testing.T) {
	binaryPath := "./quicktodo"
	
	cmd := exec.Command(binaryPath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Errorf("Version command failed: %v", err)
	}
	
	if len(output) == 0 {
		t.Error("Expected version output, got empty")
	}
	
	// Should contain version information
	outputStr := string(output)
	if !strings.Contains(strings.ToLower(outputStr), "quicktodo") {
		t.Errorf("Expected 'quicktodo' in version output: %s", outputStr)
	}
}

// TestCLIErrorHandling tests error conditions
func TestCLIErrorHandling(t *testing.T) {
	binaryPath := "./quicktodo"
	
	tests := []struct {
		name        string
		args        []string
		expectError bool
	}{
		{"invalid command", []string{"invalid-command"}, true},
		{"create without title", []string{"create-task"}, true},
		{"display non-existent task", []string{"display-task", "999"}, true},
		{"invalid status", []string{"set-task-status", "1", "invalid"}, true},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := exec.Command(binaryPath, tt.args...)
			_, err := cmd.CombinedOutput()
			
			if tt.expectError && err == nil {
				t.Errorf("Expected error for command %v, but got none", tt.args)
			}
			
			if !tt.expectError && err != nil {
				t.Errorf("Expected no error for command %v, but got: %v", tt.args, err)
			}
		})
	}
}