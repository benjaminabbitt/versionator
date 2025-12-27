#!/bin/bash
set -e

echo "=== Swift Patch Test ==="
echo "Testing versionator emit patch Package.swift"

echo "Before patch:"
cat Package.swift

# Patch Package.swift
versionator emit patch Package.swift

echo ""
echo "After patch:"
cat Package.swift

# Build and run
echo ""
echo "Building Swift application..."
swift build

echo ""
echo "Running Swift application..."
swift run

echo ""
echo "=== PASS ==="
