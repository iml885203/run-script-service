package web

import (
	"os"
	"path/filepath"
	"testing"
)

// TestVueFrontendDirectoryStructure tests that Vue frontend directory structure can be set up
func TestVueFrontendDirectoryStructure(t *testing.T) {
	// Create temporary directory for testing
	tempDir := t.TempDir()
	frontendDir := filepath.Join(tempDir, "frontend")

	// Test that we can create the Vue frontend directory structure
	err := setupVueFrontendStructure(frontendDir)
	if err != nil {
		t.Fatalf("Failed to setup Vue frontend structure: %v", err)
	}

	// Verify expected directories exist
	expectedDirs := []string{
		"src",
		"src/components",
		"src/views",
		"src/composables",
		"src/services",
	}

	for _, dir := range expectedDirs {
		dirPath := filepath.Join(frontendDir, dir)
		if _, err := os.Stat(dirPath); os.IsNotExist(err) {
			t.Errorf("Expected directory %s does not exist", dirPath)
		}
	}

	// Verify expected files exist
	expectedFiles := []string{
		"package.json",
		"vite.config.js",
		"index.html",
		"src/main.js",
		"src/App.vue",
	}

	for _, file := range expectedFiles {
		filePath := filepath.Join(frontendDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			t.Errorf("Expected file %s does not exist", filePath)
		}
	}
}

// TestVuePackageJsonStructure tests that package.json has correct Vue.js dependencies
func TestVuePackageJsonStructure(t *testing.T) {
	tempDir := t.TempDir()
	frontendDir := filepath.Join(tempDir, "frontend")

	err := setupVueFrontendStructure(frontendDir)
	if err != nil {
		t.Fatalf("Failed to setup Vue frontend structure: %v", err)
	}

	// Test package.json content
	packagePath := filepath.Join(frontendDir, "package.json")
	content, err := os.ReadFile(packagePath)
	if err != nil {
		t.Fatalf("Failed to read package.json: %v", err)
	}

	packageStr := string(content)

	// Check for required Vue.js dependencies
	expectedDeps := []string{
		"\"vue\":",
		"\"vue-router\":",
		"\"@vueuse/core\":",
		"\"vite\":",
		"\"@vitejs/plugin-vue\":",
	}

	for _, dep := range expectedDeps {
		if !containsString(packageStr, dep) {
			t.Errorf("package.json missing expected dependency: %s", dep)
		}
	}
}

// Helper function to check if string contains substring
func containsString(str, substr string) bool {
	return len(str) >= len(substr) &&
		(len(substr) == 0 || findSubstring(str, substr))
}

func findSubstring(str, substr string) bool {
	for i := 0; i <= len(str)-len(substr); i++ {
		if str[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
