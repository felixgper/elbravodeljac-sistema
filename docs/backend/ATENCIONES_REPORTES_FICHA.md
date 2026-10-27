# Atenciones, ingresos y ficha técnica — 27/10/2026

## Alcance implementado

Se conserva la aplicación administrativa, su menú y diseño. No se añade módulo
lateral. Se parte del árbol limpio en backend/arquitectura-go-mongodb; no se hace
commit, push ni merge. La base original ElBravoDelJac y otros proyectos Docker
no se utilizan ni se reinicializan. PostgreSQL conserva la operación transaccional;
MongoDB continúa para escáner/evidencias, sin duplicar clientes ni pagos.

### Recepciones y dos recorridos

`attentionType` admite `Directo` (Servicio directo, sin diagnóstico) o `Diagnostico`
(Requiere diagnóstico técnico). Las recepciones anteriores conservan Diagnostico.
El selector y el enlace posterior guían a Cotización o a Diagnóstico según el caso.

Ambos recorridos exigen cotización y aprobación antes de crear OT. Una recepción
Directo no puede recibir un diagnóstico; una Diagnostico no puede cotizar sin uno
real de esa misma recepción. El backend y triggers PostgreSQL verifican el origen.
No se insertan diagnósticos ficticios ni se cambia de origen una cotización guardada.
El tipo de una recepción con historial no puede cambiar por SQL accidental.

La API de diagnóstico admite `technicianId` explícito. La interfaz exige seleccionar
personal técnico activo; un trabajador sin ordenes.asignar solo puede registrarse
como responsable a sí mismo. Asignar otro exige permiso de asignación y rol técnico
compatible. Se preserva el cliente anterior de la API: si omite technicianId,
se registra el trabajador autenticado con permiso diagnosticos.crear, nunca un
nombre ficticio ni un usuario genérico. Fechas y autoría se conservan en PostgreSQL.

Contratos actualizados:

- POST /api/v1/recepciones: añade attentionType; al omitirse conserva Diagnostico.
- POST /api/v1/diagnosticos: añade technicianId, con validación de ID/autoridad.
- POST /api/v1/cotizaciones y PUT /api/v1/cotizaciones/{id}: receptionId permite
  el origen directo; diagnosisId real sigue siendo obligatorio para la atención
  técnica. Si solo se envía diagnosisId, Go obtiene su recepción para compatibilidad.
- Consultas de cotización/recepción/orden e historial incluyen el contexto directo.
  Permisos, referencias a catálogo y precios autorizados anteriores se mantienen.

### Reportes reales

La tabla Ingresos registrados y el gráfico **Evolución de ingresos**, colocado
inmediatamente debajo, usan exactamente los mismos puntos de la misma respuesta.
Semana/Mes agrupan pagos diarios; Año agrupa pagos mensuales. Se usan intervalos
semiabiertos y America/Lima, con semana desde lunes. Los buckets vacíos son cero.
La suma de los puntos coincide con incomeCents y con la consulta SQL independiente
a pagos. No se consideran cotizaciones pendientes ni se unen detalles a pagos.

Se reutiliza incomeChart borgoña, con SVG responsive, escala según ingresos reales,
valores monetarios con centavos y etiquetas espaciadas para 28–31 días. La tabla
conserva cada día completo; el gráfico omite solo algunas etiquetas del eje, no
puntos de ingreso. Se corrige asimismo el caso de una sola coordenada.

### Control técnico durante la orden

En Actualizar orden aparece **Control técnico · Ficha**. La ventana conserva las
categorías, 21 elementos de inspección, mantenimiento y 14 opciones de otros
servicios del PDF. Hay selección Bien/Atención/Urgente/Sin evaluar, PSI por rueda,
otros líquidos, bobinas, observaciones, kilometraje y fecha de próximo mantenimiento.
Las opciones de mantenimiento reproducen aceite, marca, filtros, bujías,
refrigerante, retén, mangueras, pastillas y fajas del original.

