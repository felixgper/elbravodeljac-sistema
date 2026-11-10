# Generadores PDF oficiales — Diseño 19

`quote_layout.go` y `sheet_layout.go`, con `document_brand.go`, `sheet_checkboxes.go` y `pdf.go`, generan cotización y ficha borgoña. El emblema oficial se lee de `assets/diseno19-original.png`; `assets/catalogo.json` define los controles existentes. No se añaden resultados, firmas ni conceptos ausentes.

La ficha utiliza tres recuadros (fugas, líquidos y neumáticos), casillas B/A/U sin marca cuando no se evaluó, mantenimiento y servicios realizados. El contenido extenso continúa en páginas adicionales sin recortarse. Las cotizaciones conservan importes, estado y revisión, con encabezados de continuación.

`assets/pagina-1.jpg` y `pagina-2.jpg` siguen leídos por la ruta interna de compatibilidad `buildDocument(true)`; se conservan aunque el generador oficial use el diseño vectorial actual. No se elimina código para forzar su retirada. El SVG antiguo del toro no tenía consumidor de lectura y se retira; el wildcard de embed no era una referencia de uso.

PostgreSQL entrega primero los bytes de `ficha_pdf` ya archivados, incluso con snapshot. Una ficha emitida no cambia al actualizar el generador. La descarga mantiene los permisos de la entrega. Si el PDF no fue archivado, el generador admite la lectura compatible desde el snapshot disponible..

No se requiere Poppler/Python para ejecución: solo para inspección externa de PDF en QA. No se modifica diseño ni paginación en la limpieza.
