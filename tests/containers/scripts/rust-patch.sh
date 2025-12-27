#!/bin/bash
set -e

echo "=== Rust Patch Test ==="
echo "Testing versionator emit patch Cargo.toml"

echo "Before patch:"
cat Cargo.toml

# Patch the Cargo.toml
versionator emit patch

echo ""
echo "After patch:"
cat Cargo.toml

# Build and run
echo ""
echo "Building Rust application..."
cargo build --release

echo ""
echo "Running Rust application..."
./target/release/testapp

echo ""
echo "=== PASS ==="
