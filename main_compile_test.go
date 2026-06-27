package main

import "testing"

// Trivial test so module-wide `go test -cover` (gremlins) can build a coverage
// binary for the root command package under the go1.25 managed toolchain.
func TestCompiles(t *testing.T) {}
