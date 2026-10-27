# Cotizaciones: creación, consulta y búsqueda remota

Corrección verificada el 27/10/2026 en la rama `backend/arquitectura-go-mongodb`.
No se hizo push ni merge. No hay migraciones ni cambios de tablas, FK o volúmenes.

## Causa y comportamiento corregido

La vista identificaba una cotización existente no solo por `?id=`, sino también
por la recepción/diagnóstico seleccionado. Al seleccionar una atención que tenía
cotización, cargaba los conceptos y precios anteriores. Si ya tenía OT, activaba
correctamente el bloqueo de negocio, pero parecía un formulario nuevo bloqueado.
El desplegable se construía con todas las recepciones cargadas en memoria.

Ahora solamente `cotizaciones?id=ID` abre un registro guardado. `cotizaciones`
(Nueva cotización) siempre abre sin atención seleccionada, sin líneas ni precios
heredados. Puede agregar/quitar líneas y escribir descripciones, cantidades,
precios y descuentos. Seleccionar una atención no navega ni reemplaza esos valores.
El diagnóstico se presenta en un panel informativo independiente.

Los enlaces con `?recepcion=ID` o `?diagnostico=ID` preparan ese origen en un
formulario nuevo. Si ya está cotizado, se informa y se ofrece Abrir cotización
existente. No se cargan sus líneas automáticamente y se impide guardar duplicados.
La comprobación también ocurre en Go dentro de la transacción: se bloquea la
recepción con FOR UPDATE antes de consultar si existe cotización e insertar.
No se borran duplicados históricos ni se crean restricciones físicas nuevas.

## Búsqueda

Se reutiliza `components/remote-picker.js`: debounce de 250 ms, cancelación de
peticiones anteriores, descarte de respuestas obsoletas, selección por ID y páginas
de 20 resultados. Escribir invalida el ID anterior; no hay selección automática.
El campo mantiene las validaciones para exigir seleccionar un resultado válido.

Nuevo endpoint autenticado con `cotizaciones.ver`:

`GET /api/v1/cotizaciones/atenciones?q=F&page=1&pageSize=20`

Filtros: nombre/apellidos, documento, placa o número de recepción; `jac_normalize`
existente ignora mayúsculas y tildes. `recepcionId` permite recuperar el origen
exacto. Count y resultados usan el mismo snapshot de PostgreSQL y parámetros SQL.
Devuelve recepción, cliente/documento, vehículo, diagnóstico finalizado cuando
corresponde e ID de cotización existente. No ofrece evaluaciones incompletas.
Una recepción directa es elegible sin diagnóstico. Las ya cotizadas se identifican
para permitir abrirlas, no volver a registrarlas.

El formulario no descarga todas las recepciones/diagnósticos para construir una
lista. Mantiene el listado de cotizaciones guardadas y recupera el contexto de
la atención seleccionada. La API de cotizaciones incluye número/código de OT,
permitiendo explicar el bloqueo aun sin permiso para listar Órdenes.

## Edición, permisos y trazabilidad

- Crear exige `cotizaciones.crear`; editar un registro exige `cotizaciones.editar`.
  La pantalla distingue ambos permisos y explica el modo de solo lectura.
- Una cotización vinculada a una OT permanece bloqueada y muestra exactamente:
  “Cotización de solo lectura: vinculada a la orden de trabajo N.º X”.
  El backend conserva el rechazo 409 a modificaciones, aunque se invoque la API.
- Sin OT, se conserva la edición explícita permitida por las reglas existentes,
  incluidos conceptos libres y referencias antiguas que se conservan al editar.
- No se cambia ningún cálculo, regla de precio, descuento, aprobación o stock.
  No se rellenan líneas desde recomendaciones ni desde catálogos.

Una rectificación de una cotización con OT no está habilitada. La alternativa a
aprobar es una cotización complementaria con autorización, vínculo explícito al
presupuesto original, motivo/actor/fecha y nueva aprobación del cliente, manteniendo
intactos los importes originales y separando su impacto en órdenes y pagos. Eso
requiere definir un contrato y trazabilidad específicos; no se usa una segunda
cotización ordinaria del mismo origen para eludir los bloqueos actuales.

## Verificación real

- `node checks/run.mjs`: cuatro suites JavaScript aprobadas.
- `checks/quotes-browser.mjs`, Chrome contra Go/PostgreSQL de desarrollo aislado:
  nueva vacía y editable, F/nombre sin tildes/documento/placa/número de recepción,
  paginación, exclusión de diagnóstico incompleto, selección de diagnóstico final,
  conservación de texto/precio escrito antes de seleccionar, servicio y repuesto
  manuales, total 235.00, edición a 255.00, F5 y recuperación por API.
- Consulta vinculada a OT: mensaje con número correcto, controles bloqueados y
  rechazo HTTP 409; posterior consulta conserva el importe autorizado.
- Atención ya cotizada: aviso y enlace explícito, sin heredar conceptos; volver
  a Nueva cotización limpia el contexto y mantiene Agregar habilitado.
- `go test ./... -count=1 -v` en Compose aislado: aprobado con PostgreSQL/MongoDB.
  Se extendió TestAssignedDiagnosisWorkflow con búsquedas normalizadas, páginas,
  permisos y duplicados. La prueba de códigos de OT usa otra recepción en lugar
  de crear una segunda cotización indebida para el mismo diagnóstico.
- `go vet ./...` en el contenedor de pruebas: aprobado (salida 0).
- El E2E Go que requiere FRONTEND_E2E=1 sigue omitido dentro de esa imagen sin
  Node/Playwright. Chrome se ejecutó separadamente; no se cuenta ese skip como éxito.
- Comparación de hashes por ID de todas las cotizaciones, detalles y órdenes
  existentes antes de la prueba: ninguna fila cambió ni desapareció. Las nuevas
  filas de la prueba aislada se conservan, sin limpiar datos.

Informes locales excluidos de Git: `backend/.cache/quotes-browser/resultado.json`,
`backend/.cache/quotes-tests.log`; comprobación de conservación en
`backend/.cache/quotes-existing-before.json`. Ninguno publica credenciales.

## Uso en localhost

Abrir http://localhost:8080/ y presionar Ctrl+F5. Entrar con la cuenta autorizada
existente; Cotizaciones → Nueva cotización. Escribir una letra, documento o placa
en Atención por cotizar, seleccionar el resultado y agregar las líneas manuales.
Guardar. Abrir un registro guardado únicamente desde su enlace Abrir.
El servidor sigue en modo REAL con las bases aisladas; no se accedió a ElBravoDelJac
original ni se reiniciaron volúmenes. Para arrancarlo si está detenido:
`.\backend\scripts\start-dev.ps1` desde la raíz, con Docker Desktop iniciado.

## Archivos de esta corrección

- `README.md`
- `backend/integration/diagnosis_workflow_test.go`
- `backend/integration/workshop_test.go`
- `backend/internal/httpapi/catalog_test.go`
- `backend/internal/httpapi/workshop.go`
- `backend/internal/platform/postgres/workshop_reads.go`
- `backend/internal/platform/postgres/workshop_writes.go`
- `checks/quotes-browser.mjs`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/views/operations.js`
- `docs/backend/CORRECCION_COTIZACIONES.md`
- `docs/backend/ESTADO_ENDPOINTS.md`
