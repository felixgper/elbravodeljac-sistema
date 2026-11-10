# Endpoints REST vigentes

Base `/api/v1`. El catálogo y sus permisos están implementados en [routes.go](../../backend/internal/httpapi/routes.go) y sus handlers. [MongoDB](../../database/mongodb/README.md) · [Pruebas aisladas](../../checks/README.md).

## Convenciones

El catálogo documental se consulta en `GET /tipos-documento` con `clientes.ver` o `pagos.ver`. Los tipos iniciales son DNI, RUC, CE y PASAPORTE. `documentType` y `documentNumber` mantienen su formato JSON. Se puede conservar una identidad histórica inactiva al editar otros campos. Cambiarla exige un tipo activo. La identidad del receptor de una entrega queda en su snapshot.

`GET /roles` devuelve `active`. Los roles disponibles para nuevas asignaciones son Gerente, Administrador y Mecánico. Un rol legado inactivo deja de ofrecerse, pero conserva los permisos de las cuentas que lo mantienen asignado. La asignación de roles y la edición de matrices requieren autorización y no permiten conceder permisos que el actor no posea. Se protegen la matriz de Gerente, el último acceso administrativo y, una vez establecido, el último Gerente activo. Cambiar roles o desactivar una cuenta revoca sus sesiones.

- JSON UTF-8. IDs SQL como strings decimales; campos JSON camelCase del formulario. Rutas {id} usan ID SQL, nunca codigo_ot.
- Dinero, km y cantidades: strings decimales exactos; respuesta con dos decimales. total/subtotal/saldo son calculados por servidor.
- Respuestas: 200 lectura/edición; 201 alta con Location; 204 eliminación/cierre de sesión. Detalle {data: objeto}; listas {data: [], meta: {page, pageSize, total}}.
- Listas: page=1, pageSize=20 (máximo 100), q y filtros documentados por módulo; orden estable por ID. Fechas desde inclusiva/hasta exclusiva, RFC3339 UTC.
- Errores: {error: {code, message, fields?}}; 400 JSON inválido, 401 sin sesión, 403 sin permiso, 404 recurso inexistente, 409 conflicto/referencias/reintento incompatible, 422 validación, 503 persistencia no disponible. Sin SQL ni secretos.
- Toda ruta de negocio requiere sesión y permisos; login es la excepción. CSRF en escrituras con cookie. Permisos y alcance del objeto se verifican en Go.
- Pagos: enviar `requestId` obligatorio y estable en los reintentos. La migración 006 incorpora `pagos.request_id` y su índice único por orden; un reintento con el mismo contenido recupera el pago, y un contenido incompatible devuelve 409. La unicidad de orden/cotización y entrega/orden no implica idempotencia genérica para las demás altas; no reintentarlas automáticamente sin comprobar su resultado.
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
| POST | `/api/v1/proveedores` | Alta con datos del proveedor |
| GET | `/api/v1/proveedores/{id}` | Detalle |
| PATCH | `/api/v1/proveedores/{id}` | Edición de datos del proveedor |
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
| PUT | `/api/v1/proveedores/{id}/repuestos/{repuestoId}` | Vincular repuesto y costo de referencia |

## DTO y validaciones vigentes

Las altas y ediciones de proveedores rechazan un valor no vacío en `parts` con 422. Para registrar repuestos y costos de referencia, utiliza las rutas de vinculación de proveedores y repuestos.

La fuente ejecutable de campos, longitudes, compatibilidad y validaciones es:

- [Clientes y vehículos](../../backend/internal/domain/models.go).
- [Recepciones, diagnósticos, cotizaciones, órdenes, pagos y entrega](../../backend/internal/domain/workshop.go).
- [Usuarios, roles, proveedores y catálogos](../../backend/internal/domain/relational.go).
- [Empleados](../../backend/internal/domain/employee.go).
- [Escáner y evidencias](../../backend/internal/domain/technical.go).

### Empleados y cuentas

`GET /empleados` y `GET /empleados/{id}` requieren `usuarios.ver`. El alta por POST requiere `usuarios.crear` y la edición parcial por PATCH requiere `usuarios.editar`. El alta acepta `name` (hasta 150 caracteres), `employeeType` (hasta 40) y `active` (true por defecto). El empleado puede existir sin cuenta y su cargo laboral no concede permisos. Se cambia su estado laboral sin borrarlo físicamente.

El POST o PATCH de usuarios acepta `employeeId`. Vincular requiere además `usuarios.roles` y se realiza en la misma transacción que la cuenta y sus roles. Un vínculo ocupado devuelve 409, una referencia inexistente devuelve 404 y un empleado inactivo devuelve 422. El nombre de una cuenta vinculada se administra desde Empleados. Desactivar el empleado no desactiva su cuenta. Los registros operativos siguen referenciando usuarios y el sistema conserva la relación de múltiples roles por cuenta.

### Recepción, diagnóstico y órdenes

Una recepción directa no crea diagnóstico. Para atención técnica, la interfaz exige un mecánico asignado. Las recepciones técnicas sin asignación admitidas por compatibilidad deben asignarse antes de diagnosticar. Los permisos de ver, asignar y supervisar diagnósticos se verifican junto al alcance del registro. Una asignación a una orden no otorga supervisión general de diagnósticos.

El diagnóstico admite Borrador o Finalizado. La finalización exige hallazgos, evaluación y recomendaciones. Las observaciones son opcionales. La edición usa `version` y la reasignación usa `assignmentVersion`. Un conflicto de versión devuelve 409. Reasignar conserva borrador e historial y cambia el acceso del técnico anterior. Un diagnóstico finalizado es inmutable y habilita la cotización técnica.

La cotización exige recepción y diagnóstico finalizado cuando corresponde. Las referencias de conceptos manuales son opcionales y validadas. La orden exige aprobación y un mecánico de ejecución, que puede ser distinto del responsable del diagnóstico. Las nuevas asignaciones de auxiliar y los cambios de `assistantId` devuelven 422. Las asignaciones históricas se conservan. Pago requiere `requestId` estable y entrega verifica saldo, identidad y conformidad.

Los PDF archivados de entrega se devuelven byte por byte. Para su generación y descarga, consulta el [contrato de PDF](../../backend/internal/platform/technicalsheet/README.md).
