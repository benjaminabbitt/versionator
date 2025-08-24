package logging

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

// InitLogger initializes the global logger with the specified output format
func InitLogger(outputFormat string) error {
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

	var err error
	logger, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

// GetSugaredLogger returns a sugared logger instance for the application to use
func GetSugaredLogger() *zap.SugaredLogger {
	if logger == nil {
		// Fallback to a basic logger if not initialized
		logger, _ = zap.NewProduction()
	}
	return logger.Sugar()
}

