---
title: Binary Embedding
description: Embed version information into compiled binaries and container images
sidebar_position: 1
---

# Embedding Version in Binaries

The real power of versionator isn't just tracking version in source control—it's getting that version **into your compiled binary** so you can always identify exactly what's running in production.

## Why Binary Embedding Matters

Consider this scenario: You're debugging a production issue at 2 AM. You SSH into a server and find a binary called `myapp`. What version is it?

Without embedded version info:
```bash
$ ./myapp --version
# ???
$ ls -la myapp
-rwxr-xr-x 1 deploy deploy 12345678 Jan 15 10:30 myapp
# Timestamp? Maybe helpful, maybe not.
```

With embedded version info:
```bash
$ ./myapp --version
myapp v2.1.3 (commit: abc1234, built: 2024-01-15T10:30:00Z)
```

**You instantly know**: the exact version, the exact commit, and when it was built.

## Two Approaches

Every language falls into one of two categories:

| Category | Languages | Mechanism |
|----------|-----------|-----------|
| **Compiled** | Go, Rust, C, C++, Java, Kotlin, C#, Swift | Inject values at compile time |
| **Interpreted** | Python, JavaScript, TypeScript, Ruby, PHP | Generate source file at build time |

Both approaches achieve the same result: version info baked into the final artifact.

---

## Live Demos

All examples below are runnable. Each includes a `justfile` with `just run` to see the embedded version in action.

### Go

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

**Run it:**
```bash
$ cd examples/go && just run
Getting version from versionator...
Building sample application with version: 0.0.13
Build completed: sample-app
./sample-app
Sample Go Application
Version: 0.0.13
```

**Source code:** [`main.go`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/main.go) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/go/justfile)

---

### Rust

**Location:** [`examples/rust/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/rust)

Rust reads environment variables at compile time with `option_env!()`:

```rust title="examples/rust/main.rs"
fn main() {
    // VERSION will be set by the compiler during build via environment variable
    let version = option_env!("VERSION").unwrap_or("0.0.0");

    println!("Sample Rust Application");
    println!("Version: {}", version);
}
```

```makefile title="examples/rust/Makefile (excerpt)"
build:
    VERSION=$$(versionator version); \
    VERSION="$$VERSION" rustc -o sample-app main.rs
```

**Run it:**
```bash
$ cd examples/rust && just run
Getting version from versionator...
Building sample application with version: 0.0.13
Build completed: sample-app
./sample-app
Sample Rust Application
Version: 0.0.13
```

**Source code:** [`main.rs`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/main.rs) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/justfile)

---

### C

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

**Run it:**
```bash
$ cd examples/c && just run
Getting version from versionator...
Building sample application with version: 0.0.13
Build completed: sample-app
./sample-app
Sample C Application
Version: 0.0.13
```

**Source code:** [`main.c`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/main.c) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/c/justfile)

---

### C++

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

**Run it:**
```bash
$ cd examples/cpp && just run
Getting version from versionator...
Building sample application with version: 0.0.13
Build completed: sample-app
./sample-app
Sample C++ Application
Version: 0.0.13
```

**Source code:** [`main.cpp`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/main.cpp) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/cpp/justfile)

---

### Java

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

**Run it:**
```bash
$ cd examples/java && just run
Getting version from versionator...
Generating BuildTime.java from template...
Building sample application with version: 0.0.13
Build completed: app/Main.class app/BuildTime.class
java app.Main
Sample Java Application
Version: 0.0.13
```

**Source code:** [`app/Main.java`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/app/Main.java) | [`app/BuildTime.tmpl.java`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/app/BuildTime.tmpl.java) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/java/justfile)

---

### Python

**Location:** [`examples/python/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/python)

Python uses `versionator emit` to generate a `_version.py` module:

```python title="examples/python/mypackage/main.py"
"""Sample application entry point."""

from . import __version__


def main():
    print("Sample Python Application")
    print(f"Version: {__version__}")


if __name__ == "__main__":
    main()
```

```makefile title="examples/python/Makefile (excerpt)"
version-file:
    versionator emit python --output mypackage/_version.py

run: version-file
    python -m mypackage.main
```

**Run it:**
```bash
$ cd examples/python && just run
Generating _version.py using versionator emit...
Version 0.0.13 written to mypackage/_version.py
python3 -m mypackage.main
Sample Python Application
Version: 0.0.13
```

**Source code:** [`mypackage/main.py`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/mypackage/main.py) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/python/justfile)