El mecánico confirma aparte qué líneas de la cotización realmente ejecutó. Solo
se aceptan IDs de detalles pertenecientes a su OT. No se presupone que un concepto
cotizado o un ejemplo del PDF fue ejecutado. Sin confirmación aparece esa ausencia;
no se inventan trabajos, inspecciones ni firmas.

GET /api/v1/ordenes/{id}/control-tecnico requiere ordenes.ver y alcance de consulta.
PUT en la misma ruta requiere ordenes.editar y asignación activa como MECANICO,
o autoridad ordenes.asignar. Guardado transaccional, autor FK a usuarios, fecha y
versión optimista. Una versión antigua devuelve 409; una orden finalizada/cancelada
o entregada es de consulta. El pago no rellena controles ni modifica inspecciones.

### Ficha y descargas posteriores

**Pagos y entrega → Atención finalizada → Exportar ficha técnica (PDF)** aparece
cuando hay entrega confirmada y saldo cero. El historial de vehículo incluye
Descargar ficha PDF para las atenciones entregadas, con permiso pagos.ver.
La API también comprueba estado finalizado, cotización aprobada, pago exacto y
entrega: no basta con habilitar el botón en JavaScript.

GET /api/v1/ordenes/{id}/ficha-tecnica.pdf devuelve application/pdf, con sesión
HttpOnly y permiso pagos.ver. Una atención aún no entregada devuelve 404, sin PDF.

Se utiliza una plantilla saneada del documento aportado: mismo fondo, logo,
colores, bloques, categorías, casillas, firmas y dos páginas A4. Los fondos se
prepararon a 300 ppp, retirando píxeles y OCR de los datos personales y selecciones
de ejemplo. No se incluye el PDF personal original en Git. El PDF final es estático,
imprimible y sin firmas inventadas. Los textos variables se dibujan sobre los
espacios originales; la tabla mantiene su cabecera y franjas.

Los datos del taller se leen desde la vista PostgreSQL datos_taller_ficha (nombre,
RUC, dirección aprobados del documento); logo/marca gráfica permanecen como diseño.
Cliente, documento, contacto, vehículo, kilometraje de recepción, motivo, fechas,
mecánico, trabajos confirmados, conceptos/precios/descuentos e inspecciones salen
de esa atención. La plantilla de producción no conserva datos personales de ejemplo.

Al entregar se congela ficha_snapshot y se genera/guarda ficha_pdf en esa misma
transacción. Son datos de PostgreSQL: la ficha no depende de localStorage ni de
estado temporal del navegador. Las descargas posteriores devuelven los mismos
bytes aunque luego cambie el nombre actual del cliente. Entregas anteriores a la
migración no reciben inspecciones inventadas: al primer pedido se reconstruye una
vez con los datos disponibles y se conserva. No puede garantizarse un dato histórico
que nunca se almacenó, por lo que la ausencia queda explícita.

Para respetar el original de tres casillas de inspección, Sin evaluar se imprime
sin marca, con leyenda expresa. Las opciones de mantenimiento no seleccionadas
permanecen vacías. La conformidad de cliente y operario mantiene su espacio en blanco.

## Migración 003_atencion_ficha.sql

Transaccional, repetible y con bloqueo asesor. No borra tablas, datos ni FK originales:

- recepciones: tipo_atencion (default Diagnostico para registros previos).
- cotizaciones: id_recepcion con FK, backfill desde su diagnóstico existente y NOT
  NULL; id_diagnostico admite NULL solo para origen directo validado por trigger.
- ordenes_trabajo: control_tecnico JSONB, versión, autor FK y fecha.
- entregas: ficha_snapshot JSONB y ficha_pdf bytea.
- Índice de recepción, triggers de coherencia/historial y vista de datos del taller.

