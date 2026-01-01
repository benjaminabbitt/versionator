package plugin

import (
	"sync"
	"testing"
)

// mockPlugin implements Plugin interface for testing
type mockPlugin struct {
	name  string
	types PluginTypeSet
}

func (m *mockPlugin) Name() string         { return m.name }
func (m *mockPlugin) Types() PluginTypeSet { return m.types }

// mockTemplateProvider implements TemplateProvider for testing
type mockTemplateProvider struct {
	mockPlugin
	vars map[string]string
}

func (m *mockTemplateProvider) GetTemplateVariables(ctx map[string]string) map[string]string {
	return m.vars
}

// mockVersioningPlugin implements VersioningPlugin for testing
type mockVersioningPlugin struct {
	mockPlugin
	pattern string
}

func (m *mockVersioningPlugin) PatternName() string                    { return m.pattern }
func (m *mockVersioningPlugin) GetVersioningConfig() *VersioningConfig { return nil }

func TestNewPluginTypeSet_CreatesSet_WithAllTypes(t *testing.T) {
	set := NewPluginTypeSet(TypeVCS, TypeHook)

	if !set.Contains(TypeVCS) {
		t.Error("expected set to contain TypeVCS")
	}
	if !set.Contains(TypeHook) {
		t.Error("expected set to contain TypeHook")
	}
	if set.Contains(TypeVersioning) {
		t.Error("expected set to not contain TypeVersioning")
	}
}

func TestPluginTypeSetSlice_WithMultipleTypes_ReturnsAll(t *testing.T) {
	set := NewPluginTypeSet(TypeVCS, TypeHook)
	slice := set.Slice()

	if len(slice) != 2 {
		t.Errorf("expected 2 types, got %d", len(slice))
	}

	hasVCS := false
	hasHook := false
	for _, pt := range slice {
		if pt == TypeVCS {
			hasVCS = true
		}
		if pt == TypeHook {
			hasHook = true
		}
	}

	if !hasVCS || !hasHook {
		t.Error("slice missing expected types")
	}
}

func TestGetPlugins_Always_ReturnsCopy(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	plugin := &mockPlugin{name: "test", types: NewPluginTypeSet(TypeVCS)}
	Register(plugin)

	plugins1 := GetPlugins()
	plugins2 := GetPlugins()

	if len(plugins1) != 1 || len(plugins2) != 1 {
		t.Fatalf("expected 1 plugin in each slice")
	}

	// Modifying one should not affect the other
	plugins1[0] = nil
	if plugins2[0] == nil {
		t.Error("modifying returned slice affected other copy")
	}
}

func TestGetVersioningPlugins_Always_ReturnsCopy(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	vp := &mockVersioningPlugin{
		mockPlugin: mockPlugin{name: "semver-test", types: NewPluginTypeSet(TypeVersioning)},
		pattern:    "semver",
	}
	Register(vp)

	plugins1 := GetVersioningPlugins()
	plugins2 := GetVersioningPlugins()

	// Modifying one should not affect the other
	delete(plugins1, "semver")
	if _, ok := plugins2["semver"]; !ok {
		t.Error("modifying returned map affected other copy")
	}
}

func TestRegister_ConcurrentAccess_NoRace(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	var wg sync.WaitGroup
	numGoroutines := 100

	// Register concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			plugin := &mockPlugin{
				name:  "test",
				types: NewPluginTypeSet(TypeVCS),
			}
			Register(plugin)
		}(i)
	}

	// Read concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = GetPlugins()
			_ = GetVersioningPlugins()
		}()
	}

	wg.Wait()

	plugins := GetPlugins()
	if len(plugins) != numGoroutines {
		t.Errorf("expected %d plugins, got %d", numGoroutines, len(plugins))
	}
}

func TestGetAllTemplateVariables_MultipleProviders_CollectsAll(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	tp1 := &mockTemplateProvider{
		mockPlugin: mockPlugin{name: "tp1", types: NewPluginTypeSet(TypeTemplateProvider)},
		vars:       map[string]string{"Var1": "value1"},
	}
	tp2 := &mockTemplateProvider{
		mockPlugin: mockPlugin{name: "tp2", types: NewPluginTypeSet(TypeTemplateProvider)},
		vars:       map[string]string{"Var2": "value2"},
	}

	RegisterTemplateProvider(tp1)
	RegisterTemplateProvider(tp2)

	vars := GetAllTemplateVariables(nil)

	if vars["Var1"] != "value1" {
		t.Errorf("expected Var1=value1, got %s", vars["Var1"])
	}
	if vars["Var2"] != "value2" {
		t.Errorf("expected Var2=value2, got %s", vars["Var2"])
	}
}

func TestGetPluginsByType_MixedTypes_FiltersCorrectly(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	vcsPlugin := &mockPlugin{name: "vcs", types: NewPluginTypeSet(TypeVCS)}
	hookPlugin := &mockPlugin{name: "hook", types: NewPluginTypeSet(TypeHook)}
	multiPlugin := &mockPlugin{name: "multi", types: NewPluginTypeSet(TypeVCS, TypeHook)}

	Register(vcsPlugin)
	Register(hookPlugin)
	Register(multiPlugin)

	vcsPlugins := GetPluginsByType(TypeVCS)
	if len(vcsPlugins) != 2 {
		t.Errorf("expected 2 VCS plugins, got %d", len(vcsPlugins))
	}

	hookPlugins := GetPluginsByType(TypeHook)
	if len(hookPlugins) != 2 {
		t.Errorf("expected 2 Hook plugins, got %d", len(hookPlugins))
	}

	versioningPlugins := GetPluginsByType(TypeVersioning)
	if len(versioningPlugins) != 0 {
		t.Errorf("expected 0 Versioning plugins, got %d", len(versioningPlugins))
	}
}

func TestGetVersioningPlugin_Registered_ReturnsPlugin(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	vp := &mockVersioningPlugin{
		mockPlugin: mockPlugin{name: "semver-plugin", types: NewPluginTypeSet(TypeVersioning)},
		pattern:    "semver",
	}
	Register(vp)

	plugin, ok := GetVersioningPlugin("semver")
	if !ok {
		t.Error("expected to find 'semver' versioning plugin")
	}
	if plugin.Name() != "semver-plugin" {
		t.Errorf("expected plugin name 'semver-plugin', got '%s'", plugin.Name())
	}

	_, ok = GetVersioningPlugin("calver")
	if ok {
		t.Error("expected not to find 'calver' versioning plugin")
	}
}

func TestIsVersioningPatternSupported_RegisteredPattern_ReturnsTrue(t *testing.T) {
	t.Cleanup(ResetRegistry)
	ResetRegistry()

	vp := &mockVersioningPlugin{
		mockPlugin: mockPlugin{name: "semver-plugin", types: NewPluginTypeSet(TypeVersioning)},
		pattern:    "semver",
	}
	Register(vp)

	if !IsVersioningPatternSupported("semver") {
		t.Error("expected 'semver' to be supported")
	}
	if IsVersioningPatternSupported("calver") {
		t.Error("expected 'calver' to not be supported")
	}
}
