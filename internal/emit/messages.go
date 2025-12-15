// Package emit messages - error and log message constants
// Exported so tests can compare against them
package emit

import "os"

// File permission constants
const (
	// FilePermission is the default permission for created files (owner rw, group/other r)
	FilePermission os.FileMode = 0644
)

// Error messages
const (
	ErrUnsupportedFormat     = "unsupported format"
	ErrTemplateRenderFail    = "failed to render template"
	ErrFileWriteFail         = "failed to write file"
	ErrOutputPathEmpty       = "output path cannot be empty"
	ErrOutputPathIsDirectory = "is a directory, not a file"
	ErrParentDirNotExist     = "does not exist"
	ErrParentNotDirectory    = "is not a directory"
)

// Log messages for structured logging
const (
	LogTemplateRendered = "template_rendered"
	LogTemplateWritten  = "template_written"
	LogEmitCompleted    = "emit_completed"
)
