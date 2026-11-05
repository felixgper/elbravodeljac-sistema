# Asignación de mecánicos y flujo de diagnósticos

Implementación y comprobación: 27/10/2026. Rama `backend/arquitectura-go-mongodb`.
Punto de partida: commit `2563a3dc10e48c8cc1ec7adc671577189a13a059`.
No se hizo commit, push ni merge de estos cambios.

## Funcionamiento

1. Recepciones: al seleccionar Requiere diagnóstico técnico, aparece el selector
   de mecánicos activos registrados en Usuarios y roles. La interfaz exige
   seleccionar uno y guarda la recepción como Pendiente de diagnóstico.
   Servicio directo oculta/deshabilita ese selector y continúa a Cotizaciones.
2. Se reutilizan `usuarios`, `usuario_roles` y `roles`; no se crea otro catálogo
   de personas. Se admiten Mecánico/Supervisor activos según la validación técnica
   existente, sin eliminar relaciones N. Un trabajador inactivo o inexistente
   devuelve 422 y revierte toda la recepción, incluido el cambio de kilometraje.
3. El mecánico encuentra Mis diagnósticos pendientes en Inicio y Diagnósticos.
   Consulta número de recepción, cliente, placa/marca/modelo, motivo, ingreso y
   estado. El formulario muestra contexto y responsable en solo lectura.
4. Hallazgos, evaluación, recomendaciones y observaciones se guardan como
   Borrador o Finalizado. Las observaciones son opcionales. Para finalizar se
   requieren los otros tres campos; cada uno admite hasta 4000 caracteres sin
   truncamiento. Actor, fechas y cambios quedan registrados por el servidor.
   Un POST sin estado crea Borrador; guardar no finaliza implícitamente.
5. Los borradores persisten tras F5. PUT requiere `version` vigente; una edición
   concurrente devuelve 409. Reasignar exige `assignmentVersion`, conserva el
   contenido del borrador e historial y revoca el acceso del anterior mecánico.
   Los finalizados son inmutables y no se reasignan ni se sobrescriben.
6. Inicio muestra Diagnósticos pendientes de cotización para quien puede cotizar:
   finalizados sin cotización, responsable, fecha y hallazgos. Crear cotización
   abre el origen correcto y muestra la evaluación sin reescribirla. No se crea
   cotización automáticamente. Al registrarla sale de la bandeja de pendientes.
7. La cotización conserva servicios/repuestos, cantidades, precios, descuentos y
   aprobación existentes; un borrador se rechaza en Go y PostgreSQL. Las órdenes
   siguen requiriendo aprobación. El mecánico de ejecución de OT es independiente
   del mecánico del diagnóstico; asignar OT no sobrescribe el responsable técnico.
8. Las bandejas se actualizan al abrir el módulo, recuperar foco/visibilidad y
   cada 60 segundos visibles, con límite mínimo de tres segundos entre consultas.
   Solo se reemplaza la bandeja, conservando los campos en edición. Inicio muestra
   diez registros recientes y el total; Diagnósticos carga todas las páginas.

## Seguridad y compatibilidad

Todos los endpoints requieren sesión JWT existente; las escrituras utilizan
cookie HttpOnly y CSRF. Nuevos permisos: `diagnosticos.asignar` y
`diagnosticos.supervisar`; aparecen en la matriz de roles existente.

El permiso `diagnosticos.ver` por sí solo no habilita consultas ajenas. Go filtra
listas, detalles, opciones auxiliares y acceso a evidencias/escáner. Un mecánico
sin supervisión ve sus diagnósticos, y puede consultar el resultado asociado a una
OT si pertenece a su equipo activo autorizado. No puede modificar diagnósticos
ajenos. Adivinar otra URL produce 404 en lectura o 403 en escritura. Asignar una
OT no concede automáticamente supervisión de diagnósticos. Los permisos de acceso
al resto de módulos se mantienen según la matriz autorizada.

En la primera aplicación de 004 se agregan permisos de ver/asignar/supervisar a
Propietario, Administrador, Supervisor y Administrativo. Administrativo recibe
además crear/editar cotizaciones, conservando sus permisos anteriores. No se
reemplaza ninguna matriz ni asignación de roles. Esta concesión inicial está
protegida para no restaurar permisos retirados cuando la migración se repite.
Las cuentas técnicas requieren `diagnosticos.ver` y `diagnosticos.crear`; la
asignación determina sobre qué registros pueden trabajar.

