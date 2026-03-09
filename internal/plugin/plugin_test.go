package plugin

import (
	"testing"
)

// =============================================================================
// TEST HELPERS
// =============================================================================

// mockPlugin is a simple plugin for testing that implements the Plugin interface.
type mockPlugin struct {
	name  string
	types PluginTypeSet
}

func (m *mockPlugin) Name() string {
	return m.name
}

func (m *mockPlugin) Types() PluginTypeSet {
	return m.types
}

// mockTemplateProvider is a template provider for testing that implements
// both Plugin and TemplateProvider interfaces.
type mockTemplateProvider struct {
	mockPlugin
	variables map[string]string
}

func (m *mockTemplateProvider) GetTemplateVariables(context map[string]string) map[string]string {
	return m.variables
}

// saveAndClearRegistry saves the current global registry state and clears it.
// Returns a cleanup function that restores the original state.
func saveAndClearRegistry() func() {
	oldPlugins := globalRegistry.plugins
	oldProviders := globalRegistry.templateProviders
	globalRegistry.plugins = nil
	globalRegistry.templateProviders = nil
	return func() {
		globalRegistry.plugins = oldPlugins
		globalRegistry.templateProviders = oldProviders
	}
}

// =============================================================================
// CORE FUNCTIONALITY
// Tests for the primary happy path operations that must work correctly.
// =============================================================================

// TestNewPluginTypeSet_SingleType validates that a PluginTypeSet can be created
// with a single plugin type.
//
// Why: The PluginTypeSet is the foundation for categorizing plugins. Creating
// a set with a single type is the most common use case and must work correctly.
//
// What: Create a PluginTypeSet with one type, verify it has exactly one element
// and contains that type.
func TestNewPluginTypeSet_SingleType(t *testing.T) {
	// Action: Create a set with a single type
	set := NewPluginTypeSet(TypeVCS)

	// Expected: Set contains exactly one element
	if len(set) != 1 {
		t.Errorf("expected 1 element, got %d", len(set))
	}

	// Expected: Set contains the specified type
	if !set.Contains(TypeVCS) {
		t.Error("expected set to contain TypeVCS")
	}
}

// TestPluginTypeSet_Contains validates that the Contains method correctly
// identifies membership in the set.
//
// Why: The Contains method is used throughout the plugin system to check
// if a plugin supports a particular capability. False positives or negatives
// would break plugin filtering.
//
// What: Create a set with specific types, verify Contains returns true for
// those types and false for others.
func TestPluginTypeSet_Contains(t *testing.T) {
	// Precondition: Create a set with VCS and TemplateProvider types
	set := NewPluginTypeSet(TypeVCS, TypeTemplateProvider)

	// Expected: Contains returns true for types in the set
	if !set.Contains(TypeVCS) {
		t.Error("expected set to contain TypeVCS")
	}
	if !set.Contains(TypeTemplateProvider) {
		t.Error("expected set to contain TypeTemplateProvider")
	}

	// Expected: Contains returns false for types not in the set
	if set.Contains(TypeOutput) {
		t.Error("expected set to NOT contain TypeOutput")
	}
	if set.Contains(TypeHook) {
		t.Error("expected set to NOT contain TypeHook")
	}
}

// TestPluginTypeSet_Slice validates that a PluginTypeSet can be converted
// to a slice for iteration.
//
// Why: Converting to a slice is necessary for displaying plugin types or
// iterating when map iteration is not suitable.
//
// What: Create a set with multiple types, convert to slice, verify all
// types are present (order is not guaranteed due to map iteration).
func TestPluginTypeSet_Slice(t *testing.T) {
	// Precondition: Create a set with two types
	set := NewPluginTypeSet(TypeVCS, TypeTemplateProvider)

	// Action: Convert to slice
	slice := set.Slice()

	// Expected: Slice has correct length
	if len(slice) != 2 {
		t.Errorf("expected slice length 2, got %d", len(slice))
	}

	// Expected: Slice contains both types (order not guaranteed)
	hasVCS := false
	hasTemplateProvider := false
	for _, pt := range slice {
		if pt == TypeVCS {
			hasVCS = true
		}
		if pt == TypeTemplateProvider {
			hasTemplateProvider = true
		}
	}
	if !hasVCS {
		t.Error("expected slice to contain TypeVCS")
	}
	if !hasTemplateProvider {
		t.Error("expected slice to contain TypeTemplateProvider")
	}
}

