package loader

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"go.uber.org/zap"
)

func TestNewLoader_Always_InitializesEmptyMaps(t *testing.T) {
	logger := zap.NewNop()
	loader := NewLoader(logger)

	if loader.EmitPlugins == nil {
		t.Error("EmitPlugins should be initialized")
	}
	if loader.BuildPlugins == nil {
		t.Error("BuildPlugins should be initialized")
	}
	if loader.PatchPlugins == nil {
		t.Error("PatchPlugins should be initialized")
	}
}

func TestIsExecutable_ExecutableFileUnix_ReturnsTrue(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	tmpDir := t.TempDir()
	execFile := filepath.Join(tmpDir, "test-exec")
	if err := os.WriteFile(execFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create executable: %v", err)
	}

	if !isExecutable(execFile) {
		t.Error("expected file with 0755 to be executable")
	}
}

func TestIsExecutable_NonExecutableFileUnix_ReturnsFalse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	tmpDir := t.TempDir()
	nonExecFile := filepath.Join(tmpDir, "test-noexec")
	if err := os.WriteFile(nonExecFile, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if isExecutable(nonExecFile) {
		t.Error("expected file with 0644 to not be executable")
	}
}

func TestIsExecutable_Directory_ReturnsFalse(t *testing.T) {
	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if isExecutable(subDir) {
		t.Error("expected directory to not be executable")
	}
}

func TestIsExecutable_NonExistentFile_ReturnsFalse(t *testing.T) {
	if isExecutable("/nonexistent/path/to/file") {
		t.Error("expected nonexistent file to not be executable")
	}
}

func TestIsExecutable_ExeExtensionWindows_ReturnsTrue(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tmpDir := t.TempDir()
	exeFile := filepath.Join(tmpDir, "test.exe")
	if err := os.WriteFile(exeFile, []byte("MZ"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !isExecutable(exeFile) {
		t.Error("expected .exe file to be executable on Windows")
	}
}

func TestIsExecutable_ComExtensionWindows_ReturnsTrue(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tmpDir := t.TempDir()
	comFile := filepath.Join(tmpDir, "test.com")
	if err := os.WriteFile(comFile, []byte("MZ"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !isExecutable(comFile) {
		t.Error("expected .com file to be executable on Windows")
	}
}

func TestIsExecutable_NonExecutableExtensionWindows_ReturnsFalse(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows-specific test")
	}

	tmpDir := t.TempDir()
	txtFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(txtFile, []byte("data"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if isExecutable(txtFile) {
		t.Error("expected .txt file to not be executable on Windows")
	}
}

func TestWindowsExecutableExtensions_Always_ContainsExeAndCom(t *testing.T) {
	expectedExtensions := []string{".exe", ".com"}

	for _, ext := range expectedExtensions {
		if _, ok := windowsExecutableExtensions[ext]; !ok {
			t.Errorf("expected windowsExecutableExtensions to contain %s", ext)
		}
	}
}

func TestWindowsExecutableExtensions_Always_ExcludesScripts(t *testing.T) {
	scriptExtensions := []string{".bat", ".cmd", ".ps1", ".vbs", ".js"}

	for _, ext := range scriptExtensions {
		if _, ok := windowsExecutableExtensions[ext]; ok {
			t.Errorf("expected windowsExecutableExtensions to NOT contain %s (script)", ext)
		}
	}
}

func TestLoaderClose_NoClients_DoesNotPanic(t *testing.T) {
	logger := zap.NewNop()
	loader := NewLoader(logger)

	loader.Close()

	if loader.clients != nil {
		t.Error("expected clients to be nil after Close")
	}
}

func TestGetPluginDirs_Always_ReturnsNonEmpty(t *testing.T) {
	dirs := getPluginDirs()

	if len(dirs) == 0 {
		t.Error("expected at least one plugin directory")
	}
}

func TestGetPluginDirs_EnvVarSet_IncludesEnvVarFirst(t *testing.T) {
	testDir := "/test/plugin/dir"
	t.Setenv("VERSIONATOR_PLUGIN_DIR", testDir)

	dirs := getPluginDirs()

	if len(dirs) == 0 || dirs[0] != testDir {
		t.Errorf("expected first dir to be %s from env var, got %v", testDir, dirs)
	}
}

func TestDiscoverAndLoad_NonExistentDir_ReturnsZero(t *testing.T) {
	// Save and restore the original function
	originalFunc := getPluginDirsFunc
	t.Cleanup(func() { getPluginDirsFunc = originalFunc })

	// Mock to return only a nonexistent directory
	getPluginDirsFunc = func() []string {
		return []string{"/nonexistent/plugin/dir"}
	}

	logger := zap.NewNop()
	loader := NewLoader(logger)

	count, err := loader.DiscoverAndLoad()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins loaded, got %d", count)
	}
}

func TestDiscoverAndLoad_EmptyDir_ReturnsZero(t *testing.T) {
	tmpDir := t.TempDir()

	// Save and restore the original function
	originalFunc := getPluginDirsFunc
	t.Cleanup(func() { getPluginDirsFunc = originalFunc })

	// Mock to return only the temp directory
	getPluginDirsFunc = func() []string {
		return []string{tmpDir}
	}

	logger := zap.NewNop()
	loader := NewLoader(logger)

	count, err := loader.DiscoverAndLoad()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins loaded, got %d", count)
	}
}

func TestLoadFromDir_EmptyDir_ReturnsZero(t *testing.T) {
	logger := zap.NewNop()
	loader := NewLoader(logger)

	tmpDir := t.TempDir()

	count, err := loader.loadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins loaded, got %d", count)
	}
}

func TestLoadFromDir_DirectoryEntry_SkipsDirectory(t *testing.T) {
	logger := zap.NewNop()
	loader := NewLoader(logger)

	tmpDir := t.TempDir()
	subDir := filepath.Join(tmpDir, "versionator-plugin-emit-test")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdirectory: %v", err)
	}

	count, err := loader.loadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins loaded (subdirs should be skipped), got %d", count)
	}
}

func TestLoadFromDir_NonPluginFile_SkipsFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Unix-specific test")
	}

	logger := zap.NewNop()
	loader := NewLoader(logger)

	tmpDir := t.TempDir()

	nonPluginFile := filepath.Join(tmpDir, "some-other-binary")
	if err := os.WriteFile(nonPluginFile, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	count, err := loader.loadFromDir(tmpDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 plugins loaded (non-plugin files should be skipped), got %d", count)
	}
}
