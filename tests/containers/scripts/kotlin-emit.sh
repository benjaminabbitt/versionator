#!/bin/bash
set -e

echo "=== Kotlin Emit Test ==="
echo "Testing versionator emit kotlin with Gradle build"

# Generate version file
mkdir -p src/main/kotlin/version
versionator emit kotlin --output src/main/kotlin/version/Version.kt

echo "Generated src/main/kotlin/version/Version.kt:"
cat src/main/kotlin/version/Version.kt

# Build
echo ""
echo "Building with Gradle..."
gradle build -q

# Run
echo ""
echo "Running Kotlin application..."
gradle run -q

echo ""
echo "=== PASS ==="
