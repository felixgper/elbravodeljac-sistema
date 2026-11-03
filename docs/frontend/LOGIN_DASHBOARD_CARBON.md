# Login e Inicio: presentación carbón

Implementado sobre `backend/arquitectura-go-mongodb`, partiendo del commit
`0be180a` y un árbol de trabajo limpio. No se realizó commit, push ni merge.

## Alcance visual

- Login de dos columnas: identidad oficial Diseño 19 a la izquierda y tarjeta
  de acceso a la derecha, fondo carbón con circuitos decorativos discretos.
- Título Bienvenido, campos visibles Correo o usuario/Contraseña, botón Iniciar
  sesión y control accesible de mostrar/ocultar contraseña.
- Se conserva el identificador recordado previamente y el código existente de
  autenticación. No se añaden registro, recuperación de contraseña ni nuevos
  servicios. El recordatorio anterior no se expone como un tercer campo visible.
- Menú lateral grafito y barra superior oscura, con los módulos y permisos
  existentes. Se conservan búsqueda de módulos, perfil y cierre de sesión.
- Inicio oscuro, con Resumen/Análisis, filtros, KPI, gráficos y tablas existentes.
  Se presenta también el mismo conteo de estados como tabla legible, usando la
  respuesta real que ya alimenta el gráfico. No se agrega un endpoint.
- Formularios de otros módulos conservados claros. El tema de Inicio está
  limitado a `.dashboard-surface .workspace`; no cambia variables de `:root`.
- Inter para interfaz y Barlow Condensed para encabezados/indicadores, fuentes
  que ya utilizaba el proyecto, con alternativas locales si no cargan.

Las imágenes de Nano Banana se usaron únicamente como referencias de diseño.
No se copiaron sus clientes, importes, órdenes ni progresos. Se conservan los
cuatro indicadores disponibles: ingresos, órdenes activas, servicios finalizados
y auxilios. No se añade «vehículos en taller» ni porcentajes sin una fuente
equivalente existente.

La semántica original de los filtros permanece intacta: las órdenes activas son
el resumen actual global del backend; ingresos, servicios y movimientos utilizan
el período consultado. El cambio visual no modifica esas consultas.

## Archivos

| Archivo | Finalidad |
|---|---|
| `dist/index.html` | Carga el CSS específico después de los estilos existentes |
| `dist/assets/css/gestion-carbon.css` | Login, navegación y superficies exclusivas de Inicio; responsive |
| `dist/assets/images/gestion-circuitos.svg` | Decoración vectorial, sin datos ni texto de la referencia |
| `dist/assets/js/app.js` | Composición visual del login, marca y clases de alcance del shell |
| `dist/assets/js/views/dashboard.js` | Tabla de los estados reales ya disponibles, conservando gráficos y cálculos |
| `checks/carbon-design-browser.mjs` | Comprobaciones y capturas reales en Chrome |
| `backend/integration/carbon_design_test.go` | Servidor HTTP y cuentas ficticias en PostgreSQL exclusivo de QA; sin código nuevo de producción Go |
| `checks/taller-browser.mjs` | Actualiza el selector de la imagen del login para su nueva ubicación |
| `docs/frontend/LOGIN_DASHBOARD_CARBON.md` | Este informe |

No se modificaron los estilos comunes `styles.css`/`atelier.css`, servicios de
autenticación o datos, backend de producción, PostgreSQL/MongoDB existentes,
migraciones, modelos, permisos ni generadores o muestras PDF aprobadas.
Se verificaron las huellas de los 318 archivos versionados iniciales: solo
cambiaron los cuatro archivos anteriores listados; los demás permanecen iguales.
Los archivos nuevos están enumerados en la tabla.

## Pruebas realizadas

Cluster nuevo `elbravodeljac-carbon-qa`, separado de desarrollo:

- PostgreSQL 18.6 en `127.0.0.1:25440`, base `elbravodeljac_test`.
- MongoDB 9.0 en `127.0.0.1:27040`, base `elbravodeljac_tecnico_test`.
- Cuentas y empleados ficticios con roles canónicos Administrador, Gerente y
  Mecánico. Contraseñas aleatorias, sin cambiar permisos de esos roles para el
  caso nuevo. La configuración privada está excluida de Git.
- No se utilizó `ElBravoDelJac` del puerto 5432 del host. No se detuvieron,
  recrearon ni limpiaron bases, contenedores o volúmenes anteriores.

| Verificación | Resultado |
|---|---|
| Login por correo válido | Aprobado contra Go/PostgreSQL |
| Login por usuario válido | Aprobado contra Go/PostgreSQL |
| Contraseña incorrecta | Error visible, sin entrada al aplicativo |
| Mostrar/ocultar contraseña | Aprobado; tipo y estado accesible cambian correctamente |
| Administrador, Gerente y Mecánico | Acceso propio y navegación correspondiente aprobados |
| Cerrar sesión y volver a ingresar | Aprobado mediante el flujo existente |
| Indicadores reales | La creación de una orden por API aumenta el KPI en uno; coincide con PostgreSQL |
| Sin registros | Comprobado al iniciar el cluster vacío; las repeticiones comprueban períodos sin movimientos sin borrar las órdenes ya creadas |
| Semana, Mes, Año y fecha | Peticiones reales y gráficos de Análisis aprobados |
| Finanzas del mecánico | KPI ausente, Análisis deshabilitado y API de reportes devuelve 403 |
| Navegación y otros módulos | Comparación de estilos computados con/sin el CSS nuevo: formularios, campos y tablas conservados |
| Responsive | 1440×900, 820×1180 y 390×844; sin desplazamiento horizontal global; menú móvil abre y cierra |
| JavaScript | Sin errores de ejecución en las páginas comprobadas |
| Regresión completa | Seis suites JavaScript y Go con navegadores/PostgreSQL/MongoDB: aprobados |
| Análisis estático | `go vet ./...` y `git diff --check`: sin errores |

La ejecución completa de `go test ./... -v -count=1` obtuvo **44 pruebas
principales y 53 subcasos aprobados, cero fallos**. Se omitieron explícitamente:

- `TestCleanRestoredCopy`: necesita otra copia restaurada y su configuración;
  no se restauró una base para esta tarea visual.
- `TestSheetFunctionalAudit`: exige su cluster independiente del puerto 25439
  y habilitación específica. No se cambió su protección para apuntarlo a 25440.
  Las demás pruebas existentes de ficha/cotización sí se ejecutaron.

Tras el último ajuste de color se repitió `TestCarbonDesignBrowser`: aprobado.
Los primeros intentos ajustaron las expectativas del test al formato numérico
de la API y al resumen global de órdenes activas; no se cambió el backend para
hacer coincidir una imagen.

## Capturas y revisión

Galería local: `backend/.cache/diseno-login-dashboard/index.html`.
Incluye referencias separadas, capturas reales de escritorio/tablet/móvil,
Análisis, rol mecánico y ejemplo de formulario claro conservado.
Los resultados y logs se conservan en la misma carpeta, excluida de Git.

Los servidores HTTP de las pruebas son efímeros y se cierran al finalizar.
El contenedor de desarrollo de localhost:8080 no se reconstruyó ni desplegó:
las capturas permiten revisar primero el diseño solicitado, usando el frontend
actual contra APIs reales del entorno aislado.
