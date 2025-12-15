package cmd

import (
	"testing"

	"github.com/benjaminabbitt/versionator/internal/emit"
)

func TestEmit_SupportedFormats(t *testing.T) {
	formats := emit.SupportedFormats()

	// All supported formats
	expectedFormats := []string{
		"python", "json", "yaml", "go", "c", "c-header", "cpp", "cpp-header",
		"js", "ts", "java", "kotlin", "csharp", "php", "swift", "ruby", "rust",
	}

	for _, expected := range expectedFormats {
		found := false
		for _, f := range formats {
			if f == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected format %s to be supported", expected)
		}
	}
}

func TestEmit_IsValidFormat(t *testing.T) {
	validFormats := []string{
		"python", "json", "yaml", "go", "c", "c-header", "cpp", "cpp-header",
		"js", "ts", "java", "kotlin", "csharp", "php", "swift", "ruby", "rust",
	}

	for _, format := range validFormats {
		if !emit.IsValidFormat(format) {
			t.Errorf("expected %s to be a valid format", format)
		}
	}

	invalidFormats := []string{"invalid", "perl", "lua", ""}

	for _, format := range invalidFormats {
		if emit.IsValidFormat(format) {
			t.Errorf("expected %s to be an invalid format", format)
		}
	}
}
