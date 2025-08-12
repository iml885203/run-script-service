package main

import (
	"os"
	"path/filepath"
	"testing"

	"run-script-service/service"
	"run-script-service/web"
)

// TestMainWebServerInitializationProcess documents the current vs desired initialization
func TestMainWebServerInitializationProcess(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Create test configuration
	configPath := filepath.Join(tempDir, "service_config.json")
	testConfig := service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
		WebPort: 8085,
	}

	err := service.SaveServiceConfig(configPath, &testConfig)
	if err != nil {
		t.Fatalf("Failed to create config: %v", err)
	}

	// Change directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test what runMultiScriptServiceWithWeb currently initializes
	fileManager := service.NewFileManager(tempDir)
	scriptManager := service.NewScriptManagerWithPath(&testConfig, configPath)
	systemMonitor := service.NewSystemMonitor()

	// Verify these components are created successfully
	if fileManager == nil {
		t.Error("File manager should be initialized")
	}
	if scriptManager == nil {
		t.Error("Script manager should be initialized")
	}
	if systemMonitor == nil {
		t.Error("System monitor should be initialized")
	}

	// THE MISSING COMPONENT: Script file manager should also be initialized
	// This is what needs to be added to main.go:
	scriptFileManager := service.NewScriptFileManager(tempDir)
	if scriptFileManager == nil {
		t.Error("Script file manager should be initialized")
	}

	// Verify scripts directory is created by script file manager
	scriptsDir := filepath.Join(tempDir, "scripts")
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		t.Error("Scripts directory should be created by script file manager initialization")
	}
}

// TestWebServerSetScriptFileManager tests that the web server can properly set the script file manager
func TestWebServerSetScriptFileManager(t *testing.T) {
	// Create temporary directory
	tempDir := t.TempDir()

	// Initialize all components as they should be in main.go
	scriptFileManager := service.NewScriptFileManager(tempDir)
	fileManager := service.NewFileManager(tempDir)
	systemMonitor := service.NewSystemMonitor()

	// Create minimal config for script manager
	config := &service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
		WebPort: 8086,
	}
	scriptManager := service.NewScriptManagerWithPath(config, "")

	// Create web server and set all components
	webServer := web.NewWebServer(nil, 8086, "test-secret")
	webServer.SetScriptManager(scriptManager)
	webServer.SetFileManager(fileManager)
	webServer.SetSystemMonitor(systemMonitor)
	webServer.SetScriptFileManager(scriptFileManager)

	// Test that script file manager works by creating a script
	testFilename := "webserver-test.sh"
	testContent := "#!/bin/bash\necho 'Web server test'"

	err := scriptFileManager.CreateScript(testFilename, testContent)
	if err != nil {
		t.Errorf("Failed to create script through script file manager: %v", err)
	}

	// Verify script was created
	scriptPath := filepath.Join(tempDir, "scripts", testFilename)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Error("Script file should have been created")
	}

	// Test listing scripts
	scripts, err := scriptFileManager.ListScripts()
	if err != nil {
		t.Errorf("Failed to list scripts: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}

	if len(scripts) > 0 && scripts[0].Filename != testFilename {
		t.Errorf("Expected filename %s, got %s", testFilename, scripts[0].Filename)
	}
}
