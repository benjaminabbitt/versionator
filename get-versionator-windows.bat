@echo off
REM Download versionator binary for Windows

echo Creating bin directory...
if not exist bin mkdir bin
echo Downloading versionator for Windows amd64...
curl -L "https://github.com/benjaminabbitt/versionator/releases/latest/download/versionator-windows-amd64.exe" -o "bin\versionator.exe"
echo Successfully downloaded versionator for Windows
echo Binary saved as: bin\versionator.exe