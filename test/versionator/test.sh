#!/bin/bash
set -e

# Test that versionator is installed and accessible
echo "Testing versionator installation..."

# Check if versionator is in PATH
if ! command -v versionator &> /dev/null; then
    echo "FAIL: versionator command not found"
    exit 1
fi
echo "PASS: versionator is in PATH"

# Check version output works
if ! versionator output version &> /dev/null; then
    # This might fail if no VERSION file exists, try help instead
    if ! versionator --help &> /dev/null; then
        echo "FAIL: versionator does not respond to commands"
        exit 1
    fi
fi
echo "PASS: versionator responds to commands"

# Check init works (creates VERSION file)
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"
git init -q
git config user.email "test@test.com"
git config user.name "Test"

if ! versionator init &> /dev/null; then
    echo "FAIL: versionator init failed"
    rm -rf "$TEMP_DIR"
    exit 1
fi

if [ ! -f VERSION ]; then
    echo "FAIL: VERSION file not created"
    rm -rf "$TEMP_DIR"
    exit 1
fi
echo "PASS: versionator init creates VERSION file"

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "All tests passed!"
