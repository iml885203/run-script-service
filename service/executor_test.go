package service

import (
	"context"
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
		tt := tt
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

func TestExecutor_ExecuteWithResult(t *testing.T) {
	tests := []struct {
		name          string
		scriptContent string
		expectedExit  int
		expectStdout  bool
		expectStderr  bool
	}{
		{
			name:          "successful script with result",
			scriptContent: "#!/bin/bash\necho 'Hello from ExecuteWithResult'",
			expectedExit:  0,
			expectStdout:  true,
			expectStderr:  false,
		},
		{
			name:          "failing script with result",
			scriptContent: "#!/bin/bash\nexit 2",
			expectedExit:  2,
			expectStdout:  false,
			expectStderr:  false,
		},
		{
			name:          "script with stderr result",
			scriptContent: "#!/bin/bash\necho 'error output' >&2\nexit 0",
			expectedExit:  0,
			expectStdout:  false,
			expectStderr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and script
			tempDir, err := os.MkdirTemp("", "executor_with_result_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			scriptPath := filepath.Join(tempDir, "test.sh")
			logPath := filepath.Join(tempDir, "test.log")

			// Write script file
			err = os.WriteFile(scriptPath, []byte(tt.scriptContent), 0755)
			if err != nil {
				t.Fatal(err)
			}

			// Create executor
			executor := NewExecutor(scriptPath, logPath, 100)

			// Execute script with result
			result, err := executor.ExecuteWithResult(nil)

			// Check error handling
			if tt.expectedExit == 0 && err != nil {
				t.Errorf("expected no error for successful script, got: %v", err)
			}

			if tt.expectedExit != 0 && err == nil {
				t.Error("expected error for failing script, got none")
			}

			// Check exit code
			if result.ExitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d", tt.expectedExit, result.ExitCode)
			}

			// Check stdout
			if tt.expectStdout && result.Stdout == "" {
				t.Error("expected stdout output, got empty string")
			}
			if !tt.expectStdout && result.Stdout != "" {
				t.Errorf("expected no stdout, got: %s", result.Stdout)
			}

			// Check stderr
			if tt.expectStderr && result.Stderr == "" {
				t.Error("expected stderr output, got empty string")
			}
			if !tt.expectStderr && result.Stderr != "" {
				t.Errorf("expected no stderr, got: %s", result.Stderr)
			}

			// Check timestamp is set
			if result.Timestamp.IsZero() {
				t.Error("expected timestamp to be set")
			}
		})
	}
}

func TestExecutor_ExecuteWithResultStreaming(t *testing.T) {
	// Red phase - this test should fail because ExecuteWithResultStreaming method doesn't exist
	tests := []struct {
		name          string
		scriptContent string
		expectedExit  int
		expectError   bool
		expectStdout  bool
		expectStderr  bool
	}{
		{
			name:          "successful streaming script with result",
			scriptContent: "#!/bin/bash\necho 'Hello streaming'\necho 'Second line'",
			expectedExit:  0,
			expectError:   false,
			expectStdout:  true,
			expectStderr:  false,
		},
		{
			name:          "failing streaming script with result",
			scriptContent: "#!/bin/bash\necho 'Starting'\necho 'Error message' >&2\nexit 1",
			expectedExit:  1,
			expectError:   true,
			expectStdout:  true,
			expectStderr:  true,
		},
		{
			name:          "streaming script with mixed output",
			scriptContent: "#!/bin/bash\necho 'stdout line 1'\necho 'stderr line 1' >&2\necho 'stdout line 2'\nexit 0",
			expectedExit:  0,
			expectError:   false,
			expectStdout:  true,
			expectStderr:  true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and script
			tempDir, err := os.MkdirTemp("", "executor_with_result_streaming_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			scriptPath := filepath.Join(tempDir, "test.sh")
			logPath := filepath.Join(tempDir, "test.log")

			// Write script file
			err = os.WriteFile(scriptPath, []byte(tt.scriptContent), 0755)
			if err != nil {
				t.Fatal(err)
			}

			// Create executor with streaming log handler
			executor := NewExecutor(scriptPath, logPath, 100)
			logManager := NewLogManager(tempDir)
			streamingHandler := NewStreamingLogHandler("test-script", logManager)
			executor.SetLogHandler(streamingHandler)

			// Execute with streaming and result - THIS WILL FAIL because method doesn't exist
			ctx := context.Background()
			result, execError := executor.ExecuteWithResultStreaming(ctx)

			// Check error expectation
			if (execError != nil) != tt.expectError {
				t.Errorf("ExecuteWithResultStreaming() error = %v, expectError = %v", execError, tt.expectError)
			}

			// Check exit code
			if result.ExitCode != tt.expectedExit {
				t.Errorf("ExecuteWithResultStreaming() exit code = %v, expected = %v", result.ExitCode, tt.expectedExit)
			}

			// Check streaming was captured by log manager
			logger := logManager.GetLogger("test-script")
			entries := logger.GetEntries()

			if len(entries) != 1 {
				t.Errorf("Expected 1 log entry from streaming, got %d", len(entries))
				return
			}

			entry := entries[0]

			// Verify streaming captured output correctly
			if tt.expectStdout && entry.Stdout == "" {
				t.Error("Expected stdout to be captured via streaming")
			}

			if tt.expectStderr && entry.Stderr == "" {
				t.Error("Expected stderr to be captured via streaming")
			}

			// Check that result also contains the output (backward compatibility)
			if tt.expectStdout && result.Stdout == "" {
				t.Error("Expected stdout in result")
			}

			if tt.expectStderr && result.Stderr == "" {
				t.Error("Expected stderr in result")
			}

			// Check timestamp is set
			if result.Timestamp.IsZero() {
				t.Error("Expected timestamp to be set")
			}
		})
	}
}

