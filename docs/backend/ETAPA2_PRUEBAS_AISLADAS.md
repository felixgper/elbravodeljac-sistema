# Etapa 2 — Pruebas vigentes y aislamiento

6 de noviembre de 2026. Rama `feature/aislamiento-pruebas-vigentes`, creada desde `main` sincronizada en `c824414`, después de fusionarse el PR #5. El cambio se limita a herramientas, fixtures, pruebas y documentación de pruebas. No modifica el aplicativo, sus generadores PDF, SQL, migraciones, modelos ni diseños.

## Qué estaba desactualizado y qué se corrigió

| Archivo o grupo | Problema anterior | Corrección |
|---|---|---|
| `backend/scripts/test-local-postgres.ps1` | Lista manual detenida en 001–006; podía heredar URI MongoDB, flags o una copia restaurada de otra sesión. | Clúster nuevo y puerto aleatorio; esquema técnico inicial y **una llamada al migrador oficial Go con 001–009**. Limpia/restaura conexiones y flags heredados, configuración PG y GOWORK; `psql -X`. |
| `test-local-databases.ps1` | Delegaba en el fallback incompleto y repartía la propiedad de los procesos. | Un único propietario crea ambos clústeres nativos nuevos. Comprueba el PID dueño del puerto MongoDB y cierra solo sus procesos, restaurando el entorno incluso al fallar. |
| `test-compose.ps1` | Nombre Compose fijo: una ejecución podía reutilizar o detener un QA anterior. | Proyecto UUID nuevo, ausencia comprobada antes de iniciar; su `finally` retira solo el proyecto creado por esa ejecución. No cambia la imagen habitual. |
| `test-audit.ps1` | Migraba/reutilizaba QA existentes y leía configuración privada histórica. | Alias del runner nuevo con navegador. Rechaza los antiguos parámetros de reutilización antes de ejecutar Docker/SQL. |
| `test-isolated.ps1` (nuevo) | Faltaba una entrada reproducible para integrar navegador y bases nuevas sin puertos fijos. | PostgreSQL/MongoDB `tmpfs` nuevos, puertos loopback aleatorios, copia temporal de fuentes, Go/JS/vet y flags explícitos. No guarda contraseñas ni reutiliza volúmenes. |
| Cinco browser de marcas, diagnóstico, cotización, revisión y atención | Leían `.env.fase-e.local` y escribían directamente en `localhost:8080`. | Fixtures por stdin, origen de prueba acreditado y `TestLegacyBrowserRegressionsIsolated`. Se conservan sus escenarios; no se eliminan por antigüedad. |
| `live-docker.mjs`, `read-persisted-docker.mjs` | Configuración privada, Auxiliar nuevo y expectativas antiguas; se afirmaba reinicio sin ejecutarlo. | Mismo harness aislado, roles vigentes, aprobación, ejecución/conformidad y consulta desde otra sesión del mismo QA. El informe debe coincidir en origen e identificador de ejecución; no se afirma reiniciar servicios. |
| Los seis browser que ya utilizaban harness Go | Algunos solo comprobaban un puerto local; Carbon/Sheet exigían 25440/25439, ocupados por QA previos. | Todos verifican la acreditación antes de Playwright. Los dos puertos fijos se sustituyen por loopback no habitual **más prueba de pertenencia a la ejecución nueva**. |
| `browser_isolation_test.go`, `isolated-fixture.mjs`, `isolation.mjs` (nuevos) | No había una guarda común verificable. | Prueba del identificador de la base, cabecera exclusiva del servidor de test y rechazo previo a cualquier credencial/navegador. Las 13 entradas se prueban negativamente contra 8080. |
| `verify.mjs`, `prototype.mjs`, `run.mjs` | La prueba del Dashboard comparaba un objeto actual con una muestra fija no relacionada del prototipo; los mensajes mezclaban categorías. | Se comprueba el store demo vigente y su aislamiento. El contrato Balsamiq se conserva en una suite histórica independiente. El coordinador etiqueta las ocho suites. |

