package git

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// GitVCSTestSuite is the test suite for GitVersionControlSystem
type GitVCSTestSuite struct {
	suite.Suite
	fs      afero.Fs
	gitVCS  *GitVersionControlSystem
	tempDir string
}

// SetupTest runs before each test
func (suite *GitVCSTestSuite) SetupTest() {
	suite.fs = afero.NewMemMapFs()
	suite.gitVCS = NewGitVCS(suite.fs)
	suite.tempDir = "/test-repo"
}

// TestNewGitVCS tests the constructor
func (suite *GitVCSTestSuite) TestNewGitVCS() {
	fs := afero.NewMemMapFs()
	vcs := NewGitVCS(fs)

	suite.NotNil(vcs)
	suite.Equal(fs, vcs.fs)
	suite.Empty(vcs.repoRoot)
}

// TestNewGitVCSDefault tests the default constructor
func (suite *GitVCSTestSuite) TestNewGitVCSDefault() {
	vcs := NewGitVCSDefault()

	suite.NotNil(vcs)
	suite.IsType(&afero.OsFs{}, vcs.fs)
	suite.Empty(vcs.repoRoot)
}

// TestName tests the Name method
func (suite *GitVCSTestSuite) TestName() {
	name := suite.gitVCS.Name()
	suite.Equal("git", name)
}

// TestGetWorkingDirWithOsFs tests getWorkingDir with OS filesystem
func (suite *GitVCSTestSuite) TestGetWorkingDirWithOsFs() {
	osVCS := NewGitVCSDefault()

	dir, err := osVCS.getWorkingDir()

	// Should not error and should return a valid path
	suite.NoError(err)
	suite.NotEmpty(dir)

	// Verify it's an actual directory path
	suite.True(filepath.IsAbs(dir))
}

// TestFindGitDir tests the findGitDir method
func (suite *GitVCSTestSuite) TestFindGitDir() {
	// Create a directory structure with .git directory
	gitDir := filepath.Join(suite.tempDir, ".git")
	err := suite.fs.MkdirAll(gitDir, 0755)
	suite.NoError(err)

	// Create subdirectories
	subDir := filepath.Join(suite.tempDir, "sub1", "sub2")
	err = suite.fs.MkdirAll(subDir, 0755)
	suite.NoError(err)

	// Test finding git dir from subdirectory
	result := suite.gitVCS.findGitDir(subDir)
	suite.Equal(filepath.ToSlash(suite.tempDir), filepath.ToSlash(result))

	// Test finding git dir from root directory
	result = suite.gitVCS.findGitDir(suite.tempDir)
	suite.Equal(filepath.ToSlash(suite.tempDir), filepath.ToSlash(result))
}

// TestFindGitDirNotFound tests findGitDir when no .git directory exists
func (suite *GitVCSTestSuite) TestFindGitDirNotFound() {
	// Create directory without .git
	err := suite.fs.MkdirAll(suite.tempDir, 0755)
	suite.NoError(err)

	result := suite.gitVCS.findGitDir(suite.tempDir)
	suite.Empty(result)
}

// TestFindGitDirWithFile tests findGitDir when .git is a file (not directory)
func (suite *GitVCSTestSuite) TestFindGitDirWithFile() {
	// Create directory and .git file (not directory)
	err := suite.fs.MkdirAll(suite.tempDir, 0755)
	suite.NoError(err)

	gitFile := filepath.Join(suite.tempDir, ".git")
	err = afero.WriteFile(suite.fs, gitFile, []byte("gitdir: /some/path"), 0644)
	suite.NoError(err)

	result := suite.gitVCS.findGitDir(suite.tempDir)
	suite.Empty(result, "Should not find git dir when .git is a file")
}

