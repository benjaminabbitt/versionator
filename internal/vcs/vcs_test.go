package vcs

import (
	"errors"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the VCS registry: registering,
// retrieving, and detecting active version control systems.
// =============================================================================

// TestVCSRegistry_RegisterVCS_AddsVCSToRegistry validates that version control
// systems can be registered and stored in the registry.
//
// Why: The registry must reliably store VCS implementations so they can be
// retrieved later. Failure here would break all VCS operations.
//
// What: Registers a single mock VCS and verifies it is stored in the registry
// with the correct key.
func TestVCSRegistry_RegisterVCS_AddsVCSToRegistry(t *testing.T) {
	// Precondition: Empty registry with no registered VCS systems
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock").AnyTimes()

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	// Action: Register a VCS with the registry
	registry.RegisterVCS(mockVCS)

	// Expected: Registry contains exactly one VCS with the correct key
	if len(registry.systems) != 1 {
		t.Errorf("Expected 1 VCS registered, got %d", len(registry.systems))
	}

	if registry.systems["mock"] != mockVCS {
		t.Error("VCS not properly registered")
	}
}

// TestVCSRegistry_GetVCS_RetrievesRegisteredVCS validates that registered VCS
// systems can be retrieved by name.
//
// Why: Components need to retrieve specific VCS implementations by name to
// perform version control operations.
//
// What: Registers a VCS and retrieves it by name, verifying the correct
// instance is returned.
func TestVCSRegistry_GetVCS_RetrievesRegisteredVCS(t *testing.T) {
	// Precondition: Registry with one registered VCS
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	// Action: Retrieve the VCS by name
	retrievedVCS := registry.GetVCS("mock")

	// Expected: The exact VCS instance that was registered is returned
	if retrievedVCS != mockVCS {
		t.Error("Expected to retrieve the mock VCS")
	}
}

// TestVCSRegistry_GetActiveVCS_ReturnsVCSInRepository validates that the
// registry correctly identifies which VCS is active in the current directory.
//
// Why: When multiple VCS systems are registered, only one is typically active
// in a given directory. The registry must detect and return the correct one.
//
// What: Registers two VCS systems where only one reports being in a repository,
// and verifies that one is returned as active.
func TestVCSRegistry_GetActiveVCS_ReturnsVCSInRepository(t *testing.T) {
	// Precondition: Registry with two VCS systems, only one reporting as active
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS1 := mock.NewMockVersionControlSystem(ctrl)
	mockVCS2 := mock.NewMockVersionControlSystem(ctrl)

	// Map iteration order is non-deterministic, so use AnyTimes()
	// First VCS is not in a repository
	mockVCS1.EXPECT().IsRepository().Return(false).AnyTimes()
	// Second VCS is in a repository
	mockVCS2.EXPECT().IsRepository().Return(true).AnyTimes()

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["vcs1"] = mockVCS1
	registry.systems["vcs2"] = mockVCS2

	// Action: Get the active VCS from the registry
	activeVCS := registry.GetActiveVCS()

	// Expected: The VCS that reports being in a repository is returned
	if activeVCS != mockVCS2 {
		t.Error("Expected mockVCS2 to be the active VCS")
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for alternate flows: listing, unregistering, and global function access.
// =============================================================================

// TestVCSRegistry_ListVCS_ReturnsAllRegisteredNames validates that the registry
// can enumerate all registered VCS system names.
//
// Why: Users and tooling need to see what VCS systems are available in the
// current configuration.
//
// What: Registers two VCS systems and verifies ListVCS returns both names.
func TestVCSRegistry_ListVCS_ReturnsAllRegisteredNames(t *testing.T) {
	// Precondition: Registry with two registered VCS systems
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS1 := mock.NewMockVersionControlSystem(ctrl)
	mockVCS2 := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["vcs1"] = mockVCS1
	registry.systems["vcs2"] = mockVCS2

	// Action: List all VCS names
	vcsNames := registry.ListVCS()

	// Expected: Both VCS names are returned (order independent)
	if len(vcsNames) != 2 {
		t.Errorf("Expected 2 VCS names, got %d", len(vcsNames))
	}

	nameMap := make(map[string]bool)
	for _, name := range vcsNames {
		nameMap[name] = true
	}

	if !nameMap["vcs1"] || !nameMap["vcs2"] {
		t.Error("Expected both vcs1 and vcs2 in the list")
	}
}

// TestVCSRegistry_UnregisterVCS_RemovesVCSFromRegistry validates that VCS
// systems can be removed from the registry.
//
// Why: Dynamic reconfiguration may require removing VCS systems, and cleanup
// during shutdown depends on this functionality.
//
// What: Registers a VCS, unregisters it, and verifies the registry is empty.
func TestVCSRegistry_UnregisterVCS_RemovesVCSFromRegistry(t *testing.T) {
	// Precondition: Registry with one registered VCS
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	// Action: Unregister the VCS
	registry.UnregisterVCS("mock")

	// Expected: Registry is now empty
	if len(registry.systems) != 0 {
		t.Errorf("Expected 0 VCS registered after unregistering, got %d", len(registry.systems))
	}
}

// TestGlobalFunctions_OperateOnGlobalRegistry validates that the package-level
// convenience functions correctly operate on the global registry.
//
// Why: The global functions provide a simplified API for common use cases.
// They must correctly delegate to the global registry instance.
//
// What: Exercises all global functions (RegisterVCS, GetVCS, GetActiveVCS,
// ListVCS, UnregisterVCS) and verifies correct behavior.
func TestGlobalFunctions_OperateOnGlobalRegistry(t *testing.T) {
	// Precondition: Save original registry to restore after test
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	originalRegistry := registry
	defer func() {
		registry = originalRegistry
	}()

	registry = &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()

	// Action: Test RegisterVCS global function
	RegisterVCS(mockVCS)

	// Expected: VCS is retrievable via GetVCS
	retrievedVCS := GetVCS("mock")
	if retrievedVCS != mockVCS {
		t.Error("Expected to retrieve the mock VCS using global function")
	}

	// Action: Test GetActiveVCS global function
	activeVCS := GetActiveVCS()

	// Expected: Mock VCS is returned as active
	if activeVCS != mockVCS {
		t.Error("Expected mock VCS to be active using global function")
	}

	// Action: Test ListVCS global function
	vcsNames := ListVCS()

	// Expected: One VCS named 'mock' in the list
	if len(vcsNames) != 1 || vcsNames[0] != "mock" {
		t.Error("Expected one VCS named 'mock' in global list")
	}

	// Action: Test UnregisterVCS global function
	UnregisterVCS("mock")

	// Expected: VCS is no longer retrievable
	retrievedVCS = GetVCS("mock")
	if retrievedVCS != nil {
		t.Error("Expected VCS to be unregistered using global function")
	}
}

// TestMockVCS_AllOperations_SuccessPath validates that all VCS interface
// methods work correctly through the mock in success scenarios.
//
// Why: Ensures the mock correctly implements all interface methods and can be
// used reliably in other tests.
//
// What: Exercises each VCS interface method with expected success responses.
func TestMockVCS_AllOperations_SuccessPath(t *testing.T) {
	// Precondition: Mock VCS configured for success responses
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	// Action: Test Name method
	mockVCS.EXPECT().Name().Return("mock-git").Times(1)
	name := mockVCS.Name()

	// Expected: Returns configured name
	if name != "mock-git" {
		t.Errorf("Expected name 'mock-git', got '%s'", name)
	}

	// Action: Test IsRepository method
	mockVCS.EXPECT().IsRepository().Return(true).Times(1)
	isRepo := mockVCS.IsRepository()

	// Expected: Returns true
	if !isRepo {
		t.Error("Expected IsRepository to return true")
	}

	// Action: Test GetRepositoryRoot method
	expectedRoot := "/path/to/repo"
	mockVCS.EXPECT().GetRepositoryRoot().Return(expectedRoot, nil).Times(1)
	root, err := mockVCS.GetRepositoryRoot()

	// Expected: Returns path without error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if root != expectedRoot {
		t.Errorf("Expected root '%s', got '%s'", expectedRoot, root)
	}

	// Action: Test IsWorkingDirectoryClean method
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil).Times(1)
	isClean, err := mockVCS.IsWorkingDirectoryClean()

	// Expected: Returns true without error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !isClean {
		t.Error("Expected working directory to be clean")
	}

	// Action: Test GetVCSIdentifier method
	expectedHash := "abc1234"
	mockVCS.EXPECT().GetVCSIdentifier(7).Return(expectedHash, nil).Times(1)
	hash, err := mockVCS.GetVCSIdentifier(7)

	// Expected: Returns hash without error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hash != expectedHash {
		t.Errorf("Expected hash '%s', got '%s'", expectedHash, hash)
	}

	// Action: Test CreateTag method
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release version 1.0.0").Return(nil).Times(1)
	err = mockVCS.CreateTag("v1.0.0", "Release version 1.0.0")

	// Expected: Returns no error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Action: Test TagExists method
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil).Times(1)
	exists, err := mockVCS.TagExists("v1.0.0")

	// Expected: Returns true without error
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected tag to exist")
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes when VCS operations encounter errors.
// =============================================================================

