package emit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/benjaminabbitt/versionator/internal/version"
)

func TestRender_Python(t *testing.T) {
	result, err := Render(FormatPython, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `__version__ = "1.2.3"`) {
		t.Errorf("expected Python version string, got: %s", result)
	}
	if !strings.Contains(result, "versionator") {
		t.Errorf("expected versionator mention in comment, got: %s", result)
	}
}

func TestRender_JSON(t *testing.T) {
	result, err := Render(FormatJSON, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `"version": "1.2.3"`) {
		t.Errorf("expected JSON version string, got: %s", result)
	}
}

func TestRender_YAML(t *testing.T) {
	result, err := Render(FormatYAML, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `version: "1.2.3"`) {
		t.Errorf("expected YAML version string, got: %s", result)
	}
}

func TestRender_Go(t *testing.T) {
	result, err := Render(FormatGo, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `Version     = "1.2.3"`) {
		t.Errorf("expected Go version string, got: %s", result)
	}
}

func TestRender_C(t *testing.T) {
	result, err := Render(FormatC, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `#define VERSION "1.2.3"`) {
		t.Errorf("expected C version define, got: %s", result)
	}
}

func TestRender_CHeader(t *testing.T) {
	result, err := Render(FormatCHeader, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `#define VERSION "1.2.3"`) {
		t.Errorf("expected C header version define, got: %s", result)
	}
	if !strings.Contains(result, "extern const char* VERSION_STRING") {
		t.Errorf("expected extern declaration, got: %s", result)
	}
}

func TestRender_CPP(t *testing.T) {
	result, err := Render(FormatCPP, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `#define VERSION "1.2.3"`) {
		t.Errorf("expected C++ version define, got: %s", result)
	}
	if !strings.Contains(result, "namespace version") {
		t.Errorf("expected namespace, got: %s", result)
	}
}

func TestRender_CPPHeader(t *testing.T) {
	result, err := Render(FormatCPPHeader, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `#define VERSION "1.2.3"`) {
		t.Errorf("expected C++ header version define, got: %s", result)
	}
	if !strings.Contains(result, "extern const char* VERSION_STRING") {
		t.Errorf("expected extern declaration, got: %s", result)
	}
}

func TestRender_JS(t *testing.T) {
	result, err := Render(FormatJS, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `export const VERSION = "1.2.3"`) {
		t.Errorf("expected JS version string, got: %s", result)
	}
}

func TestRender_TS(t *testing.T) {
	result, err := Render(FormatTS, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `export const VERSION: string = "1.2.3"`) {
		t.Errorf("expected TS version string, got: %s", result)
	}
}

func TestRender_Java(t *testing.T) {
	result, err := Render(FormatJava, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public static final String VERSION = "1.2.3"`) {
		t.Errorf("expected Java version string, got: %s", result)
	}
}

func TestRender_Kotlin(t *testing.T) {
	result, err := Render(FormatKotlin, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `const val VERSION = "1.2.3"`) {
		t.Errorf("expected Kotlin version string, got: %s", result)
	}
}

func TestRender_CSharp(t *testing.T) {
	result, err := Render(FormatCSharp, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public const string Version = "1.2.3"`) {
		t.Errorf("expected C# version string, got: %s", result)
	}
}

func TestRender_PHP(t *testing.T) {
	result, err := Render(FormatPHP, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `const VERSION = '1.2.3'`) {
		t.Errorf("expected PHP version string, got: %s", result)
	}
}

