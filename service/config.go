// Package service provides core functionality for the run-script-service daemon.
package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// ScriptConfig represents configuration for a single script
type ScriptConfig struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Interval    int    `json:"interval"` // seconds
	Enabled     bool   `json:"enabled"`
	MaxLogLines int    `json:"max_log_lines"`
	Timeout     int    `json:"timeout"` // seconds, 0 means no limit
}

// UpdateResponse represents the detailed response for script updates
type UpdateResponse struct {
	Success       bool               `json:"success"`
	Message       string             `json:"message"`
	Applied       bool               `json:"applied"`
	Scheduled     bool               `json:"scheduled"`
	Changes       []ConfigChangeInfo `json:"changes"`
	NextExecution *string            `json:"next_execution,omitempty"`
}

// ConfigChangeInfo represents information about a specific configuration change
type ConfigChangeInfo struct {
	Field    string      `json:"field"`
	OldValue interface{} `json:"old_value"`
	NewValue interface{} `json:"new_value"`
	Applied  bool        `json:"applied"`
	Reason   string      `json:"reason,omitempty"`
}

// ConfigUpdateEvent represents a configuration update event for WebSocket broadcasting
type ConfigUpdateEvent struct {
	Type       string             `json:"type"` // "config_update"
	ScriptName string             `json:"script_name"`
	Status     string             `json:"status"` // "applied", "scheduled", "failed"
	Changes    []ConfigChangeInfo `json:"changes"`
	Applied    bool               `json:"applied"`
	Scheduled  bool               `json:"scheduled"`
	Message    string             `json:"message"`
	Timestamp  string             `json:"timestamp"`
}

// ServiceConfig represents the overall service configuration
type ServiceConfig struct {
	Scripts []ScriptConfig `json:"scripts"`
	WebPort int            `json:"web_port"`
}

// Config is a legacy struct for backward compatibility
type Config struct {
	Interval int `json:"interval"`
}

// Validate checks if the script configuration is valid
func (sc *ScriptConfig) Validate() error {
	return sc.ValidateWithOptions(true)
}

// ValidateWithOptions checks if the script configuration is valid with optional file existence check
func (sc *ScriptConfig) ValidateWithOptions(checkFileExists bool) error {
	if sc.Name == "" {
		return fmt.Errorf("script name cannot be empty")
	}
	if sc.Path == "" {
		return fmt.Errorf("script path cannot be empty")
	}
	if sc.Interval < 0 {
		return fmt.Errorf("interval cannot be negative")
	}
	if sc.MaxLogLines < 0 {
		return fmt.Errorf("max_log_lines cannot be negative")
	}
	if sc.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}

	// Optionally check if script file exists and is executable
	if checkFileExists {
		scriptPath := sc.Path
		if !filepath.IsAbs(scriptPath) {
			// Convert relative path to absolute path
			workDir, err := os.Getwd()
			if err != nil {
				return fmt.Errorf("unable to get working directory: %v", err)
			}
			scriptPath = filepath.Join(workDir, sc.Path)
		}

		info, err := os.Stat(scriptPath)
		if err != nil {
			if os.IsNotExist(err) {
				return fmt.Errorf("script file does not exist: %s", sc.Path)
			}
			return fmt.Errorf("unable to access script file %s: %v", sc.Path, err)
		}

		// Check if it's a regular file (not a directory)
		if info.IsDir() {
			return fmt.Errorf("script path is a directory, not a file: %s", sc.Path)
		}

		// Check if file is executable
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("script file is not executable: %s (mode: %v)", sc.Path, info.Mode())
		}
	}

	return nil
}

// LoadConfig loads configuration from the specified file path
func LoadConfig(configPath string, config *Config) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Keep default config
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return nil // Keep default config, don't fail
	}

	if err := json.Unmarshal(data, config); err != nil {
		log.Printf("Error parsing config: %v", err)
		return nil // Keep default config, don't fail
	}

	return nil
}

// SaveConfig saves configuration to the specified file path
func SaveConfig(configPath string, config *Config) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	return nil
}

// LoadServiceConfig loads the new multi-script configuration with backward compatibility
func LoadServiceConfig(configPath string, config *ServiceConfig) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil // Keep default config
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Printf("Error reading config: %v", err)
		return nil // Keep default config, don't fail
	}

	// Try to parse as new format first
	var tempConfig ServiceConfig
	if err := json.Unmarshal(data, &tempConfig); err == nil {
		// Check if it looks like new format (has "scripts" field or "web_port" field)
		var rawConfig map[string]interface{}
		if err := json.Unmarshal(data, &rawConfig); err != nil {
			log.Printf("Error parsing config as map: %v", err)
		}

		if _, hasScripts := rawConfig["scripts"]; hasScripts || rawConfig["web_port"] != nil {
			// Successfully parsed as new format
			for i, script := range tempConfig.Scripts {
				// Only validate basic fields during config loading, not file existence
				if err := script.ValidateWithOptions(false); err != nil {
					log.Printf("Invalid script config %d: %v", i, err)
					return nil // Keep default config
				}
			}
			*config = tempConfig
			return nil
		}
	}

	// Try to parse as legacy format for backward compatibility
	var legacyConfig Config
	if err := json.Unmarshal(data, &legacyConfig); err != nil {
		log.Printf("Error parsing config: %v", err)
		return nil // Keep default config, don't fail
	}

	// Convert legacy config to new format
	config.Scripts = []ScriptConfig{
		{
			Name:        "main",
			Path:        "./run.sh", // default script path
			Interval:    legacyConfig.Interval,
			Enabled:     true,
			MaxLogLines: 100, // default
			Timeout:     0,   // no timeout
		},
	}
	if config.WebPort == 0 {
		config.WebPort = 8080 // default
	}

	return nil
}

// SaveServiceConfig saves the service configuration to file
func SaveServiceConfig(configPath string, config *ServiceConfig) error {
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %v", err)
	}

	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	return nil
}
