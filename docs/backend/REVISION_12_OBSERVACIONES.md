# Revisión de 12 observaciones

> Corrección posterior de marcas: la solución inicial basada solamente en vehículos
> registrados resultó insuficiente. [La revisión específica](CORRECCION_MARCAS.md)
> incorpora las 34 marcas aprobadas en PostgreSQL, selector buscable y Otra marca.
> Las afirmaciones históricas sobre marcas que siguen describen la entrega anterior.

Base limpia: `0b2d61e54bce5dcf65ad0096baee11cee814eeee`.
Recuperación: `respaldo/antes-observaciones-12-20261027`. Se conserva la rama actual.

## Matriz de diagnóstico previa a la implementación

| Nº | Hallazgo | Archivos/módulos | Solución prevista |
|---|---|---|---|
| 1 | La hidratación recorre todas las páginas de clientes; selects nativos sin búsqueda remota; ILIKE distingue tildes | real-workshop, ui, vistas, store.go | Autocompletado remoto paginado con debounce/cancelación y consulta normalizada indexada; directorio paginado en servidor |
| 2 | Solo existe el literal JAC; no hay tabla ni catálogo separado de marcas en este repositorio | clients-vehicles, recursos Go | Un único origen: marcas distintas de vehículos existentes como sugerencias; permitir escribir otra sin alterar registros |
| 3 | Recepción sin guardar muestra «Hoy: fecha y hora al guardar» | ui, operations | Fecha de negocio dinámica Hoy — dd/mm/aaaa, America/Lima; timestamp definitivo del servidor |
| 4 | CHECK exige exactamente una FK; no hay tipo Original/Alternativo ni descuento persistido | domain/workshop, workshop_writes/reads, migración, editor de cotización | Líneas manuales explícitas sin FK ficticias; descripción/precio/cantidad propios; referencias opcionales a catálogos; tipo y descuento por línea |
| 5 | Las opciones dependen de la instantánea global de vehículos | real-workshop, recepciones/auxilios | Consultar páginas por clientId y reconstruir las opciones; validar pertenencia también en Go |
| 6 | Ya existe rowActions con lápiz/papelera y confirmDelete | ui y vistas | Reutilizar patrón; verificar permisos, errores y FK; distinguir quitar línea no guardada de eliminar registro persistido |
| 7 | Existe unidades_auxilio y GET/POST; el formulario no permite registrar unidades y no explica catálogo vacío | operations, servicios, catálogos Go | Mensaje de catálogo vacío y alta real autorizada, sin sembrar unidades ficticias |
| 8 | Prueba opcional para avances; última prueba satisfactoria exigida para finalizar | formulario OT y validaciones Go | Explicar condición; guardar avance sin prueba; al finalizar usar prueba satisfactoria vigente o pedirla explícitamente |
| 9 | Usuarios ↔ roles N; asignaciones_ot distingue MECANICO/AUXILIAR | administration, operations, identity | Centralizar alta en Usuarios, acceso desde Mecánicos, explicar roles y funciones; preservar N y permisos efectivos |
| 10 (numeración inicial) | Campo sin identificar; aclaración posterior del usuario lo denomina observación 11 | Formularios auditados abajo | Se conserva; pendiente de identificar el control concreto |
| 11 | Dashboard sin filtros; seis meses fijos y varias métricas globales | app, dashboard-service, statsRange/statistics | Semana/Mes/Año con límites Lima semiabiertos; semana lunes; filtrar todos los indicadores y series |
| 12 | Controles y datos deben probarse juntos | suites Go/Node/Chrome/Docker | Regresión y pruebas HTTP/SQL, navegador y persistencia, con bases separadas |

## Justificación de migración

La observación 4 modifica la decisión anterior de exigir siempre catálogo. No es
posible guardar tipo de repuesto ni descuento transparente sin nuevos campos.
Se agregaron a detalle_cotizacion `tipo_concepto`, `tipo_repuesto` y
`descuento_unitario` (cero para registros anteriores). Se reemplazó solamente el
CHECK de referencia por otro que conserva la exclusión de ambas FK y admite una
línea sin referencias únicamente con tipo y descripción explícitos. Ninguna FK se
retira: siguen siendo 26 tablas y 42 FK. No se altera ni recalcula ninguna cotización
anterior. Origen TALLER/CLIENTE se conserva como «Suministrado por», independiente
de Original/Alternativo. Precio y descuento son snapshots de cada línea.

