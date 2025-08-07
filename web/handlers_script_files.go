// Package web provides script file management handlers for the HTTP API server
package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"run-script-service/service"
)

// CreateScriptFileRequest represents a request to create a new script file
type CreateScriptFileRequest struct {
	Name     string `json:"name" binding:"required"`
	Filename string `json:"filename" binding:"required"`
	Content  string `json:"content" binding:"required"`
	Interval int    `json:"interval" binding:"required"`
	Enabled  bool   `json:"enabled"`
	Timeout  int    `json:"timeout"`
}

// UpdateScriptFileRequest represents a request to update a script file
type UpdateScriptFileRequest struct {
	Content string `json:"content" binding:"required"`
}

// setupScriptFileRoutes configures script file management API routes with authentication
func (ws *WebServer) setupScriptFileRoutes() {
	if ws.scriptFileManager == nil {
		return
	}

	// Create script file routes group
	api := ws.router.Group("/api")

	// If auth middleware is available, use protected routes
	if ws.authMiddleware != nil {
		protected := api.Group("/")
		protected.Use(ws.authMiddleware.RequireAuth())

		// Script file management endpoints (protected)
		protected.GET("/script-files", ws.handleGetScriptFiles)
		protected.GET("/script-files/:filename", ws.handleGetScriptFile)
		protected.POST("/script-files", ws.handleCreateScriptFile)
		protected.PUT("/script-files/:filename", ws.handleUpdateScriptFile)
		protected.DELETE("/script-files/:filename", ws.handleDeleteScriptFile)
	} else {
		// For testing - unprotected routes
		api.GET("/script-files", ws.handleGetScriptFiles)
		api.GET("/script-files/:filename", ws.handleGetScriptFile)
		api.POST("/script-files", ws.handleCreateScriptFile)
		api.PUT("/script-files/:filename", ws.handleUpdateScriptFile)
		api.DELETE("/script-files/:filename", ws.handleDeleteScriptFile)
	}
}

// handleGetScriptFiles lists all script files
func (ws *WebServer) handleGetScriptFiles(c *gin.Context) {
	if ws.scriptFileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script file manager not initialized",
		})
		return
	}

	scripts, err := ws.scriptFileManager.ListScripts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to list scripts: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    scripts,
	})
}

// handleGetScriptFile retrieves a single script file with its content
func (ws *WebServer) handleGetScriptFile(c *gin.Context) {
	filename := c.Param("filename")

	if ws.scriptFileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script file manager not initialized",
		})
		return
	}

	script, err := ws.scriptFileManager.GetScript(filename)
	if err != nil {
		c.JSON(http.StatusNotFound, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    script,
	})
}

// handleCreateScriptFile creates a new script file and configuration
func (ws *WebServer) handleCreateScriptFile(c *gin.Context) {
	var request CreateScriptFileRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	if ws.scriptFileManager == nil || ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script managers not initialized",
		})
		return
	}

	// Create the script file
	if err := ws.scriptFileManager.CreateScript(request.Filename, request.Content); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Create script configuration
	scriptConfig := service.ScriptConfig{
		Name:        request.Name,
		Filename:    request.Filename,
		Path:        ws.scriptFileManager.GetScriptPath(request.Filename),
		Interval:    request.Interval,
		Enabled:     request.Enabled,
		MaxLogLines: 100, // Default value
		Timeout:     request.Timeout,
	}

	// Add to script manager
	if err := ws.scriptManager.AddScript(scriptConfig); err != nil {
		// If adding config failed, cleanup the created file
		ws.scriptFileManager.DeleteScript(request.Filename)

		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Failed to add script configuration: %v", err),
		})
		return
	}

	c.JSON(http.StatusCreated, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Script created successfully"},
	})
}

// handleUpdateScriptFile updates the content of an existing script file
func (ws *WebServer) handleUpdateScriptFile(c *gin.Context) {
	filename := c.Param("filename")

	var request UpdateScriptFileRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	if ws.scriptFileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script file manager not initialized",
		})
		return
	}

	if err := ws.scriptFileManager.UpdateScript(filename, request.Content); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Script updated successfully"},
	})
}

// handleDeleteScriptFile deletes a script file and removes it from configuration
func (ws *WebServer) handleDeleteScriptFile(c *gin.Context) {
	filename := c.Param("filename")

	if ws.scriptFileManager == nil || ws.scriptManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Script managers not initialized",
		})
		return
	}

	// Find and remove from script manager first (by filename)
	scripts, err := ws.scriptManager.GetScripts()
	if err == nil {
		for _, script := range scripts {
			if script.Filename == filename {
				ws.scriptManager.RemoveScript(script.Name)
				break
			}
		}
	}

	// Delete the file
	if err := ws.scriptFileManager.DeleteScript(filename); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    map[string]string{"message": "Script deleted successfully"},
	})
}
