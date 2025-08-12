package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnvLoader_LoadFromFile(t *testing.T) {
	// Create temporary .env file for testing
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	envContent := `WEB_SECRET_KEY=test-secret-key
LOG_LEVEL=debug
WEB_PORT=8080
DATABASE_PASSWORD=secret123
`

	err := os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Test loading .env file
	loader := NewEnvLoader()
	err = loader.LoadFromFile(envFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Test that values are accessible
	tests := []struct {
		key      string
		expected string
	}{
		{"WEB_SECRET_KEY", "test-secret-key"},
		{"LOG_LEVEL", "debug"},
		{"WEB_PORT", "8080"},
		{"DATABASE_PASSWORD", "secret123"},
	}

	for _, test := range tests {
		value := loader.Get(test.key)
		if value != test.expected {
			t.Errorf("Expected %s=%s, got %s", test.key, test.expected, value)
		}
	}
}

func TestEnvLoader_EnvVariablePriority(t *testing.T) {
	// Create temporary .env file
	tempDir := t.TempDir()
	envFile := filepath.Join(tempDir, ".env")

	envContent := `WEB_SECRET_KEY=file-secret
LOG_LEVEL=info
`

	err := os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Set environment variable (should take priority)
	os.Setenv("WEB_SECRET_KEY", "env-secret")
	defer os.Unsetenv("WEB_SECRET_KEY")

	loader := NewEnvLoader()
	err = loader.LoadFromFile(envFile)
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Environment variable should take priority over .env file
	if value := loader.Get("WEB_SECRET_KEY"); value != "env-secret" {
		t.Errorf("Expected WEB_SECRET_KEY=env-secret (from env), got %s", value)
	}

	// .env file value should be used when no env var exists
	if value := loader.Get("LOG_LEVEL"); value != "info" {
		t.Errorf("Expected LOG_LEVEL=info (from file), got %s", value)
	}
}

func TestEnvLoader_DefaultValues(t *testing.T) {
	loader := NewEnvLoader()

	// Test default value when key doesn't exist
	value := loader.GetWithDefault("NONEXISTENT_KEY", "default-value")
	if value != "default-value" {
		t.Errorf("Expected default-value, got %s", value)
	}

	// Test that existing value overrides default
	os.Setenv("TEST_KEY", "actual-value")
	defer os.Unsetenv("TEST_KEY")

	value = loader.GetWithDefault("TEST_KEY", "default-value")
	if value != "actual-value" {
		t.Errorf("Expected actual-value, got %s", value)
	}
}

func TestEnvLoader_FileNotExist(t *testing.T) {
	loader := NewEnvLoader()
	err := loader.LoadFromFile("nonexistent.env")

	// Should not fail when file doesn't exist
	if err != nil {
		t.Errorf("LoadFromFile should not fail when file doesn't exist, got: %v", err)
	}
}
