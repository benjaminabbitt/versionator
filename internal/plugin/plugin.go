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

	// TypeLanguage indicates a language configuration plugin
	TypeLanguage PluginType = "Language"

	// TypeVersioning indicates a versioning pattern plugin
	TypeVersioning PluginType = "Versioning"
)

// InjectionMethod represents how version information is injected into the application
type InjectionMethod string

const (
	// InjectionEmit generates a source file with version constants/variables
	// Works for all languages (compiled and interpreted)
	InjectionEmit InjectionMethod = "emit"

	// InjectionLink overrides variables at link/load time using linker flags
	// Only works for compiled languages (Go, Rust, C, C++, etc.)
	// Examples:
	//   Go:   -ldflags "-X main.Version=1.2.3"
	//   Rust: RUSTFLAGS with --cfg or build.rs
	//   C/C++: -DVERSION="1.2.3"
	InjectionLink InjectionMethod = "link"

	// InjectionPatch updates version in existing config/manifest files
	// Examples:
	//   Node.js: package.json "version" field
	//   Rust: Cargo.toml version field
	//   Python: pyproject.toml version field
	//   Java/Maven: pom.xml <version> element
	InjectionPatch InjectionMethod = "patch"
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

// EmitConfig holds configuration for source file emission
type EmitConfig struct {
	// DefaultOutputPath is the default path for the generated version file
	// Example: "internal/version/version.go", "_version.py", "src/version.rs"
	DefaultOutputPath string

	// DefaultPackageName is the default package/module name for the generated code
	// Example: "version" for Go, "" for languages that don't need it
	DefaultPackageName string

	// FileExtension is the file extension for the language
	// Example: ".go", ".py", ".rs"
	FileExtension string
}

// LinkConfig holds configuration for link-time variable injection
type LinkConfig struct {
	// VariablePath is the default variable path to override
	// Go example: "main.Version" or "github.com/user/pkg/version.Version"
	// C/C++ example: "VERSION"
	VariablePath string

	// FlagTemplate is the flag template using {{Variable}} and {{Value}} placeholders
	// Go example: "-X {{Variable}}={{Value}}"
	// C example: "-D{{Variable}}={{Value}}"
	// User incorporates output into their build command
	FlagTemplate string
}

// PatchFormat specifies the file format for config/manifest patching
type PatchFormat string

const (
	PatchFormatJSON   PatchFormat = "json"   // JSON files (package.json, composer.json)
	PatchFormatTOML   PatchFormat = "toml"   // TOML files (Cargo.toml, pyproject.toml)
	PatchFormatYAML   PatchFormat = "yaml"   // YAML files (pubspec.yaml)
	PatchFormatXML    PatchFormat = "xml"    // XML files (pom.xml, *.csproj)
	PatchFormatPLIST  PatchFormat = "plist"  // Property list (Info.plist)
	PatchFormatPython PatchFormat = "python" // Python code (setup.py)
	PatchFormatRuby   PatchFormat = "ruby"   // Ruby code (*.gemspec)
	PatchFormatSwift  PatchFormat = "swift"  // Swift comments (// VERSION: x.y.z)
)

// PatchFunc is the standard signature for patch functions.
// It takes the file content and new version, returns patched content or error.
type PatchFunc func(content, version string) (string, error)

// PatchConfig holds configuration for patching a config/manifest file
type PatchConfig struct {
	// Name is a human-readable name for this patch target
	// Example: "package.json", "Cargo.toml", "Maven POM"
	Name string

	// FilePath is the default path to the file to patch
	// Example: "package.json", "Cargo.toml", "pom.xml"
	FilePath string

	// Format specifies the file format for parsing (used for documentation/identification)
	Format PatchFormat

	// VersionPath is the path to the version field within the file
	// JSON/TOML: dot-notation like "version" or "package.version"
	// XML: XPath-like "project.version" or "/project/version"
	VersionPath string

	// Description explains what this patch target is for
	Description string

	// Patch is the function that performs the actual patching
	Patch PatchFunc
}

// LanguagePlugin is an interface for plugins that provide language-specific configuration.
// Each method returns nil if that injection method is not supported by the language.
type LanguagePlugin interface {
	Plugin

	// LanguageName returns the language identifier (e.g., "go", "python")
	LanguageName() string

	// GetEmitConfig returns configuration for source file emission.
	// Returns nil if emit is not supported (all languages should support emit).
	GetEmitConfig() *EmitConfig

	// GetBuildConfig returns configuration for link-time variable injection.
	// Returns nil if link injection is not supported (e.g., interpreted languages).
	GetBuildConfig() *LinkConfig

	// GetPatchConfigs returns configuration for manifest/config file patching.
	// Returns nil or empty slice if patching is not supported.
	GetPatchConfigs() []PatchConfig
}

// VersioningConfig holds versioning pattern configuration
// All patterns use three-dot semver (Major.Minor.Patch) as the base
type VersioningConfig struct {
	Name              string   // Pattern name (e.g., "go", "standard")
	Prefix            string   // Version prefix (e.g., "v" for Go)
	PreReleaseElements []string // Pre-release elements (joined with dashes)
	MetadataElements   []string // Metadata elements (joined with dots)
}

// VersioningPlugin is an interface for plugins that provide versioning pattern configuration
type VersioningPlugin interface {
	Plugin

	// PatternName returns the pattern identifier (e.g., "go", "standard")
	PatternName() string

	// GetVersioningConfig returns the versioning pattern configuration
	GetVersioningConfig() *VersioningConfig
}

// Registry holds all registered plugins
type Registry struct {
	plugins            []Plugin
	templateProviders  []TemplateProvider
	languagePlugins    map[string]LanguagePlugin
	versioningPlugins  map[string]VersioningPlugin
}

// globalRegistry is the default plugin registry
var globalRegistry = &Registry{
	languagePlugins:   make(map[string]LanguagePlugin),
	versioningPlugins: make(map[string]VersioningPlugin),
}

// Register adds a plugin to the global registry
func Register(p Plugin) {
	globalRegistry.plugins = append(globalRegistry.plugins, p)

	// Also register as template provider if it implements the interface
	if tp, ok := p.(TemplateProvider); ok {
		globalRegistry.templateProviders = append(globalRegistry.templateProviders, tp)
	}

	// Also register as language plugin if it implements the interface
	if lp, ok := p.(LanguagePlugin); ok {
		globalRegistry.languagePlugins[lp.LanguageName()] = lp
	}

	// Also register as versioning plugin if it implements the interface
	if vp, ok := p.(VersioningPlugin); ok {
		globalRegistry.versioningPlugins[vp.PatternName()] = vp
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

// GetLanguagePlugin returns a language plugin by name
func GetLanguagePlugin(name string) (LanguagePlugin, bool) {
	lp, ok := globalRegistry.languagePlugins[name]
	return lp, ok
}

// GetLanguagePlugins returns all registered language plugins
func GetLanguagePlugins() map[string]LanguagePlugin {
	return globalRegistry.languagePlugins
}

// GetSupportedLanguages returns the list of supported language names
func GetSupportedLanguages() []string {
	names := make([]string, 0, len(globalRegistry.languagePlugins))
	for name := range globalRegistry.languagePlugins {
		names = append(names, name)
	}
	return names
}

// IsLanguageSupported checks if a language is supported
func IsLanguageSupported(name string) bool {
	_, ok := globalRegistry.languagePlugins[name]
	return ok
}

// GetVersioningPlugin returns a versioning plugin by name
func GetVersioningPlugin(name string) (VersioningPlugin, bool) {
	vp, ok := globalRegistry.versioningPlugins[name]
	return vp, ok
}

// GetVersioningPlugins returns all registered versioning plugins
func GetVersioningPlugins() map[string]VersioningPlugin {
	return globalRegistry.versioningPlugins
}

// GetSupportedVersioningPatterns returns the list of supported versioning pattern names
func GetSupportedVersioningPatterns() []string {
	names := make([]string, 0, len(globalRegistry.versioningPlugins))
	for name := range globalRegistry.versioningPlugins {
		names = append(names, name)
	}
	return names
}

// IsVersioningPatternSupported checks if a versioning pattern is supported
func IsVersioningPatternSupported(name string) bool {
	_, ok := globalRegistry.versioningPlugins[name]
	return ok
}