La API permite recepciones técnicas todavía sin asignar por compatibilidad y las
identifica como Sin asignar. No permite diagnosticar hasta asignar un trabajador;
la interfaz de nueva recepción exige asignación. Las recepciones históricas se
pueden asignar desde su listado. Los diagnósticos previos se conservan como
Finalizado y mantienen resumen, responsable y fecha histórica; sus nuevos campos
vacíos no se rellenan con evaluaciones inventadas.

## Migración segura 004

Archivo: `backend/internal/platform/postgres/migrations/004_flujo_diagnosticos.sql`.
Aplicada solo en `elbravodeljac_dev` y `elbravodeljac_test` aisladas. No se restauró
ningún respaldo sobre bases existentes ni se accedió a `ElBravoDelJac` del puerto
5432. No se borraron registros ni se recrearon bases.

Incompatibilidad resuelta: el esquema previo no tenía asignación de diagnóstico
antes de su creación ni campos para borrador/evaluación/recomendaciones/historial.
Se amplían las dos tablas existentes, en transacción con bloqueo de migración:

- Recepciones: mecánico diagnóstico, asignador, fecha de asignación, versión e
  historial de asignaciones JSONB.
- Diagnósticos: estado, evaluación, recomendaciones, observaciones, versión,
  actualización/finalización e historial JSONB de cambios con actor y fecha.
- Dos FK nuevas hacia usuarios: responsable de diagnóstico y quien lo asignó.
  Se conservan las 26 tablas y las 44 FK anteriores a este cambio; quedan **46 FK**.
  Las 42 relaciones del esquema original permanecen, con las ampliaciones previas.
- Índices de asignación/estado y validaciones de origen de cotización y diagnóstico.

Antes/después de la primera migración: mismos 19 clientes, 28 vehículos,
19 recepciones, 15 diagnósticos, 17 cotizaciones, 12 pagos, 11 entregas y 26 usuarios;
el hash de las relaciones usuario_roles tampoco cambió. Las pruebas posteriores
agregaron sus propios registros identificados de desarrollo y los conservaron.

Se repitió también la migración completa: hashes de las filas de clientes,
vehículos, recepciones, diagnósticos, cotizaciones, pagos, entregas, usuarios,
relaciones de roles y matriz de permisos permanecieron idénticos. Los conteos de
todas las colecciones técnicas de MongoDB también permanecieron idénticos.

## Endpoints

| Operación | Uso |
|---|---|
| POST `/api/v1/recepciones` | Guarda `technicianId` y tipo de atención con asignación autorizada |
| PATCH `/api/v1/recepciones/{id}/mecanico` | Asigna/reasigna con `technicianId` y `version` de asignación |
| GET `/api/v1/me/diagnosticos` | Pendientes por asignación, paginados |
| GET `/api/v1/diagnosticos/pendientes-cotizacion` | Finalizados sin cotización, paginados |
| POST `/api/v1/diagnosticos` | Crea borrador o finalizado autorizado |
| PUT `/api/v1/diagnosticos/{id}` | Guarda/finaliza borrador con versión vigente |
| GET `/api/v1/recepciones`, GET `/api/v1/recepciones/{id}` | Contexto con alcance y asignación |
| GET `/api/v1/diagnosticos`, GET `/api/v1/diagnosticos/{id}` | Resultado con alcance, estado e historial |
| GET `/api/v1/frontend/opciones` | Opciones del formulario filtradas por autorización |
| POST `/api/v1/cotizaciones` | Origen técnico finalizado o recepción directa sin diagnóstico |

El cuerpo de diagnóstico admite `receptionId`, `status`, `findings`, `evaluation`,
`recommendations`, `notes`, `version`. `technicianId` es opcional; si se envía debe
coincidir con la asignación persistida. Identidad y fecha no se aceptan del cliente.

## Pruebas realmente ejecutadas

- `node checks/run.mjs`: cuatro suites aprobadas (verify, workflow, edge-cases,
  real-api), separando pruebas de simulación de la conexión real.
- Compose de pruebas: `go test ./... -count=1 -v`, aprobado con PostgreSQL 18 y
  MongoDB aislados; JWT/CSRF, autorizaciones, transacciones, concurrencia, pagos,
  GridFS, ficha PDF y regresiones anteriores.
- Nuevo `TestAssignedDiagnosisWorkflow`: dos mecánicos y auxiliar, asignaciones
  inválidas/inactivas, aislamiento por URL/API, estado omitido = Borrador,
  borrador no cotizable, validación de finalización, reasignación sin pérdida,
  versión desactualizada, finalizado inmutable, bandeja para cotizar, mecánico de
  OT diferente y servicio directo sin diagnóstico ficticio.
- `go vet ./...` en el contenedor de pruebas: aprobado, salida 0.
- `checks/diagnosis-browser.mjs`: Chrome real con administradora y dos mecánicos,
  asignación desde formulario, borrador, F5, finalización, actualización por foco,
  cotización con contexto técnico, directo y bloqueo 404/403 de accesos ajenos.
