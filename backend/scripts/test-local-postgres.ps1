param([string]$PostgresBin = 'C:\Program Files\PostgreSQL\18\bin')
$ErrorActionPreference = 'Stop'
$backendRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$cacheRoot = Join-Path $backendRoot '.cache'
$clusterRoot = Join-Path $cacheRoot ('pg-test-' + [guid]::NewGuid().ToString('N'))
# New directory per run. No deletion, no existing cluster/service is used.
New-Item -ItemType Directory -Path $clusterRoot -Force | Out-Null
$dataPath = Join-Path $clusterRoot 'data'
$logPath = Join-Path $clusterRoot 'postgres.log'
$schemaPath = (Resolve-Path (Join-Path $backendRoot '../database/postgresql/esquema_actual.sql')).Path
$listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 0)
$listener.Start()
$testPort = $listener.LocalEndpoint.Port
$listener.Stop()
function Invoke-Pg([string]$binary, [string[]]$arguments) {
    & (Join-Path $PostgresBin $binary) @arguments
    if ($LASTEXITCODE -ne 0) { throw "$binary falló con código $LASTEXITCODE" }
}
$started = $false
$previousDsn = $env:INTEGRATION_DATABASE_URL
Push-Location $backendRoot
try {
    # No durability benchmark: skip initial fsync only for this disposable cluster.
    Invoke-Pg 'initdb.exe' @('-D', $dataPath, '-U', 'postgres', '--encoding=UTF8', '--locale=C', '--no-sync', '-A', 'trust')
    # Trust is confined to this disposable loopback cluster and random port.
    Invoke-Pg 'pg_ctl.exe' @('-D', $dataPath, '-l', $logPath, '-o', "-p $testPort -h 127.0.0.1", '-w', 'start')
    $started = $true
    Invoke-Pg 'createdb.exe' @('-h','127.0.0.1','-p',"$testPort",'-U','postgres','elbravodeljac_test')
    Invoke-Pg 'psql.exe' @('-h','127.0.0.1','-p',"$testPort",'-U','postgres','-d','elbravodeljac_test','-q','-v','ON_ERROR_STOP=1','-f',$schemaPath)
    $env:INTEGRATION_DATABASE_URL = "postgres://postgres@127.0.0.1:$testPort/elbravodeljac_test?sslmode=disable"
    & go test ./... -count=1 -v
    $testExitCode = $LASTEXITCODE
    if ($testExitCode -ne 0) { throw "Pruebas fallidas ($testExitCode)" }
} finally {
    $env:INTEGRATION_DATABASE_URL = $previousDsn
    if ($started) { Invoke-Pg 'pg_ctl.exe' @('-D',$dataPath,'-m','fast','-w','stop') }
    Pop-Location
}
