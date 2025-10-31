
package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents the CLI configuration
type Config struct {
	APIKey string `json:"api_key"`
}

// GetConfigDir returns the path to the .skilzy config directory
func GetConfigDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(home, ".skilzy"), nil
}

// GetConfigPath returns the full path to the config file
func GetConfigPath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "config.json"), nil
}

// SaveAPIKey saves the API key to the config file
func SaveAPIKey(apiKey string) error {
	configDir, err := GetConfigDir()
	if err != nil {
		return err
	}

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDir, 0700); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	config := Config{APIKey: apiKey}
	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	configPath, err := GetConfigPath()
	if err != nil {
		return err
	}

	// Write config file with restricted permissions (0600 = read/write for owner only)
	if err := os.WriteFile(configPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// LoadAPIKey loads the API key from the config file
func LoadAPIKey() (string, error) {
	configPath, err := GetConfigPath()
	if err != nil {
		return "", err
	}

	// Check if config file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", nil // No config file, return empty string
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return "", fmt.Errorf("failed to parse config file: %w", err)
	}

	return config.APIKey, nil
}
