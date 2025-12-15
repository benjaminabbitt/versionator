package main

import "fmt"

// VERSION will be set by the linker during build
var VERSION = "0.0.0"

func main() {
	fmt.Printf("Sample Go Application\n")
	fmt.Printf("Version: %s\n", VERSION)
}
