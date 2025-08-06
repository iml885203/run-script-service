package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewAuthHandler(t *testing.T) {
	handler := NewAuthHandler("test-secret")

	if handler == nil {
		t.Fatal("expected handler to be created")
	}
	
	if handler.GetSessionManager() == nil {
		t.Fatal("expected session manager to be initialized")
	}
}

func TestLogin_ValidSecret(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler("test-secret")

	// Create test router
	router := gin.New()
	router.POST("/login", handler.Login)

	// Create login request
	loginData := map[string]string{
		"secretKey": "test-secret",
	}
	jsonData, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if !response["success"].(bool) {
		t.Fatal("expected success to be true")
	}

	// Check that session cookie is set
	cookies := w.Result().Cookies()
	sessionCookieFound := false
	for _, cookie := range cookies {
		if cookie.Name == "session" {
			sessionCookieFound = true
			if cookie.Value == "" {
				t.Fatal("expected session cookie to have a value")
			}
		}
	}
	if !sessionCookieFound {
		t.Fatal("expected session cookie to be set")
	}
}

func TestLogin_InvalidSecret(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler("test-secret")

	// Create test router
	router := gin.New()
	router.POST("/login", handler.Login)

	// Create login request with wrong secret
	loginData := map[string]string{
		"secretKey": "wrong-secret",
	}
	jsonData, _ := json.Marshal(loginData)
	req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	if response["success"].(bool) {
		t.Fatal("expected success to be false")
	}
}

func TestLogout(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler("test-secret")

	// Create a session
	token, err := handler.GetSessionManager().CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create test router
	router := gin.New()
	router.POST("/logout", handler.Logout)

	// Create logout request with session cookie
	req := httptest.NewRequest("POST", "/logout", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Verify session is destroyed
	_, valid := handler.GetSessionManager().ValidateSession(token)
	if valid {
		t.Fatal("expected session to be destroyed after logout")
	}
}

func TestAuthStatus_Authenticated(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler("test-secret")

	// Create a session
	token, err := handler.GetSessionManager().CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create test router
	router := gin.New()
	router.GET("/auth/status", handler.AuthStatus)

	// Create request with session cookie
	req := httptest.NewRequest("GET", "/auth/status", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: token,
	})

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if !data["authenticated"].(bool) {
		t.Fatal("expected authenticated to be true")
	}
}

func TestAuthStatus_Unauthenticated(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	handler := NewAuthHandler("test-secret")

	// Create test router
	router := gin.New()
	router.GET("/auth/status", handler.AuthStatus)

	// Create request without session cookie
	req := httptest.NewRequest("GET", "/auth/status", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	// Check response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data := response["data"].(map[string]interface{})
	if data["authenticated"].(bool) {
		t.Fatal("expected authenticated to be false")
	}
}
