package emit

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/benjaminabbitt/versionator/internal/vcs"
	gitVCS "github.com/benjaminabbitt/versionator/internal/vcs/git"
	"github.com/benjaminabbitt/versionator/internal/vcs/mock"
	"github.com/benjaminabbitt/versionator/internal/version"
	"github.com/golang/mock/gomock"
)

// =============================================================================
// CORE FUNCTIONALITY
// Tests demonstrating the primary purpose of the emit package: rendering
// version information into language-specific output files.
// =============================================================================

// TestRender_Python validates that Render produces correct Python version output.
//
// Why: Python is a common target format and serves as the canonical example of
// template rendering. Ensures the core rendering pipeline works end-to-end.
//
// What: Given version "1.2.3" and Python format, the output should contain
// the __version__ assignment and versionator attribution comment.
func TestRender_Python(t *testing.T) {
	// Precondition: Version string "1.2.3" and Python format
	// Action: Render with FormatPython
	result, err := Render(FormatPython, "1.2.3")

	// Expected: No error, output contains Python version syntax
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

// TestRenderTemplate_BasicVersionComponents validates that template variables
// Major, Minor, and Patch are correctly populated from a version string.
//
// Why: These are the fundamental building blocks of semantic versioning.
// All other version derivations depend on correct parsing of these components.
//
// What: A template using {{Major}}.{{Minor}}.{{Patch}} should produce "1.2.3"
// when given version "1.2.3".
func TestRenderTemplate_BasicVersionComponents(t *testing.T) {
	// Precondition: Template using individual version components
	template := `{{Major}}.{{Minor}}.{{Patch}}`

	// Action: Render with version "1.2.3"
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: Output is "1.2.3"
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got: %s", result)
	}
}

// TestRenderTemplate_MajorMinorPatch validates the convenience variable that
// combines all three version components.
//
// Why: Many use cases need the full version string without manual concatenation.
// This variable simplifies templates and reduces error opportunities.
//
// What: {{MajorMinorPatch}} should produce "1.2.3" for version "1.2.3".
func TestRenderTemplate_MajorMinorPatch(t *testing.T) {
	// Precondition: Template using MajorMinorPatch shorthand
	template := `{{MajorMinorPatch}}`

	// Action: Render with version "1.2.3"
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: Output is "1.2.3"
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got: %s", result)
	}
}

// TestBuildTemplateDataFromVersion validates that Version structs are correctly
// converted to TemplateData for rendering.
//
// Why: This function bridges the internal version representation with the
// template system. Incorrect conversion would cause all template rendering to fail.
//
// What: A Version with Major=1, Minor=2, Patch=3, Prefix="v" should produce
// TemplateData with matching string fields and computed MajorMinor/MajorMinorPatch.
func TestBuildTemplateDataFromVersion(t *testing.T) {
	// Precondition: A Version struct with known values
	vd := &version.Version{
		Prefix: "v",
		Major:  1,
		Minor:  2,
		Patch:  3,
	}

	// Action: Build template data
	data := BuildTemplateDataFromVersion(vd)

	// Expected: All fields populated correctly
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

// TestEmitToFile validates the complete workflow of rendering and writing
// version output to a file.
//
// Why: This is the primary use case - generating version files that can be
// included in projects. File I/O errors would break the entire tool.
//
// What: EmitToFile should create a file containing the rendered Python version.
func TestEmitToFile(t *testing.T) {
	// Precondition: Temporary directory for output file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "version.py")

	// Action: Emit Python format to file
	err := EmitToFile(FormatPython, "1.2.3", tmpFile)

	// Expected: File exists and contains Python version string
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(data), `__version__ = "1.2.3"`) {
		t.Errorf("expected Python version string in file, got: %s", string(data))
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests covering important alternate flows: different output formats,
// template features, and data transformations.
// =============================================================================

// TestRender_JSON validates JSON format output.
//
// Why: JSON is a universal data format used by many build systems and
// package managers (e.g., package.json). Format correctness is critical.
//
// What: Render should produce valid JSON with version field.
func TestRender_JSON(t *testing.T) {
	// Precondition: Version string and JSON format
	// Action: Render
	result, err := Render(FormatJSON, "1.2.3")

	// Expected: JSON-formatted version string
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `"version": "1.2.3"`) {
		t.Errorf("expected JSON version string, got: %s", result)
	}
}

// TestRender_YAML validates YAML format output.
//
// Why: YAML is common in configuration files and CI/CD pipelines.
// Correct YAML syntax prevents parsing errors in downstream tools.
//
// What: Render should produce valid YAML with version field.
func TestRender_YAML(t *testing.T) {
	// Precondition: Version string and YAML format
	// Action: Render
	result, err := Render(FormatYAML, "1.2.3")

	// Expected: YAML-formatted version string
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `version: "1.2.3"`) {
		t.Errorf("expected YAML version string, got: %s", result)
	}
}

// TestRender_Go validates Go source code format output.
//
// Why: Go projects need compile-time version constants. The generated
// code must be syntactically valid Go.
//
// What: Render should produce Go code with Version constant.
func TestRender_Go(t *testing.T) {
	// Precondition: Version string and Go format
	// Action: Render
	result, err := Render(FormatGo, "1.2.3")

	// Expected: Go constant declaration
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `Version     = "1.2.3"`) {
		t.Errorf("expected Go version string, got: %s", result)
	}
}

// TestRender_C validates C source code format output.
//
// Why: C projects use preprocessor macros for version constants.
// Correct syntax is required for compilation.
//
// What: Render should produce C #define for VERSION.
func TestRender_C(t *testing.T) {
	// Precondition: Version string and C format
	// Action: Render
	result, err := Render(FormatC, "1.2.3")

	// Expected: C preprocessor define
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `#define VERSION "1.2.3"`) {
		t.Errorf("expected C version define, got: %s", result)
	}
}