El fallback nativo no aplica SQL de migraciones por segunda vía. Las repeticiones deliberadas de `Migrate` dentro de pruebas de idempotencia o en bases nuevas propias de ciertos casos se conservan: no son una segunda inicialización accidental del runner. No se sustituyó `esquema_actual.sql` por `esquema_entrega.sql`.

## Expectativas históricas actualizadas

Las primeras ejecuciones hicieron visibles diferencias reales de los fixtures con el aplicativo vigente:

- Cotizaciones aprobadas sin nombre/medio de aprobación: ahora se completan esos campos. Al cambiar conceptos/precios se comprueba la invalidación y se registra una nueva aprobación antes de crear la orden.
- Actualizaciones sin versión: se envía la revisión actual. La prueba de bloqueo por OT utiliza un payload válido para comprobar el 409 correspondiente, en lugar de recibir antes un 422 de validación.
- Alta de Auxiliar/Supervisor y asignación auxiliar nueva: los escenarios actuales registran mecánicos y prueban roles N Mecánico/Administrador. Se comprueba la ausencia de controles para nuevas asignaciones auxiliares; las pruebas Go de preservación de relaciones/asignaciones históricas permanecen.
- Marcar trabajos con un control antiguo `performed_`: se utiliza el selector vigente `execution_` y se conserva la exigencia de prueba satisfactoria y conformidad de entrega.
- Login de varios casos bajo un único servidor: se agotaba el límite real de cinco intentos/minuto por IP. Cada escenario tiene ahora su propio servidor; no se debilitó el limitador.
- Historial con dos botones válidos para la misma ficha: la prueba elige explícitamente el primero y compara sus bytes con la descarga de Pagos.
- Reportes disparan consultas del período actual y su comparación histórica en paralelo. Una espera genérica capturaba a veces el mes anterior (31 días frente a 30). La regresión exige los parámetros exactos de período/fechas y concilia el total de la respuesta correcta.
- La integración inicial de la nueva guarda omitió el identificador en el fixture Chrome de ficha. Se corrigió el harness, sin modificar el aplicativo.

Se mantuvieron los rechazos, concurrencia, validaciones y aserciones de negocio. No se cambió la API para conseguir resultados exitosos.

## Resultados ejecutados

Los números son pruebas Go de primer nivel; no suman subtests ni duplican ejecuciones.

| Ejecución | PASS | FAIL | SKIP | Alcance |
|---|---:|---:|---:|---|
| Runner final `test-isolated.ps1 -Browser` | **49** | **0** | **1** | 50 casos Go; PostgreSQL 18.6 y MongoDB 9 nuevos, con Chrome. `go vet ./...`: PASS. |
| `test-compose.ps1`, Go dentro de imagen QA nueva | 43 | 0 | 7 | Sin flags Chrome. Esas seis pruebas de navegador y la copia restaurada se omiten explícitamente. |
| `test-local-postgres.ps1`, versión final | 42 | 0 | 8 | PostgreSQL nativo nuevo; sin MongoDB ni flags Chrome/copia restaurada. |
| `test-local-databases.ps1`, versión final | 43 | 0 | 7 | PostgreSQL/MongoDB nativos nuevos; sin navegador/copia restaurada. |
| `node checks/run.mjs` | 8 suites | 0 | 0 | Demo, unidad y prototipo etiquetados; sin persistencia PostgreSQL implícita. |

Se ejecutaron además el rechazo de parámetros históricos de `test-audit.ps1`, el análisis sintáctico PowerShell y la prueba de restauración de variables tras los runners nativos. Se suministraron como sentinelas destinos heredados deliberadamente inválidos para MongoDB/copia restaurada y flags de navegador: se ignoraron durante el fallback y se restauraron al retornar. Ninguno era una credencial real.

