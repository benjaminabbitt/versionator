package main

import (
	"fmt"
	"os"
	"versionator/cmd"
)

// VERSION will be set by the linker during build
var VERSION = "dev"

func main() {
	// Set the application version for the cmd package
	cmd.SetApplicationVersion(VERSION)
	
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
