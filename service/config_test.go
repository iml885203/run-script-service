package service

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfig_LoadConfig(t *testing.T) {
	tests := []struct {
		name             string
		configContent    string
		expectedInterval int
		expectError      bool
	}{
		{
			name:             "valid config",
			configContent:    `{"interval": 1800}`,
			expectedInterval: 1800,
			expectError:      false,
		},
		{
			name:             "default config when file doesn't exist",
			configContent:    "",
			expectedInterval: 3600,
			expectError:      false,
		},
		{
			name:             "invalid json",
			configContent:    `{invalid json}`,
			expectedInterval: 3600, // should keep default
			expectError:      false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory
			tempDir, err := os.MkdirTemp("", "config_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			configPath := filepath.Join(tempDir, "test_config.json")

			// Create config file if content provided
			if tt.configContent != "" {
				if writeErr := os.WriteFile(configPath, []byte(tt.configContent), 0644); writeErr != nil {
					t.Fatal(writeErr)
				}
			}

			// Test loading config
			config := &Config{Interval: 3600} // default
			err = LoadConfig(configPath, config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if config.Interval != tt.expectedInterval {
				t.Errorf("expected interval %d, got %d", tt.expectedInterval, config.Interval)
			}
		})
	}
}

func TestConfig_SaveConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "config_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "test_config.json")
	config := &Config{Interval: 2400}

	// Test saving config
	err = SaveConfig(configPath, config)
	if err != nil {
		t.Errorf("unexpected error saving config: %v", err)
	}

	// Verify file was created and has correct content
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Errorf("error reading saved config: %v", err)
	}

	expectedContent := "{\n  \"interval\": 2400\n}"
	if string(data) != expectedContent {
		t.Errorf("expected config content %q, got %q", expectedContent, string(data))
	}
}

func TestServiceConfig_LoadMultiScriptConfig(t *testing.T) {
	tests := []struct {
		name            string
		configContent   string
		expectedScripts int
		expectedWebPort int
		expectError     bool
	}{
		{
			name: "valid multi-script config",
			configContent: `{
				"scripts": [
					{
						"name": "main",
						"path": "./run.sh",
						"interval": 3600,
						"enabled": true,
						"max_log_lines": 100,
						"timeout": 300
					},
					{
						"name": "backup",
						"path": "./backup.sh",
						"interval": 86400,
						"enabled": true,
						"max_log_lines": 50,
						"timeout": 1800
					}
				],
				"web_port": 8080
			}`,
			expectedScripts: 2,
			expectedWebPort: 8080,
			expectError:     false,
		},
		{
			name: "empty scripts array",
			configContent: `{
				"scripts": [],
				"web_port": 9090
			}`,
			expectedScripts: 0,
			expectedWebPort: 9090,
			expectError:     false,
		},
		{
			name: "backward compatibility - old config format",
			configContent: `{
				"interval": 1800
			}`,
			expectedScripts: 1,    // should create default script
			expectedWebPort: 8080, // default
			expectError:     false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			tempDir, err := os.MkdirTemp("", "config_test")
			if err != nil {
				t.Fatal(err)
			}
			defer os.RemoveAll(tempDir)

			configPath := filepath.Join(tempDir, "test_config.json")
			err = os.WriteFile(configPath, []byte(tt.configContent), 0644)
			if err != nil {
				t.Fatal(err)
			}

			config := &ServiceConfig{WebPort: 8080} // default
			err = LoadServiceConfig(configPath, config)

			if tt.expectError && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if len(config.Scripts) != tt.expectedScripts {
				t.Errorf("expected %d scripts, got %d", tt.expectedScripts, len(config.Scripts))
			}

			if config.WebPort != tt.expectedWebPort {
				t.Errorf("expected web port %d, got %d", tt.expectedWebPort, config.WebPort)
			}
		})
	}
}