// TestMockVCS_GetRepositoryRoot_ReturnsError validates error propagation from
// GetRepositoryRoot.
//
// Why: Callers must handle cases where the repository root cannot be determined
// (e.g., not in a repository, permission issues).
//
// What: Configures mock to return an error and verifies it propagates correctly.
func TestMockVCS_GetRepositoryRoot_ReturnsError(t *testing.T) {
	// Precondition: Mock configured to return an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	expectedError := errors.New("not in a repository")

	// Action: Call GetRepositoryRoot which returns an error
	mockVCS.EXPECT().GetRepositoryRoot().Return("", expectedError).Times(1)
	_, err := mockVCS.GetRepositoryRoot()

	// Expected: Error is returned with correct message
	if err == nil {
		t.Error("Expected an error")
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedError.Error(), err.Error())
	}
}

// TestMockVCS_IsWorkingDirectoryClean_ReturnsError validates error propagation
// and return value when checking working directory state fails.
//
// Why: Callers must handle failures when checking directory state and should
// not assume the directory is clean when an error occurs.
//
// What: Configures mock to return an error and verifies false is returned
// along with the error.
func TestMockVCS_IsWorkingDirectoryClean_ReturnsError(t *testing.T) {
	// Precondition: Mock configured to return an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	expectedError := errors.New("not in a repository")

	// Action: Call IsWorkingDirectoryClean which returns an error
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, expectedError).Times(1)
	isClean, err := mockVCS.IsWorkingDirectoryClean()

	// Expected: Error is returned and isClean is false
	if err == nil {
		t.Error("Expected an error")
	}
	if isClean {
		t.Error("Expected working directory to not be clean when error occurs")
	}
}

