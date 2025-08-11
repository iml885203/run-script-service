package service

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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

func TestScriptLogger_LoadExistingLogs(t *testing.T) {
	tests := []struct {
		name     string
		setupLog func(string) error
		expected int
		validate func(*testing.T, []LogEntry)
	}{
		{
			name: "load JSON format entries",
			setupLog: func(logPath string) error {
				entry1 := LogEntry{
					Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					ScriptName: "test",
					ExitCode:   0,
					Stdout:     "First execution output",
					Stderr:     "",
					Duration:   1500,
				}

				entry2 := LogEntry{
					Timestamp:  time.Date(2024, 1, 15, 10, 31, 0, 0, time.UTC),
					ScriptName: "test",
					ExitCode:   1,
					Stdout:     "Second execution output",
					Stderr:     "Error message",
					Duration:   2000,
				}

				file, err := os.Create(logPath)
				if err != nil {
					return err
				}
				defer file.Close()

				data1, _ := json.Marshal(entry1)
				data2, _ := json.Marshal(entry2)
				file.Write(append(data1, '\n'))
				file.Write(append(data2, '\n'))
				return nil
			},
			expected: 2,
			validate: func(t *testing.T, entries []LogEntry) {
				if entries[0].ExitCode != 0 {
					t.Errorf("Expected first entry exit code 0, got %d", entries[0].ExitCode)
				}
				if entries[0].Stdout != "First execution output" {
					t.Errorf("Expected first entry stdout 'First execution output', got '%s'", entries[0].Stdout)
				}
				if entries[1].ExitCode != 1 {
					t.Errorf("Expected second entry exit code 1, got %d", entries[1].ExitCode)
				}
				if entries[1].Stderr != "Error message" {
					t.Errorf("Expected second entry stderr 'Error message', got '%s'", entries[1].Stderr)
				}
			},
		},
		{
			name: "handle empty log file",
			setupLog: func(logPath string) error {
				return os.WriteFile(logPath, []byte(""), 0600)
			},
			expected: 0,
			validate: func(t *testing.T, entries []LogEntry) {
				// No validation needed for empty entries
			},
		},
		{
			name: "handle malformed JSON lines",
			setupLog: func(logPath string) error {
				validEntry := LogEntry{
					Timestamp:  time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
					ScriptName: "test",
					ExitCode:   0,
					Stdout:     "Valid entry",
					Stderr:     "",
					Duration:   1000,
				}

				file, err := os.Create(logPath)
				if err != nil {
					return err
				}
				defer file.Close()

				// Write a valid entry
				data, _ := json.Marshal(validEntry)
				file.Write(append(data, '\n'))

				// Write malformed JSON
				file.WriteString("invalid json line\n")

				// Write another valid entry
				validEntry.Stdout = "Second valid entry"
				data, _ = json.Marshal(validEntry)
				file.Write(append(data, '\n'))

				return nil
			},
			expected: 2,
			validate: func(t *testing.T, entries []LogEntry) {
				if entries[0].Stdout != "Valid entry" {
					t.Errorf("Expected first valid entry, got '%s'", entries[0].Stdout)
				}
				if entries[1].Stdout != "Second valid entry" {
					t.Errorf("Expected second valid entry, got '%s'", entries[1].Stdout)
				}
			},
		},
		{
			name: "respect maxLines limit when loading",
			setupLog: func(logPath string) error {
				file, err := os.Create(logPath)
				if err != nil {
					return err
				}
				defer file.Close()

				// Write 5 entries
				for i := 0; i < 5; i++ {
					entry := LogEntry{
						Timestamp:  time.Date(2024, 1, 15, 10, 30+i, 0, 0, time.UTC),
						ScriptName: "test",
						ExitCode:   i,
						Stdout:     fmt.Sprintf("Entry %d", i),
						Stderr:     "",
						Duration:   1000,
					}
					data, _ := json.Marshal(entry)
					file.Write(append(data, '\n'))
				}
				return nil
			},
			expected: 3, // maxLines = 3 in this test
			validate: func(t *testing.T, entries []LogEntry) {
				// Should keep the last 3 entries (2, 3, 4)
				if entries[0].ExitCode != 2 {
					t.Errorf("Expected first kept entry exit code 2, got %d", entries[0].ExitCode)
				}
				if entries[2].ExitCode != 4 {
					t.Errorf("Expected last kept entry exit code 4, got %d", entries[2].ExitCode)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			baseDir := filepath.Join(os.TempDir(), "test_load_logs_"+tt.name)
			defer os.RemoveAll(baseDir)

			// Create test directory
			err := os.MkdirAll(baseDir, 0750)
			if err != nil {
				t.Fatalf("Failed to create test directory: %v", err)
			}

			logPath := filepath.Join(baseDir, "test.log")
			err = tt.setupLog(logPath)
			if err != nil {
				t.Fatalf("Failed to setup log file: %v", err)
			}

			// Use maxLines = 3 for the limit test, otherwise 100
			maxLines := 100
			if tt.name == "respect maxLines limit when loading" {
				maxLines = 3
			}

			// Create new logger and test loading
			logger := NewScriptLogger("test", baseDir, maxLines)

			// Verify entries were loaded
			entries := logger.GetEntries()
			if len(entries) != tt.expected {
				t.Fatalf("Expected %d loaded entries, got %d", tt.expected, len(entries))
			}

			// Run custom validation
			tt.validate(t, entries)
		})
	}
}

func TestLogManager_QueryLogs(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_query_logs")
	defer os.RemoveAll(baseDir)

	lm := NewLogManager(baseDir)

	// Add some test entries to multiple script loggers
	script1Logger := lm.GetLogger("script1")
	script2Logger := lm.GetLogger("script2")

	// Add entries to script1
	entry1 := &LogEntry{
		Timestamp:  time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC),
		ScriptName: "script1",
		ExitCode:   0,
		Stdout:     "Success output",
		Stderr:     "",
		Duration:   1000,
	}

	entry2 := &LogEntry{
		Timestamp:  time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
		ScriptName: "script1",
		ExitCode:   1,
		Stdout:     "Error output",
		Stderr:     "Error message",
		Duration:   2000,
	}

	script1Logger.AddEntry(entry1)
	script1Logger.AddEntry(entry2)

	// Add entries to script2
	entry3 := &LogEntry{
		Timestamp:  time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		ScriptName: "script2",
		ExitCode:   0,
		Stdout:     "Script2 success",
		Stderr:     "",
		Duration:   1500,
	}

	script2Logger.AddEntry(entry3)

	tests := []struct {
		name     string
		query    LogQuery
		expected int
		validate func(*testing.T, []LogEntry)
	}{
		{
			name:     "query all scripts",
			query:    LogQuery{},
			expected: 3,
			validate: func(t *testing.T, results []LogEntry) {
				// Should return entries from both scripts
				scriptNames := make(map[string]bool)
				for _, entry := range results {
					scriptNames[entry.ScriptName] = true
				}
				if !scriptNames["script1"] || !scriptNames["script2"] {
					t.Error("Expected entries from both script1 and script2")
				}
			},
		},
		{
			name:     "query specific script",
			query:    LogQuery{ScriptName: "script1"},
			expected: 2,
			validate: func(t *testing.T, results []LogEntry) {
				for _, entry := range results {
					if entry.ScriptName != "script1" {
						t.Errorf("Expected only script1 entries, got %s", entry.ScriptName)
					}
				}
			},
		},
		{
			name:     "query by exit code",
			query:    LogQuery{ExitCode: &[]int{0}[0]},
			expected: 2,
			validate: func(t *testing.T, results []LogEntry) {
				for _, entry := range results {
					if entry.ExitCode != 0 {
						t.Errorf("Expected only exit code 0 entries, got %d", entry.ExitCode)
					}
				}
			},
		},
		{
			name: "query by time range",
			query: LogQuery{
				StartTime: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
				EndTime:   time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC),
			},
			expected: 2,
			validate: func(t *testing.T, results []LogEntry) {
				for _, entry := range results {
					if entry.Timestamp.Before(time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)) {
						t.Errorf("Entry timestamp %v is before start time", entry.Timestamp)
					}
					if entry.Timestamp.After(time.Date(2024, 1, 15, 12, 30, 0, 0, time.UTC)) {
						t.Errorf("Entry timestamp %v is after end time", entry.Timestamp)
					}
				}
			},
		},
		{
			name:     "query with limit",
			query:    LogQuery{Limit: 2},
			expected: 2,
			validate: func(t *testing.T, results []LogEntry) {
				// Should return last 2 entries chronologically
				if len(results) > 2 {
					t.Errorf("Expected max 2 results, got %d", len(results))
				}
			},
		},
		{
			name:     "query nonexistent script",
			query:    LogQuery{ScriptName: "nonexistent"},
			expected: 0,
			validate: func(t *testing.T, results []LogEntry) {
				// No validation needed for empty results
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := lm.QueryLogs(&tt.query)
			if err != nil {
				t.Fatalf("QueryLogs failed: %v", err)
			}

			if len(results) != tt.expected {
				t.Fatalf("Expected %d results, got %d", tt.expected, len(results))
			}

			tt.validate(t, results)
		})
	}
}

