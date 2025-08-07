package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScriptFileManager_CreateScript(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	content := "#!/bin/bash\necho 'test'"

	err := manager.CreateScript("test.sh", content)
	assert.NoError(t, err)

	// Verify file was created
	filePath := filepath.Join(tmpDir, "scripts", "test.sh")
	assert.FileExists(t, filePath)

	// Verify file content
	fileContent, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(fileContent))

	// Verify file permissions (executable)
	fileInfo, err := os.Stat(filePath)
	assert.NoError(t, err)
	assert.True(t, fileInfo.Mode()&0111 != 0, "Script should be executable")
}

func TestScriptFileManager_CreateScript_InvalidExtension(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	err := manager.CreateScript("test.txt", "echo 'test'")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "must end with .sh extension")
}

func TestScriptFileManager_CreateScript_InvalidFilename(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	err := manager.CreateScript("invalid/script.sh", "echo 'test'")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid filename format")
}

func TestScriptFileManager_CreateScript_AlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	content := "#!/bin/bash\necho 'test'"

	// Create first time
	err := manager.CreateScript("test.sh", content)
	assert.NoError(t, err)

	// Try to create again
	err = manager.CreateScript("test.sh", content)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestScriptFileManager_GetScript(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	// Create test script first
	content := "#!/bin/bash\necho 'hello world'"
	err := manager.CreateScript("test.sh", content)
	require.NoError(t, err)

	// Get the script
	script, err := manager.GetScript("test.sh")
	assert.NoError(t, err)
	assert.NotNil(t, script)
	assert.Equal(t, "test.sh", script.Filename)
	assert.Equal(t, content, script.Content)
	assert.True(t, strings.HasSuffix(script.Path, "scripts/test.sh"))
	assert.Greater(t, script.Size, int64(0))
	assert.True(t, script.Modified.After(time.Now().Add(-time.Minute)))
}

func TestScriptFileManager_GetScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	script, err := manager.GetScript("nonexistent.sh")
	assert.Error(t, err)
	assert.Nil(t, script)
	assert.Contains(t, err.Error(), "not found")
}

func TestScriptFileManager_UpdateScript(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	// Create test script first
	originalContent := "#!/bin/bash\necho 'original'"
	err := manager.CreateScript("test.sh", originalContent)
	require.NoError(t, err)

	// Update the script
	newContent := "#!/bin/bash\necho 'updated'"
	err = manager.UpdateScript("test.sh", newContent)
	assert.NoError(t, err)

	// Verify content was updated
	script, err := manager.GetScript("test.sh")
	assert.NoError(t, err)
	assert.Equal(t, newContent, script.Content)
}

func TestScriptFileManager_UpdateScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	err := manager.UpdateScript("nonexistent.sh", "echo 'test'")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestScriptFileManager_ListScripts(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	// Create multiple test scripts
	scripts := []string{"backup.sh", "deploy.sh", "monitor.sh"}
	for _, scriptName := range scripts {
		err := manager.CreateScript(scriptName, "#!/bin/bash\necho '"+scriptName+"'")
		require.NoError(t, err)
	}

	// List scripts
	scriptList, err := manager.ListScripts()
	assert.NoError(t, err)
	assert.Len(t, scriptList, 3)

	// Verify all scripts are in the list
	scriptNames := make([]string, len(scriptList))
	for i, script := range scriptList {
		scriptNames[i] = script.Filename
		assert.Greater(t, script.Size, int64(0))
		assert.True(t, strings.HasSuffix(script.Path, script.Filename))
		assert.True(t, script.Modified.After(time.Now().Add(-time.Minute)))
		// Content should not be loaded in list view
		assert.Empty(t, script.Content)
	}

	for _, expectedName := range scripts {
		assert.Contains(t, scriptNames, expectedName)
	}
}

func TestScriptFileManager_ListScripts_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	scriptList, err := manager.ListScripts()
	assert.NoError(t, err)
	assert.Len(t, scriptList, 0)
}

func TestScriptFileManager_DeleteScript(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	// Create test script
	err := manager.CreateScript("test.sh", "#!/bin/bash\necho 'test'")
	require.NoError(t, err)

	// Verify it exists
	filePath := filepath.Join(tmpDir, "scripts", "test.sh")
	assert.FileExists(t, filePath)

	// Delete the script
	err = manager.DeleteScript("test.sh")
	assert.NoError(t, err)

	// Verify it's deleted
	assert.NoFileExists(t, filePath)
}

func TestScriptFileManager_DeleteScript_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	err := manager.DeleteScript("nonexistent.sh")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to delete")
}

func TestScriptFileManager_GetScriptPath(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewScriptFileManager(tmpDir)

	path := manager.GetScriptPath("test.sh")
	expectedPath := filepath.Join(tmpDir, "scripts", "test.sh")
	assert.Equal(t, expectedPath, path)
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		filename string
		valid    bool
	}{
		{"valid-script.sh", true},
		{"valid_script.sh", true},
		{"valid.123.sh", true},
		{"VaLiD-Script_123.sh", true},
		{"invalid/script.sh", false},
		{"invalid<script.sh", false},
		{"invalid>script.sh", false},
		{"invalid|script.sh", false},
		{"invalid?script.sh", false},
		{"invalid*script.sh", false},
		{"invalid:script.sh", false},
		{"invalid\"script.sh", false},
		{"", false},
		{strings.Repeat("a", 300) + ".sh", false}, // Too long
	}

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			result := isValidFilename(test.filename)
			assert.Equal(t, test.valid, result, "filename: %s", test.filename)
		})
	}
}

func TestScriptFileManager_DirectoryCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't pre-create scripts directory
	scriptsDir := filepath.Join(tmpDir, "scripts")
	assert.NoDirExists(t, scriptsDir)

	// Create manager - should create directory
	manager := NewScriptFileManager(tmpDir)
	assert.NotNil(t, manager)

	// Verify directory was created
	assert.DirExists(t, scriptsDir)

	// Verify directory permissions
	fileInfo, err := os.Stat(scriptsDir)
	assert.NoError(t, err)
	assert.True(t, fileInfo.IsDir())
	assert.Equal(t, os.FileMode(0755), fileInfo.Mode().Perm())
}
