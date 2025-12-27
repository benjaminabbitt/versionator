<?php
// Read version from composer.json
$composer = json_decode(file_get_contents('composer.json'), true);
echo "Version: " . $composer['version'] . "\n";
