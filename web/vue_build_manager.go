package web

import (
	"fmt"
	"os"
	"path/filepath"
)

// VueBuildManager handles the Vue.js frontend build process
type VueBuildManager struct {
	ProjectRoot string
	FrontendDir string
	StaticDir   string
}

// NewVueBuildManager creates a new Vue build manager
func NewVueBuildManager(projectRoot string) *VueBuildManager {
	return &VueBuildManager{
		ProjectRoot: projectRoot,
		FrontendDir: filepath.Join(projectRoot, "web", "frontend"),
		StaticDir:   filepath.Join(projectRoot, "web", "frontend", "dist"),
	}
}

// InitializeFrontendProject sets up the complete Vue.js frontend project structure
func (vbm *VueBuildManager) InitializeFrontendProject() error {
	// Create main frontend directory
	if err := os.MkdirAll(vbm.FrontendDir, 0755); err != nil {
		return fmt.Errorf("failed to create frontend directory: %w", err)
	}

	// Initialize TypeScript Vue project (new approach)
	if err := vbm.initializeTypeScriptProject(); err != nil {
		return fmt.Errorf("failed to initialize TypeScript project: %w", err)
	}

	return nil
}

// initializeTypeScriptProject sets up a complete TypeScript + Vue.js project
func (vbm *VueBuildManager) initializeTypeScriptProject() error {
	// For now, we assume the TypeScript frontend files are already in place
	// This method validates that all required TypeScript files exist

	// Check if TypeScript files exist (they should have been created by external process)
	requiredFiles := []string{
		"package.json",
		"tsconfig.json",
		"vite.config.ts",
		"vitest.config.ts",
		"index.html",
		"src/main.ts",
		"src/App.vue",
		"src/router/index.ts",
		"src/types/api.ts",
		"src/services/api.ts",
		"src/composables/useScripts.ts",
		"src/composables/useLogs.ts",
		"src/composables/useSystemMetrics.ts",
		"src/composables/useWebSocket.ts",
		"src/views/Dashboard.vue",
		"src/style.css",
		"tests/setup.ts",
		"tests/unit/services/api.test.ts",
		"tests/unit/composables/useScripts.test.ts",
		"tests/unit/components/Dashboard.test.ts",
	}

	missingFiles := []string{}
	for _, file := range requiredFiles {
		filePath := filepath.Join(vbm.FrontendDir, file)
		if _, err := os.Stat(filePath); os.IsNotExist(err) {
			missingFiles = append(missingFiles, file)
		}
	}

	if len(missingFiles) > 0 {
		return fmt.Errorf("missing required TypeScript files: %v", missingFiles)
	}

	return nil
}

// GetFrontendDir returns the frontend directory path
func (vbm *VueBuildManager) GetFrontendDir() string {
	return vbm.FrontendDir
}

// GetStaticDir returns the static directory path
func (vbm *VueBuildManager) GetStaticDir() string {
	return vbm.StaticDir
}

// CheckFrontendExists checks if the frontend directory structure exists
func (vbm *VueBuildManager) CheckFrontendExists() bool {
	requiredPaths := []string{
		filepath.Join(vbm.FrontendDir, "package.json"),
		filepath.Join(vbm.FrontendDir, "src", "main.ts"),
		filepath.Join(vbm.FrontendDir, "src", "App.vue"),
		filepath.Join(vbm.FrontendDir, "tsconfig.json"),
		filepath.Join(vbm.FrontendDir, "vite.config.ts"),
	}

	for _, path := range requiredPaths {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	return true
}
