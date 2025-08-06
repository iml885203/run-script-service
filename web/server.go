// Package web provides HTTP API server functionality
package web

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"run-script-service/auth"
	"run-script-service/service"
)

//go:embed frontend/dist/*
var frontendFS embed.FS

// WebServer represents the HTTP API server
type WebServer struct {
	router        *gin.Engine
	service       *service.Service
	scriptManager *service.ScriptManager
	fileManager   *service.FileManager
	wsHub         *WebSocketHub
	systemMonitor *service.SystemMonitor
	authHandler   *auth.AuthHandler
	authMiddleware *auth.AuthMiddleware
	port          int
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// LogEntry represents a structured log entry for the frontend
type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Message   string `json:"message"`
	Level     string `json:"level"` // "info", "warning", "error"
	Script    string `json:"script,omitempty"`
}

// NewWebServer creates a new web server instance
func NewWebServer(svc *service.Service, port int, secretKey string) *WebServer {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	// Create WebSocket hub
	wsHub := NewWebSocketHub()
	go wsHub.Run()

	// Initialize authentication
	authHandler := auth.NewAuthHandler(secretKey)
	authMiddleware := auth.NewAuthMiddleware(authHandler.GetSessionManager())

	server := &WebServer{
		router:         router,
		service:        svc,
		wsHub:          wsHub,
		authHandler:    authHandler,
		authMiddleware: authMiddleware,
		port:           port,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// SetScriptManager sets the script manager for the web server
func (ws *WebServer) SetScriptManager(sm *service.ScriptManager) {
	ws.scriptManager = sm
}

// GetWebSocketHub returns the WebSocket hub for broadcasting messages
func (ws *WebServer) GetWebSocketHub() *WebSocketHub {
	return ws.wsHub
}

// SetSystemMonitor sets the system monitor for the web server
func (ws *WebServer) SetSystemMonitor(monitor *service.SystemMonitor) {
	ws.systemMonitor = monitor
}

// StartSystemMetricsBroadcasting starts periodic system metrics broadcasting via WebSocket
func (ws *WebServer) StartSystemMetricsBroadcasting(ctx context.Context, interval time.Duration) error {
	if ws.systemMonitor == nil {
		return fmt.Errorf("system monitor not configured")
	}

	// Create event publisher that broadcasts via WebSocket
	publisher := func(msgType string, data map[string]interface{}) error {
		if ws.wsHub != nil {
			return ws.wsHub.BroadcastMessage(msgType, data)
		}
		return nil
	}

	// Start periodic broadcasting in a goroutine
	go ws.systemMonitor.StartPeriodicBroadcasting(ctx, interval, publisher)

	return nil
}

// setupRoutes configures all API routes
func (ws *WebServer) setupRoutes() {
	// Create a sub filesystem for the dist directory
	distFS, err := fs.Sub(frontendFS, "frontend/dist")
	if err != nil {
		fmt.Printf("DEBUG: embed fs.Sub failed: %v, using fallback\n", err)
		// Fallback to file system if embed fails (development mode)
		ws.router.Static("/static", "./web/frontend/dist")
		ws.router.GET("/", func(c *gin.Context) {
			c.File("./web/frontend/dist/index.html")
		})
	} else {
		fmt.Println("DEBUG: Using embedded filesystem")
		// Use embedded filesystem for static files (protected)
		staticGroup := ws.router.Group("/static", ws.authMiddleware.RequireAuth())
		staticGroup.StaticFS("/", http.FS(distFS))

		// Login route (unprotected)
		ws.router.GET("/login", func(c *gin.Context) {
			indexFile, err := distFS.Open("index.html")
			if err != nil {
				fmt.Printf("DEBUG: Failed to open embedded index.html: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
				return
			}
			defer indexFile.Close()

			c.Header("Content-Type", "text/html")
			http.ServeContent(c.Writer, c.Request, "index.html", time.Now(), indexFile.(io.ReadSeeker))
		})

		// Root route serves index.html from embedded FS (protected)
		ws.router.GET("/", ws.authMiddleware.RequireAuth(), func(c *gin.Context) {
			indexFile, err := distFS.Open("index.html")
			if err != nil {
				fmt.Printf("DEBUG: Failed to open embedded index.html: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
				return
			}
			defer indexFile.Close()

			c.Header("Content-Type", "text/html")
			http.ServeContent(c.Writer, c.Request, "index.html", time.Now(), indexFile.(io.ReadSeeker))
		})

		// Serve index.html for SPA routes (NoRoute handler) - protected  
		ws.router.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// If it's an API route, let it 404
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/ws") {
				c.JSON(http.StatusNotFound, gin.H{"error": "Not found"})
				return
			}

			// Check authentication for SPA routes
			if !ws.authMiddleware.IsAuthenticated(c) {
				c.Redirect(http.StatusFound, "/login")
				return
			}

			// For all other routes, serve index.html (Vue.js SPA)
			indexFile, err := distFS.Open("index.html")
			if err != nil {
				fmt.Printf("DEBUG: Failed to open embedded index.html in NoRoute: %v\n", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load frontend"})
				return
			}
			defer indexFile.Close()

			c.Header("Content-Type", "text/html")
			http.ServeContent(c.Writer, c.Request, "index.html", time.Now(), indexFile.(io.ReadSeeker))
		})
	}

	// WebSocket endpoint (protected)
	ws.router.GET("/ws", ws.authMiddleware.RequireAuth(), func(c *gin.Context) {
		HandleWebSocket(ws.wsHub, c)
	})

	api := ws.router.Group("/api")

	// Authentication routes (unprotected)
	auth := api.Group("/auth")
	auth.POST("/login", ws.authHandler.Login)
	auth.POST("/logout", ws.authHandler.Logout)
	auth.GET("/status", ws.authHandler.AuthStatus)

	// Protected routes (require authentication)
	protected := api.Group("/")
	protected.Use(ws.authMiddleware.RequireAuth())

	// System status endpoint
	protected.GET("/status", ws.handleStatus)

	// Script management endpoints
	protected.GET("/scripts", ws.handleGetScripts)
	protected.POST("/scripts", ws.handlePostScript)
	protected.GET("/scripts/:name", ws.handleGetScript)
	protected.PUT("/scripts/:name", ws.handleUpdateScript)
	protected.DELETE("/scripts/:name", ws.handleDeleteScript)
	protected.POST("/scripts/:name/run", ws.handleRunScript)
	protected.POST("/scripts/:name/enable", ws.handleEnableScript)
	protected.POST("/scripts/:name/disable", ws.handleDisableScript)

	// Log management endpoints
	protected.GET("/logs", ws.handleGetLogs)
	protected.GET("/logs/:script", ws.handleGetScriptLogs)
	protected.GET("/logs/raw/:script", ws.handleGetRawLogs) // New simple endpoint
	protected.DELETE("/logs/:script", ws.handleClearScriptLogs)

	// Configuration endpoints
	protected.GET("/config", ws.handleGetConfig)
	protected.PUT("/config", ws.handleUpdateConfig)
}

// handleStatus returns system status information
func (ws *WebServer) handleStatus(c *gin.Context) {
	uptime := "Unknown"
	runningScripts := 0
	totalScripts := 0

	// Get script counts if script manager is available
	if ws.scriptManager != nil {
		config := ws.scriptManager.GetConfig()
		totalScripts = len(config.Scripts)

		// Count running/enabled scripts
		for _, script := range config.Scripts {
			if script.Enabled && ws.scriptManager.IsScriptRunning(script.Name) {
				runningScripts++
			}
		}
	}

	// Calculate uptime if system monitor is available
	if ws.systemMonitor != nil {
		uptimeStr := ws.systemMonitor.GetUptime()
		if uptimeStr != "" {
			uptime = uptimeStr
		}
	}

	statusData := map[string]interface{}{
		"status":         "running",
		"uptime":         uptime,
		"runningScripts": runningScripts,
		"totalScripts":   totalScripts,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    statusData,
	})
}

// handleGetScripts returns all scripts
func (ws *WebServer) handleGetScripts(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	// Get script configs from the manager
	var scripts []map[string]interface{}
	for _, scriptConfig := range ws.scriptManager.GetConfig().Scripts {
		running := ws.scriptManager.IsScriptRunning(scriptConfig.Name)

		scripts = append(scripts, map[string]interface{}{
			"name":          scriptConfig.Name,
			"path":          scriptConfig.Path,
			"interval":      scriptConfig.Interval,
			"enabled":       scriptConfig.Enabled,
			"max_log_lines": scriptConfig.MaxLogLines,
			"timeout":       scriptConfig.Timeout,
			"running":       running,
		})
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    scripts,
	})
}

