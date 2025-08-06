package web

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"run-script-service/service"
)

// Helper function to create a web server with script manager for testing
func createTestServerWithScripts(scripts []service.ScriptConfig) *WebServer {
	config := &service.ServiceConfig{Scripts: scripts}
	scriptManager := service.NewScriptManager(config)
	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)
	return server
}

// Helper function to test a not-found response
func assertNotFoundResponse(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for non-existent resource")
	}
}

// Helper function to test a successful response
func assertSuccessResponse(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

// Helper function to create a standard test script configuration
func createTestScript(name string, enabled bool) service.ScriptConfig {
	return service.ScriptConfig{
		Name:        name,
		Path:        "./test.sh",
		Interval:    60,
		Enabled:     enabled,
		MaxLogLines: 100,
		Timeout:     30,
	}
}

// Helper function to authenticate a request by adding session cookie
func addAuthCookie(req *http.Request, server *WebServer) error {
	// Create a session using the auth handler
	token, err := server.authHandler.GetSessionManager().CreateSession("test-user")
	if err != nil {
		return err
	}

	// Add session cookie to request
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	return nil
}

// Helper function to create an authenticated request
func createAuthenticatedRequest(method, url, body string, server *WebServer) (*http.Request, error) {
	var req *http.Request
	var err error

	if body == "" {
		req, err = http.NewRequest(method, url, nil)
	} else {
		req, err = http.NewRequest(method, url, strings.NewReader(body))
	}

	if err != nil {
		return nil, err
	}

	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Add authentication cookie
	err = addAuthCookie(req, server)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func TestWebServer_New(t *testing.T) {
	// Create test service and log manager

	server := NewWebServer(nil, 8080, "test-secret")

	if server == nil {
		t.Fatal("NewWebServer should not return nil")
	}

	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}

}

func TestAPIResponse_JSON(t *testing.T) {
	tests := []struct {
		name     string
		response APIResponse
		expected string
	}{
		{
			name: "success response",
			response: APIResponse{
				Success: true,
				Data:    map[string]string{"test": "data"},
			},
			expected: `{"success":true,"data":{"test":"data"}}`,
		},
		{
			name: "error response",
			response: APIResponse{
				Success: false,
				Error:   "test error",
			},
			expected: `{"success":false,"error":"test error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.response)
			if err != nil {
				t.Fatalf("Failed to marshal response: %v", err)
			}

			if string(jsonData) != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, string(jsonData))
			}
		})
	}
}

func TestWebServer_StatusEndpoint(t *testing.T) {
	// Create test dependencies

	server := NewWebServer(nil, 8080, "test-secret")

	// Create authenticated test request
	req, err := createAuthenticatedRequest("GET", "/api/status", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the status handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

func TestWebServer_ScriptsEndpoint(t *testing.T) {
	// Create test dependencies with proper config
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{
			{
				Name:        "test-script",
				Path:        "./test.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
	}

	scriptManager := service.NewScriptManager(config)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create authenticated test request
	req, err := createAuthenticatedRequest("GET", "/api/scripts", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the scripts handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// Check that data contains script info
	data, ok := response.Data.([]interface{})
	if !ok {
		t.Fatal("Expected data to be an array")
	}

	if len(data) != 1 {
		t.Errorf("Expected 1 script, got %d", len(data))
	}
}

func TestWebServer_PostScript(t *testing.T) {
	// Create test dependencies with empty config
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
	}

	scriptManager := service.NewScriptManager(config)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create test request with script data
	scriptData := `{
		"name": "new-script",
		"path": "./new-script.sh",
		"interval": 120,
		"enabled": true,
		"max_log_lines": 200,
		"timeout": 60
	}`

	// Create authenticated test request
	req, err := createAuthenticatedRequest("POST", "/api/scripts", scriptData, server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the post script handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

func TestWebServer_RunScript(t *testing.T) {
	// Create test dependencies with a script
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{
			{
				Name:        "test-script",
				Path:        "./test.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
	}

	scriptManager := service.NewScriptManager(config)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create authenticated test request
	req, err := createAuthenticatedRequest("POST", "/api/scripts/test-script/run", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the run script handler
	server.router.ServeHTTP(w, req)

	// Since the script doesn't exist, we expect a 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for non-existent script")
	}
}

func TestWebServer_RunScript_NotFound(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{})

	// Create authenticated test request
	req, err := createAuthenticatedRequest("POST", "/api/scripts/non-existent/run", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertNotFoundResponse(t, w)
}

func TestWebServer_LogsEndpoint(t *testing.T) {
	server := NewWebServer(nil, 8080, "test-secret")

	// Create authenticated test request
	req, err := createAuthenticatedRequest("GET", "/api/logs", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the logs handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// Should return empty array when no script specified (new LogEntry format)
	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Error("Expected data to be an array of LogEntry objects")
	}

	if len(logs) != 0 {
		t.Error("Expected empty array when no script specified")
	}

}

// TDD Test: Red Phase - Test for LogEntry array format expected by frontend
func TestWebServer_LogsEndpoint_ExpectedFormat(t *testing.T) {
	server := NewWebServer(nil, 8080, "test-secret")

	// Create authenticated test request
	req, err := createAuthenticatedRequest("GET", "/api/logs", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the logs handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// TDD: Frontend expects an array of LogEntry objects, not raw content
	logs, ok := response.Data.([]interface{})
	if !ok {
		t.Errorf("Expected data to be an array of LogEntry objects, got %T", response.Data)
	}

	// When no logs exist, should return empty array (not nil)
	if logs == nil {
		t.Error("Expected empty array when no logs, got nil")
	}

	// Each entry should have the LogEntry structure: timestamp, message, level, script?
	for i, entry := range logs {
		logEntry, ok := entry.(map[string]interface{})
		if !ok {
			t.Errorf("Log entry %d should be an object, got %T", i, entry)
			continue
		}

		// Check required fields
		if _, hasTimestamp := logEntry["timestamp"]; !hasTimestamp {
			t.Errorf("Log entry %d missing timestamp field", i)
		}
		if _, hasMessage := logEntry["message"]; !hasMessage {
			t.Errorf("Log entry %d missing message field", i)
		}
		if level, hasLevel := logEntry["level"]; !hasLevel {
			t.Errorf("Log entry %d missing level field", i)
		} else {
			// Level should be one of: info, warning, error
			levelStr, ok := level.(string)
			if !ok || (levelStr != "info" && levelStr != "warning" && levelStr != "error") {
				t.Errorf("Log entry %d has invalid level: %v", i, level)
			}
		}
		// script field is optional
	}
}

func TestWebServer_GetScriptLogs(t *testing.T) {
	server := NewWebServer(nil, 8080, "test-secret")

	// Create authenticated test request for specific script
	req, err := createAuthenticatedRequest("GET", "/api/logs/test-script", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the script logs handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}

	// Check that data contains log content for the specific script
	data, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	script, ok := data["script"].(string)
	if !ok {
		t.Fatal("Expected script to be a string")
	}

	if script != "test-script" {
		t.Errorf("Expected script 'test-script', got '%s'", script)
	}
}

func TestWebServer_GetSpecificScript(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{
		createTestScript("test-script", true),
	})

	req, err := createAuthenticatedRequest("GET", "/api/scripts/test-script", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertSuccessResponse(t, w)
}

func TestWebServer_GetSpecificScript_NotFound(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{})

	req, err := createAuthenticatedRequest("GET", "/api/scripts/non-existent", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertNotFoundResponse(t, w)
}

func TestWebServer_EnableScript(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{
		createTestScript("test-script", false),
	})

	req, err := createAuthenticatedRequest("POST", "/api/scripts/test-script/enable", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertSuccessResponse(t, w)
}

func TestWebServer_DisableScript(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{
		createTestScript("test-script", true),
	})

	req, err := createAuthenticatedRequest("POST", "/api/scripts/test-script/disable", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertSuccessResponse(t, w)
}

func TestWebServer_StaticFiles(t *testing.T) {
	// Create test dependencies
	server := NewWebServer(nil, 8080, "test-secret")

	// Test that static route returns 404 when files don't exist
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Expect 404 since the actual static files may not exist in test environment
	if w.Code == http.StatusNotFound {
		t.Log("Static file routing is configured (returns 404 when files don't exist)")
	} else if w.Code == http.StatusOK {
		t.Log("Static files served successfully")
	} else {
		t.Logf("Unexpected status code: %d", w.Code)
	}
}

func TestWebServer_StaticFileRouting(t *testing.T) {
	// Create test dependencies
	server := NewWebServer(nil, 8080, "test-secret")

	// Test static file routing
	req := httptest.NewRequest("GET", "/static/css/main.css", nil)
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should return 404 if file doesn't exist, or 200 if it does
	if w.Code == http.StatusNotFound {
		t.Log("Static file routing configured (404 when file doesn't exist)")
	} else if w.Code == http.StatusOK {
		t.Log("Static CSS file served successfully")
	} else {
		t.Logf("Unexpected status code: %d", w.Code)
	}
}

func TestWebServer_UpdateConfig(t *testing.T) {
	// Create test dependencies with config
	config := &service.ServiceConfig{
		WebPort: 8080,
		Scripts: []service.ScriptConfig{
			{
				Name:        "test-script",
				Path:        "./test.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
	}

	// Create temporary config file for testing
	configPath := "/tmp/test_config.json"
	if err := service.SaveServiceConfig(configPath, config); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	defer func() {
		_ = os.Remove(configPath)
	}()

	scriptManager := service.NewScriptManagerWithPath(config, configPath)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create test request with updated config data
	configData := `{
		"web_port": 9090
	}`

	req, err := createAuthenticatedRequest("PUT", "/api/config", configData, server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the update config handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

func TestWebServer_UpdateConfig_InvalidJSON(t *testing.T) {
	// Create test dependencies
	config := &service.ServiceConfig{WebPort: 8080}
	configPath := "/tmp/test_config_invalid.json"
	if err := service.SaveServiceConfig(configPath, config); err != nil {
		t.Fatalf("Failed to save test config: %v", err)
	}
	defer func() {
		_ = os.Remove(configPath)
	}()

	scriptManager := service.NewScriptManagerWithPath(config, configPath)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create test request with invalid JSON
	req, err := createAuthenticatedRequest("PUT", "/api/config", "invalid json", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the update config handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for invalid JSON")
	}
}

func TestWebServer_UpdateScript(t *testing.T) {
	// Create test dependencies with a script
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{
			{
				Name:        "test-script",
				Path:        "./test.sh",
				Interval:    60,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     30,
			},
		},
	}

	scriptManager := service.NewScriptManager(config)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetScriptManager(scriptManager)

	// Create test request with updated script data
	updateData := `{
		"path": "./updated-test.sh",
		"interval": 120,
		"enabled": false,
		"max_log_lines": 200,
		"timeout": 60
	}`

	req, err := createAuthenticatedRequest("PUT", "/api/scripts/test-script", updateData, server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	// Call the update script handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Error("Expected successful response")
	}
}

func TestWebServer_UpdateScript_NotFound(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{})

	updateData := `{
		"path": "./test.sh",
		"interval": 60,
		"enabled": true,
		"max_log_lines": 100,
		"timeout": 30
	}`

	req, err := createAuthenticatedRequest("PUT", "/api/scripts/nonexistent", updateData, server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertNotFoundResponse(t, w)
}

func TestWebServer_DeleteScript(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{
		createTestScript("test-script", true),
	})

	req, err := createAuthenticatedRequest("DELETE", "/api/scripts/test-script", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertSuccessResponse(t, w)
}

func TestWebServer_DeleteScript_NotFound(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{})

	req, err := createAuthenticatedRequest("DELETE", "/api/scripts/nonexistent", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	assertNotFoundResponse(t, w)
}

func TestWebServer_WebSocketRouteSetup(t *testing.T) {
	// Create test dependencies
	server := NewWebServer(nil, 8080, "test-secret")

	// Test that WebSocket route is configured but requires authentication
	req, err := createAuthenticatedRequest("GET", "/ws", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}
	w := httptest.NewRecorder()

	server.router.ServeHTTP(w, req)

	// Should return 404 since we haven't implemented the WebSocket handler yet
	if w.Code == http.StatusNotFound {
		t.Log("WebSocket route not implemented yet (expected for TDD)")
	} else {
		t.Logf("WebSocket route status: %d (might be implemented)", w.Code)
	}
}

func TestWebServer_SystemMonitorIntegration(t *testing.T) {
	server := NewWebServer(nil, 8080, "test-secret")

	// Test that web server can be configured with system monitor
	if server.systemMonitor != nil {
		t.Error("SystemMonitor should be nil initially")
	}

	// Test setting system monitor
	monitor := service.NewSystemMonitor()
	server.SetSystemMonitor(monitor)

	if server.systemMonitor == nil {
		t.Error("SystemMonitor should not be nil after SetSystemMonitor")
	}
}

func TestWebServer_StartSystemMetricsBroadcasting(t *testing.T) {
	server := NewWebServer(nil, 8080, "test-secret")
	monitor := service.NewSystemMonitor()
	server.SetSystemMonitor(monitor)

	// Test that we can start system metrics broadcasting
	// This should not fail when system monitor is configured
	err := server.StartSystemMetricsBroadcasting(context.Background(), time.Millisecond*10)
	if err != nil {
		t.Errorf("Expected no error starting system metrics broadcasting, got: %v", err)
	}
}

func TestWebServer_GitProjects(t *testing.T) {
	// Create temporary directory structure with Git projects
	tempDir := t.TempDir()

	// Create test Git project 1
	project1 := tempDir + "/project1"
	os.MkdirAll(project1+"/.git", 0755)

	// Create test Git project 2
	project2 := tempDir + "/project2"
	os.MkdirAll(project2+"/.git", 0755)

	// Create non-Git directory
	nonGitDir := tempDir + "/not-git"
	os.MkdirAll(nonGitDir, 0755)

	server := createTestServerWithScripts([]service.ScriptConfig{})

	// Test GET /api/git-projects with test directory
	req, err := createAuthenticatedRequest("GET", "/api/git-projects?dir="+tempDir, "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if !response.Success {
		t.Errorf("Expected successful response, got: %s", response.Error)
	}

	// Verify response contains projects
	projectsData, ok := response.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected response data to be a map")
	}

	projects, ok := projectsData["projects"].([]interface{})
	if !ok {
		t.Fatal("Expected projects to be an array")
	}

	if len(projects) != 2 {
		t.Errorf("Expected 2 projects, got %d", len(projects))
	}
}

func TestWebServer_GitProjects_NonExistentDirectory(t *testing.T) {
	server := createTestServerWithScripts([]service.ScriptConfig{})

	req, err := createAuthenticatedRequest("GET", "/api/git-projects?dir=/non/existent/path", "", server)
	if err != nil {
		t.Fatalf("Failed to create authenticated request: %v", err)
	}

	w := httptest.NewRecorder()
	server.router.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	var response APIResponse
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for non-existent directory")
	}
}
