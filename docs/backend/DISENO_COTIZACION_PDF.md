# Diseño de cotización PDF — El Bravo del JAC

Se revisaron visualmente los dos archivos del Escritorio: `cotizacion-25.pdf`
(salida anterior) y `COTIZACIÓN - EXCEL_1323_261027_231906.pdf` (referencia
solicitada). Se adoptó la distribución de encabezado, datos en bloques, tabla de
tres columnas con bordes, agrupación, total y agradecimiento. No se incorporaron
marca, fotografía, logotipos, clientes, teléfonos ni datos de Autoservis Marcelo.

## Formato implementado

- A4, márgenes de 34 puntos, borgoña/blanco/gris y logo institucional vectorial.
  La copia SVG para Go es idéntica al recurso web y una prueba comprueba esa
  igualdad. No se descarga una imagen ni se requiere un conversor externo al
  generar el PDF. La reproducción vectorial admite los comandos del SVG actual;
  un cambio futuro del recurso deberá verificarse con estas pruebas.
- Datos del taller desde la vista existente; teléfono solo si el registro
  configurado expone `phone`. Actualmente la vista no lo configura: no se inventa
  ni se copia el teléfono del documento de referencia. Datos del cliente,
  documento, vehículo, emisión, recepción, número y versión desde PostgreSQL.
- Tabla **DESCRIPCIÓN DE SERVICIOS | CANTIDAD | IMPORTE**, con cantidades
  centradas y subtotales con prefijo S/ alineados a la derecha. Secciones
  **SERVICIOS** y **REPUESTOS** en borgoña; orden original conservado dentro de
  cada sección. Descripciones, párrafos vacíos y espacios se conservan, con
  ajuste de líneas para impresión. El texto registrado no se reemplaza.
- Precio unitario y descuento por unidad se conservan como información
  secundaria de la fila. El importe es `subtotal` guardado: no se recalcula ni
  se cobra una segunda vez una actividad descriptiva. Las líneas dentro de un
  concepto constituyen una sola fila con un solo importe. Los conceptos con
  subtotal cero indican **Sin cargo independiente** y S/ 0.00; eso no presupone
  que cualquier concepto sin precio sea un servicio realmente ejecutado.
- El modelo actual no tiene una relación entre actividades informativas y un
  concepto padre que las facture agrupadas. Por ello no se inventan paquetes ni
  cobros agrupados. Se utilizan los conceptos multilínea y los subtotales
  existentes, incluidos los de importe cero. No se cambia el formulario.
- **TOTAL COTIZADO** destacado en borgoña, igual al total registrado. Sin
  impuestos, descuentos globales ni cálculos nuevos.
- Estado Pendiente de aprobación/Aprobada/Rechazada; las no aprobadas indican
  presupuesto propuesto sin autorización de trabajo. Aprobación, persona, medio,
  fecha y versión solo cuando constan registrados. Documento y aprobación
  conservan sus identificaciones de versión separadas.
- Pie **¡GRACIAS POR CONFIAR EN NOSOTROS!**, identificación y página. Sin barras
  negras ni fotografías que dificulten imprimir.
- Filas normales completas: se trasladan a la página siguiente cuando no caben.
  Para el caso excepcional de un único concepto más alto que una página útil,
  se permite continuación identificada sin repetir cantidad ni importe.
  Encabezados de tabla, sección y datos identificativos se repiten. El total y
  los datos de aprobación se reservan juntos cuando caben en una página nueva.

Continúa el endpoint `GET /api/v1/cotizaciones/{id}/pdf` y el botón **PDF** junto
a **Abrir**. Exige `cotizaciones.ver`, permite descargar pendientes, aprobadas
y rechazadas, y realiza solo lecturas. Se exporta la versión actual guardada de
la cotización solicitada; no se agregó un selector de revisiones históricas.
La fuente sigue siendo WinAnsi, como en el mecanismo anterior: español y
caracteres occidentales; caracteres no representables provocan error explícito,
sin omitirlos ni inventar sustitutos.

