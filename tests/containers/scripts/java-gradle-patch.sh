#!/bin/bash
set -e

echo "=== Java Gradle Patch Test ==="
echo "Testing versionator emit patch build.gradle"

echo "Before patch:"
grep "version" build.gradle | head -1

# Patch build.gradle
versionator emit patch build.gradle

echo ""
echo "After patch:"
grep "version" build.gradle | head -1

# Verify the version was patched correctly
echo ""
echo "Verifying patched version..."
if grep -q "version = '1.2.3'" build.gradle; then
    echo "Version: 1.2.3"
else
    echo "ERROR: Version not patched correctly"
    exit 1
fi

# Build with Gradle to verify build.gradle is valid
echo ""
echo "Building with Gradle to verify..."
gradle build -q

echo ""
echo "=== PASS ==="
