param(
    [string]$PostgresBin = 'C:\Program Files\PostgreSQL\18\bin',
    [string]$MongoBin = '',
    [switch]$Browser
)
$ErrorActionPreference = 'Stop'
$backendRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
foreach ($binary in @('initdb.exe','pg_ctl.exe','createdb.exe','psql.exe')) {
    if (!(Test-Path -LiteralPath (Join-Path $PostgresBin $binary))) { throw "Falta herramienta PostgreSQL: $binary" }
}
foreach ($tool in @('go','node')) { if (!(Get-Command $tool -ErrorAction SilentlyContinue)) { throw "Falta $tool" } }
if ($MongoBin -and !(Test-Path -LiteralPath (Join-Path $MongoBin 'mongod.exe'))) { throw 'Falta mongod.exe' }
if ($Browser -and !$MongoBin) { throw 'El navegador integral requiere tambiÃ©n un MongoDB temporal nuevo.' }
$clusterRoot = Join-Path $backendRoot ('.cache/pg-test-' + [guid]::NewGuid().ToString('N'))
# The directory must be new. Never reuse a service, database, directory or port.
New-Item -ItemType Directory -Path $clusterRoot | Out-Null
$dataPath = Join-Path $clusterRoot 'data'
$logPath = Join-Path $clusterRoot 'postgres.log'
$schemaPath = (Resolve-Path (Join-Path $backendRoot '../database/postgresql/esquema_entrega.sql')).Path
function New-TestPort {
    $listener = [Net.Sockets.TcpListener]::new([Net.IPAddress]::Loopback,0)
    $listener.Start()
    try { return $listener.LocalEndpoint.Port } finally { $listener.Stop() }
}
function Invoke-Pg([string]$binary, [string[]]$arguments) {
    & (Join-Path $PostgresBin $binary) @arguments
    if ($LASTEXITCODE -ne 0) { throw "$binary fallÃ³ con cÃ³digo $LASTEXITCODE" }
}
# Do not inherit opt-ins, alternate DSNs, artifact paths or a connection to MongoDB.
$previous = @{}
foreach ($entry in Get-ChildItem Env:) {
    if ($entry.Name -match '^(INTEGRATION_|FRONTEND_|SHEET_|DOCUMENT_|QUOTATION_|PG)' -or $entry.Name -in @('CARBON_DESIGN_QA','LEGACY_BROWSER_E2E','POSTGRES_DSN','MONGODB_URI','MONGODB_DATABASE','GOFLAGS','GOWORK')) {
        $previous[$entry.Name] = $entry.Value
        [Environment]::SetEnvironmentVariable($entry.Name,$null,'Process')
    }
}
$keys = @('INTEGRATION_DATABASE_URL','INTEGRATION_MONGODB_URI','INTEGRATION_MONGODB_DATABASE','INTEGRATION_QA_TOKEN','POSTGRES_DSN','FRONTEND_AUDIT','FRONTEND_E2E','SHEET_FUNCTIONAL_AUDIT','SHEET_BROWSER_AUDIT','CARBON_DESIGN_QA','LEGACY_BROWSER_E2E','GOFLAGS','GOWORK')
$started = $false; $mongoProcess = $null
Push-Location $backendRoot
try {
    $testPort = New-TestPort
    Invoke-Pg 'initdb.exe' @('-D',$dataPath,'-U','postgres','--encoding=UTF8','--locale=C','--no-sync','-A','trust')
    Invoke-Pg 'pg_ctl.exe' @('-D',$dataPath,'-l',$logPath,'-o',"-p $testPort -h 127.0.0.1",'-w','start')
    $started = $true
    Invoke-Pg 'createdb.exe' @('-h','127.0.0.1','-p',"$testPort",'-U','postgres','elbravodeljac_test')
    Invoke-Pg 'psql.exe' @('-X','-h','127.0.0.1','-p',"$testPort",'-U','postgres','-d','elbravodeljac_test','-q','-v','ON_ERROR_STOP=1','-f',$schemaPath)
    $env:POSTGRES_DSN = "postgres://postgres@127.0.0.1:$testPort/elbravodeljac_test?sslmode=disable"
    $env:INTEGRATION_DATABASE_URL = $env:POSTGRES_DSN
    $env:GOFLAGS = '-mod=readonly'; $env:GOWORK = 'off'
    # Single official migration path: embeds 001â€“009. No second psql migration list.
    & go run ./cmd/db-migrate
    if ($LASTEXITCODE -ne 0) { throw 'FallÃ³ el migrador oficial en el cluster nuevo.' }
    $bytes = New-Object byte[] 32; [Security.Cryptography.RandomNumberGenerator]::Fill($bytes)
    $env:INTEGRATION_QA_TOKEN = [Convert]::ToHexString($bytes).ToLowerInvariant()
    Invoke-Pg 'psql.exe' @('-X','-h','127.0.0.1','-p',"$testPort",'-U','postgres','-d','elbravodeljac_test','-q','-v','ON_ERROR_STOP=1','-c',"ALTER DATABASE elbravodeljac_test SET jac.qa_run_token = '$env:INTEGRATION_QA_TOKEN'")
    $env:POSTGRES_DSN = $null
    if ($MongoBin) {
        $mongoPort = New-TestPort
        $mongoData = Join-Path $clusterRoot 'mongo-data'; New-Item -ItemType Directory -Path $mongoData | Out-Null
        $mongoLog = Join-Path $clusterRoot 'mongod.log'
        $mongoProcess = Start-Process -FilePath (Join-Path $MongoBin 'mongod.exe') -ArgumentList @('--dbpath',('"'+$mongoData+'"'),'--bind_ip','127.0.0.1','--port',"$mongoPort",'--logpath',('"'+$mongoLog+'"'),'--setParameter','enableTestCommands=true') -WindowStyle Hidden -PassThru
        $ready = $false
        for ($attempt=0; $attempt -lt 100; $attempt++) {
            $mongoProcess.Refresh(); if ($mongoProcess.HasExited) { throw 'TerminÃ³ MongoDB temporal.' }
            $tcp = [Net.Sockets.TcpClient]::new()
            try { $tcp.Connect('127.0.0.1',$mongoPort); $ready=$true } catch {} finally { $tcp.Dispose() }
            if ($ready) { break }; Start-Sleep -Milliseconds 200
        }
        if (!$ready) { throw 'MongoDB temporal no abriÃ³ su puerto.' }
        $owner = Get-NetTCPConnection -LocalPort $mongoPort -State Listen -ErrorAction Stop
        if (!($owner | Where-Object { $_.OwningProcess -eq $mongoProcess.Id })) { throw 'El puerto no pertenece al MongoDB reciÃ©n creado; no se usan sus datos.' }
        $env:INTEGRATION_MONGODB_URI = "mongodb://127.0.0.1:$mongoPort/?directConnection=true&retryWrites=false"
        $env:INTEGRATION_MONGODB_DATABASE = 'elbravodeljac_tecnico_test'
    }
    if ($Browser) {
        $env:FRONTEND_AUDIT='1'; $env:FRONTEND_E2E='1'; $env:CARBON_DESIGN_QA='1'
        $env:SHEET_FUNCTIONAL_AUDIT='1'; $env:SHEET_BROWSER_AUDIT='1'; $env:LEGACY_BROWSER_E2E='1'
    }
    & go test ./... -count=1 -v
    if ($LASTEXITCODE -ne 0) { throw 'Pruebas fallidas: revisar resultados, no las bases existentes.' }
} finally {
    try {
        try {
            if ($null -ne $mongoProcess) { $mongoProcess.Refresh(); if (!$mongoProcess.HasExited) { Stop-Process -Id $mongoProcess.Id; $mongoProcess.WaitForExit() } }
        } finally {
            if ($started) { Invoke-Pg 'pg_ctl.exe' @('-D',$dataPath,'-m','fast','-w','stop') }
        }
    } finally {
        foreach ($key in $keys) { [Environment]::SetEnvironmentVariable($key,$null,'Process') }
        foreach ($key in $previous.Keys) { [Environment]::SetEnvironmentVariable($key,$previous[$key],'Process') }
        Pop-Location
    }
}
