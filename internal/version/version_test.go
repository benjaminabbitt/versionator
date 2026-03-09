package version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose: reading, writing, and representing versions
// =============================================================================

// Validates that an existing valid VERSION file is read correctly.
// This is the primary happy path - users should be able to store and retrieve versions.
func TestGetCurrentVersion_ExistingValidVersion(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION file exists with valid semver content
	versionContent := "1.2.3"
	err := os.WriteFile(versionFile, []byte(versionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Read the version
	version, err := GetCurrentVersion()

	// Expected: Version is returned without error
	if err != nil {
		t.Fatalf("Expected no error reading valid version, got: %v", err)
	}
	if version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", version)
	}
}

// Validates that Save writes a valid version to the VERSION file.
// Save is the primary write operation - ensures versions persist correctly.
func TestSave_Success(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Valid version struct
	v := &Version{Major: 1, Minor: 2, Patch: 3, Prefix: "v"}

	// Action: Save the version
	err := Save(v)

	// Expected: File created with correct content
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	content, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("Failed to read VERSION file: %v", err)
	}
	expected := "v1.2.3\n"
	if string(content) != expected {
		t.Errorf("Expected VERSION content %q, got %q", expected, string(content))
	}
}

// Validates that FullString returns the complete version representation.
// String formatting is how versions are displayed to users and in output.
func TestVersion_FullString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{
			name:     "simple version",
			version:  &Version{Major: 1, Minor: 2, Patch: 3},
			expected: "1.2.3",
		},
		{
			name:     "with prefix",
			version:  &Version{Prefix: "v", Major: 1, Minor: 0, Patch: 0},
			expected: "v1.0.0",
		},
		{
			name:     "with pre-release",
			version:  &Version{Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"},
			expected: "1.0.0-alpha",
		},
		{
			name:     "with build metadata",
			version:  &Version{Major: 1, Minor: 0, Patch: 0, BuildMetadata: "build.123"},
			expected: "1.0.0+build.123",
		},
		{
			name:     "full version",
			version:  &Version{Prefix: "v", Major: 2, Minor: 3, Patch: 4, PreRelease: "rc.1", BuildMetadata: "sha.abc123"},
			expected: "v2.3.4-rc.1+sha.abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.FullString(); got != tt.expected {
				t.Errorf("FullString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows: auto-creation, increment/decrement, setters
// =============================================================================

// Validates that GetCurrentVersion creates a default VERSION file when none exists.
// First-time users shouldn't need to manually create a VERSION file.
func TestGetCurrentVersion_NoVersionFile_NoVCS(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VERSION file, no VCS
	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Default version created
	if err != nil {
		t.Fatalf("Expected no error when creating default version, got: %v", err)
	}
	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created")
	}
}

// Validates that GetCurrentVersion places VERSION in repository root when VCS is active.
// In a monorepo or subdirectory, version should be at the project root.
func TestGetCurrentVersion_NoVersionFile_WithVCS(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Precondition: Mock VCS reports being in a repository
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(tempDir, nil).AnyTimes()
	vcs.RegisterVCS(mockVCS)
	defer vcs.UnregisterVCS("mock-git")

	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Default version created in repository root
	if err != nil {
		t.Fatalf("Expected no error when creating default version, got: %v", err)
	}
	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created in repository root")
	}
}

// Validates that Increment correctly bumps the major version.
// Major bumps reset minor and patch per SemVer spec.
func TestIncrement_Major(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 1.2.3
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Increment major
	err = Increment(MajorLevel)

	// Expected: 2.0.0 (minor and patch reset)
	if err != nil {
		t.Fatalf("Expected no error incrementing major version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "2.0.0" {
		t.Errorf("Expected version '2.0.0' after major increment, got '%s'", version)
	}
}

// Validates that Increment correctly bumps the minor version.
// Minor bumps reset patch per SemVer spec.
func TestIncrement_Minor(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 1.2.3
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Increment minor
	err = Increment(MinorLevel)

	// Expected: 1.3.0 (patch reset)
	if err != nil {
		t.Fatalf("Expected no error incrementing minor version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "1.3.0" {
		t.Errorf("Expected version '1.3.0' after minor increment, got '%s'", version)
	}
}

// Validates that Increment correctly bumps the patch version.
func TestIncrement_Patch(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 1.2.3
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Increment patch
	err = Increment(PatchLevel)

	// Expected: 1.2.4
	if err != nil {
		t.Fatalf("Expected no error incrementing patch version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "1.2.4" {
		t.Errorf("Expected version '1.2.4' after patch increment, got '%s'", version)
	}
}

// Validates that Decrement correctly reduces the major version.
// Major decrements reset minor and patch to 0.
func TestDecrement_Major(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 2.3.4
	err := os.WriteFile(versionFile, []byte("2.3.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement major
	err = Decrement(MajorLevel)

	// Expected: 1.0.0 (minor and patch reset)
	if err != nil {
		t.Fatalf("Expected no error decrementing major version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "1.0.0" {
		t.Errorf("Expected version '1.0.0' after major decrement, got '%s'", version)
	}
}

// Validates that Decrement correctly reduces the minor version.
func TestDecrement_Minor(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 1.3.4
	err := os.WriteFile(versionFile, []byte("1.3.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement minor
	err = Decrement(MinorLevel)

	// Expected: 1.2.0 (patch reset)
	if err != nil {
		t.Fatalf("Expected no error decrementing minor version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "1.2.0" {
		t.Errorf("Expected version '1.2.0' after minor decrement, got '%s'", version)
	}
}

// Validates that Decrement correctly reduces the patch version.
func TestDecrement_Patch(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION at 1.2.4
	err := os.WriteFile(versionFile, []byte("1.2.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement patch
	err = Decrement(PatchLevel)

	// Expected: 1.2.3
	if err != nil {
		t.Fatalf("Expected no error decrementing patch version, got: %v", err)
	}
	version, _ := GetCurrentVersion()
	if version != "1.2.3" {
		t.Errorf("Expected version '1.2.3' after patch decrement, got '%s'", version)
	}
}

// Validates that SetPrefix updates the version prefix.
// Prefixes like "v" are common in Git tags.
func TestSetPrefix_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists without prefix
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Set prefix
	err = SetPrefix("v")

	// Expected: Prefix is set
	if err != nil {
		t.Fatalf("Expected no error setting prefix, got: %v", err)
	}
	prefix, _ := GetPrefix()
	if prefix != "v" {
		t.Errorf("Expected prefix 'v', got '%s'", prefix)
	}
}

// Validates that GetPrefix reads the prefix from an existing VERSION file.
func TestGetPrefix_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists with prefix
	err := os.WriteFile(versionFile, []byte("v1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Get prefix
	prefix, err := GetPrefix()

	// Expected: Prefix is returned
	if err != nil {
		t.Fatalf("Expected no error getting prefix, got: %v", err)
	}
	if prefix != "v" {
		t.Errorf("Expected prefix 'v', got '%s'", prefix)
	}
}

// Validates that SetPreRelease updates the pre-release identifier.
// Pre-release identifiers mark unstable versions (alpha, beta, rc).
func TestSetPreRelease_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Set pre-release
	err = SetPreRelease("alpha.1")

	// Expected: Pre-release is set
	if err != nil {
		t.Fatalf("Expected no error setting pre-release, got: %v", err)
	}
	v, _ := Load()
	if v.PreRelease != "alpha.1" {
		t.Errorf("Expected pre-release 'alpha.1', got '%s'", v.PreRelease)
	}
}

// Validates that SetMetadata updates the build metadata.
// Build metadata provides traceability (git sha, build number).
func TestSetMetadata_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Set metadata
	err = SetMetadata("build.123")

	// Expected: Metadata is set
	if err != nil {
		t.Fatalf("Expected no error setting metadata, got: %v", err)
	}
	v, _ := Load()
	if v.BuildMetadata != "build.123" {
		t.Errorf("Expected metadata 'build.123', got '%s'", v.BuildMetadata)
	}
}

// Validates that Save handles versions with pre-release and metadata.
func TestSave_WithPreReleaseAndMetadata(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Full version struct
	v := &Version{
		Major:         2,
		Minor:         0,
		Patch:         0,
		PreRelease:    "rc.1",
		BuildMetadata: "sha.abc123",
		Prefix:        "v",
	}

	// Action: Save
	err := Save(v)

	// Expected: Full version written
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	content, _ := os.ReadFile(versionFile)
	expected := "v2.0.0-rc.1+sha.abc123\n"
	if string(content) != expected {
		t.Errorf("Expected VERSION content %q, got %q", expected, string(content))
	}
}

// Validates MajorString returns only the major component.
func TestVersion_MajorString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{name: "zero major", version: &Version{Major: 0, Minor: 1, Patch: 2}, expected: "0"},
		{name: "positive major", version: &Version{Major: 5, Minor: 1, Patch: 2}, expected: "5"},
		{name: "large major", version: &Version{Major: 100, Minor: 0, Patch: 0}, expected: "100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.MajorString(); got != tt.expected {
				t.Errorf("MajorString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Validates MinorString returns only the minor component.
func TestVersion_MinorString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{name: "zero minor", version: &Version{Major: 1, Minor: 0, Patch: 2}, expected: "0"},
		{name: "positive minor", version: &Version{Major: 1, Minor: 5, Patch: 2}, expected: "5"},
		{name: "large minor", version: &Version{Major: 1, Minor: 99, Patch: 0}, expected: "99"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.MinorString(); got != tt.expected {
				t.Errorf("MinorString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Validates PatchString returns only the patch component.
func TestVersion_PatchString(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{name: "zero patch", version: &Version{Major: 1, Minor: 2, Patch: 0}, expected: "0"},
		{name: "positive patch", version: &Version{Major: 1, Minor: 2, Patch: 5}, expected: "5"},
		{name: "large patch", version: &Version{Major: 1, Minor: 0, Patch: 999}, expected: "999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.PatchString(); got != tt.expected {
				t.Errorf("PatchString() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// Validates OriginalFullSemVer preserves the original prefix in output.
func TestVersion_OriginalFullSemVer(t *testing.T) {
	tests := []struct {
		name     string
		version  *Version
		expected string
	}{
		{name: "no prefix", version: &Version{Major: 1, Minor: 2, Patch: 3}, expected: "1.2.3"},
		{name: "with prefix", version: &Version{Prefix: "v", Major: 1, Minor: 0, Patch: 0}, expected: "v1.0.0"},
		{name: "with prefix and pre-release", version: &Version{Prefix: "v", Major: 1, Minor: 0, Patch: 0, PreRelease: "alpha"}, expected: "v1.0.0-alpha"},
		{name: "with prefix and metadata", version: &Version{Prefix: "v", Major: 1, Minor: 0, Patch: 0, BuildMetadata: "build.123"}, expected: "v1.0.0+build.123"},
		{name: "full version with prefix", version: &Version{Prefix: "v", Major: 2, Minor: 3, Patch: 4, PreRelease: "rc.1", BuildMetadata: "sha.abc123"}, expected: "v2.3.4-rc.1+sha.abc123"},
		{name: "no prefix with pre-release and metadata", version: &Version{Major: 1, Minor: 0, Patch: 0, PreRelease: "beta", BuildMetadata: "build"}, expected: "1.0.0-beta+build"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.version.OriginalFullSemVer(); got != tt.expected {
				t.Errorf("OriginalFullSemVer() = %v, want %v", got, tt.expected)
			}
		})
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and recovery
// =============================================================================

// Validates that GetCurrentVersion falls back gracefully when VCS fails.
// VCS errors shouldn't block version operations - fallback to current directory.
func TestGetCurrentVersion_VCSError(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Precondition: VCS reports repository but fails to get root
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git-error").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return("", os.ErrPermission).AnyTimes()
	vcs.RegisterVCS(mockVCS)
	defer vcs.UnregisterVCS("mock-git-error")

	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Fallback to current directory succeeds
	if err != nil {
		t.Fatalf("Expected no error when VCS fails (should fallback), got: %v", err)
	}
	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created in current directory as fallback")
	}
}

// Validates that Increment rejects invalid version levels.
// Prevents silent failures from typos or API misuse.
func TestIncrement_InvalidLevel(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Increment with invalid level
	err = Increment(VersionLevel(999))

	// Expected: Error returned
	if err == nil {
		t.Error("Expected error for invalid version level, got nil")
	}
	if !contains(err.Error(), "invalid version level") {
		t.Errorf("Expected invalid version level error, got: %v", err)
	}
}

// Validates that Decrement rejects invalid version levels.
func TestDecrement_InvalidLevel(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement with invalid level
	err = Decrement(VersionLevel(999))

	// Expected: Error returned
	if err == nil {
		t.Error("Expected error for invalid version level, got nil")
	}
	if !contains(err.Error(), "invalid version level") {
		t.Errorf("Expected invalid version level error, got: %v", err)
	}
}

// Validates that Decrement fails when major would go below zero.
// Versions cannot be negative per SemVer spec.
func TestDecrement_MajorAtZero(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Major is already 0
	err := os.WriteFile(versionFile, []byte("0.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement major
	err = Decrement(MajorLevel)

	// Expected: Error - cannot go below 0
	if err == nil {
		t.Error("Expected error decrementing major version below 0, got nil")
	}
	if !contains(err.Error(), "cannot decrement major version below 0") {
		t.Errorf("Expected major version below 0 error, got: %v", err)
	}
}

// Validates that Decrement fails when minor would go below zero.
func TestDecrement_MinorAtZero(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Minor is already 0
	err := os.WriteFile(versionFile, []byte("1.0.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement minor
	err = Decrement(MinorLevel)

	// Expected: Error - cannot go below 0
	if err == nil {
		t.Error("Expected error decrementing minor version below 0, got nil")
	}
	if !contains(err.Error(), "cannot decrement minor version below 0") {
		t.Errorf("Expected minor version below 0 error, got: %v", err)
	}
}

// Validates that Decrement fails when patch would go below zero.
func TestDecrement_PatchAtZero(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Patch is already 0
	err := os.WriteFile(versionFile, []byte("1.2.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Decrement patch
	err = Decrement(PatchLevel)

	// Expected: Error - cannot go below 0
	if err == nil {
		t.Error("Expected error decrementing patch version below 0, got nil")
	}
	if !contains(err.Error(), "cannot decrement patch version below 0") {
		t.Errorf("Expected patch version below 0 error, got: %v", err)
	}
}

// Validates that Save rejects invalid versions.
// Prevents persisting invalid state that would break downstream tools.
func TestSave_InvalidVersion(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Invalid version (negative major)
	v := &Version{Major: -1, Minor: 0, Patch: 0}

	// Action: Attempt save
	err := Save(v)

	// Expected: Error returned
	if err == nil {
		t.Error("Expected error for invalid version, got nil")
	}
	if !contains(err.Error(), "invalid version") {
		t.Errorf("Expected 'invalid version' error, got: %v", err)
	}
}

// Validates that Save fails with appropriate error for read-only directory.
// Users should get clear feedback when filesystem prevents writes.
func TestSave_ReadOnlyDirectory(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("Cannot test read-only directory as root")
	}

	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Directory is read-only
	_ = os.Chmod(tempDir, 0555)
	defer func() { _ = os.Chmod(tempDir, 0755) }()

	v := &Version{Major: 1, Minor: 0, Patch: 0}

	// Action: Attempt save
	err := Save(v)

	// Expected: Error with clear message
	if err == nil {
		t.Error("Expected error when saving to read-only directory, got nil")
	}
	if !contains(err.Error(), "failed to write VERSION") {
		t.Errorf("Expected 'failed to write VERSION' error, got: %v", err)
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and unusual inputs
// =============================================================================

// Validates that empty VERSION file is treated as 0.0.0.
// Handles corrupted or manually emptied files gracefully.
func TestGetCurrentVersion_EmptyVersionFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Empty VERSION file
	err := os.WriteFile(versionFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty VERSION file: %v", err)
	}

	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Defaults to 0.0.0
	if err != nil {
		t.Fatalf("Expected no error reading empty version file, got: %v", err)
	}
	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0' for empty file, got '%s'", version)
	}
}

// Validates that whitespace-only VERSION file is treated as 0.0.0.
func TestGetCurrentVersion_WhitespaceVersionFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION file with only whitespace
	err := os.WriteFile(versionFile, []byte("  \n\t  \n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create whitespace VERSION file: %v", err)
	}

	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Defaults to 0.0.0
	if err != nil {
		t.Fatalf("Expected no error reading whitespace version file, got: %v", err)
	}
	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0' for whitespace file, got '%s'", version)
	}
}

// Validates that unparseable VERSION content defaults to 0.0.0.
// Provides graceful degradation for corrupted files.
func TestGetCurrentVersion_UnparseableVersion(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: Invalid version content
	err := os.WriteFile(versionFile, []byte("not a version"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Action: Get version
	version, err := GetCurrentVersion()

	// Expected: Defaults to 0.0.0 (lenient parsing)
	if err != nil {
		t.Errorf("Expected no error for unparseable version, got: %v", err)
	}
	if version != "0.0.0" {
		t.Errorf("Expected version '0.0.0' for unparseable content, got '%s'", version)
	}
}

// Validates that getVersionPath walks up directories to find VERSION.
// Supports running commands from subdirectories of a project.
func TestGetVersionPath_WalksUpToFindVersionFile(t *testing.T) {
	rootDir := t.TempDir()
	subDir := filepath.Join(rootDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Precondition: VERSION in root, cwd is subdir
	versionPath := filepath.Join(rootDir, versionFile)
	err = os.WriteFile(versionPath, []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(subDir)

	// Unregister VCS to test pure walk-up
	vcs.UnregisterVCS("git")
	defer func() { vcs.RegisterVCS(gitVCS.NewGitVCSDefault()) }()

	// Action: Get version path
	path, err := getVersionPath()

	// Expected: Finds VERSION in parent
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if path != versionPath {
		t.Errorf("Expected to find VERSION at '%s', got '%s'", versionPath, path)
	}
}

// Validates that getVersionPath returns current dir when no VERSION exists.
// Enables creating new VERSION files in the current location.
func TestGetVersionPath_ReturnsCurrentDirWhenNotFound(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Unregister VCS
	vcs.UnregisterVCS("git")
	defer func() { vcs.RegisterVCS(gitVCS.NewGitVCSDefault()) }()

	// Action: Get version path (no VERSION exists)
	path, err := getVersionPath()

	// Expected: Returns path in current directory
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	expectedPath := filepath.Join(tempDir, versionFile)
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}
}

// Validates that getVersionPath prefers closer VERSION files.
// Supports nested projects with their own version files.
func TestGetVersionPath_PrefersCloserVersionFile(t *testing.T) {
	rootDir := t.TempDir()
	subDir := filepath.Join(rootDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Precondition: VERSION in both root and subdir
	rootVersionPath := filepath.Join(rootDir, versionFile)
	subVersionPath := filepath.Join(subDir, versionFile)
	_ = os.WriteFile(rootVersionPath, []byte("1.0.0"), 0644)
	_ = os.WriteFile(subVersionPath, []byte("2.0.0"), 0644)

	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(subDir)

	// Unregister VCS
	vcs.UnregisterVCS("git")
	defer func() { vcs.RegisterVCS(gitVCS.NewGitVCSDefault()) }()

	// Action: Get version path
	path, err := getVersionPath()

	// Expected: Finds closer VERSION in subdir
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if path != subVersionPath {
		t.Errorf("Expected to find closer VERSION at '%s', got '%s'", subVersionPath, path)
	}
}

// Validates that valid pre-release and metadata identifiers pass validation.
func TestValidate_ValidIdentifiers(t *testing.T) {
	tests := []struct {
		name       string
		preRelease string
		metadata   string
	}{
		{name: "alphanumeric prerelease", preRelease: "alpha1", metadata: ""},
		{name: "hyphen in prerelease", preRelease: "alpha-1", metadata: ""},
		{name: "numeric prerelease", preRelease: "123", metadata: ""},
		{name: "dotted prerelease", preRelease: "alpha.1.beta", metadata: ""},
		{name: "alphanumeric metadata", preRelease: "", metadata: "build123"},
		{name: "hyphen in metadata", preRelease: "", metadata: "build-123"},
		{name: "dotted metadata", preRelease: "", metadata: "build.123.abc"},
		{name: "both prerelease and metadata", preRelease: "alpha", metadata: "build123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Major:         1,
				Minor:         0,
				Patch:         0,
				PreRelease:    tt.preRelease,
				BuildMetadata: tt.metadata,
			}
			err := v.Validate()
			if err != nil {
				t.Errorf("Expected no error for preRelease='%s' metadata='%s', got: %v",
					tt.preRelease, tt.metadata, err)
			}
		})
	}
}

// Validates that SetPrefix creates VERSION file if needed.
// Users shouldn't need to initialize VERSION before setting prefix.
func TestSetPrefix_CreatesVersionFileIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VERSION file exists
	// Action: Set prefix
	err := SetPrefix("v")

	// Expected: VERSION created with prefix
	if err != nil {
		t.Fatalf("SetPrefix should create VERSION file if needed, got: %v", err)
	}
}

// Validates that GetPrefix creates VERSION file if needed.
func TestGetPrefix_CreatesVersionFileIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VERSION file
	// Action: Get prefix
	prefix, err := GetPrefix()

	// Expected: Default prefix returned, file created
	if err != nil {
		t.Fatalf("GetPrefix should create VERSION file if needed, got: %v", err)
	}
	if prefix != "v" {
		t.Errorf("Expected default prefix 'v', got '%s'", prefix)
	}
}

// Validates that SetPreRelease creates VERSION file if needed.
func TestSetPreRelease_CreatesVersionFileIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VERSION file
	// Action: Set pre-release
	err := SetPreRelease("alpha")

	// Expected: VERSION created with pre-release
	if err != nil {
		t.Fatalf("SetPreRelease should create VERSION file if needed, got: %v", err)
	}
}

// Validates that SetMetadata creates VERSION file if needed.
func TestSetMetadata_CreatesVersionFileIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: No VERSION file
	// Action: Set metadata
	err := SetMetadata("build123")

	// Expected: VERSION created with metadata
	if err != nil {
		t.Fatalf("SetMetadata should create VERSION file if needed, got: %v", err)
	}
}

// Validates that Save overwrites existing VERSION file.
func TestSave_UpdatesExistingFile(t *testing.T) {
	tempDir := t.TempDir()
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()
	_ = os.Chdir(tempDir)

	// Precondition: VERSION exists with old content
	_ = os.WriteFile(versionFile, []byte("1.0.0\n"), 0644)

	// Action: Save new version
	v := &Version{Major: 2, Minor: 0, Patch: 0}
	err := Save(v)

	// Expected: File overwritten
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	content, _ := os.ReadFile(versionFile)
	expected := "2.0.0\n"
	if string(content) != expected {
		t.Errorf("Expected VERSION content %q, got %q", expected, string(content))
	}
}

// =============================================================================
// MINUTIAE
// Obscure scenarios, defensive checks, validation edge cases
// =============================================================================

// Validates that invalid characters in pre-release identifiers are rejected.
// Per SemVer spec: only alphanumerics, hyphens, and dots allowed.
func TestValidate_InvalidPreReleaseCharacter(t *testing.T) {
	tests := []struct {
		name       string
		preRelease string
	}{
		{name: "space character", preRelease: "alpha beta"},
		{name: "underscore", preRelease: "alpha_1"},
		{name: "special char @", preRelease: "alpha@1"},
		{name: "unicode", preRelease: "alpha.β"},
		{name: "empty part", preRelease: "alpha..beta"},
		{name: "leading dot", preRelease: ".alpha"},
		{name: "trailing dot", preRelease: "alpha."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Major:      1,
				Minor:      0,
				Patch:      0,
				PreRelease: tt.preRelease,
			}
			err := v.Validate()
			if err == nil {
				t.Errorf("Expected error for pre-release '%s', got nil", tt.preRelease)
			}
		})
	}
}

// Validates that invalid characters in build metadata are rejected.
func TestValidate_InvalidBuildMetadataCharacter(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
	}{
		{name: "space character", metadata: "build 123"},
		{name: "underscore", metadata: "build_123"},
		{name: "special char #", metadata: "build#123"},
		{name: "empty part", metadata: "build..123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := &Version{
				Major:         1,
				Minor:         0,
				Patch:         0,
				BuildMetadata: tt.metadata,
			}
			err := v.Validate()
			if err == nil {
				t.Errorf("Expected error for metadata '%s', got nil", tt.metadata)
			}
		})
	}
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
