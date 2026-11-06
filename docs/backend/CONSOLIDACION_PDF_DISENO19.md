# Consolidación de PDF Diseño 19

Validación del 5 de noviembre de 2026. Rama `feature/consolidacion-pdf-diseno19`, desde `main` sincronizada en `1e277b3c1a2c2676979c4e3a4164bb6e5db79cac`.

## Comportamiento y diseños conservados

Una descarga de ficha técnica con `entregas.ficha_pdf` no vacío devuelve **exactamente los bytes archivados**, aunque exista `ficha_snapshot`. No vuelve a interpretar el snapshot ni regenera el documento. Un PDF emitido conserva así su presentación original después de actualizar el generador; las fichas antiguas mantienen también sus diseños anteriores.

Si falta el PDF, se conserva el flujo compatible: generar desde el snapshot existente o reconstruir la atención antigua y archivar el resultado una vez. No se eliminan snapshots. Permanecen las comprobaciones actuales de autenticación, permisos, orden finalizada, cotización aprobada y saldo cancelado. No se modifica el esquema ni se añade una migración.

Los generadores, imágenes, catálogo de inspección y estilos funcionales de `main` se conservan íntegros:

- Login oscuro de dos columnas con logo Diseño 19; Dashboard carbón/borgoña.
- Cotización borgoña de `main`: estados Pendiente, Aprobada y Rechazada, revisión/aprobación e importes originales, con continuaciones de página.
- Ficha borgoña con servicios realizados, tres tarjetas de inspección, B/A/U, mantenimiento y casillas. Sin evaluar permanece sin marcar. El contenido extenso continúa en más páginas.

No se copiaron los generadores antiguos del contenedor ni `sheet_brand_deployed.go`. Su variante separada no es necesaria para compilar los generadores oficiales de `main`.

## Pruebas y resultados

| Comprobación ejecutada | Resultado |
|---|---|
| `go test ./... -json -count=1`, en copia temporal fiel a los dos archivos funcionales modificados | 46 PASS, 0 FAIL, 2 SKIP de 48 pruebas de primer nivel |
| Prueba nueva `TestApprovedBurgundyPDFLayout` y suite PDF completa | 14 PASS, 0 FAIL, 0 SKIP; incluye las 13 pruebas PDF ya contadas en la suite general |
| Auditoría adicional `TestSheetFunctionalAudit`, con navegador | PASS en otro PostgreSQL temporal exclusivo; cubre una de las dos omisiones anteriores |
| Total de casos Go distintos verificados, sin duplicar repeticiones | **48 PASS, 0 FAIL, 1 SKIP**, de 49 pruebas de primer nivel |
| `go vet ./...` | PASS |
| `node checks/run.mjs` | Seis suites PASS: verify, workflow, edge-cases, real-api, report-periods y diagnosis-text |
| PostgreSQL/MongoDB reales de QA | PASS: integraciones HTTP, permisos, flujo del taller, transacciones, GridFS, reintentos y fallos de MongoDB |
| Chrome: AuditBrowser, PhaseE, CarbonDesign y TallerFourLinkedAccounts | PASS: formularios, persistencia, roles, login por correo/usuario, logout, períodos y responsive |
| Ocho PDF ficticios nuevos comparados con el `main` aprobado de la etapa 1 | Los ocho coinciden byte por byte; 16 páginas renderizadas a PNG, sin cambiar los generadores |

`TestAttentionInspectionAndOriginalSheet` comprueba mediante HTTP real dos tipos de atención, nuevas entregas y permisos. Para cada tipo prueba un PDF archivado distinto del que produciría el generador, con snapshot válido, ausente y no interpretable. Repite cada descarga y compara bytes, snapshot, archivo y `xmin`. Una traza de las conexiones reales del store exige **cero UPDATE** en las descargas archivadas. Para registros sin PDF, exige una sola actualización al generarlo y ninguna al descargarlo nuevamente, con y sin snapshot previo.

`TestApprovedBurgundyPDFLayout` fija las coordenadas y color de cabecera/pie aprobados independientemente de las constantes del generador. Comprueba las cotizaciones de los tres estados, varias páginas y fichas con inspecciones evaluadas/no evaluadas. Se conservan las pruebas existentes del logo original, importes, agrupación, textos largos y marcado explícito de casillas.

