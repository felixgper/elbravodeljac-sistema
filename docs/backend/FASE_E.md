# Fase E · Frontend aprobado conectado a Go

Fecha: 27 de octubre de 2026. Rama: `backend/arquitectura-go-mongodb`.

**Actualización posterior:** Docker y el recorrido completo desde Chrome ya se
ejecutaron correctamente. Véase [VERIFICACION_DOCKER.md](VERIFICACION_DOCKER.md)
para el estado actual, la dirección local, inicio de sesión privado y persistencia
después de reiniciar contenedores. Las limitaciones Docker de este informe
corresponden a la ejecución inicial de Fase E.

El frontend de `dist/` consume las APIs de las fases A–D. Se conservan sus CSS,
imágenes, menú, colores y distribución. No se creó otra aplicación ni se cambió
el SQL aprobado. PostgreSQL sigue administrando 26 tablas y 42 FK.

## Pantallas y endpoints utilizados

Todas las rutas de esta tabla tienen prefijo `/api/v1`. Los listados recorren la
paginación del servidor, sin reemplazar resultados parciales por datos demo.

| Pantalla existente | Integración real |
|---|---|
| Acceso / perfil | `POST /auth/login`, `GET /auth/me`, `POST /auth/logout` |
| Clientes, alta, edición, eliminación | `GET,POST /clientes`, `PATCH,DELETE /clientes/{id}` |
| Vehículos, alta, actualización, eliminación | `GET,POST /vehiculos`, `PATCH,DELETE /vehiculos/{id}` |
| Historial y transferencia | `GET /vehiculos/{id}/historial`, `POST /vehiculos/{id}/transferencias` |
| Recepciones | `GET,POST /recepciones` |
| Diagnósticos | `GET,POST /diagnosticos` |
| Cotizaciones | `GET,POST /cotizaciones`, `PUT /cotizaciones/{id}`, `GET /servicios`, `GET /repuestos` |
| Órdenes, asignación, avances y pruebas | `GET,POST /ordenes`, `PATCH,DELETE /ordenes/{id}` |
| Pagos y entrega | `GET,POST /ordenes/{id}/pagos`, `GET,POST /ordenes/{id}/entrega` |
| Contacto para retiro | `GET,POST /clientes/{id}/contactos` |
| Auxilios | `GET,POST /auxilios`, `GET /unidades-auxilio`, `GET /mecanicos` |
| Mis tareas | `GET /me/tareas`, `PATCH /ordenes/{id}` si existe `ordenes.editar` |
| Mecánicos / asignaciones | `GET /mecanicos`; órdenes y asignaciones de la proyección auxiliar autorizada |
| Proveedores | `GET,POST /proveedores`, `PATCH,DELETE /proveedores/{id}`, `GET /proveedores/{id}/repuestos` |
| Usuarios | `GET,POST /usuarios`, `PATCH,DELETE /usuarios/{id}`, `GET /roles` |
| Matriz de permisos | `GET /permisos`, `GET,PUT /roles/{id}/permisos`, `POST /roles/{id}/permisos/restablecer` |
| Dashboard | `GET /dashboard` |
| Reportes y filtros | `GET /reportes?period=Este%20mes` o `Últimos 6 meses` |

Las pantallas consultan las entidades relacionadas que necesitan. Para formularios
cuyo rol no puede entrar a otro directorio, se agregó:

`GET /api/v1/frontend/opciones?module=ordenes&resource=clientes&page=1&pageSize=100`

Este endpoint exige sesión y `module.ver`, aplica una lista cerrada de relaciones
por pantalla y devuelve proyecciones auxiliares, no concede permisos CRUD. No
expone usuarios completos ni contraseñas. Clientes auxiliares contienen ID,
nombre y documento; vehículos contienen identificación, propietario y kilometraje.
El personal técnico contiene nombre, roles y carga de trabajo. Los trabajadores
sin acceso global a órdenes reciben referencias de sus asignaciones activas.
`mecanicos.ver` habilita la consulta del directorio de asignaciones. Contactos
desde Pagos requieren `module=pagos&resource=contactos&orderId={id}`.
Los comandos siguen verificando los permisos y reglas de PostgreSQL en Go.