// handlePostScript creates a new script
func (ws *WebServer) handlePostScript(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	var scriptConfig service.ScriptConfig
	if err := c.ShouldBindJSON(&scriptConfig); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Validate required fields
	if scriptConfig.Name == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	if scriptConfig.Path == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script path is required",
		})
		return
	}

	// Set defaults for optional fields
	if scriptConfig.Interval <= 0 {
		scriptConfig.Interval = 60 // Default to 1 minute
	}
	if scriptConfig.MaxLogLines <= 0 {
		scriptConfig.MaxLogLines = 100 // Default to 100 lines
	}

	// Add the script
	if err := ws.scriptManager.AddScript(scriptConfig); err != nil {
		c.JSON(http.StatusConflict, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"name":          scriptConfig.Name,
			"path":          scriptConfig.Path,
			"interval":      scriptConfig.Interval,
			"enabled":       scriptConfig.Enabled,
			"max_log_lines": scriptConfig.MaxLogLines,
			"timeout":       scriptConfig.Timeout,
		},
	})
}

// handleRunScript executes a script once
func (ws *WebServer) handleRunScript(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	scriptName := c.Param("name")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Run the script with a timeout context
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	if err := ws.scriptManager.RunScriptOnce(ctx, scriptName); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Note: Logs are now handled via raw file access in /logs/raw/:script endpoint

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Script %s executed successfully", scriptName),
			"script":  scriptName,
		},
	})
}

