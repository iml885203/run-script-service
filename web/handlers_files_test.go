package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"run-script-service/service"
)

// Helper function to test access denied responses
func assertAccessDeniedResponse(t *testing.T, w *httptest.ResponseRecorder) {
	t.Helper()
	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403, got %d", w.Code)
	}

	var response APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response.Success {
		t.Error("Expected failed response for denied access")
	}
}

func TestWebServer_FileOperations(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "web_files_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test dependencies
	fileManager := service.NewFileManager(tempDir)

	server := NewWebServer(nil, 8080, "test-secret")
	server.SetFileManager(fileManager)

	// Create test file
	testContent := "#!/bin/bash\necho 'Hello World'\n"
	testFile := filepath.Join(tempDir, "test.sh")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("get file - success", func(t *testing.T) {
		req, err := createAuthenticatedRequest("GET", "/api/files/test.sh", "", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

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

		// Check that data contains file content
		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		content, ok := data["content"].(string)
		if !ok {
			t.Fatal("Expected content to be a string")
		}

		if content != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, content)
		}
	})

	t.Run("get file - access denied", func(t *testing.T) {
		req, err := createAuthenticatedRequest("GET", "/api/files/../../etc/passwd", "", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assertAccessDeniedResponse(t, w)
	})

	t.Run("get file - not found", func(t *testing.T) {
		req, err := createAuthenticatedRequest("GET", "/api/files/nonexistent.sh", "", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})

	t.Run("put file - success", func(t *testing.T) {
		newContent := "#!/bin/bash\necho 'Updated content'\n"
		requestBody := `{"content": "#!/bin/bash\necho 'Updated content'\n"}`

		req, err := createAuthenticatedRequest("PUT", "/api/files/new_test.sh", requestBody, server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

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

		// Verify file was created
		createdFile := filepath.Join(tempDir, "new_test.sh")
		content, err := os.ReadFile(createdFile)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		if string(content) != newContent {
			t.Errorf("Expected content '%s', got '%s'", newContent, string(content))
		}
	})

	t.Run("put file - access denied", func(t *testing.T) {
		requestBody := `{"content": "malicious content"}`

		req, err := createAuthenticatedRequest("PUT", "/api/files/../../etc/malicious.sh", requestBody, server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("Expected status 403, got %d", w.Code)
		}

		var response APIResponse
		err = json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if response.Success {
			t.Error("Expected failed response for denied access")
		}
	})

	t.Run("put file - invalid JSON", func(t *testing.T) {
		req, err := createAuthenticatedRequest("PUT", "/api/files/test.sh", "invalid json", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
	})

	t.Run("validate file - valid script", func(t *testing.T) {
		requestBody := `{"content": "#!/bin/bash\necho 'Hello World'\nls -la"}`

		req, err := createAuthenticatedRequest("POST", "/api/files/validate", requestBody, server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

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

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		valid, ok := data["valid"].(bool)
		if !ok {
			t.Fatal("Expected valid to be a boolean")
		}

		if !valid {
			t.Error("Expected script to be valid")
		}
	})

	t.Run("validate file - script with issues", func(t *testing.T) {
		requestBody := `{"content": "#!/bin/bash\necho \"Unmatched quote\nsudo rm -rf /tmp/*"}`

		req, err := createAuthenticatedRequest("POST", "/api/files/validate", requestBody, server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

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

		data, ok := response.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		valid, ok := data["valid"].(bool)
		if !ok {
			t.Fatal("Expected valid to be a boolean")
		}

		if valid {
			t.Error("Expected script to be invalid")
		}

		issues, ok := data["issues"].([]interface{})
		if !ok {
			t.Fatal("Expected issues to be an array")
		}

		if len(issues) == 0 {
			t.Error("Expected validation issues")
		}
	})

	t.Run("list files - success", func(t *testing.T) {
		req, err := createAuthenticatedRequest("GET", "/api/files-list/", "", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

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

		data, ok := response.Data.([]interface{})
		if !ok {
			t.Fatal("Expected data to be an array")
		}

		// Should have at least the test files we created
		if len(data) == 0 {
			t.Error("Expected at least one file in listing")
		}
	})

	t.Run("list files - access denied", func(t *testing.T) {
		req, err := createAuthenticatedRequest("GET", "/api/files-list/../../etc", "", server)
		if err != nil {
			t.Fatalf("Failed to create authenticated request: %v", err)
		}
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		assertAccessDeniedResponse(t, w)
	})
}

func TestWebServer_FileOperations_NoFileManager(t *testing.T) {
	// Create web server without file manager
	server := NewWebServer(nil, 8080, "test-secret")
	// Note: not setting file manager

	t.Run("get file without file manager", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/files/test.sh", nil)
		w := httptest.NewRecorder()

		server.router.ServeHTTP(w, req)

		// When file manager is not set, routes are not registered, so we get 404
		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status 404, got %d", w.Code)
		}
	})
}
