# Correcciones de auditoría funcional: etapas 1, 2 y 3

Fecha: 27/10/2026, America/Lima. Rama: `backend/arquitectura-go-mongodb`.
No se hizo commit, push ni merge. Se conservaron el frontend, el diseño borgoña,
la arquitectura Go/PostgreSQL/MongoDB y las 26 tablas existentes.

## 1. Problemas corregidos

### Etapa 1: presupuesto aprobado y ejecución real

La cotización ahora tiene versión, huella de contenido, aprobación e historial.
La aprobación exige nombre de quien autoriza por el cliente y medio: Presencial,
WhatsApp, Llamada u Otro. El servidor registra usuario, fecha/hora, importe y copia
exacta de los conceptos, cantidades, precios y descuentos. No se inventan firmas.

Si cambia el contenido de una cotización aprobada, el servidor la guarda como
Pendiente y elimina su aprobación vigente, conservándola en el historial con los
conceptos originales. No permite cambiar precios y aprobarlos silenciosamente en
la misma actualización. Se guarda primero el cambio y se registra después la nueva
aceptación. Una versión antigua se rechaza para evitar sobrescribir otra edición.
Las cotizaciones con OT siguen inmutables. Los conceptos siguen siendo manuales;
no se reintroduce el selector de sugerencias ni se exige catálogo.

Una ampliación se crea desde Nueva ampliación autorizada en la cotización original
con OT abierta. Utiliza el mismo POST de cotizaciones con `parentQuoteId`, misma
recepción y aprobación propia. Conserva el presupuesto original y genera su propia
OT para el alcance adicional; los importes de ambas no se mezclan ni se reescribe
la orden original. Deben finalizarse y liquidarse todas las órdenes/ampliaciones
aprobadas de la atención antes de entregar el vehículo. Una ampliación aprobada
sin OT también impide dar por concluida la entrega. No se permiten ampliaciones de
una atención ya cerrada.

En el Control técnico de Órdenes se reutiliza el modal existente: cada concepto
aprobado tiene Pendiente, Realizado o No realizado, con observaciones. No realizado
exige motivo. Los conceptos ajenos se rechazan y se conserva el historial de las
versiones del control. Se mantiene el control de concurrencia por versión.

Una orden nueva o todavía abierta no puede finalizar con conceptos pendientes.
Si hay conceptos no realizados, un usuario autorizado debe aprobar el ajuste desde
el mismo Control técnico. El servidor registra actor, fecha, motivo y la versión
de ejecución conciliada. Cambiar esa ejecución invalida la conciliación anterior.
El importe a cobrar excluye los conceptos no realizados una vez autorizado el
ajuste; el presupuesto aprobado original conserva su importe. Si los anticipos
superan el importe ajustado, se exige regularizar el exceso antes de finalizar.
Se conservan las pruebas técnicas vigentes posteriores a la última reparación.
No se exige completar las inspecciones opcionales para un servicio sencillo.

### Etapa 2: anticipos, correcciones y receptor físico

Se permiten pagos desde una OT abierta y aprobada, incluyendo anticipos, parciales
y saldo final. Se rechazan órdenes canceladas, pagos superiores al saldo y pagos
ordinarios nuevos después de entregar. Orden, saldo e inserción comparten bloqueo
y transacción para impedir sobrepagos concurrentes.

Cada pago nuevo exige `requestId`, único por orden. Un reintento idéntico devuelve
el mismo pago; reutilizar la clave con otro monto/medio/referencia se rechaza.
La interfaz guarda la clave pendiente en sessionStorage durante errores de red,
normalizando el importe, y la elimina solo al recibir confirmación. No usa una
comparación de montos para eliminar pagos legítimos iguales. Un pago revertido no
puede volver a registrarse con su antigua clave.

Nuevo permiso `pagos.rectificar`, concedido inicialmente solo a Propietario y
Administrador. La matriz permite administrarlo después; repetir la migración no
restablece permisos retirados. Revertir un pago requiere motivo y clave de reintento.
Se conservan monto, fecha y responsable originales; se agrega la reversión con su
actor, fecha y motivo. No se borran pagos. El historial muestra los revertidos y su
relación con cualquier sustitución.

