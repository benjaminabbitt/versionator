---
title: Go
description: Embed version in Go binaries
sidebar_position: 1
---

# Go

**Location:** [`examples/go/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/go)

Go's linker injects string values via `-ldflags`:

```go title="examples/go/main.go"
package main

import "fmt"

// VERSION will be set by the linker during build
var VERSION = "0.0.0"

func main() {
    fmt.Printf("Sample Go Application\n")
    fmt.Printf("Version: %s\n", VERSION)
}
```

```makefile title="examples/go/Makefile (excerpt)"
build:
    VERSION=$$(versionator version); \
    go build -ldflags "-X main.VERSION=$$VERSION" -o sample-app .
```

## Run it

```bash
$ cd examples/go && just run
Getting version from versionator...
Building sample application with version: 0.0.16
Build completed: sample-app
./sample-app
Sample Go Application
Version: 0.0.16
```

## Source Code

- [`main.go`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/main.go)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/Containerfile)
