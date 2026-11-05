param([string]$MongoBin = 'C:\Program Files\MongoDB\Server\9.0\bin', [string]$PostgresBin = 'C:\Program Files\PostgreSQL\18\bin')
$ErrorActionPreference = 'Stop'
$backendRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
$root = Join-Path $backendRoot ('.cache/mongo-test-' + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $root -Force | Out-Null
$data = Join-Path $root 'data'
New-Item -ItemType Directory -Path $data | Out-Null
$log = Join-Path $root 'mongod.log'
$listener = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback,0)
$listener.Start(); $port = $listener.LocalEndpoint.Port; $listener.Stop()
$oldUri = $env:INTEGRATION_MONGODB_URI
$oldDatabase = $env:INTEGRATION_MONGODB_DATABASE
$process = $null
try {
    # Native server is only this fresh loopback directory/random port. Never use a service.
    $process = Start-Process -FilePath (Join-Path $MongoBin 'mongod.exe') -ArgumentList @('--dbpath',('"'+$data+'"'),'--bind_ip','127.0.0.1','--port',"$port",'--logpath',('"'+$log+'"'),'--setParameter','enableTestCommands=true') -WindowStyle Hidden -PassThru
    $ready = $false
    for ($attempt=0; $attempt -lt 100; $attempt++) {
        $process.Refresh()
        if ($process.HasExited) { throw 'La instancia temporal MongoDB terminó antes de aceptar conexiones.' }
        $tcp = [System.Net.Sockets.TcpClient]::new()
        try { $tcp.Connect('127.0.0.1',$port); $ready=$true } catch {} finally { $tcp.Dispose() }
        if ($ready) { break }; Start-Sleep -Milliseconds 200
    }
    if (!$ready) { throw 'MongoDB temporal no abrió su puerto local.' }
    $env:INTEGRATION_MONGODB_URI = "mongodb://127.0.0.1:$port/?directConnection=true&retryWrites=false"
    $env:INTEGRATION_MONGODB_DATABASE = 'elbravodeljac_tecnico_test'
    & (Join-Path $PSScriptRoot 'test-local-postgres.ps1') -PostgresBin $PostgresBin
} finally {
    $env:INTEGRATION_MONGODB_URI = $oldUri
    $env:INTEGRATION_MONGODB_DATABASE = $oldDatabase
    if ($null -ne $process) { $process.Refresh(); if (!$process.HasExited) { Stop-Process -Id $process.Id; $process.WaitForExit() } }
}
