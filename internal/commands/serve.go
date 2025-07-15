package commands

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"

	"quicktodo/internal/config"
	"quicktodo/internal/database"
	"quicktodo/internal/models"
)

//go:embed static
var staticFiles embed.FS

var (
	port       int
	openBrowser bool
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin for local development
	},
}

// Hub maintains the set of active clients and broadcasts messages to them
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

// Client represents a websocket client connection
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte
}

// WebSocket message types
type WSMessage struct {
	Type    string      `json:"type"`
	Data    interface{} `json:"data"`
	Project string      `json:"project,omitempty"`
}

// Global hub instance
var hub *Hub

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start a web server with a kanban board interface",
	Long: `Start a web server that provides a kanban board interface for managing tasks.
	
The server provides a REST API and a web interface for viewing and managing tasks
across all your projects.`,
	RunE: runServe,
}

func init() {
	serveCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port to run the server on")
	serveCmd.Flags().BoolVar(&openBrowser, "open", false, "Open browser automatically")
	RootCmd.AddCommand(serveCmd)
}

// newHub creates a new WebSocket hub
func newHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// run starts the hub
func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = true
			h.mu.Unlock()
			log.Printf("WebSocket client connected. Total clients: %d", len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
			h.mu.Unlock()
			log.Printf("WebSocket client disconnected. Total clients: %d", len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					delete(h.clients, client)
					close(client.send)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// broadcastUpdate sends an update to all connected clients
func (h *Hub) broadcastUpdate(msgType string, data interface{}, project string) {
	message := WSMessage{
		Type:    msgType,
		Data:    data,
		Project: project,
	}
	
	jsonData, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling WebSocket message: %v", err)
		return
	}
	
	select {
	case h.broadcast <- jsonData:
	default:
		log.Printf("Broadcast channel full, dropping message")
	}
}

// Client WebSocket handlers
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})
	
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}
	}
}

