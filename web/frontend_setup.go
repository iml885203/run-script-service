package web

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// VuePackageJSON represents the structure of package.json for Vue.js project
type VuePackageJSON struct {
	Name            string            `json:"name"`
	Version         string            `json:"version"`
	Private         bool              `json:"private"`
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

// VueFrontendStructure defines the Vue.js frontend project structure
type VueFrontendStructure struct {
	BaseDir string
}

// NewVueFrontendStructure creates a new Vue frontend structure helper
func NewVueFrontendStructure(baseDir string) *VueFrontendStructure {
	return &VueFrontendStructure{
		BaseDir: baseDir,
	}
}

// setupVueFrontendStructure creates the Vue.js frontend directory structure
func setupVueFrontendStructure(frontendDir string) error {
	structure := NewVueFrontendStructure(frontendDir)
	return structure.Create()
}

// Create builds the complete Vue.js frontend structure
func (vfs *VueFrontendStructure) Create() error {
	// Create directory structure
	if err := vfs.createDirectories(); err != nil {
		return err
	}

	// Create configuration files
	if err := vfs.createConfigurationFiles(); err != nil {
		return err
	}

	// Create source files
	if err := vfs.createSourceFiles(); err != nil {
		return err
	}

	return nil
}

// createDirectories creates the Vue.js directory structure
func (vfs *VueFrontendStructure) createDirectories() error {
	// Create main frontend directory
	if err := os.MkdirAll(vfs.BaseDir, 0755); err != nil {
		return fmt.Errorf("failed to create frontend directory: %w", err)
	}

	// Define Vue.js standard directory structure
	dirs := []string{
		"src",
		"src/components",
		"src/views",
		"src/composables",
		"src/services",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(vfs.BaseDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	return nil
}

// createConfigurationFiles creates configuration files for Vue.js project
func (vfs *VueFrontendStructure) createConfigurationFiles() error {
	configFiles := []func() error{
		vfs.createPackageJSON,
		vfs.createViteConfig,
		vfs.createIndexHTML,
	}

	for _, createFile := range configFiles {
		if err := createFile(); err != nil {
			return err
		}
	}

	return nil
}

// createSourceFiles creates source code files for Vue.js project
func (vfs *VueFrontendStructure) createSourceFiles() error {
	sourceFiles := []func() error{
		vfs.createMainJS,
		vfs.createAppVue,
	}

	for _, createFile := range sourceFiles {
		if err := createFile(); err != nil {
			return err
		}
	}

	return nil
}

// createPackageJSON creates the package.json file with Vue.js dependencies
func (vfs *VueFrontendStructure) createPackageJSON() error {
	packageJSON := VuePackageJSON{
		Name:    "run-script-service-frontend",
		Version: "1.0.0",
		Private: true,
		Scripts: map[string]string{
			"serve": "vite",
			"build": "vite build",
			"dev":   "vite",
		},
		Dependencies: map[string]string{
			"vue":          "^3.4.0",
			"vue-router":   "^4.2.0",
			"@vueuse/core": "^10.7.0",
		},
		DevDependencies: map[string]string{
			"vite":               "^5.0.0",
			"@vitejs/plugin-vue": "^4.5.0",
			"sass":               "^1.69.0",
		},
	}

	data, err := json.MarshalIndent(packageJSON, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal package.json: %w", err)
	}

	packagePath := filepath.Join(vfs.BaseDir, "package.json")
	return os.WriteFile(packagePath, data, 0644)
}

// createViteConfig creates the vite.config.js file
func (vfs *VueFrontendStructure) createViteConfig() error {
	viteConfig := `import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'

export default defineConfig({
  plugins: [vue()],
  base: '/static/',
  build: {
    outDir: '../static',
    assetsInlineLimit: 8192,
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['vue', 'vue-router']
        }
      }
    }
  }
})
`
	viteConfigPath := filepath.Join(vfs.BaseDir, "vite.config.js")
	return os.WriteFile(viteConfigPath, []byte(viteConfig), 0644)
}

// createIndexHTML creates the index.html file
func (vfs *VueFrontendStructure) createIndexHTML() error {
	indexHTML := `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>Run Script Service</title>
</head>
<body>
  <div id="app"></div>
  <script type="module" src="/src/main.js"></script>
</body>
</html>
`
	indexPath := filepath.Join(vfs.BaseDir, "index.html")
	return os.WriteFile(indexPath, []byte(indexHTML), 0644)
}

// createMainJS creates the main.js file
func (vfs *VueFrontendStructure) createMainJS() error {
	mainJS := `import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')
`
	mainJSPath := filepath.Join(vfs.BaseDir, "src", "main.js")
	return os.WriteFile(mainJSPath, []byte(mainJS), 0644)
}

// createAppVue creates the App.vue file
func (vfs *VueFrontendStructure) createAppVue() error {
	appVue := `<template>
  <div id="app">
    <h1>Run Script Service</h1>
    <p>Vue.js frontend is loading...</p>
  </div>
</template>

<script>
export default {
  name: 'App'
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  text-align: center;
  color: #2c3e50;
  margin-top: 60px;
}
</style>
`
	appVuePath := filepath.Join(vfs.BaseDir, "src", "App.vue")
	return os.WriteFile(appVuePath, []byte(appVue), 0644)
}