// TestMockVCS_GetVCSIdentifier_ReturnsError validates error propagation from
// GetVCSIdentifier.
//
// Why: Callers must handle cases where the VCS identifier (commit hash, etc.)
// cannot be retrieved.
//
// What: Configures mock to return an error and verifies it propagates correctly.
func TestMockVCS_GetVCSIdentifier_ReturnsError(t *testing.T) {
	// Precondition: Mock configured to return an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	expectedError := errors.New("not in a repository")

	// Action: Call GetVCSIdentifier which returns an error
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("", expectedError).Times(1)
	_, err := mockVCS.GetVCSIdentifier(7)

	// Expected: Error is returned
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestMockVCS_CreateTag_ReturnsError validates error propagation from CreateTag.
//
// Why: Tag creation can fail for various reasons (tag exists, permission denied,
// invalid characters). Callers must handle these failures.
//
// What: Configures mock to return an error and verifies it propagates correctly.
func TestMockVCS_CreateTag_ReturnsError(t *testing.T) {
	// Precondition: Mock configured to return an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	expectedError := errors.New("not in a repository")

	// Action: Call CreateTag which returns an error
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release").Return(expectedError).Times(1)
	err := mockVCS.CreateTag("v1.0.0", "Release")

	// Expected: Error is returned
	if err == nil {
		t.Error("Expected an error")
	}
}

// TestMockVCS_TagExists_ReturnsError validates error propagation from TagExists.
//
// Why: Checking tag existence can fail (repository access issues). Callers must
// handle these failures and not assume the tag doesn't exist.
//
// What: Configures mock to return an error and verifies it propagates correctly.
func TestMockVCS_TagExists_ReturnsError(t *testing.T) {
	// Precondition: Mock configured to return an error
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	expectedError := errors.New("not in a repository")

	// Action: Call TagExists which returns an error
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, expectedError).Times(1)
	_, err := mockVCS.TagExists("v1.0.0")

	// Expected: Error is returned
	if err == nil {
		t.Error("Expected an error")
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and uncommon scenarios.
// =============================================================================

// TestVCSRegistry_GetActiveVCS_ReturnsNilWhenNoActiveVCS validates behavior
// when no VCS reports being in a repository.
//
// Why: In directories outside any VCS repository, GetActiveVCS must return nil
// rather than a random VCS or panic.
//
// What: Registers a VCS that reports not being in a repository and verifies
// nil is returned.
func TestVCSRegistry_GetActiveVCS_ReturnsNilWhenNoActiveVCS(t *testing.T) {
	// Precondition: Registry with VCS that is not in a repository
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().IsRepository().Return(false)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	// Action: Get active VCS when none are in a repository
	activeVCS := registry.GetActiveVCS()

	// Expected: nil is returned
	if activeVCS != nil {
		t.Error("Expected no active VCS")
	}
}

// TestVCSRegistry_GetVCS_ReturnsNilForNonexistent validates behavior when
// requesting a VCS that has not been registered.
//
// Why: Code may attempt to retrieve VCS systems by name without knowing if
// they exist. The registry must return nil gracefully rather than panic.
//
// What: Attempts to retrieve a non-existent VCS and verifies nil is returned.
func TestVCSRegistry_GetVCS_ReturnsNilForNonexistent(t *testing.T) {
	// Precondition: Registry with one VCS registered under different name
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	// Action: Attempt to retrieve non-existent VCS
	nonExistentVCS := registry.GetVCS("nonexistent")

	// Expected: nil is returned
	if nonExistentVCS != nil {
		t.Error("Expected nil for non-existent VCS")
	}
}
