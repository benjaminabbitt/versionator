package vcs

import (
	"errors"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/golang/mock/gomock"
)

func TestVCSRegistry_RegisterVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock").AnyTimes()

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.RegisterVCS(mockVCS)

	if len(registry.systems) != 1 {
		t.Errorf("Expected 1 VCS registered, got %d", len(registry.systems))
	}

	if registry.systems["mock"] != mockVCS {
		t.Error("VCS not properly registered")
	}
}

func TestVCSRegistry_GetActiveVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create two mock VCS systems
	mockVCS1 := mock.NewMockVersionControlSystem(ctrl)
	mockVCS2 := mock.NewMockVersionControlSystem(ctrl)

	// First VCS is not in a repository
	mockVCS1.EXPECT().IsRepository().Return(false)
	// Second VCS is in a repository
	mockVCS2.EXPECT().IsRepository().Return(true)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["vcs1"] = mockVCS1
	registry.systems["vcs2"] = mockVCS2

	activeVCS := registry.GetActiveVCS()

	if activeVCS != mockVCS2 {
		t.Error("Expected mockVCS2 to be the active VCS")
	}
}

func TestVCSRegistry_GetActiveVCS_NoActiveVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().IsRepository().Return(false)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	activeVCS := registry.GetActiveVCS()

	if activeVCS != nil {
		t.Error("Expected no active VCS")
	}
}

func TestVCSRegistry_GetVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	retrievedVCS := registry.GetVCS("mock")
	if retrievedVCS != mockVCS {
		t.Error("Expected to retrieve the mock VCS")
	}

	nonExistentVCS := registry.GetVCS("nonexistent")
	if nonExistentVCS != nil {
		t.Error("Expected nil for non-existent VCS")
	}
}

func TestVCSRegistry_UnregisterVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["mock"] = mockVCS

	registry.UnregisterVCS("mock")

	if len(registry.systems) != 0 {
		t.Errorf("Expected 0 VCS registered after unregistering, got %d", len(registry.systems))
	}
}

func TestVCSRegistry_ListVCS(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS1 := mock.NewMockVersionControlSystem(ctrl)
	mockVCS2 := mock.NewMockVersionControlSystem(ctrl)

	registry := &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	registry.systems["vcs1"] = mockVCS1
	registry.systems["vcs2"] = mockVCS2

	vcsNames := registry.ListVCS()

	if len(vcsNames) != 2 {
		t.Errorf("Expected 2 VCS names, got %d", len(vcsNames))
	}

	// Check that both names are present (order doesn't matter)
	nameMap := make(map[string]bool)
	for _, name := range vcsNames {
		nameMap[name] = true
	}

	if !nameMap["vcs1"] || !nameMap["vcs2"] {
		t.Error("Expected both vcs1 and vcs2 in the list")
	}
}

// Test individual VCS operations using mock
func TestMockVCS_Operations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	// Test Name method
	mockVCS.EXPECT().Name().Return("mock-git").Times(1)
	name := mockVCS.Name()
	if name != "mock-git" {
		t.Errorf("Expected name 'mock-git', got '%s'", name)
	}

	// Test IsRepository method
	mockVCS.EXPECT().IsRepository().Return(true).Times(1)
	isRepo := mockVCS.IsRepository()
	if !isRepo {
		t.Error("Expected IsRepository to return true")
	}

	// Test GetRepositoryRoot method
	expectedRoot := "/path/to/repo"
	mockVCS.EXPECT().GetRepositoryRoot().Return(expectedRoot, nil).Times(1)
	root, err := mockVCS.GetRepositoryRoot()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if root != expectedRoot {
		t.Errorf("Expected root '%s', got '%s'", expectedRoot, root)
	}

	// Test IsWorkingDirectoryClean method
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(true, nil).Times(1)
	isClean, err := mockVCS.IsWorkingDirectoryClean()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !isClean {
		t.Error("Expected working directory to be clean")
	}

	// Test GetVCSIdentifier method
	expectedHash := "abc1234"
	mockVCS.EXPECT().GetVCSIdentifier(7).Return(expectedHash, nil).Times(1)
	hash, err := mockVCS.GetVCSIdentifier(7)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if hash != expectedHash {
		t.Errorf("Expected hash '%s', got '%s'", expectedHash, hash)
	}

	// Test CreateTag method
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release version 1.0.0").Return(nil).Times(1)
	err = mockVCS.CreateTag("v1.0.0", "Release version 1.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Test TagExists method
	mockVCS.EXPECT().TagExists("v1.0.0").Return(true, nil).Times(1)
	exists, err := mockVCS.TagExists("v1.0.0")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if !exists {
		t.Error("Expected tag to exist")
	}
}

// Test error scenarios
func TestMockVCS_ErrorScenarios(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)

	// Test GetRepositoryRoot with error
	expectedError := errors.New("not in a repository")
	mockVCS.EXPECT().GetRepositoryRoot().Return("", expectedError).Times(1)
	_, err := mockVCS.GetRepositoryRoot()
	if err == nil {
		t.Error("Expected an error")
	}
	if err.Error() != expectedError.Error() {
		t.Errorf("Expected error '%s', got '%s'", expectedError.Error(), err.Error())
	}

	// Test IsWorkingDirectoryClean with error
	mockVCS.EXPECT().IsWorkingDirectoryClean().Return(false, expectedError).Times(1)
	isClean, err := mockVCS.IsWorkingDirectoryClean()
	if err == nil {
		t.Error("Expected an error")
	}
	if isClean {
		t.Error("Expected working directory to not be clean when error occurs")
	}

	// Test GetVCSIdentifier with error
	mockVCS.EXPECT().GetVCSIdentifier(7).Return("", expectedError).Times(1)
	_, err = mockVCS.GetVCSIdentifier(7)
	if err == nil {
		t.Error("Expected an error")
	}

	// Test CreateTag with error
	mockVCS.EXPECT().CreateTag("v1.0.0", "Release").Return(expectedError).Times(1)
	err = mockVCS.CreateTag("v1.0.0", "Release")
	if err == nil {
		t.Error("Expected an error")
	}

	// Test TagExists with error
	mockVCS.EXPECT().TagExists("v1.0.0").Return(false, expectedError).Times(1)
	_, err = mockVCS.TagExists("v1.0.0")
	if err == nil {
		t.Error("Expected an error")
	}
}

// Test global functions with mock
func TestGlobalFunctions_WithMock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Save original registry and restore after test
	originalRegistry := registry
	defer func() {
		registry = originalRegistry
	}()

	// Create a new registry for testing
	registry = &VCSRegistry{
		systems: make(map[string]VersionControlSystem),
	}

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("mock").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()

	// Test RegisterVCS global function
	RegisterVCS(mockVCS)

	// Test GetVCS global function
	retrievedVCS := GetVCS("mock")
	if retrievedVCS != mockVCS {
		t.Error("Expected to retrieve the mock VCS using global function")
	}

	// Test GetActiveVCS global function
	activeVCS := GetActiveVCS()
	if activeVCS != mockVCS {
		t.Error("Expected mock VCS to be active using global function")
	}

	// Test ListVCS global function
	vcsNames := ListVCS()
	if len(vcsNames) != 1 || vcsNames[0] != "mock" {
		t.Error("Expected one VCS named 'mock' in global list")
	}

	// Test UnregisterVCS global function
	UnregisterVCS("mock")
	retrievedVCS = GetVCS("mock")
	if retrievedVCS != nil {
		t.Error("Expected VCS to be unregistered using global function")
	}
}
