# Esquema físico de El Bravo del JAC en SQL Server

Para preparar una copia histórica desde PostgreSQL, consulta el [procedimiento de migración de datos](MIGRACION_DATOS.md).

SQL Server 2025 es el motor relacional oficial. El esquema prepara una instalación vacía en Standard o Standard Developer. La validación automatizada del adaptador Go utiliza Standard Developer.

Ejecutar desde un cliente compatible con lotes `GO` y detenerse ante cualquier error. Para probar, crear previamente una base desechable llamada `elbravodeljac_sqlserver_qa` en una instancia aislada con nivel de compatibilidad 170. Con una sesión de cliente configurada para esa base:

```powershell
sqlcmd -S <instancia-qa> -d elbravodeljac_sqlserver_qa -b -f 65001 -i database/sqlserver/esquema_entrega.sql
sqlcmd -S <instancia-qa> -d elbravodeljac_sqlserver_qa -b -f 65001 -i database/sqlserver/pruebas_esquema.sql
```

El primer script crea el esquema en una transacción y no instala datos iniciales. El segundo carga datos sintéticos en la base desechable y verifica las reglas por lotes. Una instrucción con una fila inválida debe revertirse completa. Los casos negativos comprueban el error concreto y, cuando corresponde, que tampoco se escribió la fila válida.

- Se mantienen las 28 tablas, 208 columnas, 28 PK, 50 FK, 22 CHECK de origen, 26 secuencias, 3 vistas y 93 índices contando los de las claves. Las nueve funciones de trigger de PostgreSQL se integran en diez triggers T-SQL. Las otras tres funciones conservan sus nombres.
- Los importes mantienen precisión y escala. Las fechas usan `datetimeoffset(6)` con valor por defecto UTC. SQL Server genera ese valor al escribir, mientras que `CURRENT_TIMESTAMP` de PostgreSQL se fija al comenzar la transacción. Cuando varias filas deban compartir un mismo instante de negocio, el adaptador pasará el mismo valor explícito.
- Los textos de longitud fija de contrato usan `nvarchar` con capacidad para pares sustitutos y un CHECK de longitud en caracteres. El JSON se conserva como texto Unicode validado y las fichas archivadas como `varbinary(max)`. El adaptador debe usar una representación JSON estable para no confundir cambios de formato con cambios en datos históricos.
- La colación binaria conserva distinciones de mayúsculas y acentos en claves. SQL Server aún puede equiparar cadenas con espacios finales en índices únicos, así que la carga debe detectar esas colisiones y detenerse para revisión, sin descartar ni modificar filas automáticamente. Para búsquedas se usa una colación insensible a acentos. El índice de clientes cubre nombre y documento, pero no reemplaza el rendimiento del índice GIN para búsquedas por subcadena. El puerto del adaptador debe verificar ambas cosas con datos representativos.
- Las FK `RESTRICT` se expresan como `NO ACTION`. Correo y vínculo con empleado tienen índices únicos filtrados, con varios NULL permitidos. Las normalizaciones de escritura usan triggers `INSTEAD OF`; el futuro adaptador proporcionará los campos obligatorios, incluido el identificador de recepción al cotizar, antes de enviar el DML. Los triggers que correlacionan filas por PK rechazan actualizaciones de ese identificador.
- El adaptador Go protege el último acceso administrativo, controla versiones y toma un bloqueo sobre la orden antes de registrar pagos, rectificaciones o entregas. La futura carga histórica debe preservar IDs, secuencias, PDF e historiales anteriores a las reglas vigentes; por ejemplo, no tratar asignaciones auxiliares históricas como altas nuevas.

## Ejecutar la API con una base nueva y aislada

1. Crea `elbravodeljac_sqlserver_dev` con nivel de compatibilidad 170 y aplica únicamente `esquema_entrega.sql`. Reserva `elbravodeljac_sqlserver_qa` para pruebas automatizadas. `pruebas_esquema.sql` carga datos de prueba y se ejecuta en otra base desechable.
2. Configura `SQLSERVER_DSN` como URL `sqlserver://` que indique la base y los parámetros TLS del servidor. Quita `POSTGRES_DSN` del proceso de la API. Configura la URL y el secreto JWT como variables privadas; no guardes sus valores en el repositorio.
3. Para una instalación vacía, establece `SQLSERVER_BOOTSTRAP_CONFIRM` con el nombre exacto de la base. Proporciona `DEV_ADMIN_USERNAME` para el dueño y las contraseñas privadas `DEV_ADMIN_PASSWORD`, `JAC_ADMIN_PASSWORD` y `JAC_MECHANIC_PASSWORD`. Desde `backend`, ejecuta `go run ./cmd/sqlserver-bootstrap`. El comando crea el dueño con rol Gerente, `administradora` con rol Administrador y `mecanico` con rol Mecánico, vinculado a un empleado activo. También crea los roles, permisos y tipos documentales en la misma transacción. El comando rechaza una base con datos. Se ejecuta por separado del arranque normal.
4. Configura `APP_ORIGIN`, `HTTP_ADDR`, `JWT_SECRET_BASE64` y `FRONTEND_DIR` como se indica en el [README del backend](../../backend/README.md). Para evidencias usa un MongoDB de desarrollo o pruebas aislado con `MONGODB_URI` y `MONGODB_DATABASE`; conserva su inicialización explícita y los nombres de base admitidos. Ejecuta `go run ./cmd/api` desde `backend`.

