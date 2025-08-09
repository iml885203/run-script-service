package auth

import (
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	manager := NewSessionManager()
	if manager == nil {
		t.Fatal("expected session manager to be created")
	}
}

func TestCreateSession(t *testing.T) {
	manager := NewSessionManager()

	sessionToken, err := manager.CreateSession("test-user")
	if err != nil {
		t.Fatalf("expected session creation to succeed, got error: %v", err)
	}

	if sessionToken == "" {
		t.Fatal("expected session token to be non-empty")
	}

	if len(sessionToken) < 32 {
		t.Fatalf("expected session token to be at least 32 characters, got %d", len(sessionToken))
	}
}

func TestValidateSession(t *testing.T) {
	manager := NewSessionManager()

	// Create a session
	token, err := manager.CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Validate the session
	userID, valid := manager.ValidateSession(token)
	if !valid {
		t.Fatal("expected session to be valid")
	}

	if userID != "test-user" {
		t.Fatalf("expected user ID 'test-user', got '%s'", userID)
	}
}

func TestInvalidSession(t *testing.T) {
	manager := NewSessionManager()

	// Validate non-existent session
	userID, valid := manager.ValidateSession("invalid-token")
	if valid {
		t.Fatal("expected invalid session to be invalid")
	}

	if userID != "" {
		t.Fatalf("expected empty user ID for invalid session, got '%s'", userID)
	}
}

func TestDestroySession(t *testing.T) {
	manager := NewSessionManager()

	// Create a session
	token, err := manager.CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Verify session exists
	_, valid := manager.ValidateSession(token)
	if !valid {
		t.Fatal("expected session to be valid before destruction")
	}

	// Destroy session
	err = manager.DestroySession(token)
	if err != nil {
		t.Fatalf("failed to destroy session: %v", err)
	}

	// Verify session no longer exists
	_, valid = manager.ValidateSession(token)
	if valid {
		t.Fatal("expected session to be invalid after destruction")
	}
}

func TestSessionTimeout(t *testing.T) {
	manager := NewSessionManagerWithTimeout(1 * time.Millisecond) // Very short timeout for testing

	// Create a session
	token, err := manager.CreateSession("test-user")
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	// Wait for timeout
	time.Sleep(5 * time.Millisecond)

	// Session should be expired
	_, valid := manager.ValidateSession(token)
	if valid {
		t.Fatal("expected session to be invalid after timeout")
	}
}
