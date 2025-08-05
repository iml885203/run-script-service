package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileManager_New(t *testing.T) {
	fm := NewFileManager("/test/base")

	if fm == nil {
		t.Fatal("NewFileManager should not return nil")
	}

	if fm.baseDir != "/test/base" {
		t.Errorf("Expected baseDir '/test/base', got '%s'", fm.baseDir)
	}

	if len(fm.allowedPaths) == 0 {
		t.Error("FileManager should have default allowed paths")
	}

	if len(fm.deniedPaths) == 0 {
		t.Error("FileManager should have default denied paths")
	}
}

func TestFileManager_IsPathAllowed(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fm := NewFileManager(tempDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{
			name:     "current directory allowed",
			path:     ".",
			expected: true,
		},
		{
			name:     "scripts directory allowed",
			path:     "./scripts",
			expected: true,
		},
		{
			name:     "scripts subdirectory allowed",
			path:     "./scripts/backup.sh",
			expected: true,
		},
		{
			name:     "logs directory allowed",
			path:     "./logs",
			expected: true,
		},
		{
			name:     "system directory denied",
			path:     "/etc/passwd",
			expected: false,
		},
		{
			name:     "user directory denied",
			path:     "/home/user/.ssh/id_rsa",
			expected: false,
		},
		{
			name:     "root directory denied",
			path:     "/root/.bashrc",
			expected: false,
		},
		{
			name:     "arbitrary system path denied",
			path:     "/usr/bin/bash",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := fm.IsPathAllowed(tt.path)
			if result != tt.expected {
				t.Errorf("IsPathAllowed(%s) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFileManager_ReadFile(t *testing.T) {
	// Create temporary directory and test file
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fm := NewFileManager(tempDir)

	// Create test file
	testContent := "#!/bin/bash\necho 'Hello World'\n"
	testFile := filepath.Join(tempDir, "test.sh")
	err = os.WriteFile(testFile, []byte(testContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("read allowed file", func(t *testing.T) {
		content, err := fm.ReadFile("test.sh")
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if content.Content != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, content.Content)
		}

		if content.Path != "test.sh" {
			t.Errorf("Expected path 'test.sh', got '%s'", content.Path)
		}

		if content.Size != int64(len(testContent)) {
			t.Errorf("Expected size %d, got %d", len(testContent), content.Size)
		}
	})

	t.Run("read denied file", func(t *testing.T) {
		_, err := fm.ReadFile("/etc/passwd")
		if err == nil {
			t.Error("Expected error when reading denied file")
		}

		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("Expected 'access denied' error, got: %v", err)
		}
	})

	t.Run("read non-existent file", func(t *testing.T) {
		_, err := fm.ReadFile("nonexistent.sh")
		if err == nil {
			t.Error("Expected error when reading non-existent file")
		}
	})
}

func TestFileManager_WriteFile(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fm := NewFileManager(tempDir)

	t.Run("write to allowed path", func(t *testing.T) {
		testContent := "#!/bin/bash\necho 'Test script'\n"
		err := fm.WriteFile("test_write.sh", testContent)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}

		// Verify file was written correctly
		content, err := os.ReadFile(filepath.Join(tempDir, "test_write.sh"))
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
		}
	})

	t.Run("write to denied path", func(t *testing.T) {
		err := fm.WriteFile("/etc/test.conf", "test content")
		if err == nil {
			t.Error("Expected error when writing to denied path")
		}

		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("Expected 'access denied' error, got: %v", err)
		}
	})

	t.Run("write to subdirectory", func(t *testing.T) {
		testContent := "test log content\n"
		err := fm.WriteFile("logs/test.log", testContent)
		if err != nil {
			t.Fatalf("Failed to write to subdirectory: %v", err)
		}

		// Verify file exists and has correct content
		fullPath := filepath.Join(tempDir, "logs", "test.log")
		content, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read written file: %v", err)
		}

		if string(content) != testContent {
			t.Errorf("Expected content '%s', got '%s'", testContent, string(content))
		}
	})
}

func TestFileManager_ValidateScriptSyntax(t *testing.T) {
	fm := NewFileManager("/test")

	tests := []struct {
		name           string
		content        string
		expectedIssues int
		shouldContain  string
	}{
		{
			name: "valid script",
			content: `#!/bin/bash
echo "Hello World"
ls -la`,
			expectedIssues: 0,
		},
		{
			name: "script with rm -rf",
			content: `#!/bin/bash
rm -rf /tmp/*`,
			expectedIssues: 1,
			shouldContain:  "dangerous command",
		},
		{
			name: "script with sudo",
			content: `#!/bin/bash
sudo apt update`,
			expectedIssues: 1,
			shouldContain:  "sudo",
		},
		{
			name: "script with unmatched quotes",
			content: `#!/bin/bash
echo "Hello World
echo 'Test'`,
			expectedIssues: 1,
			shouldContain:  "Unmatched",
		},
		{
			name: "script with multiple issues",
			content: `#!/bin/bash
echo "Unmatched quote
sudo rm -rf /tmp/*`,
			expectedIssues: 3,
		},
		{
			name: "script with comments only",
			content: `#!/bin/bash
# This is a comment
# Another comment`,
			expectedIssues: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := fm.ValidateScriptSyntax(tt.content)

			if len(issues) != tt.expectedIssues {
				t.Errorf("Expected %d issues, got %d: %v", tt.expectedIssues, len(issues), issues)
			}

			if tt.shouldContain != "" && len(issues) > 0 {
				found := false
				for _, issue := range issues {
					if strings.Contains(strings.ToLower(issue), strings.ToLower(tt.shouldContain)) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected issue containing '%s', got: %v", tt.shouldContain, issues)
				}
			}
		})
	}
}

func TestFileManager_ListFiles(t *testing.T) {
	// Create temporary directory with test files
	tempDir, err := os.MkdirTemp("", "file_manager_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	fm := NewFileManager(tempDir)

	// Create test files
	testFiles := []string{"test1.sh", "test2.sh", "config.json"}
	for _, filename := range testFiles {
		err := os.WriteFile(filepath.Join(tempDir, filename), []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", filename, err)
		}
	}

	t.Run("list files in allowed directory", func(t *testing.T) {
		files, err := fm.ListFiles(".")
		if err != nil {
			t.Fatalf("Failed to list files: %v", err)
		}

		if len(files) != len(testFiles) {
			t.Errorf("Expected %d files, got %d", len(testFiles), len(files))
		}

		// Check that all test files are present
		fileNames := make(map[string]bool)
		for _, file := range files {
			fileNames[file.Name()] = true
		}

		for _, expectedFile := range testFiles {
			if !fileNames[expectedFile] {
				t.Errorf("Expected file '%s' not found in listing", expectedFile)
			}
		}
	})

	t.Run("list files in denied directory", func(t *testing.T) {
		_, err := fm.ListFiles("/etc")
		if err == nil {
			t.Error("Expected error when listing denied directory")
		}

		if !strings.Contains(err.Error(), "access denied") {
			t.Errorf("Expected 'access denied' error, got: %v", err)
		}
	})

	t.Run("list files in non-existent directory", func(t *testing.T) {
		_, err := fm.ListFiles("nonexistent")
		if err == nil {
			t.Error("Expected error when listing non-existent directory")
		}
	})
}
