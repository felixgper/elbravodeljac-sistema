# Documentos PDF — identidad Diseño 19

> Estado de diseño: iteración anterior no aprobada por el usuario. La propuesta vigente para revisión está en `PDF_PROPUESTA_COMERCIAL.md`. Este informe se conserva como antecedente.

## Cambios de presentación

Cotización y ficha técnica comparten encabezado, emblema oficial proporcionado,
Helvetica/Helvetica Bold, margen de 34 puntos, separadores borgoña, superficies
claras y pie numerado. Se conserva el PNG original sin vectorización ficticia ni
eliminación artificial de su fondo carbón. No se requieren librerías nuevas.

La cotización conserva las agrupaciones de servicios y repuestos, cantidades,
precios, descuentos, total recibido del backend, estado y versión/aprobación.
Se añadieron los datos de contacto existentes a su bloque de cliente. Las filas
largas continúan sin repetir cargos; los encabezados se repiten en cada página.

La ficha ya no se imprime sobre las fotografías de la plantilla del toro.
Utiliza texto real y tablas, sin tarjetas o filas vacías decorativas. Incluye
atención, cliente, vehículo, diagnóstico disponible, trabajos realizados,
importes, avances, pruebas, inspecciones, neumáticos, fluidos, mantenimiento,
observaciones y próximas recomendaciones. Conserva espacios de conformidad,
sin inventar firmas. Una ficha básica ocupa dos páginas; los datos extensos
pueden producir páginas adicionales con encabezado y pie consistentes.

## Datos y compatibilidad

- No se modifica ningún esquema, migración, cálculo, endpoint ni permiso.
- La consulta Go que construye la instantánea para una atención nueva incorpora
  datos que ya existen: tipo de documento, año, diagnóstico, recomendaciones,
  tipo de concepto y pruebas. Son campos adicionales del documento interno,
  necesarios para imprimirlos; no se añaden columnas ni se altera el DTO REST.
- Los valores financieros se imprimen como vienen de PostgreSQL. No se recalculan.
- Los trabajos incluidos provienen de los conceptos confirmados por la consulta
  ya existente. Las inspecciones no registradas dicen «Sin evaluar»; las opciones
  de mantenimiento sin selección no se presentan como realizadas.
- Una descarga histórica con instantánea se renderiza desde esa instantánea
  congelada, sin leer el cliente actual y sin sobrescribir `ficha_pdf` ni
  `ficha_snapshot`. Así puede mostrar la identidad nueva sin modificar el archivo
  histórico almacenado. Descargas sucesivas conservan idénticos bytes.
- Si un registro antiguo solo tiene PDF y carece de instantánea, se conserva su
  PDF original. No se reconstruye su historia con información actual.
- Los datos nuevos del diagnóstico no pueden inventarse para instantáneas
  antiguas que no los guardaron; se muestra únicamente lo disponible.
- Los módulos web de descarga no generan PDF: llaman a las rutas existentes de
  Go, que son las que utilizan estos renderizadores. No necesitaban cambios.

## Archivos

- `backend/internal/platform/technicalsheet/document_brand.go`: identidad compartida.
- `backend/internal/platform/technicalsheet/sheet_layout.go`: ficha con flujo y paginación.
- `backend/internal/platform/technicalsheet/pdf.go`: entrada de renderizado y campos opcionales de instantánea.
- `backend/internal/platform/technicalsheet/quote_layout.go`: cotización y encabezado/pie compartidos.
- `backend/internal/platform/postgres/technical_sheet.go`: consulta de datos existentes y presentación histórica sin sobrescribir el archivo.
- `backend/internal/platform/technicalsheet/document_identity_test.go`: ejemplos ficticios y pruebas de conservación/paginación.
- `backend/integration/attention_sheet_test.go`: regresión de descargas y archivo histórico inalterado.
- `checks/taller-browser.mjs`: admite continuación de la ficha sin imponer dos páginas exactas.
- `docs/backend/DISENO_COTIZACION_PDF.md`: referencia a esta evolución del diseño.
- Este informe.

## Verificación

Ejecución final del 2026-11-03:

| Comprobación | Resultado |
| --- | --- |
| `node checks/run.mjs` | Seis suites aprobadas |
| `go test ./... -v -count=1` | 42 pruebas principales y 52 subcasos aprobados |
| `go vet ./...` | Aprobado |
| `TestBusinessDocumentIdentity` y pruebas de documentos | Aprobadas: diagnóstico, observaciones extensas, trabajos, descuentos, tipos, aprobaciones y conservación de contenido |
| `TestAttentionInspectionAndOriginalSheet` | HTTP/PostgreSQL QA: atención directa/técnica, entrega, diagnóstico de instantánea, permisos y descargas repetidas; archivo e instantánea históricos sin cambios |
| `TestTallerFourLinkedAccounts` | Chrome, flujo completo y descarga de ambos PDF aprobados |
| Apertura y geometría | Poppler/pypdf/pdfplumber abrieron seis muestras, 39 páginas A4; ningún carácter fuera del área imprimible |
| Inspección visual | Cotización, ficha, continuaciones extensas y versiones en grises revisadas |
| Desarrollo | 28 tablas y restricciones PostgreSQL idénticas; documentos, índices y validadores MongoDB idénticos antes/después |
| Protección documental | 90 archivos versionados de modelos/esquemas/migraciones sin cambios; ninguna credencial privada detectada entre los diez archivos pendientes |

Se omitió `TestCleanRestoredCopy`: exige configurar una copia restaurada vacía,
que no se creó ni limpió para esta tarea. Las demás pruebas listadas sí se ejecutaron.

Los primeros intentos identificaron aserciones ligadas al máximo de dos páginas;
se actualizaron para permitir continuación, conservando el caso básico de dos
páginas. Una ejecución posterior obtuvo dos respuestas 503 en consultas de Inicio
mientras se compilaba Go en paralelo. Se conservó ese registro y se repitió la
suite una vez disponible la API: la ejecución final pasó sin esos errores.

Se actualizó únicamente el contenedor API para `http://localhost:8080/`, sin
migraciones, recreación de bases, cambios de credenciales ni operaciones sobre
contenedores ajenos. Las muestras para revisión siguen siendo exclusivamente
ficticias.

Las muestras se generan en memoria con información ficticia, sin insertar sus
clientes, direcciones ni importes en desarrollo. Se guardan localmente en
`backend/.cache/pdf-identidad/`, fuera de Git:

- `cotizacion-pendiente-diseno19.pdf`.
- `cotizacion-aprobada-diseno19.pdf`.
- `ficha-basica.pdf`.
- `ficha-tecnica-diseno19.pdf`.
- `ficha-tecnica-extensa-diseno19.pdf`.
- `cotizacion-multipagina.pdf`.

Las capturas de revisión no utilizan los datos personales de la imagen aportada
por el usuario. No se hace commit, push ni merge.
