param(
    [string]$Version = "",
    [string]$TargetOS = "linux,windows",
    [string]$Arch = "amd64"
)

$ErrorActionPreference = "Stop"
$repoRoot = Split-Path -Parent $PSScriptRoot
$publicDir = Join-Path $repoRoot "web/public"
$publicStaticDir = Join-Path $publicDir "static"
$publicIndex = Join-Path $publicDir "index.html"
$vueDir = Join-Path $repoRoot "web/vue"

Push-Location $repoRoot
try {
    Push-Location $vueDir
    try {
        npm run build
    } finally {
        Pop-Location
    }

    if (Test-Path -LiteralPath $publicStaticDir) {
        Remove-Item -LiteralPath $publicStaticDir -Recurse -Force
    }
    if (Test-Path -LiteralPath $publicIndex) {
        Remove-Item -LiteralPath $publicIndex -Force
    }
    New-Item -ItemType Directory -Path $publicDir -Force | Out-Null
    Copy-Item -Path (Join-Path $vueDir "dist/*") -Destination $publicDir -Recurse -Force

    go run github.com/rakyll/statik -src=web/public -dest=internal -f
    go test ./...

    $arguments = @("run", "./tools/package-release", "-os", $TargetOS, "-arch", $Arch)
    if ($Version) {
        $arguments += @("-version", $Version)
    }
    & go @arguments
    if ($LASTEXITCODE -ne 0) {
        throw "Release packaging failed with exit code $LASTEXITCODE"
    }
} finally {
    Pop-Location
}