Antes de entregar se puede revertir un pago incorrecto y registrar el correcto.
Después de entregar no se permite abrir saldo con una reversión aislada: la opción
Rectificar sustituye el pago por otro del mismo importe en una transacción, por
ejemplo para corregir medio/referencia, conservando ambos asientos y saldo cero.
Una devolución o modificación del importe de una atención ya entregada requiere
un procedimiento comercial de reapertura/ajuste posterior y no se permite mediante
una modificación oculta del presupuesto ni borrando el historial.

Entrega diferencia Propietario o titular, Persona autorizada y Representante de
empresa. Las dos últimas requieren un contacto activo vinculado al cliente, cuyos
datos reales se congelan como receptor. Una empresa con RUC no puede figurar como
la persona física que recoge: debe identificarse su representante.

Se registran por separado identidad/autorización y conformidad con entrega/trabajos,
con observaciones. La confirmación no constituye una firma digital. Se exige OT
finalizada, saldo neto cero y que no queden ampliaciones/órdenes pendientes.

### Etapa 3: documento, historial e indicadores

La ficha utiliza solo conceptos confirmados como realizados y el importe definitivo
conciliado. Conserva el control, observaciones, próximo mantenimiento y la explicación
del ajuste económico; no presenta como utilizado el repuesto no realizado. Las dos
páginas originales mantienen su fondo, estilo y organización. Si el texto o la
cantidad de conceptos no caben, se imprimen referencias y anexos A4 con el contenido
completo. Se inspeccionaron visualmente páginas originales y anexos generados.

La instantánea y el PDF quedan congelados al entregar. La descarga posterior desde
Pagos y entrega o historial reutiliza el mismo documento; cambiar luego datos del
cliente o rectificar el medio de pago no modifica el PDF emitido.

Servicios realizados suma cantidades de conceptos efectivamente confirmados en
órdenes finalizadas, incluyendo confirmaciones explícitas del control anterior.
No se infiere que todos los conceptos históricos se realizaron por estar cotizados.
Por eso los indicadores pueden bajar para registros antiguos sin confirmación:
sus presupuestos, pagos y atenciones no se borran ni se modifican.

Ingresos procede de una vista de movimientos: pago positivo en su fecha original
y reversión negativa en la fecha de rectificación. No borra ingresos históricos ni
multiplica montos mediante joins con conceptos. Una sustitución del mismo importe
produce reversión y nuevo pago, con efecto neto cero en esa fecha. Tabla y gráfico
usan la misma agregación; se admiten períodos netos negativos. Semana/Mes muestran
días y Año meses, con límites del negocio en America/Lima.

Las órdenes activas y la carga de mecánicos incluyen trabajo pendiente de meses
anteriores. Finalizadas/canceladas se filtran por su fecha de cierre.
Se agrega historial de atenciones por cliente y se amplía el del vehículo con
trabajos, órdenes, pagos/rectificaciones y descarga de fichas. Los datos financieros
se exponen solo con permiso pagos.ver. Los selectores de cotización aprobada y OT
para pagos reutilizan el buscador remoto, con búsqueda normalizada y páginas de 20.

## 2. Implementación reutilizada

- Servidor Go, driver PostgreSQL, transacciones y bloqueos de orden/cotización.
- JWT, cookies HttpOnly, CSRF, permisos efectivos y relaciones N de roles.
- Recepciones directas y técnicas; asignación, borradores y diagnósticos finalizados.
- Cotizaciones manuales, descuentos, precios y referencias opcionales existentes.
- Órdenes, avances, pruebas, control técnico y su modal.
- Búsqueda remota con debounce, cancelación y selección por ID.
- Identificación de contactos y entrega, plantilla PDF e instantáneas.
- MongoDB/GridFS existentes; no se duplican datos relacionales ni se añade inventario.

