# Cotizaciones PDF y organización de Inicio / Reportes

Actualización visual posterior: [nuevo diseño de cotización PDF](DISENO_COTIZACION_PDF.md)
con logo institucional y tabla de tres columnas. Este informe conserva los
resultados de la implementación inicial; la actualización detalla el formato
vigente y sus pruebas adicionales.

Implementado y comprobado el 27/10/2026 en la rama
`backend/arquitectura-go-mongodb`. Sin push, merge ni migraciones.

## Descargar una cotización

Abrir http://localhost:8080, iniciar sesión y entrar a **Cotizaciones**. En
**Cotizaciones guardadas**, cada fila ofrece **Abrir** y el botón de descarga
**PDF**. Abrir conserva la consulta y edición existentes. PDF utiliza la
cotización guardada, no los cambios sin guardar del formulario.

`GET /api/v1/cotizaciones/{id}/pdf` exige sesión JWT y `cotizaciones.ver`, incluso
al invocar directamente la URL. No exige permisos de edición ni un pago:
una cotización pendiente se puede descargar para solicitar aprobación.
Respuestas: PDF 200, sesión ausente 401, permiso ausente 403, ID inválido 400,
cotización inexistente 404. El frontend informa un fallo y no genera una descarga
exitosa cuando la API rechaza la solicitud.

Documento A4 borgoña, independiente de la ficha técnica. Contiene datos del taller
desde la vista existente, cliente/documento/contacto, vehículo/placa/año, fecha,
número, recepción, versión, conceptos y tipos, cantidades, precios, descuentos
por unidad, subtotales, total y estado. Los importes proceden de PostgreSQL:
no se recalculan ni se alteran precios o descuentos al exportar.

Si hay aprobación registrada se identifica su versión, persona, medio y fecha,
conservando la versión actual del documento por separado. Al invalidarse la
aprobación por una revisión, la descarga indica Pendiente y no reutiliza la
aprobación anterior. Una cotización histórica aprobada sin metadatos lo indica
sin inventarlos. La descarga actual no ofrece seleccionar revisiones históricas;
su historial permanece intacto en PostgreSQL.

El servidor lee una instantánea mediante un único SELECT. No almacena ni actualiza
un PDF en la cotización, y no escribe en aprobación, detalles, historial ni
entregas. Descripciones extensas continúan en otras páginas; se repiten cabecera,
tabla, identificación y número de página. No se inventan firmas.

Se reutilizó el ensamblador PDF del proyecto, separando la plantilla de la ficha
técnica del documento de cotización. Se conserva la fuente WinAnsi existente:
soporta español y caracteres occidentales; un carácter fuera de esa fuente, como
un emoji, provoca un error explícito 422 en lugar de omitir o sustituir texto.

## Diferencias entre las pantallas

| Inicio — Dashboard | Reportes |
|---|---|
| Indicadores resumidos del período | Servicios efectivamente realizados y sus cantidades |
| Órdenes activas y distribución por estado | Tabla detallada de ingresos netos de pagos |
| Recepciones, diagnósticos y entregas registrados | Gráfico lineal Evolución de ingresos |
| Carga activa por mecánico | Semana/Mes: días; Año: meses |
| Bandejas de diagnósticos pendientes y listos para cotizar, según permisos | Fecha de referencia para consultar períodos históricos |
| Enlace a Reportes con permiso | Comparación con el período de calendario anterior: ingresos y órdenes finalizadas |

Se retiró de Inicio el gráfico de ingresos que duplicaba Reportes, y el segundo
gráfico de estados que repetía el circular. La información de servicios más
frecuentes se presenta en Reportes junto a su tabla completa. Permanecen el menú,
logo, colores, componentes de tablas, gráficos y hojas de estilo originales.

Reportes reutiliza el contrato `GET /api/v1/reportes?period=...&desde=...&hasta=...`.
Los intervalos son de calendario en America/Lima, con final exclusivo y semanas
desde el lunes. Consulta por separado el período elegido y el anterior; no aplica
porcentajes sin base. Ambos son períodos completos, incluso cuando el actual aún
está en curso. Tabla, gráfico y total consumen las mismas series de pagos netos,
incluidas sus rectificaciones, sin unir pagos a conceptos de cotización ni contar
cotizaciones pendientes como ingresos. Modo demostración permanece separado.

## Pruebas realizadas

- `node checks/run.mjs`: cinco suites aprobadas, incluida `report-periods.mjs`
  (históricos, cambio de año, febrero bisiesto y conciliación de ambas series).
