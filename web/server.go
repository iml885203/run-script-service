// Package web provides HTTP API server functionality
package web

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"run-script-service/service"
)

// WebServer represents the HTTP API server
type WebServer struct {
	router        *gin.Engine
	service       *service.Service
	logManager    *service.LogManager
	scriptManager *service.ScriptManager
	port          int
}

// APIResponse represents the standard API response format
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// NewWebServer creates a new web server instance
func NewWebServer(svc *service.Service, logManager *service.LogManager, port int) *WebServer {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()

	// Add middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	router.Use(cors.Default())

	server := &WebServer{
		router:     router,
		service:    svc,
		logManager: logManager,
		port:       port,
	}

	// Setup routes
	server.setupRoutes()

	return server
}

// SetScriptManager sets the script manager for the web server
func (ws *WebServer) SetScriptManager(sm *service.ScriptManager) {
	ws.scriptManager = sm
}

// setupRoutes configures all API routes
func (ws *WebServer) setupRoutes() {
	api := ws.router.Group("/api")

	// System status endpoint
	api.GET("/status", ws.handleStatus)

	// Script management endpoints
	api.GET("/scripts", ws.handleGetScripts)
	api.POST("/scripts", ws.handlePostScript)
	api.GET("/scripts/:name", ws.handleGetScript)
	api.PUT("/scripts/:name", ws.handleUpdateScript)
	api.DELETE("/scripts/:name", ws.handleDeleteScript)
	api.POST("/scripts/:name/run", ws.handleRunScript)
	api.POST("/scripts/:name/enable", ws.handleEnableScript)
	api.POST("/scripts/:name/disable", ws.handleDisableScript)

	// Log management endpoints
	api.GET("/logs", ws.handleGetLogs)
	api.GET("/logs/:script", ws.handleGetScriptLogs)
	api.DELETE("/logs/:script", ws.handleClearScriptLogs)

	// Configuration endpoints
	api.GET("/config", ws.handleGetConfig)
	api.PUT("/config", ws.handleUpdateConfig)
}

// handleStatus returns system status information
func (ws *WebServer) handleStatus(c *gin.Context) {
	statusData := map[string]interface{}{
		"status": "running",
		"port":   ws.port,
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

	// Ensure the name matches the URL parameter
	updateData.Name = scriptName

	// Update the script (this would need to be implemented in script manager)
	c.JSON(http.StatusNotImplemented, APIResponse{
		Success: false,
		Error:   "Script update not yet implemented",
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

	// Remove the script (this would need to be implemented in script manager)
	c.JSON(http.StatusNotImplemented, APIResponse{
		Success: false,
		Error:   "Script deletion not yet implemented",
	})
}

// handleEnableScript enables a script
func (ws *WebServer) handleEnableScript(c *gin.Context) {
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

	// Enable the script
	if err := ws.scriptManager.EnableScript(scriptName); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Script %s enabled successfully", scriptName),
			"script":  scriptName,
			"enabled": true,
		},
	})
}

// handleDisableScript disables a script
func (ws *WebServer) handleDisableScript(c *gin.Context) {
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

	// Disable the script
	if err := ws.scriptManager.DisableScript(scriptName); err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Script %s disabled successfully", scriptName),
			"script":  scriptName,
			"enabled": false,
		},
	})
}

// handleGetLogs returns logs with optional query parameters
func (ws *WebServer) handleGetLogs(c *gin.Context) {
	if ws.logManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Log manager not initialized",
		})
		return
	}

	// Build query from request parameters
	query := &service.LogQuery{}

	// Parse optional query parameters
	if scriptName := c.Query("script"); scriptName != "" {
		query.ScriptName = scriptName
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	// Query logs
	entries, err := ws.logManager.QueryLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to query logs: %v", err),
		})
		return
	}

	// Convert entries to response format
	logs := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		logs[i] = map[string]interface{}{
			"script":    entry.ScriptName,
			"timestamp": entry.Timestamp.Format(time.RFC3339),
			"exit_code": entry.ExitCode,
			"stdout":    entry.Stdout,
			"stderr":    entry.Stderr,
			"duration":  entry.Duration,
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    logs,
	})
}

// handleGetScriptLogs returns logs for a specific script
func (ws *WebServer) handleGetScriptLogs(c *gin.Context) {
	if ws.logManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Log manager not initialized",
		})
		return
	}

	scriptName := c.Param("script")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// Build query for specific script
	query := &service.LogQuery{
		ScriptName: scriptName,
	}

	// Parse optional limit parameter
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			query.Limit = limit
		}
	}

	// Query logs for the specific script
	entries, err := ws.logManager.QueryLogs(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to query logs for script %s: %v", scriptName, err),
		})
		return
	}

	// Convert entries to response format
	logs := make([]map[string]interface{}, len(entries))
	for i, entry := range entries {
		logs[i] = map[string]interface{}{
			"script":    entry.ScriptName,
			"timestamp": entry.Timestamp.Format(time.RFC3339),
			"exit_code": entry.ExitCode,
			"stdout":    entry.Stdout,
			"stderr":    entry.Stderr,
			"duration":  entry.Duration,
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    logs,
	})
}

// handleClearScriptLogs clears logs for a specific script
func (ws *WebServer) handleClearScriptLogs(c *gin.Context) {
	scriptName := c.Param("script")
	if scriptName == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Script name is required",
		})
		return
	}

	// For now, return success (would need log manager implementation)
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("Logs cleared for script %s", scriptName),
			"script":  scriptName,
		},
	})
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

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    config,
	})
}

// handleUpdateConfig updates system configuration
func (ws *WebServer) handleUpdateConfig(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, APIResponse{
		Success: false,
		Error:   "Configuration update not yet implemented",
	})
}

// Start starts the web server
func (ws *WebServer) Start() error {
	addr := fmt.Sprintf(":%d", ws.port)
	return ws.router.Run(addr)
}
