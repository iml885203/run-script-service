// Package service provides file management functionality
package service

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileManager handles secure file operations
type FileManager struct {
	allowedPaths []string
	deniedPaths  []string
	baseDir      string
}

// FileContent represents file content with metadata
type FileContent struct {
	Path    string `json:"path"`
	Content string `json:"content"`
	Size    int64  `json:"size"`
	Mode    string `json:"mode"`
}

// NewFileManager creates a new file manager with security constraints
func NewFileManager(baseDir string) *FileManager {
	return &FileManager{
		baseDir: baseDir,
		allowedPaths: []string{
			".",          // Current directory
			"./scripts",  // Scripts directory
			"./logs",     // Logs directory
			"./testdata", // Test data directory
		},
		deniedPaths: []string{
			"/etc",
			"/usr",
			"/bin",
			"/sbin",
			"/root",
			"/home",
			"/var",
			"/tmp",
			"/proc",
			"/sys",
		},
	}
}

// IsPathAllowed checks if a file path is allowed for access
func (fm *FileManager) IsPathAllowed(path string) bool {
	// Clean and resolve the path
	cleanPath := filepath.Clean(path)

	// Additional security checks
	if strings.Contains(cleanPath, "..") {
		return false // Path traversal attempt
	}

	// Convert to absolute path for security checks
	var absPath string
	if filepath.IsAbs(cleanPath) {
		absPath = cleanPath

		// Check denied paths first for absolute paths
		for _, denied := range fm.deniedPaths {
			if strings.HasPrefix(absPath, denied) {
				return false
			}
		}

		// Absolute paths outside allowed system paths are denied
		return false
	} else {
		// Relative paths are relative to baseDir
		absPath = filepath.Join(fm.baseDir, cleanPath)
	}

	// Ensure the resolved path is still within baseDir
	absBaseDir, err := filepath.Abs(fm.baseDir)
	if err != nil {
		return false
	}

	absRequestPath, err := filepath.Abs(absPath)
	if err != nil {
		return false
	}

	if !strings.HasPrefix(absRequestPath, absBaseDir) {
		return false // Outside base directory
	}

	// For relative paths, check if they're within allowed directories
	for _, allowed := range fm.allowedPaths {
		allowedAbs := allowed
		if !filepath.IsAbs(allowed) {
			allowedAbs = filepath.Join(fm.baseDir, allowed)
		}

		// Normalize paths for comparison
		allowedAbs = filepath.Clean(allowedAbs)
		cleanAbsPath := filepath.Clean(absPath)

		// Allow exact match
		if cleanAbsPath == allowedAbs {
			return true
		}

		// Allow subdirectory (ensure it's actually a subdirectory, not just prefix match)
		if strings.HasPrefix(cleanAbsPath+string(filepath.Separator), allowedAbs+string(filepath.Separator)) {
			return true
		}

		// Allow files within the directory
		if strings.HasPrefix(cleanAbsPath, allowedAbs+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

// ReadFile reads a file's content safely
func (fm *FileManager) ReadFile(path string) (*FileContent, error) {
	if !fm.IsPathAllowed(path) {
		return nil, fmt.Errorf("access denied: path not allowed")
	}

	// Resolve relative path
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(fm.baseDir, path)
	}

	// Get file info
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Read file content
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return &FileContent{
		Path:    path,
		Content: string(content),
		Size:    info.Size(),
		Mode:    info.Mode().String(),
	}, nil
}

// WriteFile writes content to a file safely
func (fm *FileManager) WriteFile(path string, content string) error {
	if !fm.IsPathAllowed(path) {
		return fmt.Errorf("access denied: path not allowed")
	}

	// Resolve relative path
	fullPath := path
	if !filepath.IsAbs(path) {
		fullPath = filepath.Join(fm.baseDir, path)
	}

	// Ensure directory exists
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ValidateScriptSyntax performs basic validation on shell scripts
func (fm *FileManager) ValidateScriptSyntax(content string) []string {
	var issues []string

	lines := strings.Split(content, "\n")
	for i, line := range lines {
		lineNum := i + 1
		trimmed := strings.TrimSpace(line)

		// Skip empty lines and comments
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		// Check for potential security issues
		if strings.Contains(trimmed, "rm -rf") {
			issues = append(issues, fmt.Sprintf("Line %d: Potentially dangerous command 'rm -rf'", lineNum))
		}

		if strings.Contains(trimmed, "sudo") {
			issues = append(issues, fmt.Sprintf("Line %d: Use of 'sudo' detected", lineNum))
		}

		// Check for basic syntax issues
		if strings.Count(trimmed, "'")%2 != 0 {
			issues = append(issues, fmt.Sprintf("Line %d: Unmatched single quote", lineNum))
		}

		if strings.Count(trimmed, "\"")%2 != 0 {
			issues = append(issues, fmt.Sprintf("Line %d: Unmatched double quote", lineNum))
		}
	}

	return issues
}

// ListFiles lists files in a directory
func (fm *FileManager) ListFiles(dirPath string) ([]fs.FileInfo, error) {
	if !fm.IsPathAllowed(dirPath) {
		return nil, fmt.Errorf("access denied: path not allowed")
	}

	// Resolve relative path
	fullPath := dirPath
	if !filepath.IsAbs(dirPath) {
		fullPath = filepath.Join(fm.baseDir, dirPath)
	}

	entries, err := os.ReadDir(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var fileInfos []fs.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue // Skip files we can't get info for
		}
		fileInfos = append(fileInfos, info)
	}

	return fileInfos, nil
}