// handleGetScript returns information about a specific script
func (ws *WebServer) handleGetScript(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	scriptName := c.Param("name")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Find the script in configuration
	config := ws.scriptManager.GetConfig()
	for _, scriptConfig := range config.Scripts {
		if scriptConfig.Name == scriptName {
			running := ws.scriptManager.IsScriptRunning(scriptConfig.Name)

			scriptData := map[string]interface{}{
				"name":          scriptConfig.Name,
				"path":          scriptConfig.Path,
				"interval":      scriptConfig.Interval,
				"enabled":       scriptConfig.Enabled,
				"max_log_lines": scriptConfig.MaxLogLines,
				"timeout":       scriptConfig.Timeout,
				"running":       running,
			}

			c.JSON(http.StatusOK, APIResponse{
				Success: true,
				Data:    scriptData,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, APIResponse{
		Success: false,
		Error:   fmt.Sprintf("Script '%s' not found", scriptName),
	})
}

// handleUpdateScript updates a script configuration
func (ws *WebServer) handleUpdateScript(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	scriptName := c.Param("name")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	var updateData service.ScriptConfig
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Set defaults for optional fields
	if updateData.Interval <= 0 {
		updateData.Interval = 60 // Default to 1 minute
	}
	if updateData.MaxLogLines <= 0 {
		updateData.MaxLogLines = 100 // Default to 100 lines
	}

	// Update the script
	if err := ws.scriptManager.UpdateScript(scriptName, updateData); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message":       fmt.Sprintf("Script %s updated successfully", scriptName),
			"script":        scriptName,
			"name":          updateData.Name,
			"path":          updateData.Path,
			"interval":      updateData.Interval,
			"enabled":       updateData.Enabled,
			"max_log_lines": updateData.MaxLogLines,
			"timeout":       updateData.Timeout,
		},
	})
}

// handleDeleteScript removes a script
func (ws *WebServer) handleDeleteScript(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	scriptName := c.Param("name")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Remove the script
	if err := ws.scriptManager.RemoveScript(scriptName); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Script %s deleted successfully", scriptName),
			"script":  scriptName,
		},
	})
}

