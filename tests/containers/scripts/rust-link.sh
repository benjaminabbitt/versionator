#!/bin/bash
set -e

echo "=== Rust Link Test ==="
echo "Testing versionator emit build rust (env var injection)"

# Get build flags (env var format)
BUILD_FLAGS=$(versionator emit build rust --var VERSION)
echo "Using env var: $BUILD_FLAGS"

# Build with version injection via environment variable
echo ""
echo "Building Rust application with version injection..."
export $BUILD_FLAGS
cargo build --release

# Run
echo ""
echo "Running Rust application..."
./target/release/testapp

echo ""
echo "=== PASS ==="
