// Package web provides file operation handlers for the HTTP API server
package web

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"run-script-service/service"
)

// FileOperationRequest represents a file operation request
type FileOperationRequest struct {
	Path    string `json:"path,omitempty"`
	Content string `json:"content,omitempty"`
}

// ValidationRequest represents a script validation request
type ValidationRequest struct {
	Content string `json:"content" binding:"required"`
}

// ValidationResponse represents validation results
type ValidationResponse struct {
	Valid  bool     `json:"valid"`
	Issues []string `json:"issues,omitempty"`
}

// SetFileManager sets the file manager for the web server
func (ws *WebServer) SetFileManager(fm *service.FileManager) {
	ws.fileManager = fm

	// Setup file routes now that file manager is available
	api := ws.router.Group("/api")
	ws.setupFileRoutes(api)
}

// setupFileRoutes configures file operation API routes
func (ws *WebServer) setupFileRoutes(api *gin.RouterGroup) {
	if ws.fileManager == nil {
		return
	}

	// File operation endpoints
	api.GET("/files/*path", ws.handleGetFile)
	api.PUT("/files/*path", ws.handlePutFile)
	api.POST("/files/validate", ws.handleValidateFile)
	api.GET("/files-list/*path", ws.handleListFiles)
}

// handleGetFile reads and returns a file's content
func (ws *WebServer) handleGetFile(c *gin.Context) {
	if ws.fileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "File manager not initialized",
		})
		return
	}

	// Extract path from URL parameter
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "File path is required",
		})
		return
	}

	// Remove leading slash from path parameter
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}

	fileContent, err := ws.fileManager.ReadFile(filePath)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "access denied: path not allowed" {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "no such file or directory") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    fileContent,
	})
}

// handlePutFile writes content to a file
func (ws *WebServer) handlePutFile(c *gin.Context) {
	if ws.fileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "File manager not initialized",
		})
		return
	}

	// Extract path from URL parameter
	filePath := c.Param("path")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "File path is required",
		})
		return
	}

	// Remove leading slash from path parameter
	if filePath[0] == '/' {
		filePath = filePath[1:]
	}

	var request FileOperationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Use content from request body
	content := request.Content

	err := ws.fileManager.WriteFile(filePath, content)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "access denied: path not allowed" {
			statusCode = http.StatusForbidden
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"message": fmt.Sprintf("File '%s' written successfully", filePath),
			"path":    filePath,
		},
	})
}

// handleValidateFile validates script syntax
func (ws *WebServer) handleValidateFile(c *gin.Context) {
	if ws.fileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "File manager not initialized",
		})
		return
	}

	var request ValidationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	issues := ws.fileManager.ValidateScriptSyntax(request.Content)

	response := ValidationResponse{
		Valid:  len(issues) == 0,
		Issues: issues,
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    response,
	})
}

// handleListFiles lists files in a directory
func (ws *WebServer) handleListFiles(c *gin.Context) {
	if ws.fileManager == nil {
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "File manager not initialized",
		})
		return
	}

	// Extract path from URL parameter
	dirPath := c.Param("path")
	if dirPath == "" {
		dirPath = "." // Default to current directory
	}

	// Remove leading slash from path parameter
	if dirPath[0] == '/' {
		dirPath = dirPath[1:]
	}

	files, err := ws.fileManager.ListFiles(dirPath)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "access denied: path not allowed" {
			statusCode = http.StatusForbidden
		} else if strings.Contains(err.Error(), "no such file or directory") {
			statusCode = http.StatusNotFound
		}

		c.JSON(statusCode, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	// Convert file info to JSON-serializable format
	fileList := make([]map[string]interface{}, len(files))
	for i, file := range files {
		fileList[i] = map[string]interface{}{
			"name":     file.Name(),
			"size":     file.Size(),
			"mode":     file.Mode().String(),
			"is_dir":   file.IsDir(),
			"mod_time": file.ModTime().Unix(),
		}
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    fileList,
	})
}
