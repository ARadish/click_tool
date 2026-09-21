$ErrorActionPreference = "Stop"
$PSNativeCommandUseErrorActionPreference = $true

$root = Split-Path -Parent $PSScriptRoot
Set-Location $root

go test -tags win ./...
go install fyne.io/fyne/v2/cmd/fyne@v2.8.1

# fyne package chdirs into -src before opening -icon, so the path must be absolute.
$icon = Join-Path $root "assets/Icon.png"
if (-not (Test-Path -LiteralPath $icon)) {
    throw "Missing application icon at $icon"
}

$src = Join-Path $root "cmd/mousekeeper"
fyne package `
  -os windows `
  -release `
  -tags win `
  -src $src `
  -icon $icon `
  -name MouseKeeper `
  -appID com.local.mousekeeper `
  -appVersion 1.0.0 `
  -appBuild 1

$packaged = Join-Path $src "MouseKeeper.exe"
if (-not (Test-Path -LiteralPath $packaged)) {
    throw "fyne package did not produce $packaged"
}

$dist = Join-Path $root "dist"
New-Item -ItemType Directory -Force -Path $dist | Out-Null
Move-Item -Force $packaged (Join-Path $dist "MouseKeeper.exe")
Write-Host "Created dist/MouseKeeper.exe"
