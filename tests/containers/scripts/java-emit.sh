#!/bin/bash
set -e

echo "=== Java Emit Test ==="
echo "Testing versionator emit java"

# Generate version file
mkdir -p version
versionator emit java --output version/Version.java

echo "Generated version/Version.java:"
cat version/Version.java

# Compile
echo ""
echo "Compiling Java application..."
javac -d . version/Version.java Main.java

# Run
echo ""
echo "Running Java application..."
java Main

echo ""
echo "=== PASS ==="
