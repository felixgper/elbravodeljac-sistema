param([switch]$SkipBuild)
$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$privatePath = Join-Path $repoRoot '.env.fase-e.local'
$accessPath = Join-Path $repoRoot 'backend/.cache/acceso-desarrollo.txt'
$projectName = 'elbravodeljac-fase-e-dev'
function New-DevelopmentSecret {
    $randomBytes = New-Object byte[] 32
    [System.Security.Cryptography.RandomNumberGenerator]::Fill($randomBytes)
    return [Convert]::ToHexString($randomBytes).ToLowerInvariant()
}
function Invoke-DevelopmentCompose([string[]]$Arguments) {
    & docker compose --env-file $privatePath -p $projectName -f (Join-Path $repoRoot 'compose.dev.yaml') @Arguments
    if ($LASTEXITCODE -ne 0) { throw 'Falló Docker Compose del entorno aislado. Revisa el error anterior.' }
}
Push-Location $repoRoot
try {
    & docker info --format '{{.ServerVersion}}'
    if ($LASTEXITCODE -ne 0) { throw 'Docker no está accesible desde esta terminal.' }
    $projects = & docker compose ls -a --format json | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo inventariar Compose.' }
    $existingProject = @($projects | Where-Object { $_.Name -eq $projectName })
    if ($existingProject.Count -gt 0 -and $existingProject[0].ConfigFiles -ne (Join-Path $repoRoot 'compose.dev.yaml')) {
        throw 'El nombre reservado pertenece a otra configuración Compose; no se modifica.'
    }
    if (!(Test-Path -LiteralPath $privatePath)) {
        $signingBytes = New-Object byte[] 32
        [System.Security.Cryptography.RandomNumberGenerator]::Fill($signingBytes)
        $settings = @{
            DEV_POSTGRES_PASSWORD = (New-DevelopmentSecret)
            DEV_MONGO_PASSWORD = (New-DevelopmentSecret)
            TEST_POSTGRES_PASSWORD = (New-DevelopmentSecret)
            JWT_SECRET_BASE64 = [Convert]::ToBase64String($signingBytes)
            DEV_ADMIN_USERNAME = 'admin_pruebas'
            DEV_ADMIN_PASSWORD = (New-DevelopmentSecret)
        }
        $lines = foreach ($key in $settings.Keys) { '{0}={1}' -f $key, $settings[$key] }
        [IO.File]::WriteAllLines($privatePath, $lines, [Text.UTF8Encoding]::new($false))
    }
    $settings = @{}
    foreach ($line in [IO.File]::ReadAllLines($privatePath)) {
        if ($line -match '^([A-Z0-9_]+)=(.*)$') { $settings[$Matches[1]] = $Matches[2] }
    }
    foreach ($key in @('DEV_POSTGRES_PASSWORD','DEV_MONGO_PASSWORD','JWT_SECRET_BASE64','DEV_ADMIN_USERNAME','DEV_ADMIN_PASSWORD')) {
        if (!$settings[$key]) { throw "Falta $key en el archivo privado; no se regeneran credenciales existentes." }
    }
    Invoke-DevelopmentCompose @('config','--quiet')
    if (!$SkipBuild) { Invoke-DevelopmentCompose @('build','api','mongo-init','dev-bootstrap') }
    Invoke-DevelopmentCompose @('up','-d','--wait','postgres-dev','mongo-dev')
    Invoke-DevelopmentCompose @('run','--rm','--no-deps','dev-bootstrap')
    Invoke-DevelopmentCompose @('run','--rm','--no-deps','mongo-init')
    Invoke-DevelopmentCompose @('up','-d','api')
    $ready = $false
    for ($attempt = 0; $attempt -lt 120; $attempt++) {
        try { $status = Invoke-RestMethod 'http://localhost:8080/readyz' -TimeoutSec 3; $ready = $status.status -eq 'ready' -and $status.technicalConfigured } catch {}
        if ($ready) { break }
        Start-Sleep -Seconds 1
    }
    if (!$ready) { throw 'La aplicación no confirmó PostgreSQL y MongoDB en /readyz.' }
    New-Item -ItemType Directory -Path (Split-Path $accessPath) -Force | Out-Null
    $accessLines = @('ENTORNO AISLADO DE DESARROLLO - NO COMPARTIR', 'URL: http://localhost:8080/', ('Usuario: '+$settings.DEV_ADMIN_USERNAME), ('Contraseña: '+$settings.DEV_ADMIN_PASSWORD), '', 'Archivo local excluido de Git. La cuenta inicial no se restablece si ya existen usuarios.')
    [IO.File]::WriteAllLines($accessPath, $accessLines, [Text.UTF8Encoding]::new($false))
    Write-Host 'Aplicación REAL disponible: http://localhost:8080/'
    Write-Host "Acceso privado guardado en: $accessPath"
    Write-Host 'PostgreSQL y MongoDB no publican puertos. Otros proyectos Docker permanecen intactos.'
} finally { Pop-Location }