## 3. PostgreSQL y compatibilidad

Migraciones compatibles nuevas:

| Migración | Ampliaciones |
|---|---|
| 005_auditoria_flujo.sql | Versión/huella/aprobación/historial de cotizaciones; referencia a cotización base; versión de flujo e historial de control en OT |
| 006_pagos_entrega.sql | Clave de reintento, reversión y rectificación en pagos; receptor real y conformidad en entregas; vista de movimientos y función de importe definitivo; adaptación del trigger de entrega al saldo neto |

Se mantienen **26 tablas y todas las FK anteriores**. Se añade una FK desde
cotizaciones hacia su presupuesto base: **47 FK en total** (42 originales más las
ampliaciones de fases previas y esta referencia). No se crean tablas duplicadas.
No se ejecutaron DROP TABLE, TRUNCATE ni recreaciones de las bases existentes.

Los registros antiguos mantienen sus importes y datos. Las aprobaciones antiguas
no reciben un nombre o medio inventado: si no tienen OT, deberán recapturarse antes
de crear una nueva. Las OT abiertas adoptan la conciliación del nuevo flujo; las
finalizadas/canceladas conservan su condición histórica. No se marcan trabajos
históricos como realizados automáticamente ni se inventan inspecciones.

Las migraciones se probaron primero en elbravodeljac_test y luego se aplicaron al
entorno local elbravodeljac_dev para servir la aplicación actualizada. No se accedió
a la base original ElBravoDelJac del puerto 5432 ni se eliminaron volúmenes.

Se guardaron huellas de las 26 tablas antes de actualizar, excluyendo solamente las
columnas nuevas y el nuevo permiso autorizado para poder comparar los datos originales.
La comparación confirmó intactos los registros preexistentes. Durante la ventana
aparecieron altas concurrentes: recepción 30, cotización 24, OT 21, pago 14, entrega
13 y sus registros asociados. Se conservaron todas. Al separar esas nuevas filas,
los registros originales de las 26 tablas coincidieron exactamente con sus huellas
previas. No se atribuyen esas altas a las pruebas: los escenarios automatizados
utilizaron exclusivamente las bases test.
Los conteos de todas las colecciones de MongoDB de desarrollo permanecieron iguales.

## 4. Contratos y permisos

Se extienden los endpoints existentes de cotización, OT, control, pagos y entrega.
Solo se incorporan dos operaciones distintas que no existían:

- PATCH `/api/v1/ordenes/{id}/pagos/{paymentId}`: permiso pagos.rectificar;
  `requestId`, `reason` y sustitución opcional `replacement: {method, comprobante}`.
- GET `/api/v1/clientes/{id}/historial`: permiso clientes.ver, finanzas sujetas a pagos.ver.

POST/PUT cotizaciones acepta `version`, `approval: {name, method}` y
`parentQuoteId`. El servidor agrega los metadatos y conserva el contenido aprobado.
PUT control-tecnico acepta `execution: {idDetalle: {status, notes}}` y, para usuarios
autorizados, `adjustmentReason`. No confía en una conciliación fabricada por el cliente.
El ajuste requiere acceso al control técnico y cotizaciones.editar; la matriz de
roles sigue determinando qué cuenta administrativa/supervisora puede realizarlo.

POST pagos exige requestId; POST entrega incluye recipientKind, conformity y
conformityNotes además de los controles previos. Las APIs devuelven errores de
validación/concurrencia/permisos sin aparentar éxito. Los límites de lectura del
presupuesto y control admiten los 100 conceptos declarados, con JSON acotado a 1 MiB.

El catálogo acumulado tiene **86 rutas de aplicación**, incluyendo frontend/opciones;
las 71 originales se conservan, sin rutas nuevas 501.

## 5. Pruebas ejecutadas y resultados

- `node checks/run.mjs`: cuatro suites aprobadas (estructura, flujo demo, límites y
  servicios reales simulados). Estas suites no se presentan como conexión a bases.