La búsqueda requiere normalización sin tildes e índices; se incorporan unaccent,
pg_trgm y una función SQL de normalización para las consultas. Las migraciones son
transaccionales e idempotentes, con comando explícito limitado a bases _dev/_test.
No se aplican a ElBravoDelJac ni se reinicializan volúmenes.

## Resultado de la implementación — 27/10/2026

La aclaración del usuario denomina **10 al registro de mecánicos/auxiliares**
y **11 al campo no identificado**. Se conserva la matriz original como trazabilidad;
los filtros del dashboard también se implementaron. No se declaran resueltas las
12 observaciones: queda pendiente identificar el campo exacto.

| Tema | Resultado comprobado |
|---|---|
| Búsqueda | Autocompletado remoto por nombre/documento, sin tildes ni diferencias de mayúsculas, debounce de 250 ms, cancelación de consultas antiguas, páginas de 20 y mensaje sin coincidencias. El directorio también pagina en PostgreSQL. No descarga todos los clientes. |
| Marcas | Se eliminó la restricción visual a JAC. El repositorio no contiene una tabla independiente de marcas: `GET /api/v1/vehiculos/marcas` devuelve las ya registradas como sugerencias y permite introducir otra (máximo 40). No se inventó otro catálogo ni se modificaron vehículos anteriores. |
| Fecha | Recepción nueva muestra «Hoy — dd/mm/aaaa» calculado en America/Lima. La fecha persistida sigue generándola Go/PostgreSQL. |
| Cotizaciones | Conceptos manuales con tipo y descripción explícitos; catálogo opcional con FK validada cuando se usa. Cantidad, precio de venta, descuento por unidad, subtotal y total exactos. Original/Alternativo es independiente de Suministrado por Taller/Cliente. Los servicios no exigen tipo de repuesto. Los repuestos aportados por cliente conservan precio cero. |
| Vehículos del cliente | Se consultan todas las páginas del propietario seleccionado, se invalida la selección anterior y Go vuelve a comprobar la pertenencia. Prueba de 105 vehículos en un propietario. El enlace de alta conserva el cliente seleccionado. |
| Editar/eliminar | Se conserva el patrón existente lápiz/papelera, ayudas y confirmación. Los errores 403/409 permanecen visibles y no producen mensajes de éxito. Eliminar una línea sin guardar no equivale a borrar un registro histórico. |
| Auxilio | Unidad es una fila de `unidades_auxilio`, no el trabajador. Consulta real y mensaje cuando no existen. Alta mediante ventana autorizada usando POST existente, seleccionando el ID devuelto; no se siembran unidades ficticias. |
| Pruebas de reparación | Se permite guardar avances sin prueba. La pantalla explica que finalizar exige una prueba satisfactoria posterior a la reparación; puede reutilizar la última válida. Sin ella devuelve 422 y no muestra éxito. |
| Mecánicos y auxiliares (10 aclarada) | Alta centralizada en Usuarios, con enlaces desde Mecánicos y rol preseleccionado. Usuarios↔roles sigue siendo N: añadir Supervisor no borra Mecánico. Responsable y auxiliar se seleccionan por separado en la OT; una misma persona no puede ocupar ambos puestos simultáneamente. Guardar avances sin cambiar personal no genera asignaciones duplicadas. |
| Dashboard/reportes | Botones y selector exactamente Semana/Mes/Año. Semana de lunes a lunes, mes natural y año natural en Lima; final exclusivo. Indicadores y gráficos usan el período. Ingresos se agregan desde pagos sin multiplicarlos al unir detalles. |
| Consistencia | Se conservaron diseño, menú y estilos; solo se añadieron el desplegable y los controles necesarios. Recepciones/órdenes/auxilios incluyen etiquetas de contexto reales para no depender de una descarga global de clientes. Las listas recientes muestran los nuevos registros y los borradores de recepción conservan kilometraje al recargar. |

### Registro de un trabajador

1. Usuarios y roles → Nuevo usuario (o Mecánicos → Registrar mecánico/auxiliar).
2. Completar identificación de acceso y asignar el rol correspondiente. Solo usuarios
   con permisos de creación/asignación pueden hacerlo. Los permisos efectivos siguen
   siendo la unión de roles; el rol técnico no implica acceso administrativo.