func TestRender_Swift(t *testing.T) {
	result, err := Render(FormatSwift, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public let VERSION = "1.2.3"`) {
		t.Errorf("expected Swift version string, got: %s", result)
	}
}

func TestRender_Ruby(t *testing.T) {
	result, err := Render(FormatRuby, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `VERSION = "1.2.3"`) {
		t.Errorf("expected Ruby version string, got: %s", result)
	}
}

func TestRender_Rust(t *testing.T) {
	result, err := Render(FormatRust, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `pub const VERSION: &str = "1.2.3"`) {
		t.Errorf("expected Rust version string, got: %s", result)
	}
}

func TestRender_InvalidFormat(t *testing.T) {
	_, err := Render(Format("invalid"), "1.2.3")
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestRenderTemplate_BasicVersionComponents(t *testing.T) {
	template := `{{Major}}.{{Minor}}.{{Patch}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got: %s", result)
	}
}

func TestRenderTemplate_MajorMinorPatch(t *testing.T) {
	template := `{{MajorMinorPatch}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got: %s", result)
	}
}

func TestRenderTemplate_MajorMinor(t *testing.T) {
	template := `{{MajorMinor}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2" {
		t.Errorf("expected 1.2, got: %s", result)
	}
}

func TestRenderTemplate_Prefix(t *testing.T) {
	template := `{{Prefix}}{{MajorMinorPatch}}`
	result, err := RenderTemplate(template, "v1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "v1.2.3" {
		t.Errorf("expected v1.2.3, got: %s", result)
	}
}

func TestRenderTemplate_BuildDateComponents(t *testing.T) {
	template := `{{BuildYear}}-{{BuildMonth}}-{{BuildDay}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be in YYYY-MM-DD format
	if len(result) != 10 {
		t.Errorf("expected date length 10, got: %d (%s)", len(result), result)
	}
}

func TestRenderTemplate_BuildDateUTC(t *testing.T) {
	template := `{{BuildDateUTC}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be in YYYY-MM-DD format
	if len(result) != 10 {
		t.Errorf("expected date length 10, got: %d (%s)", len(result), result)
	}
}

func TestRenderTemplate_BuildDateTimeCompact(t *testing.T) {
	template := `{{BuildDateTimeCompact}}`
	result, err := RenderTemplate(template, "1.2.3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be in YYYYMMDDHHmmss format (14 chars)
	if len(result) != 14 {
		t.Errorf("expected compact datetime length 14, got: %d (%s)", len(result), result)
	}
}

func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	template := `{{Invalid`
	_, err := RenderTemplate(template, "1.2.3")
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestGetEmbeddedTemplate(t *testing.T) {
	template, err := GetEmbeddedTemplate(FormatPython)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(template, "{{MajorMinorPatch}}") {
		t.Errorf("expected template placeholder, got: %s", template)
	}
	if !strings.Contains(template, "__version__") {
		t.Errorf("expected Python syntax in template, got: %s", template)
	}
}

func TestGetEmbeddedTemplate_InvalidFormat(t *testing.T) {
	_, err := GetEmbeddedTemplate(Format("invalid"))
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

func TestIsValidFormat(t *testing.T) {
	tests := []struct {
		format   string
		expected bool
	}{
		{"python", true},
		{"json", true},
		{"yaml", true},
		{"go", true},
		{"c", true},
		{"c-header", true},
		{"cpp", true},
		{"cpp-header", true},
		{"js", true},
		{"ts", true},
		{"java", true},
		{"kotlin", true},
		{"csharp", true},
		{"php", true},
		{"swift", true},
		{"ruby", true},
		{"rust", true},
		{"invalid", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := IsValidFormat(tt.format)
			if result != tt.expected {
				t.Errorf("IsValidFormat(%s) = %v, expected %v", tt.format, result, tt.expected)
			}
		})
	}
}

func TestSupportedFormats(t *testing.T) {
	formats := SupportedFormats()
	if len(formats) == 0 {
		t.Error("expected at least one supported format")
	}

	// Check that python is in the list
	found := false
	for _, f := range formats {
		if f == "python" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected python to be in supported formats")
	}
}

func TestWriteToFile(t *testing.T) {
	// Create a temp directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := "test content"
	err := WriteToFile(content, tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read back and verify
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected %s, got: %s", content, string(data))
	}
}

func TestEmitToFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "version.py")

	err := EmitToFile(FormatPython, "1.2.3", tmpFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify file exists and has content
	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(data), `__version__ = "1.2.3"`) {
		t.Errorf("expected Python version string in file, got: %s", string(data))
	}
}

func TestBuildTemplateDataFromVersion(t *testing.T) {
	vd := &version.Version{
		Prefix: "v",
		Major:  1,
		Minor:  2,
		Patch:  3,
	}

	data := BuildTemplateDataFromVersion(vd)

	if data.Major != "1" {
		t.Errorf("expected Major=1, got %s", data.Major)
	}
	if data.Minor != "2" {
		t.Errorf("expected Minor=2, got %s", data.Minor)
	}
	if data.Patch != "3" {
		t.Errorf("expected Patch=3, got %s", data.Patch)
	}
	if data.MajorMinorPatch != "1.2.3" {
		t.Errorf("expected MajorMinorPatch=1.2.3, got %s", data.MajorMinorPatch)
	}
	if data.MajorMinor != "1.2" {
		t.Errorf("expected MajorMinor=1.2, got %s", data.MajorMinor)
	}
	if data.Prefix != "v" {
		t.Errorf("expected Prefix=v, got %s", data.Prefix)
	}
}

func TestBuildCompleteTemplateData(t *testing.T) {
	vd := &version.Version{
		Major: 1,
		Minor: 2,
		Patch: 3,
	}

	// Templates now use single strings with separators included
	prereleaseTemplate := "alpha-1"
	metadataTemplate := "build.123"

	data := BuildCompleteTemplateData(vd, prereleaseTemplate, metadataTemplate)

	if data.PreRelease != "alpha-1" {
		t.Errorf("expected PreRelease=alpha-1, got %s", data.PreRelease)
	}
	if data.PreReleaseWithDash != "-alpha-1" {
		t.Errorf("expected PreReleaseWithDash=-alpha-1, got %s", data.PreReleaseWithDash)
	}
	// Metadata should contain the dot separator
	if data.Metadata != "build.123" {
		t.Errorf("expected Metadata=build.123, got %s", data.Metadata)
	}
	if data.MetadataWithPlus != "+build.123" {
		t.Errorf("expected MetadataWithPlus=+build.123, got %s", data.MetadataWithPlus)
	}
}

func TestBuildCompleteTemplateData_EmptyTemplates(t *testing.T) {
	vd := &version.Version{
		Major: 1,
		Minor: 2,
		Patch: 3,
	}

	data := BuildCompleteTemplateData(vd, "", "")

	if data.PreRelease != "" {
		t.Errorf("expected empty PreRelease, got %s", data.PreRelease)
	}
	if data.PreReleaseWithDash != "" {
		t.Errorf("expected empty PreReleaseWithDash, got %s", data.PreReleaseWithDash)
	}
	if data.Metadata != "" {
		t.Errorf("expected empty Metadata, got %s", data.Metadata)
	}
	if data.MetadataWithPlus != "" {
		t.Errorf("expected empty MetadataWithPlus, got %s", data.MetadataWithPlus)
	}
}

func TestRenderTemplateList(t *testing.T) {
	vd := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := BuildTemplateDataFromVersion(vd)

	templates := []string{"a", "b", "c"}
	result := RenderTemplateList(templates, data, "-")

	if result != "a-b-c" {
		t.Errorf("expected a-b-c, got %s", result)
	}
}

func TestRenderTemplateList_DotSeparator(t *testing.T) {
	vd := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := BuildTemplateDataFromVersion(vd)

	templates := []string{"{{Major}}", "{{Minor}}", "{{Patch}}"}
	result := RenderTemplateList(templates, data, ".")

	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", result)
	}
}

