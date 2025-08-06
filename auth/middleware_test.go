package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewAuthMiddleware(t *testing.T) {
	sessionManager := NewSessionManager()
	middleware := NewAuthMiddleware(sessionManager)

	if middleware == nil {
		t.Fatal("expected middleware to be created")
	}
}

func TestAuthMiddleware_AuthenticatedRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	sessionManager := NewSessionManager()
	middleware := NewAuthMiddleware(sessionManager)

	// Create a session
	token, err := sessionManager.CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Create test router
	router := gin.New()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		userID := c.GetString("userID")
		c.JSON(http.StatusOK, gin.H{"message": "success", "userID": userID})
	})

	// Create request with valid session cookie
	req := httptest.NewRequest("GET", "/protected", nil)
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
}

func TestAuthMiddleware_UnauthenticatedRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	sessionManager := NewSessionManager()
	middleware := NewAuthMiddleware(sessionManager)

	// Create test router
	router := gin.New()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request without session cookie
	req := httptest.NewRequest("GET", "/protected", nil)

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidSessionRequest(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	sessionManager := NewSessionManager()
	middleware := NewAuthMiddleware(sessionManager)

	// Create test router
	router := gin.New()
	router.Use(middleware.RequireAuth())
	router.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// Create request with invalid session cookie
	req := httptest.NewRequest("GET", "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session",
		Value: "invalid-token",
	})

	// Execute request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify response
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}
