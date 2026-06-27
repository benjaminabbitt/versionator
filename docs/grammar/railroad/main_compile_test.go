package main

import "testing"

// Trivial test so module-wide `go test -cover` (gremlins) can build a coverage
// binary for this doc tool; excluded from mutation in .gremlins.yaml.
func TestCompiles(t *testing.T) {}
