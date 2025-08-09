package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"run-script-service/service"
)

func setupTestScriptFileServer(t *testing.T) (*WebServer, string) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create services - using nil for service since we don't need it
	svc := service.NewService(tmpDir+"/run.sh", tmpDir+"/run.log", tmpDir+"/config.json", 100)

	// Create service config for script manager
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
	}
	scriptManager := service.NewScriptManager(config)
	scriptFileManager := service.NewScriptFileManager(tmpDir)

	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a simple router without authentication for testing
	router := gin.New()
	server := &WebServer{
		router:            router,
		service:           svc,
		scriptManager:     scriptManager,
		scriptFileManager: scriptFileManager,
		// Don't set authMiddleware so routes will be unprotected
	}

	// Manually setup script file routes for testing (will use unprotected routes)
	server.setupScriptFileRoutes()

	return server, tmpDir
}

func TestHandleGetScriptFiles(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	// Create some test script files
	testScripts := []struct {
		filename string
		content  string
	}{
		{"backup.sh", "#!/bin/bash\necho 'backup'"},
		{"deploy.sh", "#!/bin/bash\necho 'deploy'"},
	}

	for _, script := range testScripts {
		err := server.scriptFileManager.CreateScript(script.filename, script.content)
		require.NoError(t, err)
	}

	// Test GET /api/script-files
	req := httptest.NewRequest("GET", "/api/script-files", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify script files are returned
	files, ok := response.Data.([]interface{})
	require.True(t, ok)
	assert.Len(t, files, 2)
}

func TestHandleGetScriptFile(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	// Create test script
	content := "#!/bin/bash\necho 'test script'"
	err := server.scriptFileManager.CreateScript("test.sh", content)
	require.NoError(t, err)

	// Test GET /api/script-files/:filename
	req := httptest.NewRequest("GET", "/api/script-files/test.sh", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify script content is returned
	scriptData, ok := response.Data.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "test.sh", scriptData["filename"])
	assert.Equal(t, content, scriptData["content"])
}

func TestHandleGetScriptFile_NotFound(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	// Test GET non-existent script
	req := httptest.NewRequest("GET", "/api/script-files/nonexistent.sh", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "not found")
}

func TestHandleCreateScriptFile(t *testing.T) {
	server, tmpDir := setupTestScriptFileServer(t)

	// Prepare request
	request := map[string]interface{}{
		"name":     "test-script",
		"filename": "test.sh",
		"content":  "#!/bin/bash\necho 'test'",
		"interval": 3600,
		"enabled":  true,
		"timeout":  0,
	}

	body, err := json.Marshal(request)
	require.NoError(t, err)

	// Test POST /api/script-files
	req := httptest.NewRequest("POST", "/api/script-files", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify file was created
	scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
	assert.FileExists(t, scriptPath)

	// Verify content
	fileContent, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho 'test'", string(fileContent))

	// Verify script was added to script manager
	scripts, err := server.scriptManager.GetScripts()
	require.NoError(t, err)
	assert.Len(t, scripts, 1)
	assert.Equal(t, "test-script", scripts[0].Name)
	assert.Equal(t, "test.sh", scripts[0].Filename)
}

func TestHandleCreateScriptFile_InvalidRequest(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	tests := []struct {
		name    string
		request map[string]interface{}
		wantErr string
	}{
		{
			name:    "missing name",
			request: map[string]interface{}{"filename": "test.sh", "content": "echo test", "interval": 3600},
			wantErr: "Invalid request",
		},
		{
			name:    "missing filename",
			request: map[string]interface{}{"name": "test", "content": "echo test", "interval": 3600},
			wantErr: "Invalid request",
		},
		{
			name:    "missing content",
			request: map[string]interface{}{"name": "test", "filename": "test.sh", "interval": 3600},
			wantErr: "Invalid request",
		},
		{
			name:    "invalid filename extension",
			request: map[string]interface{}{"name": "test", "filename": "test.txt", "content": "echo test", "interval": 3600},
			wantErr: ".sh extension",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.request)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/api/script-files", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			server.router.ServeHTTP(w, req)

			assert.Equal(t, http.StatusBadRequest, w.Code)

			var response APIResponse
			err = json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)
			assert.False(t, response.Success)
			assert.Contains(t, response.Error, tt.wantErr)
		})
	}
}

func TestHandleUpdateScriptFile(t *testing.T) {
	server, tmpDir := setupTestScriptFileServer(t)

	// Create initial script
	originalContent := "#!/bin/bash\necho 'original'"
	err := server.scriptFileManager.CreateScript("test.sh", originalContent)
	require.NoError(t, err)

	// Prepare update request
	updateRequest := map[string]string{
		"content": "#!/bin/bash\necho 'updated'",
	}

	body, err := json.Marshal(updateRequest)
	require.NoError(t, err)

	// Test PUT /api/script-files/:filename
	req := httptest.NewRequest("PUT", "/api/script-files/test.sh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify file content was updated
	scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
	fileContent, err := os.ReadFile(scriptPath)
	require.NoError(t, err)
	assert.Equal(t, "#!/bin/bash\necho 'updated'", string(fileContent))
}

func TestHandleUpdateScriptFile_NotFound(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	updateRequest := map[string]string{
		"content": "#!/bin/bash\necho 'test'",
	}

	body, err := json.Marshal(updateRequest)
	require.NoError(t, err)

	// Test PUT non-existent script
	req := httptest.NewRequest("PUT", "/api/script-files/nonexistent.sh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "not found")
}

func TestHandleDeleteScriptFile(t *testing.T) {
	server, tmpDir := setupTestScriptFileServer(t)

	// Create script file
	err := server.scriptFileManager.CreateScript("test.sh", "#!/bin/bash\necho 'test'")
	require.NoError(t, err)

	// Add to script manager
	scriptConfig := service.ScriptConfig{
		Name:     "test-script",
		Filename: "test.sh",
		Path:     server.scriptFileManager.GetScriptPath("test.sh"),
		Interval: 3600,
		Enabled:  true,
	}
	err = server.scriptManager.AddScript(scriptConfig)
	require.NoError(t, err)

	// Test DELETE /api/script-files/:filename
	req := httptest.NewRequest("DELETE", "/api/script-files/test.sh", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.True(t, response.Success)

	// Verify file was deleted
	scriptPath := filepath.Join(tmpDir, "scripts", "test.sh")
	assert.NoFileExists(t, scriptPath)

	// Verify script was removed from script manager
	scripts, err := server.scriptManager.GetScripts()
	require.NoError(t, err)
	assert.Len(t, scripts, 0)
}

func TestHandleDeleteScriptFile_NotFound(t *testing.T) {
	server, _ := setupTestScriptFileServer(t)

	// Test DELETE non-existent script
	req := httptest.NewRequest("DELETE", "/api/script-files/nonexistent.sh", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.False(t, response.Success)
	assert.Contains(t, response.Error, "failed to delete")
}