// handleScriptToggle handles both enable and disable script operations
func (ws *WebServer) handleScriptToggle(c *gin.Context, enable bool) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	scriptName := c.Param("name")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	var err error
	var action string
	if enable {
		err = ws.scriptManager.EnableScript(scriptName)
		action = "enabled"
	} else {
		err = ws.scriptManager.DisableScript(scriptName)
		action = "disabled"
	}

	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Script %s %s successfully", scriptName, action),
			"script":  scriptName,
			"enabled": enable,
		},
	})
}

// handleEnableScript enables a script
func (ws *WebServer) handleEnableScript(c *gin.Context) {
	ws.handleScriptToggle(c, true)
}

// handleDisableScript disables a script
func (ws *WebServer) handleDisableScript(c *gin.Context) {
	ws.handleScriptToggle(c, false)
}

// handleGetLogs returns structured log entries as expected by frontend
func (ws *WebServer) handleGetLogs(c *gin.Context) {
	scriptName := c.Query("script")
	limit := c.DefaultQuery("limit", "50")

	// Parse limit
	maxEntries := 50
	if parsedLimit, err := strconv.Atoi(limit); err == nil && parsedLimit > 0 {
		maxEntries = parsedLimit
	}

	// If no script specified, return aggregated logs from all scripts
	if scriptName == "" {
		allLogs := ws.getAggregatedLogs(maxEntries)
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data:    allLogs,
		})
		return
	}

	// Get logs for specific script
	scriptLogs := ws.getScriptLogs(scriptName, maxEntries)
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    scriptLogs,
	})
}

// handleGetScriptLogs returns raw log content for a specific script
func (ws *WebServer) handleGetScriptLogs(c *gin.Context) {
	scriptName := c.Param("script")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Get log file path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	logFile := filepath.Join(dir, fmt.Sprintf("%s.log", scriptName))

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"content": "",
				"script":  scriptName,
				"message": "No log file found",
			},
		})
		return
	}

	// Read raw log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read log file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"content": string(content),
			"script":  scriptName,
		},
	})
}

// handleClearScriptLogs clears logs for a specific script (simplified)
func (ws *WebServer) handleClearScriptLogs(c *gin.Context) {
	scriptName := c.Param("script")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Get log file path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	logFile := filepath.Join(dir, fmt.Sprintf("%s.log", scriptName))

	// Clear the log file by truncating it
	if err := os.Truncate(logFile, 0); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to clear log file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Logs cleared for script %s", scriptName),
			"script":  scriptName,
		},
	})
}

// handleGetRawLogs returns raw log file content (simple approach)
func (ws *WebServer) handleGetRawLogs(c *gin.Context) {
	scriptName := c.Param("script")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Get log file path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	logFile := filepath.Join(dir, fmt.Sprintf("%s.log", scriptName))

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		c.JSON(http.StatusOK, APIResponse{
			Success: true,
			Data: map[string]interface{}{
				"script":  scriptName,
				"content": "",
				"message": "No log file found",
			},
		})
		return
	}

	// Read log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to read log file: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"script":  scriptName,
			"content": string(content),
		},
	})
}

// ConfigResponse represents the configuration format expected by the frontend
type ConfigResponse struct {
	WebPort      int    `json:"webPort"`
	Interval     string `json:"interval"`
	LogRetention int    `json:"logRetention"`
	AutoRefresh  bool   `json:"autoRefresh"`
}

