$ErrorActionPreference = 'Stop'

# Package parameters
$packageName = $env:ChocolateyPackageName
$version = $env:ChocolateyPackageVersion
$toolsDir = "$(Split-Path -Parent $MyInvocation.MyCommand.Definition)"

# URLs for downloads - 64-bit only (Go static binary)
$url64 = "https://github.com/benjaminabbitt/versionator/releases/download/v$version/versionator-windows-amd64.exe"

# File paths
$exePath = Join-Path $toolsDir 'versionator.exe'

# Download arguments
$packageArgs = @{
    packageName    = $packageName
    fileFullPath   = $exePath
    url64bit       = $url64
    checksum64     = '$checksum64$'
    checksumType64 = 'sha256'
}

# Download the executable
Write-Host "Downloading $packageName v$version..."
Get-ChocolateyWebFile @packageArgs

# Verify the download
if (-not (Test-Path $exePath)) {
    throw "Download failed - executable not found at $exePath"
}

# Create shim for command-line access
Write-Host "Installing $packageName to PATH..."
Install-BinFile -Name 'versionator' -Path $exePath

Write-Host ""
Write-Host "$packageName v$version has been installed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Usage:" -ForegroundColor Cyan
Write-Host "  versionator version         - Show current version"
Write-Host "  versionator patch increment - Increment patch version"
Write-Host "  versionator --help          - Show all commands"
Write-Host ""
