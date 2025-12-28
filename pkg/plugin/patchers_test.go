package plugin

import (
	"strings"
	"testing"
)

func TestPatchJSON_ValidJSON_PatchesVersion(t *testing.T) {
	patcher := PatchJSON()
	content := `{
  "name": "my-package",
  "version": "1.0.0",
  "description": "test"
}`
	expected := `{
  "name": "my-package",
  "version": "2.0.0",
  "description": "test"
}`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestPatchJSON_NoVersionField_ReturnsUnchanged(t *testing.T) {
	patcher := PatchJSON()
	content := `{"name": "my-package", "description": "test"}`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged, got: %s", result)
	}
}

func TestPatchJSON_InvalidJSON_ReturnsError(t *testing.T) {
	patcher := PatchJSON()
	content := `{invalid json`

	_, err := patcher(content, "2.0.0")
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "invalid JSON") {
		t.Errorf("expected 'invalid JSON' error, got: %v", err)
	}
}

func TestPatchJSON_NestedVersion_OnlyPatchesTopLevel(t *testing.T) {
	patcher := PatchJSON()
	content := `{
  "version": "1.0.0",
  "dependencies": {
    "other": {
      "version": "3.0.0"
    }
  }
}`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `"version": "2.0.0"`) {
		t.Error("expected top-level version to be patched")
	}
	if !strings.Contains(result, `"version": "3.0.0"`) {
		t.Error("expected nested version to remain unchanged")
	}
}

func TestPatchJSON_PreservesSpacing(t *testing.T) {
	patcher := PatchJSON()
	content := `{"version":  "1.0.0"}`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `"version":  "2.0.0"`) {
		t.Errorf("expected spacing to be preserved, got: %s", result)
	}
}

func TestPatchTOML_DoubleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchTOML()
	content := `[package]
name = "my-crate"
version = "1.0.0"
`
	expected := `[package]
name = "my-crate"
version = "2.0.0"
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != expected {
		t.Errorf("expected:\n%s\ngot:\n%s", expected, result)
	}
}

func TestPatchTOML_SingleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchTOML()
	content := `version = '1.0.0'`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `version = '2.0.0'` {
		t.Errorf("expected version = '2.0.0', got: %s", result)
	}
}

func TestPatchTOML_NoVersionField_ReturnsUnchanged(t *testing.T) {
	patcher := PatchTOML()
	content := `[package]
name = "my-crate"
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}

func TestPatchXML_LowercaseVersion_PatchesVersion(t *testing.T) {
	patcher := PatchXML()
	content := `<project>
  <version>1.0.0</version>
  <dependencies>
    <dependency>
      <version>3.0.0</version>
    </dependency>
  </dependencies>
</project>`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "<version>2.0.0</version>") {
		t.Error("expected first version to be patched")
	}
	if !strings.Contains(result, "<version>3.0.0</version>") {
		t.Error("expected nested version to remain unchanged")
	}
}

func TestPatchXML_UppercaseVersion_PatchesVersion(t *testing.T) {
	patcher := PatchXML()
	content := `<Project>
  <Version>1.0.0</Version>
</Project>`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "<Version>2.0.0</Version>") {
		t.Errorf("expected Version to be patched, got: %s", result)
	}
}

func TestPatchXML_NoVersionElement_ReturnsUnchanged(t *testing.T) {
	patcher := PatchXML()
	content := `<project><name>test</name></project>`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}

func TestPatchYAML_BasicVersion_PatchesVersion(t *testing.T) {
	patcher := PatchYAML()
	content := `name: my-app
version: 1.0.0
description: test
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "version: 2.0.0") {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchYAML_QuotedVersion_PatchesVersion(t *testing.T) {
	patcher := PatchYAML()
	content := `version: "1.0.0"`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "version: 2.0.0") {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchYAML_NoVersionField_ReturnsUnchanged(t *testing.T) {
	patcher := PatchYAML()
	content := `name: my-app
description: test
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}

func TestPatchGradle_DoubleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchGradle()
	content := `plugins {
    id 'java'
}

version = "1.0.0"
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `version = "2.0.0"`) {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchGradle_SingleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchGradle()
	content := `version = '1.0.0'`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `version = '2.0.0'` {
		t.Errorf("expected version = '2.0.0', got: %s", result)
	}
}

func TestPatchPythonSetup_DoubleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchPythonSetup()
	content := `from setuptools import setup

setup(
    name="mypackage",
    version="1.0.0",
    author="Test"
)
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `version="2.0.0"`) {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchPythonSetup_SingleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchPythonSetup()
	content := `setup(version='1.0.0')`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `version='2.0.0'`) {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchPythonSetup_NoVersion_ReturnsUnchanged(t *testing.T) {
	patcher := PatchPythonSetup()
	content := `setup(name="mypackage")`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}

func TestPatchRubyGemspec_DoubleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchRubyGemspec()
	content := `Gem::Specification.new do |spec|
  spec.name = "mygem"
  spec.version = "1.0.0"
end
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, `.version = "2.0.0"`) {
		t.Errorf("expected version to be patched, got: %s", result)
	}
}

func TestPatchRubyGemspec_SingleQuotes_PatchesVersion(t *testing.T) {
	patcher := PatchRubyGemspec()
	content := `spec.version = '1.0.0'`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != `spec.version = '2.0.0'` {
		t.Errorf("expected spec.version = '2.0.0', got: %s", result)
	}
}

func TestPatchRubyGemspec_NoVersion_ReturnsUnchanged(t *testing.T) {
	patcher := PatchRubyGemspec()
	content := `spec.name = "mygem"`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}

func TestPatchSwiftPackage_WithVersionComment_PatchesVersion(t *testing.T) {
	patcher := PatchSwiftPackage()
	content := `// swift-tools-version:5.5
// VERSION: 1.0.0

import PackageDescription
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(result, "// VERSION: 2.0.0") {
		t.Errorf("expected version comment to be patched, got: %s", result)
	}
}

func TestPatchSwiftPackage_NoVersionComment_ReturnsUnchanged(t *testing.T) {
	patcher := PatchSwiftPackage()
	content := `// swift-tools-version:5.5

import PackageDescription
`

	result, err := patcher(content, "2.0.0")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != content {
		t.Errorf("expected content to be unchanged")
	}
}
