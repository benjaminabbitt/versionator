package main

import (
	"fmt"
	"os"

	"github.com/benjaminabbitt/versionator/cmd"

	// Import VCS implementations for auto-registration
	// Git VCS also registers as a TemplateProvider plugin
	_ "github.com/benjaminabbitt/versionator/internal/vcs/git"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