// handleGetConfig returns system configuration
func (ws *WebServer) handleGetConfig(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	config := ws.scriptManager.GetConfig()

	// Convert to frontend-expected format
	response := ConfigResponse{
		WebPort:      config.WebPort,
		Interval:     "1h", // default interval as string
		LogRetention: 100,  // default log retention
		AutoRefresh:  true, // default auto-refresh setting
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// handleUpdateConfig updates system configuration
func (ws *WebServer) handleUpdateConfig(c *gin.Context) {
	if ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script manager not initialized",
		})
		return
	}

	var updateData map[string]interface{}
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Get current configuration
	config := ws.scriptManager.GetConfig()

	// Update web port if provided (handle both camelCase and snake_case)
	if webPort, ok := updateData["webPort"]; ok {
		if port, isFloat := webPort.(float64); isFloat {
			if port >= 1 && port <= 65535 {
				config.WebPort = int(port)
			} else {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error:   "Web port must be between 1 and 65535",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Web port must be a number",
			})
			return
		}
	} else if webPort, ok := updateData["web_port"]; ok {
		if port, isFloat := webPort.(float64); isFloat {
			if port >= 1 && port <= 65535 {
				config.WebPort = int(port)
			} else {
				c.JSON(http.StatusBadRequest, APIResponse{
					Success: false,
					Error:   "Web port must be between 1 and 65535",
				})
				return
			}
		} else {
			c.JSON(http.StatusBadRequest, APIResponse{
				Success: false,
				Error:   "Web port must be a number",
			})
			return
		}
	}

	// Save updated configuration
	if err := ws.scriptManager.SaveConfig(); err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to save configuration: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": "Configuration updated successfully",
			"config":  config,
		},
	})
}

// Start starts the web server
func (ws *WebServer) Start() error {
	addr := fmt.Sprintf(":%d", ws.port)
	return ws.router.Run(addr)
}

// getAggregatedLogs returns logs from all scripts in LogEntry format
func (ws *WebServer) getAggregatedLogs(maxEntries int) []LogEntry {
	// Initialize with non-nil slice to ensure JSON serializes as [] not null
	allLogs := make([]LogEntry, 0)

	if ws.scriptManager == nil {
		return allLogs
	}

	// Get all configured scripts
	config := ws.scriptManager.GetConfig()
	for _, script := range config.Scripts {
		scriptLogs := ws.getScriptLogs(script.Name, maxEntries)
		allLogs = append(allLogs, scriptLogs...)
	}

	// If we have more logs than requested, truncate to most recent
	if len(allLogs) > maxEntries {
		allLogs = allLogs[len(allLogs)-maxEntries:]
	}

	return allLogs
}

// getScriptLogs returns logs for a specific script in LogEntry format
func (ws *WebServer) getScriptLogs(scriptName string, maxEntries int) []LogEntry {
	// Initialize with non-nil slice to ensure JSON serializes as [] not null
	logs := make([]LogEntry, 0)

	// Get log file path
	dir, err := os.Executable()
	if err != nil {
		dir, _ = os.Getwd()
	} else {
		dir = filepath.Dir(dir)
	}

	logFile := filepath.Join(dir, fmt.Sprintf("%s.log", scriptName))

	// Check if log file exists
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		return logs // Return empty array
	}

	// Read log file content
	content, err := os.ReadFile(logFile)
	if err != nil {
		return logs // Return empty array on error
	}

	// Parse log content into LogEntry objects
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Simple log parsing - assume format: timestamp level message
		// For now, create basic LogEntry objects from raw log lines
		logEntry := LogEntry{
			Timestamp: time.Now().Format(time.RFC3339), // Default timestamp
			Message:   line,
			Level:     "info", // Default level
			Script:    scriptName,
		}

		// Try to extract timestamp and level from line if possible
		if len(line) > 19 && line[10] == 'T' { // ISO timestamp format
			if timestampEnd := strings.Index(line[20:], " "); timestampEnd > 0 {
				logEntry.Timestamp = line[:20+timestampEnd]
				remaining := strings.TrimSpace(line[20+timestampEnd:])

				// Check for level indicators
				if strings.Contains(strings.ToLower(remaining), "error") {
					logEntry.Level = "error"
				} else if strings.Contains(strings.ToLower(remaining), "warn") {
					logEntry.Level = "warning"
				}

				logEntry.Message = remaining
			}
		}

		logs = append(logs, logEntry)
	}

	// Limit to maxEntries (most recent)
	if len(logs) > maxEntries {
		logs = logs[len(logs)-maxEntries:]
	}

	return logs
}
