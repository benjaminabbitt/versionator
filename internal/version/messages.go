// Package version messages - error and log message constants
// Exported so tests can compare against them
package version

import "os"

// File permission constants
const (
	// FilePermission is the default permission for created files (owner rw, group/other r)
	FilePermission os.FileMode = 0644
)

// Error messages
const (
	ErrCannotDecrementMajor = "cannot decrement major version below 0"
	ErrCannotDecrementMinor = "cannot decrement minor version below 0"
	ErrCannotDecrementPatch = "cannot decrement patch version below 0"
	ErrInvalidVersionLevel  = "invalid version level"
	ErrCustomKeyNotFound    = "custom key not found"
)

// Log messages for structured logging
const (
	LogVersionLoaded      = "version_loaded"
	LogVersionSaved       = "version_saved"
	LogVersionIncremented = "version_incremented"
	LogVersionDecremented = "version_decremented"
	LogVersionMigrated    = "version_migrated"
	LogVersionParsed      = "version_parsed"
	LogVersionParseError  = "version_parse_error"
	LogVersionCreated     = "version_created"
	LogCustomVarSet       = "custom_var_set"
	LogCustomVarDeleted   = "custom_var_deleted"
	LogVersionSet         = "version_set"
	LogPrefixSet          = "prefix_set"
	LogFileReadError      = "file_read_error"
	LogFileWriteError     = "file_write_error"
)
