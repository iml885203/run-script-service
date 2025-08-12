package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware provides authentication middleware for HTTP requests
type AuthMiddleware struct {
	sessionManager *SessionManager
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(sessionManager *SessionManager) *AuthMiddleware {
	return &AuthMiddleware{
		sessionManager: sessionManager,
	}
}

// IsAuthenticated checks if the current request is authenticated
func (am *AuthMiddleware) IsAuthenticated(c *gin.Context) bool {
	// Get session cookie
	sessionCookie, err := c.Request.Cookie("session")
	if err != nil {
		return false
	}

	// Validate session
	_, valid := am.sessionManager.ValidateSession(sessionCookie.Value)
	return valid
}

// RequireAuth returns a Gin middleware that requires authentication
func (am *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get session cookie
		sessionCookie, err := c.Request.Cookie("session")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Authentication required",
			})
			c.Abort()
			return
		}

		// Validate session
		userID, valid := am.sessionManager.ValidateSession(sessionCookie.Value)
		if !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "Invalid or expired session",
			})
			c.Abort()
			return
		}

		// Set user ID in context for handlers to use
		c.Set("userID", userID)
		c.Next()
	}
}
