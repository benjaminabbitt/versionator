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

func TestGetCurrentVersion_NoVersionFile_NoVCS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Test getting version when no VERSION file exists and no VCS
	// The version package will fallback to current directory when no VCS is active
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error when creating default version, got: %v", err)
	}

	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}

	// Verify VERSION file was created
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created")
	}
}

func TestGetCurrentVersion_NoVersionFile_WithVCS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock VCS that reports being in a repository
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(tempDir, nil).AnyTimes()

	// Register the mock VCS
	vcs.RegisterVCS(mockVCS)
	defer vcs.UnregisterVCS("mock-git")

	// Test getting version when no VERSION file exists but VCS is available
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error when creating default version, got: %v", err)
	}

	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}

	// Verify VERSION file was created in repository root
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created in repository root")
	}
}

func TestGetCurrentVersion_ExistingValidVersion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create a legacy VERSION file with valid content (will be migrated)
	versionContent := "1.2.3"
	err := os.WriteFile(versionFile, []byte(versionContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test getting existing version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading valid version, got: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("Expected version '1.2.3', got '%s'", version)
	}
}

func TestGetCurrentVersion_EmptyVersionFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create an empty legacy VERSION file (will be migrated as 0.0.0)
	err := os.WriteFile(versionFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty VERSION file: %v", err)
	}

	// Test getting version from empty file - empty version parses as 0.0.0
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading empty version file, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0' for empty file, got '%s'", version)
	}
}

func TestGetCurrentVersion_WhitespaceVersionFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create a legacy VERSION file with whitespace (will be migrated as 0.0.0)
	err := os.WriteFile(versionFile, []byte("  \n\t  \n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create whitespace VERSION file: %v", err)
	}

	// Test getting version from whitespace file - parses as 0.0.0
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading whitespace version file, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0' for whitespace file, got '%s'", version)
	}
}

func TestGetCurrentVersion_UnparseableVersion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create a VERSION file with content that doesn't parse as a valid semver
	// The parser is lenient - unparseable strings result in 0.0.0
	err := os.WriteFile(versionFile, []byte("not a version"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test getting unparseable version - should return 0.0.0 (no error)
	version, err := GetCurrentVersion()
	if err != nil {
		t.Errorf("Expected no error for unparseable version, got: %v", err)
	}

	// Unparseable versions default to 0.0.0
	if version != "0.0.0" {
		t.Errorf("Expected version '0.0.0' for unparseable content, got '%s'", version)
	}
}

func TestGetCurrentVersion_VCSError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock VCS that reports being in a repository but fails to get root
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git-error").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return("", os.ErrPermission).AnyTimes()

	// Register the mock VCS
	vcs.RegisterVCS(mockVCS)
	defer vcs.UnregisterVCS("mock-git-error")

	// Test getting version when VCS fails - should fallback to current directory
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error when VCS fails (should fallback), got: %v", err)
	}

	if version != "0.0.1" {
		t.Errorf("Expected default version '0.0.1', got '%s'", version)
	}

	// Verify VERSION file was created in current directory (fallback)
	versionPath := filepath.Join(tempDir, versionFile)
	if _, err := os.Stat(versionPath); os.IsNotExist(err) {
		t.Error("Expected VERSION file to be created in current directory as fallback")
	}
}

func TestIncrement_Major(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test major increment
	err = Increment(MajorLevel)
	if err != nil {
		t.Fatalf("Expected no error incrementing major version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after increment, got: %v", err)
	}

	if version != "2.0.0" {
		t.Errorf("Expected version '2.0.0' after major increment, got '%s'", version)
	}
}

func TestIncrement_Minor(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test minor increment
	err = Increment(MinorLevel)
	if err != nil {
		t.Fatalf("Expected no error incrementing minor version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after increment, got: %v", err)
	}

	if version != "1.3.0" {
		t.Errorf("Expected version '1.3.0' after minor increment, got '%s'", version)
	}
}

func TestIncrement_Patch(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test patch increment
	err = Increment(PatchLevel)
	if err != nil {
		t.Fatalf("Expected no error incrementing patch version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after increment, got: %v", err)
	}

	if version != "1.2.4" {
		t.Errorf("Expected version '1.2.4' after patch increment, got '%s'", version)
	}
}

func TestIncrement_InvalidLevel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test invalid level
	err = Increment(VersionLevel(999))
	if err == nil {
		t.Error("Expected error for invalid version level, got nil")
	}

	if !contains(err.Error(), "invalid version level") {
		t.Errorf("Expected invalid version level error, got: %v", err)
	}
}

