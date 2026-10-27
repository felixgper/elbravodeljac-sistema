# Endpoints REST propuestos

Base `/api/v1`. 71 operaciones reservadas; todas responden **501 NOT_IMPLEMENTED** en el esqueleto. No implican implementación ni acceso a bases. Contratos de campos y bloqueos: CORRESPONDENCIA.md y ARQUITECTURA.md.

## Convenciones

- JSON UTF-8. IDs SQL como strings decimales; campos JSON camelCase del formulario. Rutas {id} usan ID SQL, nunca codigo_ot.
- Dinero, km y cantidades: strings decimales exactos; respuesta con dos decimales. total/subtotal/saldo son calculados por servidor.
- Éxito futuro: 200 lectura/edición; 201 alta con Location; 204 eliminación/cierre de sesión. Detalle {data: objeto}; listas {data: [], meta: {page, pageSize, total}}.
- Listas: page=1, pageSize=20 (máximo 100), q y filtros documentados por módulo; orden estable por ID. Fechas desde inclusiva/hasta exclusiva, RFC3339 UTC.
- Errores: {error: {code, message, fields?}}; 400 JSON inválido, 401 sin sesión, 403 sin permiso, 404 recurso inexistente, 409 conflicto/referencias/reintento incompatible, 422 validación, 503 persistencia no disponible. Sin SQL ni secretos.
- Toda ruta de negocio requiere sesión y permisos; login es la excepción. CSRF en escrituras con cookie. Permisos y alcance del objeto se verifican en Go.
- Pagos, entregas y creación de OT: acordar y persistir Idempotency-Key antes de habilitar reintentos automáticos. El SQL actual no tiene almacenamiento de claves; restricción única de OT/entrega no reemplaza deduplicación de pagos.
- Catálogos nuevos: alta explícita con permiso de editar cotizaciones, sin crear módulos visuales. No crear por inferencia silenciosa.

## Catálogo

### Acceso

| Método | Ruta | Operación |
|---|---|---|
| POST | `/api/v1/auth/login` | Autenticar; usuario/correo y contraseña |
| POST | `/api/v1/auth/logout` | Revocar sesión |
| GET | `/api/v1/auth/me` | Actor y permisos efectivos |

### Inicio

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/dashboard` | Indicadores y gráficos derivados |

### Clientes

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/clientes` | Buscar por nombre, documento o placa |
| POST | `/api/v1/clientes` | Registrar cliente |
| GET | `/api/v1/clientes/{id}` | Detalle |
| PATCH | `/api/v1/clientes/{id}` | Editar campos permitidos |
| DELETE | `/api/v1/clientes/{id}` | Eliminar solo sin referencias |
| GET | `/api/v1/clientes/{id}/contactos` | Contactos vinculados |
| POST | `/api/v1/clientes/{id}/contactos` | Alta de contacto; no autoriza retiro |

### Vehículos

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/vehiculos` | Filtrar por placa y clienteId |
| POST | `/api/v1/vehiculos` | Alta con historial inicial |
| GET | `/api/v1/vehiculos/{id}` | Detalle y propietario |
| PATCH | `/api/v1/vehiculos/{id}` | Actualizar datos conservando placa |
| DELETE | `/api/v1/vehiculos/{id}` | Eliminar solo sin historial operativo; tratar historial inicial en transacción |
| POST | `/api/v1/vehiculos/{id}/transferencias` | Transferencia verificada e historial |
| GET | `/api/v1/vehiculos/{id}/historial` | Atenciones, diagnósticos, OT y transferencias |

### Recepciones

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/recepciones` | Listado; filtro pendienteDiagnostico |
| POST | `/api/v1/recepciones` | Registrar y actualizar kilometraje |
| GET | `/api/v1/recepciones/{id}` | Consulta guardada |

### Diagnósticos

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/diagnosticos` | Listado; filtro recepcionId |
| POST | `/api/v1/diagnosticos` | Guardar resumen y técnico autenticado |
| GET | `/api/v1/diagnosticos/{id}` | Detalle |
| GET | `/api/v1/diagnosticos/{id}/escaneres` | Resultados técnicos MongoDB |
| POST | `/api/v1/diagnosticos/{id}/escaneres` | Capturar resultado; requestId requerido |
| GET | `/api/v1/diagnosticos/{id}/evidencias` | Metadatos técnicos MongoDB |
| POST | `/api/v1/diagnosticos/{id}/evidencias` | Subir evidencia multipart; requestId requerido |
| GET | `/api/v1/diagnosticos/{id}/evidencias/{evidenciaId}/archivo` | Descarga autorizada desde GridFS |

### Cotizaciones

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/cotizaciones` | Listado; diagnosticoId, estado, sinOrden |
| POST | `/api/v1/cotizaciones` | Guardar cabecera y líneas |
| GET | `/api/v1/cotizaciones/{id}` | Detalle con totales |
| PUT | `/api/v1/cotizaciones/{id}` | Reemplazar líneas/estado; bloquear si tiene OT |

### Órdenes de trabajo

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/ordenes` | Listado; estado, búsqueda y asignación |
| POST | `/api/v1/ordenes` | Crear desde cotización aprobada |
| GET | `/api/v1/ordenes/{id}` | Detalle completo, asignaciones, avances, pruebas e historial |
| PATCH | `/api/v1/ordenes/{id}` | Actualizar estado, personal, avance y prueba atómicamente |
| DELETE | `/api/v1/ordenes/{id}` | Eliminar solo según reglas; conservar con pagos/entrega |

### Pagos y entrega

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/ordenes/{id}/pagos` | Pagos y resumen total/pagado/saldo |
| POST | `/api/v1/ordenes/{id}/pagos` | Registrar pago; bloqueo y control de saldo |
| GET | `/api/v1/ordenes/{id}/entrega` | Consultar retiro; 404 si no existe |
| POST | `/api/v1/ordenes/{id}/entrega` | Verificar receptor, saldo y estado; guardar snapshots |