3. Crear o editar una OT autorizada y seleccionar Mecánico responsable y Auxiliar.
   Estas funciones se guardan en `asignaciones_ot`, sin duplicar trabajadores.

### Auditoría del campo no identificado (11 aclarada)

No es posible atribuir el problema a un control concreto sin identificar pantalla
u operación. No se eliminó ningún campo por suposición. Se verificaron estos casos:

| Pantalla/control | Finalidad y motivo de bloqueo |
|---|---|
| Recepción: fecha | Generada por el sistema. Recepciones ya guardadas son de consulta. |
| Diagnóstico: cliente, vehículo, síntomas | Proceden de la recepción; no son una segunda edición del propietario. Un diagnóstico persistido no ofrece modificación en el contrato vigente. |
| Cotización: vehículo | Derivado del diagnóstico. Si ya existe OT, conceptos/autorización quedan bloqueados para preservar su importe histórico. Tipo de repuesto no aplica a servicios. |
| OT: código, contexto, fecha | Generados o derivados. Asignaciones exigen `ordenes.asignar`; una orden finalizada, cancelada o entregada bloquea edición según sus reglas. |
| OT: Prueba realizada | Opcional al guardar avances; condición de finalización explicada y probada. |
| Pagos/entrega | Exigen estado finalizado, saldo/importe válidos y permisos. Entregar requiere saldo cero y contacto receptor válido. |
| Proveedores: repuestos vinculados | Consulta de la relación existente; se aclaró el texto informativo. No se convirtió en texto libre ni se agregó especialidad. |
| Usuarios/matriz | Propietario, autoasignación y último acceso administrativo conservan bloqueos de seguridad. Asignar roles exige permiso específico. |

La identificación del campo pendiente requiere una futura indicación de pantalla
(y, si es posible, texto visible). No bloquea las correcciones ya entregadas.

## Cambios de base de datos y seguridad

Migración `backend/internal/platform/postgres/migrations/001_revision_funcional.sql`,
embebida por Go y aplicada con `cmd/db-migrate`. Transacción y bloqueo asesor;
ejecución repetible. `start-dev.ps1` la ejecuta antes de arrancar la API y no cambia
cuentas existentes. DSN restringido a bases de desarrollo/pruebas `_dev`/`_test`.

- Tres columnas nuevas en detalle_cotizacion (tipo, tipo de repuesto, descuento).
- Sustitución justificada de un CHECK de referencia; ambas FK originales continúan.
- CHECK de descuento y tipo; extensiones unaccent/pg_trgm, función e índice GIN.
- Se conservan **26 tablas y 42 FK**. No se modificó el SQL exportado original.
- Tipos históricos sin clasificación permanecen sin clasificar; descuentos previos
  son cero y no se recalculan importes existentes.
- Antes/después de migrar desarrollo coincidieron: 2 clientes, 3 vehículos,
  2 cotizaciones (140.00), 2 órdenes y pagos 140.00. Después se añadieron únicamente
  registros marcados de revisión mediante las pruebas funcionales.
- Sin DROP TABLE, TRUNCATE, eliminación de volúmenes, restauración sobre bases reales
  ni conexión a ElBravoDelJac en 5432. El proyecto comercial ajeno no se reinició.
- Credenciales, capturas e informes con datos locales están excluidos de Git.

## Pruebas realmente ejecutadas

| Prueba | Resultado |
|---|---|
| `node checks/run.mjs` | Cuatro suites aprobadas: demo, flujo, límites y contratos de servicios reales. Estas pruebas Node no sustituyen la integración con BD. |
| `go test ./... -count=1 -v` en Compose tests | Aprobado con PostgreSQL 18.6 y MongoDB 9.0 temporales: fases A–D, autenticación/CSRF, permisos, transacciones, FK, concurrencia, último administrador, GridFS, reintentos y fallos MongoDB. |
| `TestFunctionalReviewPostgreSQL` | Aprobado: 1500 clientes, normalización/paginación/índice, 105 vehículos, marcas, propietario incorrecto, conceptos manuales, descuento, precio independiente del catálogo, migración repetible, asignaciones sin duplicados, pruebas/finalización y comparación SQL de ingresos/estados por período. Sigue habiendo 26 tablas/42 FK. |
| `go vet ./...` en imagen tests | Aprobado, salida 0. |
| `node checks/review-browser.mjs` en Chrome Windows | Aprobado contra Go, PostgreSQL y MongoDB del entorno dev: búsqueda F, dos vehículos del cliente, F5/borrador, cotización manual 305.00, roles N, asignaciones, avance sin prueba, error 422 sin éxito, unidad/auxilio, tres períodos y permiso 403. Sin errores JavaScript. |
| `node checks/read-persisted-docker.mjs` | Aprobado antes del nuevo recorrido: registros originales cliente/OT/pago/entrega/escáner/evidencia seguían consultables tras desplegar la actualización. En esta revisión no se reiniciaron PostgreSQL/MongoDB. |
| `node checks/live-docker.mjs` en Chrome Windows | Aprobado de nuevo: cliente → vehículo → recepción → diagnóstico → cotización → OT → reparación/prueba → pago → entrega; consultas tras recarga, MongoDB/GridFS, idempotencia, CSRF y permisos. |
| Revisión de capturas | Cotización y dashboard Semana conservan el estilo, menú y componentes existentes a 1440 px. |
| `git diff --check` | Sin errores; Git avisa únicamente de conversión LF/CRLF en archivos Go de Windows. |

