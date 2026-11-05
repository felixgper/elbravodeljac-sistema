# Correcciones de las pruebas previas a consolidación

Fecha: 3 de noviembre de 2026. Rama: `backend/arquitectura-go-mongodb`.

## Cambios limitados al entorno y al fixture

1. `backend/Dockerfile.test` copiaba backend y dist, pero no el archivo que `TestCatalogRolesMigrationAndHTTP` abre mediante `../../database/postgresql/esquema_actual.sql`. Se añade una única instrucción COPY desde ese archivo oficial a `/workspace/database/postgresql/esquema_actual.sql`. Las migraciones 001–009 ya estaban incluidas/embebidas en Go y montadas por Compose: no faltaban ni se duplican. No se cambian Compose, SQL ni estructura PostgreSQL.
2. `backend/integration/phase_c_test.go` suponía otros administradores de un clúster acumulado. Ahora crea una base con nombre aleatorio dentro del clúster explícito `elbravodeljac_test`, carga el esquema oficial y llama a `Store.Migrate`, sin copiar definiciones SQL ni mantener otra lista de migraciones. Así controla sus cuentas sin alterar las de otras suites. La base adicional no se elimina mediante SQL; desaparece junto al clúster efímero creado para la ejecución.

El fixture verifica dos administradores iniciales, retiro permitido de permisos cuando queda otro acceso válido, rechazo de escalamiento, rechazo y rollback al retirar la última matriz, desactivar/eliminar la última cuenta o quitarle su rol. Usa un operador con permisos de edición/eliminación/roles pero sin usuarios.ver para comprobar que no sustituye un acceso administrativo completo. La cuenta usada para probar DELETE no tiene referencias operativas, por lo que un error de FK no puede ocultar la validación de último administrador.

Se restauran permisos mediante la API autorizada. Se reautentican las cuentas reactivadas y se prueba concurrencia: exactamente una desactivación prospera y queda un administrador válido. La reautenticación usa otra instancia HTTP de pruebas con el mismo store y clave; conserva el límite real de cinco intentos por ventana y las validaciones de JWT. No se modifica la lógica del aplicativo para hacer pasar las pruebas.

## Reproducción

El script oficial `backend/scripts/test-compose.ps1` ejecuta `go test ./... -count=1 -v` en Compose, con contraseña efímera, PostgreSQL/MongoDB de pruebas y red interna. No dirigirlo a desarrollo ni producción.

En esta revisión se usó `compose.test.yaml` con nombres de proyecto aleatorios nuevos. Las bases eran tmpfs, sin volúmenes persistentes ni puertos publicados. Para las tres pruebas Chrome se crearon otros proyectos nuevos, con puertos localhost automáticos y red propia. No se usó localhost:8080 ni se reutilizaron los clústeres QA existentes. Se retiraron únicamente los proyectos nuevos al terminar, sin limpiezas globales de Docker.

## Resultados

- Seis scripts JavaScript de `checks/run.mjs`: pasaron. Se ejecutó Node directamente, equivalente al script check de package.json; npm no está disponible en PATH.
- `go test ./... -count=1` local, sin variables de integración: pasó; las pruebas que requieren servicios se omiten en esa ejecución.
- Fase C: `go test ./integration -run '^TestPhaseCRelationalHTTP$' -count=3 -shuffle=on -v` pasó tres veces, con bases adicionales independientes.
- La primera repetición de Compose confirmó resuelto el archivo SQL faltante y detectó un 429 del fixture ampliado al reautenticar. Se corrigió exclusivamente la ventana de pruebas descrita arriba y se repitió desde una imagen nueva. La evidencia de esa ejecución inicial se conserva; no se presenta como aprobada.

