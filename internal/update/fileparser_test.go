package update

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDaselFileParser_Read_JSON_ParsesCorrectly(t *testing.T) {
	// Create temp file
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	content := `{"name": "test", "version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)

	require.NoError(t, err)
	assert.Equal(t, FormatJSON, format)
	dataMap := data.(map[string]any)
	assert.Equal(t, "test", dataMap["name"])
	assert.Equal(t, "1.0.0", dataMap["version"])
}

func TestDaselFileParser_Read_YAML_ParsesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yaml")
	content := `name: test
version: 1.0.0`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)

	require.NoError(t, err)
	assert.Equal(t, FormatYAML, format)
	dataMap := data.(map[string]any)
	assert.Equal(t, "test", dataMap["name"])
	assert.Equal(t, "1.0.0", dataMap["version"])
}

func TestDaselFileParser_Read_TOML_ParsesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.toml")
	content := `name = "test"
version = "1.0.0"`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)

	require.NoError(t, err)
	assert.Equal(t, FormatTOML, format)
	dataMap := data.(map[string]any)
	assert.Equal(t, "test", dataMap["name"])
	assert.Equal(t, "1.0.0", dataMap["version"])
}

func TestDaselFileParser_Read_YML_Extension_ParsesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.yml")
	content := `version: 2.0.0`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)

	require.NoError(t, err)
	assert.Equal(t, FormatYAML, format)
	dataMap := data.(map[string]any)
	assert.Equal(t, "2.0.0", dataMap["version"])
}

func TestDaselFileParser_Read_FileNotFound_ReturnsError(t *testing.T) {
	parser := NewDaselFileParser()
	_, _, err := parser.Read("/nonexistent/file.json")

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrFileNotFound)
}

func TestDaselFileParser_Read_UnsupportedFormat_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.txt")
	require.NoError(t, os.WriteFile(filePath, []byte("content"), 0644))

	parser := NewDaselFileParser()
	_, _, err := parser.Read(filePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrUnsupportedFormat)
}

func TestDaselFileParser_Read_InvalidJSON_ReturnsError(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	content := `{invalid json`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	_, _, err := parser.Read(filePath)

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrFileParseFailed)
}

func TestDaselFileParser_ReadWithFormat_ExplicitOverride(t *testing.T) {
	tmpDir := t.TempDir()
	// File with .txt extension but contains JSON
	filePath := filepath.Join(tmpDir, "data.txt")
	content := `{"version": "3.0.0"}`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.ReadWithFormat(filePath, "json")

	require.NoError(t, err)
	assert.Equal(t, FormatJSON, format)
	dataMap := data.(map[string]any)
	assert.Equal(t, "3.0.0", dataMap["version"])
}

func TestDaselFileParser_Write_JSON_WritesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.json")

	parser := NewDaselFileParser()
	data := map[string]any{"name": "test", "version": "1.0.0"}
	err := parser.Write(filePath, data, FormatJSON)

	require.NoError(t, err)

	// Read back and verify
	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), `"name": "test"`)
	assert.Contains(t, string(content), `"version": "1.0.0"`)
}

func TestDaselFileParser_Write_YAML_WritesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.yaml")

	parser := NewDaselFileParser()
	data := map[string]any{"name": "test", "version": "1.0.0"}
	err := parser.Write(filePath, data, FormatYAML)

	require.NoError(t, err)

	content, err := os.ReadFile(filePath)
	require.NoError(t, err)
	assert.Contains(t, string(content), "name: test")
	assert.Contains(t, string(content), "version: 1.0.0")
}

func TestDaselFileParser_Write_TOML_WritesCorrectly(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "output.toml")

	parser := NewDaselFileParser()
	data := map[string]any{"name": "test", "version": "1.0.0"}
	err := parser.Write(filePath, data, FormatTOML)

	require.NoError(t, err)

	// Read back and verify values (TOML may use single or double quotes)
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	assert.Equal(t, "test", readBackMap["name"])
	assert.Equal(t, "1.0.0", readBackMap["version"])
}

func TestDaselFileParser_Select_SimpleKey_ReturnsValue(t *testing.T) {
	parser := NewDaselFileParser()
	data := map[string]any{"version": "1.2.3"}

	result, err := parser.Select(data, "version")

	require.NoError(t, err)
	assert.Equal(t, "1.2.3", result)
}

func TestDaselFileParser_Select_NestedKey_ReturnsValue(t *testing.T) {
	parser := NewDaselFileParser()
	data := map[string]any{
		"package": map[string]any{
			"version": "2.0.0",
		},
	}

	result, err := parser.Select(data, "package.version")

	require.NoError(t, err)
	assert.Equal(t, "2.0.0", result)
}

func TestDaselFileParser_Select_NonexistentPath_ReturnsError(t *testing.T) {
	parser := NewDaselFileParser()
	data := map[string]any{"name": "test"}

	_, err := parser.Select(data, "nonexistent.path")

	require.Error(t, err)
	assert.Contains(t, err.Error(), ErrPathNotFound)
}

func TestDaselFileParser_Put_SimpleKey_UpdatesValue(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")
	content := `{"version": "1.0.0"}`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := data.(map[string]any)

	err = parser.Put(&dataMap, "version", "2.0.0")
	require.NoError(t, err)

	// Write back and verify
	err = parser.Write(filePath, dataMap, format)
	require.NoError(t, err)

	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	assert.Equal(t, "2.0.0", readBackMap["version"])
}

func TestDaselFileParser_Put_NestedKey_UpdatesValue(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.toml")
	content := `[package]
name = "test"
version = "1.0.0"
`
	require.NoError(t, os.WriteFile(filePath, []byte(content), 0644))

	parser := NewDaselFileParser()
	data, format, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := data.(map[string]any)

	err = parser.Put(&dataMap, "package.version", "3.0.0")
	require.NoError(t, err)

	// Write back and verify
	err = parser.Write(filePath, dataMap, format)
	require.NoError(t, err)

	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	pkg := readBackMap["package"].(map[string]any)
	assert.Equal(t, "3.0.0", pkg["version"])
}

func TestDaselFileParser_RoundTrip_JSON(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "package.json")
	original := `{
  "name": "myapp",
  "version": "1.0.0",
  "dependencies": {}
}`
	require.NoError(t, os.WriteFile(filePath, []byte(original), 0644))

	parser := NewDaselFileParser()

	// Read
	data, format, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := data.(map[string]any)

	// Modify
	err = parser.Put(&dataMap, "version", "2.0.0")
	require.NoError(t, err)

	// Write
	err = parser.Write(filePath, dataMap, format)
	require.NoError(t, err)

	// Verify
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	assert.Equal(t, "2.0.0", readBackMap["version"])
	assert.Equal(t, "myapp", readBackMap["name"])
}

func TestDaselFileParser_RoundTrip_TOML_Cargo(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Cargo.toml")
	original := `[package]
name = "myapp"
version = "1.0.0"
`
	require.NoError(t, os.WriteFile(filePath, []byte(original), 0644))

	parser := NewDaselFileParser()

	// Read
	data, format, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := data.(map[string]any)

	// Modify
	err = parser.Put(&dataMap, "package.version", "2.0.0")
	require.NoError(t, err)

	// Write
	err = parser.Write(filePath, dataMap, format)
	require.NoError(t, err)

	// Verify
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	pkg := readBackMap["package"].(map[string]any)
	assert.Equal(t, "2.0.0", pkg["version"])
	assert.Equal(t, "myapp", pkg["name"])
}

func TestDaselFileParser_RoundTrip_YAML_HelmChart(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "Chart.yaml")
	original := `apiVersion: v2
name: myapp
version: 1.0.0
appVersion: 1.0.0
`
	require.NoError(t, os.WriteFile(filePath, []byte(original), 0644))

	parser := NewDaselFileParser()

	// Read
	data, format, err := parser.Read(filePath)
	require.NoError(t, err)
	dataMap := data.(map[string]any)

	// Modify both version and appVersion
	err = parser.Put(&dataMap, "version", "2.0.0")
	require.NoError(t, err)
	err = parser.Put(&dataMap, "appVersion", "2.0.0")
	require.NoError(t, err)

	// Write
	err = parser.Write(filePath, dataMap, format)
	require.NoError(t, err)

	// Verify
	readBack, _, err := parser.Read(filePath)
	require.NoError(t, err)
	readBackMap := readBack.(map[string]any)
	assert.Equal(t, "2.0.0", readBackMap["version"])
	assert.Equal(t, "2.0.0", readBackMap["appVersion"])
	assert.Equal(t, "myapp", readBackMap["name"])
}