## Verificación

Las pruebas usan exclusivamente `elbravodeljac_test` y
`elbravodeljac_tecnico_test`, nunca ElBravoDelJac ni los registros de desarrollo.

Resultado final: **aprobadas** las cinco suites JavaScript, Go completo con
PostgreSQL/MongoDB, auditoría HTTP/Chrome y `go vet`. La muestra aprobada de cinco
conceptos ocupa una página A4 con su total y aprobación; el caso extremo de
65 conceptos ocupa 19 páginas sin perder el último concepto ni el total. Se
revisaron visualmente la muestra, la primera página extensa y la última.

- `node checks/run.mjs`: las cinco suites del frontend.
- `go test ./... -count=1 -v` en Compose aislado: backend, PostgreSQL y MongoDB.
- Auditoría HTTP y Chrome en Windows mediante
  `.\backend\scripts\test-audit.ps1`: permisos, estados, descarga junto a Abrir,
  flujo existente hasta entrega y ficha técnica, y comparación de huellas antes
  y después de descargar. Se conservan los errores 401/403/400/404 existentes.
- Se añadió una cotización ficticia persistida con tres servicios, dos
  repuestos, descripciones multilínea, cantidades, descuentos y conceptos sin
  cargo. Total de PostgreSQL **S/ 515.00**: 135 + 0 + 180 + 200 + 0. Se descargó
  en estado Pendiente, Aprobada y Rechazada, sin modificar datos al exportar.
- Pruebas unitarias del logo, agrupación, conceptos sin duplicar, conservación
  de espacios/párrafos, datos de aprobación incompletos, 40 filas normales sin
  corte y 65 conceptos con una descripción muy extensa sin perder el final.
- `go vet ./...` y `git diff --check`.
- Se renderizan con Poppler los documentos de prueba y se revisan encabezado,
  columnas, bordes, cantidades, importes, total, aprobación, pie y continuaciones.

El script guarda resultados en `backend/.cache/audit-tests.log`,
`audit-native.log` y `quote-redesign/test-run.log`. Los PDF de los tres estados
están en `backend/.cache/audit-pdf/nuevo-diseno-*.pdf`. La vista previa entregada
está en `output/pdf/cotizacion-vista-previa.pdf` y su PNG, ignorados por Git.
Son datos ficticios de prueba; los ejemplos personales de los archivos de
referencia no se convirtieron en constantes del sistema.

## Archivos

- `backend/internal/platform/technicalsheet/quote.go`: instantánea y ajuste de
  texto conservando espacios/párrafos.
- `backend/internal/platform/technicalsheet/quote_layout.go`: maquetación,
  tabla, agrupación, alineación monetaria, paginación, estado, total y pie.
- `backend/internal/platform/technicalsheet/quote_logo.go`: dibujo vectorial del
  SVG institucional embebido.
- `backend/internal/platform/technicalsheet/assets/bravo-borgona.svg`: mismo
  recurso institucional que usa la web.
- `backend/internal/platform/technicalsheet/quote_test.go`: pruebas de diseño.
- `backend/integration/audit_corrections_test.go`: cotización ficticia persistida
  y descargas de los tres estados.
- `.gitignore`: exclusión específica de vistas previas generadas.
- Este documento y enlace desde `COTIZACIONES_PDF_REPORTES.md`.

No se modificaron formulario web, CSS, endpoints de diagnósticos/órdenes,
generador/plantillas de ficha técnica, tablas, claves foráneas, precios o
credenciales. Se conserva la rama actual, sin commit, push ni merge.

Se reconstruyó y actualizó únicamente `api` del proyecto Compose de desarrollo
existente. Las huellas de las 26 tablas fueron idénticas antes/después de esa
actualización. No se ejecutaron migraciones ni se reiniciaron servicios de bases
de datos. La descarga utiliza el botón PDF existente en Cotizaciones.