func TestLogManager_ClearLogs(t *testing.T) {
	baseDir := filepath.Join(os.TempDir(), "test_clear_logs")
	defer os.RemoveAll(baseDir)

	lm := NewLogManager(baseDir)
	logger := lm.GetLogger("test_script")

	// Add some entries
	entry := &LogEntry{
		Timestamp:  time.Now(),
		ScriptName: "test_script",
		ExitCode:   0,
		Stdout:     "Test output",
		Stderr:     "",
		Duration:   1000,
	}

	logger.AddEntry(entry)
	logger.AddEntry(entry)

	// Verify entries exist
	if len(logger.GetEntries()) != 2 {
		t.Fatalf("Expected 2 entries before clearing, got %d", len(logger.GetEntries()))
	}

	// Clear logs
	err := lm.ClearLogs("test_script")
	if err != nil {
		t.Fatalf("Failed to clear logs: %v", err)
	}

	// Verify entries are cleared
	if len(logger.GetEntries()) != 0 {
		t.Errorf("Expected 0 entries after clearing, got %d", len(logger.GetEntries()))
	}

	// Test clearing nonexistent script
	err = lm.ClearLogs("nonexistent")
	if err == nil {
		t.Error("Expected error when clearing logs for nonexistent script")
	}
}

// 🔴 Red Phase: Write failing test for parseLegacyFormat() method (0% coverage)
func TestScriptLogger_parseLegacyFormat(t *testing.T) {
	t.Run("should parse legacy format with single entry", func(t *testing.T) {
		// Create legacy format log content
		legacyContent := `[2025-08-02 11:26:16] Exit code: 0
STDOUT: Hello World
STDOUT: This is multi-line
output from script
--------------------------------------------------`

		// Create ScriptLogger
		logger := &ScriptLogger{
			scriptName: "test-script",
			entries:    make([]LogEntry, 0),
		}

		// Create scanner from the content
		scanner := bufio.NewScanner(strings.NewReader(legacyContent))
		scanner.Scan()
		firstLine := strings.TrimSpace(scanner.Text())

		// This test should initially fail because parseLegacyFormat() has 0% coverage
		logger.parseLegacyFormat(scanner, firstLine)

		// Verify the parsed entry
		entries := logger.GetEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(entries))
		}

		if len(entries) > 0 {
			entry := entries[0]
			if entry.ScriptName != "test-script" {
				t.Errorf("Expected script name 'test-script', got '%s'", entry.ScriptName)
			}
			if entry.ExitCode != 0 {
				t.Errorf("Expected exit code 0, got %d", entry.ExitCode)
			}
			expectedStdout := "Hello World\nThis is multi-line\noutput from script"
			if entry.Stdout != expectedStdout {
				t.Errorf("Expected stdout '%s', got '%s'", expectedStdout, entry.Stdout)
			}
			// Check timestamp parsing
			expectedTime := time.Date(2025, 8, 2, 11, 26, 16, 0, time.UTC)
			if !entry.Timestamp.Equal(expectedTime) {
				t.Errorf("Expected timestamp %v, got %v", expectedTime, entry.Timestamp)
			}
		}
	})

	t.Run("should parse legacy format with multiple entries", func(t *testing.T) {
		// Create legacy format log content with multiple entries
		legacyContent := `[2025-08-02 11:26:16] Exit code: 0
STDOUT: First execution
--------------------------------------------------
[2025-08-02 12:30:45] Exit code: 1
STDOUT: Second execution failed
STDOUT: Error message
--------------------------------------------------`

		// Create ScriptLogger
		logger := &ScriptLogger{
			scriptName: "test-script",
			entries:    make([]LogEntry, 0),
		}

		// Create scanner from the content
		scanner := bufio.NewScanner(strings.NewReader(legacyContent))
		scanner.Scan()
		firstLine := strings.TrimSpace(scanner.Text())

		logger.parseLegacyFormat(scanner, firstLine)

		// Verify the parsed entries
		entries := logger.GetEntries()
		if len(entries) != 2 {
			t.Errorf("Expected 2 entries, got %d", len(entries))
		}

		if len(entries) >= 2 {
			// First entry
			entry1 := entries[0]
			if entry1.ExitCode != 0 {
				t.Errorf("Expected first entry exit code 0, got %d", entry1.ExitCode)
			}
			if entry1.Stdout != "First execution" {
				t.Errorf("Expected first entry stdout 'First execution', got '%s'", entry1.Stdout)
			}

			// Second entry
			entry2 := entries[1]
			if entry2.ExitCode != 1 {
				t.Errorf("Expected second entry exit code 1, got %d", entry2.ExitCode)
			}
			expectedStdout2 := "Second execution failed\nError message"
			if entry2.Stdout != expectedStdout2 {
				t.Errorf("Expected second entry stdout '%s', got '%s'", expectedStdout2, entry2.Stdout)
			}
		}
	})

	t.Run("should handle legacy format without final separator", func(t *testing.T) {
		// Create legacy format log content without final separator
		legacyContent := `[2025-08-02 11:26:16] Exit code: 0
STDOUT: Output without separator`

		// Create ScriptLogger
		logger := &ScriptLogger{
			scriptName: "test-script",
			entries:    make([]LogEntry, 0),
		}

		// Create scanner from the content
		scanner := bufio.NewScanner(strings.NewReader(legacyContent))
		scanner.Scan()
		firstLine := strings.TrimSpace(scanner.Text())

		logger.parseLegacyFormat(scanner, firstLine)

		// Verify the parsed entry (should handle missing final separator)
		entries := logger.GetEntries()
		if len(entries) != 1 {
			t.Errorf("Expected 1 entry, got %d", len(entries))
		}

		if len(entries) > 0 {
			entry := entries[0]
			if entry.Stdout != "Output without separator" {
				t.Errorf("Expected stdout 'Output without separator', got '%s'", entry.Stdout)
			}
		}
	})
}
