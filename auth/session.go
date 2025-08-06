// Package auth provides authentication and session management
package auth

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Session represents an authentication session
type Session struct {
	UserID    string
	CreatedAt time.Time
	ExpiresAt time.Time
}

// SessionManager manages authentication sessions
type SessionManager struct {
	sessions map[string]*Session
	timeout  time.Duration
	mutex    sync.RWMutex
}

// NewSessionManager creates a new session manager with default 24 hour timeout
func NewSessionManager() *SessionManager {
	return NewSessionManagerWithTimeout(24 * time.Hour)
}

// NewSessionManagerWithTimeout creates a new session manager with custom timeout
func NewSessionManagerWithTimeout(timeout time.Duration) *SessionManager {
	return &SessionManager{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}
}

// CreateSession creates a new session and returns the session token
func (sm *SessionManager) CreateSession(userID string) (string, error) {
	// Generate cryptographically secure session token
	token, err := generateSessionToken()
	if err != nil {
		return "", fmt.Errorf("failed to generate session token: %v", err)
	}

	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// Create session
	session := &Session{
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(sm.timeout),
	}

	sm.sessions[token] = session

	return token, nil
}

// ValidateSession validates a session token and returns the user ID if valid
func (sm *SessionManager) ValidateSession(token string) (string, bool) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	session, exists := sm.sessions[token]
	if !exists {
		return "", false
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		// Clean up expired session
		go func() {
			sm.mutex.Lock()
			defer sm.mutex.Unlock()
			delete(sm.sessions, token)
		}()
		return "", false
	}

	return session.UserID, true
}

// DestroySession removes a session
func (sm *SessionManager) DestroySession(token string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	delete(sm.sessions, token)
	return nil
}

// generateSessionToken generates a cryptographically secure random session token
func generateSessionToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
