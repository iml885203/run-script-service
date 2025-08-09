package service

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestDebugLogger_NewDebugLogger(t *testing.T) {
	// Red phase - this test should fail initially
	logger := NewDebugLogger()
	if logger == nil {
		t.Error("Expected logger to be created, got nil")
	}
}

func TestDebugLogger_EnabledByEnvironment(t *testing.T) {
	// Red phase - test debug logger respects environment variable
	tests := []struct {
		name     string
		envValue string
		expected bool
	}{
		{
			name:     "debug enabled with true",
			envValue: "true",
			expected: true,
		},
		{
			name:     "debug enabled with 1",
			envValue: "1",
			expected: true,
		},
		{
			name:     "debug disabled with false",
			envValue: "false",
			expected: false,
		},
		{
			name:     "debug disabled when empty",
			envValue: "",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			os.Setenv("DEBUG", tt.envValue)
			defer os.Unsetenv("DEBUG")

			logger := NewDebugLogger()
			if logger.IsEnabled() != tt.expected {
				t.Errorf("Expected IsEnabled() = %v, got %v", tt.expected, logger.IsEnabled())
			}
		})
	}
}

func TestDebugLogger_LogOutput(t *testing.T) {
	// Red phase - test that debug messages are output when enabled
	var buf bytes.Buffer

	// This should fail until we implement debug logger
	logger := NewDebugLogger()
	logger.SetOutput(&buf)
	logger.Enable()

	logger.Debugf("test debug message: %s", "value")

	output := buf.String()
	if !strings.Contains(output, "test debug message: value") {
		t.Errorf("Expected debug output to contain message, got: %s", output)
	}
}

func TestDebugLogger_NoOutputWhenDisabled(t *testing.T) {
	// Red phase - test that no output when disabled
	var buf bytes.Buffer

	logger := NewDebugLogger()
	logger.SetOutput(&buf)
	logger.Disable() // Explicitly disable

	logger.Debugf("this should not appear")

	output := buf.String()
	if output != "" {
		t.Errorf("Expected no debug output when disabled, got: %s", output)
	}
}