// TestIsRepository tests the IsRepository method
func (suite *GitVCSTestSuite) TestIsRepository() {
	// Test without .git directory
	isRepo := suite.gitVCS.IsRepository()
	suite.False(isRepo)

	// Create .git directory structure
	gitDir := filepath.Join("/", ".git")
	err := suite.fs.MkdirAll(gitDir, 0755)
	suite.NoError(err)

	// Test with .git directory
	isRepo = suite.gitVCS.IsRepository()
	suite.True(isRepo)
	suite.Equal("/", suite.gitVCS.repoRoot)
}

// TestGetRepositoryRoot tests the GetRepositoryRoot method
func (suite *GitVCSTestSuite) TestGetRepositoryRoot() {
	// Test when no repository is found
	_, err := suite.gitVCS.GetRepositoryRoot()
	suite.Error(err)
	suite.Contains(err.Error(), "not a git repository")

	// Create .git directory
	gitDir := filepath.Join("/", ".git")
	err = suite.fs.MkdirAll(gitDir, 0755)
	suite.NoError(err)

	// Test when repository is found
	root, err := suite.gitVCS.GetRepositoryRoot()
	suite.NoError(err)
	suite.Equal("/", root)
	suite.Equal("/", suite.gitVCS.repoRoot)
}

// TestGetRepositoryRootCached tests that GetRepositoryRoot returns cached value
func (suite *GitVCSTestSuite) TestGetRepositoryRootCached() {
	// Set cached root
	suite.gitVCS.repoRoot = "/cached/path"

	root, err := suite.gitVCS.GetRepositoryRoot()
	suite.NoError(err)
	suite.Equal("/cached/path", root)
}

// TestGetHashLength tests the GetHashLength method with memory filesystem
func (suite *GitVCSTestSuite) TestGetHashLength() {
	length := suite.gitVCS.GetHashLength()
	suite.Equal(7, length, "Should return default length for memory filesystem")
}

// TestGetHashLengthWithEnvVar tests GetHashLength with environment variable
func (suite *GitVCSTestSuite) TestGetHashLengthWithEnvVar() {
	// Test with OS filesystem to allow environment variable usage
	osVCS := NewGitVCSDefault()

	// Set environment variable
	originalValue := os.Getenv("VERSIONATOR_HASH_LENGTH")
	defer func() {
		if originalValue != "" {
			os.Setenv("VERSIONATOR_HASH_LENGTH", originalValue)
		} else {
			os.Unsetenv("VERSIONATOR_HASH_LENGTH")
		}
	}()

	os.Setenv("VERSIONATOR_HASH_LENGTH", "10")
	length := osVCS.GetHashLength()
	suite.Equal(10, length)

	// Test with invalid value
	os.Setenv("VERSIONATOR_HASH_LENGTH", "invalid")
	length = osVCS.GetHashLength()
	suite.Equal(7, length, "Should return default for invalid env var")

	// Test with out of range value
	os.Setenv("VERSIONATOR_HASH_LENGTH", "50")
	length = osVCS.GetHashLength()
	suite.Equal(7, length, "Should return default for out of range value")

	// Test with zero value
	os.Setenv("VERSIONATOR_HASH_LENGTH", "0")
	length = osVCS.GetHashLength()
	suite.Equal(7, length, "Should return default for zero value")
}

// TestOpenRepository tests the openRepository method error handling
func (suite *GitVCSTestSuite) TestOpenRepositoryError() {
	// Test when no repository root is available
	_, err := suite.gitVCS.openRepository()
	suite.Error(err)

	// Test with invalid repository path
	suite.gitVCS.repoRoot = "/invalid/path"
	_, err = suite.gitVCS.openRepository()
	suite.Error(err)
	suite.Contains(err.Error(), "failed to open git repository")
}

