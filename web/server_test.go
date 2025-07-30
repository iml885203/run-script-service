package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"run-script-service/service"
)

func TestWebServer_New(t *testing.T) {
	// Create test service and log manager
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)

	if server == nil {
		t.Fatal("NewWebServer should not return nil")
	}

	if server.port != 8080 {
		t.Errorf("Expected port 8080, got %d", server.port)
	}

	if server.service != svc {
		t.Error("Service not properly assigned")
	}

	if server.logManager != logManager {
		t.Error("LogManager not properly assigned")
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
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)

	// Create test request
	req := httptest.NewRequest("GET", "/api/status", nil)
	w := httptest.NewRecorder()

	// Call the status handler
	server.router.ServeHTTP(w, req)

	// Check response
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
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)
	server.SetScriptManager(scriptManager)

	// Create test request
	req := httptest.NewRequest("GET", "/api/scripts", nil)
	w := httptest.NewRecorder()

	// Call the scripts handler
	server.router.ServeHTTP(w, req)

	// Check response
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
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)
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

	req := httptest.NewRequest("POST", "/api/scripts", strings.NewReader(scriptData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Call the post script handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", w.Code)
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
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)
	server.SetScriptManager(scriptManager)

	// Create test request
	req := httptest.NewRequest("POST", "/api/scripts/test-script/run", nil)
	w := httptest.NewRecorder()

	// Call the run script handler
	server.router.ServeHTTP(w, req)

	// Since the script doesn't exist, we expect a 404
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for non-existent script")
	}
}

func TestWebServer_RunScript_NotFound(t *testing.T) {
	// Create test dependencies with empty config
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
	}

	scriptManager := service.NewScriptManager(config)
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)
	server.SetScriptManager(scriptManager)

	// Create test request for non-existent script
	req := httptest.NewRequest("POST", "/api/scripts/non-existent/run", nil)
	w := httptest.NewRecorder()

	// Call the run script handler
	server.router.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for non-existent script")
	}
}

func TestWebServer_LogsEndpoint(t *testing.T) {
	// Create test dependencies
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)

	// Create test request
	req := httptest.NewRequest("GET", "/api/logs", nil)
	w := httptest.NewRecorder()

	// Call the logs handler
	server.router.ServeHTTP(w, req)

	// Check response
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
