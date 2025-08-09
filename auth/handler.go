package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthHandler handles authentication-related HTTP requests
type AuthHandler struct {
	sessionManager *SessionManager
	secretKey      string
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	SecretKey string `json:"secretKey" binding:"required"`
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(secretKey string) *AuthHandler {
	sessionManager := NewSessionManagerWithTimeout(24 * time.Hour) // 24-hour session timeout
	return &AuthHandler{
		sessionManager: sessionManager,
		secretKey:      secretKey,
	}
}

// GetSessionManager returns the session manager instance
func (ah *AuthHandler) GetSessionManager() *SessionManager {
	return ah.sessionManager
}

// Login handles the login endpoint
func (ah *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request body",
		})
		return
	}

	// Verify secret key
	if req.SecretKey != ah.secretKey {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "Invalid secret key",
		})
		return
	}

	// Create session (using "authenticated" as user ID since we only have secret-based auth)
	sessionToken, err := ah.sessionManager.CreateSession("authenticated")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create session",
		})
		return
	}

	// Set session cookie
	c.SetCookie(
		"session",                   // name
		sessionToken,                // value
		int(24*time.Hour.Seconds()), // max age (24 hours)
		"/",                         // path
		"",                          // domain
		false,                       // secure (set to true in production with HTTPS)
		true,                        // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Login successful",
	})
}

// Logout handles the logout endpoint
func (ah *AuthHandler) Logout(c *gin.Context) {
	// Get session cookie
	sessionCookie, err := c.Request.Cookie("session")
	if err == nil {
		// Destroy the session
		ah.sessionManager.DestroySession(sessionCookie.Value)
	}

	// Clear the session cookie
	c.SetCookie(
		"session",
		"",
		-1,    // max age (expired)
		"/",   // path
		"",    // domain
		false, // secure
		true,  // httpOnly
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Logout successful",
	})
}

// AuthStatus returns the current authentication status
func (ah *AuthHandler) AuthStatus(c *gin.Context) {
	// Check if session cookie exists
	sessionCookie, err := c.Request.Cookie("session")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"authenticated": false,
			},
		})
		return
	}

	// Validate session
	_, valid := ah.sessionManager.ValidateSession(sessionCookie.Value)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"authenticated": valid,
		},
	})
}
