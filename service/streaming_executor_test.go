// Package service provides core functionality for the run-script-service daemon.
package service

import (
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
