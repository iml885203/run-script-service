package service

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
)

type Config struct {
	Interval int `json:"interval"`
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

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	return nil
}
