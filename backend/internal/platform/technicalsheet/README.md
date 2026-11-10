# Generadores PDF oficiales — Diseño 19

`quote_layout.go` y `sheet_layout.go`, con `document_brand.go`, `sheet_checkboxes.go` y `pdf.go`, generan cotización y ficha borgoña. El emblema oficial se lee de `assets/diseno19-original.png`; `assets/catalogo.json` define los controles existentes. No se añaden resultados, firmas ni conceptos ausentes.

La ficha utiliza tres recuadros (fugas, líquidos y neumáticos), casillas B/A/U sin marca cuando no se evaluó, mantenimiento y servicios realizados. El contenido extenso continúa en páginas adicionales sin recortarse. Las cotizaciones conservan importes, estado y revisión, con encabezados de continuación.

La ruta de compatibilidad `buildDocument(true)` utiliza `assets/pagina-1.jpg` y `assets/pagina-2.jpg`. El generador oficial utiliza el diseño vectorial.

PostgreSQL entrega primero los bytes de `ficha_pdf` ya archivados, incluso con snapshot. Una ficha emitida no cambia al actualizar el generador. La descarga mantiene los permisos de la entrega. Si falta el PDF archivado, el sistema puede generarlo a partir del snapshot disponible.

Poppler y Python se utilizan para la inspección externa de PDF en QA.
