// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestStreamingExecutor_Interface(t *testing.T) {
	// Test that StreamingExecutor interface is properly defined
	var _ StreamingExecutor = (*Executor)(nil) // This should compile when interface is implemented
}

func TestStreamingLogWriter_WriteStreamLine(t *testing.T) {
	// Red phase - test should fail initially
	tests := []struct {
		name       string
		streamType string
		line       string
		wantError  bool
	}{
		{
			name:       "stdout line",
			streamType: "STDOUT",
			line:       "test output line",
			wantError:  false,
		},
		{
			name:       "stderr line",
			streamType: "STDERR",
			line:       "error message",
			wantError:  false,
		},
		{
			name:       "empty line",
			streamType: "STDOUT",
			line:       "",
			wantError:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// This test will fail until we implement StreamingLogWriter
			writer := NewStreamingLogWriter("/tmp/test.log", 4096, 100*time.Millisecond)
			defer writer.Close()

			timestamp := time.Now()
			err := writer.WriteStreamLine(timestamp, tt.streamType, tt.line)

			if (err != nil) != tt.wantError {
				t.Errorf("WriteStreamLine() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestLogHandler_Interface(t *testing.T) {
	// Test LogHandler interface requirements
	handler := &MockLogHandler{}

	// Test HandleLogLine
	timestamp := time.Now()
	handler.HandleLogLine(timestamp, "STDOUT", "test line")

	if len(handler.LogLines) != 1 {
		t.Errorf("Expected 1 log line, got %d", len(handler.LogLines))
	}

	if handler.LogLines[0].Line != "test line" {
		t.Errorf("Expected 'test line', got '%s'", handler.LogLines[0].Line)
	}
}

// MockLogHandler for testing
type MockLogHandler struct {
	LogLines       []MockLogLine
	ExecutionStart *time.Time
	ExecutionEnd   *MockExecutionEnd
}

type MockLogLine struct {
	Timestamp  time.Time
	StreamType string
	Line       string
}

type MockExecutionEnd struct {
	Timestamp time.Time
	ExitCode  int
}

func (m *MockLogHandler) HandleLogLine(timestamp time.Time, streamType string, line string) {
	m.LogLines = append(m.LogLines, MockLogLine{
		Timestamp:  timestamp,
		StreamType: streamType,
		Line:       line,
	})
}

func (m *MockLogHandler) HandleExecutionStart(timestamp time.Time) {
	m.ExecutionStart = &timestamp
}

func (m *MockLogHandler) HandleExecutionEnd(timestamp time.Time, exitCode int) {
	m.ExecutionEnd = &MockExecutionEnd{
		Timestamp: timestamp,
		ExitCode:  exitCode,
	}
}

func TestStreamOutput_Processing(t *testing.T) {
	// Test streaming output processing with mock data
	mockOutput := "line 1\nline 2\nline 3\n"
	reader := strings.NewReader(mockOutput)

	handler := &MockLogHandler{}

	// This will fail until we implement streamOutput
	executor := NewExecutor("", "", 100)
	executor.streamOutput(reader, "STDOUT", handler)

	expectedLines := []string{"line 1", "line 2", "line 3"}
	if len(handler.LogLines) != len(expectedLines) {
		t.Errorf("Expected %d lines, got %d", len(expectedLines), len(handler.LogLines))
	}

	for i, expectedLine := range expectedLines {
		if i < len(handler.LogLines) && handler.LogLines[i].Line != expectedLine {
			t.Errorf("Expected line %d to be '%s', got '%s'", i, expectedLine, handler.LogLines[i].Line)
		}
	}
}

func TestExecuteWithStreaming_RealImplementation(t *testing.T) {
	// Red phase - this test should fail until we implement real streaming
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/test_script.sh"
	logPath := tmpDir + "/test.log"

	// Create a test script that outputs to both stdout and stderr
	scriptContent := `#!/bin/bash
echo "stdout line 1"
echo "stderr line 1" >&2
echo "stdout line 2"
echo "stderr line 2" >&2
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	executor := NewExecutor(scriptPath, logPath, 100)
	handler := &MockLogHandler{}
	executor.SetLogHandler(handler)

	ctx := context.Background()
	result := executor.ExecuteWithStreaming(ctx)

	// Check that we got some log lines processed
	if len(handler.LogLines) == 0 {
		t.Error("Expected streaming log lines, but got none")
	}

	// Check that execution completed successfully
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Check that both HandleExecutionStart and HandleExecutionEnd were called
	if handler.ExecutionStart == nil {
		t.Error("Expected HandleExecutionStart to be called")
	}

	if handler.ExecutionEnd == nil {
		t.Error("Expected HandleExecutionEnd to be called")
	}
}

func TestStreamingLogHandler_Integration(t *testing.T) {
	// Red phase - test streaming log handler integration with log manager
	tmpDir := t.TempDir()
	logManager := NewLogManager(tmpDir)
	scriptName := "test-script"

	// Create streaming log handler that integrates with log manager
	handler := NewStreamingLogHandler(scriptName, logManager)

	// Simulate streaming log events
	startTime := time.Now()
	handler.HandleExecutionStart(startTime)

	handler.HandleLogLine(time.Now(), "STDOUT", "line 1")
	handler.HandleLogLine(time.Now(), "STDERR", "error line 1")
	handler.HandleLogLine(time.Now(), "STDOUT", "line 2")

	endTime := time.Now()
	handler.HandleExecutionEnd(endTime, 0)

	// Check that log manager has the completed entry
	logger := logManager.GetLogger(scriptName)
	entries := logger.GetEntries()

	if len(entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(entries))
	}

	if len(entries) > 0 {
		entry := entries[0]
		if entry.ExitCode != 0 {
			t.Errorf("Expected exit code 0, got %d", entry.ExitCode)
		}

		if !strings.Contains(entry.Stdout, "line 1") || !strings.Contains(entry.Stdout, "line 2") {
			t.Errorf("Expected stdout to contain streaming lines, got: %s", entry.Stdout)
		}

		if !strings.Contains(entry.Stderr, "error line 1") {
			t.Errorf("Expected stderr to contain error line, got: %s", entry.Stderr)
		}
	}
}

func TestExecuteWithStreaming_LogManagerIntegration(t *testing.T) {
	// Integration test: Executor + StreamingLogHandler + LogManager
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/test_script.sh"
	logPath := tmpDir + "/test.log"

	// Create a test script
	scriptContent := `#!/bin/bash
echo "Hello from stdout"
echo "Error from stderr" >&2
echo "Another stdout line"
exit 0
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Set up executor with streaming log handler
	executor := NewExecutor(scriptPath, logPath, 100)
	logManager := NewLogManager(tmpDir)
	streamingHandler := NewStreamingLogHandler("integration-test", logManager)
	executor.SetLogHandler(streamingHandler)

	// Execute with streaming
	ctx := context.Background()
	result := executor.ExecuteWithStreaming(ctx)

	// Verify execution result
	if result.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", result.ExitCode)
	}

	// Verify log manager received the streamed data
	logger := logManager.GetLogger("integration-test")
	entries := logger.GetEntries()

	if len(entries) != 1 {
		t.Errorf("Expected 1 log entry, got %d", len(entries))
		return
	}

	entry := entries[0]

	// Check that streaming captured all output properly
	if !strings.Contains(entry.Stdout, "Hello from stdout") {
		t.Errorf("Expected stdout to contain 'Hello from stdout', got: %s", entry.Stdout)
	}

	if !strings.Contains(entry.Stdout, "Another stdout line") {
		t.Errorf("Expected stdout to contain 'Another stdout line', got: %s", entry.Stdout)
	}

	if !strings.Contains(entry.Stderr, "Error from stderr") {
		t.Errorf("Expected stderr to contain 'Error from stderr', got: %s", entry.Stderr)
	}

	if entry.ExitCode != 0 {
		t.Errorf("Expected logged exit code 0, got %d", entry.ExitCode)
	}
}

func TestExecuteWithStreaming_TimeoutSupport(t *testing.T) {
	// Red phase - this test should fail until we implement timeout support
	tmpDir := t.TempDir()
	scriptPath := tmpDir + "/timeout_script.sh"
	logPath := tmpDir + "/timeout.log"

	// Create a long-running test script
	scriptContent := `#!/bin/bash
echo "Starting long operation"
sleep 5
echo "Operation completed"
`
	err := os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatal(err)
	}

	// Create executor with timeout capability
	executor := NewExecutorWithTimeout(scriptPath, logPath, 100, 2*time.Second)
	handler := &MockLogHandler{}
	executor.SetLogHandler(handler)

	startTime := time.Now()
	ctx := context.Background()
	result := executor.ExecuteWithStreaming(ctx)
	duration := time.Since(startTime)

	// Should timeout after ~2 seconds, not complete the full 5-second sleep
	if duration > 3*time.Second {
		t.Errorf("Expected execution to timeout in ~2 seconds, but took %v", duration)
	}

	// Should have a timeout exit code (typically -1 or specific timeout code)
	if result.ExitCode == 0 {
		t.Error("Expected non-zero exit code due to timeout")
	}

	// Handler should receive timeout notification
	if handler.ExecutionEnd == nil {
		t.Error("Expected HandleExecutionEnd to be called for timeout")
	}

	// Should have received some log lines before timeout
	if len(handler.LogLines) == 0 {
		t.Error("Expected to receive log lines before timeout")
	}

	// Should contain the "Starting long operation" but not "Operation completed"
	foundStart := false
	foundComplete := false
	for _, line := range handler.LogLines {
		if line.Line == "Starting long operation" {
			foundStart = true
		}
		if line.Line == "Operation completed" {
			foundComplete = true
		}
	}

	if !foundStart {
		t.Error("Expected to find 'Starting long operation' in logs")
	}
	if foundComplete {
		t.Error("Should not have found 'Operation completed' due to timeout")
	}
}
