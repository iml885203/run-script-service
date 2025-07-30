package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestService_NewService(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "run.sh")
	logPath := filepath.Join(tempDir, "run.log")
	configPath := filepath.Join(tempDir, "service_config.json")

	service := NewService(scriptPath, logPath, configPath, 100)

	if service.config.Interval != 3600 {
		t.Errorf("expected default interval 3600, got %d", service.config.Interval)
	}

	if service.scriptPath != scriptPath {
		t.Errorf("expected scriptPath %s, got %s", scriptPath, service.scriptPath)
	}

	if service.logPath != logPath {
		t.Errorf("expected logPath %s, got %s", logPath, service.logPath)
	}

	if service.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, service.configPath)
	}

	if service.maxLines != 100 {
		t.Errorf("expected maxLines 100, got %d", service.maxLines)
	}
}

func TestService_SetInterval(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "service_config.json")
	service := NewService("", "", configPath, 100)

	err = service.SetInterval(1800)
	if err != nil {
		t.Errorf("unexpected error setting interval: %v", err)
	}

	if service.config.Interval != 1800 {
		t.Errorf("expected interval 1800, got %d", service.config.Interval)
	}

	// Verify config was saved
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("error reading saved config: %v", err)
	}

	expectedContent := "{\n  \"interval\": 1800\n}"
	if string(data) != expectedContent {
		t.Errorf("expected config content %q, got %q", expectedContent, string(data))
	}
}

func TestService_Start_Stop(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a simple test script
	scriptPath := filepath.Join(tempDir, "test_script.sh")
	scriptContent := "#!/bin/bash\necho 'test execution'"
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	logPath := filepath.Join(tempDir, "test.log")
	configPath := filepath.Join(tempDir, "service_config.json")

	service := NewService(scriptPath, logPath, configPath, 100)

	// Set a short interval for testing
	service.config.Interval = 1 // 1 second

	// Start service in a goroutine
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	done := make(chan bool)
	go func() {
		service.Start(ctx)
		done <- true
	}()

	// Let it run for a bit
	time.Sleep(2500 * time.Millisecond)

	// Stop the service
	service.Stop()

	// Wait for service to finish
	select {
	case <-done:
		// Service stopped successfully
	case <-time.After(1 * time.Second):
		t.Error("service did not stop within timeout")
	}

	// Check that log file was created and has content
	if _, statErr := os.Stat(logPath); os.IsNotExist(statErr) {
		t.Error("log file was not created")
	}

	// Verify log has some content
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(data) == 0 {
		t.Error("log file is empty, expected some content")
	}
}

func TestService_LoadExistingConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "service_config.json")

	// Create a config file with custom interval
	configContent := `{"interval": 2400}`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatal(err)
	}

	service := NewService("", "", configPath, 100)

	if service.config.Interval != 2400 {
		t.Errorf("expected loaded interval 2400, got %d", service.config.Interval)
	}
}

func TestService_ShowConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "service_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	scriptPath := filepath.Join(tempDir, "run.sh")
	logPath := filepath.Join(tempDir, "run.log")
	configPath := filepath.Join(tempDir, "service_config.json")

	service := NewService(scriptPath, logPath, configPath, 100)

	// This test mainly verifies the function doesn't panic
	// Full output testing would require capturing stdout
	service.ShowConfig()
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name     string
		seconds  int
		expected string
	}{
		{
			name:     "seconds only",
			seconds:  30,
			expected: "30s",
		},
		{
			name:     "single minute",
			seconds:  60,
			expected: "1m",
		},
		{
			name:     "multiple minutes",
			seconds:  300,
			expected: "5m",
		},
		{
			name:     "single hour",
			seconds:  3600,
			expected: "1h",
		},
		{
			name:     "multiple hours",
			seconds:  7200,
			expected: "2h",
		},
		{
			name:     "zero seconds",
			seconds:  0,
			expected: "0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%d) = %q, expected %q", tt.seconds, result, tt.expected)
			}
		})
	}
}
