# Identidad Diseño 19, diagnóstico unificado e Inicio integrado

## Alcance

Implementación sobre `backend/arquitectura-go-mongodb`, a partir de `ca7ca68`.
No se crean migraciones, tablas, permisos ni cuentas en desarrollo. No se ejecuta
la limpieza operativa anterior. No se modifica Balsamiq ni los modelos académicos.

## Identidad

- El PNG oficial se conserva íntegro en `dist/assets/images/diseno19/original.png`.
  SHA-256: `36c2ddeef8a979c608e764cb7291ca18443266a74cb88a7ead86e05979f55d67`.
- Login: composición horizontal en placa carbón con borde acero para fondo claro.
- Menú: escudo compacto y nombre legible; emblema reducido en menú contraído y favicon.
- Los SVG son contenedores de la imagen raster original, con encuadres proporcionales.
  **No son trazados vectoriales editables ni tienen transparencia artificial.**
  Se conserva el fondo original para evitar halos o reconstrucciones del diseño.
- Paleta borgoña/carbón/acero, áreas de trabajo claras y controles existentes.
- Las cotizaciones nuevas usan el escudo oficial. Los PDF históricos almacenados
  y la plantilla de dos páginas de ficha técnica permanecen intactos.
- Los recursos antiguos se conservan como respaldo, sin utilizarlos en el shell.

## Diagnóstico compatible

La interfaz tiene tres campos editables: Diagnóstico técnico, Reparaciones o
servicios recomendados y Observaciones adicionales opcionales. El motivo de
ingreso sigue visible en solo lectura.

El DTO admite `technicalDiagnosis` (hasta 8200 caracteres, suficiente para unir
los dos textos anteriores de 4000 más sus etiquetas). Se almacena en `resumen`,
con `evaluacion` vacía para nuevas capturas. El contrato anterior `findings` /
`evaluation` sigue aceptándose. No se aceptan ambos contratos simultáneamente.
No se duplican textos para satisfacer una validación oculta.

Al consultar registros antiguos, valores diferentes se muestran juntos y
etiquetados; valores iguales no se duplican. Antes de actualizar un borrador se
conserva una instantánea de ambos textos, recomendaciones y observaciones en el
historial JSON ya existente, dentro de la misma transacción y con bloqueo de fila.
La interfaz permite revisar ese historial. Se conservan las restricciones de
mecánico asignado, versión, concurrencia e inmutabilidad tras finalizar.

## Inicio: Resumen y Análisis

- El menú tiene una única entrada Inicio; se retira Reportes del menú.
- Resumen muestra indicadores operativos, estados, carga por mecánico y bandejas
  de diagnóstico existentes. Los importes requieren `reportes.ver`.
- Análisis reutiliza el componente y consultas de Reportes: servicios realizados,
  ingresos, gráfico, tabla y comparación del período anterior.
- Un filtro compartido Semana/Mes/Año y fecha de referencia usa America/Lima.
  Se almacena el resultado por pestaña y filtro: alternar pestañas no repite
  consultas; aplicar filtros renueva los resultados.
- `#/reportes` redirige a Inicio/Análisis respetando permisos. La búsqueda de
  módulos encuentra Inicio al buscar Reportes, Análisis o Dashboard.
- Se conserva `GET /api/v1/reportes` y el permiso histórico `reportes.ver` en la
  matriz. No se modifican roles ni asignaciones existentes.
- Go elimina `monthlyIncome`, `monthlyIncomeCents` e `income` de `/dashboard`
  cuando no hay `reportes.ver`, e indica `financialVisible:false`. La protección
  se verifica mediante HTTP, además de ocultar los componentes financieros.

## Verificación

Ejecución final del 2026-11-03:

| Comprobación | Resultado |
| --- | --- |
| `node checks/run.mjs` | Seis suites aprobadas |
| `go test ./... -v -count=1` | 41 pruebas principales y 52 subcasos aprobados; ningún fallo |
| `go vet ./...` | Aprobado |
| `TestDesign19DiagnosisCompatibilityAndFinance` | HTTP real: contrato nuevo/anterior, historial, inmutabilidad, 403 y ocultación financiera |
| `TestTallerFourLinkedAccounts` | Chrome con Gerente, Administrador y dos mecánicos; 17 verificaciones integrales aprobadas |
| Filtros | Semana/Mes/Año, fecha histórica, seis peticiones de análisis (actual/anterior por período), sin duplicación al alternar pestañas |
| Flujo de taller | Cliente, vehículo, recepción, diagnóstico, cotización, aprobación, orden, ejecución, pruebas, pagos, entrega y descarga posterior |
| MongoDB | Escáner, GridFS, reintentos, autorizaciones y fallos controlados en QA |
| Visual | Login PC/móvil, menú normal/contraído, favicon, diagnóstico/historial, Resumen y Análisis; imágenes revisadas |
| Desarrollo | Huellas idénticas en las 28 tablas y restricciones PostgreSQL; documentos, índices y validadores MongoDB idénticos antes/después |
| Archivos protegidos | 103 archivos de esquemas/migraciones/modelos idénticos; ninguna credencial privada detectada en los archivos pendientes |

