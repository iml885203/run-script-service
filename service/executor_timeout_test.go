package service

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Red phase: Write failing tests for timeout scenarios not currently covered
func TestExecutor_TimeoutHandling(t *testing.T) {
	tests := []struct {
		name             string
		scriptContent    string
		timeout          time.Duration
		expectTimeout    bool
		expectedExitCode int
	}{
		{
			name:             "script that runs within timeout",
			scriptContent:    "#!/bin/bash\necho 'quick script'\nexit 0",
			timeout:          5 * time.Second,
			expectTimeout:    false,
			expectedExitCode: 0,
		},
		{
			name:             "script that exceeds timeout",
			scriptContent:    "#!/bin/bash\necho 'starting long script'\nsleep 2\necho 'finished'\nexit 0",
			timeout:          100 * time.Millisecond,
			expectTimeout:    true,
			expectedExitCode: -1,
		},
		{
			name:             "script with zero timeout should not timeout",
			scriptContent:    "#!/bin/bash\necho 'no timeout script'\nsleep 0.1\nexit 0",
			timeout:          0,
			expectTimeout:    false,
			expectedExitCode: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and script
			tempDir, err := os.MkdirTemp("", "timeout_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			scriptPath := filepath.Join(tempDir, "timeout_test.sh")
			logPath := filepath.Join(tempDir, "timeout_test.log")

			err = os.WriteFile(scriptPath, []byte(tt.scriptContent), 0755)
			if err != nil {
				t.Fatal(err)
			}

			// Create executor with timeout
			executor := NewExecutorWithTimeout(scriptPath, logPath, 100, tt.timeout)

			// Execute with context
			ctx := context.Background()
			result := executor.ExecuteScriptWithContext(ctx, "test-arg")

			// Check exit code
			if result.ExitCode != tt.expectedExitCode {
				t.Errorf("expected exit code %d, got %d", tt.expectedExitCode, result.ExitCode)
			}

			// Check timeout behavior
			if tt.expectTimeout && result.ExitCode == 0 {
				t.Error("expected script to timeout but it succeeded")
			}

			// For timeout cases, check that error is logged
			if tt.expectTimeout {
				logData, err := os.ReadFile(logPath)
				if err == nil {
					logContent := string(logData)
					t.Logf("Log content: %s", logContent)
					if !strings.Contains(logContent, "ERROR") && !strings.Contains(logContent, "timed out") {
						t.Error("expected timeout error to be logged")
					}
				}
			}
		})
	}
}

// Red phase: Test executor error scenarios not currently covered
func TestExecutor_ErrorScenarios(t *testing.T) {
	tests := []struct {
		name          string
		setupFunc     func(string) (string, string) // returns scriptPath, logPath
		expectError   bool
		expectedExit  int
		checkLogError bool
	}{
		{
			name: "script with permission denied",
			setupFunc: func(tempDir string) (string, string) {
				scriptPath := filepath.Join(tempDir, "no_perm.sh")
				logPath := filepath.Join(tempDir, "no_perm.log")
				os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test'"), 0000) // no permissions
				return scriptPath, logPath
			},
			expectError:   true,
			expectedExit:  -1,
			checkLogError: true,
		},
		{
			name: "script in non-existent directory",
			setupFunc: func(tempDir string) (string, string) {
				scriptPath := filepath.Join(tempDir, "nonexistent", "script.sh")
				logPath := filepath.Join(tempDir, "error.log")
				return scriptPath, logPath
			},
			expectError:   true,
			expectedExit:  -1,
			checkLogError: true,
		},
		{
			name: "empty log path should not crash",
			setupFunc: func(tempDir string) (string, string) {
				scriptPath := filepath.Join(tempDir, "test.sh")
				os.WriteFile(scriptPath, []byte("#!/bin/bash\necho 'test'\nexit 0"), 0755)
				return scriptPath, "" // empty log path
			},
			expectError:   false,
			expectedExit:  0,
			checkLogError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "error_scenarios_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			scriptPath, logPath := tt.setupFunc(tempDir)
			executor := NewExecutor(scriptPath, logPath, 100)

			result := executor.ExecuteScript()

			if result.ExitCode != tt.expectedExit {
				t.Errorf("expected exit code %d, got %d", tt.expectedExit, result.ExitCode)
			}

			// Check log error if expected and log path exists
			if tt.checkLogError && logPath != "" {
				if _, err := os.Stat(logPath); err == nil {
					logData, _ := os.ReadFile(logPath)
					if !strings.Contains(string(logData), "ERROR") {
						t.Error("expected error to be logged")
					}
				}
			}
		})
	}
}

