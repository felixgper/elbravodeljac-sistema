# Empleados y tipo de empleado

Implementación en `feature/empleados-tipo-empleado`, desde la rama funcional
`backend/arquitectura-go-mongodb`. Los diagramas V19/V20 permanecen en
`docs/propuesta-modelo-v20`, commit `086f166`; no se fusionan con esta rama.

## Modelo y compatibilidad

La migración aditiva `007_empleados.sql` crea `empleados` con `id_empleado`,
`nombre` (150 caracteres), `tipo_empleado` (40 caracteres) y `activo`. Añade
`usuarios.id_empleado`, nullable, con FK restrictiva y unicidad. Un empleado
puede no tener cuenta; una cuenta se vincula como máximo a un empleado y un
empleado como máximo a una cuenta. El esquema resultante tiene 27 tablas y
48 FK, contando las migraciones funcionales anteriores.

Las cuentas anteriores conservan su ID, contraseña, estado y roles. No se
infieren empleados ni cargos a partir de nombres o roles. Para cada cuenta
humana existente, un administrador verifica la identidad, registra al empleado
y lo selecciona al editar la cuenta. Las cuentas técnicas o pendientes de
revisión pueden permanecer sin vínculo. Un vínculo ocupado se rechaza con
409 y uno existente no se reemplaza silenciosamente. No hay borrado físico de
empleados: se cambia su estado laboral.

`empleados.nombre` es el nombre laboral principal. Dos triggers mantienen
`usuarios.nombre` compatible con los lectores actuales. El nombre de una cuenta
vinculada se edita desde Empleados. Diagnósticos, recepciones, órdenes y demás
referencias continúan apuntando a los mismos usuarios; no se migran sus FK.
Un empleado sin cuenta todavía no puede recibir asignaciones técnicas que
referencian usuarios. Crear la cuenta y conceder el rol técnico correspondiente
habilita el flujo existente.

El cargo laboral es texto validado y no otorga permisos: un empleado Mecánico
puede tener roles de acceso Supervisor y Mecánico simultáneamente. Cambiarlo a
Administrador no concede acceso administrativo. Desactivar al empleado tampoco
revoca automáticamente su cuenta: para impedir el acceso se desactiva el usuario
con las protecciones administrativas existentes.

## API y seguridad

Base `/api/v1`, sesión JWT en cookie HttpOnly y CSRF para escrituras.

| Método | Ruta | Permiso |
|---|---|---|
| GET | `/empleados` | `usuarios.ver` |
| GET | `/empleados/{id}` | `usuarios.ver` |
| POST | `/empleados` | `usuarios.crear` |
| PATCH | `/empleados/{id}` | `usuarios.editar` |

Entrada: `name`, `employeeType`, `active` (booleano). Alta: nombre y cargo
obligatorios, activo por defecto. Edición parcial. Lista: paginación y `q`.
Respuesta: `id`, `name`, `employeeType`, `active`, `userId` y `username` (null
si no hay cuenta). Campos demasiado largos o vacíos: 422, sin truncar.

POST/PATCH `/usuarios` acepta `employeeId`. Vincular requiere además
`usuarios.roles`, se prohíbe cambiar el propio vínculo y se realiza en la misma
transacción que la cuenta y los roles. Una referencia inexistente devuelve 404;
un empleado inactivo, 422. Las ediciones laborales de un administrador exigen
`usuarios.roles`. Se conserva la relación N de roles y la protección del último
administrador. Los cambios de roles revocan la sesión existente y exigen entrar
de nuevo, como antes.

## Uso en pantalla

1. En modo real: **Usuarios → Empleados → Registrar empleado**.
2. Introducir nombre, tipo/cargo laboral y estado; guardar. No se crea cuenta.
3. Si necesita acceso, pulsar **Crear cuenta**, completar usuario y credenciales
   y seleccionar el rol autorizado. El nombre proviene del empleado.
4. Para una cuenta anterior, usar **Usuarios → Editar → Empleado**.
5. Los roles adicionales se administran con los controles existentes y se
   conservan las asignaciones previas.

No se añade un módulo al menú lateral ni se modifican estilos. El modo demo
permanece separado y no simula persistencia de empleados.

## Migración y pruebas aisladas

No aplicar esta migración a `ElBravoDelJac` durante las pruebas. El script
`backend/scripts/test-audit.ps1` usa exclusivamente `elbravodeljac_test` y
`elbravodeljac_tecnico_test`. Compose monta 007 para clusters de pruebas nuevos;
el script aplica únicamente 007 al cluster de navegador ya existente. No usa
restauración sobre bases reales ni elimina volúmenes. Las credenciales locales
permanecen en archivos ignorados por Git.

