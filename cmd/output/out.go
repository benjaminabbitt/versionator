package output

import (
	"github.com/benjaminabbitt/versionator/pkg/plugin/loader"
)

// PluginLoader is set by the parent cmd package during initialization
var PluginLoader *loader.Loader

// UseDefaultMarker is the marker for "flag provided without value"
const UseDefaultMarker = "\x00DEFAULT\x00"
