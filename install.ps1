# Install the licenses CLI on Windows.
#   irm https://raw.githubusercontent.com/forjd/licenses-cli/main/install.ps1 | iex
# Env: LICENSES_INSTALL_DIR  (default $env:LOCALAPPDATA\Programs\licenses)
#      LICENSES_VERSION      (default latest, e.g. v0.1.0)
#      LICENSES_DOWNLOAD_URL (override the download base URL, e.g. a mirror;
#                             checksums.txt is fetched from the same place)

# A script block, so `iex` does not leave variables or preferences in the caller's session.
& {
  $ErrorActionPreference = 'Stop'
  $ProgressPreference = 'SilentlyContinue' # Invoke-WebRequest is much faster without the progress bar
  # Windows PowerShell 5.1 can default to TLS 1.0, which GitHub rejects.
  [Net.ServicePointManager]::SecurityProtocol = [Net.ServicePointManager]::SecurityProtocol -bor [Net.SecurityProtocolType]::Tls12

  $repo = 'forjd/licenses-cli'
  $version = if ($env:LICENSES_VERSION) { $env:LICENSES_VERSION } else { 'latest' }
  if ($version -ne 'latest' -and -not $version.StartsWith('v')) { $version = "v$version" }

  # PROCESSOR_ARCHITEW6432 is set when 32-bit or emulated PowerShell runs on a 64-bit OS.
  $osArch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
  $arch = switch ($osArch) {
    'ARM64' { 'arm64' }
    'AMD64' { 'amd64' }
    default { throw "Unsupported architecture: $osArch" }
  }

  $base = if ($version -eq 'latest') {
    "https://github.com/$repo/releases/latest/download"
  } else {
    "https://github.com/$repo/releases/download/$version"
  }
  if ($env:LICENSES_DOWNLOAD_URL) { $base = $env:LICENSES_DOWNLOAD_URL }
  $archive = "licenses_windows_$arch.zip"
  $dir = if ($env:LICENSES_INSTALL_DIR) { $env:LICENSES_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Programs\licenses' }

  $tmp = Join-Path ([IO.Path]::GetTempPath()) ([Guid]::NewGuid())
  New-Item -ItemType Directory -Path $tmp | Out-Null
  try {
    Write-Host "Downloading $archive..."
    Invoke-WebRequest "$base/$archive" -OutFile (Join-Path $tmp $archive) -UseBasicParsing
    Invoke-WebRequest "$base/checksums.txt" -OutFile (Join-Path $tmp 'checksums.txt') -UseBasicParsing

    $want = Get-Content (Join-Path $tmp 'checksums.txt') |
      ForEach-Object { $f = -split $_; if ($f.Count -eq 2 -and $f[1] -eq $archive) { $f[0] } } |
      Select-Object -First 1
    $got = (Get-FileHash (Join-Path $tmp $archive) -Algorithm SHA256).Hash.ToLower()
    if (-not $want -or $want -ne $got) { throw "Checksum mismatch for $archive" }

    Expand-Archive (Join-Path $tmp $archive) -DestinationPath $tmp -Force
    New-Item -ItemType Directory -Path $dir -Force | Out-Null
    Move-Item (Join-Path $tmp 'licenses.exe') (Join-Path $dir 'licenses.exe') -Force
  } finally {
    Remove-Item $tmp -Recurse -Force -ErrorAction SilentlyContinue
  }

  # Edit the registry value directly so entries such as %USERPROFILE%\bin stay
  # unexpanded and the value keeps its REG_EXPAND_SZ type.
  $key = [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
  try {
    $userPath = $key.GetValue('Path', '', 'DoNotExpandEnvironmentNames')
    $parts = @($userPath -split ';' | Where-Object { $_ })
    if ($parts -notcontains $dir) {
      $key.SetValue('Path', (($parts + $dir) -join ';'), 'ExpandString')
      # Setting any user variable broadcasts the change to Explorer, so new terminals see it.
      [Environment]::SetEnvironmentVariable('LICENSES_INSTALL_REFRESH', '1', 'User')
      [Environment]::SetEnvironmentVariable('LICENSES_INSTALL_REFRESH', $null, 'User')
      $env:Path = "$env:Path;$dir"
      Write-Host "Added $dir to your user PATH (restart your terminal)."
    }
  } finally {
    $key.Close()
  }
  Write-Host "Installed $(& (Join-Path $dir 'licenses.exe') -version) to $dir\licenses.exe"
}