Permanecen implementadas las 71 operaciones originales y las tres ampliaciones
de C. La nueva consulta auxiliar es una operación adicional, fuera del catálogo.
No se agregaron respuestas 501.

## Comportamiento y seguridad

- Go sirve `/` en modo real y `/demo/` en demostración. Solo publica el HTML y
  archivos públicos de `dist/assets/`, sin listar directorios ni servir `.env`,
  SQL o código del backend. Las APIs y el frontend comparten origen.
- El servidor inyecta el modo en el HTML. La demo servida con Python sigue
  identificada como demostración; no se convierte automáticamente al modo real.
- Sesión real y CSRF viven en memoria. JWT permanece en cookie HttpOnly,
  SameSite Strict. Las escrituras usan `credentials: same-origin` y
  `X-CSRF-Token`; Go comprueba también Origin. No hay secretos en JavaScript.
- Cada navegación consulta la sesión actual y sus permisos efectivos, unión de
  todos los roles. Un nombre de rol no concede acceso. Un 401 elimina la sesión
  y la instantánea de datos. Las respuestas de vistas antiguas no sustituyen la
  vista nueva. Los borradores se separan por modo y usuario.
- Modo real nunca lee ni guarda las tablas demo de localStorage. No hay respaldo
  automático a datos ficticios cuando falla una API. Los datos reales consultados
  solo se mantienen en una instantánea en memoria para las vistas existentes.
- Formularios y confirmaciones esperan la API antes de notificar, cerrar o
  navegar. Errores HTTP y de conexión se presentan en el formulario; no generan
  confirmaciones falsas. Los controles de guardado se deshabilitan durante la petición.
- Cotizaciones seleccionan ID real del catálogo dentro de la misma celda
  «Concepto». No se adivinan referencias por nombre. El servidor conserva stock,
  precios, aprobación y restricciones. Importes permanecen decimales en HTTP;
  las vistas usan centavos exactos y rechazan totales fuera del rango seguro.
- El código visible de la OT se distingue de su ID usado en las URLs. Las
  bitácoras separadas de avances, pruebas y estados conservan sus eventos.
- Seleccionar otro rol agrega la relación N; no envía `removeRoleId`. Se muestran
  los roles actuales, se conserva el resto y se evita reasignar roles al editar
  solo un nombre. Go bloquea autoasignación y pérdida del último administrador.
- Datos del vehículo y transferencia son dos comandos independientes. Si se
  intentan modificar juntos, el formulario exige guardar primero los datos y
  después confirmar la transferencia, evitando anunciar una operación parcialmente guardada.

## MongoDB y límites de las pantallas aprobadas

`technical-service.js` implementa listados y registro de escáner, listado/carga
de evidencias y URL de descarga con la misma cookie y autorización PostgreSQL:

- `GET,POST /diagnosticos/{id}/escaneres`
- `GET,POST /diagnosticos/{id}/evidencias`
- `GET /diagnosticos/{id}/evidencias/{evidenciaId}/archivo`

La carga usa multipart, clave `requestId` conservada en reintentos y máximo 10 MiB.
Solo PDF, PNG y JPEG; el backend verifica el contenido. Se probaron estos servicios
desde JavaScript en el navegador contra MongoDB y GridFS reales, incluida la descarga.

**Pendiente visual:** el formulario aprobado de diagnóstico no contiene controles
de archivos o resultados de escáner. No se insertó una pantalla distinta. Falta
acordar esos controles y enlazarlos con los servicios ya probados.

