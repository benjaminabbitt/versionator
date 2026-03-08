package mode

import (
	"testing"

	"github.com/benjaminabbitt/versionator/internal/config"
	"github.com/benjaminabbitt/versionator/internal/version"
)

func TestGetMode_DefaultsToRelease(t *testing.T) {
	mode := GetMode(nil)
	if mode.Name() != "release" {
		t.Errorf("expected release mode, got %s", mode.Name())
	}
}

func TestGetMode_ReleaseMode(t *testing.T) {
	cfg := &config.Config{
		Mode: config.ModeConfig{
			Type: "release",
		},
	}
	mode := GetMode(cfg)
	if mode.Name() != "release" {
		t.Errorf("expected release mode, got %s", mode.Name())
	}
}

func TestGetMode_ContinuousDeliveryMode(t *testing.T) {
	cfg := &config.Config{
		Mode: config.ModeConfig{
			Type: "continuous-delivery",
			ContinuousDelivery: config.CDModeConfig{
				PrereleaseTemplate: "build-{{CommitsSinceTag}}",
				MetadataTemplate:   "{{ShortHash}}",
			},
		},
	}
	mode := GetMode(cfg)
	if mode.Name() != "continuous-delivery" {
		t.Errorf("expected continuous-delivery mode, got %s", mode.Name())
	}
}

func TestGetMode_EmptyTypeDefaultsToRelease(t *testing.T) {
	cfg := &config.Config{}
	mode := GetMode(cfg)
	if mode.Name() != "release" {
		t.Errorf("expected release mode for empty type, got %s", mode.Name())
	}
}

func TestReleaseMode_Name(t *testing.T) {
	mode := &ReleaseMode{}
	if mode.Name() != "release" {
		t.Errorf("expected 'release', got %s", mode.Name())
	}
}

func TestReleaseMode_IsReleaseMode(t *testing.T) {
	mode := &ReleaseMode{}
	if !mode.IsReleaseMode() {
		t.Error("expected IsReleaseMode to return true")
	}
}

func TestReleaseMode_GetPreRelease(t *testing.T) {
	mode := &ReleaseMode{}
	v := &version.Version{
		Major:      1,
		Minor:      2,
		Patch:      3,
		PreRelease: "alpha.1",
	}

	result, err := mode.GetPreRelease(v, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "alpha.1" {
		t.Errorf("expected 'alpha.1', got %s", result)
	}
}

func TestReleaseMode_GetMetadata(t *testing.T) {
	mode := &ReleaseMode{}
	v := &version.Version{
		Major:         1,
		Minor:         2,
		Patch:         3,
		BuildMetadata: "build.123",
	}

	result, err := mode.GetMetadata(v, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "build.123" {
		t.Errorf("expected 'build.123', got %s", result)
	}
}

func TestContinuousDeliveryMode_Name(t *testing.T) {
	mode := &ContinuousDeliveryMode{}
	if mode.Name() != "continuous-delivery" {
		t.Errorf("expected 'continuous-delivery', got %s", mode.Name())
	}
}

func TestContinuousDeliveryMode_IsReleaseMode(t *testing.T) {
	mode := &ContinuousDeliveryMode{}
	if mode.IsReleaseMode() {
		t.Error("expected IsReleaseMode to return false")
	}
}

func TestContinuousDeliveryMode_GetPreRelease_DefaultTemplate(t *testing.T) {
	mode := &ContinuousDeliveryMode{}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := map[string]string{
		"CommitsSinceTag": "42",
	}

	result, err := mode.GetPreRelease(v, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "build-42" {
		t.Errorf("expected 'build-42', got %s", result)
	}
}

func TestContinuousDeliveryMode_GetPreRelease_CustomTemplate(t *testing.T) {
	mode := &ContinuousDeliveryMode{
		PrereleaseTemplate: "dev-{{CommitsSinceTag}}-{{ShortHash}}",
	}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := map[string]string{
		"CommitsSinceTag": "5",
		"ShortHash":       "abc1234",
	}

	result, err := mode.GetPreRelease(v, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "dev-5-abc1234" {
		t.Errorf("expected 'dev-5-abc1234', got %s", result)
	}
}

func TestContinuousDeliveryMode_GetMetadata_DefaultTemplate(t *testing.T) {
	mode := &ContinuousDeliveryMode{}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := map[string]string{
		"ShortHash": "abc1234",
	}

	result, err := mode.GetMetadata(v, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "abc1234" {
		t.Errorf("expected 'abc1234', got %s", result)
	}
}

func TestContinuousDeliveryMode_GetMetadata_CustomTemplate(t *testing.T) {
	mode := &ContinuousDeliveryMode{
		MetadataTemplate: "{{BuildDateTimeCompact}}.{{ShortHash}}",
	}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}
	data := map[string]string{
		"BuildDateTimeCompact": "20241215103045",
		"ShortHash":            "abc1234",
	}

	result, err := mode.GetMetadata(v, data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "20241215103045.abc1234" {
		t.Errorf("expected '20241215103045.abc1234', got %s", result)
	}
}

func TestContinuousDeliveryMode_GetPreRelease_EmptyData(t *testing.T) {
	mode := &ContinuousDeliveryMode{
		PrereleaseTemplate: "static-prerelease",
	}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}

	result, err := mode.GetPreRelease(v, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "static-prerelease" {
		t.Errorf("expected 'static-prerelease', got %s", result)
	}
}

func TestContinuousDeliveryMode_GetMetadata_EmptyData(t *testing.T) {
	mode := &ContinuousDeliveryMode{
		MetadataTemplate: "static-metadata",
	}
	v := &version.Version{Major: 1, Minor: 2, Patch: 3}

	result, err := mode.GetMetadata(v, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "static-metadata" {
		t.Errorf("expected 'static-metadata', got %s", result)
	}
}

func TestRenderTemplate_EmptyTemplate(t *testing.T) {
	result, err := renderTemplate("", nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "" {
		t.Errorf("expected empty string, got %s", result)
	}
}

func TestRenderTemplate_WithData(t *testing.T) {
	data := map[string]string{
		"Foo": "bar",
		"Baz": "qux",
	}

	result, err := renderTemplate("{{Foo}}-{{Baz}}", data)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "bar-qux" {
		t.Errorf("expected 'bar-qux', got %s", result)
	}
}