- Compose temporal aislado: `go test ./... -count=1 -v`, aprobado. Incluye pruebas
  previas de JWT/CSRF, permisos, FKs, transacciones, concurrencia, MongoDB/GridFS y PDF.
- TestAuditCorrectionsPostgreSQL: directo y diagnóstico; aprobación requerida y
  revocada al editar, versiones antiguas, presupuesto histórico protegido, ampliación,
  trabajos no realizados/ajuste, anticipo, reintentos concurrentes con una sola fila,
  sobrepago rechazado, reversión y sustitución autorizadas, receptor físico/conformidad,
  saldo cero, nueva descarga inmutable, anexos y reportes netos.
- Se comprueban cantidades ejecutadas y órdenes activas creadas dos meses antes.
- TestAuditBrowserIsolated: Chrome real con administradora/mecánico y Juan Pérez con
  Toyota Corolla, ambos recorridos completos, formularios, anticipo, reintento,
  reversión, aprobación renovada, ajuste, dos confirmaciones de entrega, rectificación
  tras entrega, PDF, historial, gráfico Semana/Mes/Año y persistencia tras F5.
- `go vet ./...`: aprobado. Pruebas unitarias HTTP/dominio/PDF adicionales aprobadas.
- PDF: se generaron y renderizaron ejemplos normales de dos páginas y un documento
  de cinco páginas para 24 conceptos extensos con observaciones; revisión visual de originales y anexos.

Dentro de la imagen Go se omiten los E2E que necesitan Node/Playwright. El nuevo
E2E AuditBrowser sí se ejecutó nativamente desde Windows contra las bases test,
con Chrome, además de la suite de contenedor. El E2E antiguo PhaseE no se ejecutó
como navegador en esta fase y no se cuenta como pasado.

Al navegar se observaron consultas canceladas al abandonar pantallas. El tracer
confirmó cancelaciones de contexto, **cero errores SQL inesperados**, y el navegador
exigió cero respuestas HTTP 5xx observadas y cero errores JavaScript. Los mensajes
genéricos del servidor por esas cancelaciones no se ocultan en el registro de pruebas.

Informes locales excluidos de Git: backend/.cache/audit-tests.log,
audit-native.log, audit-final.log, audit-browser/resultado.json, audit-pdf/ y
las huellas audit-preservation-before/after.json. No contienen contraseñas reales
publicadas en Git. Los accesos locales permanecen en archivos privados ignorados.

## 6. Qué queda fuera y límites conocidos

- Inventario avanzado: aplazado expresamente. Se preservan las referencias opcionales
  a servicios/repuestos para incorporarlo después; no se simula consumo de stock.
- Devoluciones o cambios del importe de una atención entregada requieren un proceso
  específico de ajuste comercial posterior. Se protege el saldo cerrado; sí se puede
  corregir medio/referencia mediante sustitución del mismo importe.
- La fuente original PDF admite Windows-1252 (incluye español). Caracteres fuera de
  esa tipografía requieren incorporar una fuente Unicode; se devuelve error explícito
  y se conserva la instantánea, en lugar de sustituir o eliminar caracteres.
- La conformidad registrada no es una firma electrónica ni una verificación externa
  automática de un mensaje de WhatsApp; la registra una persona autorizada.
- Scripts históricos de navegador que apuntaban directamente a localhost:8080 no se
  usaron para crear fixtures en esta fase. Para la auditoría usar test-audit.ps1.

## 7. Cómo verificar desde Windows

Aplicación actualizada: **http://localhost:8080/**. Ctrl+F5 para cargar el JavaScript
nuevo. Si está detenida, ejecutar `.\backend\scripts\start-dev.ps1` desde la raíz,
con Docker Desktop iniciado. El acceso existente se conserva en
`backend/.cache/acceso-desarrollo.txt`; no se publican contraseñas en este informe.

Prueba automática, sin altas de ejemplo en la base de trabajo:

```powershell
.\backend\scripts\test-audit.ps1
```

