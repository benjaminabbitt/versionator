---
title: Java
description: Embed version in Java applications
sidebar_position: 5
---

# Java

**Location:** [`examples/java/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/java)

Java generates a source file from a template at build time:

```java title="examples/java/app/Main.java"
package app;

import static app.BuildTime.VERSION;

public class Main {
    public static void main(String[] args) {
        System.out.println("Sample Java Application");
        System.out.println("Version: " + VERSION);
    }
}
```

The Makefile generates `BuildTime.java` from a template:

```makefile title="examples/java/Makefile (excerpt)"
build:
    VERSION=$$(versionator version); \
    sed -e "s/@VERSION@/$${VERSION}/g" BuildTime.java.tmpl > BuildTime.java; \
    javac Main.java BuildTime.java
```

## Run it

```bash
$ cd examples/java && just run
Getting version from versionator...
Generating BuildTime.java from template...
Building sample application with version: 0.0.16
Build completed: app/Main.class app/BuildTime.class
java app.Main
Sample Java Application
Version: 0.0.16
```

## Source Code

- [`app/Main.java`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/app/Main.java)
- [`app/BuildTime.tmpl.java`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/app/BuildTime.tmpl.java)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/Containerfile)