// TestRender_CHeader validates C header file format output.
//
// Why: C header files need include guards and extern declarations
// for proper multi-file usage. Invalid headers break compilation.
//
// What: Render should produce C header with #define and extern declaration.
func TestRender_CHeader(t *testing.T) {
	// Precondition: Version string and C header format
	// Action: Render
	result, err := Render(FormatCHeader, "1.2.3")

	// Expected: Header with define and extern
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

// TestRender_CPP validates C++ source code format output.
//
// Why: C++ projects may use namespaces for version constants.
// The output must be valid C++ syntax.
//
// What: Render should produce C++ code with namespace and defines.
func TestRender_CPP(t *testing.T) {
	// Precondition: Version string and C++ format
	// Action: Render
	result, err := Render(FormatCPP, "1.2.3")

	// Expected: C++ with namespace
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

// TestRender_CPPHeader validates C++ header file format output.
//
// Why: C++ headers need include guards and extern declarations
// similar to C, but with namespace support.
//
// What: Render should produce C++ header with defines and extern declaration.
func TestRender_CPPHeader(t *testing.T) {
	// Precondition: Version string and C++ header format
	// Action: Render
	result, err := Render(FormatCPPHeader, "1.2.3")

	// Expected: Header with define and extern
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

// TestRender_JS validates JavaScript ES module format output.
//
// Why: Modern JavaScript uses ES modules. The export syntax must be
// correct for import/export to work.
//
// What: Render should produce ES module export statement.
func TestRender_JS(t *testing.T) {
	// Precondition: Version string and JS format
	// Action: Render
	result, err := Render(FormatJS, "1.2.3")

	// Expected: ES module export
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `export const VERSION = "1.2.3"`) {
		t.Errorf("expected JS version string, got: %s", result)
	}
}

// TestRender_TS validates TypeScript format output with type annotations.
//
// Why: TypeScript requires type annotations for compile-time type checking.
// Missing annotations would cause TypeScript compilation to infer less precise types.
//
// What: Render should produce TypeScript with explicit string type.
func TestRender_TS(t *testing.T) {
	// Precondition: Version string and TypeScript format
	// Action: Render
	result, err := Render(FormatTS, "1.2.3")

	// Expected: TypeScript with type annotation
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `export const VERSION: string = "1.2.3"`) {
		t.Errorf("expected TS version string, got: %s", result)
	}
}

// TestRender_Java validates Java class format output.
//
// Why: Java version constants must be public static final String
// to be accessible from other classes and immutable.
//
// What: Render should produce Java constant declaration.
func TestRender_Java(t *testing.T) {
	// Precondition: Version string and Java format
	// Action: Render
	result, err := Render(FormatJava, "1.2.3")

	// Expected: Java constant
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public static final String VERSION = "1.2.3"`) {
		t.Errorf("expected Java version string, got: %s", result)
	}
}

// TestRender_Kotlin validates Kotlin format output.
//
// Why: Kotlin uses const val for compile-time constants.
// This syntax differs from Java and must be correct.
//
// What: Render should produce Kotlin const val declaration.
func TestRender_Kotlin(t *testing.T) {
	// Precondition: Version string and Kotlin format
	// Action: Render
	result, err := Render(FormatKotlin, "1.2.3")

	// Expected: Kotlin constant
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `const val VERSION = "1.2.3"`) {
		t.Errorf("expected Kotlin version string, got: %s", result)
	}
}

// TestRender_CSharp validates C# format output.
//
// Why: C# uses public const string for version constants.
// Correct casing and modifiers are required for C# compilation.
//
// What: Render should produce C# constant declaration.
func TestRender_CSharp(t *testing.T) {
	// Precondition: Version string and C# format
	// Action: Render
	result, err := Render(FormatCSharp, "1.2.3")

	// Expected: C# constant
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public const string Version = "1.2.3"`) {
		t.Errorf("expected C# version string, got: %s", result)
	}
}

// TestRender_PHP validates PHP format output.
//
// Why: PHP class constants use different syntax than variables.
// Single quotes are preferred for strings without interpolation.
//
// What: Render should produce PHP const declaration with single quotes.
func TestRender_PHP(t *testing.T) {
	// Precondition: Version string and PHP format
	// Action: Render
	result, err := Render(FormatPHP, "1.2.3")

	// Expected: PHP constant with single quotes
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `const VERSION = '1.2.3'`) {
		t.Errorf("expected PHP version string, got: %s", result)
	}
}

// TestRender_Swift validates Swift format output.
//
// Why: Swift uses public let for module-level constants.
// Access modifiers must be correct for cross-module usage.
//
// What: Render should produce Swift public let declaration.
func TestRender_Swift(t *testing.T) {
	// Precondition: Version string and Swift format
	// Action: Render
	result, err := Render(FormatSwift, "1.2.3")

	// Expected: Swift public constant
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `public let VERSION = "1.2.3"`) {
		t.Errorf("expected Swift version string, got: %s", result)
	}
}

// TestRender_Ruby validates Ruby format output.
//
// Why: Ruby uses uppercase constants. The syntax must follow
// Ruby naming conventions.
//
// What: Render should produce Ruby constant assignment.
func TestRender_Ruby(t *testing.T) {
	// Precondition: Version string and Ruby format
	// Action: Render
	result, err := Render(FormatRuby, "1.2.3")

	// Expected: Ruby constant
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `VERSION = "1.2.3"`) {
		t.Errorf("expected Ruby version string, got: %s", result)
	}
}

// TestRender_Rust validates Rust format output.
//
// Why: Rust uses pub const with explicit type annotation for
// static string constants. Lifetime annotation &str is required.
//
// What: Render should produce Rust pub const with &str type.
func TestRender_Rust(t *testing.T) {
	// Precondition: Version string and Rust format
	// Action: Render
	result, err := Render(FormatRust, "1.2.3")

	// Expected: Rust constant with type
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result, `pub const VERSION: &str = "1.2.3"`) {
		t.Errorf("expected Rust version string, got: %s", result)
	}
}

// TestRenderTemplate_MajorMinor validates the two-component version shorthand.
//
// Why: Some systems (like Docker tags) use Major.Minor without patch.
// This variable provides that common format.
//
// What: {{MajorMinor}} should produce "1.2" for version "1.2.3".
func TestRenderTemplate_MajorMinor(t *testing.T) {
	// Precondition: Template using MajorMinor
	template := `{{MajorMinor}}`

	// Action: Render with full version
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: Only Major.Minor
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "1.2" {
		t.Errorf("expected 1.2, got: %s", result)
	}
}

