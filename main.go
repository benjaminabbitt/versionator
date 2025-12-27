package main

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/cmd"

	// Import VCS implementations for auto-registration
	// Git VCS also registers as a TemplateProvider plugin
	_ "github.com/benjaminabbitt/versionator/internal/vcs/git"

	// Import language plugins for auto-registration
	_ "github.com/benjaminabbitt/versionator/internal/languages/c"
	_ "github.com/benjaminabbitt/versionator/internal/languages/cpp"
	_ "github.com/benjaminabbitt/versionator/internal/languages/csharp"
	_ "github.com/benjaminabbitt/versionator/internal/languages/golang"
	_ "github.com/benjaminabbitt/versionator/internal/languages/java-gradle"
	_ "github.com/benjaminabbitt/versionator/internal/languages/java-maven"
	_ "github.com/benjaminabbitt/versionator/internal/languages/javascript"
	_ "github.com/benjaminabbitt/versionator/internal/languages/php"
	_ "github.com/benjaminabbitt/versionator/internal/languages/python"
	_ "github.com/benjaminabbitt/versionator/internal/languages/python-setuptools"
	_ "github.com/benjaminabbitt/versionator/internal/languages/ruby"
	_ "github.com/benjaminabbitt/versionator/internal/languages/rust"
	_ "github.com/benjaminabbitt/versionator/internal/languages/swift"
	_ "github.com/benjaminabbitt/versionator/internal/languages/typescript"

	// Import versioning pattern plugins for auto-registration
	_ "github.com/benjaminabbitt/versionator/internal/versioning/goversion"
	_ "github.com/benjaminabbitt/versionator/internal/versioning/standard"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
