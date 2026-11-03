# Propuesta visual comercial de cotización y ficha técnica

**Estado: pendiente de revisión del usuario. No es un diseño aprobado.**

## Revisión inicial y conservación

Se partió de `backend/arquitectura-go-mongodb`, último commit `b50f3a5`, con diez
archivos locales pendientes de la iteración anterior. Se revisaron `git status`,
`git diff`, los renderizadores y las diferencias respecto de Git. Se conservó
una copia local del estado previo, sin resetear ni descartar cambios.

En Git, la ficha usaba dos fondos JPG y una maquetación fija; la cotización tenía
un encabezado más simple. El intento local anterior ya había incorporado datos
técnicos disponibles, la instantánea congelada y la protección del archivo
histórico. Esas mejoras se conservaron. En particular,
`backend/internal/platform/postgres/technical_sheet.go` conserva exactamente
los mismos bytes que tenía al comenzar esta revisión: no se amplían consultas,
campos de datos ni reglas de negocio en esta iteración visual.

## Criterio gráfico

- Encabezado borgoña con el emblema oficial Diseño 19 y nombre del taller.
- Tipografía existente, títulos destacados, texto normal legible y fondos blancos.
- Datos de cliente y vehículo agrupados, bordes discretos y filas alternadas.
- Tablas con encabezados de alto contraste, grupos de servicios/repuestos y
  precios unitarios/descuentos conservados como detalle de cada concepto.
- Totales destacados a la derecha, sin recalcular valores financieros.
- Pie de marca y numeración real «PÁGINA N DE M» en cada página.
- Sin marca de agua: se priorizó la lectura y la impresión.

Las imágenes recibidas solo se utilizaron como referencia estética. No se
copiaron personas, RUC, direcciones, placas, fechas, importes, servicios,
condiciones ni evaluaciones. No se agregaron VIN, motor, asesor, garantías,
vigencias, facturación ni controles nuevos.

## Distribución y comparación

| Muestra | Intento anterior | Propuesta |
| --- | --- | --- |
| Cotización sencilla | Encabezado lineal y datos dispersos | Encabezado comercial, información agrupada, tabla y total destacados |
| Ficha con diagnóstico y pruebas | Tres páginas, una con abundante espacio vacío | Dos páginas con los mismos datos |
| Ficha con controles completos | Riesgo de página solo para firmas | Dos páginas; controles y cierre se conservan juntos |
| Ficha sencilla | Controles partidos entre páginas | Bloque técnico completo en una página cuando cabe |
| Ficha extensa | Cierre aislado al final | Observaciones, próximo mantenimiento y conformidades viajan juntos |

La paginación mide los bloques con las mismas funciones de dibujo antes de decidir
si deben continuar. No impone dos páginas ni reduce la fuente para forzar el ajuste.
Los títulos y las primeras filas permanecen juntos. Las filas normales no se
separan; un concepto excepcionalmente mayor que una página utiliza continuación
explícita y no repite su cantidad o importe.

## Fuentes, datos y funcionalidad

Ambos PDF siguen siendo documentos independientes. Las rutas, JWT, permisos,
frontend, cálculos y fuentes de datos conservan su funcionamiento. No se añadió
ningún campo de formulario ni de base de datos. Las mejoras funcionales locales
previas de diagnóstico y descargas históricas permanecen intactas.

No se accedió a las bases reales o de desarrollo para crear o modificar registros.
Las operaciones de integración se ejecutaron solo en `elbravodeljac_test` y
`elbravodeljac_tecnico_test`, del proyecto QA existente del 2026-11-03. No se
recrearon bases ni volúmenes ni se ejecutó limpieza. No se reinició ni actualizó
el contenedor API de desarrollo en esta revisión: la propuesta se entrega en los
PDF y en el servidor HTTP de pruebas, para revisión visual antes de publicarla.

## Archivos

Esta revisión ajusta:

1. `backend/internal/platform/technicalsheet/document_brand.go`: encabezado, paleta,
   títulos, paneles y pies con total de páginas compartidos.
2. `backend/internal/platform/technicalsheet/quote_layout.go`: bloques de datos,
   tabla, filas alternadas, total y aplicación del estilo comercial.
