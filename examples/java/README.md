# Java Example

This directory contains a Java example demonstrating how to use versionator with Java applications.

## Important Notice

⚠️ **Warning**: The Java code in this directory is largely untested and may have errors. Use with caution and thoroughly test before using in production environments.

## Contents

- `app/Main.java` - Main application class that displays the version
- `app/BuildTime.java` - Generated class containing build-time version information
- `justfile` - Just build configuration
- `Makefile` - Make build configuration

## Usage

You can build and run the example using either Just or Make:

```bash
# Using Just
just build
just run

# Using Make
make build
make run
```

The build process will inject the current version from versionator into the BuildTime.java file.