package buildinfo

import "testing"

func TestVersion_DefaultValue(t *testing.T) {
	// Version should have a default value when not set via ldflags
	if Version == "" {
		t.Error("Version should not be empty")
	}
	if Version != "dev" {
		t.Errorf("Version default should be 'dev', got %q", Version)
	}
}
