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

	// NOTE: External plugins (emit, build, patch) are loaded from:
	//   - $VERSIONATOR_PLUGIN_DIR
	//   - $XDG_CONFIG_HOME/versionator/plugins
	//   - ~/.versionator/plugins
	//   - /usr/local/lib/versionator/plugins
	//   - /usr/lib/versionator/plugins
	// Build plugins with: just plugins
	// Install plugins with: just plugins-install
)

func main() {
	exitCode := run()
	os.Exit(exitCode)
}

// run executes the CLI and returns an exit code.
// Separating this from main() allows defer to work before os.Exit.
func run() int {
	// Ensure plugin cleanup runs on all exit paths
	defer cmd.Cleanup()

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}
