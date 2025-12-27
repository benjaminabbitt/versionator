# C Plugin

The c plugin provides versionator support for C projects.

## Overview

C projects can use version information through header files or preprocessor defines at compile time.

## Injection Methods

| Method | Description |
|--------|-------------|
| `emit` | Generate `version.h` header file with `#define` macros |
| `link` | Generate preprocessor flags for compile-time injection |

### Emit

Generate a header file with version macros:

```bash
versionator emit c
# Creates version.h
```

Generated file:

```c
#ifndef VERSION_H
#define VERSION_H

#define VERSION "1.2.3"
#define VERSION_MAJOR 1
#define VERSION_MINOR 2
#define VERSION_PATCH 3
#define GIT_HASH "abc1234"

#endif
```

Usage in your code:

```c
#include "version.h"
#include <stdio.h>

int main() {
    printf("Version: %s\n", VERSION);
    return 0;
}
```

### Link (Preprocessor Defines)

Inject version at compile time without generating files:

```bash
# Get compiler flags
versionator emit build c
# Output: -DVERSION="1.2.3" -DVERSION_MAJOR=1 ...

# Use with gcc/clang
gcc $(versionator emit build c) -o app main.c
```

Usage in your code:

```c
#include <stdio.h>

#ifndef VERSION
#define VERSION "unknown"
#endif

int main() {
    printf("Version: %s\n", VERSION);
    return 0;
}
```

## Build System Integration

### Makefile

```makefile
VERSION_FLAGS := $(shell versionator emit build c)

app: main.c
	$(CC) $(VERSION_FLAGS) -o $@ $<
```

### CMake

```cmake
execute_process(
    COMMAND versionator emit build c
    OUTPUT_VARIABLE VERSION_FLAGS
    OUTPUT_STRIP_TRAILING_WHITESPACE
)
separate_arguments(VERSION_FLAGS)
add_compile_options(${VERSION_FLAGS})
```

## Configuration

Default output path: `version.h`

Override with:

```bash
versionator emit c --output include/version.h
```

## See Also

- [cpp](../cpp/) - C++ projects (uses namespace and constexpr)
