#!/bin/bash
set -e

echo "=== Rust Emit Test ==="
echo "Testing versionator emit rust"

# Generate version file
versionator emit rust --output src/version.rs

echo "Generated src/version.rs:"
cat src/version.rs

# Build and run
echo ""
echo "Building Rust application..."
cargo build --release

echo ""
echo "Running Rust application..."
./target/release/testapp

echo ""
echo "=== PASS ==="
