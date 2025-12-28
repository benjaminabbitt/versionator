package plugin

import (
	"testing"
	"time"
)

func TestPluginTypeConstants_Always_AreCorrectlyDefined(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"PluginTypeEmit", PluginTypeEmit, "emit"},
		{"PluginTypeBuild", PluginTypeBuild, "build"},
		{"PluginTypePatch", PluginTypePatch, "patch"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("expected %s = %q, got %q", tc.name, tc.expected, tc.constant)
			}
		})
	}
}

func TestPluginPrefixConstants_Always_AreCorrectlyDefined(t *testing.T) {
	tests := []struct {
		name     string
		constant string
		expected string
	}{
		{"PluginPrefixEmit", PluginPrefixEmit, "versionator-plugin-emit-"},
		{"PluginPrefixBuild", PluginPrefixBuild, "versionator-plugin-build-"},
		{"PluginPrefixPatch", PluginPrefixPatch, "versionator-plugin-patch-"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.constant != tc.expected {
				t.Errorf("expected %s = %q, got %q", tc.name, tc.expected, tc.constant)
			}
		})
	}
}

func TestHandshake_Always_IsProperlyConfigured(t *testing.T) {
	if Handshake.ProtocolVersion != 1 {
		t.Errorf("expected ProtocolVersion = 1, got %d", Handshake.ProtocolVersion)
	}
	if Handshake.MagicCookieKey != "VERSIONATOR_PLUGIN" {
		t.Errorf("expected MagicCookieKey = VERSIONATOR_PLUGIN, got %s", Handshake.MagicCookieKey)
	}
	if Handshake.MagicCookieValue != "v1" {
		t.Errorf("expected MagicCookieValue = v1, got %s", Handshake.MagicCookieValue)
	}
}

func TestPluginMaps_Always_ContainCorrectTypes(t *testing.T) {
	if _, ok := EmitPluginMap[PluginTypeEmit]; !ok {
		t.Error("EmitPluginMap should contain PluginTypeEmit key")
	}
	if _, ok := BuildPluginMap[PluginTypeBuild]; !ok {
		t.Error("BuildPluginMap should contain PluginTypeBuild key")
	}
	if _, ok := PatchPluginMap[PluginTypePatch]; !ok {
		t.Error("PatchPluginMap should contain PluginTypePatch key")
	}
}

// Mock implementations for testing server-side behavior

type mockEmitImpl struct {
	name          string
	format        string
	fileExtension string
	defaultOutput string
	emitContent   string
	emitError     error
}

func (m *mockEmitImpl) Name() string          { return m.name }
func (m *mockEmitImpl) Format() string        { return m.format }
func (m *mockEmitImpl) FileExtension() string { return m.fileExtension }
func (m *mockEmitImpl) DefaultOutput() string { return m.defaultOutput }
func (m *mockEmitImpl) Emit(vars map[string]string) (string, error) {
	if m.emitError != nil {
		return "", m.emitError
	}
	return m.emitContent, nil
}

type mockBuildImpl struct {
	name       string
	format     string
	flags      string
	flagsError error
}

func (m *mockBuildImpl) Name() string   { return m.name }
func (m *mockBuildImpl) Format() string { return m.format }
func (m *mockBuildImpl) GenerateFlags(vars map[string]string) (string, error) {
	if m.flagsError != nil {
		return "", m.flagsError
	}
	return m.flags, nil
}

type mockPatchImpl struct {
	name        string
	filePattern string
	description string
	patchResult string
	patchError  error
}

func (m *mockPatchImpl) Name() string        { return m.name }
func (m *mockPatchImpl) FilePattern() string { return m.filePattern }
func (m *mockPatchImpl) Description() string { return m.description }
func (m *mockPatchImpl) Patch(content, version string) (string, error) {
	if m.patchError != nil {
		return "", m.patchError
	}
	return m.patchResult, nil
}

func TestEmitGRPCServerGetInfo_ValidImpl_ReturnsCorrectInfo(t *testing.T) {
	impl := &mockEmitImpl{
		name:          "test-emit",
		format:        "test",
		fileExtension: ".test",
		defaultOutput: "version.test",
	}
	server := &EmitGRPCServer{Impl: impl}

	info, err := server.GetInfo(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "test-emit" {
		t.Errorf("expected Name = test-emit, got %s", info.Name)
	}
	if info.Format != "test" {
		t.Errorf("expected Format = test, got %s", info.Format)
	}
	if info.FileExtension != ".test" {
		t.Errorf("expected FileExtension = .test, got %s", info.FileExtension)
	}
	if info.DefaultOutput != "version.test" {
		t.Errorf("expected DefaultOutput = version.test, got %s", info.DefaultOutput)
	}
}

func TestBuildGRPCServerGetInfo_ValidImpl_ReturnsCorrectInfo(t *testing.T) {
	impl := &mockBuildImpl{
		name:   "test-build",
		format: "test",
	}
	server := &BuildGRPCServer{Impl: impl}

	info, err := server.GetInfo(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "test-build" {
		t.Errorf("expected Name = test-build, got %s", info.Name)
	}
	if info.Format != "test" {
		t.Errorf("expected Format = test, got %s", info.Format)
	}
}

func TestPatchGRPCServerGetInfo_ValidImpl_ReturnsCorrectInfo(t *testing.T) {
	impl := &mockPatchImpl{
		name:        "test-patch",
		filePattern: "*.test",
		description: "Test patcher",
	}
	server := &PatchGRPCServer{Impl: impl}

	info, err := server.GetInfo(nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Name != "test-patch" {
		t.Errorf("expected Name = test-patch, got %s", info.Name)
	}
	if info.FilePattern != "*.test" {
		t.Errorf("expected FilePattern = *.test, got %s", info.FilePattern)
	}
	if info.Description != "Test patcher" {
		t.Errorf("expected Description = Test patcher, got %s", info.Description)
	}
}

func TestRpcTimeout_Always_IsTenSeconds(t *testing.T) {
	expected := 10 * time.Second
	if rpcTimeout != expected {
		t.Errorf("expected rpcTimeout = %v, got %v", expected, rpcTimeout)
	}
}
