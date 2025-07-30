package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	// Create test dependencies
	svc := &service.Service{}
	logManager := &service.LogManager{}

	server := NewWebServer(svc, logManager, 8080)

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