// TestRegister validates that plugins can be registered and retrieved
// from the global registry.
//
// Why: Plugin registration is the core mechanism for extensibility. Without
// working registration, no plugins would be discoverable.
//
// What: Register a mock plugin, verify it appears in GetPlugins results
// with the correct name.
func TestRegister(t *testing.T) {
	// Precondition: Clear registry to isolate test
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Precondition: Create a mock plugin
	plugin := &mockPlugin{
		name:  "test-plugin",
		types: NewPluginTypeSet(TypeVCS),
	}

	// Action: Register the plugin
	Register(plugin)

	// Expected: Plugin is retrievable
	plugins := GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
	if plugins[0].Name() != "test-plugin" {
		t.Errorf("expected name 'test-plugin', got '%s'", plugins[0].Name())
	}
}

// TestRegisterTemplateProvider validates that template providers are
// registered both as providers and as plugins.
//
// Why: Template providers need to be accessible via GetTemplateProviders
// for variable collection, and via GetPlugins for general plugin discovery.
//
// What: Register a template provider, verify it appears in both provider
// and plugin lists.
func TestRegisterTemplateProvider(t *testing.T) {
	// Precondition: Clear registry to isolate test
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Precondition: Create a mock template provider
	provider := &mockTemplateProvider{
		mockPlugin: mockPlugin{
			name:  "test-provider",
			types: NewPluginTypeSet(TypeTemplateProvider),
		},
		variables: map[string]string{
			"TestVar": "test-value",
		},
	}

	// Action: Register the template provider
	RegisterTemplateProvider(provider)

	// Expected: Provider is in the providers list
	providers := GetTemplateProviders()
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}

	// Expected: Provider is also in the plugins list
	plugins := GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}
}

// TestGetAllTemplateVariables validates that template variables are
// collected from all registered providers.
//
// Why: Template rendering depends on aggregating variables from all
// providers. Missing variables would break template output.
//
// What: Register multiple providers with different variables, verify
// GetAllTemplateVariables returns all of them.
func TestGetAllTemplateVariables(t *testing.T) {
	// Precondition: Clear registry to isolate test
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Precondition: Register two providers with different variables
	provider1 := &mockTemplateProvider{
		mockPlugin: mockPlugin{name: "provider1", types: NewPluginTypeSet(TypeTemplateProvider)},
		variables:  map[string]string{"Var1": "value1"},
	}
	provider2 := &mockTemplateProvider{
		mockPlugin: mockPlugin{name: "provider2", types: NewPluginTypeSet(TypeTemplateProvider)},
		variables:  map[string]string{"Var2": "value2"},
	}
	RegisterTemplateProvider(provider1)
	RegisterTemplateProvider(provider2)

	// Action: Get all template variables
	vars := GetAllTemplateVariables(nil)

	// Expected: Variables map is not nil
	if vars == nil {
		t.Error("expected non-nil variables map")
	}

	// Expected: All provider variables are present
	if vars["Var1"] != "value1" {
		t.Errorf("expected Var1='value1', got '%s'", vars["Var1"])
	}
	if vars["Var2"] != "value2" {
		t.Errorf("expected Var2='value2', got '%s'", vars["Var2"])
	}
}

// =============================================================================
// KEY VARIATIONS
// Tests for important alternate flows and multi-type scenarios.
// =============================================================================

