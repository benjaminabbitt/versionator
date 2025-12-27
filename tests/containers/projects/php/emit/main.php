#!/usr/bin/env php
<?php
// Test application that requires generated version file

require_once 'version.php';

use App\Version;

echo "Version: " . Version::VERSION . "\n";
echo "Major: " . Version::MAJOR . ", Minor: " . Version::MINOR . ", Patch: " . Version::PATCH . "\n";
