package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// ScriptFile represents a script file with metadata
type ScriptFile struct {
	Name     string    `json:"name"`
	Filename string    `json:"filename"`
	Content  string    `json:"content"`
	Path     string    `json:"path"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

// ScriptFileManager manages script files in the dedicated scripts directory
type ScriptFileManager struct {
	scriptsDir string
	mutex      sync.RWMutex
}

// NewScriptFileManager creates a new ScriptFileManager with the specified base directory
func NewScriptFileManager(baseDir string) *ScriptFileManager {
	scriptsDir := filepath.Join(baseDir, "scripts")

	// Ensure scripts directory exists
	if err := os.MkdirAll(scriptsDir, 0755); err != nil {
		// Log warning but don't fail - tests might still want to proceed
		fmt.Printf("Warning: Failed to create scripts directory: %v\n", err)
	}

	return &ScriptFileManager{
		scriptsDir: scriptsDir,
	}
}

// CreateScript creates a new script file with the given content
func (sfm *ScriptFileManager) CreateScript(filename, content string) error {
	sfm.mutex.Lock()
	defer sfm.mutex.Unlock()

	// Validate file extension
	if !strings.HasSuffix(filename, ".sh") {
		return fmt.Errorf("script filename must end with .sh extension")
	}

	// Validate filename format
	if !isValidFilename(filename) {
		return fmt.Errorf("invalid filename format")
	}

	filePath := filepath.Join(sfm.scriptsDir, filename)

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return fmt.Errorf("script file already exists: %s", filename)
	}

	// Create file with executable permissions
	if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to create script file: %v", err)
	}

	return nil
}

// GetScript retrieves a script file with its content and metadata
func (sfm *ScriptFileManager) GetScript(filename string) (*ScriptFile, error) {
	sfm.mutex.RLock()
	defer sfm.mutex.RUnlock()

	filePath := filepath.Join(sfm.scriptsDir, filename)

	// Read file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("script file not found: %s", filename)
		}
		return nil, fmt.Errorf("failed to get script file info: %v", err)
	}

	// Read file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read script file: %v", err)
	}

	return &ScriptFile{
		Filename: filename,
		Content:  string(content),
		Path:     filePath,
		Size:     fileInfo.Size(),
		Modified: fileInfo.ModTime(),
	}, nil
}

// UpdateScript updates the content of an existing script file
func (sfm *ScriptFileManager) UpdateScript(filename, content string) error {
	sfm.mutex.Lock()
	defer sfm.mutex.Unlock()

	filePath := filepath.Join(sfm.scriptsDir, filename)

	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("script file not found: %s", filename)
	}

	// Update file content with executable permissions
	if err := os.WriteFile(filePath, []byte(content), 0755); err != nil {
		return fmt.Errorf("failed to update script file: %v", err)
	}

	return nil
}

// ListScripts returns a list of all script files in the scripts directory
func (sfm *ScriptFileManager) ListScripts() ([]*ScriptFile, error) {
	sfm.mutex.RLock()
	defer sfm.mutex.RUnlock()

	var scripts []*ScriptFile

	err := filepath.WalkDir(sfm.scriptsDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and non-.sh files
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".sh") {
			return nil
		}

		fileInfo, err := d.Info()
		if err != nil {
			return err
		}

		scripts = append(scripts, &ScriptFile{
			Filename: d.Name(),
			Path:     path,
			Size:     fileInfo.Size(),
			Modified: fileInfo.ModTime(),
			// Don't load content for list view - performance optimization
		})

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to list scripts: %v", err)
	}

	return scripts, nil
}

// DeleteScript removes a script file from the scripts directory
func (sfm *ScriptFileManager) DeleteScript(filename string) error {
	sfm.mutex.Lock()
	defer sfm.mutex.Unlock()

	filePath := filepath.Join(sfm.scriptsDir, filename)

	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to delete script file: %v", err)
	}

	return nil
}

// GetScriptPath returns the full path for a script filename
func (sfm *ScriptFileManager) GetScriptPath(filename string) string {
	return filepath.Join(sfm.scriptsDir, filename)
}

// isValidFilename validates that a filename contains only safe characters
func isValidFilename(filename string) bool {
	// Must not be empty and must be reasonable length
	if len(filename) == 0 || len(filename) > 255 {
		return false
	}

	// Define allowed characters: letters, numbers, dots, underscores, hyphens
	validChars := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789._-"

	for _, char := range filename {
		if !strings.ContainsRune(validChars, char) {
			return false
		}
	}

	return true
}
