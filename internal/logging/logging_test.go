package logging

import (
	"testing"
)

func TestInitLogger_ConsoleFormat(t *testing.T) {
	err := InitLogger("console")
	if err != nil {
		t.Fatalf("Expected no error initializing console logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Test that we can get a sugared logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

func TestInitLogger_JSONFormat(t *testing.T) {
	err := InitLogger("json")
	if err != nil {
		t.Fatalf("Expected no error initializing JSON logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Test that we can get a sugared logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

func TestInitLogger_DevelopmentFormat(t *testing.T) {
	err := InitLogger("development")
	if err != nil {
		t.Fatalf("Expected no error initializing development logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Test that we can get a sugared logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

func TestInitLogger_DefaultFormat(t *testing.T) {
	// Test with empty string (should default to console)
	err := InitLogger("")
	if err != nil {
		t.Fatalf("Expected no error initializing default logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Test that we can get a sugared logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

func TestInitLogger_UnknownFormat(t *testing.T) {
	// Test with unknown format (should default to console)
	err := InitLogger("unknown-format")
	if err != nil {
		t.Fatalf("Expected no error initializing logger with unknown format, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	// Test that we can get a sugared logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

func TestGetSugaredLogger_WithoutInit(t *testing.T) {
	// Reset logger to nil to test fallback behavior
	originalLogger := logger
	logger = nil
	defer func() {
		logger = originalLogger
	}()

	// Should create a fallback logger
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available even without initialization")
	}

	// Verify that logger was created as fallback
	if logger == nil {
		t.Fatal("Expected fallback logger to be created")
	}
}

func TestGetSugaredLogger_MultipleCallsConsistent(t *testing.T) {
	// Initialize logger first
	err := InitLogger("console")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Get sugared logger multiple times
	logger1 := GetSugaredLogger()
	logger2 := GetSugaredLogger()

	if logger1 == nil || logger2 == nil {
		t.Fatal("Expected both sugared loggers to be available")
	}

	// Both should be based on the same underlying logger
	// We can't directly compare them, but we can verify they're both functional
	logger1.Info("Test message 1")
	logger2.Info("Test message 2")
}

func TestInitLogger_Reinitialize(t *testing.T) {
	// Initialize with one format
	err := InitLogger("json")
	if err != nil {
		t.Fatalf("Failed to initialize JSON logger: %v", err)
	}

	firstLogger := logger

	// Reinitialize with different format
	err = InitLogger("development")
	if err != nil {
		t.Fatalf("Failed to reinitialize with development logger: %v", err)
	}

	// Logger should be different instance
	if logger == firstLogger {
		t.Error("Expected logger to be reinitialized with new instance")
	}

	// Should still be functional
	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available after reinitialization")
	}
}

func TestLoggerFormats_AllSupported(t *testing.T) {
	formats := []string{"console", "json", "development", "", "invalid"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			err := InitLogger(format)
			if err != nil {
				t.Fatalf("Failed to initialize logger with format '%s': %v", format, err)
			}

			if logger == nil {
				t.Fatalf("Logger not initialized for format '%s'", format)
			}

			// Test basic functionality
			sugaredLogger := GetSugaredLogger()
			if sugaredLogger == nil {
				t.Fatalf("Sugared logger not available for format '%s'", format)
			}

			// Test that we can log without panicking
			sugaredLogger.Info("Test log message for format: " + format)
		})
	}
}

func TestLogger_ThreadSafety(t *testing.T) {
	// Initialize logger
	err := InitLogger("console")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Test concurrent access to GetSugaredLogger
	done := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()

			sugaredLogger := GetSugaredLogger()
			if sugaredLogger == nil {
				t.Errorf("Goroutine %d: Expected sugared logger to be available", id)
				return
			}

			// Log a message to test functionality
			sugaredLogger.Infof("Test message from goroutine %d", id)
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

// Benchmark tests to ensure logging performance is reasonable
func BenchmarkInitLogger_Console(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InitLogger("console")
	}
}

func BenchmarkInitLogger_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		InitLogger("json")
	}
}

func BenchmarkGetSugaredLogger(b *testing.B) {
	InitLogger("console")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetSugaredLogger()
	}
}
