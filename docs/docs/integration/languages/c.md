---
title: C
description: Embed version in C binaries
sidebar_position: 3
---

# C

**Location:** [`examples/c/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/c)

C uses preprocessor defines (`-D`) to inject values:

```c title="examples/c/main.c"
#include <stdio.h>

// VERSION will be set by the compiler during build
#ifndef VERSION
#define VERSION "0.0.0"
#endif

int main() {
    printf("Sample C Application\n");
    printf("Version: %s\n", VERSION);
    return 0;
}
```

```makefile title="examples/c/Makefile (excerpt)"
build:
    VERSION=$$(versionator version); \
    gcc -DVERSION="\"$$VERSION\"" -o sample-app main.c
```

## Run it

```bash
$ cd examples/c && just run
Getting version from versionator...
Building sample application with version: 0.0.16
Build completed: sample-app
./sample-app
Sample C Application
Version: 0.0.16
```

## Source Code

- [`main.c`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/main.c)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/Containerfile)
