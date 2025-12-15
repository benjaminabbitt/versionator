// Package config messages - error and log message constants
// Exported so tests can compare against them
package config

import "os"

// File permission constants
const (
	// FilePermission is the default permission for created files (owner rw, group/other r)
	FilePermission os.FileMode = 0644
)

// Error messages
const (
	ErrConfigNotFound       = "config file not found"
	ErrConfigParseFail      = "failed to parse config file"
	ErrInvalidTemplateSyntax = "invalid template syntax"
)

// Log messages for structured logging
const (
	LogConfigLoaded  = "config_loaded"
	LogConfigSaved   = "config_saved"
	LogConfigCreated = "config_created"
)
