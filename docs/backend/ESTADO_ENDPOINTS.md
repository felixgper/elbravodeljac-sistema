# Estado de las 71 operaciones REST

Estado al completar Fase C. Base de datos real no utilizada. El test TestOriginalCatalogRelationalCoverage comprueba el registro de los 66 handlers relacionales y las cinco reservas MongoDB.

| Método | Ruta | Módulo | Estado |
|---|---|---|---|
| POST | `/api/v1/auth/login` | Acceso | Implementada |
| POST | `/api/v1/auth/logout` | Acceso | Implementada |
| GET | `/api/v1/auth/me` | Acceso | Implementada |
| GET | `/api/v1/dashboard` | Inicio | Implementada |
| GET | `/api/v1/clientes` | Clientes | Implementada |
| POST | `/api/v1/clientes` | Clientes | Implementada |
| GET | `/api/v1/clientes/{id}` | Clientes | Implementada |
| PATCH | `/api/v1/clientes/{id}` | Clientes | Implementada |
| DELETE | `/api/v1/clientes/{id}` | Clientes | Implementada |
| GET | `/api/v1/clientes/{id}/contactos` | Clientes | Implementada |
| POST | `/api/v1/clientes/{id}/contactos` | Clientes | Implementada |
| GET | `/api/v1/vehiculos` | Vehículos | Implementada |
| POST | `/api/v1/vehiculos` | Vehículos | Implementada |
| GET | `/api/v1/vehiculos/{id}` | Vehículos | Implementada |
| PATCH | `/api/v1/vehiculos/{id}` | Vehículos | Implementada |
| DELETE | `/api/v1/vehiculos/{id}` | Vehículos | Implementada |
| POST | `/api/v1/vehiculos/{id}/transferencias` | Vehículos | Implementada |
| GET | `/api/v1/vehiculos/{id}/historial` | Vehículos | Implementada |
| GET | `/api/v1/recepciones` | Recepciones | Implementada |
| POST | `/api/v1/recepciones` | Recepciones | Implementada |
| GET | `/api/v1/recepciones/{id}` | Recepciones | Implementada |
| GET | `/api/v1/diagnosticos` | Diagnósticos | Implementada |
| POST | `/api/v1/diagnosticos` | Diagnósticos | Implementada |
| GET | `/api/v1/diagnosticos/{id}` | Diagnósticos | Implementada |
| GET | `/api/v1/diagnosticos/{id}/escaneres` | Diagnósticos | 501: MongoDB pendiente |
| POST | `/api/v1/diagnosticos/{id}/escaneres` | Diagnósticos | 501: MongoDB pendiente |
| GET | `/api/v1/diagnosticos/{id}/evidencias` | Diagnósticos | 501: MongoDB pendiente |
| POST | `/api/v1/diagnosticos/{id}/evidencias` | Diagnósticos | 501: MongoDB pendiente |
| GET | `/api/v1/diagnosticos/{id}/evidencias/{evidenciaId}/archivo` | Diagnósticos | 501: MongoDB pendiente |
| GET | `/api/v1/cotizaciones` | Cotizaciones | Implementada |
| POST | `/api/v1/cotizaciones` | Cotizaciones | Implementada |
| GET | `/api/v1/cotizaciones/{id}` | Cotizaciones | Implementada |
| PUT | `/api/v1/cotizaciones/{id}` | Cotizaciones | Implementada |
| GET | `/api/v1/ordenes` | Órdenes de trabajo | Implementada |
| POST | `/api/v1/ordenes` | Órdenes de trabajo | Implementada |
| GET | `/api/v1/ordenes/{id}` | Órdenes de trabajo | Implementada |
| PATCH | `/api/v1/ordenes/{id}` | Órdenes de trabajo | Implementada |
| DELETE | `/api/v1/ordenes/{id}` | Órdenes de trabajo | Implementada |
| GET | `/api/v1/ordenes/{id}/pagos` | Pagos y entrega | Implementada |
| POST | `/api/v1/ordenes/{id}/pagos` | Pagos y entrega | Implementada |
| GET | `/api/v1/ordenes/{id}/entrega` | Pagos y entrega | Implementada |
| POST | `/api/v1/ordenes/{id}/entrega` | Pagos y entrega | Implementada |
| GET | `/api/v1/auxilios` | Auxilio y tareas | Implementada |
| POST | `/api/v1/auxilios` | Auxilio y tareas | Implementada |
| GET | `/api/v1/auxilios/{id}` | Auxilio y tareas | Implementada |
| GET | `/api/v1/me/tareas` | Auxilio y tareas | Implementada |
| GET | `/api/v1/mecanicos` | Mecánicos | Implementada |
| GET | `/api/v1/asignaciones` | Mecánicos | Implementada |
| GET | `/api/v1/proveedores` | Proveedores | Implementada |
| POST | `/api/v1/proveedores` | Proveedores | Implementada |
| GET | `/api/v1/proveedores/{id}` | Proveedores | Implementada |
| PATCH | `/api/v1/proveedores/{id}` | Proveedores | Implementada |
| DELETE | `/api/v1/proveedores/{id}` | Proveedores | Implementada |
| GET | `/api/v1/reportes` | Reportes | Implementada |
| GET | `/api/v1/usuarios` | Usuarios, roles y permisos | Implementada |
| POST | `/api/v1/usuarios` | Usuarios, roles y permisos | Implementada |
| GET | `/api/v1/usuarios/{id}` | Usuarios, roles y permisos | Implementada |
| PATCH | `/api/v1/usuarios/{id}` | Usuarios, roles y permisos | Implementada |
| DELETE | `/api/v1/usuarios/{id}` | Usuarios, roles y permisos | Implementada |
| GET | `/api/v1/roles` | Usuarios, roles y permisos | Implementada |
| GET | `/api/v1/permisos` | Usuarios, roles y permisos | Implementada |
| GET | `/api/v1/roles/{id}/permisos` | Usuarios, roles y permisos | Implementada |
| PUT | `/api/v1/roles/{id}/permisos` | Usuarios, roles y permisos | Implementada |
| POST | `/api/v1/roles/{id}/permisos/restablecer` | Usuarios, roles y permisos | Implementada |
| GET | `/api/v1/servicios` | Catálogos de apoyo | Implementada |
| POST | `/api/v1/servicios` | Catálogos de apoyo | Implementada |
| GET | `/api/v1/repuestos` | Catálogos de apoyo | Implementada |
| POST | `/api/v1/repuestos` | Catálogos de apoyo | Implementada |
| GET | `/api/v1/unidades-auxilio` | Catálogos de apoyo | Implementada |
| GET | `/api/v1/proveedores/{id}/repuestos` | Catálogos de apoyo | Implementada |
| PUT | `/api/v1/proveedores/{id}/repuestos/{repuestoId}` | Catálogos de apoyo | Implementada |

Además del catálogo original se implementaron tres operaciones de apoyo:

- `POST /api/v1/unidades-auxilio`: alta autorizada.
- `PATCH /api/v1/auxilios/{id}`: reasignar técnico/unidad o cambiar estado.
- `DELETE /api/v1/proveedores/{id}/repuestos/{repuestoId}`: desvinculación explícita.

Total: 66/71 originales implementadas; 69 operaciones de negocio reales contando estas tres ampliaciones. GET /healthz y GET /readyz son operacionales y no se cuentan. Un rechazo 401/403/409/422 por reglas de seguridad o validación no es un placeholder.
