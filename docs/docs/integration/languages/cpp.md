---
title: C++
description: Embed version in C++ binaries
sidebar_position: 4
---

# C++

**Location:** [`examples/cpp/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/cpp)

Same approach as C—preprocessor defines:

```cpp title="examples/cpp/main.cpp"
#include <iostream>

// VERSION will be set by the compiler during build
#ifndef VERSION
#define VERSION "0.0.0"
#endif

int main() {
    std::cout << "Sample C++ Application" << std::endl;
    std::cout << "Version: " << VERSION << std::endl;
    return 0;
}
```

```makefile title="examples/cpp/Makefile (excerpt)"
build:
    VERSION=$$(versionator version); \
    g++ -DVERSION="\"$$VERSION\"" -o sample-app main.cpp
```

## Run it

```bash
$ cd examples/cpp && just run
Getting version from versionator...
Building sample application with version: 0.0.16
Build completed: sample-app
./sample-app
Sample C++ Application
Version: 0.0.16
```

## Source Code

- [`main.cpp`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/main.cpp)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/Containerfile)
