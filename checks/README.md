# Pruebas del aplicativo vigente

Las categorías no son intercambiables: un store demo o un `fetch` simulado no acredita persistencia ni permisos de PostgreSQL. Los nombres históricos de archivos se conservan para mantener su trazabilidad.

| Entrada | Categoría | Qué valida / fuente |
|---|---|---|
| `verify.mjs` | Demostración | Sesión demo y Dashboard derivado del store actual; aislamiento de objetos devueltos. Ya no compara el Dashboard actual con la muestra fija del prototipo. |
| `workflow.mjs`, `edge-cases.mjs` | Demo / unidad | Recorrido del taller, límites y cálculos en memoria/localStorage, con datos ficticios. |
| `real-api.mjs` | Unidad con `fetch` simulado | DTO, cookies/CSRF, errores y propagación de permisos; sin conexión a una base real. |
| `report-periods.mjs`, `diagnosis-text.mjs` | Unidad | Períodos, conciliación y compatibilidad de texto. |
| `isolation.mjs` | Unidad de seguridad de pruebas | Rechaza destinos inseguros antes de hacer HTTP; prueba la acreditación con una respuesta simulada y bloquea las 13 entradas de navegador antes de cargar Playwright o credenciales. |
| `prototype.mjs` | Contrato histórico | Balsamiq V14.1, 23 pantallas y correspondencia con navegación; los valores ilustrativos no se presentan como indicadores actuales. No modifica el prototipo. |
| `audit-browser.mjs`, `frontend-browser.mjs`, `carbon-design-browser.mjs`, `taller-browser.mjs`, `sheet-functional-browser.mjs` | Integración HTTP + Chrome | Los ejecutan sus harness Go con credenciales ficticias, Go real y bases temporales. Login/Dashboard, módulos, permisos, controles, pagos, entrega, MongoDB y PDF. |
| `brands-browser.mjs`, `diagnosis-browser.mjs`, `quotes-browser.mjs`, `review-browser.mjs`, `attention-browser.mjs` | Regresiones históricas actualizadas | `TestLegacyBrowserRegressionsIsolated` conserva y ejecuta estos escenarios con la interfaz y reglas vigentes. |
| `live-docker.mjs`, `read-persisted-docker.mjs` | Persistencia en QA | Mismo harness temporal: recorrido y GridFS, seguido por otra sesión que consulta los registros. El nombre histórico se conserva; ya no leen `.env.fase-e.local` ni apuntan al Docker habitual. No afirman un reinicio de servicios que no ejecutan. |
| `clean-copy-browser.mjs` | Auditoría histórica específica | Solo mediante su harness y una copia QA explícita; requiere precondiciones adicionales. No forma parte de la inicialización corriente ni debe usar una instalación real. |

## Comandos

`node checks/run.mjs` ejecuta ocho suites de demo/unidad/prototipo. El coordinador imprime su categoría. No inicia bases ni ejecuta los escenarios de navegador de escritura.

Para integración nueva y aislada desde PowerShell en la raíz:

```powershell
./backend/scripts/test-isolated.ps1
# Con Node, Go, Chrome y Playwright configurados:
$env:PLAYWRIGHT_MODULE = 'ruta/a/playwright'
$env:BRAVO_BROWSER = 'C:/Program Files/Google/Chrome/Application/chrome.exe'
./backend/scripts/test-isolated.ps1 -Browser
```

El runner crea otro proyecto Compose con UUID, PostgreSQL/MongoDB `tmpfs`, puertos loopback aleatorios y una copia temporal de las fuentes. Inicializa la base vacía con `esquema_entrega.sql`. El migrador detecta su marca, prepara el catálogo documental y no vuelve a aplicar 001–009. Las pruebas de compatibilidad que parten de `000_base_legacy.sql` sí ejercitan las migraciones 001–009. Usa contraseñas efímeras en memoria, guarda evidencia ignorada y deja el QA identificado para revisión; no detiene otros proyectos. `-TestPattern` permite elegir escenarios. `-GoBuildCache` permite reutilizar únicamente compilaciones, nunca bases o volúmenes; `-count=1` evita reutilizar resultados de pruebas.

Alternativas existentes conservadas:

- `backend/scripts/test-compose.ps1`: ejecuta Go dentro de una imagen de pruebas y crea un proyecto nuevo sin puertos publicados. Al terminar retira solamente ese proyecto cuya ausencia comprobó antes de crearlo. No compila/reemplaza la imagen del aplicativo habitual. Las suites Chrome no se habilitan en esa entrada.
- `backend/scripts/test-local-postgres.ps1`: crea un clúster PostgreSQL nativo nuevo, puerto aleatorio y esquema inicial; llama una vez a `go run ./cmd/db-migrate` para preparar el catálogo documental. No mantiene otra lista SQL de migraciones. Después ejecuta Go y cierra únicamente ese clúster. Requiere PostgreSQL 18, Go y Node.
- `backend/scripts/test-local-databases.ps1`: delega al mismo propietario y añade un proceso MongoDB nativo nuevo con directorio/puerto propios; verifica que el puerto pertenece al PID recién creado. `-Browser` habilita además los harness actuales.
- `backend/scripts/test-audit.ps1`: alias histórico del runner nuevo con navegador. Rechaza los antiguos parámetros de reutilización antes de consultar Docker o ejecutar SQL.

Los runners limpian/restauran sus variables de conexión y opt-ins del proceso. No heredan un `INTEGRATION_CLEAN_DATABASE_URL`, una URI MongoDB ni flags de navegador de otra ejecución. El nativo desactiva también configuración PostgreSQL de entorno y usa `psql -X`, sin ejecutar el `.psqlrc` del usuario.

## Acreditación del navegador

Todos los scripts de navegador exigen un fixture por stdin, un origen exacto `http://127.0.0.1:puerto` distinto de 8080 y un identificador aleatorio de ejecución. Rechazan usuario/contraseña en URL, rutas, query, fragmentos, orígenes externos y redirecciones durante la comprobación.

El harness Go comprueba por consulta de lectura que la base de pruebas fue marcada por el preparador con el identificador de esta ejecución (`jac.qa_run_token`). Solo entonces su servidor HTTP de pruebas anuncia `X-Bravo-Test-Run` en `/healthz`; JavaScript lo verifica antes de abrir Chrome. El aplicativo de producción no contiene ese middleware ni anuncia esa cabecera. Cambiar solo una variable JavaScript o indicar otro puerto no acredita el servidor habitual.

Cada escenario histórico tiene su propio servidor y limitador de Login. No se debilita ni se reinicia el limitador del aplicativo para conseguir resultados exitosos. Las credenciales ficticias viajan por stdin; no se guardan archivos de acceso. La consulta de persistencia exige el mismo origen e identificador resumido del QA que generó su informe.

Esta protección previene conexiones accidentales mediante los runners soportados; no sustituye JWT/CSRF/permisos ni es una barrera contra un operador que altere deliberadamente el harness o marque manualmente otra base. No apuntar variables de integración a instalaciones existentes.

`go test ./...` sin DSN o flags omite las integraciones correspondientes. **SKIP significa no ejecutada**, no aprobada.
