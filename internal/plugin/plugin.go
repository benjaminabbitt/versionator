package plugin

// PluginType represents the type of plugin capability.
// Values match interface names for reflective discovery.
type PluginType string

const (
	// TypeVCS indicates a version control system plugin
	TypeVCS PluginType = "VCS"

	// TypeTemplateProvider indicates a plugin that provides template variables
	TypeTemplateProvider PluginType = "TemplateProvider"

	// TypeOutput indicates an output format plugin
	TypeOutput PluginType = "Output"

	// TypeHook indicates a lifecycle hook plugin
	TypeHook PluginType = "Hook"
)

// PluginTypeSet represents a set of plugin types
type PluginTypeSet map[PluginType]struct{}

// NewPluginTypeSet creates a new set from the given types
func NewPluginTypeSet(types ...PluginType) PluginTypeSet {
	set := make(PluginTypeSet)
	for _, t := range types {
		set[t] = struct{}{}
	}
	return set
}

// Contains checks if the set contains the given type
func (s PluginTypeSet) Contains(t PluginType) bool {
	_, ok := s[t]
	return ok
}

// Slice returns the types as a slice
func (s PluginTypeSet) Slice() []PluginType {
	result := make([]PluginType, 0, len(s))
	for t := range s {
		result = append(result, t)
	}
	return result
}

// Plugin is the base interface all plugins must implement
type Plugin interface {
	// Name returns the name of the plugin
	Name() string

	// Types returns the set of plugin types this plugin implements.
	// Type names match interface names for reflective discovery.
	Types() PluginTypeSet
}

// TemplateProvider is an interface for plugins that can provide template variables
type TemplateProvider interface {
	Plugin

	// GetTemplateVariables returns plugin-specific template variables
	// The context map provides access to existing variables the plugin might need
	// (e.g., ShortHash for creating prefixed hash variables)
	GetTemplateVariables(context map[string]string) map[string]string
}

// Registry holds all registered plugins
type Registry struct {
	plugins           []Plugin
	templateProviders []TemplateProvider
}

// globalRegistry is the default plugin registry
var globalRegistry = &Registry{}

// Register adds a plugin to the global registry
func Register(p Plugin) {
	globalRegistry.plugins = append(globalRegistry.plugins, p)

	// Also register as template provider if it implements the interface
	if tp, ok := p.(TemplateProvider); ok {
		globalRegistry.templateProviders = append(globalRegistry.templateProviders, tp)
	}
}

// RegisterTemplateProvider adds a template provider to the global registry
func RegisterTemplateProvider(provider TemplateProvider) {
	globalRegistry.templateProviders = append(globalRegistry.templateProviders, provider)
	globalRegistry.plugins = append(globalRegistry.plugins, provider)
}

// GetAllTemplateVariables collects template variables from all registered plugins
func GetAllTemplateVariables(context map[string]string) map[string]string {
	result := make(map[string]string)
	for _, provider := range globalRegistry.templateProviders {
		vars := provider.GetTemplateVariables(context)
		for k, v := range vars {
			result[k] = v
		}
	}
	return result
}

// GetPlugins returns all registered plugins
func GetPlugins() []Plugin {
	return globalRegistry.plugins
}

// GetTemplateProviders returns all registered template providers
func GetTemplateProviders() []TemplateProvider {
	return globalRegistry.templateProviders
}

// GetPluginsByType returns all plugins that implement a specific type
func GetPluginsByType(pluginType PluginType) []Plugin {
	var result []Plugin
	for _, p := range globalRegistry.plugins {
		if p.Types().Contains(pluginType) {
			result = append(result, p)
		}
	}
	return result
}