Se mantienen **26 tablas y las 42 FK originales**, y se añaden **2 FK**, total **44**.
Las pruebas comprueban los 42 nombres originales individualmente. Se actualizó
la guarda del bootstrap a la evolución aprobada; no se restablecen usuarios/roles.
El SQL exportado original no se modifica. Compose y el script local alternativo
aplican 001, 002 y 003 en ese orden, solamente en las bases _dev/_test autorizadas.

En desarrollo, antes/después de migrar coincidieron: 15 clientes, 23 vehículos,
12 recepciones, 11 diagnósticos, 10 cotizaciones, total cotizaciones 2025.00,
pagos 500.00 y 4 entregas. Los registros añadidos después son pruebas identificadas.
Los archivos de comparación están en backend/.cache/ficha/antes.json y despues.json,
excluidos de Git. No se recreó el PostgreSQL persistente ni se reiniciaron sus datos.

## Pruebas y resultados

- go test ./... -count=1 -v en Compose aislado: aprobado. Incluye PostgreSQL 18.6 y
  MongoDB 9.0 temporales, regresión A–D, marcas y revisión funcional anterior.
- TestAttentionInspectionAndOriginalSheet: aprobado en ambos recorridos, incluyendo
  cotización rechazada antes de diagnóstico/aprobación, ausencia de diagnóstico
  ficticio, mecánico responsable, permisos 403/404, control persistido, versión
  antigua 409, cierre de control, rechazo de entrega sin pago, generación PDF al
  entregar, dos páginas, repetición con bytes idénticos tras cambiar cliente,
  historial, conciliación SQL y las 42 FK originales conservadas. Se comprobó también que un motivo extenso no
  recorta el snapshot ni pierde la entrega: sheetReady=false y descarga 422.
- TestControlValidation y TestOriginalTwoPagePDF: aprobados. Rechazo de opciones
  desconocidas, PSI inválido, textos excesivos y fechas inválidas; sin recorte.
- go vet ./... en la imagen Docker: aprobado. También pasaron go test ./... y
  go vet ./... nativos de Windows sobre el código final. En la ejecución nativa
  no se configuraron DSN de integración: esos casos se omiten allí; las pruebas
  reales PostgreSQL/MongoDB son las ejecutadas y aprobadas en Compose, no ese SKIP.
  Se probó asimismo el ID de mecánico y la plantilla con los campos destacados.
- node checks/run.mjs: las cuatro suites aprobadas (lógica/contratos, no BD simulada
  presentada como integración).
- node checks/attention-browser.mjs en Chrome Windows: aprobado. Usuario mecánico
  real en sesión separada, guardado y F5, dos flujos, pago/entrega, descarga desde
  Pagos e historial con igualdad de bytes y tres filtros con gráfico conciliado.
- node checks/live-docker.mjs: regresión del recorrido con CSRF/permisos y
  MongoDB/GridFS aprobada. El navegador opcional dentro del contenedor Go permanece
  SKIP: Chrome se ejecuta por separado con Playwright en Windows.
- Renderizado con Poppler y revisión visual de ambas páginas frente al original:
  dos páginas A4 595.5 × 842.25 puntos; sin restos de datos personales de ejemplo,
  firmas ficticias ni casillas seleccionadas por defecto. Se corrigieron máscaras,
  alineación de detalle, texto de cabecera y coordenadas antes de verificar de nuevo.

Informes y PDF de atenciones de prueba: backend/.cache/attention-browser/.
No se publican informes locales, credenciales ni los PDF que contienen datos de prueba.
Se verificaron además los tres períodos a 1440 y 390 px: sin etiquetas superpuestas
ni desbordamiento del gráfico. Se volvieron a consultar los registros persistidos
PostgreSQL/MongoDB tras desplegar Go. Una primera lectura inmediata no llegó al
panel por la ventana de inicios de sesión consecutivos de las pruebas; se repitió
sin modificar el limitador y pasó. Las bases de datos no se reiniciaron para ello.

## Límites explícitos de la plantilla

