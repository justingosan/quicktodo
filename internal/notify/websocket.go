package notify

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"quicktodo/internal/config"
	"quicktodo/internal/models"
	"strings"
	"time"
)

// NotificationMessage represents a change notification
type NotificationMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Project   string      `json:"project"`
	Timestamp time.Time   `json:"timestamp"`
}

// NotifyWebServer sends a notification to any running web server instances
func NotifyWebServer(cfg *config.Config, msgType string, data interface{}, projectName string) error {
	// Try to send via HTTP to running server first
	if err := sendHTTPNotification(msgType, data, projectName); err == nil {
		return nil
	}
	
	// Fallback to file-based notification
	return writeNotificationFile(cfg, msgType, data, projectName)
}

// sendHTTPNotification tries to send notification to running web server
func sendHTTPNotification(msgType string, data interface{}, projectName string) error {
	// Try common development ports
	ports := []int{8080, 3000, 8000, 8086, 9000, 8001, 8008}
	
	notification := NotificationMessage{
		Type:      msgType,
		Data:      data,
		Project:   projectName,
		Timestamp: time.Now(),
	}
	
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return err
	}
	
	for _, port := range ports {
		url := fmt.Sprintf("http://localhost:%d/api/notify", port)
		resp, err := http.Post(url, "application/json", strings.NewReader(string(jsonData)))
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return nil
		}
		if resp != nil {
			resp.Body.Close()
		}
	}
	
	return fmt.Errorf("no running web server found")
}

// writeNotificationFile writes notification to a file that the web server can monitor
func writeNotificationFile(cfg *config.Config, msgType string, data interface{}, projectName string) error {
	notificationDir := filepath.Join(cfg.DataDir, "notifications")
	if err := os.MkdirAll(notificationDir, 0755); err != nil {
		return fmt.Errorf("failed to create notification directory: %w", err)
	}
	
	notification := NotificationMessage{
		Type:      msgType,
		Data:      data,
		Project:   projectName,
		Timestamp: time.Now(),
	}
	
	jsonData, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}
	
	// Use timestamp to create unique filename
	filename := fmt.Sprintf("%d_%s_%s.json", time.Now().UnixNano(), projectName, msgType)
	filePath := filepath.Join(notificationDir, filename)
	
	return os.WriteFile(filePath, jsonData, 0644)
}

// NotifyTaskCreated sends a task creation notification
func NotifyTaskCreated(cfg *config.Config, task *models.Task, projectName string) error {
	return NotifyWebServer(cfg, "task_created", task, projectName)
}

// NotifyTaskUpdated sends a task update notification
func NotifyTaskUpdated(cfg *config.Config, task *models.Task, projectName string) error {
	return NotifyWebServer(cfg, "task_updated", task, projectName)
}

// NotifyTaskDeleted sends a task deletion notification
func NotifyTaskDeleted(cfg *config.Config, taskID int, title string, projectName string) error {
	data := map[string]interface{}{
		"id":    taskID,
		"title": title,
	}
	return NotifyWebServer(cfg, "task_deleted", data, projectName)
}