**Historial de fallos conservado:** primer barrido integral, 47 PASS/2 FAIL/1 SKIP; segundo, 48/1/1; repetición dirigida histórica, 0/1/0; cuarto barrido, 48/1/1. Esos fallos se debían a los fixtures/esperas descritos arriba. El quinto barrido integral es el resultado final 49/0/1. Los logs fallidos no se eliminaron ni se presentan como aprobados. La prueba positiva de la guarda se mantuvo como unidad con respuesta HTTP simulada; la acreditación HTTP verdadera se verificó por separado en Go/Chrome.

### Recorridos cubiertos en la ejecución final

- Login por correo/usuario, contraseña incorrecta, logout, cookies HttpOnly, CSRF, rate limit y permisos efectivos.
- Dashboard carbón/borgoña: indicadores contra PostgreSQL, períodos, análisis, responsive y restricciones financieras para mecánicos.
- Clientes, vehículos, marcas/búsqueda/edición, historial y restricciones de integridad.
- Recepciones directas/técnicas, mecánico responsable, borrador/finalización, aislamiento entre cuentas y cotización elegible.
- Cotizaciones: descripción libre, referencias existentes, versiones, aprobación, edición/F5, bloqueo por OT y PDF pendiente/aprobado. Las pruebas PDF cubren también Rechazada, varios conceptos y contenido extenso.
- Órdenes, roles N/cargo laboral, asignación actual, ejecución, avances, pruebas y cierre; transacciones y errores/concurrencia.
- Pagos, reintentos, rectificación, saldo, entrega y conformidad; fichas archivadas byte por byte sin UPDATE y compatibilidad sin archivo.
- MongoDB/GridFS: escáner/evidencias, carga/consulta/descarga, reintentos, autorización PostgreSQL, fallos y compensaciones. Consulta posterior desde otra sesión del QA.

Las 14 pruebas PDF del paquete `technicalsheet` pasaron en las suites generales. También se inspeccionaron capturas reales de Login/Dashboard y renderizados de la cotización borgoña y controles de ficha descargados del QA final. Los 220 archivos protegidos revisados de frontend, backend funcional, SQL y modelos no presentan diferencias contra la base de esta rama. Balsamiq y los modelos no se movieron ni regeneraron.

## Pendiente y límites

Solo queda omitida en el barrido integral **`TestCleanRestoredCopy`**: exige una copia restaurada específica `jac_restore_<fecha>` en 25435, su correspondiente MongoDB, acreditación de ejecución y acceso privado previo. No se restauró ninguna base ni se leyó ese acceso para satisfacerla. No bloquea los recorridos de bases nuevas ejecutados aquí; su resultado es SKIP, no PASS.

Los comandos sin navegador no acreditan Chrome. Los casos demo y `fetch` simulado no acreditan bases ni seguridad del servidor. Las evidencias corresponden a fixtures ficticios y versiones de herramientas ensayadas, no a una prueba de despliegue de producción. El Docker habitual de `localhost:8080` no fue actualizado.

## Cómo se evita afectar datos existentes

1. UUID de proyecto/directorio nuevo y comprobación de ausencia antes de crear servicios; ninguna reutilización de los QA anteriores.
2. Docker usa solo `tmpfs` y publicaciones loopback aleatorias. El nativo usa un directorio PostgreSQL nuevo y un MongoDB nuevo cuyo puerto pertenece al PID creado.
3. Los runners limpian/restauran las variables heredadas de conexiones y opt-ins. Sus DSN apuntan exclusivamente a `elbravodeljac_test`/`elbravodeljac_tecnico_test` recién preparados; las credenciales efímeras no se guardan en Git.
4. El preparador marca únicamente su base nueva con `jac.qa_run_token`. El harness lo comprueba por lectura antes de crear fixtures y su servidor de pruebas anuncia `X-Bravo-Test-Run` en `/healthz`. JavaScript verifica el identificador antes de cargar Playwright o leer informes.
5. `localhost:8080`, otros hosts, credenciales en URL, rutas/query/fragmentos y redirecciones en esa comprobación se rechazan. Un cambio aislado de puerto o una bandera JavaScript no acredita al servidor habitual.
6. Fixtures por stdin y copia temporal de fuentes sin archivos ignorados: no se lee `.env.fase-e.local` ni se genera el anterior archivo privado de acceso de diagnósticos.
7. La cabecera/middleware de acreditación existe solo en archivos `_test.go`; no añade endpoints, permisos o funcionalidades al backend empresarial. JWT/CSRF y las restricciones originales siguen verificándose.