**Proveedores:** «especialidad» no existe en PostgreSQL. En modo real ese espacio
es consulta de repuestos vinculados al catálogo y se deshabilita como texto libre,
con explicación visible. El servicio rechaza expresamente texto de especialidad;
no lo coloca en contacto ni lo descarta. Altas, edición y eliminación de los
campos compatibles funcionan. Faltan controles aprobados para editar vínculos y
costos de repuestos; los endpoints de C permanecen disponibles.

Tampoco se inventaron pantallas de alta de catálogos ni de reasignación posterior
de auxilios: las APIs existentes admiten esas operaciones, pero el diseño actual
solo contiene selección de catálogos y registro inicial de auxilio. El auxiliar
sin `ordenes.editar` consulta sus tareas sin botón de actualización habilitado.

## Pruebas y reproducción

No se utilizó `ElBravoDelJac`. La suite nativa crea clústeres nuevos con nombres
exactos de prueba, directorios nuevos en `backend/.cache/`, puertos aleatorios
loopback y cuentas efímeras; importa el esquema solo en esas bases nuevas y
detiene ambos procesos al terminar. No ejecuta DROP TABLE ni TRUNCATE.

**Resultado final:** pruebas Node aprobadas, suite Go completa aprobada con ambas
bases y navegador habilitados, y `go vet ./...` sin hallazgos. Se generaron
capturas de las 15 rutas en `.cache/frontend-E-<sufijo>/` y se revisaron
visualmente Dashboard, Clientes y la matriz. Los estilos e imágenes no tienen
diferencias en Git.

Pruebas ejecutadas durante esta fase:

1. `node checks/verify.mjs`, `workflow.mjs`, `edge-cases.mjs`: demo, correspondencia
   de 23 pantallas/13 módulos, flujo completo y casos límite existentes.
2. `node checks/real-api.mjs`: pruebas unitarias del transporte con dobles HTTP
   controlados, DTO, cookies/CSRF, permisos efectivos, roles N, referencias
   obligatorias, decimales, paginación, errores 409/red/401 y limpieza de datos.
   Estos dobles no se presentan como prueba de conexión a una base de datos.
3. `go test ./... -count=1 -v` a través de `test-local-databases.ps1`: integración
   A/B/C/D con PostgreSQL 18 y MongoDB 9 reales, reglas, FK, transacciones,
   concurrencia, autorización y fallos MongoDB. Confirma 26 tablas y 42 FK.
4. `TestPhaseEFrontendBrowserPostgreSQLMongoDB`: Edge sin interfaz + Playwright,
   Go HTTP real, PostgreSQL y MongoDB reales. Recorre formularios y 15 rutas de
   consulta; comprueba cookies HttpOnly, matriz de permisos, denegaciones 403,
   ausencia de datos demo y fallo de red sin éxito falso. Prueba también los
   servicios de escáner/GridFS, descarga e idempotencia desde el navegador.
5. `go vet ./...`: análisis estático. `node --check` sobre los módulos JavaScript.
   Prueba HTTP adicional de publicación estática y bloqueo de archivos privados.

Docker no pudo ejecutarse: falta la tubería `dockerDesktopLinuxEngine`.
La ejecución de contenedores Compose queda bloqueada; las pruebas nativas sí
se ejecutaron con ambos motores. `npm` no está en PATH en este entorno; se
usaron los comandos Node directamente, equivalentes a `npm run check`.

Para repetir las pruebas de frontend sin npm, desde la raíz:

```powershell
node checks/run.mjs
```

Para repetir las pruebas integradas con navegador, requiere Node, Playwright,
Edge, PostgreSQL 18 y MongoDB 9. `PLAYWRIGHT_MODULE` puede apuntar a la instalación
local del paquete Playwright; si está instalado en node_modules no hace falta:

```powershell
$env:FRONTEND_E2E = '1'
# Solo si Playwright está instalado fuera del proyecto:
# $env:PLAYWRIGHT_MODULE = 'RUTA_ABSOLUTA_AL_PAQUETE_PLAYWRIGHT'
# Opcional, otro Chromium instalado: $env:BRAVO_BROWSER = 'RUTA_AL_EXE'
./backend/scripts/test-local-databases.ps1
Set-Location backend
go vet ./...
```

