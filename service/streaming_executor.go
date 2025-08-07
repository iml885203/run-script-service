// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"bufio"
	"context"
	"io"
	"os"
	"strings"
	"sync"
	"time"
)

// StreamingExecutor defines the interface for executing scripts with streaming output
type StreamingExecutor interface {
	ExecuteWithStreaming(ctx context.Context, args ...string) *ExecutionResult
	SetLogHandler(handler LogHandler)
}

// LogHandler defines the interface for handling streaming log events
type LogHandler interface {
	HandleLogLine(timestamp time.Time, stream string, line string)
	HandleExecutionStart(timestamp time.Time)
	HandleExecutionEnd(timestamp time.Time, exitCode int)
}

// StreamingLogWriter handles real-time log writing with buffering
type StreamingLogWriter struct {
	logPath       string
	file          *os.File
	mutex         sync.Mutex
	buffer        *bufio.Writer
	flushInterval time.Duration
	bufferSize    int
	closed        bool
	stopChan      chan struct{}
	flushTicker   *time.Ticker
}

// NewStreamingLogWriter creates a new streaming log writer
func NewStreamingLogWriter(logPath string, bufferSize int, flushInterval time.Duration) *StreamingLogWriter {
	writer := &StreamingLogWriter{
		logPath:       logPath,
		bufferSize:    bufferSize,
		flushInterval: flushInterval,
		stopChan:      make(chan struct{}),
	}

	// Initialize the writer
	if err := writer.init(); err != nil {
		// Return a writer that will fail on WriteStreamLine
		return writer
	}

	return writer
}

// init initializes the log writer
func (w *StreamingLogWriter) init() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed {
		return nil
	}

	file, err := os.OpenFile(w.logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	w.file = file
	w.buffer = bufio.NewWriterSize(file, w.bufferSize)

	// Start flush ticker
	w.flushTicker = time.NewTicker(w.flushInterval)
	go w.flushLoop()

	return nil
}

// flushLoop periodically flushes the buffer
func (w *StreamingLogWriter) flushLoop() {
	for {
		select {
		case <-w.flushTicker.C:
			w.flush()
		case <-w.stopChan:
			return
		}
	}
}

// WriteStreamLine writes a single log line with timestamp and stream type
func (w *StreamingLogWriter) WriteStreamLine(timestamp time.Time, streamType string, line string) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed || w.buffer == nil {
		return nil // Fail silently for closed writer
	}

	logEntry := timestamp.Format("2006-01-02 15:04:05") + " [" + streamType + "] " + line + "\n"
	_, err := w.buffer.WriteString(logEntry)
	return err
}

// flush flushes the buffer to disk
func (w *StreamingLogWriter) flush() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.buffer != nil && !w.closed {
		_ = w.buffer.Flush()
	}
}

// Close closes the streaming log writer
func (w *StreamingLogWriter) Close() error {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.closed {
		return nil
	}

	w.closed = true
	close(w.stopChan)

	if w.flushTicker != nil {
		w.flushTicker.Stop()
	}

	if w.buffer != nil {
		_ = w.buffer.Flush()
	}

	if w.file != nil {
		return w.file.Close()
	}

	return nil
}

// streamOutput processes output from a reader line by line and sends to log handler
func (e *Executor) streamOutput(reader io.Reader, streamType string, handler LogHandler) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		line := scanner.Text()
		timestamp := time.Now()
		handler.HandleLogLine(timestamp, streamType, line)
	}
}

// StreamingLogHandler integrates streaming output with the log manager
type StreamingLogHandler struct {
	scriptName  string
	logManager  *LogManager
	startTime   time.Time
	stdoutLines []string
	stderrLines []string
	mutex       sync.Mutex
}

// NewStreamingLogHandler creates a new streaming log handler
func NewStreamingLogHandler(scriptName string, logManager *LogManager) *StreamingLogHandler {
	return &StreamingLogHandler{
		scriptName:  scriptName,
		logManager:  logManager,
		stdoutLines: make([]string, 0),
		stderrLines: make([]string, 0),
	}
}

// HandleLogLine processes individual log lines during streaming
func (h *StreamingLogHandler) HandleLogLine(timestamp time.Time, stream string, line string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	switch stream {
	case "STDOUT":
		h.stdoutLines = append(h.stdoutLines, line)
	case "STDERR":
		h.stderrLines = append(h.stderrLines, line)
	}
}

// HandleExecutionStart records the start time of execution
func (h *StreamingLogHandler) HandleExecutionStart(timestamp time.Time) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.startTime = timestamp
	h.stdoutLines = make([]string, 0)
	h.stderrLines = make([]string, 0)
}

// HandleExecutionEnd finalizes the log entry and adds it to the log manager
func (h *StreamingLogHandler) HandleExecutionEnd(timestamp time.Time, exitCode int) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Create final log entry with accumulated output
	entry := &LogEntry{
		Timestamp:  h.startTime,
		ScriptName: h.scriptName,
		ExitCode:   exitCode,
		Stdout:     strings.Join(h.stdoutLines, "\n"),
		Stderr:     strings.Join(h.stderrLines, "\n"),
		Duration:   timestamp.Sub(h.startTime).Nanoseconds() / 1e6, // Convert to milliseconds
	}

	// Add to log manager
	logger := h.logManager.GetLogger(h.scriptName)
	_ = logger.AddEntry(entry) // Ignore error for now - could be logged separately
}
