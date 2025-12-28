package main

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/cmd"

	// Import VCS implementations for auto-registration
	// Git VCS also registers as a TemplateProvider plugin
	_ "github.com/benjaminabbitt/versionator/internal/vcs/git"

	// Import versioning pattern plugins for auto-registration
	_ "github.com/benjaminabbitt/versionator/internal/versioning/goversion"
	_ "github.com/benjaminabbitt/versionator/internal/versioning/standard"

	// NOTE: Language plugins are now loaded as external plugins from:
	//   - $VERSIONATOR_PLUGIN_DIR
	//   - $XDG_CONFIG_HOME/versionator/plugins
	//   - ~/.versionator/plugins
	//   - /usr/local/lib/versionator/plugins
	//   - /usr/lib/versionator/plugins
	// Build plugins with: just plugins
	// Install plugins with: just plugins-install
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
