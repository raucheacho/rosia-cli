# Rosia CLI Installation Script for Windows (PowerShell)
# Usage: iwr -useb https://raw.githubusercontent.com/raucheacho/rosia-cli/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "raucheacho/rosia-cli"
$BinaryName = "rosia.exe"
$InstallDir = "$env:LOCALAPPDATA\rosia\bin"
$ConfigDir = "$env:USERPROFILE\.rosia"
$ProfilesDir = "$ConfigDir\profiles"

# Colors for output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "amd64" }
        "x86" { return "386" }
        default {
            Write-ColorOutput "Error: Unsupported architecture $arch" "Red"
            exit 1
        }
    }
}

# Get latest release version from GitHub
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-ColorOutput "Error: Failed to fetch latest version" "Red"
        Write-ColorOutput $_.Exception.Message "Red"
        exit 1
    }
}

# Download and install binary
function Install-Binary {
    param(
        [string]$Version,
        [string]$Arch
    )
    
    $platform = "windows_$Arch"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/rosia_$($Version.TrimStart('v'))_$platform.zip"
    $tmpDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()
    $zipFile = "$tmpDir\rosia.zip"
    
    Write-ColorOutput "Downloading Rosia CLI $Version for $platform..." "Blue"
    
    try {
        # Create temporary directory
        New-Item -ItemType Directory -Path $tmpDir -Force | Out-Null
        
        # Download archive
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipFile -UseBasicParsing
        
        # Extract archive
        Write-ColorOutput "Extracting archive..." "Blue"
        Expand-Archive -Path $zipFile -DestinationPath $tmpDir -Force
        
        # Create install directory
        if (-not (Test-Path $InstallDir)) {
            New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        }
        
        # Move binary to install directory
        Move-Item -Path "$tmpDir\$BinaryName" -Destination "$InstallDir\$BinaryName" -Force
        
        # Copy profiles if they exist
        if (Test-Path "$tmpDir\profiles") {
            Write-ColorOutput "Installing default profiles..." "Blue"
            if (-not (Test-Path $ProfilesDir)) {
                New-Item -ItemType Directory -Path $ProfilesDir -Force | Out-Null
            }
            Copy-Item -Path "$tmpDir\profiles\*" -Destination $ProfilesDir -Recurse -Force
        }
        
        # Cleanup
        Remove-Item -Path $tmpDir -Recurse -Force
        
        Write-ColorOutput "✓ Rosia CLI installed successfully to $InstallDir\$BinaryName" "Green"
    }
    catch {
        Write-ColorOutput "Error: Installation failed" "Red"
        Write-ColorOutput $_.Exception.Message "Red"
        if (Test-Path $tmpDir) {
            Remove-Item -Path $tmpDir -Recurse -Force
        }
        exit 1
    }
}

# Add to PATH
function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -notlike "*$InstallDir*") {
        Write-ColorOutput "Adding $InstallDir to PATH..." "Blue"
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        $env:Path = "$env:Path;$InstallDir"
        Write-ColorOutput "✓ Added to PATH (restart your terminal for changes to take effect)" "Green"
    }
    else {
        Write-ColorOutput "✓ Install directory already in PATH" "Green"
    }
}

# Setup configuration directory
function Setup-Config {
    if (-not (Test-Path $ConfigDir)) {
        Write-ColorOutput "Creating configuration directory at $ConfigDir..." "Blue"
        New-Item -ItemType Directory -Path $ConfigDir -Force | Out-Null
        New-Item -ItemType Directory -Path "$ConfigDir\trash" -Force | Out-Null
        New-Item -ItemType Directory -Path "$ConfigDir\plugins" -Force | Out-Null
        New-Item -ItemType Directory -Path $ProfilesDir -Force | Out-Null
    }
    
    # Create default config if it doesn't exist
    $configFile = "$env:USERPROFILE\.rosiarc.json"
    if (-not (Test-Path $configFile)) {
        Write-ColorOutput "Creating default configuration file..." "Blue"
        $defaultConfig = @{
            trash_retention_days = 3
            profiles = @("node", "python", "rust", "flutter", "go")
            ignore_paths = @()
            plugins = @()
            concurrency = 0
            telemetry_enabled = $false
        } | ConvertTo-Json -Depth 10
        
        Set-Content -Path $configFile -Value $defaultConfig -Encoding UTF8
        Write-ColorOutput "✓ Default configuration created at $configFile" "Green"
    }
}

# Verify installation
function Test-Installation {
    # Refresh PATH for current session
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    
    try {
        $version = & "$InstallDir\$BinaryName" version 2>&1 | Select-Object -First 1
        Write-ColorOutput "✓ Installation verified: $version" "Green"
        return $true
    }
    catch {
        Write-ColorOutput "Error: Installation verification failed" "Red"
        Write-ColorOutput "Please restart your terminal and try running 'rosia version'" "Yellow"
        return $false
    }
}

# Print usage instructions
function Show-Usage {
    Write-Host ""
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" "Green"
    Write-ColorOutput "  Rosia CLI has been installed successfully!" "Green"
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" "Green"
    Write-Host ""
    Write-ColorOutput "Quick Start:" "Blue"
    Write-Host "  rosia scan C:\projects          # Scan for cleanable files"
    Write-Host "  rosia ui C:\projects            # Launch interactive TUI"
    Write-Host "  rosia clean                     # Clean detected targets"
    Write-Host "  rosia stats                     # View cleaning statistics"
    Write-Host ""
    Write-ColorOutput "Configuration:" "Blue"
    Write-Host "  Config file: $env:USERPROFILE\.rosiarc.json"
    Write-Host "  Profiles:    $ProfilesDir"
    Write-Host "  Trash:       $ConfigDir\trash"
    Write-Host ""
    Write-ColorOutput "Documentation:" "Blue"
    Write-Host "  rosia --help                    # Show help"
    Write-Host "  https://github.com/$Repo"
    Write-Host ""
    Write-ColorOutput "Note: Please restart your terminal for PATH changes to take effect" "Yellow"
    Write-Host ""
}

# Main installation flow
function Main {
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━��━━━━━━━━━━━━━━━━━━━━━━━━━━" "Blue"
    Write-ColorOutput "  Rosia CLI Installer" "Blue"
    Write-ColorOutput "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" "Blue"
    Write-Host ""
    
    # Detect architecture
    $arch = Get-Architecture
    Write-ColorOutput "Detected architecture: $arch" "Blue"
    
    # Get latest version
    $version = Get-LatestVersion
    Write-ColorOutput "Latest version: $version" "Blue"
    Write-Host ""
    
    # Install binary
    Install-Binary -Version $version -Arch $arch
    
    # Add to PATH
    Add-ToPath
    
    # Setup configuration
    Setup-Config
    
    # Verify installation
    Write-Host ""
    Test-Installation | Out-Null
    
    # Print usage instructions
    Show-Usage
}

# Run main function
Main