---

### JavaScript

**Location:** [`examples/javascript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/javascript)

JavaScript generates a `version.js` module:

```javascript title="examples/javascript/src/index.js"
import { VERSION } from './version.js';

function main() {
    console.log('Sample JavaScript Application');
    console.log(`Version: ${VERSION}`);
}

main();
```

```makefile title="examples/javascript/Makefile (excerpt)"
version-file:
    versionator emit js --output src/version.js

run: version-file
    node src/index.js
```

**Run it:**
```bash
$ cd examples/javascript && just run
Generating version.js using versionator emit...
Version 0.0.13 written to src/version.js
node src/index.js
Sample JavaScript Application
Version: 0.0.13
```

**Source code:** [`src/index.js`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/src/index.js) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/javascript/justfile)

---

### TypeScript

**Location:** [`examples/typescript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/typescript)

TypeScript generates a typed `version.ts` module:

```typescript title="examples/typescript/src/index.ts"
import { VERSION } from './version.js';

function main(): void {
    console.log('Sample TypeScript Application');
    console.log(`Version: ${VERSION}`);
}

main();
```

```makefile title="examples/typescript/Makefile (excerpt)"
version-file:
    versionator emit ts --output src/version.ts

build: version-file
    npx tsc

run: build
    node dist/index.js
```

**Run it:**
```bash
$ cd examples/typescript && just run
Generating version.ts using versionator emit...
Version 0.0.13 written to src/version.ts
Building TypeScript package...
Build completed!
node dist/index.js
Sample TypeScript Application
Version: 0.0.13
```

**Source code:** [`src/index.ts`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/src/index.ts) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/typescript/justfile)

---

### Ruby

**Location:** [`examples/ruby/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/ruby)

Ruby generates a `version.rb` module with a `Versionator` namespace:

```ruby title="examples/ruby/lib/mypackage.rb"
require_relative "mypackage/version"

module Mypackage
  def self.hello
    puts "Sample Ruby Application"
    puts "Version: #{Versionator::VERSION}"
  end
end
```

```makefile title="examples/ruby/Makefile (excerpt)"
version-file:
    versionator emit ruby --output lib/mypackage/version.rb

run: version-file
    ruby -I lib -e "require 'mypackage'; Mypackage.hello"
```

**Run it:**
```bash
$ cd examples/ruby && just run
Generating version.rb using versionator emit...
Version 0.0.13 written to lib/mypackage/version.rb
ruby -I lib -e "require 'mypackage'; Mypackage.hello"
Sample Ruby Application
Version: 0.0.13
```

**Source code:** [`lib/mypackage.rb`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/lib/mypackage.rb) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/ruby/justfile)

---

### PHP

PHP generates a version class:

```bash
versionator emit php --output src/Version.php
```

Generated file:

```php
<?php
namespace MyApp;

class Version {
    public const VERSION = "1.2.3";
    public const MAJOR = 1;
    public const MINOR = 2;
    public const PATCH = 3;
}
```

---

### Kotlin

```bash
versionator emit kotlin --output src/main/kotlin/Version.kt
```

---

### C#

```bash
versionator emit csharp --output src/Version.cs
```

---

### Swift

```bash
versionator emit swift --output Sources/Version.swift
```

---

## JSON / YAML

For configuration files or API responses:

```bash
# JSON
versionator emit json --output version.json

# YAML
versionator emit yaml --output version.yml
```

JSON output:

```json
{
  "version": "1.2.3",
  "major": 1,
  "minor": 2,
  "patch": 3
}
```

---

## Docker / Containers

**Location:** [`examples/docker/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/docker)

Container images embed version info in two places:
1. **The binary inside** (using the language-specific approach above)
2. **OCI image labels** (for image inspection without running)

```dockerfile title="examples/docker/Dockerfile"
# Build arguments
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

FROM golang:1.21-alpine AS builder

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

WORKDIR /app
COPY . .

# Inject version at compile time
RUN go build -ldflags "\
    -X main.Version=${VERSION} \
    -X main.GitCommit=${GIT_COMMIT} \
    -X main.BuildDate=${BUILD_DATE}" \
    -o /app/sample-app

FROM alpine:3.19

ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE

# OCI Image Labels
LABEL org.opencontainers.image.version="${VERSION}"
LABEL org.opencontainers.image.revision="${GIT_COMMIT}"
LABEL org.opencontainers.image.created="${BUILD_DATE}"

COPY --from=builder /app/sample-app /usr/local/bin/sample-app

ENTRYPOINT ["sample-app"]
```

