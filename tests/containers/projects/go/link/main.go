package main

import "fmt"

// Version is injected at build time via -ldflags
var Version = "dev"

func main() {
	fmt.Printf("Version: %s\n", Version)
}
