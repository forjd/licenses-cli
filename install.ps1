# Install the licenses CLI on Windows.
#   irm https://raw.githubusercontent.com/forjd/licenses-cli/main/install.ps1 | iex
# Env: LICENSES_INSTALL_DIR (default $env:LOCALAPPDATA\Programs\licenses)
#      LICENSES_VERSION     (default latest, e.g. v0.1.0)
$ErrorActionPreference = 'Stop'

$repo = 'forjd/licenses-cli'
$version = if ($env:LICENSES_VERSION) { $env:LICENSES_VERSION } else { 'latest' }
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$base = if ($version -eq 'latest') {
  "https://github.com/$repo/releases/latest/download"
} else {
  "https://github.com/$repo/releases/download/$version"
}
$archive = "licenses_windows_$arch.zip"
$dir = if ($env:LICENSES_INSTALL_DIR) { $env:LICENSES_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\licenses' }

$tmp = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid())
New-Item -ItemType Directory -Path $tmp | Out-Null
try {
  Write-Host "Downloading $archive..."
  Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
  Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

  $want = (Get-Content (Join-Path $tmp 'checksums.txt') | Where-Object { $_ -match " $([regex]::Escape($archive))$" }) -split ' ' | Select-Object -First 1
  $got = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash.ToLower()
  if (-not $want -or $want -ne $got) { throw "Checksum mismatch for $archive" }

  Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force
  New-Item -ItemType Directory -Path $dir -Force | Out-Null
  Move-Item (Join-Path $tmp 'licenses.exe') (Join-Path $dir 'licenses.exe') -Force
} finally {
  Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
}

$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if (($userPath -split ';') -notcontains $dir) {
  [Environment]::SetEnvironmentVariable('Path', "$userPath;$dir", 'User')
  $env:Path = "$env:Path;$dir"
  Write-Host "Added $dir to your user PATH (restart your terminal)."
}
Write-Host "Installed $(& (Join-Path $dir 'licenses.exe') -version) to $dir\licenses.exe"
