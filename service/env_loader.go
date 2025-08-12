package service

import (
	"bufio"
	"os"
	"strings"
)

// EnvLoader handles loading environment variables from .env files
// with proper priority handling (environment variables > .env file > default)
type EnvLoader struct {
	values map[string]string
}

// NewEnvLoader creates a new environment variable loader
func NewEnvLoader() *EnvLoader {
	return &EnvLoader{
		values: make(map[string]string),
	}
}

// LoadFromFile loads environment variables from a .env file
// Environment variables take priority over file values
func (e *EnvLoader) LoadFromFile(filepath string) error {
	// File not existing is not an error - just continue with env vars only
	file, err := os.Open(filepath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value pairs
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Only set if not already set in environment
		if _, exists := e.values[key]; !exists {
			e.values[key] = value
		}
	}

	return scanner.Err()
}

// Get retrieves a value, checking environment variables first, then .env file values
func (e *EnvLoader) Get(key string) string {
	// Check environment variables first (highest priority)
	if value := os.Getenv(key); value != "" {
		return value
	}

	// Check .env file values
	if value, exists := e.values[key]; exists {
		return value
	}

	return ""
}

// GetWithDefault retrieves a value with a default fallback
func (e *EnvLoader) GetWithDefault(key, defaultValue string) string {
	if value := e.Get(key); value != "" {
		return value
	}
	return defaultValue
}