El test opcional `TestPhaseEFrontendBrowserPostgreSQLMongoDB` aparece **SKIP** en
el contenedor porque no se habilitó `FRONTEND_E2E`/runtime Node allí. No se presenta
como aprobado: el navegador se ejecutó por separado en Chrome Windows, con las dos
suites descritas. La primera consulta desde el sandbox falló por EACCES al socket
localhost; se repitió fuera del sandbox y pasó. No es un fallo de la aplicación.

Informes locales ignorados: `backend/.cache/review-browser/resultado.json`,
`backend/.cache/live-docker/resultado.json`; capturas en la primera carpeta.
Los datos masivos solo se insertaron en PostgreSQL tmpfs de pruebas; las pruebas
Chrome dejan unos pocos registros identificados en la base dev para consultarlos.

## Ejecución exacta en Windows

Desde PowerShell, en la raíz del repositorio y con Docker Desktop en Engine running:

```powershell
.\backend\scripts\start-dev.ps1
```

No usar SkipBuild al actualizar esta versión. El script usa exclusivamente el
proyecto `elbravodeljac-fase-e-dev`, crea/levanta sus servicios, aplica la migración
y sirve frontend y API en el mismo origen. Abrir **http://localhost:8080/** en Chrome.
Usuario inicial `admin_pruebas`; contraseña privada en
`backend/.cache/acceso-desarrollo.txt` (no cambia si la base ya tiene usuarios).
No publicar `.env.fase-e.local` ni el archivo de acceso. `/demo/` sigue separado.
PostgreSQL y MongoDB de desarrollo no publican puertos al host.

Repetir integración aislada desde esa misma raíz:

```powershell
docker compose --env-file .env.fase-e.local -p elbravodeljac-fase-e-tests -f compose.test.yaml up --build --abort-on-container-exit --exit-code-from tests --attach tests
docker compose --env-file .env.fase-e.local -p elbravodeljac-fase-e-tests -f compose.test.yaml run --rm --no-deps tests go vet ./...
node checks/run.mjs
$env:PLAYWRIGHT_MODULE = "C:\Users\Usuario\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules\playwright"
# Si esa ruta no existe, indicar la ruta absoluta al paquete Playwright instalado.
node checks/review-browser.mjs
node checks/live-docker.mjs
```

Para esta computadora, la ruta comprobada a Playwright es
`C:\Users\Usuario\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules\playwright`.
Chrome se detecta en Program Files/Google/Chrome/Application/chrome.exe.
No usar `down -v` ni apuntar estos comandos a otros proyectos/bases.

## Estado de Git y archivos

Rama conservada: `backend/arquitectura-go-mongodb`. Sin merge con main ni push.
No se creó un commit nuevo. Punto de recuperación previo:
`respaldo/antes-observaciones-12-20261027`, commit
`0b2d61e54bce5dcf65ad0096baee11cee814eeee`.
Los cambios quedan disponibles para revisión y posterior commit.

## Auditoría previa al commit — 27/10/2026

La cifra correcta es **42 archivos: 31 modificados y 11 nuevos**. Las 12
observaciones son temas funcionales, no un número de archivos. GitHub Desktop
incluye también archivos nuevos que `git diff --stat` aún no muestra.

