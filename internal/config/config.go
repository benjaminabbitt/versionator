package config

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

const configFile = ".versionator.yaml"

// Config holds configuration for version suffix behavior
type Config struct {
	Prefix  string        `yaml:"prefix"`
	Suffix  SuffixConfig  `yaml:"suffix"`
	Logging LoggingConfig `yaml:"logging"`
}

// SuffixConfig holds suffix-specific configuration
type SuffixConfig struct {
	Type    string    `yaml:"type"`
	Enabled bool      `yaml:"enabled"`
	Git     GitConfig `yaml:"git"`
}

// GitConfig holds git-specific configuration
type GitConfig struct {
	HashLength int `yaml:"hashLength"`
}

// LoggingConfig holds logging-specific configuration
type LoggingConfig struct {
	Output string `yaml:"output"` // console, json, development
}

// ReadConfig reads the configuration from .versionator.yaml file
func ReadConfig() (*Config, error) {
	config := &Config{
		Prefix: "v", // default prefix
		Suffix: SuffixConfig{
			Type:    "git",
			Enabled: false, // default to disabled
			Git: GitConfig{
				HashLength: 7, // default hash length
			},
		},
		Logging: LoggingConfig{
			Output: "console", // default to human-readable console output
		},
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// Config file doesn't exist, return default config
			return config, nil
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	err = yaml.Unmarshal(data, config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

func WriteConfig(config *Config) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Add a comment header
	content := "# Versionator Configuration\n" + string(data)
	return os.WriteFile(configFile, []byte(content), 0644)
}
