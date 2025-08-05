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

// setupVueFrontendStructure creates the Vue.js frontend directory structure
func setupVueFrontendStructure(frontendDir string) error {
	// Create main frontend directory
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		return fmt.Errorf("failed to create frontend directory: %w", err)
	}

	// Create subdirectories
	dirs := []string{
		"src",
		"src/components",
		"src/views",
		"src/composables",
		"src/services",
	}

	for _, dir := range dirs {
		dirPath := filepath.Join(frontendDir, dir)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dirPath, err)
		}
	}

	// Create package.json
	if err := createPackageJSON(frontendDir); err != nil {
		return fmt.Errorf("failed to create package.json: %w", err)
	}

	// Create vite.config.js
	if err := createViteConfig(frontendDir); err != nil {
		return fmt.Errorf("failed to create vite.config.js: %w", err)
	}

	// Create index.html
	if err := createIndexHTML(frontendDir); err != nil {
		return fmt.Errorf("failed to create index.html: %w", err)
	}

	// Create main.js
	if err := createMainJS(frontendDir); err != nil {
		return fmt.Errorf("failed to create main.js: %w", err)
	}

	// Create App.vue
	if err := createAppVue(frontendDir); err != nil {
		return fmt.Errorf("failed to create App.vue: %w", err)
	}

	return nil
}

// createPackageJSON creates the package.json file with Vue.js dependencies
func createPackageJSON(frontendDir string) error {
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

	packagePath := filepath.Join(frontendDir, "package.json")
	return os.WriteFile(packagePath, data, 0644)
}

// createViteConfig creates the vite.config.js file
func createViteConfig(frontendDir string) error {
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
	viteConfigPath := filepath.Join(frontendDir, "vite.config.js")
	return os.WriteFile(viteConfigPath, []byte(viteConfig), 0644)
}

// createIndexHTML creates the index.html file
func createIndexHTML(frontendDir string) error {
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
	indexPath := filepath.Join(frontendDir, "index.html")
	return os.WriteFile(indexPath, []byte(indexHTML), 0644)
}

// createMainJS creates the main.js file
func createMainJS(frontendDir string) error {
	mainJS := `import { createApp } from 'vue'
import App from './App.vue'

createApp(App).mount('#app')
`
	mainJSPath := filepath.Join(frontendDir, "src", "main.js")
	return os.WriteFile(mainJSPath, []byte(mainJS), 0644)
}

// createAppVue creates the App.vue file
func createAppVue(frontendDir string) error {
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
	appVuePath := filepath.Join(frontendDir, "src", "App.vue")
	return os.WriteFile(appVuePath, []byte(appVue), 0644)
}