### Auxilio y tareas

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/auxilios` | Consultar atenciones externas |
| POST | `/api/v1/auxilios` | Registrar auxilio |
| GET | `/api/v1/auxilios/{id}` | Detalle |
| GET | `/api/v1/me/tareas` | OT activas del actor, sin ID fijo demo |

### Mecánicos

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/mecanicos` | Usuarios técnicos y conteo de OT activas |
| GET | `/api/v1/asignaciones` | Asignaciones activas; filtros ordenId/usuarioId |

### Proveedores

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/proveedores` | Buscar nombre/documento |
| POST | `/api/v1/proveedores` | Alta; resolver parts antes de implementación |
| GET | `/api/v1/proveedores/{id}` | Detalle |
| PATCH | `/api/v1/proveedores/{id}` | Editar; resolver parts antes de implementación |
| DELETE | `/api/v1/proveedores/{id}` | Eliminar solo sin referencias |

### Reportes

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/reportes` | Servicios e ingresos; desde/hasta |

### Usuarios, roles y permisos

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/usuarios` | Listado sin hashes |
| POST | `/api/v1/usuarios` | Alta y asignación de rol |
| GET | `/api/v1/usuarios/{id}` | Detalle sin hashes |
| PATCH | `/api/v1/usuarios/{id}` | Editar usuario/estado/rol |
| DELETE | `/api/v1/usuarios/{id}` | Solo sin referencias; preservar administración |
| GET | `/api/v1/roles` | Catálogo real |
| GET | `/api/v1/permisos` | Catálogo modulo.accion |
| GET | `/api/v1/roles/{id}/permisos` | Matriz por rol |
| PUT | `/api/v1/roles/{id}/permisos` | Reemplazar matriz validada atómicamente |
| POST | `/api/v1/roles/{id}/permisos/restablecer` | Restaurar política inicial acordada con auditoría |

### Catálogos de apoyo

| Método | Ruta | Operación |
|---|---|---|
| GET | `/api/v1/servicios` | Catálogo para líneas de cotización |
| POST | `/api/v1/servicios` | Alta explícita autorizada para concepto nuevo |
| GET | `/api/v1/repuestos` | Catálogo para líneas de cotización |
| POST | `/api/v1/repuestos` | Alta explícita autorizada para repuesto nuevo |
| GET | `/api/v1/unidades-auxilio` | Unidades disponibles con ID/código |
| GET | `/api/v1/proveedores/{id}/repuestos` | Relación y costos de referencia |
| PUT | `/api/v1/proveedores/{id}/repuestos/{repuestoId}` | Vincular repuesto y costo; contrato futuro de catálogo |

## Payloads de escritura (propuesta para implementación)

Campos de alta requeridos salvo que se indique opcional; PATCH admite subconjunto de campos editables. Solo recepción/diagnóstico/auxilio/OT toman el actor de sesión. En todos los módulos las fechas de auditoría y creadores son del servidor.

| Operación | Campos de entrada |
|---|---|
| Login | username (usuario o correo), password; remember opcional solo para recordar usuario, no contraseña |
| Cliente | name, documentType, documentNumber, phone, address; email opcional |
| Vehículo | clientId, plate, brand, model, year, mileage; notes opcional. PATCH no admite plate/clientId; propietario solo por transferencia. |
| Transferencia | newClientId, transferVerified=true; observation opcional |
| Contacto | name, documentType, documentNumber, phone, relation; clientId viene de ruta |
| Recepción | clientId, vehicleId, mileage, symptoms |
| Diagnóstico | receptionId, findings |
| Cotización | diagnosisId, status, items: [{concept, type, serviceId o sparePartId, quantity, unitPrice, origin}]; PUT reemplaza lista completa; total no es entrada |
| Nueva OT | quoteId, mechanicId; assistantId opcional |
| Actualizar OT | state, note; mechanicId/assistantId requieren permiso asignar; test/result opcionales, juntos cuando hay prueba |
| Pago | amount, method; comprobante opcional; orderId viene de ruta |
| Entrega | kind, method, verified=true; contactId requerido si Tercero; snapshots desde SQL |
| Auxilio | clientId, vehicleId, location, incident, description, technicianId, unitId (sustituye unit textual) |
| Proveedor | name, ruc, phone, parts; parts pendiente de decisión de persistencia, no habilitar alta que lo descarte |
| Usuario | name, username, email, roleId, active; password requerido solo en alta. roleId se resuelve desde select role de la UI |
| Matriz de permisos | permissions: {modulo: [accion, ...]}; reemplazo completo; validar contra catálogo y reglas de acceso obligatorio |
| Restablecer permisos | Sin body; política por defecto del servidor, nunca desde navegador |
| Servicio / repuesto | name, referencePrice; stock inicial opcional solo repuesto (decimal), active opcional |
| Repuesto/proveedor | referenceCost; IDs en ruta |
| Escáner | requestId, equipo, dtcs, mediciones; capturadoEn opcional (hora de captura del equipo), capturadoPor/registradoEn del servidor |
| Evidencia | multipart: archivo, requestId, descripcion; tamaño/hash/MIME verificados por servidor |

La correspondencia SQL completa y los límites de longitud están en CORRESPONDENCIA.md y ESQUEMA_POSTGRESQL.md. Esto es un contrato de arquitectura: DTO formales y OpenAPI detallado se escribirán al resolver las diferencias, antes de activar los handlers. El catálogo Go y estas rutas deben mantenerse alineados.