`TestEmployeesPostgreSQL` comprueba la migración repetida sin alterar cuentas ni
RBAC, cuentas antiguas, empleado sin cuenta, validación, permisos, vínculo único,
rollback y separación cargo/roles. Chrome comprueba el formulario y la creación
de cuenta vinculada con dos roles. La suite completa también verifica los flujos
previos del taller y MongoDB.

Resultados verificados el 29/10/2026:

- `node checks/run.mjs`: cinco suites aprobadas.
- `go test ./...` en Docker con PostgreSQL y MongoDB: aprobado, incluida
  `TestEmployeesPostgreSQL`, restricciones, transacciones, JWT/CSRF y fallos de MongoDB.
- Chrome nativo: `TestAuditBrowserIsolated` aprobado (17,01 s), empleado sin
  cuenta, cuenta vinculada, dos roles y flujo previo completo hasta entrega/PDF.
- Integración nativa de auditoría y empleados: aprobada (22,083 s en conjunto).
- `go vet ./...`: aprobado. `git diff --check`: sin errores de espacios.

La primera ejecución encontró expectativas antiguas del número de tablas y
reutilización de una sesión revocada por cambio de roles. Se corrigieron las
pruebas y el aislamiento del administrador de prueba. Chrome necesitó crear
explícitamente el rol Supervisor en su fixture; no se presupone que ya exista
en una base vacía. Las repeticiones finalizaron correctamente.

El contenedor omite deliberadamente las dos suites de navegador. La auditoría
Chrome se ejecutó después en Windows con las bases aisladas: no se interpreta
un test omitido como aprobado. Los logs de esta ejecución quedan en
`backend/.cache/audit-tests.log`, `audit-native.log` y
`employees-browser-resume.log`; la captura e informe están en
`backend/.cache/audit-browser/`. Esos archivos no se incluyen en Git.

Despliegue: aplicar 007 antes de iniciar esta versión del backend. El comando de
migración general incluye 001–007; para actualizar un entorno ya migrado se
recomienda ejecutar únicamente 007 tras confirmar el nombre de la base aislada.
La implementación no exige ejecutar migraciones en producción para revisar la rama.

Tras las pruebas se aplicó **solo 007** al proyecto aislado
`elbravodeljac-fase-e-dev`, base `elbravodeljac_dev`, y se reconstruyó el servicio
`api`. Antes y después se compararon recuentos y hashes de todas las filas de
las tablas anteriores, excluyendo únicamente el nuevo atributo `id_empleado`:
coinciden. No se cargaron empleados de ejemplo en desarrollo ni se ejecutaron
migraciones en la base real. MongoDB de desarrollo no recibió escrituras de prueba.

Aplicación actualizada: `http://localhost:8080/`. Se verificaron `/readyz`
(PostgreSQL y MongoDB), inicio de sesión, GET empleados autorizado y cierre de
sesión con CSRF. Se utiliza la cuenta de desarrollo existente; el acceso privado
está en `backend/.cache/acceso-desarrollo.txt`, ignorado por Git. Para registrar
un empleado abrir **Usuarios y roles → Empleados → Registrar empleado**.

## Archivos de la implementación

- Migración: `backend/internal/platform/postgres/migrations/007_empleados.sql`.
- Backend nuevo: `internal/domain/employee.go`,
  `internal/httpapi/employees.go`, `internal/platform/postgres/employees.go`
  (rutas relativas a `backend/`).
- Backend adaptado: `domain/relational.go`, `httpapi/relational.go`,
  `platform/postgres/identity.go`, `migrate.go` y `relational.go`.
- Frontend: nueva vista `dist/assets/js/views/employees.js`, adaptación de
  `views/administration.js`, `services/real-state.js` y `services/real-workshop.js`.
- Pruebas nuevas: `backend/internal/domain/employee_test.go` y
  `backend/integration/employees_test.go`.
- Pruebas adaptadas: `backend/integration/audit_browser_test.go`,
  `brands_test.go`, `phase_a_test.go`, `phase_c_test.go`, `review_test.go`,
  `workshop_test.go`; conteo de rutas en `backend/internal/httpapi/catalog_test.go`.
- Entorno y navegador: `compose.test.yaml`, `backend/scripts/test-audit.ps1`
  y `checks/audit-browser.mjs`.
- Documentación: `README.md`, este informe, `docs/backend/ENDPOINTS.md`
  y `docs/backend/ESTADO_ENDPOINTS.md`.

No se modifican los diagramas V19/V20 ni se hace commit, push o merge.
