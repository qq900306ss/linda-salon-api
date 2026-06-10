# Builds the Lambda deployment package (function.zip) on Windows.
# The zip is created by scripts/zip.go so the bootstrap binary keeps the
# unix executable bit (Compress-Archive would drop it).

$ErrorActionPreference = "Stop"

Push-Location (Join-Path $PSScriptRoot "..")
try {
    $env:GOOS = "linux"
    $env:GOARCH = "arm64"
    $env:CGO_ENABLED = "0"

    Write-Host "Building bootstrap (linux/arm64)..."
    go build -tags lambda.norpc -o bootstrap ./cmd/lambda
    if ($LASTEXITCODE -ne 0) { throw "go build failed" }

    # Reset env so the zip helper builds for the host platform.
    Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue

    Write-Host "Packaging function.zip..."
    go run ./scripts/zip.go bootstrap function.zip
    if ($LASTEXITCODE -ne 0) { throw "zip packaging failed" }

    Remove-Item bootstrap -Force
    Write-Host "Done: function.zip"
}
finally {
    Pop-Location
}