// TestMethodsRequiringRealGitRepo tests methods that need real git repo
// These test error handling when repository operations fail
func (suite *GitVCSTestSuite) TestMethodsRequiringRealGitRepo() {
	// Set up a fake repository root that won't be a real git repo
	suite.gitVCS.repoRoot = "/fake/repo"

	// Test IsWorkingDirectoryClean
	clean, err := suite.gitVCS.IsWorkingDirectoryClean()
	suite.Error(err)
	suite.False(clean)

	// Test GetVCSIdentifier
	id, err := suite.gitVCS.GetVCSIdentifier(7)
	suite.Error(err)
	suite.Empty(id)

	// Test CreateTag
	err = suite.gitVCS.CreateTag("v1.0.0", "Test tag")
	suite.Error(err)

	// Test TagExists
	exists, err := suite.gitVCS.TagExists("v1.0.0")
	suite.Error(err)
	suite.False(exists)
}

// TestGetVCSIdentifierValidation tests the GetVCSIdentifier parameter validation
func (suite *GitVCSTestSuite) TestGetVCSIdentifierValidation() {
	// Test invalid length values
	_, err := suite.gitVCS.GetVCSIdentifier(0)
	suite.Error(err)
	suite.Contains(err.Error(), "invalid hash length")

	_, err = suite.gitVCS.GetVCSIdentifier(-1)
	suite.Error(err)
	suite.Contains(err.Error(), "invalid hash length")

	_, err = suite.gitVCS.GetVCSIdentifier(41)
	suite.Error(err)
	suite.Contains(err.Error(), "invalid hash length")

	// Valid length should not error in parameter validation (though it will error on repo operations)
	_, err = suite.gitVCS.GetVCSIdentifier(7)
	suite.Error(err)
	// Error should be about repository, not parameter validation
	suite.NotContains(err.Error(), "invalid hash length")
}

// TestRunSuite runs the test suite
func TestGitVCSTestSuite(t *testing.T) {
	suite.Run(t, new(GitVCSTestSuite))
}

// Additional table-driven tests for edge cases
func TestFindGitDirEdgeCases(t *testing.T) {

	tests := []struct {
		name        string
		setupFunc   func(afero.Fs) string
		expected    string
		description string
	}{
		{
			name: "empty_path",
			setupFunc: func(fs afero.Fs) string {
				return ""
			},
			expected:    "",
			description: "should return empty for empty path",
		},
		{
			name: "root_path_no_git",
			setupFunc: func(fs afero.Fs) string {
				return "/"
			},
			expected:    "",
			description: "should return empty when no .git found at root",
		},
		{
			name: "nested_git_dirs",
			setupFunc: func(fs afero.Fs) string {
				// Create nested structure with multiple .git dirs
				fs.MkdirAll("/repo/.git", 0755)
				fs.MkdirAll("/repo/sub/nested/.git", 0755)
				return "/repo/sub/nested/deep"
			},
			expected:    "/repo/sub/nested",
			description: "should find nearest .git directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			gitVCS := NewGitVCS(fs)
			startPath := tt.setupFunc(fs)

			result := gitVCS.findGitDir(startPath)
			assert.Equal(t, tt.expected, filepath.ToSlash(result), tt.description)
		})
	}
}

// TestMemoryFilesystemIsolation ensures tests don't affect real filesystem
func TestMemoryFilesystemIsolation(t *testing.T) {
	fs := afero.NewMemMapFs()
	gitVCS := NewGitVCS(fs)

	// Verify we're using memory filesystem
	_, isMemFs := fs.(*afero.MemMapFs)
	require.True(t, isMemFs, "Should be using memory filesystem")

	// Create files in memory filesystem
	err := afero.WriteFile(fs, "/test-file", []byte("test content"), 0644)
	require.NoError(t, err)

	// Verify file exists in memory filesystem
	exists, err := afero.Exists(fs, "/test-file")
	require.NoError(t, err)
	require.True(t, exists)

	// Verify file doesn't exist on real filesystem
	_, err = os.Stat("/test-file")
	require.Error(t, err, "File should not exist on real filesystem")
	require.True(t, os.IsNotExist(err), "Should be file not found error")

	// Verify getWorkingDir returns "/" for memory filesystem
	dir, err := gitVCS.getWorkingDir()
	require.NoError(t, err)
	require.Equal(t, "/", dir, "Memory filesystem should return root as working dir")
}
