package logging

import (
	"testing"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests validating the primary happy-path behavior of the logging package.
// =============================================================================

// TestInitLogger_ConsoleFormat validates that the console output format
// initializes a functional logger.
//
// Why: Console format is the primary human-readable output mode for interactive
// CLI usage. This test ensures the core initialization path works correctly.
//
// What: Calling InitLogger with "console" format should succeed, create a valid
// logger instance, and provide a usable sugared logger for application logging.
func TestInitLogger_ConsoleFormat(t *testing.T) {
	// Precondition: None - testing fresh initialization

	// Action: Initialize logger with console format
	err := InitLogger("console")

	// Expected: No error, logger and sugared logger both available
	if err != nil {
		t.Fatalf("Expected no error initializing console logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

// TestInitLogger_JSONFormat validates that the JSON output format initializes
// a functional logger for structured logging.
//
// Why: JSON format is critical for production log aggregation systems (ELK,
// Splunk, etc.). This test ensures structured logging works correctly.
//
// What: Calling InitLogger with "json" format should succeed and produce a
// logger capable of outputting structured JSON log entries.
func TestInitLogger_JSONFormat(t *testing.T) {
	// Precondition: None - testing fresh initialization

	// Action: Initialize logger with JSON format
	err := InitLogger("json")

	// Expected: No error, logger and sugared logger both available
	if err != nil {
		t.Fatalf("Expected no error initializing JSON logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

// TestGetSugaredLogger_MultipleCallsConsistent validates that multiple calls
// to GetSugaredLogger return functional loggers based on the same underlying
// logger instance.
//
// Why: Application code may call GetSugaredLogger from many locations. This
// test ensures consistent behavior across the application.
//
// What: After initialization, multiple calls to GetSugaredLogger should all
// return valid, functional sugared loggers.
func TestGetSugaredLogger_MultipleCallsConsistent(t *testing.T) {
	// Precondition: Logger must be initialized
	err := InitLogger("console")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Action: Get sugared logger multiple times
	logger1 := GetSugaredLogger()
	logger2 := GetSugaredLogger()

	// Expected: Both loggers should be available and functional
	if logger1 == nil || logger2 == nil {
		t.Fatal("Expected both sugared loggers to be available")
	}

	// Verify both are functional by logging (no panic expected)
	logger1.Info("Test message 1")
	logger2.Info("Test message 2")
}

// =============================================================================
// KEY VARIATIONS
// Tests validating important alternate flows and configuration options.
// =============================================================================

// TestInitLogger_DevelopmentFormat validates that the development output
// format initializes correctly with verbose debugging features.
//
// Why: Development format provides enhanced debugging with stack traces and
// verbose output. Developers rely on this for local troubleshooting.
//
// What: Calling InitLogger with "development" format should succeed and create
// a logger configured for enhanced developer feedback.
func TestInitLogger_DevelopmentFormat(t *testing.T) {
	// Precondition: None - testing fresh initialization

	// Action: Initialize logger with development format
	err := InitLogger("development")

	// Expected: No error, logger and sugared logger both available
	if err != nil {
		t.Fatalf("Expected no error initializing development logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

// TestInitLogger_Reinitialize validates that the logger can be reinitialized
// with a different format during runtime.
//
// Why: Some applications may need to switch log formats dynamically (e.g.,
// based on configuration reload or user preference).
//
// What: Calling InitLogger twice with different formats should create new
// logger instances each time, with the second call replacing the first.
func TestInitLogger_Reinitialize(t *testing.T) {
	// Precondition: Initialize with one format first
	err := InitLogger("json")
	if err != nil {
		t.Fatalf("Failed to initialize JSON logger: %v", err)
	}
	firstLogger := logger

	// Action: Reinitialize with different format
	err = InitLogger("development")

	// Expected: New logger instance created, old one replaced
	if err != nil {
		t.Fatalf("Failed to reinitialize with development logger: %v", err)
	}

	if logger == firstLogger {
		t.Error("Expected logger to be reinitialized with new instance")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available after reinitialization")
	}
}

// TestLoggerFormats_AllSupported validates that all documented format strings
// result in successful logger initialization.
//
// Why: The API contract promises support for specific format strings. This
// table-driven test ensures all documented formats work correctly.
//
// What: Each format string (console, json, development, empty, invalid) should
// initialize without error and produce a functional logger.
func TestLoggerFormats_AllSupported(t *testing.T) {
	formats := []string{"console", "json", "development", "", "invalid"}

	for _, format := range formats {
		t.Run("format_"+format, func(t *testing.T) {
			// Precondition: None per subtest

			// Action: Initialize logger with the given format
			err := InitLogger(format)

			// Expected: No error, functional logger created
			if err != nil {
				t.Fatalf("Failed to initialize logger with format '%s': %v", format, err)
			}

			if logger == nil {
				t.Fatalf("Logger not initialized for format '%s'", format)
			}

			sugaredLogger := GetSugaredLogger()
			if sugaredLogger == nil {
				t.Fatalf("Sugared logger not available for format '%s'", format)
			}

			// Verify logging does not panic
			sugaredLogger.Info("Test log message for format: " + format)
		})
	}
}

// =============================================================================
// ERROR HANDLING
// Tests validating expected failure modes and recovery behavior.
// =============================================================================

// TestGetSugaredLogger_WithoutInit validates that GetSugaredLogger creates a
// fallback logger when called before InitLogger.
//
// Why: Defensive programming - if application code calls GetSugaredLogger
// before initialization (e.g., during startup race conditions), the system
// should not crash but provide a safe fallback.
//
// What: When logger is nil, GetSugaredLogger should automatically create a
// no-op fallback logger rather than returning nil or panicking.
func TestGetSugaredLogger_WithoutInit(t *testing.T) {
	// Precondition: Reset logger to nil to simulate uninitialized state
	originalLogger := logger
	logger = nil
	defer func() {
		logger = originalLogger
	}()

	// Action: Request sugared logger without prior initialization
	sugaredLogger := GetSugaredLogger()

	// Expected: Fallback logger created automatically
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available even without initialization")
	}

	if logger == nil {
		t.Fatal("Expected fallback logger to be created")
	}
}

// =============================================================================
// EDGE CASES
// Tests validating boundary conditions and unusual inputs.
// =============================================================================

// TestInitLogger_DefaultFormat validates that an empty string defaults to
// the quiet (no-op) logger for CLI usage.
//
// Why: CLI tools often need silent operation by default. An empty format
// string should not cause an error but should produce a no-op logger.
//
// What: Calling InitLogger with "" should succeed and create a functional
// (albeit silent) logger instance.
func TestInitLogger_DefaultFormat(t *testing.T) {
	// Precondition: None

	// Action: Initialize with empty string (should default to quiet mode)
	err := InitLogger("")

	// Expected: No error, silent but functional logger created
	if err != nil {
		t.Fatalf("Expected no error initializing default logger, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

// TestInitLogger_UnknownFormat validates graceful handling of unrecognized
// format strings by defaulting to quiet mode.
//
// Why: User configuration errors should not crash the application. Unknown
// formats should fall back to safe defaults.
//
// What: Calling InitLogger with an unrecognized format string should succeed
// and create a no-op logger rather than returning an error.
func TestInitLogger_UnknownFormat(t *testing.T) {
	// Precondition: None

	// Action: Initialize with unknown format string
	err := InitLogger("unknown-format")

	// Expected: No error, fallback to quiet logger
	if err != nil {
		t.Fatalf("Expected no error initializing logger with unknown format, got: %v", err)
	}

	if logger == nil {
		t.Fatal("Expected logger to be initialized")
	}

	sugaredLogger := GetSugaredLogger()
	if sugaredLogger == nil {
		t.Fatal("Expected sugared logger to be available")
	}
}

// =============================================================================
// MINUTIAE
// Tests validating obscure scenarios and non-functional requirements.
// =============================================================================

// TestLogger_ThreadSafety validates that the logging package is safe for
// concurrent access from multiple goroutines.
//
// Why: Real applications call logging from multiple goroutines. Race
// conditions could cause crashes or data corruption.
//
// What: Multiple concurrent goroutines calling GetSugaredLogger and logging
// messages should not cause panics, data races, or inconsistent behavior.
func TestLogger_ThreadSafety(t *testing.T) {
	// Precondition: Logger must be initialized
	err := InitLogger("console")
	if err != nil {
		t.Fatalf("Failed to initialize logger: %v", err)
	}

	// Action: Concurrent access from multiple goroutines
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

	// Expected: All goroutines complete without error
	for i := 0; i < 10; i++ {
		<-done
	}
}

// =============================================================================
// BENCHMARKS
// Performance tests to ensure logging initialization is efficient.
// =============================================================================

// BenchmarkInitLogger_Console measures the performance of initializing
// a console-format logger.
//
// Why: Logger initialization may occur at startup or during configuration
// reloads. Excessive initialization time impacts application responsiveness.
func BenchmarkInitLogger_Console(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = InitLogger("console")
	}
}

// BenchmarkInitLogger_JSON measures the performance of initializing
// a JSON-format logger.
//
// Why: JSON loggers may have different initialization costs due to encoder
// configuration. This benchmark ensures JSON mode remains performant.
func BenchmarkInitLogger_JSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = InitLogger("json")
	}
}

// BenchmarkGetSugaredLogger measures the performance of retrieving the
// sugared logger instance.
//
// Why: GetSugaredLogger is called frequently throughout application code.
// It must be extremely fast to avoid impacting application performance.
func BenchmarkGetSugaredLogger(b *testing.B) {
	_ = InitLogger("console")
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		GetSugaredLogger()
	}
}
