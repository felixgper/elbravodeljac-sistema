# Catálogo documental y tres roles

Validado el 29 de octubre de 2026 en `feature/catalogo-documentos-tres-roles`,
creada desde `backend/arquitectura-go-mongodb`. Sin commit, push ni merge.
Los modelos académicos V19–V22 permanecen en sus ramas documentales.

## Comparación

| Elemento | Antes | Después |
|---|---|---|
| Esquema físico | 27 tablas / 48 FK | 28 tablas / 50 FK |
| Identidad de clientes/contactos | Tipo textual | Catálogo y dos FK obligatorias; JSON compatible |
| Identidad en entregas | Snapshot del receptor | Conservado, independiente de cambios posteriores |
| Oferta de roles | Seis roles anteriores | Gerente, Administrador, Mecánico |
| Cargo laboral | `empleados.tipo_empleado` | Independiente de los roles N de usuario |
| Orden nueva | Mecánico y auxiliar opcional | Mecánico obligatorio; sin auxiliar nuevo |
| Auxiliar histórico | Asignaciones `AUXILIAR` | Filas conservadas y consultables |
| Auxilio mecánico | Atención fuera del taller | Conservado; técnico con rol Mecánico |

## Migraciones y compatibilidad

001–007 no cambian. `008_tipos_documento.sql` agrega el catálogo con PK,
código único, denominación y estado activo, las FK NOT NULL de clientes y
contactos y un índice único `(id_tipo_documento, numero_documento)` de clientes.
Valores iniciales: DNI, RUC, CE y PASAPORTE. Las equivalencias conocidas se
normalizan sin cambiar números; una colisión entre alias aborta toda la
transacción, sin fusionar personas. Los códigos desconocidos reciben una
entrada inactiva `HIST_…`, conservando su texto original.

Se mantiene `documentType`/`documentNumber` en JSON, incluido `Pasaporte`.
La columna textual queda como compatibilidad, sincronizada con la FK mediante
triggers. Una identidad histórica inactiva puede conservarse al editar otros
campos; cambiar su identidad exige un tipo activo. Se validan formato y
longitudes sin truncar. Los formularios consultan el catálogo real y mantienen
«Razón social» para RUC; el modo demostración tiene datos separados. Las
entregas conservan sus snapshots y PDF, sin agregarles FK al catálogo.

`009_tres_roles.sql` agrega `roles.activo`, que controla la disponibilidad
para nuevas asignaciones. Propietario cambia a Gerente conservando su ID.
El Administrador anterior queda como `Administrador (legado)`, conservando
cuentas y permisos; se crea un Administrador operativo. Supervisor,
Administrativo, Auxiliar y demás roles preexistentes se retiran para nuevas
asignaciones sin inferir equivalencias de cuentas. Una colisión previa con los
nombres de destino detiene la migración para revisión. La marca de transición
impide que reaplicarla reponga permisos retirados posteriormente.

Los permisos heredados siguen siendo efectivos durante la revisión:
`activo=false` retira la oferta del rol, no bloquea silenciosamente sus cuentas.
La interfaz muestra las asignaciones actuales y permite retirar una heredada
solo mediante selección explícita autorizada. Editar un nombre no convierte
un rol antiguo en Mecánico ni borra roles adicionales.

RBAC N se conserva; no hay endpoint para crear un cuarto rol, ni tabla
`usuario_permisos`. Un rol aprovisionado explícitamente por SQL en el futuro
no obliga a rehacer RBAC, pero requiere otra decisión. Las pruebas crean
roles propios para no modificar las matrices del producto.

## Seguridad

| Rol nuevo | Política inicial |
|---|---|
| Gerente | 40 permisos de dirección, usuarios, matriz y operaciones |
| Administrador | 34 permisos operativos; sin módulo usuarios ni `pagos.rectificar` |
| Mecánico | 8: inicio, vehículos, recepciones, diagnósticos ver/crear, órdenes ver/editar y auxilio ver |

Fuente: `backend/internal/domain/permission-policy.json`. Se consultan los
permisos vigentes en PostgreSQL en cada petición, sin lista fija en JWT.
La asignación de roles/matrices exige permisos y no permite conceder códigos
que el actor no posee. Se bloquean cambios propios de seguridad y la edición
de la matriz Gerente. Transacciones serializan las mutaciones y protegen el
último acceso administrativo y, una vez establecido, el último Gerente activo.
Cookies HttpOnly, SameSite, CSRF y revocación siguen activos. Cambiar roles o
desactivar una cuenta invalida sus sesiones; revocar permisos afecta las
peticiones siguientes. Reiniciar Go exige volver a iniciar sesión.