- `checks/live-docker.mjs`: Chrome real, recorrido completo hasta pago y entrega,
  recarga y consulta persistida, GridFS/reintentos, CSRF y 14 vistas sin errores JS.
- Se inspeccionó la captura real de la bandeja del mecánico, manteniendo el estilo.

`TestPhaseEFrontendBrowserPostgreSQLMongoDB` se omite dentro de `go test` porque
requiere `FRONTEND_E2E=1` y Node/Playwright, que esa imagen Go no contiene. No se
cuenta como pasado: la comprobación de navegador se ejecutó separadamente con los
dos scripts Chrome anteriores contra Go/PostgreSQL/MongoDB del entorno aislado.
`npm` no está en PATH en esta terminal; se ejecutó directamente el mismo comando
Node al que apunta `npm run check`, con resultado correcto.

Informes locales sin contraseñas: `backend/.cache/diagnosis-browser/resultado.json`,
`backend/.cache/live-docker/resultado.json` y `backend/.cache/diagnosis-tests.log`.
Las evidencias locales y las credenciales están excluidas de Git.

## Verificación manual en Windows

En la raíz del repositorio, con Docker Desktop iniciado:

```powershell
.\backend\scripts\start-dev.ps1
```

Abrir http://localhost:8080/ en Chrome; `/` es modo REAL. Los accesos privados de
administradora y mecánico están en `backend/.cache/acceso-diagnosticos.txt`.
No se incluyen sus contraseñas en documentación, JavaScript ni Git. Para probar
ambas cuentas a la vez, usar Chrome normal y una ventana de incógnito (orígenes
idénticos), ya que comparten cookie dentro de una misma sesión del navegador.

Se dejó la recepción **23**, asignada al mecánico indicado en ese archivo, para
revisión manual. Abrir Mis diagnósticos pendientes, guardar borrador y recargar;
la administradora todavía no debe verla como cotizable. Finalizar con los tres
campos técnicos obligatorios y volver al Inicio de la administradora para crear
la cotización. También puede crear otra recepción desde la cuenta administrativa.

El servidor publica únicamente 127.0.0.1:8080. Probar desde otras computadoras
requiere acordar una dirección LAN/HTTPS, origen y cookies seguros; no se ha
expuesto el entorno local a la red. No hay pendientes del flujo solicitado en
localhost. La carga visual de archivos técnicos continúa con el alcance previamente
acordado: APIs MongoDB disponibles; no se añaden controles nuevos en este cambio.

## Archivos modificados y creados

- `README.md`
- `backend/cmd/dev-bootstrap/main.go`
- `backend/integration/attention_sheet_test.go`
- `backend/integration/brands_test.go`
- `backend/integration/diagnosis_workflow_test.go` (nuevo)
- `backend/integration/phase_a_test.go`
- `backend/integration/phase_c_test.go`
- `backend/integration/phase_d_test.go`
- `backend/integration/review_test.go`
- `backend/integration/workshop_test.go`
- `backend/internal/domain/permission-policy.json`
- `backend/internal/domain/workshop.go`
- `backend/internal/httpapi/catalog_test.go`
- `backend/internal/httpapi/diagnosis_workflow.go` (nuevo)
- `backend/internal/httpapi/frontend_options.go`
- `backend/internal/httpapi/workshop.go`
- `backend/internal/platform/postgres/diagnosis_workflow.go` (nuevo)
- `backend/internal/platform/postgres/frontend.go`
- `backend/internal/platform/postgres/migrate.go`
- `backend/internal/platform/postgres/migrations/004_flujo_diagnosticos.sql` (nuevo)
- `backend/internal/platform/postgres/technical.go`
- `backend/internal/platform/postgres/workshop_reads.go`
- `backend/internal/platform/postgres/workshop_writes.go`
- `backend/scripts/test-local-postgres.ps1`
- `checks/attention-browser.mjs`
- `checks/diagnosis-browser.mjs` (nuevo)
- `checks/frontend-browser.mjs`
- `checks/live-docker.mjs`
- `checks/review-browser.mjs`
- `compose.test.yaml`
- `dist/assets/js/app.js`
- `dist/assets/js/components/diagnosis-inbox.js` (nuevo)
- `dist/assets/js/data/permissions.js`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/services/workshop-service.js`
- `dist/assets/js/views/operations.js`
- `docs/backend/ESTADO_ENDPOINTS.md`
- `docs/backend/FLUJO_DIAGNOSTICOS_ASIGNADOS.md` (nuevo)
