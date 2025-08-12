package service

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// DebugLogger provides conditional debug logging
type DebugLogger struct {
	enabled bool
	output  io.Writer
}

// NewDebugLogger creates a new debug logger
func NewDebugLogger() *DebugLogger {
	logger := &DebugLogger{
		output: os.Stdout,
	}

	// Check environment variable for debug setting
	debugEnv := strings.ToLower(os.Getenv("DEBUG"))
	logger.enabled = debugEnv == "true" || debugEnv == "1"

	return logger
}

// IsEnabled returns whether debug logging is enabled
func (d *DebugLogger) IsEnabled() bool {
	return d.enabled
}

// Enable enables debug logging
func (d *DebugLogger) Enable() {
	d.enabled = true
}

// Disable disables debug logging
func (d *DebugLogger) Disable() {
	d.enabled = false
}

// SetOutput sets the output destination for debug messages
func (d *DebugLogger) SetOutput(w io.Writer) {
	d.output = w
}

// Debugf prints a debug message if debugging is enabled
func (d *DebugLogger) Debugf(format string, args ...interface{}) {
	if !d.enabled {
		return
	}

	message := fmt.Sprintf(format, args...)
	fmt.Fprintln(d.output, message)
}
