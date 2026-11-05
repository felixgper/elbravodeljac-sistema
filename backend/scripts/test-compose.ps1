$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$previousPassword = $env:TEST_POSTGRES_PASSWORD
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
# Hex has no reserved URL/Compose characters. No password file is written.
$env:TEST_POSTGRES_PASSWORD = [Convert]::ToHexString($bytes).ToLowerInvariant()
Push-Location $repoRoot
try {
    & docker compose -f compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests
    if ($LASTEXITCODE -ne 0) { throw 'Docker Compose o las pruebas fallaron.' }
} finally {
    & docker compose -f compose.test.yaml down
    $env:TEST_POSTGRES_PASSWORD = $previousPassword
    Pop-Location
}