Clasificación: backend 22 (incluye pruebas/scripts), checks 4, frontend dist 10,
documentación docs 3, README raíz 1 y Compose 2. Todos pertenecen a esta revisión;
no se detectaron cambios de otro proyecto, archivos generados ni cambios ajenos
al alcance. Se revisaron los diffs y el contenido de los archivos nuevos.

La comprobación de los 42 archivos no encontró valores de las contraseñas/clave
JWT del entorno privado, claves privadas, tokens comunes ni URIs con credenciales
literales. Las conexiones Compose usan variables. No hay temporales, logs,
capturas, ejecutables, respaldos ni carpetas de dependencias entre los cambios.
`.env.fase-e.local` y `backend/.cache/` siguen ignorados. El SQL exportado original,
la página comercial y `.gitignore` no cambiaron. Las credenciales demo existentes
no son credenciales del entorno real y no se añadieron secretos nuevos.

**Pruebas repetidas en esta auditoría:** cuatro suites Node, Go completo con
PostgreSQL/MongoDB aislados, go vet, Chrome review-browser y live-docker: aprobadas,
salida 0. El navegador opcional dentro del contenedor continúa SKIP por falta de
habilitación/runtime allí; Chrome Windows sí se ejecutó, no se simularon resultados.
La lectura de los datos previos con read-persisted-docker también pasó antes de
crear los nuevos registros de revisión. No se reiniciaron las BD de desarrollo.

Consulta directa final (solo lectura): PostgreSQL `elbravodeljac_dev` conserva
26 tablas/42 FK, 9 clientes, 13 vehículos, 8 cotizaciones, 8 órdenes, 4 pagos y
3 entregas. OT-1 continúa finalizada, con pago 120.00 y entrega existente.
MongoDB `elbravodeljac_tecnico_dev` contiene 3 escáneres, 3 evidencias listas,
3 archivos GridFS y 3 bloques. El escáner y evidencia originales de la primera
verificación siguen presentes. `/readyz` devuelve ready y technicalConfigured=true.
Los nuevos registros pertenecen únicamente a las pruebas funcionales en desarrollo.
No se conectó al PostgreSQL original ni se alteraron volúmenes ajenos.

**Observaciones:** se comprobaron búsqueda, marcas, fecha, conceptos/descuentos,
vehículos, acciones autorizadas, unidad de auxilio, prueba opcional/obligatoria,
registro/asignación de trabajadores y períodos. La revisión integral pasó para
los recorridos probados. Sigue pendiente el campo desconocido (11 según la última
aclaración; 10 en el texto original). El registro de trabajadores corresponde a
9 original/10 aclarada; dashboard es 11 original. No se afirma que el campo
pendiente esté corregido ni se eliminó por suposición.

**Conclusión:** estos 42 archivos pueden incluirse en un commit de correcciones
funcionales, con el pendiente documentado. No se realizó commit, push ni merge;
la rama permanece backend/arquitectura-go-mongodb. Resumen sugerido:
`Corrige búsqueda, cotizaciones, trabajadores, auxilio y períodos del taller`.

Inventario auditado, archivo por archivo:

