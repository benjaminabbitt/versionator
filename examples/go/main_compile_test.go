package main

import "testing"

// Trivial test so module-wide `go test -cover` (gremlins) can build a coverage
// binary for this sample app; this file is excluded from mutation in .gremlins.yaml.
func TestCompiles(t *testing.T) {}
