#!/bin/bash
set -e

echo "=== Java Maven Emit Test ==="
echo "Testing versionator emit java with Maven build"

# Generate version file
mkdir -p src/main/java/version
versionator emit java --output src/main/java/version/Version.java

echo "Generated src/main/java/version/Version.java:"
cat src/main/java/version/Version.java

# Build with Maven
echo ""
echo "Building with Maven..."
mvn package -q -DskipTests

# Run
echo ""
echo "Running Java application..."
java -jar target/testapp-1.0.0.jar

echo ""
echo "=== PASS ==="
