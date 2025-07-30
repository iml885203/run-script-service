package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecutor_ExecuteScript(t *testing.T) {
	tests := []struct {
		name          string
		scriptContent string
		expectedExit  int
		expectStdout  bool
		expectStderr  bool
	}{
		{
			name:          "successful script",
			scriptContent: "#!/bin/bash\necho 'Hello World'",
			expectedExit:  0,
			expectStdout:  true,
			expectStderr:  false,
		},
		{
			name:          "failing script",
			scriptContent: "#!/bin/bash\nexit 1",
			expectedExit:  1,
			expectStdout:  false,
			expectStderr:  false,
		},
		{
			name:          "script with stderr",
			scriptContent: "#!/bin/bash\necho 'error' >&2\nexit 0",
			expectedExit:  0,
			expectStdout:  false,
			expectStderr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and script
			tempDir, err := os.MkdirTemp("", "executor_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			scriptPath := filepath.Join(tempDir, "test_script.sh")
			err = os.WriteFile(scriptPath, []byte(tt.scriptContent), 0755)
			if err != nil {
				t.Fatal(err)
			}

			logPath := filepath.Join(tempDir, "test.log")

			executor := NewExecutor(scriptPath, logPath, 100)
			result := executor.ExecuteScript()

			if result.ExitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d", tt.expectedExit, result.ExitCode)
			}

			if tt.expectStdout && result.Stdout == "" {
				t.Error("expected stdout but got none")
			}
			if !tt.expectStdout && result.Stdout != "" {
				t.Errorf("expected no stdout but got: %s", result.Stdout)
			}

			if tt.expectStderr && result.Stderr == "" {
				t.Error("expected stderr but got none")
			}
			if !tt.expectStderr && result.Stderr != "" {
				t.Errorf("expected no stderr but got: %s", result.Stderr)
			}

			// Check that log was written
			if _, err := os.Stat(logPath); os.IsNotExist(err) {
				t.Error("log file was not created")
			}
		})
	}
}

func TestExecutor_TrimLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "trim_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logPath := filepath.Join(tempDir, "test.log")

	// Create a log file with more than maxLines
	maxLines := 5
	lines := make([]string, 10)
	for i := 0; i < 10; i++ {
		lines[i] = "line " + string(rune('0'+i))
	}

	content := strings.Join(lines, "\n") + "\n"
	err = os.WriteFile(logPath, []byte(content), 0644)
	if err != nil {
		t.Fatal(err)
	}

	executor := NewExecutor("", logPath, maxLines)
	err = executor.TrimLog()
	if err != nil {
		t.Errorf("unexpected error trimming log: %v", err)
	}

	// Check that only the last maxLines lines remain
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	resultLines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(resultLines) != maxLines {
		t.Errorf("expected %d lines after trim, got %d", maxLines, len(resultLines))
	}

	// Check that it kept the last lines
	expectedLastLine := "line 9"
	if !strings.Contains(resultLines[len(resultLines)-1], expectedLastLine) {
		t.Errorf("expected last line to contain %q, got %q", expectedLastLine, resultLines[len(resultLines)-1])
	}
}

func TestExecutor_WriteLog(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "write_log_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logPath := filepath.Join(tempDir, "test.log")
	executor := NewExecutor("", logPath, 100)

	testContent := "test log entry"
	err = executor.WriteLog(testContent)
	if err != nil {
		t.Errorf("unexpected error writing log: %v", err)
	}

	// Check that content was written
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != testContent {
		t.Errorf("expected log content %q, got %q", testContent, string(data))
	}
}

func TestExecutor_LogError(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "log_error_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	logPath := filepath.Join(tempDir, "test.log")
	executor := NewExecutor("", logPath, 100)

	// Test the logError method indirectly by creating a script that doesn't exist
	// This will trigger the logError method in ExecuteScript
	executor.scriptPath = filepath.Join(tempDir, "nonexistent.sh")
	result := executor.ExecuteScript()

	// Check that the script failed
	if result.ExitCode == 0 {
		t.Error("expected non-zero exit code for nonexistent script")
	}

	// Check that error was logged
	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatal(err)
	}

	logContent := string(data)
	if !strings.Contains(logContent, "ERROR:") {
		t.Error("expected error message in log")
	}

	if !strings.Contains(logContent, "----") {
		t.Error("expected error separator in log")
	}
}
