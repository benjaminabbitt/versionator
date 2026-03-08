package main

import "fmt"

// Version info will be set by the linker during build
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

func main() {
	fmt.Println("Sample Docker Application")
	fmt.Printf("Version: %s (commit: %s, built: %s)\n", Version, GitCommit, BuildDate)
}
