# Corrección funcional del catálogo de marcas — 27/10/2026

## Causa comprobada

La API anterior hacía SELECT DISTINCT marca FROM vehiculos: solo ofrecía marcas
que ya tenían vehículos. No era el catálogo de 34 marcas aprobado por el usuario.
Además, Nuevo vehículo precargaba JAC en un input/datalist: Chrome filtraba las
sugerencias por ese texto. La revisión anterior no cubría esta necesidad completa.
No faltaban vehículos que hubiera que restaurar; faltaba el catálogo independiente
de las marcas utilizadas. No se consultó ni modificó la web comercial.

## Solución aplicada

- Migración `002_marcas_vehiculos.sql`: crea/reemplaza la **vista** PostgreSQL
  `public.marcas_vehiculos`. Su definición persistente contiene las 34 marcas
  indicadas por el usuario y une las marcas adicionales guardadas en vehículos.
- No es otra lista visual en JavaScript. `GET /api/v1/vehiculos/marcas` consulta
  esa única vista, con sesión y permiso `vehiculos.ver`. Alta y edición consumen
  la misma respuesta y el mismo componente de selector.
- El catálogo elimina duplicados por nombre normalizado (sin tildes, sin diferencia
  de mayúsculas, sin espacios extremos), priorizando la escritura aprobada.
- Se conservan 26 tablas y 42 FK: no se agrega una tabla de marcas ni una nueva FK.
  La columna existente `vehiculos.marca varchar(40)` sigue almacenando la marca.
- El selector empieza vacío al crear, abre todas las marcas al hacer clic y permite
  buscar por primeras letras sin distinguir tildes o mayúsculas. Reutiliza estilos
  de los desplegables existentes; no se modificó CSS ni el diseño de las pantallas.
- «Otra marca» es una acción del selector, no el nombre que se almacena. Muestra
  un campo obligatorio para el nombre real. Se comprueba longitud hasta 40, letras/
  números y signos habituales, sin controles, etiquetas HTML ni nombres reservados.
  Go vuelve a validar todas las altas y modificaciones, aunque se omita la interfaz.
- Al guardar, Go usa la escritura canónica del catálogo cuando ya existe. Una
  marca personalizada se guarda en el vehículo, reaparece tras F5 y queda disponible
  en nuevas altas/ediciones mientras haya un vehículo que la utilice. Las 34 marcas
  base permanecen disponibles aunque no tengan vehículos. No se implementa un CRUD
  independiente de fabricantes ni se confunde la demostración con los datos reales.
- Las demás pantallas solo muestran la marca del vehículo elegido: no mantienen
  selectores ni listas independientes de fabricantes.

Marcas base: JAC, Toyota, Nissan, Hyundai, Kia, Chevrolet, Ford, Volkswagen, Suzuki,
Mitsubishi, Mazda, Renault, Chery, Changan, Haval, BYD, MG, Great Wall, BMW,
Mercedes-Benz, Audi, Volvo, Peugeot, Citroën, BAIC, Jinbei, DFM, Brilliance, Changhe,
Mahindra, SsangYong, Daewoo, Land Rover y Honda.

## Migración y conservación de datos

La migración es transaccional y repetible, con bloqueo asesor. No contiene DELETE,
UPDATE de vehículos, DROP TABLE, TRUNCATE ni cambios sobre columnas/FK existentes.
El comando db-migrate aplica 001 y 002, en ese orden, solo en bases _dev/_test.
Compose de pruebas y el script alternativo PostgreSQL local aplican ambos archivos.
No se modificó `database/postgresql/esquema_actual.sql`.

Se aplicó con `backend/scripts/start-dev.ps1` únicamente a `elbravodeljac_dev` del
proyecto `elbravodeljac-fase-e-dev`. Antes/después coincidieron exactamente los
13 vehículos y su huella MD5 ordenada: `8c0454178b5a142aa52600f917ecf4d3`.
Los nuevos vehículos posteriores son registros explícitos de pruebas funcionales.
No se modificó ElBravoDelJac ni se expuso PostgreSQL al puerto 5432 del host.
No se reiniciaron ni eliminaron bases/volúmenes de otros proyectos Docker.

