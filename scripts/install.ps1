# Pecel Windows Installation Script
param(
    [string]$Action = "install",
    [string]$Version = "v1.0.0"
)

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "yourusername/pecel"
$BinaryName = "pecel.exe"
$InstallDir = "$env:USERPROFILE\bin"
$TempDir = "$env:TEMP\pecel-install"

# Colors
$Green = "`e[32m"
$Yellow = "`e[33m"
$Red = "`e[31m"
$Reset = "`e[0m"

function Write-Info {
    Write-Host "${Green}[INFO]${Reset} $($args[0])"
}

function Write-Warn {
    Write-Host "${Yellow}[WARN]${Reset} $($args[0])"
}

function Write-Error {
    Write-Host "${Red}[ERROR]${Reset} $($args[0])"
    exit 1
}

# Determine platform
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { "amd64" }
}

$Platform = "windows"

function Install-Binary {
    Write-Info "Downloading pecel $Version for $Platform/$Arch..."
    
    $DownloadUrl = "https://github.com/$Repo/releases/download/$Version/${BinaryName}-${Platform}-${Arch}.exe"
    
    # Create temp directory
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $TempDir -Force | Out-Null
    
    # Download binary
    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile "$TempDir\$BinaryName"
    } catch {
        Write-Error "Failed to download binary: $_"
    }
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    # Install binary
    Copy-Item "$TempDir\$BinaryName" "$InstallDir\$BinaryName" -Force
    
    # Add to PATH if not already present
    $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($CurrentPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to user PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$CurrentPath;$InstallDir", "User")
        $env:Path += ";$InstallDir"
    }
    
    # Cleanup
    Remove-Item -Path $TempDir -Recurse -Force
    
    Write-Info "Installation completed!"
    Write-Info "Run 'pecel --help' to get started"
}

function Uninstall {
    $BinaryPath = "$InstallDir\$BinaryName"
    if (Test-Path $BinaryPath) {
        Write-Info "Removing $BinaryPath..."
        Remove-Item -Path $BinaryPath -Force
        
        # Remove from PATH
        $CurrentPath = [Environment]::GetEnvironmentVariable("Path", "User")
        $NewPath = $CurrentPath -replace [regex]::Escape($InstallDir), "" -replace ";;", ";"
        [Environment]::SetEnvironmentVariable("Path", $NewPath.TrimEnd(';'), "User")
        
        Write-Info "Pecel has been uninstalled"
    } else {
        Write-Warn "Pecel is not installed"
    }
}

function Update {
    Write-Info "Checking for updates..."
    
    try {
        $LatestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $LatestVersion = $LatestRelease.tag_name
        
        if ($LatestVersion -ne $Version) {
            Write-Info "New version available: $LatestVersion"
            $Version = $LatestVersion
            Install-Binary
        } else {
            Write-Info "You have the latest version ($Version)"
        }
    } catch {
        Write-Error "Failed to check for updates: $_"
    }
}

# Main execution
switch ($Action.ToLower()) {
    "install" {
        Install-Binary
    }
    "update" {
        Update
    }
    "uninstall" {
        Uninstall
    }
    default {
        Write-Error "Unknown action: $Action"
        Write-Host "Usage: .\install.ps1 [install|update|uninstall]"
        exit 1
    }
}