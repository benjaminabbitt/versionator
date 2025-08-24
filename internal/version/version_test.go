package version

import (
	"github.com/spf13/afero"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/suite"
	"versionator/internal/vcs/mock"
)

// VersionTestSuite contains the test suite for version operations
type VersionTestSuite struct {
	suite.Suite
	fs afero.Fs
	sut *Version
}

// SetupTest runs before each test method
func (suite *VersionTestSuite) SetupTest() {
	suite.fs = afero.NewMemMapFs()
	suite.sut = NewVersion(suite.fs, ".", nil) // Pass nil VCS for most tests
}

func (suite *VersionTestSuite) TestGetCurrentVersion_NoVersionFile_NoVCS() {
	// Test getting version when no VERSION file exists and no VCS
	// The version package will fallback to current directory when no VCS is active
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error when creating default version")
	suite.Equal("0.0.0", version, "Expected default version '0.0.0'")

	// Verify VERSION file was created at the correct path
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	_, err = suite.fs.Stat(filePath)
	suite.NoError(err, "Expected VERSION file to be created")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_NoVersionFile_WithVCS() {
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	// Create a memfs-friendly repo root and create it in the filesystem
	repoRoot := filepath.Join(string(os.PathSeparator), "repo", "root")
	suite.NoError(suite.fs.MkdirAll(repoRoot, 0755), "Failed to create repo root")

	// Create mock VCS that reports being in a repository
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(repoRoot, nil).AnyTimes()

	// Create Version instance with mock VCS
	versionWithVCS := NewVersion(suite.fs, ".", mockVCS)

	// Test getting version when no VERSION file exists but VCS is available
	version, err := versionWithVCS.GetCurrentVersion()
	suite.NoError(err, "Expected no error when creating default version")
	suite.Equal("0.0.0", version, "Expected default version '0.0.0'")

	// Verify VERSION file was created in repository root
	versionPath := filepath.Join(repoRoot, versionFile)
	_, err = suite.fs.Stat(versionPath)
	suite.NoError(err, "Expected VERSION file to be created")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_ExistingValidVersion() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create a VERSION file with valid content
	versionContent := "1.2.3"
	err = afero.WriteFile(suite.fs, filePath, []byte(versionContent), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test getting existing version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading valid version")
	suite.Equal("1.2.3", version, "Expected version '1.2.3'")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_EmptyVersionFile() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create an empty VERSION file
	err = afero.WriteFile(suite.fs, filePath, []byte(""), 0644)
	suite.NoError(err, "Failed to create empty VERSION file")

	// Test getting version from empty file
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading empty version file")
	suite.Equal("0.0.0", version, "Expected default version '0.0.0' for empty file")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_WhitespaceVersionFile() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create a VERSION file with whitespace
	err = afero.WriteFile(suite.fs, filePath, []byte("  \n\t  \n"), 0644)
	suite.NoError(err, "Failed to create whitespace VERSION file")

	// Test getting version from whitespace file
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading whitespace version file")
	suite.Equal("0.0.0", version, "Expected default version '0.0.0' for whitespace file")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_InvalidVersion() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create a VERSION file with invalid content
	err = afero.WriteFile(suite.fs, filePath, []byte("invalid-version"), 0644)
	suite.NoError(err, "Failed to create invalid VERSION file")

	// Test getting invalid version
	_, err = suite.sut.GetCurrentVersion()
	suite.Error(err, "Expected error reading invalid version")
	suite.Contains(err.Error(), "invalid version format", "Expected invalid version format error")
}

func (suite *VersionTestSuite) TestGetCurrentVersion_VCSError() {
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	// Create mock VCS that reports being in a repository but fails to get root
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return("", os.ErrPermission).AnyTimes()

	// Create Version instance with mock VCS
	versionWithVCS := NewVersion(suite.fs, ".", mockVCS)

	// Get the expected fallback file path and create parent directories
	filePath, err := versionWithVCS.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")

	// Test getting version when VCS fails - should fallback to current directory
	version, err := versionWithVCS.GetCurrentVersion()
	suite.NoError(err, "Expected no error when VCS fails (should fallback)")
	suite.Equal("0.0.0", version, "Expected default version '0.0.0'")

	// Verify VERSION file was created in current directory (fallback)
	_, err = suite.fs.Stat(filePath)
	suite.NoError(err, "Expected VERSION file to be created in current directory as fallback")
}

func (suite *VersionTestSuite) TestIncrement_Major() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test major increment
	err = suite.sut.Increment(MajorLevel)
	suite.NoError(err, "Expected no error incrementing major version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after increment")
	suite.Equal("2.0.0", version, "Expected version '2.0.0' after major increment")
}

func (suite *VersionTestSuite) TestIncrement_Minor() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test minor increment
	err = suite.sut.Increment(MinorLevel)
	suite.NoError(err, "Expected no error incrementing minor version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after increment")
	suite.Equal("1.3.0", version, "Expected version '1.3.0' after minor increment")
}