func TestDecrement_Major(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("2.3.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test major decrement
	err = Decrement(MajorLevel)
	if err != nil {
		t.Fatalf("Expected no error decrementing major version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after decrement, got: %v", err)
	}

	if version != "1.0.0" {
		t.Errorf("Expected version '1.0.0' after major decrement, got '%s'", version)
	}
}

func TestDecrement_Minor(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.3.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test minor decrement
	err = Decrement(MinorLevel)
	if err != nil {
		t.Fatalf("Expected no error decrementing minor version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after decrement, got: %v", err)
	}

	if version != "1.2.0" {
		t.Errorf("Expected version '1.2.0' after minor decrement, got '%s'", version)
	}
}

func TestDecrement_Patch(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.4"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test patch decrement
	err = Decrement(PatchLevel)
	if err != nil {
		t.Fatalf("Expected no error decrementing patch version, got: %v", err)
	}

	// Verify new version
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading version after decrement, got: %v", err)
	}

	if version != "1.2.3" {
		t.Errorf("Expected version '1.2.3' after patch decrement, got '%s'", version)
	}
}

func TestDecrement_MajorAtZero(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version at 0.x.x
	err := os.WriteFile(versionFile, []byte("0.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test major decrement at zero
	err = Decrement(MajorLevel)
	if err == nil {
		t.Error("Expected error decrementing major version below 0, got nil")
	}

	if !contains(err.Error(), "cannot decrement major version below 0") {
		t.Errorf("Expected major version below 0 error, got: %v", err)
	}
}

func TestDecrement_MinorAtZero(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version with minor at 0
	err := os.WriteFile(versionFile, []byte("1.0.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test minor decrement at zero
	err = Decrement(MinorLevel)
	if err == nil {
		t.Error("Expected error decrementing minor version below 0, got nil")
	}

	if !contains(err.Error(), "cannot decrement minor version below 0") {
		t.Errorf("Expected minor version below 0 error, got: %v", err)
	}
}

func TestDecrement_PatchAtZero(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version with patch at 0
	err := os.WriteFile(versionFile, []byte("1.2.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test patch decrement at zero
	err = Decrement(PatchLevel)
	if err == nil {
		t.Error("Expected error decrementing patch version below 0, got nil")
	}

	if !contains(err.Error(), "cannot decrement patch version below 0") {
		t.Errorf("Expected patch version below 0 error, got: %v", err)
	}
}

func TestDecrement_InvalidLevel(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Test invalid level
	err = Decrement(VersionLevel(999))
	if err == nil {
		t.Error("Expected error for invalid version level, got nil")
	}

	if !contains(err.Error(), "invalid version level") {
		t.Errorf("Expected invalid version level error, got: %v", err)
	}
}

func TestGetVersionPath_WalksUpToFindVersionFile(t *testing.T) {
	// Create temp directory structure: /tmp/root/subdir
	rootDir := t.TempDir()
	subDir := filepath.Join(rootDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create VERSION file in root
	versionPath := filepath.Join(rootDir, versionFile)
	err = os.WriteFile(versionPath, []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Change to subdir
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(subDir)

	// Unregister any VCS to test pure walk-up behavior
	vcs.UnregisterVCS("git")
	defer func() {
		// Re-register git VCS after test
		vcs.RegisterVCS(gitVCS.NewGitVCS())
	}()

	// Should find VERSION in parent directory
	path, err := getVersionPath()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != versionPath {
		t.Errorf("Expected to find VERSION at '%s', got '%s'", versionPath, path)
	}
}

func TestGetVersionPath_ReturnsCurrentDirWhenNotFound(t *testing.T) {
	// Create isolated temp directory with no VERSION file
	tempDir := t.TempDir()

	// Change to temp dir
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Unregister any VCS
	vcs.UnregisterVCS("git")
	defer func() {
		vcs.RegisterVCS(gitVCS.NewGitVCS())
	}()

	// Should return path in current directory (for creating new VERSION)
	path, err := getVersionPath()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	expectedPath := filepath.Join(tempDir, versionFile)
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetVersionPath_PrefersCloserVersionFile(t *testing.T) {
	// Create temp directory structure: /tmp/root/subdir with VERSION in both
	rootDir := t.TempDir()
	subDir := filepath.Join(rootDir, "subdir")
	err := os.MkdirAll(subDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}

	// Create VERSION files in both directories
	rootVersionPath := filepath.Join(rootDir, versionFile)
	subVersionPath := filepath.Join(subDir, versionFile)
	err = os.WriteFile(rootVersionPath, []byte("1.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create root VERSION file: %v", err)
	}
	err = os.WriteFile(subVersionPath, []byte("2.0.0"), 0644)
	if err != nil {
		t.Fatalf("Failed to create sub VERSION file: %v", err)
	}

	// Change to subdir
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(subDir)

	// Unregister any VCS
	vcs.UnregisterVCS("git")
	defer func() {
		vcs.RegisterVCS(gitVCS.NewGitVCS())
	}()

	// Should find VERSION in current directory (closest)
	path, err := getVersionPath()
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if path != subVersionPath {
		t.Errorf("Expected to find closer VERSION at '%s', got '%s'", subVersionPath, path)
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

func TestValidate_InvalidPreReleaseCharacter(t *testing.T) {
	tests := []struct {
		name       string
		preRelease string
	}{
		{"space character", "alpha beta"},
		{"underscore", "alpha_1"},
		{"special char @", "alpha@1"},
		{"unicode", "alpha.β"},
		{"empty part", "alpha..beta"},
		{"leading dot", ".alpha"},
		{"trailing dot", "alpha."},
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

func TestValidate_InvalidBuildMetadataCharacter(t *testing.T) {
	tests := []struct {
		name     string
		metadata string
	}{
		{"space character", "build 123"},
		{"underscore", "build_123"},
		{"special char #", "build#123"},
		{"empty part", "build..123"},
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

func TestValidate_ValidIdentifiers(t *testing.T) {
	tests := []struct {
		name       string
		preRelease string
		metadata   string
	}{
		{"alphanumeric prerelease", "alpha1", ""},
		{"hyphen in prerelease", "alpha-1", ""},
		{"numeric prerelease", "123", ""},
		{"dotted prerelease", "alpha.1.beta", ""},
		{"alphanumeric metadata", "build123", ""},
		{"hyphen in metadata", "build-123", ""},
		{"dotted metadata", "build.123.abc", ""},
		{"both prerelease and metadata", "alpha", "build123"},
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

func TestSetPrefix_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Set prefix
	err = SetPrefix("v")
	if err != nil {
		t.Fatalf("Expected no error setting prefix, got: %v", err)
	}

	// Verify prefix was set
	prefix, err := GetPrefix()
	if err != nil {
		t.Fatalf("Expected no error getting prefix, got: %v", err)
	}
	if prefix != "v" {
		t.Errorf("Expected prefix 'v', got '%s'", prefix)
	}
}

func TestGetPrefix_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create version with prefix
	err := os.WriteFile(versionFile, []byte("v1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	prefix, err := GetPrefix()
	if err != nil {
		t.Fatalf("Expected no error getting prefix, got: %v", err)
	}
	if prefix != "v" {
		t.Errorf("Expected prefix 'v', got '%s'", prefix)
	}
}

func TestSetPreRelease_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Set pre-release
	err = SetPreRelease("alpha.1")
	if err != nil {
		t.Fatalf("Expected no error setting pre-release, got: %v", err)
	}

	// Verify pre-release was set
	v, err := Load()
	if err != nil {
		t.Fatalf("Expected no error loading version, got: %v", err)
	}
	if v.PreRelease != "alpha.1" {
		t.Errorf("Expected pre-release 'alpha.1', got '%s'", v.PreRelease)
	}
}

func TestSetMetadata_Success(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Create initial version
	err := os.WriteFile(versionFile, []byte("1.2.3"), 0644)
	if err != nil {
		t.Fatalf("Failed to create VERSION file: %v", err)
	}

	// Set metadata
	err = SetMetadata("build.123")
	if err != nil {
		t.Fatalf("Expected no error setting metadata, got: %v", err)
	}

	// Verify metadata was set
	v, err := Load()
	if err != nil {
		t.Fatalf("Expected no error loading version, got: %v", err)
	}
	if v.BuildMetadata != "build.123" {
		t.Errorf("Expected metadata 'build.123', got '%s'", v.BuildMetadata)
	}
}

func TestSetPrefix_ErrorWhenNoVersionFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Don't create VERSION file - SetPrefix creates one if needed
	err := SetPrefix("v")
	if err != nil {
		t.Fatalf("SetPrefix should create VERSION file if needed, got: %v", err)
	}
}

func TestGetPrefix_CreatesVersionFileIfNeeded(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Don't create VERSION file - GetPrefix creates one if needed
	prefix, err := GetPrefix()
	if err != nil {
		t.Fatalf("GetPrefix should create VERSION file if needed, got: %v", err)
	}
	// Default prefix from config is "v"
	if prefix != "v" {
		t.Errorf("Expected default prefix 'v', got '%s'", prefix)
	}
}

func TestSetPreRelease_ErrorWhenNoVersionFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Don't create VERSION file - SetPreRelease creates one if needed
	err := SetPreRelease("alpha")
	if err != nil {
		t.Fatalf("SetPreRelease should create VERSION file if needed, got: %v", err)
	}
}

func TestSetMetadata_ErrorWhenNoVersionFile(t *testing.T) {
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(originalDir) }()
	_ = os.Chdir(tempDir)

	// Don't create VERSION file - SetMetadata creates one if needed
	err := SetMetadata("build123")
	if err != nil {
		t.Fatalf("SetMetadata should create VERSION file if needed, got: %v", err)
	}
}
