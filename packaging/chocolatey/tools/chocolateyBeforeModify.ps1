$ErrorActionPreference = 'Stop'

# This script runs before install, upgrade, or uninstall
# Used to stop any running processes or services before modification

$packageName = $env:ChocolateyPackageName

Write-Host "Preparing to modify $packageName..."

# Stop any running versionator processes (unlikely for CLI tool, but good practice)
$processes = Get-Process -Name 'versionator' -ErrorAction SilentlyContinue
if ($processes) {
    Write-Host "Stopping running versionator processes..."
    $processes | Stop-Process -Force
    Start-Sleep -Seconds 1
}
