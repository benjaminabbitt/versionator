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
)

// Log messages for structured logging
const (
	LogFileUpdated      = "file_updated"
	LogUpdateSkipped    = "update_skipped"
	LogUpdatesApplied   = "updates_applied"
	LogValidatingConfig = "validating_update_config"
)