```makefile title="examples/docker/Makefile (excerpt)"
docker-build:
    VERSION=$$(versionator version); \
    COMMIT=$$(versionator version -t "{{ShortHash}}"); \
    DATE=$$(versionator version -t "{{BuildDateTimeUTC}}"); \
    docker build \
        --build-arg VERSION=$$VERSION \
        --build-arg GIT_COMMIT=$$COMMIT \
        --build-arg BUILD_DATE=$$DATE \
        -t sample-app:$$VERSION .
```

**Run it:**
```bash
$ cd examples/docker && just show-version
Version from versionator:
  VERSION=0.0.13
  GIT_COMMIT=ba4ecb3
  BUILD_DATE=2026-03-08T18:52:29Z

$ just docker-build
Building Docker image with:
  VERSION=0.0.13
  GIT_COMMIT=ba4ecb3
  BUILD_DATE=2026-03-08T18:52:29Z
...

$ just docker-run
Running sample-app:0.0.13
Sample Docker Application
Version: 0.0.13 (commit: ba4ecb3, built: 2026-03-08T18:52:29Z)
```

**Source code:** [`main.go`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/main.go) | [`Dockerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/Dockerfile) | [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/docker/justfile)

---

## Running All Demos

From the repository root:

```bash
# Build versionator first
just build

# Run all examples
for dir in examples/*/; do
    echo "=== $dir ==="
    (cd "$dir" && just run 2>/dev/null || echo "skipped")
done
```

---

## The Pattern

Every example follows the same pattern:

1. **Makefile** calls `versionator version` to get the current version
2. **Build step** injects that version (compile-time for compiled languages, file generation for interpreted)
3. **Application** displays the embedded version at runtime

The version is **baked in**. It doesn't read from a file at runtime. It doesn't query an API. It's part of the binary itself.

---

## Template Variables

Use these versionator template variables for richer version info:

| Variable | Use Case | Example |
|----------|----------|---------|
| `{{MajorMinorPatch}}` | Clean version | `1.2.3` |
| `{{Prefix}}{{MajorMinorPatch}}` | Prefixed version | `v1.2.3` |
| `{{ShortHash}}` | Git commit (7 chars) | `abc1234` |
| `{{BuildDateTimeUTC}}` | ISO 8601 timestamp | `2024-01-15T10:30:00Z` |
| `{{BranchName}}` | Current branch | `main` |

See [Template Variables](../templates/variables) for the complete reference.

---

## Custom Templates

The examples above use **built-in templates** via `versionator emit <lang>`. For custom namespaces, additional fields, or different file structures, use custom templates with `--template-file`.

### Dump and Customize

```bash
# Dump Python template
versionator emit dump python > custom_python.tmpl

# Edit custom_python.tmpl...

# Use custom template
versionator emit --template-file custom_python.tmpl --output _version.py
```

### Custom Template Examples

Each interpreted language has a `-custom` example demonstrating the `--template-file` workflow:

| Language | Built-in Template | Custom Template |
|----------|------------------|-----------------|
| Python | [`examples/python/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/python) | [`examples/python-custom/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/python-custom) |
| JavaScript | [`examples/javascript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/javascript) | [`examples/javascript-custom/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/javascript-custom) |
| TypeScript | [`examples/typescript/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/typescript) | [`examples/typescript-custom/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/typescript-custom) |
| Ruby | [`examples/ruby/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/ruby) | [`examples/ruby-custom/`](https://github.com/benjaminabbitt/versionator/tree/master/examples/ruby-custom) |

**Built-in examples** use `versionator emit <lang>` — simple, zero configuration.

**Custom examples** use `versionator emit --template-file` — for custom namespaces (e.g., `Mypackage::VERSION` instead of `Versionator::VERSION`) or additional fields like `GIT_HASH` and `BUILD_DATE`.

---

## Best Practices

1. **Add generated files to .gitignore**: Don't commit version files
2. **Generate at build time**: Run emit in build scripts, not manually
3. **Use appropriate approach**: Inject for compiled, generate for interpreted
4. **Include in CI**: Ensure version files are generated in CI/CD

---

## See Also

- [CI/CD Integration](./cicd) - Automate version injection in pipelines
- [Makefiles and Just](./makefiles) - Build tool integration
- [Template Variables](../templates/variables) - All available template variables
