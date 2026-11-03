# Revisión de presentación: segunda página con recuadros

## Corrección vigente

Se conserva el logo 19, encabezado borgoña, primera página, importes y datos
funcionales. La segunda página recupera la organización de la referencia:

- Tres tarjetas con borde redondeado: Fugas de fluidos, Revisión de líquidos y
  Neumáticos; columnas B/A/U alineadas. Sin evaluar significa tres casillas vacías.
- Mantenimiento dentro de un recuadro, en dos columnas, con separación entre
  etiquetas y opciones, divisores discretos y sin opciones ajenas al catálogo.
- Otros servicios dentro de un recuadro común, en tres columnas, sin cuadrícula.
- Casillas de 8,5 puntos (antes 7); texto principal del control de 9,2 puntos
  (antes 8–8,5). Se utilizan las medidas de Helvetica para distribuir las palabras
  correctamente, en lugar de comprimir o recortar los textos.
- Las presiones se conservan; se abrevia «copil.» en sus etiquetas auxiliares
  para separar claramente las cifras. Otros líquidos y bobinas conservan sus
  valores cuando existen.
- Las fichas normal, vacía, multiselección y completa conservan dos páginas.
  Las notas extensas continúan en una tercera. Los cierres se mantienen juntos
  cuando caben, sin forzar el tamaño de letra.

## Coherencia de los ejemplos y origen de los valores

La revisión fue exclusivamente sobre archivos y ejemplos en memoria: no se
consultaron ni modificaron registros reales.

El ejemplo anterior de dos aceites y otros servicios fue construido en
`TestSheetCheckboxExamples`. No procedía de PostgreSQL ni se tomó de la imagen.
Su multiselección se conserva como caso sintético de regresión, en
`backend/.cache/ficha-recuadros/ficha-casillas-multiseleccion.pdf`. También se
conserva el PDF anterior en el respaldo.

El nuevo ejemplo de presentación tiene un servicio de mantenimiento de frenos
por S/ 90 y pastillas por S/ 170, total S/ 260. Solo selecciona pastillas
semicerámicas; no declara aceite utilizado ni otros servicios. Las comprobaciones
ficticias de líquido de frenos y presión de aire se muestran como Bien; el resto
queda sin evaluar. No se alteraron los conceptos ni los importes de la primera
página.

La interfaz existente registra `maintenance` como arrays de casillas y
`otherServices` como una lista separada de `execution`/`performedItems`.
`Control.Validate` comprueba IDs de catálogo y duplicados, pero no establece
exclusividad entre viscosidades ni vincula cada otro servicio a una línea
cobrable. Esa limitación preexistente se documenta: no se cambian reglas para
ocultarla. El generador conserva todas las selecciones recibidas y no infiere
materiales ni trabajos desde una cotización. Los datos persistidos provienen del
control técnico de la orden incluido en la instantánea de entrega; esa lógica
no fue modificada.

## Pruebas y comparación

- `go -C backend test ./internal/platform/technicalsheet -v -count=1`: PASS;
  13 pruebas principales y 3 subcasos de cotizaciones.
- `go -C backend vet ./...`: correcto, sin diagnósticos.
- `git diff --check`: sin errores.
- Cinco ejemplos ficticios, once páginas A4: normal 2, multiselección 2,
  vacío 2, completo 2, notas extensas 3. Se verifican estados, columna correcta,
  conteo exacto de marcas, selecciones múltiples y ausencia de inferencias.
- La primera página del ejemplo principal conserva idéntico contenido PDF
  descomprimido respecto de la copia anterior, incluidos sus importes.
- Cotizaciones pendiente y aprobada regeneradas idénticas byte a byte a sus
  versiones anteriores; no se modificó su generador ni los componentes compartidos.
- Renderizado con Poppler y revisión visual. A4, márgenes, encabezados y contenido
  final de notas extensas comprobados. Poppler avisa de fuentes de sustitución
  Symbol/ArialUnicode no disponibles; las páginas usan Helvetica y las casillas
  se dibujan como vectores. Se comprobaron visualmente letras, acentos y marcas.

No se ejecutaron pruebas HTTP ni pruebas contra bases de datos en esta tarea:
los cambios afectan solo la generación visual del PDF. No se reinició el servidor.

## Archivos y protección

Esta revisión cambia únicamente:

1. `backend/internal/platform/technicalsheet/sheet_layout.go`.
2. `backend/internal/platform/technicalsheet/sheet_checkboxes.go`.
3. `backend/internal/platform/technicalsheet/sheet_checkboxes_test.go`.
4. `docs/backend/FICHA_CASILLAS.md`.

Se conservaron los 14 archivos inicialmente pendientes y se creó un respaldo
recuperable en `backend/.cache/ficha-recuadros/antes/`, con rutas originales.
Se verificaron las huellas de 215 archivos protegidos; permanecen intactos,
incluidos catálogo, validación, frontend, persistencia y generadores compartidos.

