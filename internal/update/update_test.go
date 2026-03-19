package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/emit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func newTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	require.NoError(t, err)
	return logger
}

func TestUpdater_UpdateFiles_SingleFile_Success(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "package.json")
	content := `{"name": "myapp", "version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "version",
			Template: "{{MajorMinorPatch}}",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{
		Major:           "2",
		Minor:           "0",
		Patch:           "0",
		MajorMinorPatch: "2.0.0",
	}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)

	// Verify file was updated
	parser := NewDaselFileParser()
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := readBack.(map[string]any)
	assert.Equal(t, "2.0.0", dataMap["version"])
	assert.Equal(t, "myapp", dataMap["name"])
}

func TestUpdater_UpdateFiles_MultipleFiles_AllUpdated(t *testing.T) {
	tmpDir := t.TempDir()

	// Create package.json
	pkgPath := filepath.Join(tmpDir, "package.json")
	require.NoError(t, os.WriteFile(pkgPath, []byte(`{"version": "1.0.0"}`), 0644))

	// Create Chart.yaml
	chartPath := filepath.Join(tmpDir, "Chart.yaml")
	require.NoError(t, os.WriteFile(chartPath, []byte("version: 1.0.0\nappVersion: 1.0.0\n"), 0644))

	configs := []config.UpdateConfig{
		{File: pkgPath, Path: "version", Template: "{{MajorMinorPatch}}"},
		{File: chartPath, Path: "version", Template: "{{MajorMinorPatch}}"},
		{File: chartPath, Path: "appVersion", Template: "{{MajorMinorPatch}}"},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "3.0.0"}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)

	parser := NewDaselFileParser()

	// Verify package.json
	pkgData, _, err := parser.Read(pkgPath)
	require.NoError(t, err)
	assert.Equal(t, "3.0.0", pkgData.(map[string]any)["version"])

	// Verify Chart.yaml
	chartData, _, err := parser.Read(chartPath)
	require.NoError(t, err)
	chartMap := chartData.(map[string]any)
	assert.Equal(t, "3.0.0", chartMap["version"])
	assert.Equal(t, "3.0.0", chartMap["appVersion"])
}

func TestUpdater_UpdateFiles_TOML_CargoToml(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Cargo.toml")
	content := `[package]
name = "myapp"
version = "1.0.0"
`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "package.version",
			Template: "{{MajorMinorPatch}}",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "2.5.0"}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)

	parser := NewDaselFileParser()
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	pkg := readBack.(map[string]any)["package"].(map[string]any)
	assert.Equal(t, "2.5.0", pkg["version"])
	assert.Equal(t, "myapp", pkg["name"])
}

func TestUpdater_UpdateFiles_WithPreRelease(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Chart.yaml")
	content := `version: 1.0.0
appVersion: 1.0.0
`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "appVersion",
			Template: "{{MajorMinorPatch}}{{PreReleaseWithDash}}",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{
		MajorMinorPatch:    "2.0.0",
		PreRelease:         "alpha.1",
		PreReleaseWithDash: "-alpha.1",
	}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)

	parser := NewDaselFileParser()
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	assert.Equal(t, "2.0.0-alpha.1", readBack.(map[string]any)["appVersion"])
}

func TestUpdater_UpdateFiles_FileNotFound_ReturnsError(t *testing.T) {
	configs := []config.UpdateConfig{
		{
			File:     "/nonexistent/file.json",
			Path:     "version",
			Template: "{{MajorMinorPatch}}",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "1.0.0"}

	err := updater.UpdateFiles(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrFileNotFound)
}

func TestUpdater_UpdateFiles_PathNotFound_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(filePath, []byte(`{"name": "test"}`), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "nonexistent.path",
			Template: "{{MajorMinorPatch}}",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "1.0.0"}

	err := updater.UpdateFiles(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrInvalidSelector)
}

func TestUpdater_UpdateFiles_InvalidTemplate_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(filePath, []byte(`{"version": "1.0.0"}`), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "version",
			Template: "{{Invalid",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "1.0.0"}

	err := updater.UpdateFiles(data)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrTemplateRender)
}

func TestUpdater_GetFilesToCommit_ReturnsUpdatedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "a.json")
	file2 := filepath.Join(tmpDir, "b.yaml")
	require.NoError(t, os.WriteFile(file1, []byte(`{"version": "1.0.0"}`), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("version: 1.0.0\n"), 0644))

	configs := []config.UpdateConfig{
		{File: file1, Path: "version", Template: "{{MajorMinorPatch}}"},
		{File: file2, Path: "version", Template: "{{MajorMinorPatch}}"},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "2.0.0"}

	// Before update, should be empty
	assert.Empty(t, updater.GetFilesToCommit())

	err := updater.UpdateFiles(data)
	require.NoError(t, err)

	// After update, should contain both files
	files := updater.GetFilesToCommit()
	assert.Len(t, files, 2)
	assert.Contains(t, files, file1)
	assert.Contains(t, files, file2)
}

func TestUpdater_ValidateConfig_AllFilesExist_Success(t *testing.T) {
	tmpDir := t.TempDir()
	file1 := filepath.Join(tmpDir, "test.json")
	require.NoError(t, os.WriteFile(file1, []byte(`{"version": "1.0.0"}`), 0644))

	configs := []config.UpdateConfig{
		{File: file1, Path: "version", Template: "{{MajorMinorPatch}}"},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))

	err := updater.ValidateConfig()

	require.NoError(t, err)
}

func TestUpdater_ValidateConfig_FileMissing_ReturnsError(t *testing.T) {
	configs := []config.UpdateConfig{
		{File: "/nonexistent/file.json", Path: "version", Template: "{{MajorMinorPatch}}"},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))

	err := updater.ValidateConfig()

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrFileNotFound)
}

func TestUpdater_UpdateFiles_ExplicitFormat_Success(t *testing.T) {
	tmpDir := t.TempDir()
	// File with .txt extension but contains JSON
	filePath := filepath.Join(tmpDir, "version.txt")
	require.NoError(t, os.WriteFile(filePath, []byte(`{"version": "1.0.0"}`), 0644))

	configs := []config.UpdateConfig{
		{
			File:     filePath,
			Path:     "version",
			Template: "{{MajorMinorPatch}}",
			Format:   "json",
		},
	}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "2.0.0"}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)

	parser := NewDaselFileParser()
	readBack, _, err := parser.ReadWithFormat(filePath, "json")
	require.NoError(t, err)
	assert.Equal(t, "2.0.0", readBack.(map[string]any)["version"])
}

func TestUpdater_UpdateFiles_EmptyConfigs_Success(t *testing.T) {
	configs := []config.UpdateConfig{}

	updater := NewUpdater(configs, NewDaselFileParser(), newTestLogger(t))
	data := emit.TemplateData{MajorMinorPatch: "1.0.0"}

	err := updater.UpdateFiles(data)

	require.NoError(t, err)
	assert.Empty(t, updater.GetFilesToCommit())
}
