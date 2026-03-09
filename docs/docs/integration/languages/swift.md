---
title: Swift
description: Embed version in Swift applications
sidebar_position: 8
---

# Swift

**Location:** [`examples/swift/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/swift)

Swift generates a `Version.swift` file with global constants using `versionator output emit`:

```swift title="examples/swift/main.swift"
print("Sample Swift Application")
print("Version: \(VERSION)")
```

```makefile title="examples/swift/Makefile (excerpt)"
version-file:
    versionator output emit swift --output Version.swift

build: version-file
    swiftc -o sample-app main.swift Version.swift
```

## Run it

```bash
$ cd examples/swift && just run
Generating Version.swift using versionator emit...
Building Swift application...
Build completed: sample-app
./sample-app
Sample Swift Application
Version: 0.0.16
```

## Source Code

- [`main.swift`](https://github.com/benjaminabbitt/versionator/blob/master/examples/swift/main.swift)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/swift/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/swift/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/swift/Containerfile)
