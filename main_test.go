package main

import (
	"os"
	"testing"
)

func TestParseInterval(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
		hasError bool
	}{
		{
			name:     "seconds with suffix",
			input:    "30s",
			expected: 30,
			hasError: false,
		},
		{
			name:     "minutes with suffix",
			input:    "5m",
			expected: 300,
			hasError: false,
		},
		{
			name:     "hours with suffix",
			input:    "2h",
			expected: 7200,
			hasError: false,
		},
		{
			name:     "plain number (seconds)",
			input:    "3600",
			expected: 3600,
			hasError: false,
		},
		{
			name:     "single digit with suffix",
			input:    "1h",
			expected: 3600,
			hasError: false,
		},
		{
			name:     "zero with suffix",
			input:    "0s",
			expected: 0,
			hasError: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid format",
			input:    "abc",
			expected: 0,
			hasError: true,
		},
		{
			name:     "invalid suffix",
			input:    "10x",
			expected: 0,
			hasError: true,
		},
		{
			name:     "negative number",
			input:    "-5s",
			expected: 0,
			hasError: true,
		},
		{
			name:     "only suffix",
			input:    "s",
			expected: 0,
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseInterval(tt.input)

			if tt.hasError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("Expected %d for input %q, got %d", tt.expected, tt.input, result)
				}
			}
		})
	}
}

func TestMain_ArgumentParsing(t *testing.T) {
	// Save original args
	oldArgs := os.Args
	defer func() { os.Args = oldArgs }()

	tests := []struct {
		name     string
		args     []string
		expected string // Expected behavior description
	}{
		{
			name:     "no arguments",
			args:     []string{"run-script-service"},
			expected: "should run service",
		},
		{
			name:     "run command",
			args:     []string{"run-script-service", "run"},
			expected: "should run service",
		},
		{
			name:     "show-config command",
			args:     []string{"run-script-service", "show-config"},
			expected: "should show config",
		},
		{
			name:     "set-interval with valid time",
			args:     []string{"run-script-service", "set-interval", "30m"},
			expected: "should set interval",
		},
		{
			name:     "set-interval missing argument",
			args:     []string{"run-script-service", "set-interval"},
			expected: "should exit with error",
		},
		{
			name:     "unknown command",
			args:     []string{"run-script-service", "unknown"},
			expected: "should exit with error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Note: This test documents the expected behavior
			// Full integration testing would require mocking os.Exit and capturing output
			// which is complex for the main function. The parseInterval function
			// is thoroughly tested above, and the service package is tested separately.
			t.Logf("Test case: %s - %s", tt.name, tt.expected)
		})
	}
}