Usa los proyectos propios elbravodeljac-fase-e-tests y elbravodeljac-audit-test,
las bases elbravodeljac_test / elbravodeljac_tecnico_test, puertos loopback 25432 y
27029 para el navegador y un servidor Go temporal en puerto libre. Conserva las
filas del entorno de navegador y no borra volúmenes ni detiene proyectos ajenos.
La suite de contenedor usa otro PostgreSQL temporal en tmpfs, descartable al detenerse. Requiere Go, Node,
Chrome y Playwright; si falta una dependencia, falla indicando el motivo. La clave
privada de pruebas se conserva en `.env.audit-tests.local`, excluido de Git.

Recorrido manual:

1. Nueva cotización: seleccionar atención y escribir conceptos. Para aprobar,
   registrar nombre y medio. Si se modifica, guardar como pendiente y recapturar aceptación.
2. Crear OT y asignar mecánico. Puede registrar anticipos desde Pagos y entrega.
3. En Control técnico, marcar cada trabajo; para No realizado indicar motivo.
   Una cuenta autorizada confirma el ajuste económico. Dejar inspecciones opcionales sin evaluar.
4. Registrar prueba satisfactoria vigente, finalizar, cobrar saldo definitivo.
5. Identificar receptor, verificar autorización y registrar conformidad; entregar.
6. Descargar ficha; volver a consultarla desde historial del cliente/vehículo.
7. Consultar reportes: cantidades realmente realizadas e ingresos netos concilian
   con los pagos y reversiones del período.

## Archivos modificados y creados

Total: **46 archivos**. Logs, PDF de prueba y claves locales excluidos de Git.

- `README.md`
- `backend/cmd/dev-bootstrap/main.go`
- `backend/integration/attention_sheet_test.go`
- `backend/integration/audit_browser_test.go`
- `backend/integration/audit_corrections_test.go`
- `backend/integration/brands_test.go`
- `backend/integration/phase_a_test.go`
- `backend/integration/phase_c_test.go`
- `backend/integration/review_test.go`
- `backend/integration/workshop_test.go`
- `backend/internal/domain/audit.go`
- `backend/internal/domain/permission-policy.json`
- `backend/internal/domain/workshop.go`
- `backend/internal/httpapi/catalog_test.go`
- `backend/internal/httpapi/payment_correction.go`
- `backend/internal/httpapi/resources.go`
- `backend/internal/httpapi/technical_sheet.go`
- `backend/internal/httpapi/workshop.go`
- `backend/internal/platform/postgres/audit_workflow.go`
- `backend/internal/platform/postgres/frontend.go`
- `backend/internal/platform/postgres/migrate.go`
- `backend/internal/platform/postgres/migrations/005_auditoria_flujo.sql`
- `backend/internal/platform/postgres/migrations/006_pagos_entrega.sql`
- `backend/internal/platform/postgres/statistics.go`
- `backend/internal/platform/postgres/technical_sheet.go`
- `backend/internal/platform/postgres/workshop_reads.go`
- `backend/internal/platform/postgres/workshop_writes.go`
- `backend/internal/platform/technicalsheet/control.go`
- `backend/internal/platform/technicalsheet/pdf.go`
- `backend/internal/platform/technicalsheet/pdf_test.go`
- `backend/scripts/test-audit.ps1`
- `backend/scripts/test-local-postgres.ps1`
- `checks/audit-browser.mjs`
- `compose.audit-browser.yaml`
- `compose.test.yaml`
- `dist/assets/js/app.js`
- `dist/assets/js/components/charts.js`
- `dist/assets/js/components/technical-sheet.js`
- `dist/assets/js/data/permissions.js`
- `dist/assets/js/services/real-workshop.js`
- `dist/assets/js/services/workshop-service.js`
- `dist/assets/js/views/clients-vehicles.js`
- `dist/assets/js/views/operations.js`
- `dist/assets/js/views/payments.js`
- `docs/backend/CORRECCIONES_AUDITORIA_FUNCIONAL.md`
- `docs/backend/ESTADO_ENDPOINTS.md`
