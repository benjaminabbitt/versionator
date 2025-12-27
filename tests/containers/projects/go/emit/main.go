package main

import (
	"fmt"

	"testapp/version"
)

func main() {
	fmt.Printf("Version: %s\n", version.Version)
	fmt.Printf("Major: %d, Minor: %d, Patch: %d\n", version.Major, version.Minor, version.Patch)
}
