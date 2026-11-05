# Fase B: proceso principal del taller implementado

Rama `backend/arquitectura-go-mongodb`. Se añadieron 24 operaciones reales a las
17 de fase A: **41 de las 71 operaciones del catálogo están implementadas**.
No se modificaron el SQL aprobado, las 26 tablas, las 42 FK, dist/ ni proveedores.
No se conectó a ElBravoDelJac. La configuración sigue admitiendo únicamente las
bases separadas elbravodeljac_test/elbravodeljac_dev.

## Endpoints añadidos y comprobados por HTTP/PostgreSQL

Base `/api/v1`. Todas las rutas exigen JWT; escrituras exigen Origin y CSRF.

| Métodos | Ruta | Permiso |
|---|---|---|
| GET, POST | /recepciones | recepciones.ver / recepciones.crear |
| GET | /recepciones/{id} | recepciones.ver |
| GET, POST | /diagnosticos | diagnosticos.ver / diagnosticos.crear |
| GET | /diagnosticos/{id} | diagnosticos.ver |
| GET, POST | /cotizaciones | cotizaciones.ver / cotizaciones.crear |
| GET, PUT | /cotizaciones/{id} | cotizaciones.ver / cotizaciones.editar |
| GET, POST | /ordenes | ordenes.ver / ordenes.crear |
| GET, PATCH, DELETE | /ordenes/{id} | ordenes.ver / ordenes.editar / ordenes.eliminar |
| GET, POST | /ordenes/{id}/pagos | pagos.ver / pagos.registrar |
| GET, POST | /ordenes/{id}/entrega | pagos.ver / pagos.entregar |
| GET | /servicios, /repuestos | cotizaciones.ver |
| GET | /mecanicos, /asignaciones | mecanicos.ver |
| GET | /me/tareas | auxilio.ver, limitado al actor asignado |

Asignación, avances y pruebas se realizan dentro de **PATCH /ordenes/{id}** en
una transacción; no se obliga al cliente a coordinar tres escrituras HTTP. El GET
de la orden incluye asignaciones actuales/anteriores, avances, pruebas y estados.
GET /asignaciones permite consultar asignaciones actuales y filtrar ordenId/usuarioId.

Listas: page/pageSize (máximo 100) y meta de paginación. Recepciones admite clientId
y pendienteDiagnostico; diagnósticos recepcionId; cotizaciones diagnosticoId, estado,
sinOrden; órdenes estado y q por código, placa o cliente. Count y filas comparten
un snapshot de lectura. IDs y valores decimales se transportan como strings.

## Payloads principales

Los siguientes ejemplos son contratos, no registros devueltos por una base real.
Los ID de ejemplo deben sustituirse por los obtenidos de los endpoints.

```json
{
  "clientId": "1", "vehicleId": "1", "mileage": "25300.50",
  "symptoms": "Ruido en frenos"
}
```

Diagnóstico: receptionId y findings. Técnico y fechas se obtienen del servidor.

```json
{
  "diagnosisId": "1", "status": "Aprobada",
  "items": [
    {"serviceId": "1", "quantity": "1.00", "unitPrice": "120.00"},
    {"sparePartId": "1", "origin": "Taller", "quantity": "1.00", "unitPrice": "35.00"}
  ]
}
```

PUT de cotización reemplaza cabecera/detalles atómicamente, conservando diagnosisId.
Cada línea requiere exactamente serviceId o sparePartId existente y activo. No
acepta concepto libre sin referencia ni crea catálogos silenciosamente. Concept es
una descripción opcional; si se omite toma el nombre real del catálogo. El servidor
calcula subtotales y total con aritmética entera/big.Int y redondeo half-up. Repuesto
aportado por cliente requiere precio cero. Se validan límites NUMERIC(10,2)/(12,2)
y un máximo de 100 líneas. No se acepta total/subtotal calculado por el cliente.

Nueva OT: quoteId, mechanicId y assistantId opcional.

```json
{
  "state": "Finalizada", "note": "Pruebas verificadas",
  "test": "Prueba de ruta", "result": "Satisfactoria"
}
```

PATCH exige state y note. mechanicId/assistantId opcionales solo con ordenes.asignar;
assistantId vacío quita la asignación actual. Nunca se eliminan asignaciones antiguas:
se desactivan y se agregan las nuevas, con roles MECANICO/AUXILIAR. Personal activo
con rol Mecánico/Supervisor para responsable y Auxiliar para ayudante. No se modifican
los roles de ningún usuario. Cada actualización deja un avance y evento de estado.

Pago: amount, method (Transferencia/Efectivo/Tarjeta), comprobante opcional (50).
Entrega: kind (Titular / representante o Tercero), method (Confirmación del titular,
Documento o Representante facultado), verified=true y contactId solo para Tercero.
El receptor se copia a snapshots SQL; no se confía en nombre/documento enviados.

## Reglas y concurrencia

- Recepción bloquea vehículo, valida propietario/km y guarda recepción + km en
  una transacción. El cliente histórico se conserva tras futuras transferencias.
