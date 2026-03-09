package logging

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger   *zap.Logger
	loggerMu sync.RWMutex
)

// InitLogger initializes the global logger with the specified output format
// This function is thread-safe and can be called multiple times
//
// Supported formats:
//   - "quiet" (default): No logging output - suitable for CLI usage
//   - "console": Human-readable colored output
//   - "json": Structured JSON output for log aggregation
//   - "development": Verbose development output with stack traces
func InitLogger(outputFormat string) error {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	var config zap.Config

	switch outputFormat {
	case "quiet", "none", "":
		// No-op logger for CLI - discards all output
		logger = zap.NewNop()
		return nil
	case "json":
		config = zap.NewProductionConfig()
	case "development":
		config = zap.NewDevelopmentConfig()
	case "console":
		config = zap.NewProductionConfig()
		config.Encoding = "console"
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	default:
		// Unknown format, default to quiet
		logger = zap.NewNop()
		return nil
	}

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// GetSugaredLogger returns a sugared logger instance for the application to use
// This function is thread-safe
func GetSugaredLogger() *zap.SugaredLogger {
	return GetLogger().Sugar()
}

// GetLogger returns the raw zap.Logger instance
// This function is thread-safe and initializes a default logger if none exists
func GetLogger() *zap.Logger {
	loggerMu.RLock()
	l := logger
	loggerMu.RUnlock()

	if l != nil {
		return l
	}

	// Need to initialize - acquire write lock
	loggerMu.Lock()
	defer loggerMu.Unlock()

	// Double-check after acquiring write lock
	if logger != nil {
		return logger
	}

	// Initialize default no-op logger (quiet mode for CLI)
	logger = zap.NewNop()
	return logger
}