// TestRenderTemplate_Prefix validates version prefix handling.
//
// Why: Many versioning schemes use "v" prefix (v1.2.3). Templates
// must preserve this prefix when requested.
//
// What: {{Prefix}}{{MajorMinorPatch}} should produce "v1.2.3" for "v1.2.3".
func TestRenderTemplate_Prefix(t *testing.T) {
	// Precondition: Template combining Prefix and MajorMinorPatch
	template := `{{Prefix}}{{MajorMinorPatch}}`

	// Action: Render with prefixed version
	result, err := RenderTemplate(template, "v1.2.3")

	// Expected: Prefix preserved
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "v1.2.3" {
		t.Errorf("expected v1.2.3, got: %s", result)
	}
}

// TestRenderTemplate_BuildDateComponents validates build timestamp variables.
//
// Why: Build metadata often includes build date for traceability.
// Individual date components allow flexible formatting.
//
// What: {{BuildYear}}-{{BuildMonth}}-{{BuildDay}} should produce YYYY-MM-DD format.
func TestRenderTemplate_BuildDateComponents(t *testing.T) {
	// Precondition: Template using build date components
	template := `{{BuildYear}}-{{BuildMonth}}-{{BuildDay}}`

	// Action: Render
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: Date in YYYY-MM-DD format (10 characters)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 10 {
		t.Errorf("expected date length 10, got: %d (%s)", len(result), result)
	}
}

// TestRenderTemplate_BuildDateUTC validates the combined UTC date variable.
//
// Why: Single date variable is simpler for common use cases.
// UTC ensures consistent dates across timezones.
//
// What: {{BuildDateUTC}} should produce YYYY-MM-DD format.
func TestRenderTemplate_BuildDateUTC(t *testing.T) {
	// Precondition: Template using BuildDateUTC
	template := `{{BuildDateUTC}}`

	// Action: Render
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: YYYY-MM-DD format
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 10 {
		t.Errorf("expected date length 10, got: %d (%s)", len(result), result)
	}
}

// TestRenderTemplate_BuildDateTimeCompact validates compact timestamp format.
//
// Why: Compact timestamps (YYYYMMDDHHmmss) are useful for build numbers
// and sortable filenames.
//
// What: {{BuildDateTimeCompact}} should produce 14-character timestamp.
func TestRenderTemplate_BuildDateTimeCompact(t *testing.T) {
	// Precondition: Template using compact datetime
	template := `{{BuildDateTimeCompact}}`

	// Action: Render
	result, err := RenderTemplate(template, "1.2.3")

	// Expected: 14-char format (YYYYMMDDHHmmss)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 14 {
		t.Errorf("expected compact datetime length 14, got: %d (%s)", len(result), result)
	}
}

// TestGetEmbeddedTemplate validates retrieval of embedded template content.
//
// Why: Users may want to inspect or modify default templates.
// This function exposes the embedded templates for customization.
//
// What: GetEmbeddedTemplate should return template with placeholders and language syntax.
func TestGetEmbeddedTemplate(t *testing.T) {
	// Precondition: Python format (known to exist)
	// Action: Get embedded template
	template, err := GetEmbeddedTemplate(FormatPython)

	// Expected: Template with Mustache placeholders and Python syntax
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

// TestIsValidFormat validates format name validation across all supported formats.
//
// Why: Users specify formats by name. Invalid formats should be rejected
// early with clear feedback rather than cryptic template errors.
//
// What: All known format names should be valid; unknown names should be invalid.
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
			// Action: Check format validity
			result := IsValidFormat(tt.format)

			// Expected: Match expected validity
			if result != tt.expected {
				t.Errorf("IsValidFormat(%s) = %v, expected %v", tt.format, result, tt.expected)
			}
		})
	}
}

