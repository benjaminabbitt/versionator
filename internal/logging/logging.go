package logging

import (
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger    *zap.Logger
	loggerMu  sync.RWMutex
	verbosity int
)

// VerbosityToLevel converts a verbosity count to a zap log level
// 0 = Warn (default - only warnings and errors)
// 1 = Info
// 2+ = Debug
func VerbosityToLevel(v int) zapcore.Level {
	switch {
	case v >= 2:
		return zapcore.DebugLevel
	case v == 1:
		return zapcore.InfoLevel
	default:
		return zapcore.WarnLevel
	}
}

// InitLoggerWithVerbosity initializes the global logger with format and verbosity
// This function is thread-safe and can be called multiple times
func InitLoggerWithVerbosity(outputFormat string, v int) error {
	loggerMu.Lock()
	defer loggerMu.Unlock()

	verbosity = v
	level := VerbosityToLevel(v)

	var config zap.Config

	switch outputFormat {
	case "json":
		config = zap.NewProductionConfig()
	case "development":
		config = zap.NewDevelopmentConfig()
	case "console":
		fallthrough
	default:
		config = zap.NewProductionConfig()
		config.Encoding = "console"
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	config.Level = zap.NewAtomicLevelAt(level)

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// InitLogger initializes the global logger with the specified output format
// Uses default verbosity (0 = Warn level)
// This function is thread-safe and can be called multiple times
func InitLogger(outputFormat string) error {
	return InitLoggerWithVerbosity(outputFormat, 0)
}

// GetVerbosity returns the current verbosity level
func GetVerbosity() int {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	return verbosity
}

// ResetVerbosity resets the verbosity level to 0 (for testing)
func ResetVerbosity() {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	verbosity = 0
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

	// Initialize default production logger
	logger, _ = zap.NewProduction()
	return logger
}