3. `backend/internal/platform/technicalsheet/sheet_layout.go`: composición compacta,
   tabla de trabajos unificada y bloques de control/cierre medidos para evitar
   páginas casi vacías.
4. `backend/internal/platform/technicalsheet/pdf.go`: color de texto de presentación
   y comentario de renderizado coherente con las instantáneas históricas.
5. `backend/internal/platform/technicalsheet/document_identity_test.go`: muestras
   ficticias sencillas, varios conceptos/trabajos, controles completos y casos
   extensos; comprobación del cierre sin una tercera página huérfana.
6. `backend/integration/attention_sheet_test.go`: caso extenso de 3.852 caracteres,
   dentro del límite real de 4.000. El texto de 1.600 caracteres del caso anterior
   ahora cabe sin anexo; se conserva la comprobación de paginación y datos completos
   con un caso realmente más largo.
7. `docs/backend/DOCUMENTOS_PDF_DISENO19.md`: identificación del informe anterior
   como iteración no aprobada, sin borrar su contenido.
8. Este informe.

Además se mantienen, sin cambios respecto al inicio de esta tarea:

- `backend/internal/platform/postgres/technical_sheet.go`.
- `checks/taller-browser.mjs`.
- `docs/backend/DISENO_COTIZACION_PDF.md`.

En conjunto Git conserva los diez archivos locales anteriores y añade este informe:
**once archivos pendientes**, sin temporales ni muestras PDF incorporados a Git.

## Pruebas y revisión

| Comprobación ejecutada | Resultado |
| --- | --- |
| `node checks/run.mjs` | Seis suites aprobadas |
| `go test ./... -v -count=1` en QA | 42 pruebas principales y 52 subcasos aprobados |
| `go vet ./...` | Aprobado |
| `go test ./internal/platform/technicalsheet -v -count=1` | Generación, precios/descuentos originales, versiones, fila larga, conceptos sin duplicar y cierre sin páginas huérfanas aprobados |
| Integración HTTP/PostgreSQL | Atención directa/técnica, entrega, documento extenso, instantánea íntegra y archivo histórico sin sobrescribir aprobados |
| Navegador | Descargas, permisos, flujo del taller y demás regresiones de la suite aprobadas |
| PDF reales de muestra | Nueve archivos abiertos y renderizados con Poppler; 42 páginas A4 |
| Geometría y numeración | Todas las páginas dentro de márgenes; «PÁGINA N DE M» comprobado en cada página |
| Revisión visual | Todas las páginas en galerías, páginas principales y continuaciones revisadas; comparación anterior/propuesta incluida |
| Galería Chrome | Diez secciones, 51 imágenes de vista previa/comparación, cero imágenes rotas y cero errores JavaScript |
| Protección | 91 archivos de modelos/esquemas/migraciones y consulta Go de ficha idénticos al inicio; cambios previos protegidos |

Se omitió `TestCleanRestoredCopy` porque no se configuró una copia restaurada
vacía. No se ejecutó limpieza para habilitarla ni se cuenta como aprobada.

La revisión visual detectó dos problemas que se corrigieron: el cierre aislado
al final de la ficha extensa y la división innecesaria de controles en la ficha
sencilla. Se agruparon mediante medición previa del contenido; no se ocultaron
campos, no se redujo el texto principal ni se forzó un número fijo de páginas.

El test HTTP que exigía un anexo con 1.600 caracteres se adaptó a 3.852 caracteres,
que realmente necesitan continuación en la nueva composición. Se mantuvo su
comprobación de snapshot completo y exportación correcta. La ejecución final
completa pasó después de esos ajustes.

Las muestras de revisión básica también usan datos de taller ficticios. Las
imágenes de referencia no se incorporaron a los PDF, capturas ni fixtures.

La galería local `backend/.cache/pdf-revision-comercial/vistas-previas.html`
contiene todos los PDF ficticios, las imágenes de cada página y la comparación
con el intento anterior. El manifiesto de páginas, capturas de Chrome, registros
Go y copia inicial permanecen excluidos de Git. Los ejemplos no utilizan datos
personales reales y no se han publicado.

No se hizo commit, push ni merge. La aprobación visual queda pendiente del usuario.
