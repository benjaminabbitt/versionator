#!/bin/bash
set -e

echo "=== Java Maven Patch Test ==="
echo "Testing versionator emit patch pom.xml"

echo "Before patch:"
cat pom.xml | grep '<version>' | head -1

# Patch the pom.xml
versionator emit patch pom.xml

echo ""
echo "After patch:"
cat pom.xml | grep '<version>' | head -1

# Verify the version was patched correctly
echo ""
echo "Verifying patched version..."
if grep -q '<version>1.2.3</version>' pom.xml; then
    echo "Version: 1.2.3"
else
    echo "ERROR: Version not patched correctly"
    exit 1
fi

# Build with Maven to verify pom.xml is valid
echo ""
echo "Building with Maven to verify..."
mvn package -q -DskipTests

echo ""
echo "=== PASS ==="