- Diagnóstico respeta la FK de recepción y el índice único por recepción.
- Guardar cotización y crear OT comparten bloqueo sobre cotizaciones. No se puede
  cambiar una cotización que ya tiene OT; tampoco crear dos OT para la misma.
- La OT solo nace aprobada; el código se deriva del ID de su secuencia existente,
  sin max+1, nuevas columnas ni secuencias. Se guarda junto con personal/historial.
- PATCH bloquea la OT. Rechaza edición de OT finalizada, cancelada o entregada.
  Un operador sin ordenes.asignar solo actualiza OT donde está asignado activamente.
  Consultas de órdenes también se restringen a asignación cuando el usuario no
  tiene permisos de gestión/creación/eliminación o consulta de pagos.
- La última prueba debe ser satisfactoria y no anterior a la última reparación.
  Reparar de nuevo invalida resultados anteriores, incluso sin cambiar de estado.
  Los eventos usan tiempo monotónico por OT bajo bloqueo (al menos un microsegundo
  tras el evento anterior). Una prueba de la misma actualización cuenta después
  de la reparación registrada en ese comando atómico. Fallar la validación revierte
  también el avance, la prueba, el estado y las reasignaciones de ese comando.
- Borrar OT está limitado a órdenes sin avances, pruebas, pagos ni entrega. Se
  limpia solo su historial inicial/asignaciones dentro de la transacción; las OT
  con ejecución conservan trazabilidad y devuelven 409. No se usa CASCADE.
- Pago requiere OT finalizada. Bloquea la misma fila de OT que actualización,
  borrado y entrega, recalcula saldo y rechaza sobrepago, cero y decimales inválidos.
- Entrega requiere OT finalizada, saldo exactamente cero, verificación expresa y
  contacto activo del cliente de la recepción. Dos entregas concurrentes producen
  una entrega y un conflicto. No se borra ni actualiza el historial transaccional.

Los triggers existentes se mantienen; Go complementa las reglas que no garantiza
solo el dump. No se crean tablas para idempotencia: no deben habilitarse reintentos
automáticos de pagos hasta acordar persistencia de claves de idempotencia. Dos pagos
parciales intencionales siguen siendo operaciones distintas, sujetos al saldo.

## Diferencia pendiente para integrar el frontend

El formulario visual actual admite conceptos de texto, pero SQL exige referencias.
La solución aplicada en la API es consultar /servicios y /repuestos y exigir sus IDs.
Al conectar el frontend se deberá asociar cada concepto al catálogo existente;
no se resolverá por coincidencia ambigua de nombre ni se quitará el CHECK SQL.
No hay incompatibilidad que obligue a cambiar el esquema para implementar estas APIs.
La demo y el diseño siguen intactos; aún no se activó la conexión del navegador.

## Qué sigue pendiente (30 operaciones del catálogo)

- Dashboard y reportes.
- Alta/consulta de auxilios y unidades de auxilio (Mis tareas GET sí está implementado).
- Administración de usuarios, roles y permisos. La autenticación/lectura de permisos
  ya funciona; no se han alterado las asignaciones N:M.
- Proveedores y su relación con repuestos; propuesta de especialidad aún sin aplicar.
- Altas de servicios/repuestos: esta fase consulta y utiliza catálogos existentes.
- Cinco operaciones técnicas de escáner/evidencias: requieren MongoDB.

## Próxima fase MongoDB

Falta incorporar driver/configuración, repositorios, validadores BSON e índices,
GridFS (o almacenamiento acordado), límites/validación de archivos, autorización
contra diagnóstico/OT SQL, claves de reintento, compensación de archivos huérfanos
y pruebas de fallos entre motores. Se conserva database/mongodb como propuesta.
Diagnóstico/resumen, pagos y demás transacciones permanecen en PostgreSQL; MongoDB
solo contendrá resultados de escáner y evidencias. No se promete transacción común.

## Pruebas ejecutadas

Se ejecutó scripts/test-local-postgres.ps1: clúster nuevo, puerto aleatorio y base
elbravodeljac_test con el esquema aprobado, detenido al finalizar. Las pruebas usan
registros sintéticos generados exclusivamente para pruebas, pero el servidor HTTP,
handlers, transacciones y motor PostgreSQL son reales: no hay respuestas simuladas.

Pasaron la suite de fase A y los siete grupos de TestWorkshopHTTPPostgreSQL:
recepciones/diagnósticos; catálogos/cotizaciones/rollback; creación/asignaciones/códigos;
permisos por objeto/avances/pruebas; pagos concurrentes; entrega/snapshots/concurrencia;
JWT/permisos/CSRF y comprobación final de 26 tablas/42 FK. También pasaron las pruebas
unitarias de precisión, redondeo y referencias, go vet ./..., y las tres verificaciones
del frontend (verify.mjs, workflow.mjs y edge-cases.mjs). git diff --check no encontró
errores de espacio. La comparación de todos los archivos de dist/ y del SQL contra
HEAD confirmó que están intactos (normalizando únicamente los finales CRLF/LF).

El script omite el fsync inicial de initdb exclusivamente para el clúster descartable;
no desactiva las transacciones de PostgreSQL ni es una prueba de durabilidad ante crash.
