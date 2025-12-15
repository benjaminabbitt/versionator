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
	ErrMajorVersionNegative  = "major version cannot be negative"
	ErrMinorVersionNegative  = "minor version cannot be negative"
	ErrPatchVersionNegative  = "patch version cannot be negative"
	ErrCannotDecrementMajor  = "cannot decrement major version below 0"
	ErrCannotDecrementMinor  = "cannot decrement minor version below 0"
	ErrCannotDecrementPatch  = "cannot decrement patch version below 0"
	ErrInvalidVersionLevel   = "invalid version level"
	ErrInvalidPreRelease     = "invalid pre-release identifier"
	ErrInvalidMetadata       = "invalid build metadata"
	ErrEmptyIdentifierPart   = "identifier part cannot be empty"
	ErrInvalidIdentifierChar = "invalid character in identifier"
	ErrCustomKeyNotFound     = "custom key not found"
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
	LogPrefixSet          = "prefix_set"
	LogFileReadError      = "file_read_error"
	LogFileWriteError     = "file_write_error"
)
