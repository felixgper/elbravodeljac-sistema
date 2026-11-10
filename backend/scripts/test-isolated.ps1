param([switch]$Browser, [string]$TestPattern = '.', [string]$GoBuildCache = '')
$ErrorActionPreference = 'Stop'
$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
foreach ($tool in @('docker','git','go','node')) {
    if (!(Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Falta $tool" }
}
if ($Browser -and !$env:PLAYWRIGHT_MODULE) { throw 'Configura PLAYWRIGHT_MODULE para las suites Chrome.' }
function New-QAPort {
    $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,0)
    $listener.Start()
    try { return $listener.LocalEndpoint.Port } finally { $listener.Stop() }
}
$id = [guid]::NewGuid().ToString('N')
$project = 'elbravodeljac-isolated-' + $id
$runRoot = Join-Path $repoRoot ('backend/.cache/isolated-' + $id)
$workspace = Join-Path $runRoot 'proyecto'
New-Item -ItemType Directory -Path $workspace | Out-Null
# No private/untracked ignored files, no .git, and no existing .cache artifacts.
Push-Location $repoRoot
try {
    $files = & git ls-files --cached --others --exclude-standard
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo inventariar el proyecto.' }
    foreach ($file in $files) {
        $source = Join-Path $repoRoot $file
        if (!(Test-Path -LiteralPath $source -PathType Leaf)) { continue }
        $target = Join-Path $workspace $file
        New-Item -ItemType Directory -Path (Split-Path $target) -Force | Out-Null
        Copy-Item -LiteralPath $source -Destination $target
    }
} finally { Pop-Location }
$pgPort = New-QAPort; $mongoPort = New-QAPort
while ($mongoPort -eq $pgPort) { $mongoPort = New-QAPort }
$override = Join-Path $runRoot 'compose.qa.yaml'
[IO.File]::WriteAllText($override,@"
services:
  postgres-test:
    ports: ["127.0.0.1:${pgPort}:5432"]
  mongo-test:
    ports: ["127.0.0.1:${mongoPort}:27017"]
networks:
  test-only:
    internal: false
"@,[Text.UTF8Encoding]::new($false))
$previous = @{}
$keys = @('TEST_POSTGRES_PASSWORD','GOFLAGS','GOCACHE','GOTMPDIR','GOWORK')
foreach ($entry in Get-ChildItem Env:) {
    if ($entry.Name -match '^(INTEGRATION_|FRONTEND_|SHEET_|DOCUMENT_|QUOTATION_)' -or $entry.Name -in ($keys + @('CARBON_DESIGN_QA','LEGACY_BROWSER_E2E','POSTGRES_DSN','MONGODB_URI','MONGODB_DATABASE'))) {
        $previous[$entry.Name]=$entry.Value
        [Environment]::SetEnvironmentVariable($entry.Name,$null,'Process')
    }
}
$keys += @('INTEGRATION_DATABASE_URL','INTEGRATION_MONGODB_URI','INTEGRATION_MONGODB_DATABASE','INTEGRATION_QA_TOKEN','FRONTEND_AUDIT','FRONTEND_E2E','SHEET_FUNCTIONAL_AUDIT','SHEET_BROWSER_AUDIT','CARBON_DESIGN_QA','LEGACY_BROWSER_E2E')
$compose = @('compose','-p',$project,'-f',(Join-Path $repoRoot 'compose.test.yaml'),'-f',$override)
$report = @{project=$project; workingDirectory=$repoRoot; workspace=$workspace; postgresPort=$pgPort; mongoPort=$mongoPort; tmpfs=$true; credentialsSaved=$false; browser=[bool]$Browser}
Push-Location $workspace
try {
    $bytes=New-Object byte[] 32; [Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:TEST_POSTGRES_PASSWORD=[Convert]::ToHexString($bytes).ToLowerInvariant()
    [Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:INTEGRATION_QA_TOKEN=[Convert]::ToHexString($bytes).ToLowerInvariant()
    $existing = & docker @compose ps -a -q
    if ($LASTEXITCODE -ne 0 -or $existing) { throw 'El proyecto QA no es nuevo; no se reutiliza.' }
    & docker @compose up -d --wait postgres-test mongo-test 1> (Join-Path $runRoot 'docker.log') 2> (Join-Path $runRoot 'docker-stderr.log')
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo crear el QA nuevo.' }
    # Only the NEW test database is stamped. A mere JS opt-in cannot attest the API.
    "ALTER DATABASE elbravodeljac_test SET jac.qa_run_token = '$env:INTEGRATION_QA_TOKEN';" | & docker @compose exec -T postgres-test psql -X -U postgres -d elbravodeljac_test -v ON_ERROR_STOP=1 -q
    if ($LASTEXITCODE -ne 0) { throw 'No se pudo acreditar la base nueva.' }
    $env:INTEGRATION_DATABASE_URL="postgres://postgres:$env:TEST_POSTGRES_PASSWORD@127.0.0.1:$pgPort/elbravodeljac_test?sslmode=disable"
    $env:INTEGRATION_MONGODB_URI="mongodb://127.0.0.1:$mongoPort/?directConnection=true&retryWrites=false"
    $env:INTEGRATION_MONGODB_DATABASE='elbravodeljac_tecnico_test'
    $env:GOCACHE=if ($GoBuildCache) { [IO.Path]::GetFullPath($GoBuildCache) } else { Join-Path $runRoot 'go-build' }
    $env:GOTMPDIR=Join-Path $runRoot 'go-tmp'
    New-Item -ItemType Directory -Path $env:GOTMPDIR | Out-Null
    $env:GOFLAGS='-mod=readonly'; $env:GOWORK='off'
    if ($Browser) {
        $env:FRONTEND_AUDIT='1'; $env:FRONTEND_E2E='1'; $env:CARBON_DESIGN_QA='1'
        $env:SHEET_FUNCTIONAL_AUDIT='1'; $env:SHEET_BROWSER_AUDIT='1'; $env:LEGACY_BROWSER_E2E='1'
    }
    & node checks/run.mjs 1> (Join-Path $runRoot 'javascript.log') 2> (Join-Path $runRoot 'javascript-stderr.log')
    $report.javascriptExitCode=$LASTEXITCODE
    Push-Location backend
    try {
        $env:POSTGRES_DSN = $env:INTEGRATION_DATABASE_URL
        try {
            & go run ./cmd/db-migrate
            if ($LASTEXITCODE -ne 0) { throw 'No se pudo preparar el catálogo documental del QA nuevo.' }
        } finally { Remove-Item Env:POSTGRES_DSN -ErrorAction SilentlyContinue }
        & go test ./... -run $TestPattern -json -count=1 -timeout=20m 1> (Join-Path $runRoot 'go-test.log') 2> (Join-Path $runRoot 'go-test-stderr.log')
        $report.goTestExitCode=$LASTEXITCODE
        & go vet ./... 1> (Join-Path $runRoot 'go-vet.log') 2> (Join-Path $runRoot 'go-vet-stderr.log')
        $report.goVetExitCode=$LASTEXITCODE
    } finally { Pop-Location }
    $report | ConvertTo-Json | Set-Content -LiteralPath (Join-Path $runRoot 'resultado.json') -Encoding utf8
    Write-Output "Evidencias QA: $runRoot"
    Write-Output "Proyecto: $project (solo recursos nuevos; no se detuvo ningún servicio previo)"
    if ($report.javascriptExitCode -ne 0 -or $report.goTestExitCode -ne 0 -or $report.goVetExitCode -ne 0) { throw 'Hay pruebas fallidas. Revisar los logs ignorados; no se modifican bases existentes para corregirlas.' }
} finally {
    foreach ($key in $keys) { [Environment]::SetEnvironmentVariable($key,$null,'Process') }
    foreach ($key in $previous.Keys) { [Environment]::SetEnvironmentVariable($key,$previous[$key],'Process') }
    Pop-Location
}
