// Package main provides tests for CLI functionality
package main

import (
	"os"
	"path/filepath"
	"testing"

	"run-script-service/service"
)

func TestCLICommands(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()
	scriptPath := filepath.Join(tempDir, "test.sh")
	logPath := filepath.Join(tempDir, "test.log")
	configPath := filepath.Join(tempDir, "service_config.json")

	// Create a simple test script
	err := os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test'\n"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Test add-script command
	t.Run("add-script command", func(t *testing.T) {
		args := []string{"run-script-service", "add-script", "--name=test", "--path=" + scriptPath, "--interval=30s"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for add-script command")
		}

		// Verify config was updated
		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		found := false
		for _, script := range config.Scripts {
			if script.Name == "test" && script.Path == scriptPath && script.Interval == 30 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Script was not added to configuration")
		}
	})

	t.Run("list-scripts command", func(t *testing.T) {
		// First add a script
		config := &service.ServiceConfig{
			Scripts: []service.ScriptConfig{
				{
					Name:        "test1",
					Path:        scriptPath,
					Interval:    60,
					Enabled:     true,
					MaxLogLines: 100,
					Timeout:     0,
				},
				{
					Name:        "test2",
					Path:        scriptPath,
					Interval:    120,
					Enabled:     false,
					MaxLogLines: 50,
					Timeout:     30,
				},
			},
			WebPort: 8080,
		}
		err := service.SaveServiceConfig(configPath, config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		args := []string{"run-script-service", "list-scripts"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for list-scripts command")
		}
	})

	t.Run("enable-script command", func(t *testing.T) {
		args := []string{"run-script-service", "enable-script", "test2"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for enable-script command")
		}

		// Verify script was enabled
		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		found := false
		for _, script := range config.Scripts {
			if script.Name == "test2" && script.Enabled {
				found = true
				break
			}
		}
		if !found {
			t.Error("Script was not enabled")
		}
	})

	t.Run("disable-script command", func(t *testing.T) {
		args := []string{"run-script-service", "disable-script", "test1"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for disable-script command")
		}

		// Verify script was disabled
		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		found := false
		for _, script := range config.Scripts {
			if script.Name == "test1" && !script.Enabled {
				found = true
				break
			}
		}
		if !found {
			t.Error("Script was not disabled")
		}
	})

	t.Run("remove-script command", func(t *testing.T) {
		args := []string{"run-script-service", "remove-script", "test2"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for remove-script command")
		}

		// Verify script was removed
		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		for _, script := range config.Scripts {
			if script.Name == "test2" {
				t.Error("Script was not removed from configuration")
			}
		}
	})

	t.Run("run-script command", func(t *testing.T) {
		args := []string{"run-script-service", "run-script", "test1"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for run-script command")
		}
	})
}

func TestParseScriptFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected map[string]string
		hasError bool
	}{
		{
			name: "valid flags",
			args: []string{"--name=test", "--path=./test.sh", "--interval=30s"},
			expected: map[string]string{
				"name":     "test",
				"path":     "./test.sh",
				"interval": "30s",
			},
			hasError: false,
		},
		{
			name:     "missing required flag",
			args:     []string{"--name=test", "--interval=30s"},
			expected: nil,
			hasError: true,
		},
		{
			name: "with optional flags",
			args: []string{"--name=test", "--path=./test.sh", "--interval=30s", "--timeout=60", "--max-log-lines=200"},
			expected: map[string]string{
				"name":          "test",
				"path":          "./test.sh",
				"interval":      "30s",
				"timeout":       "60",
				"max-log-lines": "200",
			},
			hasError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseScriptFlags(tt.args)

			if tt.hasError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			for key, expected := range tt.expected {
				if actual, ok := result[key]; !ok || actual != expected {
					t.Errorf("Expected %s=%s, got %s=%s", key, expected, key, actual)
				}
			}
		})
	}
}