// TestSupportedFormats validates that the list of supported formats is non-empty
// and includes expected formats.
//
// Why: The supported formats list is used for help text and validation.
// An empty or incomplete list would confuse users.
//
// What: SupportedFormats should return a list containing at least "python".
func TestSupportedFormats(t *testing.T) {
	// Action: Get supported formats
	formats := SupportedFormats()

	// Expected: Non-empty list containing python
	if len(formats) == 0 {
		t.Error("expected at least one supported format")
	}

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

// TestWriteToFile validates basic file writing functionality.
//
// Why: File output is a core feature. Write failures would prevent
// any version file generation.
//
// What: WriteToFile should create a file with the specified content.
func TestWriteToFile(t *testing.T) {
	// Precondition: Temporary directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")

	content := "test content"

	// Action: Write to file
	err := WriteToFile(content, tmpFile)

	// Expected: File exists with correct content
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != content {
		t.Errorf("expected %s, got: %s", content, string(data))
	}
}

// TestBuildCompleteTemplateData validates building template data with
// pre-release and metadata fields populated.
//
// Why: Pre-release and metadata are essential for development versions
// and build traceability. Correct formatting ensures SemVer compliance.
//
// What: Pre-release and metadata templates should be rendered with proper prefixes.
func TestBuildCompleteTemplateData(t *testing.T) {
	// Precondition: Version and pre-release/metadata templates
	vd := &version.Version{
		Major: 1,
		Minor: 2,
		Patch: 3,
	}

	prereleaseTemplate := "alpha-1"
	metadataTemplate := "build.123"

	// Action: Build complete template data
	data := BuildCompleteTemplateData(vd, prereleaseTemplate, metadataTemplate)

	// Expected: Pre-release and metadata with correct prefixes
	if data.PreRelease != "alpha-1" {
		t.Errorf("expected PreRelease=alpha-1, got %s", data.PreRelease)
	}
	if data.PreReleaseWithDash != "-alpha-1" {
		t.Errorf("expected PreReleaseWithDash=-alpha-1, got %s", data.PreReleaseWithDash)
	}
	if data.Metadata != "build.123" {
		t.Errorf("expected Metadata=build.123, got %s", data.Metadata)
	}
	if data.MetadataWithPlus != "+build.123" {
		t.Errorf("expected MetadataWithPlus=+build.123, got %s", data.MetadataWithPlus)
	}
}

// TestRenderTemplateList validates joining multiple template results.
//
// Why: Some version formats require joining multiple components
// (e.g., pre-release identifiers). The list renderer handles this.
//
// What: Templates ["a", "b", "c"] with "-" separator should produce "a-b-c".
func TestRenderTemplateList(t *testing.T) {
	// Precondition: Version and list of templates
	vd := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := BuildTemplateDataFromVersion(vd)

	templates := []string{"a", "b", "c"}

	// Action: Render template list
	result := RenderTemplateList(templates, data, "-")

	// Expected: Joined result
	if result != "a-b-c" {
		t.Errorf("expected a-b-c, got %s", result)
	}
}

// TestRenderTemplateList_DotSeparator validates template list with dot separator.
//
// Why: Metadata uses dot separators per SemVer. The separator parameter
// must be respected.
//
// What: Version component templates with "." separator should produce "1.2.3".
func TestRenderTemplateList_DotSeparator(t *testing.T) {
	// Precondition: Templates using version variables
	vd := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := BuildTemplateDataFromVersion(vd)

	templates := []string{"{{Major}}", "{{Minor}}", "{{Patch}}"}

	// Action: Render with dot separator
	result := RenderTemplateList(templates, data, ".")

	// Expected: Dot-separated version
	if result != "1.2.3" {
		t.Errorf("expected 1.2.3, got %s", result)
	}
}

// TestMergeCustomVars validates merging custom variables into TemplateData.
//
// Why: Custom variables allow user-defined template values beyond built-ins.
// They must be properly merged with override semantics.
//
// What: New keys should be added; existing keys should be overridden.
func TestMergeCustomVars(t *testing.T) {
	// Precondition: TemplateData with existing custom variable
	data := TemplateData{
		Major: "1",
		Minor: "2",
		Patch: "3",
		Custom: map[string]string{
			"ExistingKey": "original",
		},
	}

	extraVars := map[string]string{
		"NewKey":      "new-value",
		"ExistingKey": "overridden",
	}

	// Action: Merge custom vars
	MergeCustomVars(&data, extraVars)

	// Expected: Both new and overridden values present
	if data.Custom["NewKey"] != "new-value" {
		t.Errorf("expected NewKey='new-value', got %s", data.Custom["NewKey"])
	}
	if data.Custom["ExistingKey"] != "overridden" {
		t.Errorf("expected ExistingKey='overridden', got %s", data.Custom["ExistingKey"])
	}
}

// TestTemplateDataToStringMap validates conversion to string map format.
//
// Why: Some systems (like environment variable export) need version data
// as a flat string map. This conversion must include all fields.
//
// What: All standard fields and custom/plugin variables should appear in map.
func TestTemplateDataToStringMap(t *testing.T) {
	// Precondition: Fully populated TemplateData
	data := TemplateData{
		Major:              "1",
		Minor:              "2",
		Patch:              "3",
		MajorMinorPatch:    "1.2.3",
		MajorMinor:         "1.2",
		Prefix:             "v",
		PreRelease:         "alpha",
		PreReleaseWithDash: "-alpha",
		Metadata:           "build123",
		MetadataWithPlus:   "+build123",
		Hash:               "abc123def456",
		ShortHash:          "abc123d",
		BranchName:         "main",
		Custom: map[string]string{
			"CustomVar": "custom-value",
		},
		PluginVariables: map[string]string{
			"PluginVar": "plugin-value",
		},
	}

	// Action: Convert to map
	result := TemplateDataToStringMap(data)

	// Expected: Standard fields present
	if result["Major"] != "1" {
		t.Errorf("expected Major='1', got %s", result["Major"])
	}
	if result["MajorMinorPatch"] != "1.2.3" {
		t.Errorf("expected MajorMinorPatch='1.2.3', got %s", result["MajorMinorPatch"])
	}
	if result["PreRelease"] != "alpha" {
		t.Errorf("expected PreRelease='alpha', got %s", result["PreRelease"])
	}
	if result["Metadata"] != "build123" {
		t.Errorf("expected Metadata='build123', got %s", result["Metadata"])
	}
	if result["Hash"] != "abc123def456" {
		t.Errorf("expected Hash='abc123def456', got %s", result["Hash"])
	}

	// Expected: Custom variables merged
	if result["CustomVar"] != "custom-value" {
		t.Errorf("expected CustomVar='custom-value', got %s", result["CustomVar"])
	}

	// Expected: Plugin variables merged
	if result["PluginVar"] != "plugin-value" {
		t.Errorf("expected PluginVar='plugin-value', got %s", result["PluginVar"])
	}
}

// TestEmitTemplateToFile validates custom template rendering and file output.
//
// Why: Users may need custom output formats not covered by built-in templates.
// This combines template rendering with file output.
//
// What: Custom template should be rendered and written to file.
func TestEmitTemplateToFile(t *testing.T) {
	// Precondition: Temporary directory and custom template
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "version.txt")

	template := "Version: {{Major}}.{{Minor}}.{{Patch}}"

	// Action: Emit template to file
	err := EmitTemplateToFile(template, "1.2.3", tmpFile)

	// Expected: File contains rendered template
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(tmpFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if !strings.Contains(string(data), "Version: 1.2.3") {
		t.Errorf("expected 'Version: 1.2.3' in file, got: %s", string(data))
	}
}

// TestRenderTemplateWithData validates rendering with pre-populated TemplateData.
//
// Why: Sometimes template data is built separately from rendering.
// This function allows using pre-constructed TemplateData.
//
// What: Template with custom variable should render correctly.
func TestRenderTemplateWithData(t *testing.T) {
	// Precondition: TemplateData with custom variable
	data := TemplateData{
		Major:           "1",
		Minor:           "2",
		Patch:           "3",
		MajorMinorPatch: "1.2.3",
		Custom: map[string]string{
			"AppName": "myapp",
		},
	}

	template := "{{AppName}} v{{MajorMinorPatch}}"

	// Action: Render with data
	result, err := RenderTemplateWithData(template, data)

	// Expected: Custom variable and version both rendered
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != "myapp v1.2.3" {
		t.Errorf("expected 'myapp v1.2.3', got %s", result)
	}
}

// TestGetVCSInfo_WithMockVCS validates VCS information gathering with mock.
//
// Why: VCS info is essential for commit-based versioning. The gathering
// must correctly populate all fields from VCS operations.
//
// What: All VCS fields should be populated from mock responses.
func TestGetVCSInfo_WithMockVCS(t *testing.T) {
	// Precondition: Mock VCS with known values
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(t.TempDir(), nil).AnyTimes()

	expectedDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	mockVCS.EXPECT().GetVCSIdentifier(40).Return("abc123def456789012345678901234567890dead", nil).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("feature/test-branch", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(expectedDate, nil).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(42, nil).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("def456", nil).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(3, nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthor().Return("Test Author", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("test@example.com", nil).AnyTimes()

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Action: Get VCS info
	info := getVCSInfo()

	// Expected: All fields populated correctly
	if info.Identifier != "abc123def456789012345678901234567890dead" {
		t.Errorf("expected Identifier='abc123def456789012345678901234567890dead', got %s", info.Identifier)
	}
	if info.IdentifierShort != "abc123d" {
		t.Errorf("expected IdentifierShort='abc123d', got %s", info.IdentifierShort)
	}
	if info.IdentifierMedium != "abc123def456" {
		t.Errorf("expected IdentifierMedium='abc123def456', got %s", info.IdentifierMedium)
	}
	if info.BranchName != "feature/test-branch" {
		t.Errorf("expected BranchName='feature/test-branch', got %s", info.BranchName)
	}
	if !info.CommitDate.Equal(expectedDate) {
		t.Errorf("expected CommitDate=%v, got %v", expectedDate, info.CommitDate)
	}
	if info.CommitsSinceTag != 42 {
		t.Errorf("expected CommitsSinceTag=42, got %d", info.CommitsSinceTag)
	}
	if info.VersionSourceHash != "def456" {
		t.Errorf("expected VersionSourceHash='def456', got %s", info.VersionSourceHash)
	}
	if info.UncommittedChanges != 3 {
		t.Errorf("expected UncommittedChanges=3, got %d", info.UncommittedChanges)
	}
	if info.CommitAuthor != "Test Author" {
		t.Errorf("expected CommitAuthor='Test Author', got %s", info.CommitAuthor)
	}
	if info.CommitAuthorEmail != "test@example.com" {
		t.Errorf("expected CommitAuthorEmail='test@example.com', got %s", info.CommitAuthorEmail)
	}
}

// TestFormatVCSFields validates VCS field formatting with various inputs.
//
// Why: VCS data must be formatted consistently for templates.
// Dirty flag, padding, and date formatting are especially important.
//
// What: VCSInfo should be converted to formatted strings correctly.
func TestFormatVCSFields(t *testing.T) {
	t.Run("with commit date and commits since tag", func(t *testing.T) {
		// Precondition: VCSInfo with known values
		info := VCSInfo{
			CommitsSinceTag:    42,
			UncommittedChanges: 3,
			CommitDate:         mustParseTime("2024-01-15T10:30:00Z"),
		}

		// Action: Format fields
		fields := formatVCSFields(info)

		// Expected: All fields formatted correctly
		if fields.CommitsSinceTag != "42" {
			t.Errorf("expected CommitsSinceTag='42', got %s", fields.CommitsSinceTag)
		}
		if fields.BuildNumberPadded != "0042" {
			t.Errorf("expected BuildNumberPadded='0042', got %s", fields.BuildNumberPadded)
		}
		if fields.UncommittedChanges != "3" {
			t.Errorf("expected UncommittedChanges='3', got %s", fields.UncommittedChanges)
		}
		if fields.Dirty != "dirty" {
			t.Errorf("expected Dirty='dirty', got %s", fields.Dirty)
		}
		if fields.CommitYear != "2024" {
			t.Errorf("expected CommitYear='2024', got %s", fields.CommitYear)
		}
		if fields.CommitMonth != "01" {
			t.Errorf("expected CommitMonth='01', got %s", fields.CommitMonth)
		}
		if fields.CommitDay != "15" {
			t.Errorf("expected CommitDay='15', got %s", fields.CommitDay)
		}
	})

	t.Run("no uncommitted changes", func(t *testing.T) {
		// Precondition: VCSInfo with no uncommitted changes
		info := VCSInfo{
			UncommittedChanges: 0,
		}

		// Action: Format fields
		fields := formatVCSFields(info)

		// Expected: Dirty flag empty
		if fields.Dirty != "" {
			t.Errorf("expected Dirty='', got %s", fields.Dirty)
		}
	})

	t.Run("negative commits since tag (no tags)", func(t *testing.T) {
		// Precondition: VCSInfo with -1 (no tags)
		info := VCSInfo{
			CommitsSinceTag: -1,
		}

		// Action: Format fields
		fields := formatVCSFields(info)

		// Expected: Empty strings for commit-related fields
		if fields.CommitsSinceTag != "" {
			t.Errorf("expected empty CommitsSinceTag for -1, got %s", fields.CommitsSinceTag)
		}
		if fields.BuildNumberPadded != "" {
			t.Errorf("expected empty BuildNumberPadded for -1, got %s", fields.BuildNumberPadded)
		}
	})
}

// TestFormatBuildTime validates build timestamp formatting.
//
// Why: Build timestamps are used for traceability and unique build IDs.
// All time formats must be consistent and correctly sized.
//
// What: Build time fields should have correct formats and lengths.
func TestFormatBuildTime(t *testing.T) {
	// Action: Format build time
	bt := formatBuildTime()

	// Expected: RFC3339 format for DateTime
	if !strings.Contains(bt.DateTime, "T") || !strings.HasSuffix(bt.DateTime, "Z") {
		t.Errorf("expected RFC3339 format, got %s", bt.DateTime)
	}

	// Expected: 14 chars for compact format (YYYYMMDDHHmmss)
	if len(bt.DateCompact) != 14 {
		t.Errorf("expected 14 char compact date, got %d: %s", len(bt.DateCompact), bt.DateCompact)
	}

	// Expected: 10 chars for date only (YYYY-MM-DD)
	if len(bt.DateOnly) != 10 {
		t.Errorf("expected 10 char date, got %d: %s", len(bt.DateOnly), bt.DateOnly)
	}

	// Expected: 4 chars for year
	if len(bt.Year) != 4 {
		t.Errorf("expected 4 char year, got %d: %s", len(bt.Year), bt.Year)
	}

	// Expected: 2 chars for month
	if len(bt.Month) != 2 {
		t.Errorf("expected 2 char month, got %d: %s", len(bt.Month), bt.Month)
	}

	// Expected: 2 chars for day
	if len(bt.Day) != 2 {
		t.Errorf("expected 2 char day, got %d: %s", len(bt.Day), bt.Day)
	}
}

// =============================================================================
// ERROR HANDLING
// Tests verifying expected failure modes and error messages.
// =============================================================================

// TestRender_InvalidFormat validates error handling for unsupported formats.
//
// Why: Users may mistype format names. Clear error messages help them
// identify and fix the problem.
//
// What: Render with unknown format should return an error.
func TestRender_InvalidFormat(t *testing.T) {
	// Precondition: Invalid format name
	// Action: Attempt to render
	_, err := Render(Format("invalid"), "1.2.3")

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// TestRenderTemplate_InvalidTemplate validates error handling for malformed templates.
//
// Why: User-provided templates may have syntax errors. The system must
// fail gracefully with helpful error messages.
//
// What: Malformed Mustache template should return error.
func TestRenderTemplate_InvalidTemplate(t *testing.T) {
	// Precondition: Unclosed Mustache tag
	template := `{{Invalid`

	// Action: Attempt to render
	_, err := RenderTemplate(template, "1.2.3")

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

// TestGetEmbeddedTemplate_InvalidFormat validates error for non-existent template.
//
// Why: Requesting templates for invalid formats must fail clearly
// rather than returning empty or corrupted data.
//
// What: GetEmbeddedTemplate with unknown format should return error.
func TestGetEmbeddedTemplate_InvalidFormat(t *testing.T) {
	// Precondition: Invalid format name
	// Action: Attempt to get template
	_, err := GetEmbeddedTemplate(Format("invalid"))

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// TestValidateOutputPath_EmptyPath validates rejection of empty output paths.
//
// Why: Empty paths would cause undefined behavior in file operations.
// This must be caught early with a clear error message.
//
// What: ValidateOutputPath with empty string should return specific error.
func TestValidateOutputPath_EmptyPath(t *testing.T) {
	// Precondition: Empty path
	// Action: Validate
	err := ValidateOutputPath("")

	// Expected: Error with ErrOutputPathEmpty message
	if err == nil {
		t.Error("expected error for empty path")
	}
	if !strings.Contains(err.Error(), ErrOutputPathEmpty) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathEmpty, err)
	}
}

// TestValidateOutputPath_IsDirectory validates rejection of directory paths.
//
// Why: Attempting to write version content to a directory would fail.
// Early validation prevents confusing errors.
//
// What: ValidateOutputPath with directory should return specific error.
func TestValidateOutputPath_IsDirectory(t *testing.T) {
	// Precondition: Existing directory
	tmpDir := t.TempDir()

	// Action: Validate directory path
	err := ValidateOutputPath(tmpDir)

	// Expected: Error with ErrOutputPathIsDirectory message
	if err == nil {
		t.Error("expected error when path is a directory")
	}
	if !strings.Contains(err.Error(), ErrOutputPathIsDirectory) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathIsDirectory, err)
	}
}

// TestValidateOutputPath_ParentNotExist validates rejection when parent directory missing.
//
// Why: Creating deeply nested files without existing parents would fail.
// Validation catches this before attempting the write.
//
// What: ValidateOutputPath with non-existent parent should return error.
func TestValidateOutputPath_ParentNotExist(t *testing.T) {
	// Precondition: Path with non-existent parent
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "nonexistent", "file.txt")

	// Action: Validate
	err := ValidateOutputPath(tmpFile)

	// Expected: Error with ErrParentDirNotExist message
	if err == nil {
		t.Error("expected error when parent directory doesn't exist")
	}
	if !strings.Contains(err.Error(), ErrParentDirNotExist) {
		t.Errorf("expected error containing %q, got: %v", ErrParentDirNotExist, err)
	}
}

// TestValidateOutputPath_ParentIsFile validates rejection when parent is a file.
//
// Why: Path components that are files cannot contain children.
// This invalid path structure must be detected.
//
// What: ValidateOutputPath where parent is a file should return error.
func TestValidateOutputPath_ParentIsFile(t *testing.T) {
	// Precondition: File masquerading as directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")

	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	invalidPath := filepath.Join(tmpFile, "child.txt")

	// Action: Validate
	err := ValidateOutputPath(invalidPath)

	// Expected: Error returned
	if err == nil {
		t.Error("expected error when parent is a file, not a directory")
	}
}

// TestWriteToFile_ValidatesPath validates that WriteToFile performs path validation.
//
// Why: WriteToFile should validate paths before attempting I/O
// to provide clear error messages.
//
// What: Writing to a directory should fail with validation error.
func TestWriteToFile_ValidatesPath(t *testing.T) {
	// Precondition: Directory path
	tmpDir := t.TempDir()

	// Action: Attempt to write to directory
	err := WriteToFile("content", tmpDir)

	// Expected: Error with validation message
	if err == nil {
		t.Error("expected error when writing to a directory")
	}
	if !strings.Contains(err.Error(), ErrOutputPathIsDirectory) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathIsDirectory, err)
	}
}