func TestDirtyFlag(t *testing.T) {
	if dirtyFlag(0) != "" {
		t.Error("expected empty string for 0 uncommitted changes")
	}
	if dirtyFlag(1) != "dirty" {
		t.Error("expected 'dirty' for 1 uncommitted change")
	}
	if dirtyFlag(100) != "dirty" {
		t.Error("expected 'dirty' for 100 uncommitted changes")
	}
}

func TestValidateOutputPath_EmptyPath(t *testing.T) {
	err := ValidateOutputPath("")
	if err == nil {
		t.Error("expected error for empty path")
	}
	if !strings.Contains(err.Error(), ErrOutputPathEmpty) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathEmpty, err)
	}
}

func TestValidateOutputPath_IsDirectory(t *testing.T) {
	tmpDir := t.TempDir()

	err := ValidateOutputPath(tmpDir)
	if err == nil {
		t.Error("expected error when path is a directory")
	}
	if !strings.Contains(err.Error(), ErrOutputPathIsDirectory) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathIsDirectory, err)
	}
}

func TestValidateOutputPath_ExistingFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existing.txt")

	// Create the file
	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := ValidateOutputPath(tmpFile)
	if err != nil {
		t.Errorf("expected no error for existing file, got: %v", err)
	}
}

func TestValidateOutputPath_NewFileInExistingDir(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new.txt")

	err := ValidateOutputPath(tmpFile)
	if err != nil {
		t.Errorf("expected no error for new file in existing directory, got: %v", err)
	}
}

func TestValidateOutputPath_ParentNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nonexistent", "file.txt")

	err := ValidateOutputPath(tmpFile)
	if err == nil {
		t.Error("expected error when parent directory doesn't exist")
	}
	if !strings.Contains(err.Error(), ErrParentDirNotExist) {
		t.Errorf("expected error containing %q, got: %v", ErrParentDirNotExist, err)
	}
}

func TestValidateOutputPath_ParentIsFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")

	// Create a file that will act as a "parent"
	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Try to use the file as a parent directory
	invalidPath := filepath.Join(tmpFile, "child.txt")

	err := ValidateOutputPath(invalidPath)
	if err == nil {
		t.Error("expected error when parent is a file, not a directory")
	}
}

func TestValidateOutputPath_FileInCurrentDir(t *testing.T) {
	// File in current directory (no path separator)
	err := ValidateOutputPath("simple.txt")
	if err != nil {
		t.Errorf("expected no error for simple filename, got: %v", err)
	}
}

func TestWriteToFile_ValidatesPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Try to write to a directory
	err := WriteToFile("content", tmpDir)
	if err == nil {
		t.Error("expected error when writing to a directory")
	}
	if !strings.Contains(err.Error(), ErrOutputPathIsDirectory) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathIsDirectory, err)
	}
}

func TestWriteToFile_EmptyPath(t *testing.T) {
	err := WriteToFile("content", "")
	if err == nil {
		t.Error("expected error for empty path")
	}
	if !strings.Contains(err.Error(), ErrOutputPathEmpty) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathEmpty, err)
	}
}