Las órdenes no envían `assistantId`; el backend devuelve 422 al intentar
asignarlo o modificarlo. El DTO conserva ese campo para rechazar claramente
clientes antiguos. Un trigger bloquea también inserciones, reasignaciones y
reactivaciones SQL de `AUXILIAR`. Reasignar mecánico no modifica esas filas.
Los DTO históricos incluyen el nombre del participante aunque ya no forme
parte del directorio de mecánicos. Mis tareas y edición técnica exigen
asignación activa `MECANICO`. Auxilio mecánico permanece independiente.

## Conservación comprobada en desarrollo

Se inventariaron cuentas, roles y permisos antes de migrar. Por autorización
expresa se agregó Gerente a la única cuenta administrativa de pruebas,
conservando también su asignación heredada y exactamente sus 40 permisos.
No se convirtió automáticamente ninguna cuenta Supervisor/Administrativo/Auxiliar.

La comparación antes/después verificó igualdad de huellas y conteos de las
23 tablas de datos transaccionales, incluidas las 35 cuentas, los 26 clientes,
el contacto autorizado y las 36 asignaciones (25 mecánicas y 11 auxiliares).
Se excluyeron de la comparación únicamente las dos FK nuevas de las personas;
sus números y demás datos permanecieron iguales. Los permisos efectivos de
todas las cuentas permanecieron idénticos. También coincidieron conteos y
huellas SHA-256 de las cuatro colecciones MongoDB, incluidos archivos técnicos.

Los siete roles almacenados incluyen tres disponibles y cuatro heredados.
Las seis asignaciones Supervisor y catorce Auxiliar requieren revisión humana
de cuentas; hay usuarios con más de un rol, por lo que no se suman como personas.
La asignación Administrador legado de la cuenta ahora Gerente también se
conserva hasta que se solicite retirarla explícitamente.

Inventarios y huellas se guardan en `backend/.cache/catalogo-roles/`, excluido
de Git. No se publican nombres de cuentas ni registros personales.
No se accedió a la base real `ElBravoDelJac` ni se reiniciaron volúmenes.

## API, Windows y pruebas

Nuevo `GET /api/v1/tipos-documento`: sesión y `clientes.ver` o `pagos.ver`.
Respuesta `{data: [{id, code, name, active, documentType, label}]}`.
No se añade CRUD público de tipos. `/roles` añade `active`; rutas existentes
de clientes, contactos, usuarios, diagnósticos, órdenes y auxilios conservan
sus contratos salvo el rechazo explícito de auxiliares nuevos.

Desarrollo real: `http://localhost:8080/`. Las credenciales existentes están
en `backend/.cache/acceso-desarrollo.txt`, excluido de Git. La cuenta autorizada
ahora tiene Gerente. La API se reconstruyó sin reiniciar PostgreSQL ni MongoDB.

Pruebas PowerShell desde la raíz:

```powershell
# Solo la primera vez; falla si el proyecto ya existe, nunca lo recrea.
.\backend\scripts\test-audit.ps1 -StartIsolated
# Ejecuciones posteriores sobre el mismo cluster.
.\backend\scripts\test-audit.ps1
```

Requiere Docker Desktop, Go, Node, Playwright y Chrome/Edge. Proyecto separado
`elbravodeljac-catalogo-roles-test`: PostgreSQL en loopback 25433 y MongoDB
27030. No utiliza el PostgreSQL original 5432. Las pruebas de transición crean
bases `jac_catalog_migration_…` nuevas dentro del cluster de pruebas y las
conservan para inspección; no restauran estructuras sobre bases existentes.
El cluster de pruebas usa tmpfs: no detenerlo si se desean conservar sus datos.
El script no ejecuta down, reinicios ni eliminación de contenedores/volúmenes.

Resultados ejecutados:

- `node checks/run.mjs`: cinco suites aprobadas, incluyendo catálogo real,
  conservación de roles N y ausencia de `assistantId` en nuevas peticiones.
- `go test ./... -v -count=1`: aprobado con PostgreSQL y MongoDB reales de
  pruebas; no se omitieron los dos recorridos de navegador.
- `go vet ./...`: aprobado.
- Migración con alias duplicado: rollback completo; luego migración válida,
  conservación de identidades desconocidas, FK, unicidad y reaplicación segura.
- HTTP: Gerente, Administrador, Mecánico, último Gerente, autoescalamiento,
  JWT inválido, cuenta inactiva, revocación de permisos y cambio de rol en sesión.
- Recorridos reales: cliente/vehículo, directo y diagnóstico, aprobación,
  órdenes sin auxiliar, avances/pruebas, pagos, contacto, entrega, PDF e historial;
  auxilio independiente, usuarios/empleados y matriz. MongoDB/GridFS, reintentos
  y fallo de servicio también forman parte de las suites aprobadas.
