# Read development; restore ONLY into a freshly named database in the QA cluster.
param([ValidatePattern('^elbravodeljac-taller-[0-9]+-test$')][string]$TestProject='elbravodeljac-taller-20261102-test',
 [ValidatePattern('^compose\.auditoria(?:-taller|-[0-9]{8})\.yaml$')][string]$TestOverride='compose.auditoria-taller.yaml')
$ErrorActionPreference='Stop'
$repo=(Resolve-Path (Join-Path $PSScriptRoot '../..')).Path
Push-Location $repo
try{
 $projects=& docker compose ls -a --format json | ConvertFrom-Json
 if($LASTEXITCODE -ne 0){throw 'No se puede comprobar la propiedad de los proyectos'}
 foreach($owned in @(@{name='elbravodeljac-fase-e-dev';files=@('compose.dev.yaml')},@{name=$TestProject;files=@('compose.test.yaml',$TestOverride)})){
  $entry=@($projects|Where-Object {$_.Name -eq $owned.name})
  if($entry.Count -ne 1){throw 'Falta un proyecto autorizado; no se crean servicios desde este respaldo'}
  $actual=@($entry[0].ConfigFiles -split ','|ForEach-Object {[IO.Path]::GetFullPath($_)}|Sort-Object)
  $expected=@($owned.files|ForEach-Object {[IO.Path]::GetFullPath((Join-Path $repo $_))}|Sort-Object)
  if(($actual -join ',') -ne ($expected -join ',')){throw 'El proyecto pertenece a otra configuración; no se toca'}
 }
 $stamp=Get-Date -Format 'yyyyMMddHHmmss'
 $folder=Join-Path $repo "backend/.cache/respaldo-auditoria-$stamp"
 New-Item -ItemType Directory -Path $folder -ErrorAction Stop|Out-Null
 $dev=@('compose','--env-file','.env.fase-e.local','-p','elbravodeljac-fase-e-dev','-f','compose.dev.yaml')
 $qa=@('compose','--env-file','.env.audit-tests.local','-p',$TestProject,'-f','compose.test.yaml','-f',$TestOverride)
 $restorePG='jac_restore_'+$stamp;$restoreMongo='elbravodeljac_restore_'+$stamp
 function Read-PG([string[]]$argsCompose,[string]$service,[string]$database,[string]$sql){
  $output=$sql|& docker @argsCompose exec -T $service psql -X -U postgres -d $database -At -v ON_ERROR_STOP=1
  if($LASTEXITCODE -ne 0){throw 'Falló consulta de inventario/restauración'}
  return ($output -join "`n")
 }
 function Inventory-PG([string[]]$argsCompose,[string]$service,[string]$database){
  $tables=(Read-PG $argsCompose $service $database "SELECT tablename FROM pg_tables WHERE schemaname='public' ORDER BY 1") -split "`n"
  $parts=foreach($table in $tables){if($table -notmatch '^[a-z_]+$'){throw 'Tabla no prevista'};"SELECT '$table' AS name,count(*)::int AS rows,md5(coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text)::text,'')) AS fingerprint FROM $table t"}
  $rows=Read-PG $argsCompose $service $database ('SELECT jsonb_agg(x) FROM ('+($parts -join ' UNION ALL ')+') x')|ConvertFrom-Json
  $constraints=Read-PG $argsCompose $service $database "SELECT jsonb_agg(jsonb_build_object('table',c.conrelid::regclass::text,'name',c.conname,'type',c.contype,'definition',pg_get_constraintdef(c.oid)) ORDER BY c.conrelid::regclass::text,c.conname) FROM pg_constraint c JOIN pg_namespace n ON n.oid=c.connamespace WHERE n.nspname='public'"|ConvertFrom-Json
  $indexes=Read-PG $argsCompose $service $database "SELECT jsonb_agg(jsonb_build_object('name',indexname,'definition',indexdef) ORDER BY indexname) FROM pg_indexes WHERE schemaname='public'"|ConvertFrom-Json
  return @{tables=$rows;constraints=$constraints;indexes=$indexes}
 }
 $before=Inventory-PG $dev 'postgres-dev' 'elbravodeljac_dev'
 $before|ConvertTo-Json -Depth 14|Set-Content (Join-Path $folder 'postgres-antes.json') -Encoding utf8
 & docker @dev exec -T postgres-dev pg_dump -U postgres -d elbravodeljac_dev -Fc --no-owner --no-acl -f /tmp/bravo-audit.dump
 if($LASTEXITCODE -ne 0){throw 'Falló pg_dump'}
 & docker @dev cp postgres-dev:/tmp/bravo-audit.dump (Join-Path $folder 'postgres.dump')
 if($LASTEXITCODE -ne 0){throw 'Falló copia de pg_dump'}
 & docker @dev exec -T mongo-dev sh -c 'exec mongodump --username "$MONGO_INITDB_ROOT_USERNAME" --password "$MONGO_INITDB_ROOT_PASSWORD" --authenticationDatabase admin --db elbravodeljac_tecnico_dev --archive=/tmp/bravo-audit.archive.gz --gzip' *> (Join-Path $folder 'mongodump.log')
 if($LASTEXITCODE -ne 0){throw 'Falló mongodump'}
 & docker @dev cp mongo-dev:/tmp/bravo-audit.archive.gz (Join-Path $folder 'mongo.archive.gz')
 if($LASTEXITCODE -ne 0){throw 'Falló copia MongoDB'}
 $null=Read-PG $qa 'postgres-test' 'elbravodeljac_test' ("CREATE DATABASE "+$restorePG)
 & docker @qa cp (Join-Path $folder 'postgres.dump') postgres-test:/tmp/bravo-restore.dump
 if($LASTEXITCODE -ne 0){throw 'Falló copia QA'}
 & docker @qa exec -T postgres-test pg_restore -U postgres -d $restorePG --exit-on-error --no-owner --no-acl /tmp/bravo-restore.dump
 if($LASTEXITCODE -ne 0){throw 'Falló restauración PG; se conserva para inspección'}
 & docker @qa cp (Join-Path $folder 'mongo.archive.gz') mongo-test:/tmp/bravo-restore.archive.gz
 if($LASTEXITCODE -ne 0){throw 'Falló copia Mongo QA'}
 & docker @qa exec -T mongo-test mongorestore --archive=/tmp/bravo-restore.archive.gz --gzip --nsInclude 'elbravodeljac_tecnico_dev.*' --nsFrom 'elbravodeljac_tecnico_dev.*' --nsTo ($restoreMongo+'.*') *> (Join-Path $folder 'mongorestore.log')
 if($LASTEXITCODE -ne 0){throw 'Falló restauración Mongo QA'}
 $after=Inventory-PG $dev 'postgres-dev' 'elbravodeljac_dev';$restored=Inventory-PG $qa 'postgres-test' $restorePG
 if(($before|ConvertTo-Json -Depth 14 -Compress) -ne ($after|ConvertTo-Json -Depth 14 -Compress)){throw 'Desarrollo cambió durante respaldo; repetir con pausa autorizada'}
 # pg_restore may distribute text[] casts onto unbounded varchar literals.
 # Normalize only this equivalent literal-array spelling; compare every other
 # part of every CHECK/FK/PK/UNIQUE definition exactly.
 function Canonical-Constraint([string]$definition){
  $value=[regex]::Replace($definition,"\(('(?:[^']|'')*'::character varying)\)::text",'$1')
  return [regex]::Replace($value,"\(ARRAY\[((?:'(?:[^']|'')*'::character varying(?:, )?)+)\]\)::text\[\]",'ARRAY[$1]')
 }
 foreach($constraint in $before.constraints){$constraint.definition=Canonical-Constraint $constraint.definition}
 foreach($constraint in $restored.constraints){$constraint.definition=Canonical-Constraint $constraint.definition}
 foreach($key in @('tables','constraints','indexes')){if(($before.$key|ConvertTo-Json -Depth 14 -Compress) -ne ($restored.$key|ConvertTo-Json -Depth 14 -Compress)){throw "Restauración PG diferente: $key"}}
 $mongoInventory='const crypto=require("crypto");const d=db.getSiblingDB(DATABASE);print(JSON.stringify(d.getCollectionNames().sort().map(n=>({name:n,rows:d.getCollection(n).countDocuments({}),hash:crypto.createHash("sha256").update(EJSON.stringify(d.getCollection(n).find({}).sort({_id:1}).toArray())).digest("hex"),indexes:d.getCollection(n).getIndexes().sort((a,b)=>a.name.localeCompare(b.name)),options:d.getCollectionInfos({name:n})[0].options}))));'
 $sourceScript='const a=db.getSiblingDB("admin").auth(process.env.MONGO_INITDB_ROOT_USERNAME,process.env.MONGO_INITDB_ROOT_PASSWORD);if(!a.ok)throw Error("auth");'+$mongoInventory.Replace('DATABASE','"elbravodeljac_tecnico_dev"')
 $sourceMongo=& docker @dev exec -T mongo-dev mongosh --quiet --eval $sourceScript
 if($LASTEXITCODE -ne 0){throw 'Falló inventario Mongo origen'}
 $targetMongo=& docker @qa exec -T mongo-test mongosh --quiet --eval $mongoInventory.Replace('DATABASE',('"'+$restoreMongo+'"'))
 if($LASTEXITCODE -ne 0){throw 'Falló inventario Mongo restaurado'}
 if(($sourceMongo|ConvertFrom-Json|ConvertTo-Json -Depth 30 -Compress) -ne ($targetMongo|ConvertFrom-Json|ConvertTo-Json -Depth 30 -Compress)){throw 'Mongo restaurado diferente: documentos, GridFS, índices o validadores'}
 $sourceMongo|Set-Content (Join-Path $folder 'mongo-inventario.json') -Encoding utf8
 $restored|ConvertTo-Json -Depth 14|Set-Content (Join-Path $folder 'postgres-restaurado.json') -Encoding utf8
 Get-FileHash (Join-Path $folder 'postgres.dump'),(Join-Path $folder 'mongo.archive.gz')|Select-Object @{n='file';e={Split-Path $_.Path -Leaf}},Hash|ConvertTo-Json|Set-Content (Join-Path $folder 'SHA256.json') -Encoding utf8
 @{verified=$true;sourcePostgres='elbravodeljac_dev';sourceMongo='elbravodeljac_tecnico_dev';restorePostgres=$restorePG;restoreMongo=$restoreMongo;testProject=$TestProject;folder=$folder;operationalDeletion=$false}|ConvertTo-Json|Set-Content (Join-Path $folder 'resultado.json') -Encoding utf8
 Write-Host "RESPALDOS RESTAURADOS Y VERIFICADOS. Carpeta privada: $folder"
 Write-Host "PostgreSQL restaurado: $restorePG; MongoDB restaurado: $restoreMongo. No se ejecutó limpieza."
}finally{Pop-Location}
