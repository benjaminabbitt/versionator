// Package loader provides plugin discovery and loading functionality.
package loader

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/benjaminabbitt/versionator/pkg/plugin"
	goplugin "github.com/hashicorp/go-plugin"
	"go.uber.org/zap"
)

// Loader handles plugin discovery and loading.
type Loader struct {
	logger  *zap.Logger
	clients []*goplugin.Client

	EmitPlugins  map[string]plugin.EmitPluginInterface
	BuildPlugins map[string]plugin.BuildPluginInterface
	PatchPlugins map[string]plugin.PatchPluginInterface
}

// NewLoader creates a new plugin loader.
func NewLoader(logger *zap.Logger) *Loader {
	return &Loader{
		logger:       logger,
		EmitPlugins:  make(map[string]plugin.EmitPluginInterface),
		BuildPlugins: make(map[string]plugin.BuildPluginInterface),
		PatchPlugins: make(map[string]plugin.PatchPluginInterface),
	}
}

// getPluginDirsFunc is the function used to get plugin directories.
// Can be replaced in tests for isolation.
var getPluginDirsFunc = getPluginDirs

// getPluginDirs returns the directories to search for plugins.
func getPluginDirs() []string {
	var dirs []string

	// Environment variable override (highest priority)
	if envDir := os.Getenv("VERSIONATOR_PLUGIN_DIR"); envDir != "" {
		dirs = append(dirs, envDir)
	}

	// User-specific plugin directory
	if configDir, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(configDir, "versionator", "plugins"))
	}

	// Home directory plugins
	if homeDir, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(homeDir, ".versionator", "plugins"))
	}

	// System-wide plugin directory (Unix-like systems)
	if runtime.GOOS != "windows" {
		dirs = append(dirs, "/usr/local/lib/versionator/plugins")
		dirs = append(dirs, "/usr/lib/versionator/plugins")
	}

	return dirs
}

// DiscoverAndLoad discovers and loads all available plugins.
func (l *Loader) DiscoverAndLoad() (int, error) {
	loaded := 0

	for _, dir := range getPluginDirsFunc() {
		n, err := l.loadFromDir(dir)
		if err != nil {
			l.logger.Debug("error loading plugins from directory",
				zap.String("dir", dir),
				zap.Error(err))
			continue
		}
		loaded += n
	}

	return loaded, nil
}

// loadFromDir loads all plugins from a directory.
func (l *Loader) loadFromDir(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, err
	}

	loaded := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		path := filepath.Join(dir, name)

		if !isExecutable(path) {
			continue
		}

		// Determine plugin type from name prefix
		var pluginType string
		switch {
		case strings.HasPrefix(name, plugin.PluginPrefixEmit):
			pluginType = plugin.PluginTypeEmit
		case strings.HasPrefix(name, plugin.PluginPrefixBuild):
			pluginType = plugin.PluginTypeBuild
		case strings.HasPrefix(name, plugin.PluginPrefixPatch):
			pluginType = plugin.PluginTypePatch
		default:
			continue
		}

		if err := l.loadPlugin(path, pluginType); err != nil {
			l.logger.Warn("failed to load plugin",
				zap.String("path", path),
				zap.Error(err))
			continue
		}

		loaded++
	}

	return loaded, nil
}

// loadPlugin loads a single plugin from the given path.
func (l *Loader) loadPlugin(path, pluginType string) error {
	l.logger.Debug("loading plugin", zap.String("path", path), zap.String("type", pluginType))

	var pluginMap map[string]goplugin.Plugin
	switch pluginType {
	case plugin.PluginTypeEmit:
		pluginMap = plugin.EmitPluginMap
	case plugin.PluginTypeBuild:
		pluginMap = plugin.BuildPluginMap
	case plugin.PluginTypePatch:
		pluginMap = plugin.PatchPluginMap
	default:
		return fmt.Errorf("unknown plugin type: %s", pluginType)
	}

	client := goplugin.NewClient(&goplugin.ClientConfig{
		HandshakeConfig:  plugin.Handshake,
		Plugins:          pluginMap,
		Cmd:              exec.Command(path),
		AllowedProtocols: []goplugin.Protocol{goplugin.ProtocolGRPC},
	})

	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to connect: %w", err)
	}

	raw, err := rpcClient.Dispense(pluginType)
	if err != nil {
		client.Kill()
		return fmt.Errorf("failed to dispense: %w", err)
	}

	l.clients = append(l.clients, client)

	switch pluginType {
	case plugin.PluginTypeEmit:
		p, ok := raw.(plugin.EmitPluginInterface)
		if !ok {
			client.Kill()
			return fmt.Errorf("plugin does not implement EmitPluginInterface")
		}
		l.EmitPlugins[p.Format()] = p
		l.logger.Info("loaded emit plugin",
			zap.String("name", p.Name()),
			zap.String("format", p.Format()))

	case plugin.PluginTypeBuild:
		p, ok := raw.(plugin.BuildPluginInterface)
		if !ok {
			client.Kill()
			return fmt.Errorf("plugin does not implement BuildPluginInterface")
		}
		l.BuildPlugins[p.Format()] = p
		l.logger.Info("loaded build plugin",
			zap.String("name", p.Name()),
			zap.String("format", p.Format()))

	case plugin.PluginTypePatch:
		p, ok := raw.(plugin.PatchPluginInterface)
		if !ok {
			client.Kill()
			return fmt.Errorf("plugin does not implement PatchPluginInterface")
		}
		l.PatchPlugins[p.FilePattern()] = p
		l.logger.Info("loaded patch plugin",
			zap.String("name", p.Name()),
			zap.String("pattern", p.FilePattern()))
	}

	return nil
}

// Close terminates all loaded plugins.
func (l *Loader) Close() {
	for _, client := range l.clients {
		client.Kill()
	}
	l.clients = nil
}

// windowsExecutableExtensions contains file extensions that represent
// binary executables on Windows. Only binary formats are included since
// go-plugin requires spawning actual executables, not scripts.
var windowsExecutableExtensions = map[string]struct{}{
	".exe": {}, // Windows PE executable (primary, what Go produces)
	".com": {}, // DOS/Windows command file (legacy binary format)
}

// isExecutable checks if a file is executable.
func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Must be a regular file
	if !info.Mode().IsRegular() {
		return false
	}

	if runtime.GOOS == "windows" {
		// On Windows, check for known binary executable extensions
		ext := strings.ToLower(filepath.Ext(path))
		_, ok := windowsExecutableExtensions[ext]
		return ok
	}

	// On Unix, check executable permission bits
	return info.Mode()&0111 != 0
}