// TestNewPluginTypeSet_MultipleTypes validates that a PluginTypeSet can hold
// multiple plugin types simultaneously.
//
// Why: Many plugins provide multiple capabilities (e.g., VCS that also
// provides template variables). The set must correctly track all types.
//
// What: Create a set with three types, verify all are present and the
// Contains method works for each.
func TestNewPluginTypeSet_MultipleTypes(t *testing.T) {
	// Action: Create a set with multiple types
	set := NewPluginTypeSet(TypeVCS, TypeTemplateProvider, TypeOutput)

	// Expected: Set contains correct number of elements
	if len(set) != 3 {
		t.Errorf("expected 3 elements, got %d", len(set))
	}

	// Expected: All types are contained
	if !set.Contains(TypeVCS) {
		t.Error("expected set to contain TypeVCS")
	}
	if !set.Contains(TypeTemplateProvider) {
		t.Error("expected set to contain TypeTemplateProvider")
	}
	if !set.Contains(TypeOutput) {
		t.Error("expected set to contain TypeOutput")
	}
}

// TestGetPluginsByType validates that plugins can be filtered by their
// declared type capabilities.
//
// Why: Different subsystems need access to specific plugin types. The
// VCS system needs VCS plugins, template rendering needs template providers.
//
// What: Register plugins of different types including a multi-type plugin,
// verify GetPluginsByType correctly filters by type.
func TestGetPluginsByType(t *testing.T) {
	// Precondition: Clear registry to isolate test
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Precondition: Register plugins of different types
	vcsPlugin := &mockPlugin{
		name:  "vcs-plugin",
		types: NewPluginTypeSet(TypeVCS),
	}
	templatePlugin := &mockPlugin{
		name:  "template-plugin",
		types: NewPluginTypeSet(TypeTemplateProvider),
	}
	multiPlugin := &mockPlugin{
		name:  "multi-plugin",
		types: NewPluginTypeSet(TypeVCS, TypeTemplateProvider),
	}
	Register(vcsPlugin)
	Register(templatePlugin)
	Register(multiPlugin)

	// Action/Expected: Get VCS plugins - should include vcs-plugin and multi-plugin
	vcsPlugins := GetPluginsByType(TypeVCS)
	if len(vcsPlugins) != 2 {
		t.Errorf("expected 2 VCS plugins, got %d", len(vcsPlugins))
	}

	// Action/Expected: Get template providers - should include template-plugin and multi-plugin
	templatePlugins := GetPluginsByType(TypeTemplateProvider)
	if len(templatePlugins) != 2 {
		t.Errorf("expected 2 template provider plugins, got %d", len(templatePlugins))
	}

	// Action/Expected: Get hook plugins - none registered
	hookPlugins := GetPluginsByType(TypeHook)
	if len(hookPlugins) != 0 {
		t.Errorf("expected 0 hook plugins, got %d", len(hookPlugins))
	}
}

// TestRegister_WithTemplateProvider validates that registering a plugin
// that implements TemplateProvider also registers it as a provider.
//
// Why: Plugins implementing TemplateProvider should be auto-discovered
// even when registered via Register() rather than RegisterTemplateProvider().
// This enables simpler registration for multi-capability plugins.
//
// What: Register a TemplateProvider via Register(), verify it appears in
// both plugin and provider lists, and its variables are accessible.
func TestRegister_WithTemplateProvider(t *testing.T) {
	// Precondition: Clear registry to isolate test
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Precondition: Create a dual-purpose plugin implementing TemplateProvider
	provider := &mockTemplateProvider{
		mockPlugin: mockPlugin{
			name:  "dual-plugin",
			types: NewPluginTypeSet(TypeVCS, TypeTemplateProvider),
		},
		variables: map[string]string{"DualVar": "dual-value"},
	}

	// Action: Register via Register() not RegisterTemplateProvider()
	Register(provider)

	// Expected: Plugin is in the plugins list
	plugins := GetPlugins()
	if len(plugins) != 1 {
		t.Errorf("expected 1 plugin, got %d", len(plugins))
	}

	// Expected: Plugin is also auto-registered as a provider
	providers := GetTemplateProviders()
	if len(providers) != 1 {
		t.Errorf("expected 1 provider, got %d", len(providers))
	}

	// Expected: Variables are accessible
	vars := GetAllTemplateVariables(nil)
	if vars["DualVar"] != "dual-value" {
		t.Errorf("expected DualVar='dual-value', got '%s'", vars["DualVar"])
	}
}