func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
			
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Load project registry
	registryPath := cfg.GetProjectsPath()
	registry, err := database.LoadProjectRegistry(registryPath)
	if err != nil {
		return fmt.Errorf("failed to load project registry: %w", err)
	}

	// Check if current directory is a registered project
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	currentProject, isCurrentProject := registry.GetProjectByPath(currentDir)
	if !isCurrentProject {
		// No project found in current directory - show helpful message
		fmt.Printf("🚧 No QuickTodo project found in current directory: %s\n\n", currentDir)
		fmt.Println("To use QuickTodo in this directory:")
		fmt.Printf("  quicktodo init <project-name>\n\n")
		fmt.Println("Or navigate to an existing project directory and run 'quicktodo serve' again.")
		fmt.Println("\nAvailable projects:")
		
		projects := registry.ListProjects()
		if len(projects) == 0 {
			fmt.Println("  (No projects found)")
		} else {
			for _, proj := range projects {
				fmt.Printf("  - %s (%s)\n", proj.Name, proj.Path)
			}
		}
		
		fmt.Printf("\nStarting web server anyway... You can manage all projects at http://localhost:%d\n", port)
	} else {
		fmt.Printf("📁 Detected project: %s\n", currentProject.Name)
		fmt.Printf("🌐 Starting web server at http://localhost:%d\n", port)
		
		// Update last accessed time for the current project
		if err := registry.UpdateLastAccessed(currentProject.Name); err != nil && verbose {
			fmt.Fprintf(os.Stderr, "Warning: failed to update last accessed time: %v\n", err)
		}
	}

	// Initialize WebSocket hub
	hub = newHub()
	go hub.run()

	mux := http.NewServeMux()

	// WebSocket route
	mux.HandleFunc("/ws", handleWebSocket)

	// API routes
	mux.HandleFunc("/api/projects", corsMiddleware(handleProjects(registry)))
	mux.HandleFunc("/api/projects/", corsMiddleware(handleProjectTasks(cfg, registry)))
	mux.HandleFunc("/api/current-project", corsMiddleware(handleCurrentProject(currentProject, isCurrentProject)))
	mux.HandleFunc("/api/notify", corsMiddleware(handleNotification))

	// Static files - serve from embedded files with proper path stripping
	staticSubFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return fmt.Errorf("failed to create sub filesystem: %w", err)
	}
	staticHandler := http.FileServer(http.FS(staticSubFS))
	mux.Handle("/", staticHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// Graceful shutdown
	done := make(chan bool, 1)
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
		close(done)
	}()

	fmt.Printf("Starting server on http://localhost:%d\n", port)
	fmt.Println("Press Ctrl+C to stop")

	if openBrowser {
		go func() {
			time.Sleep(1 * time.Second)
			openURL(fmt.Sprintf("http://localhost:%d", port))
		}()
	}

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	<-done
	fmt.Println("\nServer stopped")
	return nil
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func handleProjects(registry *database.ProjectRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		projects := make([]map[string]interface{}, 0)
		for name, projectInfo := range registry.ListProjects() {
			projects = append(projects, map[string]interface{}{
				"name": name,
				"path": projectInfo.Path,
				"created_at": projectInfo.CreatedAt,
				"last_accessed": projectInfo.LastAccessed,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(projects)
	}
}

func handleProjectTasks(cfg *config.Config, registry *database.ProjectRegistry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/api/projects/"), "/")
		if len(parts) < 1 || parts[0] == "" {
			http.Error(w, "Project name required", http.StatusBadRequest)
			return
		}

		projectName := parts[0]
		_, exists := registry.GetProjectByName(projectName)
		if !exists {
			http.Error(w, "Project not found", http.StatusNotFound)
			return
		}

		dbPath := cfg.GetProjectDatabasePath(projectName)
		db, err := loadProjectDatabase(dbPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to load project: %v", err), http.StatusInternalServerError)
			return
		}

		// Handle specific task operations
		if len(parts) >= 3 && parts[1] == "tasks" {
			taskID := parts[2]
			switch r.Method {
			case http.MethodGet:
				handleGetTask(w, r, db, taskID)
			case http.MethodPut:
				handleUpdateTask(w, r, db, taskID, projectName, cfg, dbPath)
			case http.MethodDelete:
				handleDeleteTask(w, r, db, taskID, projectName, cfg, dbPath)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		// Handle tasks collection
		if len(parts) == 2 && parts[1] == "tasks" {
			switch r.Method {
			case http.MethodGet:
				handleGetTasks(w, r, db)
			case http.MethodPost:
				handleCreateTask(w, r, db, projectName, cfg, dbPath)
			default:
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
			return
		}

		http.Error(w, "Invalid endpoint", http.StatusBadRequest)
	}
}

func handleGetTasks(w http.ResponseWriter, r *http.Request, db *models.ProjectDatabase) {
	tasks := db.ListTasks(nil) // Get all tasks with no filter
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

func handleGetTask(w http.ResponseWriter, r *http.Request, db *models.ProjectDatabase, taskID string) {
	id, err := strconv.Atoi(taskID)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func handleCreateTask(w http.ResponseWriter, r *http.Request, db *models.ProjectDatabase, projectName string, cfg *config.Config, dbPath string) {
	var input struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		Priority    string `json:"priority"`
		AssignedTo  string `json:"assigned_to,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate priority
	priority := models.Priority(strings.ToLower(input.Priority))
	if input.Priority != "" && !models.IsValidPriority(string(priority)) {
		priority = models.PriorityMedium // Default
	}
	if input.Priority == "" {
		priority = models.PriorityMedium
	}

	// Create task
	task := models.NewTaskWithDetails(db.NextID, input.Title, input.Description, priority)
	if input.AssignedTo != "" {
		task.AssignTo(input.AssignedTo)
	}

	if err := db.AddTask(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to add task: %v", err), http.StatusInternalServerError)
		return
	}

	if err := saveProjectDatabase(db, dbPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save task: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast task creation to WebSocket clients
	if hub != nil {
		hub.broadcastUpdate("task_created", task, projectName)
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func handleUpdateTask(w http.ResponseWriter, r *http.Request, db *models.ProjectDatabase, taskID string, projectName string, cfg *config.Config, dbPath string) {
	id, err := strconv.Atoi(taskID)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	task, err := db.GetTask(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Apply updates
	if title, ok := updates["title"].(string); ok {
		task.UpdateTitle(title)
	}
	if description, ok := updates["description"].(string); ok {
		task.UpdateDescription(description)
	}
	if status, ok := updates["status"].(string); ok {
		if models.IsValidStatus(status) {
			task.UpdateStatus(models.Status(status))
		}
	}
	if priority, ok := updates["priority"].(string); ok {
		if models.IsValidPriority(priority) {
			task.UpdatePriority(models.Priority(priority))
		}
	}
	if assignedTo, ok := updates["assigned_to"].(string); ok {
		task.AssignTo(assignedTo)
	}

	task.UpdatedAt = time.Now()

	if err := db.UpdateTask(task); err != nil {
		http.Error(w, fmt.Sprintf("Failed to update task: %v", err), http.StatusInternalServerError)
		return
	}

	if err := saveProjectDatabase(db, dbPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save project: %v", err), http.StatusInternalServerError)
		return
	}

	// Broadcast task update to WebSocket clients
	if hub != nil {
		hub.broadcastUpdate("task_updated", task, projectName)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func handleDeleteTask(w http.ResponseWriter, r *http.Request, db *models.ProjectDatabase, taskID string, projectName string, cfg *config.Config, dbPath string) {
	id, err := strconv.Atoi(taskID)
	if err != nil {
		http.Error(w, "Invalid task ID", http.StatusBadRequest)
		return
	}

	// Get task before deletion for sync purposes
	task, err := db.GetTask(id)
	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if err := db.DeleteTask(id); err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	if err := saveProjectDatabase(db, dbPath); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save project: %v", err), http.StatusInternalServerError)
		return
	}

	// Sync to TODO list if enabled
	syncToTodoList(task, projectName, "delete", cfg)

	// Broadcast task deletion to WebSocket clients
	if hub != nil {
		hub.broadcastUpdate("task_deleted", map[string]interface{}{
			"id": task.ID,
			"title": task.Title,
		}, projectName)
	}

	w.WriteHeader(http.StatusNoContent)
}

// handleWebSocket upgrades HTTP connections to WebSocket for real-time updates
func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	
	client.hub.register <- client
	
	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// handleNotification receives notifications from CLI commands and broadcasts to WebSocket clients
func handleNotification(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	var notification struct {
		Type      string      `json:"type"`
		Data      interface{} `json:"data"`
		Project   string      `json:"project"`
		Timestamp time.Time   `json:"timestamp"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Broadcast to WebSocket clients
	if hub != nil {
		hub.broadcastUpdate(notification.Type, notification.Data, notification.Project)
	}
	
	w.WriteHeader(http.StatusOK)
}

// handleCurrentProject returns information about the current project (if any)
func handleCurrentProject(currentProject *database.ProjectInfo, isCurrentProject bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		response := map[string]interface{}{
			"has_current_project": isCurrentProject,
			"current_project":     nil,
		}
		
		if isCurrentProject && currentProject != nil {
			response["current_project"] = map[string]interface{}{
				"name": currentProject.Name,
				"path": currentProject.Path,
			}
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// Note: loadProjectDatabase and saveProjectDatabase functions are defined in other command files

func openURL(url string) {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	default:
		return
	}

	exec.Command(cmd, args...).Start()
}