func (suite *VersionTestSuite) TestIncrement_Patch() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test patch increment
	err = suite.sut.Increment(PatchLevel)
	suite.NoError(err, "Expected no error incrementing patch version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after increment")
	suite.Equal("1.2.4", version, "Expected version '1.2.4' after patch increment")
}

func (suite *VersionTestSuite) TestIncrement_InvalidLevel() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test invalid level
	err = suite.sut.Increment(VersionLevel(999))
	suite.Error(err, "Expected error for invalid version level")
	suite.Contains(err.Error(), "invalid version level", "Expected invalid version level error")
}

func (suite *VersionTestSuite) TestDecrement_Major() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("2.3.4"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test major decrement
	err = suite.sut.Decrement(MajorLevel)
	suite.NoError(err, "Expected no error decrementing major version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after decrement")
	suite.Equal("1.0.0", version, "Expected version '1.0.0' after major decrement")
}

func (suite *VersionTestSuite) TestDecrement_Minor() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.3.4"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test minor decrement
	err = suite.sut.Decrement(MinorLevel)
	suite.NoError(err, "Expected no error decrementing minor version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after decrement")
	suite.Equal("1.2.0", version, "Expected version '1.2.0' after minor decrement")
}

func (suite *VersionTestSuite) TestDecrement_Patch() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.4"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test patch decrement
	err = suite.sut.Decrement(PatchLevel)
	suite.NoError(err, "Expected no error decrementing patch version")

	// Verify new version
	version, err := suite.sut.GetCurrentVersion()
	suite.NoError(err, "Expected no error reading version after decrement")
	suite.Equal("1.2.3", version, "Expected version '1.2.3' after patch decrement")
}

func (suite *VersionTestSuite) TestDecrement_MajorAtZero() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version at 0.x.x
	err = afero.WriteFile(suite.fs, filePath, []byte("0.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test major decrement at zero
	err = suite.sut.Decrement(MajorLevel)
	suite.Error(err, "Expected error decrementing major version below 0")
	suite.Contains(err.Error(), "cannot decrement major version below 0", "Expected major version below 0 error")
}

func (suite *VersionTestSuite) TestDecrement_MinorAtZero() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version with minor at 0
	err = afero.WriteFile(suite.fs, filePath, []byte("1.0.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test minor decrement at zero
	err = suite.sut.Decrement(MinorLevel)
	suite.Error(err, "Expected error decrementing minor version below 0")
	suite.Contains(err.Error(), "cannot decrement minor version below 0", "Expected minor version below 0 error")
}

func (suite *VersionTestSuite) TestDecrement_PatchAtZero() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version with patch at 0
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.0"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test patch decrement at zero
	err = suite.sut.Decrement(PatchLevel)
	suite.Error(err, "Expected error decrementing patch version below 0")
	suite.Contains(err.Error(), "cannot decrement patch version below 0", "Expected patch version below 0 error")
}

func (suite *VersionTestSuite) TestDecrement_InvalidLevel() {
	// Get the expected file path and create parent directories
	filePath, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Failed to get version file path")
	suite.NoError(suite.fs.MkdirAll(filepath.Dir(filePath), 0755), "Failed to create parent directories")
	
	// Create initial version
	err = afero.WriteFile(suite.fs, filePath, []byte("1.2.3"), 0644)
	suite.NoError(err, "Failed to create VERSION file")

	// Test invalid level
	err = suite.sut.Decrement(VersionLevel(999))
	suite.Error(err, "Expected error for invalid version level")
	suite.Contains(err.Error(), "invalid version level", "Expected invalid version level error")
}

func (suite *VersionTestSuite) TestGetVersionFilePath_WithVCS() {
	ctrl := gomock.NewController(suite.T())
	defer ctrl.Finish()

	// Create mock VCS
	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return("/repo/root", nil).Times(1)

	// Create Version instance with mock VCS
	versionWithVCS := NewVersion(suite.fs, ".", mockVCS)

	// Test getting version file path with VCS
	path, err := versionWithVCS.getVersionFilePath()
	suite.NoError(err, "Expected no error getting version file path")

	expectedPath := filepath.Join("/repo/root", versionFile)
	suite.Equal(expectedPath, path, "Expected correct path with VCS")
}

func (suite *VersionTestSuite) TestGetVersionFilePath_NoVCS() {
	// Test getting version file path without VCS
	// This test assumes no active VCS is registered that would interfere
	path, err := suite.sut.getVersionFilePath()
	suite.NoError(err, "Expected no error getting version file path")

	expectedPath := filepath.Join(suite.sut.workingDir, versionFile)
	suite.Equal(expectedPath, path, "Expected correct path without VCS")
}

// TestVersionTestSuite runs the test suite
func TestVersionTestSuite(t *testing.T) {
	suite.Run(t, new(VersionTestSuite))
}