- `go test ./... -count=1 -v` en Compose aislado: aprobado, con PostgreSQL 18.6 y
  MongoDB 9.0; incluye autenticación, permisos, transacciones, restricciones y
  fallos de MongoDB. Las pruebas de navegador optativas se omiten dentro del
  contenedor; la auditoría Chrome se ejecutó separadamente en Windows.
- `go test ./integration -run 'TestAudit(CorrectionsPostgreSQL|BrowserIsolated)$'
  -v -count=1`: aprobado contra las dos bases aisladas. Go sirve frontend y APIs
  desde un origen efímero. Chrome descarga PDF pendientes y aprobados en atención
  directa y con diagnóstico, abre la fila correcta y conserva sus datos al
  descargar. Comprueba el mensaje ante una descarga rechazada, Semana/Mes/Año,
  febrero de 2025, comparación, Dashboard y el flujo hasta pago/entrega.
- Integración HTTP: 200 para usuario con solo `cotizaciones.ver`, 401 sin sesión,
  403 sin permiso, 400 ID inválido, 404 inexistente. Huellas de cotización,
  aprobación, historial y detalles idénticas antes/después de dos descargas.
- PDF unitario: Pendiente, Aprobada, Rechazada; aprobación versión 8/documento
  versión 12; plantilla de ficha técnica independiente; 65 conceptos con uno
  extenso continuado, último concepto y total conservados. Generó 18 páginas A4.
- Poppler y revisión visual: PDF pendientes/aprobados de datos persistidos de
  prueba y primera/última página del PDF extenso, sin recortes ni solapamientos.
  También se revisaron capturas de Inicio y Reportes.
- `go vet ./...`: aprobado. `git diff --check` y comprobación sintáctica JS:
  aprobados.

Los registros de prueba se crearon exclusivamente en `elbravodeljac_test` y
`elbravodeljac_tecnico_test`. No se utilizó ElBravoDelJac ni el puerto 5432 del
equipo. El entorno de desarrollo se actualizó recreando solo su contenedor Go:
`/readyz` devolvió 200. Las huellas de las 26 tablas de desarrollo y los conteos
de las colecciones MongoDB/GridFS permanecieron idénticos antes y después.
No se reiniciaron bases ni contenedores ajenos.

Evidencias locales ignoradas por Git: `backend/.cache/audit-tests.log`,
`audit-native.log`, `audit-browser/resultado.json`, PDF/capturas en `audit-browser/`
y `quotation-pdf/`. Las capturas/PDF de prueba no se incorporan al repositorio.

## Archivos de esta mejora

Nuevos:

- `backend/internal/httpapi/quote_pdf.go`
- `backend/internal/platform/postgres/quote_pdf.go`
- `backend/internal/platform/technicalsheet/quote.go`
- `backend/internal/platform/technicalsheet/quote_test.go`
- `dist/assets/js/components/quote-pdf.js`
- `checks/report-periods.mjs`
- `docs/backend/COTIZACIONES_PDF_REPORTES.md`

Modificados:

- `backend/internal/httpapi/workshop.go`
- `backend/internal/httpapi/catalog_test.go`
- `backend/internal/platform/technicalsheet/pdf.go`
- `backend/integration/audit_corrections_test.go`
- `dist/assets/js/views/operations.js`
- `dist/assets/js/views/administration.js`
- `dist/assets/js/app.js`
- `dist/assets/js/services/dashboard-service.js`
- `dist/assets/js/services/real-workshop.js`
- `checks/audit-browser.mjs`
- `checks/run.mjs`
- `checks/workflow.mjs`
- `docs/backend/ESTADO_ENDPOINTS.md`
- `README.md`

No se cambiaron tablas, claves foráneas, datos históricos, CSS, dependencias ni
credenciales. Los archivos privados de configuración permanecen ignorados.

## Ejecutar y repetir en Windows

La aplicación actualizada está disponible en http://localhost:8080 con las
cuentas existentes de desarrollo. Si los contenedores están detenidos:

```powershell
docker compose --env-file .env.fase-e.local -p elbravodeljac-fase-e-dev -f compose.dev.yaml up -d postgres-dev mongo-dev api
```

Para reconstruir cambios futuros: añadir `--build` al comando anterior. No se
necesitan migraciones para esta mejora. El archivo privado debe ser el existente;
no se deben reemplazar contraseñas de las bases ya inicializadas.

Pruebas completas: `.\backend\scripts\test-audit.ps1`. Requiere Docker Desktop,
Go, Node, Chrome y `PLAYWRIGHT_MODULE` apuntando a Playwright. El script comprueba
la propiedad de sus proyectos Compose y usa solamente los puertos aislados
25432/27029 para el navegador. Desde `backend/`, `go test ./...` sin las variables
de integración omite expresamente las pruebas que necesitan bases.
