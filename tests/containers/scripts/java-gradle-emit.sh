#!/bin/bash
set -e

echo "=== Java Gradle Emit Test ==="
echo "Testing versionator emit java with Gradle build"

# Generate version file
mkdir -p src/main/java/version
versionator emit java --output src/main/java/version/Version.java

echo "Generated src/main/java/version/Version.java:"
cat src/main/java/version/Version.java

# Build
echo ""
echo "Building with Gradle..."
gradle build -q

# Run
echo ""
echo "Running Java application..."
gradle run -q

echo ""
echo "=== PASS ==="