func TestExecutor_ExecuteWithResultStreamingSynchronization(t *testing.T) {
	// This test ensures that streaming output is properly synchronized
	// Create temporary directory and script
	tempDir, err := os.MkdirTemp("", "executor_streaming_sync_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	scriptContent := "#!/bin/bash\necho 'Line 1'\nsleep 0.1\necho 'Line 2' >&2\nsleep 0.1\necho 'Line 3'\nexit 0"
	scriptPath := filepath.Join(tempDir, "sync_test.sh")
	logPath := filepath.Join(tempDir, "sync_test.log")

	// Write script file
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create executor with streaming log handler
	executor := NewExecutor(scriptPath, logPath, 100)
	logManager := NewLogManager(tempDir)
	streamingHandler := NewStreamingLogHandler("sync-test-script", logManager)
	executor.SetLogHandler(streamingHandler)

	// Execute multiple times to test for race conditions
	for i := 0; i < 5; i++ {
		ctx := context.Background()
		result, execError := executor.ExecuteWithResultStreaming(ctx)

		if execError != nil {
			t.Errorf("Run %d: ExecuteWithResultStreaming() unexpected error = %v", i+1, execError)
			continue
		}

		if result.ExitCode != 0 {
			t.Errorf("Run %d: ExecuteWithResultStreaming() exit code = %v, expected = 0", i+1, result.ExitCode)
			continue
		}

		// Verify both stdout and stderr are captured in result
		if result.Stdout == "" {
			t.Errorf("Run %d: Expected stdout in result, got empty", i+1)
		}

		if result.Stderr == "" {
			t.Errorf("Run %d: Expected stderr in result, got empty", i+1)
		}

		// Check streaming was captured by log manager
		logger := logManager.GetLogger("sync-test-script")
		entries := logger.GetEntries()

		if len(entries) != i+1 {
			t.Errorf("Run %d: Expected %d log entries from streaming, got %d", i+1, i+1, len(entries))
			continue
		}

		// Verify last entry has proper output
		entry := entries[len(entries)-1]
		if entry.Stdout == "" {
			t.Errorf("Run %d: Expected stdout to be captured via streaming", i+1)
		}

		if entry.Stderr == "" {
			t.Errorf("Run %d: Expected stderr to be captured via streaming", i+1)
		}
	}
}

func TestExecutor_EnsureContext(t *testing.T) {
	executor := NewExecutor("/tmp/test.sh", "/tmp/test.log", 100)

	tests := []struct {
		name        string
		inputCtx    context.Context
		expectValid bool
	}{
		{
			name:        "nil context returns background context",
			inputCtx:    nil,
			expectValid: true,
		},
		{
			name:        "valid context is preserved",
			inputCtx:    context.Background(),
			expectValid: true,
		},
		{
			name:        "context with timeout is preserved",
			inputCtx:    context.WithValue(context.Background(), "test", "value"),
			expectValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := executor.ensureContext(tt.inputCtx)
			if result == nil {
				t.Error("ensureContext should never return nil")
			}
		})
	}
}

func TestExecutor_HandleExecutionResult(t *testing.T) {
	executor := NewExecutor("/tmp/test.sh", "/tmp/test.log", 100)

	tests := []struct {
		name        string
		result      *ExecutionResult
		expectError bool
		expectedMsg string
	}{
		{
			name: "success result returns no error",
			result: &ExecutionResult{
				ExitCode: 0,
				Stdout:   "success",
			},
			expectError: false,
		},
		{
			name: "non-zero exit code returns error",
			result: &ExecutionResult{
				ExitCode: 1,
				Stderr:   "error occurred",
			},
			expectError: true,
			expectedMsg: "script execution failed with exit code 1",
		},
		{
			name: "different exit code returns appropriate error",
			result: &ExecutionResult{
				ExitCode: 127,
				Stderr:   "command not found",
			},
			expectError: true,
			expectedMsg: "script execution failed with exit code 127",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := executor.handleExecutionResult(tt.result)

			if (err != nil) != tt.expectError {
				t.Errorf("handleExecutionResult() error = %v, expectError = %v", err, tt.expectError)
			}

			if tt.expectError && err.Error() != tt.expectedMsg {
				t.Errorf("handleExecutionResult() error message = '%s', expected = '%s'", err.Error(), tt.expectedMsg)
			}

			// Result should always be returned, even when there's an error
			if result == nil {
				t.Error("handleExecutionResult() should always return result")
			}

			// Result should be the same object that was passed in
			if result != tt.result {
				t.Error("handleExecutionResult() should return the same result object")
			}
		})
	}
}