La prueba opcional `TestCleanRestoredCopy` se omitió porque requiere una copia
restaurada vacía expresamente configurada. No se volvió a limpiar desarrollo
para realizarla. Sí se probaron gráficos vacíos en las suites JavaScript,
consultas históricas sin ingresos y PostgreSQL con registros QA.

Hubo una recuperación automática transitoria de PostgreSQL durante un intento;
se consultó su estado y se esperó disponibilidad, sin reiniciarlo ni borrar datos.
Las aserciones nuevas de navegador se ajustaron para contar solo campos editables,
consultar la pestaña activa y respetar el acceso operativo vigente del mecánico.
La ejecución final completa pasó. La compilación inicial de Go en el contenedor
retrasó localhost; ya quedó disponible en `http://localhost:8080/`.

Las capturas nuevas de login/Inicio de desarrollo usan únicamente autenticación
y consultas. El diagnóstico histórico mostrado proviene de QA. La captura
anterior de diagnóstico se conservó de la auditoría previa sobre la misma
versión inicial; las anteriores de login, Inicio y menú se tomaron antes de
esta actualización. Se conservan en `antes/` y `despues/` dentro de la carpeta
local de evidencias. No se publican capturas ni registros privados en Git.
Las evidencias locales se guardan en `backend/.cache/identidad-19/` (excluidas de Git).
Los datos ficticios se generan únicamente en `elbravodeljac_test` y
`elbravodeljac_tecnico_test`, dentro del proyecto QA del 2026-11-03.

## Archivos de esta actualización

41 archivos pendientes: código, recursos oficiales, documentación y adaptación de
pruebas a los tres campos y al traslado de Reportes. Los registros de ejecución,
credenciales e imágenes con datos de desarrollo permanecen excluidos de Git.

### Frontend y marca

- `dist/assets/css/atelier.css`
- `dist/assets/js/app.js`
- `dist/assets/js/components/navigation.js`
- `dist/assets/js/data/permissions.js`
- `dist/assets/js/services/dashboard-service.js`
- `dist/assets/js/services/permissions.js`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/services/workshop-service.js`
- `dist/assets/js/views/administration.js`
- `dist/assets/js/views/operations.js`
- `dist/index.html`
- `dist/assets/images/diseno19/README.md`
- `dist/assets/images/diseno19/emblema.svg`
- `dist/assets/images/diseno19/favicon.svg`
- `dist/assets/images/diseno19/fondo-claro.svg`
- `dist/assets/images/diseno19/fondo-oscuro.svg`
- `dist/assets/images/diseno19/horizontal.svg`
- `dist/assets/images/diseno19/original.png`
- `dist/assets/js/services/diagnosis-text.js`
- `dist/assets/js/views/dashboard.js`

### Backend y pruebas HTTP

- `backend/internal/domain/workshop.go`
- `backend/internal/httpapi/relational.go`
- `backend/internal/platform/postgres/diagnosis_workflow.go`
- `backend/internal/platform/technicalsheet/pdf.go`
- `backend/internal/platform/technicalsheet/quote_logo.go`
- `backend/internal/platform/technicalsheet/quote_test.go`
- `backend/integration/design19_test.go`
- `backend/internal/platform/technicalsheet/assets/diseno19-original.png`

### Verificaciones JavaScript y navegador

- `checks/attention-browser.mjs`
- `checks/audit-browser.mjs`
- `checks/clean-copy-browser.mjs`
- `checks/diagnosis-browser.mjs`
- `checks/frontend-browser.mjs`
- `checks/quotes-browser.mjs`
- `checks/review-browser.mjs`
- `checks/run.mjs`
- `checks/taller-browser.mjs`
- `checks/verify.mjs`
- `checks/diagnosis-text.mjs`

### Documentación

- `docs/backend/DISENO_COTIZACION_PDF.md`
- `docs/backend/IDENTIDAD19_DIAGNOSTICO_DASHBOARD.md`

