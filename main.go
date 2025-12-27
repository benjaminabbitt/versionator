package main

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/cmd"

	// Import VCS implementations for auto-registration
	// Git VCS also registers as a TemplateProvider plugin
	_ "github.com/benjaminabbitt/versionator/internal/vcs/git"

	// Import language plugins for auto-registration
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/c"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/cpp"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/csharp"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/golang"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/java-gradle"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/java-maven"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/javascript"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/php"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/python"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/python-setuptools"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/ruby"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/rust"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/swift"
	_ "github.com/benjaminabbitt/versionator/pkg/plugin/lang/typescript"

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