## Pruebas ejecutadas

- `go test ./... -count=1 -v` en Compose aislado: aprobado. PostgreSQL 18.6 y MongoDB
  9.0 temporales; conserva la regresión A–D y la revisión funcional anterior.
- Nueva `TestVehicleBrandsPostgreSQL`: aprobada. Comprueba las 34 marcas antes de
  registrar vehículos, permiso 403, reejecución de migración sin alterar vehículos,
  altas/lecturas/ediciones Toyota, Hyundai, BMW, Honda y personalizada, normalización,
  rechazo de vacío/Otra marca/marcador interno/41 caracteres/HTML/control, deduplicado
  y conservación de 26 tablas/42 FK.
- `go vet ./...` en la imagen de pruebas: aprobado, salida 0.
- `node checks/run.mjs`: cuatro suites aprobadas; son pruebas de contratos/lógica,
  no se presentan como sustituto de las pruebas con BD y navegador.
- Nueva `node checks/brands-browser.mjs` en Chrome Windows: aprobada contra el modo
  real en Go/PostgreSQL. Comprueba que el desplegable ofrece las 34 marcas, búsqueda
  por prefijo, creación/consulta/edición/F5 de Toyota, Hyundai, BMW y Honda, cambio
  de marca en edición, Otra marca obligatoria y personalizada reutilizable tras F5.
  No se detectaron errores JavaScript. Informes excluidos de Git en
  `backend/.cache/brands-browser/resultado.json`.
- `node checks/review-browser.mjs`: regresión de observaciones aprobada con el nuevo
  selector. `node checks/live-docker.mjs`: recorrido completo, permisos/CSRF y GridFS
  aprobados; `read-persisted-docker.mjs`: lectura persistente aprobada. El primer
  recorrido se bloqueó porque la prueba omitía marca y suponía el antiguo JAC por
  defecto: se corrigió la prueba para elegir JAC explícitamente y pasó.
- `git diff --check` y sintaxis de los scripts JavaScript modificados: aprobados.
- El navegador opcional dentro del contenedor Go continúa SKIP porque no se habilita
  FRONTEND_E2E/runtime Node allí. La prueba Chrome Windows anterior sí se ejecutó.

## Ejecutar en Windows

Desde la raíz del repositorio, con Docker Desktop activo:

```powershell
.\backend\scripts\start-dev.ps1
```

No usar SkipBuild para incorporar el componente y la nueva migración.
Abrir **http://localhost:8080/** (modo real). Cuenta inicial `admin_pruebas`, contraseña
privada en `backend/.cache/acceso-desarrollo.txt`, excluida de Git. Recargar pestañas
abiertas antes de la actualización. `/demo/` continúa separado y sus datos locales
no constituyen el catálogo PostgreSQL.

Para repetir Chrome, con Playwright disponible en esta computadora:

```powershell
$env:PLAYWRIGHT_MODULE = "C:\Users\Usuario\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules\playwright"
node checks/brands-browser.mjs
```

## Git y archivos

Se partió del árbol limpio, commit `d8ebc98` (Modificaciones necesarias).
Rama conservada: backend/arquitectura-go-mongodb. Sin commit, push ni merge automático.

Archivos modificados o creados en esta corrección:

- `backend/integration/brands_test.go`
- `backend/internal/domain/models.go`
- `backend/internal/platform/postgres/brands.go`
- `backend/internal/platform/postgres/migrate.go`
- `backend/internal/platform/postgres/migrations/002_marcas_vehiculos.sql`
- `backend/internal/platform/postgres/store.go`
- `backend/scripts/test-local-postgres.ps1`
- `checks/brands-browser.mjs`
- `checks/frontend-browser.mjs`
- `checks/live-docker.mjs`
- `checks/review-browser.mjs`
- `compose.test.yaml`
- `dist/assets/js/components/brand-picker.js`
- `dist/assets/js/views/clients-vehicles.js`
- `docs/backend/CORRECCION_MARCAS.md`
- `docs/backend/ESTADO_ENDPOINTS.md`
- `docs/backend/REVISION_12_OBSERVACIONES.md`
