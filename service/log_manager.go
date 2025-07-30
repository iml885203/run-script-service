// Package service provides core functionality for the run-script-service daemon
package service

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// LogManager manages multiple script loggers
type LogManager struct {
	loggers map[string]*ScriptLogger
	baseDir string
	mutex   sync.RWMutex
}

// ScriptLogger handles logging for a specific script
type ScriptLogger struct {
	scriptName string
	logPath    string
	maxLines   int
	entries    []LogEntry
	mutex      sync.RWMutex
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp  time.Time `json:"timestamp"`
	ScriptName string    `json:"script_name"`
	ExitCode   int       `json:"exit_code"`
	Stdout     string    `json:"stdout"`
	Stderr     string    `json:"stderr"`
	Duration   int64     `json:"duration_ms"`
}

// LogQuery defines criteria for querying logs
type LogQuery struct {
	ScriptName string    `json:"script_name,omitempty"`
	StartTime  time.Time `json:"start_time,omitempty"`
	EndTime    time.Time `json:"end_time,omitempty"`
	ExitCode   *int      `json:"exit_code,omitempty"`
	Limit      int       `json:"limit,omitempty"`
}

// NewLogManager creates a new LogManager instance
func NewLogManager(baseDir string) *LogManager {
	return &LogManager{
		loggers: make(map[string]*ScriptLogger),
		baseDir: baseDir,
	}
}

// GetLogger returns a logger for the specified script, creating one if it doesn't exist
func (lm *LogManager) GetLogger(scriptName string) *ScriptLogger {
	lm.mutex.Lock()
	defer lm.mutex.Unlock()

	if logger, exists := lm.loggers[scriptName]; exists {
		return logger
	}

	logger := NewScriptLogger(scriptName, lm.baseDir, 100) // Default max lines
	lm.loggers[scriptName] = logger
	return logger
}

// NewScriptLogger creates a new ScriptLogger instance
func NewScriptLogger(scriptName, baseDir string, maxLines int) *ScriptLogger {
	logPath := filepath.Join(baseDir, fmt.Sprintf("%s.log", scriptName))

	logger := &ScriptLogger{
		scriptName: scriptName,
		logPath:    logPath,
		maxLines:   maxLines,
		entries:    make([]LogEntry, 0),
	}

	// Ensure log directory exists
	_ = os.MkdirAll(baseDir, 0750) // Ignore error - logger will still work, file ops may fail later

	return logger
}

// AddEntry adds a new log entry to the script logger
func (sl *ScriptLogger) AddEntry(entry *LogEntry) error {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	// Add entry to in-memory storage
	sl.entries = append(sl.entries, *entry)

	// Maintain maxLines limit
	if len(sl.entries) > sl.maxLines {
		sl.entries = sl.entries[len(sl.entries)-sl.maxLines:]
	}

	// Write to file
	return sl.writeToFile(entry)
}

// writeToFile writes a log entry to the log file
func (sl *ScriptLogger) writeToFile(entry *LogEntry) error {
	file, err := os.OpenFile(sl.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	defer file.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal log entry: %w", err)
	}

	_, err = file.Write(append(data, '\n'))
	if err != nil {
		return fmt.Errorf("failed to write log entry: %w", err)
	}

	return nil
}

// GetEntries returns all log entries for this script
func (sl *ScriptLogger) GetEntries() []LogEntry {
	sl.mutex.RLock()
	defer sl.mutex.RUnlock()

	// Return a copy to prevent external modification
	entries := make([]LogEntry, len(sl.entries))
	copy(entries, sl.entries)
	return entries
}

// QueryLogs queries logs across all managed scripts
func (lm *LogManager) QueryLogs(query *LogQuery) ([]LogEntry, error) {
	lm.mutex.RLock()
	defer lm.mutex.RUnlock()

	var results []LogEntry

	// If specific script is requested
	if query.ScriptName != "" {
		if logger, exists := lm.loggers[query.ScriptName]; exists {
			entries := logger.GetEntries()
			results = append(results, entries...)
		}
	} else {
		// Query all scripts
		for _, logger := range lm.loggers {
			entries := logger.GetEntries()
			results = append(results, entries...)
		}
	}

	// Apply filters
	filtered := make([]LogEntry, 0)
	for i := range results {
		if lm.matchesQuery(&results[i], query) {
			filtered = append(filtered, results[i])
		}
	}

	// Apply limit
	if query.Limit > 0 && len(filtered) > query.Limit {
		filtered = filtered[len(filtered)-query.Limit:]
	}

	return filtered, nil
}

// matchesQuery checks if a log entry matches the query criteria
func (lm *LogManager) matchesQuery(entry *LogEntry, query *LogQuery) bool {
	// Check time range
	if !query.StartTime.IsZero() && entry.Timestamp.Before(query.StartTime) {
		return false
	}
	if !query.EndTime.IsZero() && entry.Timestamp.After(query.EndTime) {
		return false
	}

	// Check exit code
	if query.ExitCode != nil && entry.ExitCode != *query.ExitCode {
		return false
	}

	return true
}
