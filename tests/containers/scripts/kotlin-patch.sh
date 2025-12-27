#!/bin/bash
set -e

echo "=== Kotlin Patch Test ==="
echo "Testing versionator emit patch build.gradle.kts"

echo "Before patch:"
cat build.gradle.kts

# Patch the build.gradle.kts
versionator emit patch

echo ""
echo "After patch:"
cat build.gradle.kts

# Build with Gradle
echo ""
echo "Building with Gradle..."
gradle build -q

# Run
echo ""
echo "Running Kotlin application..."
gradle run -q

echo ""
echo "=== PASS ==="