func TestScriptConfig_Validation(t *testing.T) {
	tests := []struct {
		name        string
		script      ScriptConfig
		expectValid bool
	}{
		{
			name: "valid script config",
			script: ScriptConfig{
				Name:        "test",
				Path:        "./test.sh",
				Interval:    3600,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     300,
			},
			expectValid: true,
		},
		{
			name: "empty name should be invalid",
			script: ScriptConfig{
				Name:        "",
				Path:        "./test.sh",
				Interval:    3600,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     300,
			},
			expectValid: false,
		},
		{
			name: "empty path should be invalid",
			script: ScriptConfig{
				Name:        "test",
				Path:        "",
				Interval:    3600,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     300,
			},
			expectValid: false,
		},
		{
			name: "negative interval should be invalid",
			script: ScriptConfig{
				Name:        "test",
				Path:        "./test.sh",
				Interval:    -1,
				Enabled:     true,
				MaxLogLines: 100,
				Timeout:     300,
			},
			expectValid: false,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			// Use ValidateWithOptions(false) to skip file existence check in tests
			err := tt.script.ValidateWithOptions(false)
			isValid := err == nil

			if isValid != tt.expectValid {
				t.Errorf("expected valid=%v, got valid=%v, error=%v", tt.expectValid, isValid, err)
			}
		})
	}
}

func TestEnhancedConfig_LoadWithEnvFile(t *testing.T) {
	tempDir := t.TempDir()

	// Create .env file
	envFile := filepath.Join(tempDir, ".env")
	envContent := `WEB_SECRET_KEY=test-secret-from-env
LOG_LEVEL=debug
WEB_PORT=9090
`
	err := os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Create service config file
	configFile := filepath.Join(tempDir, "service_config.json")
	configContent := `{
  "scripts": [
    {
      "name": "test-script",
      "path": "./test.sh",
      "interval": 300,
      "enabled": true,
      "max_log_lines": 100,
      "timeout": 0
    }
  ],
  "web_port": 8080
}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	// Test enhanced configuration loading
	enhancedConfig := NewEnhancedConfig()
	err = enhancedConfig.LoadWithEnv(configFile, envFile)
	if err != nil {
		t.Fatalf("LoadWithEnv failed: %v", err)
	}

	// Test that service config was loaded
	if len(enhancedConfig.Config.Scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(enhancedConfig.Config.Scripts))
	}

	if enhancedConfig.Config.Scripts[0].Name != "test-script" {
		t.Errorf("Expected script name 'test-script', got %s", enhancedConfig.Config.Scripts[0].Name)
	}

	// Test that env values are accessible
	if secret := enhancedConfig.GetEnv("WEB_SECRET_KEY"); secret != "test-secret-from-env" {
		t.Errorf("Expected WEB_SECRET_KEY='test-secret-from-env', got '%s'", secret)
	}

	if logLevel := enhancedConfig.GetEnv("LOG_LEVEL"); logLevel != "debug" {
		t.Errorf("Expected LOG_LEVEL='debug', got '%s'", logLevel)
	}
}

func TestEnhancedConfig_GetWebPort(t *testing.T) {
	tempDir := t.TempDir()

	// Create .env file with WEB_PORT
	envFile := filepath.Join(tempDir, ".env")
	envContent := `WEB_PORT=9090`
	err := os.WriteFile(envFile, []byte(envContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create .env file: %v", err)
	}

	// Create service config with different port
	configFile := filepath.Join(tempDir, "service_config.json")
	configContent := `{"scripts": [], "web_port": 8080}`
	err = os.WriteFile(configFile, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create config file: %v", err)
	}

	enhancedConfig := NewEnhancedConfig()
	err = enhancedConfig.LoadWithEnv(configFile, envFile)
	if err != nil {
		t.Fatalf("LoadWithEnv failed: %v", err)
	}

	// Environment variable should take priority over JSON config
	webPort := enhancedConfig.GetWebPort()
	if webPort != 9090 {
		t.Errorf("Expected web port 9090 (from env), got %d", webPort)
	}
}
