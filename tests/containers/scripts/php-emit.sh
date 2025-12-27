#!/bin/bash
set -e

echo "=== PHP Emit Test ==="
echo "Testing versionator emit php"

# Generate version file
versionator emit php --output version.php

echo "Generated version.php:"
cat version.php

# Run
echo ""
echo "Running PHP application..."
php main.php

echo ""
echo "=== PASS ==="
