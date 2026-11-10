$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$previousPassword = $env:TEST_POSTGRES_PASSWORD
$project = 'elbravodeljac-compose-tests-' + [guid]::NewGuid().ToString('N')
$runRoot = Join-Path $repoRoot "backend/.cache/$project"
New-Item -ItemType Directory -Path $runRoot | Out-Null
$owned = $false
$failed = $false
$bytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
# Hex has no reserved URL/Compose characters. No password file is written.
$env:TEST_POSTGRES_PASSWORD = [Convert]::ToHexString($bytes).ToLowerInvariant()
Push-Location $repoRoot
try {
    $label = "label=com.docker.compose.project=$project"
    $existing = & docker container ls -aq --filter $label
    if ($LASTEXITCODE -ne 0 -or $existing) { throw 'El proyecto no es nuevo; no se reutiliza.' }
    $existing = & docker network ls -q --filter $label
    if ($LASTEXITCODE -ne 0 -or $existing) { throw 'El proyecto QA ya tiene una red; no se reutiliza.' }
    $owned = $true
    & docker compose -p $project -f compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests 2>&1 |
        Tee-Object -FilePath (Join-Path $runRoot 'compose-test.log')
    if ($LASTEXITCODE -ne 0) { throw 'Docker Compose o las pruebas fallaron.' }
} catch {
    $failed = $true
    throw
} finally {
    try {
        # Only a project proven absent before this run can be stopped/removed.
        if ($owned) {
            & docker compose -p $project -f compose.test.yaml down *> (Join-Path $runRoot 'docker-cleanup.log')
            if ($LASTEXITCODE -ne 0) {
                if (!$failed) { throw "No se pudo retirar completamente el QA $project. Consulta docker-cleanup.log." }
                Write-Warning "No se pudo retirar completamente el QA $project. Consulta docker-cleanup.log."
            }
        }
    } finally {
        $env:TEST_POSTGRES_PASSWORD = $previousPassword
        Pop-Location
        Write-Output "Evidencias QA: $runRoot"
    }
}
