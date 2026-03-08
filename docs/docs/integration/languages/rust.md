---
title: Rust
description: Embed version in Rust binaries
sidebar_position: 2
---

# Rust

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

## Run it

```bash
$ cd examples/rust && just run
Getting version from versionator...
Building sample application with version: 0.0.16
Build completed: sample-app
./sample-app
Sample Rust Application
Version: 0.0.16
```

## Source Code

- [`main.rs`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/main.rs)
- [`justfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/justfile)
- [`Makefile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/Makefile)
- [`Containerfile`](https://github.com/benjaminabbitt/versionator/blob/master/examples/rust/Containerfile)