El workflow de la rama de migración levanta instancias desechables en GitHub Actions. La suite usa `SQLSERVER_QA_DSN` y `SQLSERVER_QA_MONGO_URI` y crea datos sintéticos. Este procedimiento no importa datos desde PostgreSQL.

## Instalación local con Docker Compose

Usa Docker Desktop con contenedores Linux y PowerShell 7 desde la raíz del repositorio. El proyecto `elbravodeljac-sqlserver-dev` tiene tres volúmenes propios. Solo Go publica un puerto en loopback. Standard Developer se usa exclusivamente para desarrollo y pruebas.

Para reanudar una instalación existente, ejecuta `./backend/scripts/start-dev.ps1`. Añade `-Restart` para reiniciar los tres contenedores. El script verifica y reutiliza los montajes existentes. Los pasos siguientes son para la primera instalación.

Los límites de contenedor suman 4,125 GiB para SQL Server, MongoDB y Go. SQL Server dispone de 2.560 MiB dentro de un contenedor de 3 GiB, con 512 MiB para procesos auxiliares. MongoDB conserva 768 MiB y Go 384 MiB. Estos límites son techos, no mediciones de consumo. Antes de arrancar, comprueba que Windows tenga al menos 5,125 GiB físicos disponibles para dejar 1 GiB de reserva y que Docker tenga margen para los tres contenedores. Durante la validación, comprueba la reserva de Windows, el uso máximo de los contenedores y la ausencia de reinicios inesperados.

Crea el archivo privado una sola vez y conserva una copia protegida fuera del repositorio. Las contraseñas de las tres cuentas se proporcionan al contenedor de inicialización de uso único.

```powershell
$existing = @(docker ps -aq --filter 'label=com.docker.compose.project=elbravodeljac-sqlserver-dev')
if ($LASTEXITCODE) { throw 'No se pudo comprobar el estado de Docker' }
$volumes = @(docker volume ls -q --filter 'label=com.docker.compose.project=elbravodeljac-sqlserver-dev')
if ($LASTEXITCODE) { throw 'No se pudieron comprobar los volúmenes' }
if ($existing.Count -or $volumes.Count) { throw 'Ya hay recursos de SQL Server. Usa start-dev.ps1 o revisa la instalación existente.' }
if (Test-Path .env.sqlserver.local) { throw 'El archivo privado ya existe' }
function New-JacPassword {
    'Aa9' + [Convert]::ToHexString([Security.Cryptography.RandomNumberGenerator]::GetBytes(24))
}
$jwt = [Convert]::ToBase64String([Security.Cryptography.RandomNumberGenerator]::GetBytes(32))
@(
    "JAC_SQL_SA_PASSWORD=$(New-JacPassword)"
    "JAC_SQL_APP_PASSWORD=$(New-JacPassword)"
    "JAC_MONGO_ROOT_PASSWORD=$(New-JacPassword)"
    "JAC_MONGO_APP_PASSWORD=$(New-JacPassword)"
    "JAC_JWT_SECRET_BASE64=$jwt"
    'JAC_HTTP_HOST=localhost'
    'JAC_HTTP_PORT=8080'
) | Set-Content -LiteralPath .env.sqlserver.local -Encoding utf8NoBOM
Remove-Variable jwt

function Invoke-JacCompose {
    docker compose --env-file .env.sqlserver.local -f compose.sqlserver.dev.yaml @args
    if ($LASTEXITCODE -ne 0) { throw 'Falló Docker Compose. Detén la secuencia y revisa el error.' }
}
Invoke-JacCompose config --quiet
Invoke-JacCompose pull sqlserver mongo
Invoke-JacCompose build api
```

Inicia los motores y aprovisiona la instalación una sola vez. Guarda las tres contraseñas en un gestor de contraseñas. Cada contraseña debe tener entre 12 y 72 bytes.

```powershell
Invoke-JacCompose up -d --wait --wait-timeout 180 sqlserver mongo
Invoke-JacCompose exec -T sqlserver /bin/sh /setup/init-clean.sh
try {
    $env:DEV_ADMIN_USERNAME = Read-Host 'Usuario del dueño'
    $env:DEV_ADMIN_PASSWORD = Read-Host 'Contraseña del dueño' -MaskInput
    $env:JAC_ADMIN_PASSWORD = Read-Host 'Contraseña de administradora' -MaskInput
    $env:JAC_MECHANIC_PASSWORD = Read-Host 'Contraseña de mecanico' -MaskInput
    Invoke-JacCompose run --rm --no-deps sqlserver-bootstrap
} finally {
    Remove-Item Env:DEV_ADMIN_USERNAME, Env:DEV_ADMIN_PASSWORD, Env:JAC_ADMIN_PASSWORD, Env:JAC_MECHANIC_PASSWORD -ErrorAction SilentlyContinue
}
Invoke-JacCompose run --rm --no-deps mongo-init
Invoke-JacCompose up -d --no-deps --no-build api
```

Abre `http://localhost:8080` cuando `/healthz` y `/readyz` respondan correctamente. Usa ese mismo origen al iniciar sesión con las tres cuentas iniciales. Comprueba las 28 tablas, 208 columnas y 50 FK, los cuatro tipos documentales y las colecciones e índices técnicos de MongoDB. Utiliza volúmenes independientes si ejecutas pruebas con operaciones sintéticas.

Si un paso de inicialización falla, inspecciona el estado antes de reintentarlo. El script SQL rechaza una base o login que ya exista.
