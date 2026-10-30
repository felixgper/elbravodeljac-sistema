param([switch]$SkipContainerSuite,[switch]$StartIsolated)
$ErrorActionPreference='Stop'
$repoRoot=(Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
$privatePath=Join-Path $repoRoot '.env.audit-tests.local'
$testProject='elbravodeljac-catalogo-roles-test'

$testCompose=Join-Path $repoRoot 'compose.test.yaml'
$browserCompose=Join-Path $repoRoot 'compose.catalogo-test.yaml'
$trackedEnv=@('INTEGRATION_DATABASE_URL','INTEGRATION_MONGODB_URI','INTEGRATION_MONGODB_DATABASE','FRONTEND_AUDIT','FRONTEND_E2E','PLAYWRIGHT_MODULE','AUDIT_ARTIFACT_DIR')
$previous=@{};foreach($key in $trackedEnv){$previous[$key]=[Environment]::GetEnvironmentVariable($key,'Process')}
function Assert-OwnedCompose([string]$name,[string[]]$files){
 $projects=& docker compose ls -a --format json | ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'No se pudo inventariar Docker Compose.'}
 $existing=@($projects|Where-Object {$_.Name -eq $name})
 if($existing.Count){$actual=@($existing[0].ConfigFiles -split ','|ForEach-Object {[IO.Path]::GetFullPath($_)}|Sort-Object);$expected=@($files|ForEach-Object {[IO.Path]::GetFullPath($_)}|Sort-Object);if(($actual -join ',') -ne ($expected -join ',')){throw "El proyecto $name pertenece a otra configuración; no se modifica."}}
}
Push-Location $repoRoot
try {
 & docker info --format '{{.ServerVersion}}';if($LASTEXITCODE -ne 0){throw 'Docker Desktop no está disponible.'}
 foreach($tool in @('go','node')){if(!(Get-Command $tool -ErrorAction SilentlyContinue)){throw "Falta $tool en PATH."}}
 if(!(Test-Path -LiteralPath $privatePath)){$randomBytes=New-Object byte[] 32;[Security.Cryptography.RandomNumberGenerator]::Fill($randomBytes);[IO.File]::WriteAllText($privatePath,('TEST_POSTGRES_PASSWORD='+[Convert]::ToHexString($randomBytes).ToLowerInvariant()),[Text.UTF8Encoding]::new($false))}
 $settings=@{};foreach($line in [IO.File]::ReadAllLines($privatePath)){if($line -match '^([A-Z0-9_]+)=(.*)$'){$settings[$Matches[1]]=$Matches[2]}}
 if(!$settings.TEST_POSTGRES_PASSWORD){throw 'Falta la contraseña de pruebas en el archivo privado.'}
 # Reuse an already-running owned test cluster's password rather than replace it.
 # The script never resets databases or volumes. On first use the private file
 # must contain the same test password used to start an existing owned project.
 Assert-OwnedCompose $testProject @($testCompose,$browserCompose)
 if($StartIsolated){
  $existing=& docker compose --env-file $privatePath -p $testProject -f $testCompose -f $browserCompose ps -a -q
  if($existing){throw 'El proyecto ya existe. Reutilízalo sin -StartIsolated; no se recrean servicios ni volúmenes.'}
  & docker compose --env-file $privatePath -p $testProject -f $testCompose -f $browserCompose up -d --wait postgres-test mongo-test
  if($LASTEXITCODE -ne 0){throw 'No se pudo iniciar el nuevo proyecto aislado.'}
 }
 if(!$env:PLAYWRIGHT_MODULE){$env:PLAYWRIGHT_MODULE=Join-Path $env:USERPROFILE '.cache/codex-runtimes/codex-primary-runtime/dependencies/node/node_modules/playwright'}
 if(!(Test-Path -LiteralPath $env:PLAYWRIGHT_MODULE)){throw 'Configura PLAYWRIGHT_MODULE con la ruta de Playwright para ejecutar Chrome.'}
 New-Item -ItemType Directory -Path backend/.cache -Force|Out-Null
 & node checks/run.mjs;if($LASTEXITCODE -ne 0){throw 'Fallaron las suites JavaScript.'}
 # Keep the existing owned cluster: no up/recreate/down/reset.
 foreach($migration in @('007_empleados.sql','008_tipos_documento.sql','009_tres_roles.sql')){
  Get-Content -LiteralPath (Join-Path backend/internal/platform/postgres/migrations $migration) -Raw | & docker compose --env-file $privatePath -p $testProject -f $testCompose -f $browserCompose exec -T postgres-test psql -X -U postgres -d elbravodeljac_test -v ON_ERROR_STOP=1
  if($LASTEXITCODE -ne 0){throw "Falló $migration en la base aislada existente. No se reinician contenedores."}
 }
 $env:INTEGRATION_DATABASE_URL='postgres://postgres:'+$settings.TEST_POSTGRES_PASSWORD+'@127.0.0.1:25433/elbravodeljac_test?sslmode=disable'
 $env:INTEGRATION_MONGODB_URI='mongodb://127.0.0.1:27030/?directConnection=true&retryWrites=false'
 $env:INTEGRATION_MONGODB_DATABASE='elbravodeljac_tecnico_test'
 $env:FRONTEND_AUDIT='1';$env:FRONTEND_E2E='1';$env:AUDIT_ARTIFACT_DIR=Join-Path $repoRoot 'backend/.cache/audit-pdf'
 Push-Location backend
 try {
  & go test ./... -v -count=1 *> .cache/audit-native.log
  if($LASTEXITCODE -ne 0){throw 'Falló la prueba integral aislada: revisa backend/.cache/audit-native.log'}
  & go vet ./...
  if($LASTEXITCODE -ne 0){throw 'Falló go vet ./...'}
 }finally{Pop-Location}
 Write-Host 'JavaScript, Go/PostgreSQL/MongoDB, Chrome y go vet: APROBADOS.'
 Write-Host 'Bases usadas: elbravodeljac_test y elbravodeljac_tecnico_test. Se conservan los registros del entorno de navegador; no se recrearon bases ni volúmenes.'
 Write-Host 'Las bases y la API de desarrollo no se utilizaron para crear datos de prueba.'
}finally{foreach($key in $trackedEnv){[Environment]::SetEnvironmentVariable($key,$previous[$key],'Process')};Pop-Location}
