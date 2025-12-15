// Package cmd messages - error and log message constants
// Exported so tests can compare against them
package cmd

import "os"

// File permission constants
const (
	// FilePermission is the default permission for created files (owner rw, group/other r)
	FilePermission os.FileMode = 0644
)

// Error messages
const (
	ErrLoadingVersion    = "error loading version"
	ErrCustomKeyNotFound = "custom key not found"
)

// Log messages for structured logging
const (
	LogCommandStarted   = "command_started"
	LogCommandCompleted = "command_completed"
	LogCommandFailed    = "command_failed"
	LogTagCreated       = "tag_created"
	LogEmitCompleted    = "emit_completed"
)