- Auditoría de Chrome y suite de navegador Fase E: aprobadas. Las pruebas
  antiguas se adaptaron a aprobación explícita, ejecución conciliada,
  conformidad y paginación; no se relajaron las reglas del producto.

Evidencia local: `backend/.cache/audit-native.log`,
`backend/.cache/catalogo-roles/suite.log`, `backend/.cache/audit-browser/`.
Pendiente: revisar individualmente los roles heredados y aprobar una futura
migración empresarial; ninguna migración se aplicó a datos reales de producción.

## Propuesta para los diagramas (sin regenerarlos)

- Chen: nueva entidad TIPOS DOCUMENTO y dos relaciones con CLIENTES y CONTACTOS
  DEL CLIENTE. Cada persona tiene un tipo; un tipo identifica cero o muchas personas.
- Crow's Foot: dos relaciones 1 a 0:N; participación obligatoria del cliente/contacto.
  EMPLEADOS conserva su cargo independiente de USUARIOS y los roles N.
- Físico: tabla de cuatro columnas, dos FK NOT NULL, índice único compuesto
  y `roles.activo`. Mantener FK de actores hacia USUARIOS.
- Asignaciones: mecánico responsable en operaciones nuevas; AUXILIAR histórico.
  Auxilio mecánico sigue siendo una entidad distinta. Entregas mantiene snapshots.

## Archivos y dependencias afectados

Backend: migraciones 008/009 y su registro; catálogo PostgreSQL/HTTP y router;
modelos/validación/política RBAC; clientes, identidad, lecturas y escrituras de
órdenes/directorio técnico; bootstrap de desarrollo. No cambian dependencias Go.

Frontend: catálogo documental y estado/carga real; formularios de clientes y
contactos; operaciones, administración y permisos; servicios reales y modo
demostración separado. Sin cambios de CSS, colores, menú ni estructura visual.

Validación: `catalog_roles_test.go`, fixtures de integración afectados por el
nuevo esquema/roles, suites workflow/edge-cases/real-api y navegadores; Compose
de pruebas y override dedicado; script de pruebas, README y endpoints.

### Inventario de archivos de esta rama

- `README.md`
- `backend/cmd/dev-bootstrap/main.go`
- `backend/integration/audit_browser_test.go`
- `backend/integration/audit_corrections_test.go`
- `backend/integration/brands_test.go`
- `backend/integration/catalog_roles_test.go`
- `backend/integration/phase_a_test.go`
- `backend/integration/phase_c_test.go`
- `backend/integration/review_test.go`
- `backend/integration/workshop_test.go`
- `backend/internal/domain/models.go`
- `backend/internal/domain/permission-policy.json`
- `backend/internal/domain/workshop.go`
- `backend/internal/httpapi/documents.go`
- `backend/internal/httpapi/router.go`
- `backend/internal/platform/postgres/diagnosis_workflow.go`
- `backend/internal/platform/postgres/documents.go`
- `backend/internal/platform/postgres/identity.go`
- `backend/internal/platform/postgres/migrate.go`
- `backend/internal/platform/postgres/migrations/008_tipos_documento.sql`
- `backend/internal/platform/postgres/migrations/009_tres_roles.sql`
- `backend/internal/platform/postgres/relational.go`
- `backend/internal/platform/postgres/store.go`
- `backend/internal/platform/postgres/workshop_reads.go`
- `backend/internal/platform/postgres/workshop_writes.go`
- `backend/scripts/test-audit.ps1`
- `checks/audit-browser.mjs`
- `checks/edge-cases.mjs`
- `checks/frontend-browser.mjs`
- `checks/real-api.mjs`
- `checks/workflow.mjs`
- `compose.catalogo-test.yaml`
- `compose.test.yaml`
- `dist/assets/js/data/permissions.js`
- `dist/assets/js/data/seed.js`
- `dist/assets/js/services/document-types.js`
- `dist/assets/js/services/real-state.js`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/services/workshop-service.js`
- `dist/assets/js/views/administration.js`
- `dist/assets/js/views/clients-vehicles.js`
- `dist/assets/js/views/operations.js`
- `dist/assets/js/views/payments.js`
- `docs/backend/CATALOGO_DOCUMENTOS_TRES_ROLES.md`
- `docs/backend/ENDPOINTS.md`

Comprobación adicional del desarrollo actualizado: Chrome confirmó `/readyz` con PostgreSQL/MongoDB, login Gerente con 40 permisos, cuatro tipos documentales, tres roles disponibles, formulario de orden sin Auxiliar y cierre de sesión. Solo se realizaron operaciones de autenticación, sin nuevas escrituras de negocio. Evidencia local: `backend/.cache/catalogo-roles/dev-result.json`.
