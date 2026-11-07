# Limpieza del repositorio — entregables oficiales

Fecha: 6 de noviembre de 2026. Rama `feature/limpieza-entregables-oficiales`, creada desde `main` sincronizada en `4cb8e7ccf5d071cdda32e15954a4b8be579a31f3` (PR #6 fusionado).

## Resultado y alcance

Se retiran **83 archivos rastreados** justificados. Se actualizan guías y referencias; no se cambia código funcional, pruebas, configuraciones, SQL, migraciones ni los diagramas oficiales. Los únicos recursos funcionales retirados son dos copias de un SVG antiguo del toro, sin consumidor de lectura/URL, y una propuesta JSON de MongoDB que no se ejecutaba. El contrato BSON embebido permanece intacto.

No se ejecutó limpieza del sistema de archivos privados: cada retirada fue sobre una ruta exacta, rastreada, dentro del repositorio y con SHA-256 coincidente con la línea base. No hubo borrado recursivo, `git clean`, reset, modificación de historia ni eliminación de ramas. Los archivos ignorados y datos locales permanecen fuera del commit.

## Archivos retirados

El [inventario completo por archivo](LIMPIEZA_ARCHIVOS.csv) registra las 83 rutas, tamaño original, SHA-256, justificación y commit desde el que recuperarlas.

| Grupo | Archivos | Justificación / resolución de dependencias |
|---|---:|---|
| `docs/modelos/V23_Chen_Corregido/` | 22 | Sustituido por el Chen V24 oficial, nativo y autónomo. La explicación V24 ahora remite a los modelos/SQL vigentes; las referencias de procedencia histórica quedan en Git. |
| Auxiliares en las dos carpetas V24 | 27 | Manifiestos/validaciones ligados a un estado histórico, exportadores con rutas privadas y caché GraphViewer, vistas previas/capturas de bloques redundantes y comparación V23/V24. Se conservan fuentes editables, exportaciones completas, inventarios y generadores independientes. |
| Informes antiguos `docs/backend/*.md` | 25 | Fases A–E, auditorías puntuales y diseños anteriores, algunos expresamente no aprobados. Contratos actuales en backend/README, ENDPOINTS y las guías vigentes conservadas. |
| Guías iniciales en `docs/` | 4 | Guías de demo, recorrido antiguo y propuesta de integración futura. Sustituidas por README, backend/README y checks/README. |
| `docs/previews/` | 2 | Propuestas visuales descartadas; sin consumidores. No se retiran capturas QA de los diseños oficiales. |
| `bravo-borgona.svg` en frontend y assets PDF | 2 | No había referencias funcionales a esos logos. Un wildcard `go:embed assets/*` no implica que se lea el SVG; la lectura del logo oficial utiliza Diseño 19. Compilación y PDF comprobados tras retirarlo. |
| `database/mongodb/collections.schema.json` | 1 | Contrato JSON propuesto, sin loader de código. Se corrigen sus dos README consumidores; el validador BSON real es `backend/internal/platform/mongodb/validators.json`. |

Para recuperar un antecedente: consultar la ruta del CSV en el [árbol anterior a esta limpieza](https://github.com/felixgper/elbravodeljac-sistema/tree/4cb8e7ccf5d071cdda32e15954a4b8be579a31f3) o usar `git show COMMIT:ruta`. No restaurarlo sobre la versión oficial sin revisión. Git conserva los blobs: la reducción de la copia de trabajo no borra ni reduce por sí sola toda la historia del repositorio.

## Archivos conservados y motivo

| Ruta o grupo | Uso / motivo de conservación |
|---|---|
| `docs/Balsamiq_V14_1_REPARADO.bmpr`, `docs/balsamiq-v14.1.json` | Prototipo V14.1 editable e inventario de 23 pantallas; el JSON lo consume checks/prototype.mjs. No se confunde con estadísticas ni estilos del aplicativo real. |
| `docs/modelos/V24_Chen_Resumido_Exposicion/El_Bravo_del_JAC_V24_Chen_Resumido.*` | Único Chen vigente: editable, PDF, PNG y SVG del mismo modelo, 13 entidades / 15 relaciones. Sin cambios de bytes. |
| `docs/modelos/Logico_CrowsFoot_Sin_Tipos/` | Único lógico oficial sin tipos: 28 tablas, 208 atributos, 50 FK. Se conservan editable, PDF, PNG y validación original. |
| `docs/modelos/V24_Logico_CrowsFoot/El_Bravo_del_JAC_V24_Logico_Completo.*` | Referencia física detallada V24 **con tipos**; nombre histórico preservado para no romper rutas. Una versión, no un segundo lógico oficial. |
| Fuentes V24: ESPECIFICACION_V24.json, RELACIONES_V24.json, ESQUEMA_VERIFICADO.json, INVENTARIO_FK.csv, REGLAS_POR_FK.json, generar_v24.py, inventariar_esquema.py, generar_logico.py | Fuentes de mantenimiento autónomas. No necesitan V23. No se ejecutaron ni modificaron; pueden sobrescribir salidas y solo deben usarse bajo revisión. El Draw.io completo permite editar/exportar la entrega sin los exportadores privados retirados. |
| Inventario, cardinalidades, explicación y limitaciones de los modelos | Mantienen criterios NULL/UNIQUE, puentes N:M, diagnóstico opcional y separación cargo laboral/acceso. Las limitaciones no se presentan como cambios aplicados. |
| `database/postgresql/esquema_entrega.sql`, `docs/modelos/COMPARACION_BASE_ACTUAL.md` | Entrega estructural verificada el 5/11/2026 y alcance de esa consulta. No modificados ni ejecutados en esta etapa. |
| `database/postgresql/esquema_actual.sql`, migraciones 001–009 y su registro Go | Dependencias reales de inicialización y pruebas; no se sustituyen por el SQL de entrega. |
| `backend/` Go, go.mod/sum, Dockerfiles, Compose, scripts, `.env.example`, `.gitignore`, `.dockerignore` | Compilación, configuración e inicialización preservadas. Overrides QA y utilidades manuales históricas quedan por compatibilidad/propiedad de recursos: no se altera código para eliminarlos artificialmente. |
| `checks/`, `backend/integration/`, pruebas Go y runners aislados | Todas conservadas, incluidas regresiones de nombre antiguo que siguen validando comportamientos vigentes. No se retira ninguna prueba para obtener PASS. |
| `dist/` HTML/JS/CSS, Diseño 19 y recursos PDF | Código visual y funcional intacto, salvo el SVG sin consumidor. styles.css, atelier.css y gestion-carbon.css se cargan conjuntamente. |
| `dist/assets/images/referencia-cobre-carbon.png` | Todavía referenciada en styles.css. Se mantiene aunque el tema posterior prevalezca visualmente. |
| `technicalsheet/assets/pagina-1.jpg`, pagina-2.jpg y catalogo.json | Los JPG siguen leídos por buildDocument(true) de compatibilidad; el catálogo define controles reales. No se eliminan lectores ni reglas para retirar los recursos. |
| `technicalsheet/assets/diseno19-original.png`, `dist/assets/images/diseno19/` | Logo aprobado usado por Go/frontend, incluidas variantes con rutas públicas distintas. No se deduplican por apariencia o hash sin revisar su contrato. |
| `docs/backend/evidencias/consolidacion-pdf-diseno19/` | Cinco capturas QA de los diseños oficiales conservadas. No son propuestas descartadas ni capturas nuevas de esta etapa. |
| backend/README, ENDPOINTS, EMPLEADOS, CATALOGO_DOCUMENTOS_TRES_ROLES, FLUJO_DIAGNOSTICOS_ASIGNADOS, CONSOLIDACION_PDF_DISENO19, ETAPA2_PRUEBAS_AISLADAS | Instalación, contratos, permisos, PDF históricos y aislamiento. ESTADO_ENDPOINTS queda identificado como regresión de las 71 operaciones originales, no total actual. |
| `docs/backend/limpieza/`, scripts operativos existentes | Herramientas manuales históricas conservadas y etiquetadas; no son parte de esta limpieza ni fueron ejecutadas. No autorizan acceso a datos reales. |

## Dependencias y enlaces comprobados

Validación estática de la copia final: enlaces Markdown locales, imports/URLs JavaScript (incluidos imports evaluados en el navegador respecto de dist), HTML, CSS, rutas públicas de imágenes, Go embed/ReadFile, COPY de Docker y binds Compose. **Sin rutas faltantes.** Los enlaces externos se conservan como referencias; no se afirma haber navegado cada URL externa.

Se comprobaron 91 enlaces Markdown locales, 174 imports JS, 12 URLs de módulos, 4 referencias públicas de assets, 2 URLs CSS, 5 referencias HTML, 12 patrones embed, 3 lecturas explícitas de assets, 5 entradas COPY y 11 binds Compose. La búsqueda cruzada de las rutas retiradas no dejó consumidores en código ni referencias documentales activas; el CSV las enumera como historia, no como enlaces locales.

**224 archivos protegidos conservaron SHA-256:** fuentes funcionales, pruebas, Docker/Compose, SQL/migraciones, recursos actuales y entregables oficiales. No se regeneraron modelos. Comprobación XML: cada detallado conserva 28 tablas, 208 atributos, PK, 50 extremos Crow’s Foot según NULL/UNIQUE y una página; el lógico no muestra tipos. Chen conserva una página y 30 extremos de sus 15 relaciones. Se revisaron también JSON/fuentes de V24 para descartar dependencia de archivos V23 retirados.

## Pruebas ejecutadas: PASS, FAIL y SKIP

El runner `backend/scripts/test-isolated.ps1 -Browser` creó exclusivamente `elbravodeljac-isolated-122113e5cd78444398bd5bc9dfeacf97`: PostgreSQL 18.6 (loopback 53453) y MongoDB 9 (53454), ambos con tmpfs y credenciales efímeras en memoria. Esquema inicial y 001–009 aplicados en la base **nueva y vacía**, sin reutilizar volúmenes. La consulta de solo lectura al catálogo del QA confirmó 28 tablas base, 208 columnas de tablas y 50 FK; otras 8 columnas pertenecen a tres vistas y no se cuentan como atributos de tablas. Solo se reutilizó caché de compilación Go, no resultados (`-count=1`) ni bases.

| Verificación | PASS | FAIL | SKIP / límite |
|---|---:|---:|---|
| JavaScript: verify, workflow, edge-cases, real-api, report-periods, diagnosis-text, isolation, prototype | 8 suites | 0 | 0; demo/unidad/prototipo, no sustituyen integración real |
| `go test ./... -run . -json -count=1 -timeout=20m` | 49 pruebas principales | 0 | 1 prueba: TestCleanRestoredCopy |
| Go incluyendo subpruebas (mismo comando, no otra ejecución) | 112 resultados PASS | 0 | La misma prueba SKIP, sin doble conteo |
| `go vet ./...` | Exit 0 | 0 | Sin omisión |
| Dependencias, enlaces, hashes y modelos | PASS | 0 | No es una nueva auditoría de bases existentes ni de todas las URLs externas |

Siete paquetes Go sin archivos de pruebas se reportan aparte como `[no test files]`; no son siete pruebas aprobadas. Siete paquetes con pruebas terminaron PASS.

Integraciones HTTP/Chrome aprobadas: autenticación/JWT/CSRF/permisos efectivos y último acceso administrativo; empleados/roles; clientes/vehículos/marcas; recepción directa y con diagnóstico asignado; cotización/revisión/aprobación/ampliación; OT/asignación/avances/pruebas; pagos/rectificación/idempotencia/entrega/historial; PostgreSQL/MongoDB/GridFS, validaciones y fallos; Dashboard/reportes; PDFs nuevos e históricos archivados byte por byte. Las regresiones históricas aisladas también pasaron.

**SKIP:** TestCleanRestoredCopy requiere opt-in `INTEGRATION_CLEAN_DATABASE_URL`, una copia QA restaurada y vacía con nombre/puertos específicos y acreditación de ejecución nueva, además del MongoDB correspondiente. No se cumplen esas precondiciones ni se reutiliza un QA antiguo para forzar PASS. No es un bloqueo de compilación, uso corriente o limpieza documental; queda como auditoría específica pendiente. Los runners nativos alternativos y el runner Compose de pruebas dentro de imagen no se repitieron en esta etapa porque sus fuentes/configuración permanecen idénticas; no se presentan como ejecuciones nuevas.

## Estado de los diseños y entregables

Login de dos columnas/Diseño 19 y Dashboard carbón/borgoña: código inalterado y capturas Chrome del QA revisadas, con fixtures ficticios. Cotización borgoña y ficha con tres recuadros/casillas B/A/U: generadores intactos y pruebas PASS; se renderizó e inspeccionó una cotización pendiente y una ficha QA de tres páginas por su contenido, sin recortar ni forzar dos páginas. Las pruebas incluyen estados de cotización y contenido largo. No se rediseña ni se modifica la paginación aprobada.

Balsamiq V14.1, Chen V24, lógico sin tipos, referencia física V24 y SQL de entrega se conservan byte por byte. El README enlaza directamente todos los oficiales, diferencia su finalidad y documenta instalación actual y pruebas aisladas.

## Protección del entorno y evidencias

No se reconstruyó, detuvo ni reemplazó el Docker habitual. La API de localhost:8080 conserva identificador d7f0f06c5eab, imagen y arranque del 5/11/2026; PostgreSQL/MongoDB habituales conservan sus identificadores y fechas de arranque. No se ejecutó SQL ni operaciones MongoDB contra esas bases, ni escrituras HTTP en localhost:8080. No se afirma un nuevo conteo de registros reales: no fue necesario acceder a ellos. Otros QA permanecen intactos; el QA nuevo queda identificado para revisión, sin eliminación de recursos.

Evidencias temporales **ignoradas**, no incluidas: backend/.cache/etapa3-limpieza (inventario/hash/validador/renderizados) y backend/.cache/isolated-122113e5cd78444398bd5bc9dfeacf97 (resultados JS/Go/vet y copia QA). No se agregan contraseñas, tokens, .env privados, respaldos ni datos de clientes reales. El CSV contiene solo rutas/hashes/tamaños/razones.

## Antes y después

| Medida | Antes | Después |
|---|---:|---:|
| Archivos rastreados/previstos para el commit | 341 | 261 |
| Retiradas / guías existentes editadas / documentos nuevos | — | 83 / 15 / 3 |
| Versiones conceptuales ofrecidas como carpetas activas | V23 y V24 | Solo V24 |
| Capturas descartadas y vistas redundantes retiradas | Presentes | 13 retiradas; exportaciones oficiales completas conservadas |
| Bytes de archivos retirados (sin contar reducciones de documentos editados) | 4.532.808 | 0 en copia de trabajo; recuperables desde Git |

Los tres documentos nuevos son este informe, LIMPIEZA_ARCHIVOS.csv y la advertencia de uso manual de docs/backend/limpieza/README.md. Se reducen además guías anteriores para evitar instrucciones de demo/fase inicial presentadas como vigentes. La historia Git no se reescribe. El Pull Request se prepara para revisión; no autoriza merge ni despliegue del Docker habitual.
