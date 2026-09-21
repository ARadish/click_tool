$ErrorActionPreference = "Stop"

go test -tags win ./...
go install fyne.io/fyne/v2/cmd/fyne@v2.8.1

fyne package `
  -os windows `
  -release `
  -tags win `
  -src ./cmd/mousekeeper `
  -icon ./assets/Icon.png `
  -name MouseKeeper `
  -appID com.local.mousekeeper `
  -appVersion 1.0.0 `
  -appBuild 1

New-Item -ItemType Directory -Force -Path dist | Out-Null
Move-Item -Force MouseKeeper.exe dist/MouseKeeper.exe
Write-Host "Created dist/MouseKeeper.exe"

