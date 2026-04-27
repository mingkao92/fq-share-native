param(
  [Parameter(Mandatory = $true)]
  [string]$ExtensionId,

  [ValidateSet("chrome", "chromium")]
  [string]$Browser = "chrome",

  [string]$Repo = $(if ($env:FQ_SHARE_GITHUB_REPO) { $env:FQ_SHARE_GITHUB_REPO } else { "mingkao92/fq-share-native" }),

  [string]$Version = "latest"
)

$ErrorActionPreference = "Stop"

function Get-ArchLabel {
  $arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture
  switch ($arch) {
    "X64" { return "amd64" }
    "Arm64" { return "arm64" }
    default { throw "Unsupported architecture: $arch" }
  }
}

$Arch = Get-ArchLabel
$Asset = "free-quick-share-host-windows-$Arch.zip"
if ($Version -eq "latest") {
  $DownloadUrl = "https://github.com/$Repo/releases/latest/download/$Asset"
} else {
  $DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Asset"
}

$TempDir = Join-Path $env:TEMP ("fq-share-install-" + [System.Guid]::NewGuid().ToString("N"))
$ZipPath = Join-Path $TempDir $Asset

try {
  New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

  Write-Host "Downloading: $DownloadUrl"
  Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath
  Expand-Archive -LiteralPath $ZipPath -DestinationPath $TempDir -Force

  $HostBinary = Join-Path $TempDir "free-quick-share-host.exe"
  if (-not (Test-Path $HostBinary)) {
    $Candidate = Get-ChildItem -Path $TempDir -Recurse -File -Filter "free-quick-share-host.exe" | Select-Object -First 1
    if (-not $Candidate) {
      throw "Downloaded archive does not contain free-quick-share-host.exe"
    }
    $HostBinary = $Candidate.FullName
  }

  & $HostBinary install --extension-id $ExtensionId --browser $Browser
  if ($LASTEXITCODE -ne 0) {
    throw "Native host install command failed with exit code $LASTEXITCODE"
  }
} finally {
  if (Test-Path $TempDir) {
    Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
  }
}
