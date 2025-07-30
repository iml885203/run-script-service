package service

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLogManager_NewLogManager(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_logs")
	defer os.RemoveAll(baseDir)

	lm := NewLogManager(baseDir)

	if lm.baseDir != baseDir {
		t.Errorf("Expected baseDir %s, got %s", baseDir, lm.baseDir)
	}

	if lm.loggers == nil {
		t.Error("Expected loggers map to be initialized")
	}
}

func TestLogEntry_JSONSerialization(t *testing.T) {
	entry := LogEntry{
		Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		ScriptName: "backup",
		ExitCode:   0,
		Stdout:     "Backup completed successfully",
		Stderr:     "",
		Duration:   1500,
	}

	data, err := json.Marshal(entry)
	if err != nil {
		t.Fatalf("Failed to marshal LogEntry: %v", err)
	}

	var unmarshaled LogEntry
	err = json.Unmarshal(data, &unmarshaled)
	if err != nil {
		t.Fatalf("Failed to unmarshal LogEntry: %v", err)
	}

	if unmarshaled.ScriptName != entry.ScriptName {
		t.Errorf("Expected ScriptName %s, got %s", entry.ScriptName, unmarshaled.ScriptName)
	}

	if unmarshaled.ExitCode != entry.ExitCode {
		t.Errorf("Expected ExitCode %d, got %d", entry.ExitCode, unmarshaled.ExitCode)
	}
}

func TestScriptLogger_AddEntry(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_logs")
	defer os.RemoveAll(baseDir)

	logger := NewScriptLogger("test", baseDir, 100)

	entry := LogEntry{
		Timestamp:  time.Now(),
		ScriptName: "test",
		ExitCode:   0,
		Stdout:     "Test output",
		Stderr:     "",
		Duration:   100,
	}

	err := logger.AddEntry(&entry)
	if err != nil {
		t.Fatalf("Failed to add entry: %v", err)
	}

	if len(logger.entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(logger.entries))
	}

	if logger.entries[0].ScriptName != "test" {
		t.Errorf("Expected ScriptName 'test', got %s", logger.entries[0].ScriptName)
	}
}

func TestScriptLogger_MaxLines(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_logs")
	defer os.RemoveAll(baseDir)

	maxLines := 3
	logger := NewScriptLogger("test", baseDir, maxLines)

	// Add more entries than maxLines
	for i := 0; i < 5; i++ {
		entry := LogEntry{
			Timestamp:  time.Now(),
			ScriptName: "test",
			ExitCode:   i,
			Stdout:     "Test output",
			Stderr:     "",
			Duration:   100,
		}
		logger.AddEntry(&entry)
	}

	if len(logger.entries) > maxLines {
		t.Errorf("Expected max %d entries, got %d", maxLines, len(logger.entries))
	}

	// Check that the latest entries are kept
	if logger.entries[len(logger.entries)-1].ExitCode != 4 {
		t.Errorf("Expected last entry ExitCode 4, got %d", logger.entries[len(logger.entries)-1].ExitCode)
	}
}

func TestLogManager_GetLogger(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_logs")
	defer os.RemoveAll(baseDir)

	lm := NewLogManager(baseDir)

	logger1 := lm.GetLogger("script1")
	logger2 := lm.GetLogger("script1") // Same script name

	if logger1 != logger2 {
		t.Error("Expected same logger instance for same script name")
	}

	logger3 := lm.GetLogger("script2")
	if logger1 == logger3 {
		t.Error("Expected different logger instances for different script names")
	}
}

func TestLogQuery_Empty(t *testing.T) {
	query := LogQuery{}

	if query.Limit != 0 {
		t.Errorf("Expected default Limit 0, got %d", query.Limit)
	}
}
