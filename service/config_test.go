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
				err := os.WriteFile(configPath, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatal(err)
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