// TestWriteToFile_EmptyPath validates WriteToFile rejects empty paths.
//
// Why: Empty paths must be caught at the validation layer
// rather than causing low-level I/O errors.
//
// What: WriteToFile with empty path should return validation error.
func TestWriteToFile_EmptyPath(t *testing.T) {
	// Precondition: Empty path
	// Action: Attempt to write
	err := WriteToFile("content", "")

	// Expected: Error with validation message
	if err == nil {
		t.Error("expected error for empty path")
	}
	if !strings.Contains(err.Error(), ErrOutputPathEmpty) {
		t.Errorf("expected error containing %q, got: %v", ErrOutputPathEmpty, err)
	}
}

// TestEmitTemplateToFile_InvalidTemplate validates error handling for bad templates.
//
// Why: Invalid templates should fail before creating output files
// to avoid leaving partial/empty files.
//
// What: EmitTemplateToFile with invalid template should error and not create file.
func TestEmitTemplateToFile_InvalidTemplate(t *testing.T) {
	// Precondition: Temporary directory and invalid template
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "version.txt")

	template := "{{Invalid"

	// Action: Attempt to emit
	err := EmitTemplateToFile(template, "1.2.3", tmpFile)

	// Expected: Error returned, file not created
	if err == nil {
		t.Error("expected error for invalid template")
	}

	if _, statErr := os.Stat(tmpFile); !os.IsNotExist(statErr) {
		t.Error("file should not be created when template is invalid")
	}
}

