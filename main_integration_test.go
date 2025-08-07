package main

import (
	"os"
	"path/filepath"
	"testing"

	"run-script-service/service"
)

// TestScriptFileManagerInitialization tests that the script file manager is properly initialized
// when running the service with web interface
func TestScriptFileManagerInitialization(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create test configuration file
	configPath := filepath.Join(tempDir, "service_config.json")
	testConfig := service.ServiceConfig{
		Scripts: []service.ScriptConfig{},
		WebPort: 8081, // Use different port to avoid conflicts
	}

	// Save configuration
	err := service.SaveServiceConfig(configPath, &testConfig)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create .env file for test
	envPath := filepath.Join(tempDir, ".env")
	envContent := "WEB_SECRET_KEY=test-secret-key\nWEB_PORT=8081\n"
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Change to temp directory to simulate running from there
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Test the enhanced config loading (this is what runMultiScriptServiceWithWeb uses)
	enhancedConfig, err := loadEnhancedConfig(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to load enhanced config: %v", err)
	}

	// Verify config was loaded correctly
	if enhancedConfig.Config.WebPort != 8081 {
		t.Errorf("Expected web port 8081, got %d", enhancedConfig.Config.WebPort)
	}

	// Verify that the script file manager would be initialized
	// (We simulate what runMultiScriptServiceWithWeb does)
	scriptsDir := filepath.Join(tempDir, "scripts")
	scriptFileManager := service.NewScriptFileManager(tempDir)

	// Verify scripts directory was created
	if _, err := os.Stat(scriptsDir); os.IsNotExist(err) {
		t.Errorf("Scripts directory was not created at %s", scriptsDir)
	}

	// Test script file manager functionality
	testScript := "test-script.sh"
	testContent := "#!/bin/bash\necho 'Test script'"

	// Create a script file
	err = scriptFileManager.CreateScript(testScript, testContent)
	if err != nil {
		t.Errorf("Failed to create script: %v", err)
	}

	// Verify script was created in the scripts directory
	scriptPath := filepath.Join(scriptsDir, testScript)
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		t.Errorf("Script file was not created at %s", scriptPath)
	}

	// Verify we can retrieve the script
	scriptFile, err := scriptFileManager.GetScript(testScript)
	if err != nil {
		t.Errorf("Failed to get script: %v", err)
	}

	if scriptFile.Content != testContent {
		t.Errorf("Expected script content %q, got %q", testContent, scriptFile.Content)
	}

	// Test listing scripts
	scripts, err := scriptFileManager.ListScripts()
	if err != nil {
		t.Errorf("Failed to list scripts: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}

	if scripts[0].Filename != testScript {
		t.Errorf("Expected script filename %q, got %q", testScript, scripts[0].Filename)
	}
}

// TestMainWebServerIntegration tests that the web server integration works with script file manager
func TestMainWebServerIntegration(t *testing.T) {
	// Create temporary directory for test
	tempDir := t.TempDir()

	// Create test configuration file with a test script
	configPath := filepath.Join(tempDir, "service_config.json")
	testConfig := service.ServiceConfig{
		Scripts: []service.ScriptConfig{
			{
				Name:        "test-script",
				Path:        "./scripts/test.sh",
				Interval:    60,
				Enabled:     false, // Keep disabled for test
				MaxLogLines: 50,
				Timeout:     0,
			},
		},
		WebPort: 8082, // Use different port
	}

	// Save configuration
	err := service.SaveServiceConfig(configPath, &testConfig)
	if err != nil {
		t.Fatalf("Failed to create test config: %v", err)
	}

	// Create .env file for test
	envPath := filepath.Join(tempDir, ".env")
	envContent := "WEB_SECRET_KEY=test-secret-key-2\nWEB_PORT=8082\n"
	err = os.WriteFile(envPath, []byte(envContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Change to temp directory
	originalWd, _ := os.Getwd()
	defer os.Chdir(originalWd)
	os.Chdir(tempDir)

	// Create a minimal scripts directory and script file for the test
	scriptsDir := filepath.Join(tempDir, "scripts")
	err = os.MkdirAll(scriptsDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create scripts directory: %v", err)
	}

	scriptPath := filepath.Join(scriptsDir, "test.sh")
	scriptContent := "#!/bin/bash\necho 'Integration test script'"
	err = os.WriteFile(scriptPath, []byte(scriptContent), 0755)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Test configuration components that runMultiScriptServiceWithWeb would use
	enhancedConfig, err := loadEnhancedConfig(configPath, envPath)
	if err != nil {
		t.Fatalf("Failed to load enhanced config: %v", err)
	}

	// Test script file manager initialization
	scriptFileManager := service.NewScriptFileManager(tempDir)

	// Verify the script file manager can find the existing script
	existingScripts, err := scriptFileManager.ListScripts()
	if err != nil {
		t.Errorf("Failed to list existing scripts: %v", err)
	}

	if len(existingScripts) != 1 {
		t.Errorf("Expected 1 existing script, got %d", len(existingScripts))
	}

	if len(existingScripts) > 0 && existingScripts[0].Filename != "test.sh" {
		t.Errorf("Expected script filename 'test.sh', got %q", existingScripts[0].Filename)
	}

	// Test script manager initialization with the config
	scriptManager := service.NewScriptManagerWithPath(&enhancedConfig.Config, configPath)

	// Verify script manager has the expected scripts
	scripts, err := scriptManager.GetScripts()
	if err != nil {
		t.Errorf("Failed to get scripts from script manager: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script in script manager, got %d", len(scripts))
	}

	if len(scripts) > 0 && scripts[0].Name != "test-script" {
		t.Errorf("Expected script name 'test-script', got %q", scripts[0].Name)
	}
}