- Segunda ejecución completa: `go test ./... -count=1 -v` en la imagen reconstruida y clúster nuevo pasó, con **42 pruebas principales aprobadas, cero fallidas y seis omitidas**. Entre las aprobadas están la migración del catálogo/roles, Fase C, el último Gerente, JWT/CSRF, el recorrido del taller, MongoDB/GridFS, cotizaciones PDF y fichas técnicas. No se cuentan las subpruebas como pruebas principales adicionales.
- De las seis omitidas en Compose, tres se ejecutaron aparte y pasaron con Chrome y otros clústeres nuevos, según la tabla siguiente. Resultado combinado: 45 de las 48 pruebas principales disponibles verificadas y tres opcionales pendientes, sin contar las repeticiones de Fase C como cobertura adicional.
- `go vet ./...` sobre el código final: pasó localmente y en la imagen nueva de pruebas. La revisión de diferencias no presenta errores de whitespace.

## Las seis pruebas anteriormente omitidas

| Prueba | Resultado de esta revisión / condición |
| --- | --- |
| TestPhaseEFrontendBrowserPostgreSQLMongoDB | Pasó con FRONTEND_E2E=1, Chrome, Node/Playwright y bases nuevas. Flujo completo, 15 vistas, permisos, CSRF, MongoDB/GridFS y persistencia. |
| TestAuditBrowserIsolated | Pasó con FRONTEND_AUDIT=1 y otro clúster nuevo; auditoría integral del navegador y API. |
| TestTallerFourLinkedAccounts | Pasó con FRONTEND_AUDIT=1 y otro clúster nuevo; vínculos laborales, actores, diagnóstico, auxilios y entrega archivada. |
| TestCarbonDesignBrowser | No ejecutada: exige CARBON_DESIGN_QA=1 y PostgreSQL nuevo en 127.0.0.1:25440. Ese puerto está ocupado por un QA previo; no se reutiliza ni detiene. Requiere una ventana con el puerto libre o una futura parametrización revisada del harness. |
| TestSheetFunctionalAudit | No ejecutada: exige SHEET_FUNCTIONAL_AUDIT=1 y PostgreSQL nuevo en 127.0.0.1:25439, ocupado por otro QA previo. SHEET_BROWSER_AUDIT=1 habilita además su formulario Chrome. Las pruebas ordinarias de PDF/entrega son distintas y no se cuentan como esta auditoría. |
| TestCleanRestoredCopy | No ejecutada: exige INTEGRATION_CLEAN_DATABASE_URL a una copia QA jac_restore con timestamp en 25435, MongoDB en 27032 y el acceso privado esperado por el fixture. Los puertos están ocupados y no se restaura ni utiliza una cuenta existente para habilitarla. Verifica una operación histórica de limpieza/restauración, no la inicialización ordinaria desde cero. |

Las tres pruebas opcionales pendientes siguen pendientes: no se cambiaron guardas ni se marcaron como aprobadas. Las pruebas de creación de bases nuevas y migraciones se verifican por las suites ordinarias de integración y los tres clústeres nuevos de navegador, sin restaurar respaldos empresariales.

## Alcance de protección e integración

Ningún cambio en Login, Dashboard, diseño de módulos, PDFs, modelos, migraciones, reglas, roles/permisos o cuentas reales. Las huellas de los 323 archivos versionados previos solo difieren en Dockerfile.test y phase_c_test.go; este informe es el único archivo documental añadido.

Main sigue siendo antecesora del backend publicado: 0 commits exclusivos de main frente a 28 de backend al efectuar fetch. No se anticipan conflictos de integración para esas refs. Las correcciones locales deben revisarse, guardarse y publicarse antes de abrir un PR que realmente las incluya. No hubo commit, push, merge, borrado de ramas ni despliegue.

Archivos propuestos para revisión antes de commit/push: `backend/Dockerfile.test`, `backend/integration/phase_c_test.go` y este informe. Los dos fallos reproducibles están resueltos y las pruebas pertinentes pasaron; las tres auditorías opcionales pendientes no se presentan como cobertura realizada. La apertura del PR queda a la espera de la revisión solicitada por el usuario, ya que el remoto todavía no contiene estas correcciones locales.

Los logs, fixtures y herramientas de ejecución quedan en `backend/.cache/correccion-integracion/`, ignorados por Git. No deben publicarse como código ni incluir contraseñas/cookies en evidencias. Los secretos de bases de prueba se generaron en memoria y se redactaron de sus logs.