| Archivo | Estado | Motivo |
|---|---|---|
| `README.md` | Modificado | Enlace a la revisión y períodos actuales de reportes. |
| `backend/README.md` | Modificado | Documenta la migración y el nuevo contrato de cotizaciones. |
| `backend/cmd/db-migrate/main.go` | Nuevo | Comando explícito para migrar únicamente bases de desarrollo/pruebas. |
| `backend/integration/review_test.go` | Nuevo | Pruebas HTTP/SQL con 1500 clientes, 105 vehículos, cotizaciones, asignaciones y períodos. |
| `backend/internal/domain/manual_quote_test.go` | Nuevo | Valida conceptos manuales, descuento y tipo de repuesto. |
| `backend/internal/domain/models.go` | Modificado | Incluye resúmenes de vehículos del cliente de la página consultada. |
| `backend/internal/domain/workshop.go` | Modificado | Valida conceptos manuales, importes exactos y personas distintas en asignaciones. |
| `backend/internal/httpapi/brands.go` | Nuevo | Handler protegido para consultar marcas existentes. |
| `backend/internal/httpapi/catalog_test.go` | Modificado | Actualiza conteo por la nueva ruta sin quitar cobertura de las 71 originales. |
| `backend/internal/httpapi/periods_test.go` | Nuevo | Comprueba límites Semana/Mes/Año en Lima, incluso cambio de año UTC. |
| `backend/internal/httpapi/relational.go` | Modificado | Calcula períodos reales y conserva compatibilidad de consultas anteriores. |
| `backend/internal/httpapi/resources.go` | Modificado | Registra GET vehiculos/marcas con permiso vehiculos.ver. |
| `backend/internal/platform/postgres/brands.go` | Nuevo | Obtiene marcas distintas de vehículos persistidos. |
| `backend/internal/platform/postgres/frontend.go` | Modificado | Búsqueda/filtrado y paginación de opciones con alcance autorizado. |
| `backend/internal/platform/postgres/migrate.go` | Nuevo | Incluye la migración SQL en el binario Go. |
| `backend/internal/platform/postgres/migrations/001_revision_funcional.sql` | Nuevo | Migración transaccional: tres campos, CHECK, normalización e índice; conserva tablas/FK. |
| `backend/internal/platform/postgres/relational.go` | Modificado | Añade etiquetas reales a respuestas de auxilio sin cargar todos los clientes. |
| `backend/internal/platform/postgres/statistics.go` | Modificado | Aplica períodos a métricas, incluye servicios manuales y evita multiplicar ingresos. |
| `backend/internal/platform/postgres/store.go` | Modificado | Búsqueda normalizada y página de clientes con vehículos relacionados. |
| `backend/internal/platform/postgres/workshop_reads.go` | Modificado | Lee tipo/descuento de cotización y contexto real de recepciones/órdenes. |
| `backend/internal/platform/postgres/workshop_writes.go` | Modificado | Persiste conceptos/descuentos y evita asignaciones redundantes. |
| `backend/scripts/start-dev.ps1` | Modificado | Construye/ejecuta migrador antes de iniciar API en desarrollo aislado. |
| `backend/scripts/test-local-postgres.ps1` | Modificado | Aplica migración en la base nueva de prueba del script alternativo. |
| `checks/frontend-browser.mjs` | Modificado | Adapta el navegador de integración al buscador y a los períodos actuales. |
| `checks/live-docker.mjs` | Modificado | Selecciona clientes con el control real durante el recorrido Docker. |
| `checks/real-api.mjs` | Modificado | Verifica DTO manual, descuento y nuevo rango del dashboard. |
| `checks/review-browser.mjs` | Nuevo | Prueba las correcciones en Chrome y conserva informes solo en .cache ignorado. |
| `compose.dev.yaml` | Modificado | Servicio de migración de desarrollo sin publicar puertos de BD. |
| `compose.test.yaml` | Modificado | Aplica migración al iniciar PostgreSQL temporal de pruebas. |
| `dist/assets/css/atelier.css` | Modificado | Nueve líneas para desplegable de búsqueda; conserva el estilo existente. |
| `dist/assets/js/app.js` | Modificado | Conecta filtros del dashboard y contexto de carga de formularios. |
| `dist/assets/js/components/remote-picker.js` | Nuevo | Buscador remoto con ID real, debounce, cancelación y vehículos del propietario. |
| `dist/assets/js/components/ui.js` | Modificado | Listas remotas paginadas y fecha automática en Lima. |
| `dist/assets/js/services/dashboard-service.js` | Modificado | Separa cálculos demo y consulta real por período. |
| `dist/assets/js/services/real-workshop.js` | Modificado | Adapta DTO, opciones paginadas, marcas, unidades y evita descarga global de clientes. |
| `dist/assets/js/services/workshop-service.js` | Modificado | Compatibilidad demo de descuentos y contexto con etiquetas reales. |
| `dist/assets/js/views/administration.js` | Modificado | Alta de trabajadores desde Mecánicos/Usuarios, explicación de roles y períodos. |
| `dist/assets/js/views/clients-vehicles.js` | Modificado | Directorios paginados, buscadores, marcas libres e historial existente. |
| `dist/assets/js/views/operations.js` | Modificado | Conecta vehículos, fecha, cotización flexible, reglas de prueba y alta de unidad. |
| `docs/backend/ESTADO_ENDPOINTS.md` | Modificado | Documenta marcas, períodos y contrato actualizado de cotizaciones. |
| `docs/backend/FASE_E.md` | Modificado | Distingue las limitaciones históricas de los cambios posteriores. |
| `docs/backend/REVISION_12_OBSERVACIONES.md` | Nuevo | Matriz, migración, resultados, campo pendiente e inventario auditado de los 42 archivos. |