Es una defensa contra errores de configuración mediante las entradas soportadas. No pretende impedir que un operador altere deliberadamente las pruebas o marque manualmente otra base. No apuntar variables de integración a instalaciones existentes.

## QA de la etapa anterior y recursos de esta ejecución

Se comprobaron etiquetas Compose, directorio/configuración de origen y montajes, sin leer variables secretas:

| Proyecto / servicios previos | Puertos | Propiedad y almacenamiento | Decisión |
|---|---|---|---|
| `elbravodeljac-pdf-qa-b2309371`: postgres-test y mongo-test | 25440 / 27040 | Este repositorio, `compose.test.yaml` y override ignorado de consolidación PDF; solo `tmpfs`, sin volúmenes persistentes de datos. | **Conservados**. Pueden cerrarse cuando ya no se necesiten sus fixtures de revisión; detenerlos descarta esos datos ficticios en memoria. |
| `elbravodeljac-pdf-sheet-qa-3fde4f77`: postgres-test | 25439 | Mismo repositorio y override específico de ficha; `tmpfs`, sin volumen persistente de datos. | **Conservado**, con la misma condición. |
| `elbravodeljac-fase-e-dev`: API/PostgreSQL/MongoDB | 8080 y red privada | Entorno habitual con sus volúmenes. | **Intacto**, sin reconstrucción, reinicio ni migración. |

El QA final nuevo es `elbravodeljac-isolated-96714022f95941ca9d9a0b9edc07abfe`, PostgreSQL 53986 y MongoDB 53987; permanece iniciado para revisión de sus datos ficticios. Los ocho procesos de los cuatro QA fallidos creados por esta etapa se detuvieron **después** de comprobar sus UUID, configuración de este repositorio, `tmpfs` y ausencia de volúmenes de datos. No se eliminaron esos contenedores, imágenes ni archivos; se conservan sus logs. El runner Compose independiente retiró solo su proyecto nuevo comprobado, sin `down -v` ni limpiezas generales. Los procesos nativos nuevos se cerraron en su `finally`.

No se detuvo ningún QA anterior ni servicio ajeno. No se ejecutó `docker prune`, `git clean`, restauración, limpieza de mapas/prototipos/documentación ni operaciones contra las bases reales.

## Repetición y evidencias

La [guía de pruebas](../../checks/README.md) clasifica todas las entradas y documenta los comandos. Evidencia local ignorada:

- `backend/.cache/isolated-96714022f95941ca9d9a0b9edc07abfe/`: `resultado.json`, Go JSON, vet, ocho suites JS y, bajo la copia del proyecto, capturas/PDF/regresiones Chrome.
- `backend/.cache/etapa2-aislamiento/native/`: resultados de los dos fallbacks y comprobación del entorno heredado.
- `backend/.cache/etapa2-aislamiento/compose-only.log`: ejecución y retirada del proyecto Docker propio.
- `backend/.cache/etapa2-aislamiento/inventario-docker.json`, `qa-fallidos-cerrados.json`, `integridad-aplicativo.json` y `render/`: propiedad, cierre acotado, protección de fuentes y revisión visual.

Estas carpetas no se incluyen en el commit. No se publican contraseñas, tokens, respaldos, registros personales ni documentos históricos. El Pull Request se prepara para revisión; no autoriza merge ni despliegue.