Entrega visual ficticia en `output/pdf/`: `ficha-tecnica-recuadros.pdf` y sus
imágenes `ficha-tecnica-recuadros-1.png` y `ficha-tecnica-recuadros-2.png`.
Comparación referencia/antes/después y ejemplos adicionales:
`backend/.cache/ficha-recuadros/revision.html` y `comparacion.png`.
No se sobrescribieron documentos históricos. No se tocaron PostgreSQL, MongoDB,
migraciones, cuentas ni reglas. No se realizó commit, push ni merge.

---

## Antecedente: primera recuperación de casillas, sustituida visualmente por esta revisión

# Ficha técnica: recuperación de casillas

Corrección visual sobre el diseño comercial aprobado, en la rama
`backend/arquitectura-go-mongodb`. No cambia la estructura de datos ni las reglas
del taller. No se realizaron commit, push ni merge.

## Representación

- Los 21 controles del catálogo existente muestran tres casillas B/A/U:
  Bien, Atención y Urgente. Solo se marca el estado registrado. Un control
  ausente o Sin evaluar queda sin marcas; la leyenda explica su significado.
- Las 41 opciones de los 12 grupos de mantenimiento se imprimen completas.
  Solo se marcan los valores de `Control.Maintenance`, respetando sus arrays
  existentes y las selecciones múltiples permitidas por el modelo.
- Los 14 otros servicios muestran casillas basadas en `Control.OtherServices`.
  No se infieren materiales ni trabajos a partir de conceptos cotizados.
- Las presiones PSI conservan sus valores y se presentan junto a Neumáticos.
  Otros líquidos y bobinas conservan su texto; si ambos existen se distribuyen
  en dos columnas para mantener las fichas normales en dos páginas.
- Observaciones largas continúan en otra página sin recortes. Se conservan
  importes, observaciones, próximo mantenimiento y espacios de conformidad.

No se añadieron campos del ejemplo visual. Se mantiene el logo 19, el encabezado
borgoña, el pie de página y las tablas de la primera página.

## Archivos de esta corrección

1. `backend/internal/platform/technicalsheet/sheet_layout.go`: presentación del control.
2. `backend/internal/platform/technicalsheet/sheet_checkboxes.go`: dibujo y distribución de casillas.
3. `backend/internal/platform/technicalsheet/sheet_checkboxes_test.go`: estados, selecciones y ejemplos ficticios.
4. `docs/backend/FICHA_CASILLAS.md`: este informe.

Había 11 archivos pendientes antes de esta tarea. Se conservan todos; el archivo
de distribución ya estaba entre ellos. Con tres archivos nuevos, Git muestra
14 archivos pendientes. Esto no significa que los 14 hayan cambiado en esta tarea.

## Protección y comparación

Respaldo recuperable de los archivos inicialmente pendientes, con sus rutas
originales, en `backend/.cache/ficha-casillas/antes/`. La caché está excluida de Git.

Se verificaron las huellas SHA-256 de 215 archivos protegidos: ninguno cambió.
Incluyen frontend, operaciones, autenticación, modelo de control, catálogo,
migraciones y generadores compartidos de documentos.

La primera página del ejemplo nuevo y del ejemplo anterior tiene exactamente
el mismo contenido PDF descomprimido. Las cotizaciones ficticias pendiente y
aprobada regeneradas son idénticas byte a byte a las copias anteriores.

## Verificación ejecutada

- `go -C backend test ./internal/platform/technicalsheet -v -count=1`: PASS,
  13 pruebas principales y 3 subcasos de cotizaciones.
- `go -C backend vet ./...`: terminó correctamente, sin diagnósticos.
- `git diff --check`: sin errores de espacios.
- Casillas: estado Bien, Atención, Urgente, Sin evaluar y ausente; columna
  correcta, máximo una selección por inspección y conteo exacto de marcas.
- Mantenimiento: selección múltiple conservada, opciones vacías y ausencia de
  selección automática desde conceptos cotizados o trabajos de la primera página.
- Cuatro ejemplos validados con las reglas existentes de `Control.Validate`:
  mixto, sin selecciones, todos los grupos con selecciones y notas extensas.
- Renderizado Poppler e inspección visual de los ejemplos; formato A4 y texto
  dentro de los márgenes. Tres ejemplos de dos páginas y uno de tres: nueve
  páginas en total. Las notas extensas conservan su texto final.

Artefactos locales en `backend/.cache/ficha-casillas/`:

- `ficha-casillas.pdf`: ejemplo principal ficticio corregido.
- `ficha-1.png` y `ficha-2.png`: vistas previas del ejemplo principal.
- `vistas-previas.html`: comparación antes/después y variantes.
- `validacion.json`, `tests.log`, `vet.log`: resultados de comprobaciones.

Son pruebas de generación PDF en memoria; no se ejecutaron integraciones HTTP
ni conexiones a bases de datos para esta corrección. No se modificaron bases,
documentos históricos almacenados, cotizaciones ni reglas de negocio. El servidor
local no se reconstruyó ni reinició: estos PDF permiten revisar la plantilla
antes de desplegarla.
