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

	t.Run("logs command - all scripts", func(t *testing.T) {
		args := []string{"run-script-service", "logs", "--all"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for logs command")
		}
	})

	t.Run("logs command - specific script", func(t *testing.T) {
		args := []string{"run-script-service", "logs", "--script=test1"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for logs command")
		}
	})

	t.Run("logs command - with filters", func(t *testing.T) {
		args := []string{"run-script-service", "logs", "--script=test1", "--exit-code=0", "--limit=10"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for logs command")
		}
	})

	t.Run("clear-logs command - specific script", func(t *testing.T) {
		args := []string{"run-script-service", "clear-logs", "--script=test1"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for clear-logs command")
		}
	})

	t.Run("clear-logs command - missing script", func(t *testing.T) {
		args := []string{"run-script-service", "clear-logs"}
		_, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for clear-logs command without --script flag")
		}
	})

	t.Run("add-script command - duplicate name", func(t *testing.T) {
		// First add a script
		config := &service.ServiceConfig{
			Scripts: []service.ScriptConfig{
				{
					Name:        "existing",
					Path:        scriptPath,
					Interval:    60,
					Enabled:     true,
					MaxLogLines: 100,
					Timeout:     0,
				},
			},
			WebPort: 8080,
		}
		err := service.SaveServiceConfig(configPath, config)
		if err != nil {
			t.Fatalf("Failed to save config: %v", err)
		}

		// Try to add script with same name
		args := []string{"run-script-service", "add-script", "--name=existing", "--path=" + scriptPath, "--interval=30s"}
		_, err = handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for duplicate script name")
		}
		if err != nil && !contains(err.Error(), "already exists") {
			t.Errorf("Expected 'already exists' error, got: %v", err)
		}
	})

	t.Run("add-script command - invalid script path", func(t *testing.T) {
		// Try to add script with non-existent path
		args := []string{"run-script-service", "add-script", "--name=invalid", "--path=/nonexistent/script.sh", "--interval=30s"}
		_, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for non-existent script path")
		}
		if err != nil && !contains(err.Error(), "invalid script configuration") {
			t.Errorf("Expected 'invalid script configuration' error, got: %v", err)
		}
	})

	t.Run("add-script command - with optional parameters", func(t *testing.T) {
		// Test with max-log-lines and timeout parameters
		args := []string{"run-script-service", "add-script", "--name=optional", "--path=" + scriptPath, "--interval=1m", "--max-log-lines=200", "--timeout=30"}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false for add-script command")
		}

		// Verify config was updated with optional parameters
		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		found := false
		for _, script := range config.Scripts {
			if script.Name == "optional" && script.MaxLogLines == 200 && script.Timeout == 30 {
				found = true
				break
			}
		}
		if !found {
			t.Error("Script with optional parameters was not added correctly")
		}
	})

	testInvalidParameterFallback := func(name, flag, value string, checkFunc func(service.ScriptConfig) bool) {
		args := []string{"run-script-service", "add-script", "--name=" + name, "--path=" + scriptPath, "--interval=1m", flag + "=" + value}
		result, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err != nil {
			t.Errorf("Expected no error (should use default), got: %v", err)
		}

		if result.shouldRunService {
			t.Error("Expected shouldRunService to be false")
		}

		var config service.ServiceConfig
		err = service.LoadServiceConfig(configPath, &config)
		if err != nil {
			t.Fatalf("Failed to load config: %v", err)
		}

		found := false
		for _, script := range config.Scripts {
			if script.Name == name && checkFunc(script) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Script should use default %s when invalid value provided", flag)
		}
	}

	t.Run("add-script command - invalid max-log-lines", func(t *testing.T) {
		testInvalidParameterFallback("invalid-max-log", "--max-log-lines", "invalid",
			func(s service.ScriptConfig) bool { return s.MaxLogLines == 100 })
	})

	t.Run("add-script command - invalid timeout", func(t *testing.T) {
		testInvalidParameterFallback("invalid-timeout", "--timeout", "invalid",
			func(s service.ScriptConfig) bool { return s.Timeout == 0 })
	})

	t.Run("add-script command - invalid interval", func(t *testing.T) {
		// Test with invalid interval format
		args := []string{"run-script-service", "add-script", "--name=invalid-interval", "--path=" + scriptPath, "--interval=invalid"}
		_, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for invalid interval")
		}
		if err != nil && !contains(err.Error(), "invalid interval") {
			t.Errorf("Expected 'invalid interval' error, got: %v", err)
		}
	})

	t.Run("logs command - invalid exit code", func(t *testing.T) {
		args := []string{"run-script-service", "logs", "--exit-code=invalid"}
		_, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for invalid exit-code")
		}
		if err != nil && !contains(err.Error(), "invalid exit-code") {
			t.Errorf("Expected 'invalid exit-code' error, got: %v", err)
		}
	})

	t.Run("logs command - invalid limit", func(t *testing.T) {
		args := []string{"run-script-service", "logs", "--limit=invalid"}
		_, err := handleCommand(args, scriptPath, logPath, configPath, 100)

		if err == nil {
			t.Error("Expected error for invalid limit")
		}
		if err != nil && !contains(err.Error(), "invalid limit") {
			t.Errorf("Expected 'invalid limit' error, got: %v", err)
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
