#!/bin/bash
set -e

echo "=== PHP Patch Test ==="
echo "Testing versionator emit patch composer.json"

echo "Before patch:"
cat composer.json

# Patch the composer.json
versionator emit patch

echo ""
echo "After patch:"
cat composer.json

# Run PHP to read version
echo ""
echo "Running PHP application..."
php main.php

echo ""
echo "=== PASS ==="
