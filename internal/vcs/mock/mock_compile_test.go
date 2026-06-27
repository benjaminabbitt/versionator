package mock

import "testing"

// Trivial test so `go test -cover` can build a coverage binary for this
// otherwise test-less package; without it, module-wide coverage (gremlins) fails
// with "go: no such tool covdata" under the go1.25 managed toolchain.
func TestPackageCompiles(t *testing.T) {}
