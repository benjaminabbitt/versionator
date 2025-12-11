$ErrorActionPreference = 'Stop'

$packageName = $env:ChocolateyPackageName
$toolsDir = "$(Split-Path -Parent $MyInvocation.MyCommand.Definition)"

Write-Host "Uninstalling $packageName..."

# Remove shim from PATH
Uninstall-BinFile -Name 'versionator'

# Remove executable
$exePath = Join-Path $toolsDir 'versionator.exe'
if (Test-Path $exePath) {
    Remove-Item $exePath -Force
    Write-Host "Removed executable: $exePath"
}

Write-Host "$packageName has been uninstalled successfully." -ForegroundColor Green