// TestEmitTemplateToFile_InvalidPath validates error handling for bad output paths.
//
// Why: Invalid paths should be rejected with clear error messages
// even when the template itself is valid.
//
// What: EmitTemplateToFile with non-existent parent should error.
func TestEmitTemplateToFile_InvalidPath(t *testing.T) {
	// Precondition: Non-existent parent directory
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "nonexistent", "dir", "file.txt")

	template := "{{Major}}.{{Minor}}.{{Patch}}"

	// Action: Attempt to emit
	err := EmitTemplateToFile(template, "1.2.3", invalidPath)

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid path")
	}
}

// TestEmitToFile_InvalidFormat validates error handling for unsupported formats.
//
// Why: Invalid formats must be rejected clearly rather than
// producing empty or corrupted output files.
//
// What: EmitToFile with unknown format should return error.
func TestEmitToFile_InvalidFormat(t *testing.T) {
	// Precondition: Temporary file and invalid format
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "version.txt")

	// Action: Attempt to emit
	err := EmitToFile(Format("invalid"), "1.2.3", tmpFile)

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid format")
	}
}

// TestRenderTemplateWithData_InvalidTemplate validates error for bad templates
// when using pre-built TemplateData.
//
// Why: Template syntax errors must be caught regardless of how
// the template data was constructed.
//
// What: RenderTemplateWithData with malformed template should error.
func TestRenderTemplateWithData_InvalidTemplate(t *testing.T) {
	// Precondition: Valid data, invalid template
	data := TemplateData{Major: "1"}

	template := "{{Invalid"

	// Action: Attempt to render
	_, err := RenderTemplateWithData(template, data)

	// Expected: Error returned
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

// TestGetVCSInfo_NoVCS validates graceful handling when no VCS is available.
//
// Why: Versionator should work in non-VCS directories by providing
// empty/default VCS values rather than failing.
//
// What: getVCSInfo without VCS should return defaults with CommitsSinceTag=-1.
func TestGetVCSInfo_NoVCS(t *testing.T) {
	// Precondition: Unregister all VCS
	vcs.UnregisterVCS("git")
	defer vcs.RegisterVCS(gitVCS.NewGitVCSDefault())

	// Action: Get VCS info
	info := getVCSInfo()

	// Expected: Default empty values
	if info.CommitsSinceTag != -1 {
		t.Errorf("expected CommitsSinceTag=-1 when no VCS, got %d", info.CommitsSinceTag)
	}
	if info.Identifier != "" {
		t.Errorf("expected empty Identifier when no VCS, got %s", info.Identifier)
	}
	if info.BranchName != "" {
		t.Errorf("expected empty BranchName when no VCS, got %s", info.BranchName)
	}
}

// TestGetVCSInfo_WithErrors validates graceful handling when VCS operations fail.
//
// Why: VCS commands can fail for various reasons (permissions, corruption).
// The system should return defaults rather than crashing.
//
// What: getVCSInfo with failing VCS should return defaults.
func TestGetVCSInfo_WithErrors(t *testing.T) {
	// Precondition: Mock VCS returning errors for all operations
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(t.TempDir(), nil).AnyTimes()

	testErr := os.ErrNotExist
	mockVCS.EXPECT().GetVCSIdentifier(40).Return("", testErr).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("", testErr).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(time.Time{}, testErr).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(0, testErr).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("", testErr).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(0, testErr).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthor().Return("", testErr).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("", testErr).AnyTimes()

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Action: Get VCS info
	info := getVCSInfo()

	// Expected: Empty/default values
	if info.Identifier != "" {
		t.Errorf("expected empty Identifier with error, got %s", info.Identifier)
	}
	if info.BranchName != "" {
		t.Errorf("expected empty BranchName with error, got %s", info.BranchName)
	}
	if info.CommitsSinceTag != -1 {
		t.Errorf("expected CommitsSinceTag=-1 with error, got %d", info.CommitsSinceTag)
	}
}

// =============================================================================
// EDGE CASES
// Tests covering boundary conditions and unusual but valid inputs.
// =============================================================================

// TestValidateOutputPath_ExistingFile validates that existing files are accepted.
//
// Why: Overwriting existing version files is a valid use case.
// The validation should not reject existing file paths.
//
// What: ValidateOutputPath with existing file should succeed.
func TestValidateOutputPath_ExistingFile(t *testing.T) {
	// Precondition: Existing file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "existing.txt")

	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	// Action: Validate existing file path
	err := ValidateOutputPath(tmpFile)

	// Expected: No error
	if err != nil {
		t.Errorf("expected no error for existing file, got: %v", err)
	}
}

// TestValidateOutputPath_NewFileInExistingDir validates acceptance of new file paths.
//
// Why: Creating new files in existing directories is a valid use case.
// Path validation should not require the file to already exist.
//
// What: ValidateOutputPath with new filename in existing dir should succeed.
func TestValidateOutputPath_NewFileInExistingDir(t *testing.T) {
	// Precondition: New file in existing directory
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "new.txt")

	// Action: Validate new file path
	err := ValidateOutputPath(tmpFile)

	// Expected: No error
	if err != nil {
		t.Errorf("expected no error for new file in existing directory, got: %v", err)
	}
}

