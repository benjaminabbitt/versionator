---
title: Kotlin
description: Embed version in Kotlin applications
sidebar_position: 6
---

# Kotlin

**Location:** [`examples/kotlin/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/kotlin)

Kotlin generates a `Version.kt` object at build time using `versionator emit`:

```kotlin title="examples/kotlin/Main.kt"
package app

import version.Version

fun main() {
    println("Sample Kotlin Application")
    println("Version: ${Version.VERSION}")
}
```

```makefile title="examples/kotlin/Makefile (excerpt)"
version-file:
    versionator emit kotlin --output Version.kt

build: version-file
    kotlinc Main.kt Version.kt -include-runtime -d sample-app.jar
```

## Run it

```bash
$ cd examples/kotlin && just run
Generating Version.kt using versionator emit...
Building Kotlin application...
Build completed: sample-app.jar
java -jar sample-app.jar
Sample Kotlin Application
Version: 0.0.16
```

## Source Code

- [`Main.kt`](https://github.com/benjaminabbitt/versionator/blob/master/examples/kotlin/Main.kt)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/kotlin/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/kotlin/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/kotlin/Containerfile)
