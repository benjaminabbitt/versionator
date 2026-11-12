package update

import "os"

// FilePermission is the default permission for written files
const FilePermission os.FileMode = 0644

// Error messages
const (
	ErrFileNotFound      = "file not found"
	ErrPathNotFound      = "path not found in file"
	ErrInvalidSelector   = "invalid selector syntax"
	ErrUnsupportedFormat = "unsupported file format"
	ErrFileParseFailed   = "failed to parse file"
	ErrFileWriteFailed   = "failed to write file"
	ErrTemplateRender    = "failed to render template"
	ErrValueNotString    = "value at path is not a string"
	// ErrCargoVersionInvalid is raised before anything is written when a
	// Cargo.toml version would be set to a value Cargo refuses: a 4-part
	// Major.Minor.Patch.Revision, a "v" prefix, or anything else not SemVer 2.0.0.
	ErrCargoVersionInvalid = "not a valid Cargo version (SemVer 2.0.0: Major.Minor.Patch[-pre][+build], no prefix, no 4th part)"
)

// Log messages for structured logging
const (
	LogFileUpdated      = "file_updated"
	LogUpdateSkipped    = "update_skipped"
	LogUpdatesApplied   = "updates_applied"
	LogValidatingConfig = "validating_update_config"
)