// TestValidateOutputPath_FileInCurrentDir validates simple filenames without path.
//
// Why: Users may specify just a filename without directory prefix.
// This should be treated as a file in the current directory.
//
// What: ValidateOutputPath with simple filename should succeed.
func TestValidateOutputPath_FileInCurrentDir(t *testing.T) {
	// Precondition: Simple filename (no directory)
	// Action: Validate
	err := ValidateOutputPath("simple.txt")

	// Expected: No error
	if err != nil {
		t.Errorf("expected no error for simple filename, got: %v", err)
	}
}

// TestBuildCompleteTemplateData_EmptyTemplates validates handling of empty templates.
//
// Why: Empty pre-release and metadata templates should result in
// empty fields, not errors or placeholder text.
//
// What: BuildCompleteTemplateData with empty templates should leave fields empty.
func TestBuildCompleteTemplateData_EmptyTemplates(t *testing.T) {
	// Precondition: Version with empty templates
	vd := &version.Version{
		Major: 1,
		Minor: 2,
		Patch: 3,
	}

	// Action: Build with empty templates
	data := BuildCompleteTemplateData(vd, "", "")

	// Expected: Empty pre-release and metadata fields
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

// TestMergeCustomVars_NilCustomMap validates initialization of nil Custom map.
//
// Why: TemplateData may have a nil Custom map initially.
// MergeCustomVars must initialize it before adding values.
//
// What: MergeCustomVars with nil Custom should initialize and add values.
func TestMergeCustomVars_NilCustomMap(t *testing.T) {
	// Precondition: TemplateData with nil Custom
	data := TemplateData{
		Major:  "1",
		Custom: nil,
	}

	extraVars := map[string]string{
		"Key": "value",
	}

	// Action: Merge
	MergeCustomVars(&data, extraVars)

	// Expected: Custom initialized with values
	if data.Custom == nil {
		t.Error("Custom map should be initialized")
	}
	if data.Custom["Key"] != "value" {
		t.Errorf("expected Key='value', got %s", data.Custom["Key"])
	}
}

// TestRenderTemplateList_EmptyTemplates validates filtering of empty templates.
//
// Why: Template lists may contain empty strings from configuration.
// These should be silently filtered rather than producing empty segments.
//
// What: Empty and whitespace-only templates should be skipped.
func TestRenderTemplateList_EmptyTemplates(t *testing.T) {
	// Precondition: List with empty templates
	data := TemplateData{Major: "1", Minor: "2", Patch: "3"}

	templates := []string{"a", "", "  ", "b"}

	// Action: Render list
	result := RenderTemplateList(templates, data, "-")

	// Expected: Empty templates filtered out
	if result != "a-b" {
		t.Errorf("expected 'a-b' (empty templates skipped), got %s", result)
	}
}

// TestGetVCSInfo_ShortIdentifier validates handling of short commit identifiers.
//
// Why: Some VCS systems may return shorter identifiers than expected.
// Truncation logic must handle identifiers shorter than target length.
//
// What: Short identifier should be used as-is without buffer overflow.
func TestGetVCSInfo_ShortIdentifier(t *testing.T) {
	// Precondition: Mock VCS returning short identifier
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockVCS := mock.NewMockVersionControlSystem(ctrl)
	mockVCS.EXPECT().Name().Return("git").AnyTimes()
	mockVCS.EXPECT().IsRepository().Return(true).AnyTimes()
	mockVCS.EXPECT().GetRepositoryRoot().Return(t.TempDir(), nil).AnyTimes()

	mockVCS.EXPECT().GetVCSIdentifier(40).Return("abc", nil).AnyTimes()
	mockVCS.EXPECT().GetBranchName().Return("main", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitDate().Return(time.Now(), nil).AnyTimes()
	mockVCS.EXPECT().GetCommitsSinceTag().Return(0, nil).AnyTimes()
	mockVCS.EXPECT().GetLastTagCommit().Return("", nil).AnyTimes()
	mockVCS.EXPECT().GetUncommittedChanges().Return(0, nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthor().Return("", nil).AnyTimes()
	mockVCS.EXPECT().GetCommitAuthorEmail().Return("", nil).AnyTimes()

	vcs.UnregisterVCS("git")
	vcs.RegisterVCS(mockVCS)
	defer func() {
		vcs.UnregisterVCS("git")
		vcs.RegisterVCS(gitVCS.NewGitVCSDefault())
	}()

	// Action: Get VCS info
	info := getVCSInfo()

	// Expected: Short identifier preserved without truncation errors
	if info.IdentifierShort != "abc" {
		t.Errorf("expected IdentifierShort='abc', got %s", info.IdentifierShort)
	}
	if info.IdentifierMedium != "abc" {
		t.Errorf("expected IdentifierMedium='abc', got %s", info.IdentifierMedium)
	}
}

// =============================================================================
// MINUTIAE
// Obscure scenarios and helper function tests.
// =============================================================================

// TestDirtyFlag validates the dirty flag helper function.
//
// Why: The dirty flag is used in version strings to indicate uncommitted changes.
// Correct logic is essential for accurate version reporting.
//
// What: 0 should produce empty, positive numbers should produce "dirty".
func TestDirtyFlag(t *testing.T) {
	// Precondition/Action/Expected: Various uncommitted change counts
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

// TestFormatPreReleaseNumber validates pre-release number formatting.
//
// Why: Pre-release numbers may be negative (-1 indicates none) or zero.
// The formatting must handle all cases correctly.
//
// What: Negative returns empty, zero and positive return string representation.
func TestFormatPreReleaseNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{name: "negative returns empty", input: -1, expected: ""},
		{name: "zero returns string zero", input: 0, expected: "0"},
		{name: "positive returns string", input: 42, expected: "42"},
		{name: "large number", input: 9999, expected: "9999"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Action: Format number
			result := formatPreReleaseNumber(tt.input)

			// Expected: Match expected output
			if result != tt.expected {
				t.Errorf("formatPreReleaseNumber(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestRenderTemplateList_InvalidTemplate validates handling of invalid templates in lists.
//
// Why: One bad template in a list should not prevent valid templates
// from rendering. Invalid templates are silently skipped.
//
// What: Invalid templates should be filtered, valid ones rendered.
func TestRenderTemplateList_InvalidTemplate(t *testing.T) {
	// Precondition: List with invalid template
	data := TemplateData{Major: "1", Minor: "2", Patch: "3"}

	templates := []string{"valid", "{{invalid", "also-valid"}

	// Action: Render list
	result := RenderTemplateList(templates, data, "-")

	// Expected: Invalid template skipped
	if result != "valid-also-valid" {
		t.Errorf("expected 'valid-also-valid', got %s", result)
	}
}

// =============================================================================
// TEST HELPERS
// =============================================================================

// mustParseTime parses a time string in RFC3339 format, panicking on error.
// Used only in tests where the time string is known to be valid.
func mustParseTime(s string) (t time.Time) {
	t, _ = time.Parse(time.RFC3339, s)
	return
}
