# SSHM Windows Installation Script
# Usage: irm https://raw.githubusercontent.com/Gu1llaum-3/sshm/main/install/windows.ps1 | iex

param(
    [string]$InstallDir = "$env:USERPROFILE\bin",
    [switch]$Force = $false
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput($ForegroundColor) {
    $fc = $host.UI.RawUI.ForegroundColor
    $host.UI.RawUI.ForegroundColor = $ForegroundColor
    if ($args) {
        Write-Output $args
    }
    $host.UI.RawUI.ForegroundColor = $fc
}

function Write-Info { Write-ColorOutput Green $args }
function Write-Warning { Write-ColorOutput Yellow $args }
function Write-Error { Write-ColorOutput Red $args }

Write-Info "🚀 Installing SSHM - SSH Manager"
Write-Info ""

# Check if SSHM is already installed
$existingSSHM = Get-Command sshm -ErrorAction SilentlyContinue
if ($existingSSHM -and -not $Force) {
    $currentVersion = & sshm --version 2>$null | Select-String "version" | ForEach-Object { $_.ToString().Split()[-1] }
    Write-Warning "SSHM is already installed (version: $currentVersion)"
    $response = Read-Host "Do you want to continue with the installation? (y/N)"
    if ($response -ne "y" -and $response -ne "Y") {
        Write-Info "Installation cancelled."
        exit 0
    }
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
Write-Info "Detected platform: Windows ($arch)"

# Get latest version
Write-Info "Fetching latest version..."
try {
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/Gu1llaum-3/sshm/releases/latest"
    $latestVersion = $latestRelease.tag_name
    Write-Info "Latest version: $latestVersion"
} catch {
    Write-Error "Failed to fetch latest version"
    exit 1
}

# Download binary
$fileName = "sshm-windows-$arch.zip"
$downloadUrl = "https://github.com/Gu1llaum-3/sshm/releases/download/$latestVersion/$fileName"
$tempFile = "$env:TEMP\$fileName"

Write-Info "Downloading $fileName..."
try {
    Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
} catch {
    Write-Error "Download failed"
    exit 1
}

# Create installation directory
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

# Extract archive
Write-Info "Extracting..."
try {
    Expand-Archive -Path $tempFile -DestinationPath $env:TEMP -Force
    $extractedBinary = "$env:TEMP\sshm-windows-$arch.exe"
    $targetPath = "$InstallDir\sshm.exe"
    
    Move-Item -Path $extractedBinary -Destination $targetPath -Force
} catch {
    Write-Error "Extraction failed"
    exit 1
}

# Clean up
Remove-Item $tempFile -Force

# Check PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$InstallDir*") {
    Write-Warning "The directory $InstallDir is not in your PATH."
    Write-Info "Adding to user PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$InstallDir", "User")
    Write-Info "Please restart your terminal to use the 'sshm' command."
}

Write-Info ""
Write-Info "✅ SSHM successfully installed to: $targetPath"
Write-Info "You can now use the 'sshm' command!"

# Verify installation
if (Test-Path $targetPath) {
    Write-Info ""
    Write-Info "Verifying installation..."
    & $targetPath --version
}