Se respeta la restricción de **dos páginas**. No existe espacio ilimitado en ese
formato: se ajustan líneas/tamaño legible y se admiten hasta 16 conceptos en el
espacio disponible; descripciones/observaciones muy extensas pueden excederlo antes.
El renderizador no trunca ni sustituye caracteres no representables en Windows-1252.
En esos casos conserva los datos y la entrega, indica sheetReady=false en la API y
la pantalla, y la descarga devuelve 422 con el motivo. No afirma que haya generado
el PDF. Un anexo o páginas adicionales requeriría cambiar la decisión de dos páginas;
no se añadió arbitrariamente. El caso habitual de los recorridos probados sí genera
la ficha completa automáticamente. Las firmas permanecen para conformidad manual.

## Ejecución Windows

En PowerShell desde la raíz, con Docker Desktop activo:

```powershell
.\backend\scripts\start-dev.ps1
```

Abrir http://localhost:8080/ (modo real), recargando pestañas abiertas previamente.
Cuenta inicial admin_pruebas, contraseña privada en backend/.cache/acceso-desarrollo.txt.
El script aplica la migración antes del bootstrap y API. No usar SkipBuild al actualizar.
No se requiere Python/PyMuPDF en producción: Go incorpora las plantillas saneadas.

Para repetir navegador en esta computadora:

```powershell
$env:PLAYWRIGHT_MODULE = "C:\Users\Usuario\.cache\codex-runtimes\codex-primary-runtime\dependencies\node\node_modules\playwright"
node checks/attention-browser.mjs
```

## Archivos modificados o creados

- `README.md`
- `backend/cmd/dev-bootstrap/main.go`
- `backend/go.mod`
- `backend/integration/attention_sheet_test.go`
- `backend/integration/brands_test.go`
- `backend/integration/phase_a_test.go`
- `backend/integration/phase_c_test.go`
- `backend/integration/review_test.go`
- `backend/integration/workshop_test.go`
- `backend/internal/domain/workshop.go`
- `backend/internal/domain/workshop_test.go`
- `backend/internal/httpapi/catalog_test.go`
- `backend/internal/httpapi/frontend_options.go`
- `backend/internal/httpapi/technical_sheet.go`
- `backend/internal/httpapi/workshop.go`
- `backend/internal/platform/postgres/frontend.go`
- `backend/internal/platform/postgres/migrate.go`
- `backend/internal/platform/postgres/migrations/003_atencion_ficha.sql`
- `backend/internal/platform/postgres/statistics.go`
- `backend/internal/platform/postgres/store.go`
- `backend/internal/platform/postgres/technical_sheet.go`
- `backend/internal/platform/postgres/workshop_reads.go`
- `backend/internal/platform/postgres/workshop_writes.go`
- `backend/internal/platform/technicalsheet/README.md`
- `backend/internal/platform/technicalsheet/assets/catalogo.json`
- `backend/internal/platform/technicalsheet/assets/pagina-1.jpg`
- `backend/internal/platform/technicalsheet/assets/pagina-2.jpg`
- `backend/internal/platform/technicalsheet/control.go`
- `backend/internal/platform/technicalsheet/pdf.go`
- `backend/internal/platform/technicalsheet/pdf_test.go`
- `backend/scripts/test-local-postgres.ps1`
- `checks/attention-browser.mjs`
- `checks/frontend-browser.mjs`
- `checks/live-docker.mjs`
- `checks/review-browser.mjs`
- `compose.test.yaml`
- `dist/assets/js/components/charts.js`
- `dist/assets/js/components/technical-sheet.js`
- `dist/assets/js/services/api-service.js`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/services/workshop-service.js`
- `dist/assets/js/views/administration.js`
- `dist/assets/js/views/clients-vehicles.js`
- `dist/assets/js/views/operations.js`
- `dist/assets/js/views/payments.js`
- `docs/backend/ATENCIONES_REPORTES_FICHA.md`
- `docs/backend/ESTADO_ENDPOINTS.md`
