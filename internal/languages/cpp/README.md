# C++ Plugin

The cpp plugin provides versionator support for C++ projects.

## Overview

C++ projects can use version information through header files with modern C++ features (namespaces, constexpr) or preprocessor defines at compile time.

## Injection Methods

| Method | Description |
|--------|-------------|
| `emit` | Generate `version.hpp` header file with constexpr constants |
| `link` | Generate preprocessor flags for compile-time injection |

### Emit

Generate a header file with version constants:

```bash
versionator emit cpp
# Creates version.hpp
```

Generated file:

```cpp
#ifndef VERSION_HPP
#define VERSION_HPP

namespace version {
    constexpr const char* VERSION = "1.2.3";
    constexpr int MAJOR = 1;
    constexpr int MINOR = 2;
    constexpr int PATCH = 3;
    constexpr const char* GIT_HASH = "abc1234";
}

#endif
```

Usage in your code:

```cpp
#include "version.hpp"
#include <iostream>

int main() {
    std::cout << "Version: " << version::VERSION << std::endl;
    return 0;
}
```

### Link (Preprocessor Defines)

Inject version at compile time without generating files:

```bash
# Get compiler flags
versionator emit build cpp
# Output: -DVERSION="1.2.3" -DVERSION_MAJOR=1 ...

# Use with g++/clang++
g++ $(versionator emit build cpp) -o app main.cpp
```

Usage in your code:

```cpp
#include <iostream>

#ifndef VERSION
#define VERSION "unknown"
#endif

int main() {
    std::cout << "Version: " << VERSION << std::endl;
    return 0;
}
```

## Build System Integration

### Makefile

```makefile
VERSION_FLAGS := $(shell versionator emit build cpp)

app: main.cpp
	$(CXX) $(VERSION_FLAGS) -o $@ $<
```

### CMake

```cmake
execute_process(
    COMMAND versionator emit build cpp
    OUTPUT_VARIABLE VERSION_FLAGS
    OUTPUT_STRIP_TRAILING_WHITESPACE
)
separate_arguments(VERSION_FLAGS)
add_compile_options(${VERSION_FLAGS})
```

## Configuration

Default output path: `version.hpp`

Override with:

```bash
versionator emit cpp --output include/version.hpp
```

## See Also

- [c](../c/) - C projects (uses #define macros)