Las omisiones se distinguen de las aprobaciones: la suite general omitió inicialmente `TestSheetFunctionalAudit` por su flag/puerto dedicado y `TestCleanRestoredCopy`. La primera se ejecutó después con éxito, incluyendo Chrome, en el puerto QA 25439. **Solo queda sin ejecutar `TestCleanRestoredCopy`**: exige una copia restaurada específica `jac_restore_<fecha>` en 25435 y un acceso privado previo. No se utilizó una restauración ni configuración privada de un entorno existente para satisfacer esa condición.

Las pruebas generales y de Chrome utilizaron un proyecto Compose nuevo, PostgreSQL 18.6 en 25440 y MongoDB 9 en 27040, almacenamiento `tmpfs` y contraseñas efímeras en memoria. La auditoría específica de ficha utilizó otro proyecto PostgreSQL `tmpfs` en 25439. Son instancias nuevas, con inicialización oficial únicamente sobre bases vacías. No se usó PostgreSQL original de 5432, la instancia ajena de 5433 ni los volúmenes del desarrollo habitual.

El entorno temporal contiene una copia de los archivos versionados; los scripts de navegador escribieron sus fixtures y evidencias solo allí. No se ejecutaron los scripts manuales que apuntan al `localhost:8080` habitual. Los proyectos QA se dejaron iniciados, sin detener, reconstruir ni reemplazar servicios anteriores.

## Evidencias visuales

Capturas reales de Chrome y renderizados de los PDF de prueba. **Todos los nombres, importes, documentos y resultados visibles pertenecen a fixtures ficticios de QA**, no a registros de la empresa. No se incluyen contraseñas, cookies, tokens, snapshots reales ni respaldos.

| Evidencia | Archivo |
|---|---|
| Login aprobado, Chrome | [Ver captura](evidencias/consolidacion-pdf-diseno19/login-qa.png) |
| Dashboard aprobado con datos persistidos de QA, Chrome | [Ver captura](evidencias/consolidacion-pdf-diseno19/dashboard-qa.png) |
| Cotización borgoña pendiente, render PDF | [Ver captura](evidencias/consolidacion-pdf-diseno19/cotizacion-borgona.png) |
| Ficha técnica, primera página | [Ver captura](evidencias/consolidacion-pdf-diseno19/ficha-pagina-1.png) |
| Ficha técnica, segunda página y controles | [Ver captura](evidencias/consolidacion-pdf-diseno19/ficha-pagina-2.png) |

Las muestras Pendiente/Aprobada/Rechazada tienen una página, la cotización de 24 conceptos tres páginas, las fichas sin evaluar/parcial/completa dos páginas y la ficha extensa cuatro. No se forzó una paginación diferente a la aprobada. Los PDF ficticios, hashes comparativos, logs JSON y demás páginas renderizadas quedan localmente en `backend/.cache/consolidacion-pdf-diseno19/`, ignorados por Git.

## Alcance del cambio y actualización pendiente

El único cambio funcional está en `postgres.Store.SheetPDF`: prioridad al PDF archivado. El resto son regresiones automatizadas y este informe con cinco capturas ficticias. No se modificaron Login, Dashboard, otros módulos, generadores PDF, assets originales, migraciones, modelos de datos ni bases existentes. No se realizó limpieza documental.

Para verificaciones locales sin conexiones configuradas: `node checks/run.mjs`, y desde `backend`, `go test ./...` y `go vet ./...`. Las integraciones se omiten si faltan sus variables; eso no constituye aprobación. Para repetirlas deben prepararse proyectos QA **nuevos y exclusivos**, usando `compose.test.yaml`, sus nueve migraciones oficiales y flags de navegador; no apuntar las variables de integración a instalaciones existentes ni usar `start-dev.ps1` como preparación de pruebas.

El Docker habitual de `localhost:8080` **no se actualizó**. Continúa sirviendo su versión previa hasta que se apruebe el despliegue. El Pull Request consolida la cotización borgoña oficial y conserva las fichas ya emitidas; no autoriza por sí mismo un merge ni la reconstrucción del servicio habitual.