// =============================================================================
// ERROR HANDLING
// Tests for expected failure modes and graceful degradation.
// Note: The plugin system is designed to fail gracefully with empty returns
// rather than errors, so these tests verify that behavior.
// =============================================================================

// TestGetAllTemplateVariables_EmptyRegistry validates that GetAllTemplateVariables
// returns an empty map when no providers are registered.
//
// Why: An empty registry should not cause panics or nil pointer errors.
// The system should gracefully return an empty map.
//
// What: Clear the registry, call GetAllTemplateVariables, verify it returns
// a non-nil empty map.
func TestGetAllTemplateVariables_EmptyRegistry(t *testing.T) {
	// Precondition: Clear registry to simulate empty state
	cleanup := saveAndClearRegistry()
	defer cleanup()

	// Action: Get template variables from empty registry
	vars := GetAllTemplateVariables(nil)

	// Expected: Returns non-nil empty map, not nil
	if vars == nil {
		t.Error("expected non-nil variables map even with empty registry")
	}
	if len(vars) != 0 {
		t.Errorf("expected empty map, got %d entries", len(vars))
	}
}

// =============================================================================
// EDGE CASES
// Tests for boundary conditions and unusual but valid inputs.
// =============================================================================

// TestNewPluginTypeSet_Empty validates that an empty PluginTypeSet can be
// created and behaves correctly.
//
// Why: Edge case where a plugin might declare no types (though unusual).
// The system should handle this without panics.
//
// What: Create a PluginTypeSet with no arguments, verify it's empty.
func TestNewPluginTypeSet_Empty(t *testing.T) {
	// Action: Create an empty set
	set := NewPluginTypeSet()

	// Expected: Set is empty
	if len(set) != 0 {
		t.Errorf("expected empty set, got %d elements", len(set))
	}
}

// TestNewPluginTypeSet_DuplicatesDeduped validates that duplicate types
// in the input are deduplicated.
//
// Why: Callers might accidentally pass the same type multiple times.
// The set semantics require each type appears at most once.
//
// What: Create a set with duplicate types, verify the result has no duplicates.
func TestNewPluginTypeSet_DuplicatesDeduped(t *testing.T) {
	// Action: Create a set with duplicate types
	set := NewPluginTypeSet(TypeVCS, TypeVCS, TypeVCS)

	// Expected: Duplicates are removed, only one element
	if len(set) != 1 {
		t.Errorf("expected 1 element after dedup, got %d", len(set))
	}
}

// TestPluginTypeSet_Slice_Empty validates that converting an empty set
// to a slice returns an empty slice.
//
// Why: Callers iterating over the slice should not encounter nil or
// unexpected behavior with an empty set.
//
// What: Create an empty set, convert to slice, verify empty slice returned.
func TestPluginTypeSet_Slice_Empty(t *testing.T) {
	// Precondition: Create an empty set
	set := NewPluginTypeSet()

	// Action: Convert to slice
	slice := set.Slice()

	// Expected: Slice is empty (not nil, but zero length)
	if len(slice) != 0 {
		t.Errorf("expected empty slice, got %d elements", len(slice))
	}
}

// =============================================================================
// MINUTIAE
// Tests for implementation details and constant verification.
// =============================================================================

// TestPluginTypeConstants validates that the plugin type constants have
// their expected string values.
//
// Why: These constants may be used for serialization, logging, or
// configuration. Changing their values would be a breaking change.
//
// What: Verify each PluginType constant has its expected string value.
func TestPluginTypeConstants(t *testing.T) {
	// Expected: Each constant has its documented value
	if TypeVCS != "VCS" {
		t.Errorf("expected TypeVCS='VCS', got '%s'", TypeVCS)
	}
	if TypeTemplateProvider != "TemplateProvider" {
		t.Errorf("expected TypeTemplateProvider='TemplateProvider', got '%s'", TypeTemplateProvider)
	}
	if TypeOutput != "Output" {
		t.Errorf("expected TypeOutput='Output', got '%s'", TypeOutput)
	}
	if TypeHook != "Hook" {
		t.Errorf("expected TypeHook='Hook', got '%s'", TypeHook)
	}
}
