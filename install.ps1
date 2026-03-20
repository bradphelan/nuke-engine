# nuke-build installer for Windows
# Usage: iex (irm https://raw.githubusercontent.com/bradphelan/nuke-engine/main/install.ps1)

$ErrorActionPreference = 'Stop'

$repo    = 'bradphelan/nuke-engine'
$binary  = 'nuke-build'
$installDir = "$env:USERPROFILE\.local\bin"

Write-Host "Installing $binary..."

$release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
$version = $release.tag_name
$asset   = $release.assets | Where-Object { $_.name -like '*windows-amd64*' } | Select-Object -First 1

if (-not $asset) {
    Write-Error "No Windows AMD64 asset found in release $version"
    exit 1
}

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$dest = Join-Path $installDir "$binary.exe"
Write-Host "Downloading $($asset.name) ($version)..."
Invoke-WebRequest $asset.browser_download_url -OutFile $dest

Write-Host "Installed to $dest"

# Add to PATH for this session if not already present
if ($env:PATH -notlike "*$installDir*") {
    $env:PATH = "$installDir;$env:PATH"
    Write-Host ""
    Write-Host "NOTE: Add $installDir to your PATH to use $binary in new terminals:"
    Write-Host '  [System.Environment]::SetEnvironmentVariable("PATH", "$env:USERPROFILE\.local\bin;$env:PATH", "User")'
}
