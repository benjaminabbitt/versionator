package version

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"versionator/internal/vcs"
	"versionator/internal/vcs/mock"
)

func TestGetCurrentVersion_NoVersionFile_NoVCS(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Test getting version when no VERSION file exists and no VCS
	// The version package will fallback to current directory when no VCS is active
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error when creating default version, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0', got '%s'", version)
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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0', got '%s'", version)
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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file with valid content
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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create an empty VERSION file
	err := os.WriteFile(versionFile, []byte(""), 0644)
	if err != nil {
		t.Fatalf("Failed to create empty VERSION file: %v", err)
	}

	// Test getting version from empty file
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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file with whitespace
	err := os.WriteFile(versionFile, []byte("  \n\t  \n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create whitespace VERSION file: %v", err)
	}

	// Test getting version from whitespace file
	version, err := GetCurrentVersion()
	if err != nil {
		t.Fatalf("Expected no error reading whitespace version file, got: %v", err)
	}

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0' for whitespace file, got '%s'", version)
	}
}

func TestGetCurrentVersion_InvalidVersion(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

	// Create a VERSION file with invalid content
	err := os.WriteFile(versionFile, []byte("invalid-version"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid VERSION file: %v", err)
	}

	// Test getting invalid version
	_, err = GetCurrentVersion()
	if err == nil {
		t.Error("Expected error reading invalid version, got nil")
	}

	if !contains(err.Error(), "invalid version format") {
		t.Errorf("Expected invalid version format error, got: %v", err)
	}
}

func TestGetCurrentVersion_VCSError(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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

	if version != "0.0.0" {
		t.Errorf("Expected default version '0.0.0', got '%s'", version)
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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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
	defer os.Chdir(originalDir)
	os.Chdir(tempDir)

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

func TestGetVersionFilePath_WithVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock-git-path").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return("/repo/root", nil).Times(1)

	// Register the mock VCS
	vcs.RegisterVCS(mockVCS)
	defer vcs.UnregisterVCS("mock-git-path")

	// Test getting version file path with VCS
	path, err := getVersionFilePath()
	if err != nil {
		t.Fatalf("Expected no error getting version file path, got: %v", err)
	}

	expectedPath := filepath.Join("/repo/root", versionFile)
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
	}
}

func TestGetVersionFilePath_NoVCS(t *testing.T) {
	// Test getting version file path without VCS
	// This test assumes no active VCS is registered that would interfere
	path, err := getVersionFilePath()
	if err != nil {
		t.Fatalf("Expected no error getting version file path, got: %v", err)
	}

	cwd, _ := os.Getwd()
	expectedPath := filepath.Join(cwd, versionFile)
	if path != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, path)
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
