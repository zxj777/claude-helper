# PowerShell installation script for Windows
param(
    [string]$InstallDir = "$env:LOCALAPPDATA\claude-helper"
)

$ErrorActionPreference = "Stop"

# Configuration
$RepoOwner = "zxj777"
$RepoName = "claude-helper"
$BinaryName = "claude-helper.exe"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

Write-Host "Fetching latest release info..."
$LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$RepoOwner/$RepoName/releases/latest"
$Tag = $LatestRelease.tag_name

if (-not $Tag) {
    Write-Error "Failed to get latest release tag"
    exit 1
}

Write-Host "Latest version: $Tag"

# Construct download URL
$BinaryFile = "claude-helper-windows-$Arch.exe"
$DownloadUrl = "https://github.com/$RepoOwner/$RepoName/releases/download/$Tag/$BinaryFile"

Write-Host "Downloading $DownloadUrl..."

# Create install directory
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# Download binary
$OutputPath = Join-Path $InstallDir $BinaryName
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $OutputPath
} catch {
    Write-Error "Failed to download binary: $_"
    exit 1
}

# Add to user PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to user PATH"
}

Write-Host "Installation complete!"
Write-Host "Please restart your terminal and run 'claude-helper --help'"
Write-Host "Or run directly: $OutputPath"