Sin `FRONTEND_E2E=1`, Go informa SKIP de la prueba de navegador; sin las variables
de integración informa SKIP de las bases. Un SKIP no demuestra una conexión.

## Errores detectados y corregidos

Se corrigió una carrera entre respuestas de sesión durante navegación rápida,
que podía dejar una vista cargando. Solo la respuesta de la consulta de sesión
vigente actualiza la sesión; las vistas antiguas no concluyen el render nuevo.
También se alinearon los DTO, referencias de catálogo, códigos de OT y longitudes
con PostgreSQL. Los gráficos escapan nombres procedentes de la base y presentan
meses legibles, conservando el diseño.

Durante la ampliación del ensayo de navegador se corrigieron sus selectores
(meta `description` frente a textarea), la búsqueda de filas paginadas y la
lectura de respuestas DELETE 204. Esos ensayos fallidos no se contaron como
aprobados: la última suite completa terminó correctamente. Docker continúa
bloqueado por falta de su motor, según se detalla arriba.

## Ejecutar la aplicación completa en Windows

1. Tener un PostgreSQL **separado** `elbravodeljac_dev` con el esquema aprobado,
   una cuenta de desarrollo con bcrypt y permisos asignados, y catálogos válidos.
   No reutilizar la base real ni repetir el respaldo sobre ella. La API no
   restaura SQL, siembra usuarios ni transforma contraseñas demo. Si la base de
   desarrollo está vacía, primero debe provisionarse su cuenta inicial autorizada.
2. Tener MongoDB aislado y usar `elbravodeljac_tecnico_dev`. Configurar privadamente
   las variables que aparecen debajo. `.env.example` es referencia: Go no lo
   carga automáticamente. No enviar los valores por Git ni ponerlos en JavaScript.
3. Desde `backend/`, generar la clave en memoria y configurar el origen exacto:

```powershell
$jwtBytes = New-Object byte[] 32
[System.Security.Cryptography.RandomNumberGenerator]::Fill($jwtBytes)
$env:JWT_SECRET_BASE64 = [Convert]::ToBase64String($jwtBytes)
$env:APP_ORIGIN = 'http://localhost:8080'
$env:HTTP_ADDR = '127.0.0.1:8080'
$env:COOKIE_SECURE = 'false' # exclusivamente HTTP local
$env:FRONTEND_DIR = '../dist'
# Definir POSTGRES_DSN y MONGODB_URI privadamente con los puertos/credenciales
# de las instancias aisladas. POSTGRES_DSN debe terminar en elbravodeljac_dev.
$env:MONGODB_DATABASE = 'elbravodeljac_tecnico_dev'
go run ./cmd/mongo-init
# Continuar solo si la inicialización anterior terminó sin error.
go run ./cmd/api
```

4. Abrir `http://localhost:8080/` e iniciar sesión con la cuenta de desarrollo.
   Usar exactamente localhost, no alternarlo con 127.0.0.1 porque Origin cambia.
   Comprobar `/readyz`: valida ambos motores y la inicialización MongoDB. El
   enlace `/demo/` abre los ejemplos separados, sin persistirlos en las bases.
5. Para Docker, `compose.dev.yaml` ya publica Go en localhost:8080 e incluye
   ambos motores aislados. Configurar las variables privadas DEV_POSTGRES_PASSWORD,
   DEV_MONGO_PASSWORD y JWT_SECRET_BASE64, levantar postgres-dev/mongo-dev,
   ejecutar el perfil de inicialización MongoDB y después api. La imagen ahora
   incluye `dist/`. No se verificó su ejecución porque el motor Docker no estaba activo.

En HTTPS se exige COOKIE_SECURE=true. No usar Live Server/Python para el modo real:
esos servidores no publican la API Go bajo el mismo origen.
