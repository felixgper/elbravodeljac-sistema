$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$previousPassword = $env:TEST_POSTGRES_PASSWORD
$project = 'elbravodeljac-compose-tests-' + [guid]::NewGuid().ToString('N')
$owned = $false
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
# Hex has no reserved URL/Compose characters. No password file is written.
$env:TEST_POSTGRES_PASSWORD = [Convert]::ToHexString($bytes).ToLowerInvariant()
Push-Location $repoRoot
try {
    $existing = & docker compose -p $project -f compose.test.yaml ps -a -q
    if ($LASTEXITCODE -ne 0 -or $existing) { throw 'El proyecto no es nuevo; no se reutiliza.' }
    $owned = $true
    & docker compose -p $project -f compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests
    if ($LASTEXITCODE -ne 0) { throw 'Docker Compose o las pruebas fallaron.' }
} finally {
    try {
        # Only a project proven absent before this run can be stopped/removed.
        if ($owned) { & docker compose -p $project -f compose.test.yaml down }
    } finally {
        $env:TEST_POSTGRES_PASSWORD = $previousPassword
        Pop-Location
    }
}
