param([switch]$Restart)

$ErrorActionPreference = 'Stop'
$project = 'elbravodeljac-sqlserver-dev'

function Invoke-Docker {
    & docker @args
    if ($LASTEXITCODE -ne 0) { throw "Docker falló al ejecutar $($args[0])." }
}

$containers = @{}
foreach ($service in 'sqlserver', 'mongo', 'api') {
    $name = "$project-$service-1"
    $match = @(Invoke-Docker ps -a --filter "name=^/$name$" --format '{{.Names}}')
    if ($match.Count -ne 1 -or $match[0] -ne $name) {
        throw "Falta $name. Para una instalación nueva consulta database/sqlserver/README.md."
    }
    $identity = Invoke-Docker inspect --format '{{index .Config.Labels "com.docker.compose.project"}}/{{index .Config.Labels "com.docker.compose.service"}}' $name
    if ($identity -ne "$project/$service") { throw "Identidad Compose inesperada para $name." }
    $containers[$service] = Invoke-Docker inspect --format '{{.Id}}' $name
}

$mounts = @{
    sqlserver = @("$($project)_sql-data:/var/opt/mssql")
    mongo = @("$($project)_mongo-data:/data/db", "$($project)_mongo-config:/data/configdb")
}
foreach ($service in $mounts.Keys) {
    $actual = @(Invoke-Docker inspect --format '{{range .Mounts}}{{if eq .Type "volume"}}{{println (printf "%s:%s" .Name .Destination)}}{{end}}{{end}}' $containers[$service])
    foreach ($mount in $mounts[$service]) {
        if ($mount -notin $actual) { throw "Falta el montaje $mount en $service." }
        $volume = ($mount -split ':', 2)[0]
        $owner = Invoke-Docker volume inspect --format '{{index .Labels "com.docker.compose.project"}}' $volume
        if ($owner -ne $project) { throw "El volumen $volume no pertenece a $project." }
    }
}

$keys = @(Invoke-Docker inspect --format '{{range .Config.Env}}{{println (index (split . "=") 0)}}{{end}}' $containers.api)
if ('SQLSERVER_DSN' -notin $keys -or 'MONGODB_URI' -notin $keys -or 'POSTGRES_DSN' -in $keys) {
    throw 'La API existente no está configurada para SQL Server y MongoDB.'
}

if ($Restart) {
    Invoke-Docker stop $containers.api | Out-Null
    Invoke-Docker stop $containers.sqlserver $containers.mongo | Out-Null
}

Invoke-Docker start $containers.sqlserver $containers.mongo | Out-Null
$deadline = (Get-Date).AddSeconds(150)
do {
    $health = @(Invoke-Docker inspect --format '{{.State.Status}}/{{.State.Health.Status}}' $containers.sqlserver $containers.mongo)
    if (($health -join ',') -eq 'running/healthy,running/healthy') { break }
    if ($health -match '^(exited|dead)/' -or (Get-Date) -ge $deadline) { throw 'SQL Server o MongoDB no está saludable.' }
    Start-Sleep -Seconds 2
} while ($true)

Invoke-Docker start $containers.api | Out-Null
$deadline = (Get-Date).AddSeconds(60)
do {
    try {
        $ready = Invoke-RestMethod 'http://localhost:8080/readyz' -TimeoutSec 3
        if ($ready.status -eq 'ready' -and $ready.technicalConfigured) {
            Write-Host 'Aplicación REAL disponible en http://localhost:8080/ con SQL Server y MongoDB.'
            return
        }
    } catch {
        # Esperar a que la API abra el puerto y compruebe los dos motores.
    }
    Start-Sleep -Seconds 2
} while ((Get-Date) -lt $deadline)
throw 'La API no confirmó SQL Server y MongoDB en /readyz.'
