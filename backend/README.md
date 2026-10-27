# Backend Go: fases A, B, C y D

APIs reales de acceso, Clientes, contactos, Vehículos y proceso del taller:
recepción, diagnóstico, cotización, OT, asignaciones, avances/pruebas, pagos y entrega.
[Estado, contratos y pruebas](../docs/backend/FASE_D.md).
[Inventario de las 71 operaciones](../docs/backend/ESTADO_ENDPOINTS.md).
Incluye usuarios/roles/matriz, auxilios, proveedores, catálogos, Dashboard/reportes
y resultados de escáner/evidencias con MongoDB y GridFS (71/71 del catálogo).
Requiere Go 1.26 o superior; verificado con Go 1.27.1.

## Pruebas unitarias y análisis estático

Desde backend/:

```powershell
go mod download
go test ./... -count=1 -v
go vet ./...
```

Sin INTEGRATION_DATABASE_URL, la integración muestra SKIP: eso no prueba PostgreSQL.
La integración usa HTTP real (httptest.Server), PostgreSQL real y el esquema aprobado.

## PostgreSQL aislado en Windows

Desde la raíz del repositorio, con PostgreSQL 18 y Go instalados:

```powershell
./backend/scripts/test-local-postgres.ps1
# Para otra instalación:
./backend/scripts/test-local-postgres.ps1 -PostgresBin 'D:/PostgreSQL/18/bin'
```

Crea un clúster nuevo en backend/.cache/pg-test-<guid>, puerto loopback aleatorio y
base elbravodeljac_test. Inicializa SOLO esta base nueva con el esquema aprobado,
ejecuta todas las pruebas y detiene el clúster en finally. No usa el servicio,
credenciales ni puerto habituales. Trust se limita al clúster local descartable.
Los archivos detenidos quedan en .cache, excluidos de Git, para revisar fallos.
Se requieren permisos de conexión TCP local. No se elimina ninguna carpeta existente.

La suite genera sus propias contraseñas, usuarios y roles, sin datos reales ni
accesos fijos. Solo admite la base exacta elbravodeljac_test. No modifica el esquema.

## Docker Compose de pruebas

Con Docker Engine activo y PowerShell 7, desde la raíz:

```powershell
./backend/scripts/test-compose.ps1
```

Genera contraseña efímera en el entorno, usa compose.test.yaml, ejecuta el contenedor
Go y baja los contenedores al finalizar. PostgreSQL usa tmpfs, no publica puertos y
está en una red interna. Monta el SQL original de solo lectura para inicializar el
contenedor nuevo. No hay volumen persistente ni contraseña fija en Git.
El servicio Go usa backend/Dockerfile.test. Para validar sintaxis sin mostrar secretos:
`docker compose -f compose.test.yaml config --quiet` (con la variable de prueba configurada).

## Arrancar la API

.env.example es referencia; Go NO lo carga automáticamente. Define privadamente:
POSTGRES_DSN, JWT_SECRET_BASE64, APP_ORIGIN y opcionalmente HTTP_ADDR/COOKIE_SECURE.
La fase A rechaza bases distintas de elbravodeljac_test/elbravodeljac_dev antes de
conectar. El servidor no crea tablas, restaura dumps, siembra usuarios ni migra hashes.
Las cuentas para pruebas manuales deben provisionarse con bcrypt en la base separada;
la suite de integración crea cuentas efímeras y no deja una contraseña fija.

Generar clave en memoria con PowerShell 7, sin compartir ni publicar su valor:

```powershell
$jwtBytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($jwtBytes)
$env:JWT_SECRET_BASE64 = [Convert]::ToBase64String($jwtBytes)
$env:APP_ORIGIN = 'http://localhost:8080'
$env:COOKIE_SECURE = 'false' # excepción solo para HTTP loopback
# Configura POSTGRES_DSN privadamente hacia la base separada.
# Desde backend/:
go run ./cmd/api
```

Por defecto escucha 127.0.0.1:8080. En HTTPS usa Secure=true. Login exige Origin
exacto y application/json. Devuelve cookie HttpOnly bravo_access y csrfToken.
GET /api/v1/auth/me recupera csrfToken. Escrituras requieren cookie, Origin y
X-CSRF-Token. Logout revoca de inmediato. No hay CORS abierto; la integración del
frontend utilizará un proxy del mismo origen. La demo y dist/ permanecen intactos.

## Organización

- cmd/api: configuración, pool, servidor y cierre ordenado.
- internal/httpapi: middleware, handlers y catálogo de rutas futuras.
- internal/auth: JWT, CSRF y revocación en memoria.
- internal/config: secretos y bloqueo de bases reales en fase A.
- internal/application: puerto Store tipado.
- internal/domain: DTO, validadores y precisión decimal.
- internal/platform/postgres: SQL parametrizado y transacciones.
- internal/platform/mongodb: driver oficial, validadores, índices, GridFS y conciliación.
- integration: suite HTTP/PostgreSQL aislada.

El permiso efectivo une los códigos de TODOS los roles existentes. No hay bypass
por nombre de rol ni edición de asignaciones en login. Las sesiones y el limitador
son de una instancia: reiniciar invalida sesiones. Antes de producción faltan estado
compartido, despliegue TLS/proxy y observabilidad. No hay refresh token.

## MongoDB y pruebas completas

Configurar MONGODB_URI y MONGODB_DATABASE (_tecnico_dev o _tecnico_test; ver .env.example).
Desde backend/: `go run ./cmd/mongo-init` inicializa explícitamente validadores e
índices del entorno separado; `go run ./cmd/api` usa esa configuración sin alterarlos.
Sin URI, las APIs SQL siguen operativas y las técnicas informan 503, no éxito falso.

Con PostgreSQL 18 y MongoDB instalados, desde la raíz:

```powershell
./backend/scripts/test-local-databases.ps1
```

Se puede indicar -PostgresBin y -MongoBin. Crea dos instancias temporales aisladas,
ejecuta todas las pruebas A–D y las cierra. Sin variables INTEGRATION_MONGODB_URI,
INTEGRATION_MONGODB_DATABASE e INTEGRATION_DATABASE_URL, la suite D muestra SKIP.

compose.test.yaml incluye ahora ambos motores; compose.dev.yaml ofrece desarrollo
aislado con inicialización MongoDB explícita. Comandos, configuración privada,
contratos multipart, reintentos y resultados comprobados en FASE_D.md.