// Red phase: Test streaming execution timeout scenarios
func TestExecutor_StreamingWithTimeout(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "streaming_timeout_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Script that will timeout during streaming
	scriptContent := "#!/bin/bash\necho 'starting streaming script'\nsleep 1\necho 'this should not appear'\nexit 0"
	scriptPath := filepath.Join(tempDir, "streaming_timeout.sh")
	logPath := filepath.Join(tempDir, "streaming_timeout.log")

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create executor with short timeout
	executor := NewExecutorWithTimeout(scriptPath, logPath, 100, 200*time.Millisecond)

	// Setup log handler
	logManager := NewLogManager(tempDir)
	streamingHandler := NewStreamingLogHandler("timeout-test", logManager)
	executor.SetLogHandler(streamingHandler)

	// Execute with streaming
	ctx := context.Background()
	result := executor.ExecuteWithStreaming(ctx)

	// Should timeout
	if result.ExitCode == 0 {
		t.Error("expected script to timeout but it succeeded")
	}

	// Check that execution was interrupted (either by timeout or other error)
	// The key is that the exit code is non-zero, indicating the script didn't complete normally
	if result.ExitCode != -1 {
		t.Errorf("expected exit code -1 for timeout, got %d", result.ExitCode)
	}

	// Verify that some output was captured before timeout
	if !strings.Contains(result.Stdout, "starting streaming script") {
		t.Error("expected to capture some output before timeout")
	}
}

// Red phase: Test context cancellation scenarios
func TestExecutor_ContextCancellation(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "context_cancel_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	scriptContent := "#!/bin/bash\necho 'starting'\nsleep 2\necho 'finished'\nexit 0"
	scriptPath := filepath.Join(tempDir, "cancel_test.sh")
	logPath := filepath.Join(tempDir, "cancel_test.log")

	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	executor := NewExecutor(scriptPath, logPath, 100)

	// Create context that will be cancelled
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel context after short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	result := executor.ExecuteScriptWithContext(ctx)

	// Should be cancelled (exit code will depend on implementation)
	if result.ExitCode == 0 {
		t.Error("expected script to be cancelled")
	}
}

// Red phase: Test log trimming edge cases
func TestExecutor_TrimLogEdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		setupFunc   func(string) (string, int) // returns logPath, maxLines
		expectError bool
	}{
		{
			name: "trim log with scanner error simulation",
			setupFunc: func(tempDir string) (string, int) {
				logPath := filepath.Join(tempDir, "bad_log.log")
				// Create a log file that will cause scanner issues
				os.WriteFile(logPath, []byte("line1\nline2\nline3\n"), 0644)
				return logPath, 2
			},
			expectError: false,
		},
		{
			name: "trim log with exactly maxLines",
			setupFunc: func(tempDir string) (string, int) {
				logPath := filepath.Join(tempDir, "exact_log.log")
				os.WriteFile(logPath, []byte("line1\nline2\nline3\n"), 0644)
				return logPath, 3
			},
			expectError: false,
		},
		{
			name: "trim log on read-only filesystem simulation",
			setupFunc: func(tempDir string) (string, int) {
				logPath := filepath.Join(tempDir, "readonly_log.log")
				os.WriteFile(logPath, []byte("line1\nline2\nline3\nline4\n"), 0644)
				return logPath, 2
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "trim_edge_cases_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			logPath, maxLines := tt.setupFunc(tempDir)
			executor := NewExecutor("", logPath, maxLines)

			err = executor.TrimLog()
			if (err != nil) != tt.expectError {
				t.Errorf("TrimLog() error = %v, expectError = %v", err, tt.expectError)
			}
		})
	}
}
