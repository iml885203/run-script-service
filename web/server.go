// Package web provides HTTP API server functionality
package web

import (
	"fmt"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"run-script-service/service"
)

// WebServer represents the HTTP API server
type WebServer struct {
	router     *gin.Engine
	service    *service.Service
	logManager *service.LogManager
	port       int
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

// setupRoutes configures all API routes
func (ws *WebServer) setupRoutes() {
	api := ws.router.Group("/api")

	// System status endpoint
	api.GET("/status", ws.handleStatus)

	// Script management endpoints
	api.GET("/scripts", ws.handleGetScripts)
	api.GET("/logs", ws.handleGetLogs)
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
	// For now, return a placeholder response
	// In a full implementation, this would use the script manager
	scripts := []map[string]interface{}{
		{
			"name":     "example",
			"path":     "./example.sh",
			"interval": 60,
			"enabled":  true,
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    scripts,
	})
}

// handleGetLogs returns logs
func (ws *WebServer) handleGetLogs(c *gin.Context) {
	// For now, return a placeholder response
	// In a full implementation, this would use the log manager
	logs := []map[string]interface{}{
		{
			"script":    "example",
			"timestamp": "2025-07-31T06:00:00Z",
			"output":    "Script executed successfully",
			"exit_code": 0,
		},
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    logs,
	})
}

// Start starts the web server
func (ws *WebServer) Start() error {
	addr := fmt.Sprintf(":%d", ws.port)
	return ws.router.Run(addr)
}
