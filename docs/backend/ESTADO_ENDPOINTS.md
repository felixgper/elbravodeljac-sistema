# Estado de las 71 operaciones REST

Catálogo de A–D, conservado y vuelto a probar en Fase E. Bases reales del usuario no utilizadas. TestOriginalCatalogRelationalCoverage comprueba los 71 handlers originales y sus permisos. MongoDB, GridFS y PostgreSQL se probaron juntos en instancias temporales.

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
| GET | `/api/v1/diagnosticos/{id}/escaneres` | Diagnósticos | Implementada |
| POST | `/api/v1/diagnosticos/{id}/escaneres` | Diagnósticos | Implementada |
| GET | `/api/v1/diagnosticos/{id}/evidencias` | Diagnósticos | Implementada |
| POST | `/api/v1/diagnosticos/{id}/evidencias` | Diagnósticos | Implementada |
| GET | `/api/v1/diagnosticos/{id}/evidencias/{evidenciaId}/archivo` | Diagnósticos | Implementada |
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

Total: 71/71 originales implementadas; 74 operaciones de negocio reales contando estas tres ampliaciones. GET /healthz y GET /readyz son operacionales y no se cuentan. Un rechazo 401/403/409/422 por reglas de seguridad o validación no es un placeholder.


Fase E agrega `GET /api/v1/frontend/opciones` para las proyecciones auxiliares
permitidas de cada formulario, con sesión, permisos y alcance por asignación.
No cambia las 71 operaciones originales: son 75 rutas de aplicación contando
las cuatro ampliaciones. `/`, `/demo/` y los archivos estáticos no se cuentan
como operaciones de negocio. [Pantallas y endpoints utilizados](FASE_E.md).


Revisión funcional 27/10/2026 agrega `GET /api/v1/vehiculos/marcas`, con permiso
`vehiculos.ver`: catálogo central de la vista PostgreSQL `marcas_vehiculos`, con
las 34 marcas aprobadas y marcas adicionales de vehículos persistidos. Son **76 rutas
de aplicación** contando `frontend/opciones`, las tres ampliaciones C y marcas.
Las 71 originales siguen implementadas, sin introducir respuestas 501.
`GET /dashboard?period=Semana|Mes|Año` y `GET /reportes?period=Semana|Mes|Año`
usan límites del negocio en Lima. `POST /cotizaciones` / `PUT /cotizaciones/{id}` permite conceptos
manuales con `type`, `concept`, `quantity`, `unitPrice`, `discount` y
`partKind` (Original/Alternativo para repuesto); `serviceId`/`sparePartId` son opcionales
y se validan cuando se envían. Ver [migración y pruebas](REVISION_12_OBSERVACIONES.md).

[Corrección de origen de marcas, migración y pruebas](CORRECCION_MARCAS.md).


Mejoras de atención/ficha (27/10/2026): GET y PUT
`/api/v1/ordenes/{id}/control-tecnico`, y GET
`/api/v1/ordenes/{id}/ficha-tecnica.pdf`. Son 79 operaciones de aplicación contando
las ampliaciones anteriores; las 71 originales siguen implementadas. Control
requiere ordenes.ver/editar y alcance de asignación; PDF exige pagos.ver, saldo cero
y entrega. Reportes entrega días en Semana/Mes y meses en Año. Ver
[contratos, migración 003 y pruebas](ATENCIONES_REPORTES_FICHA.md).


## Asignación y diagnósticos (27/10/2026)

Se incorporaron cuatro operaciones funcionales, verificadas con PostgreSQL:

| Método | Ruta | Permiso y alcance |
|---|---|---|
| PATCH | `/api/v1/recepciones/{id}/mecanico` | `diagnosticos.asignar`; mecánico activo, versión de asignación |
| GET | `/api/v1/me/diagnosticos` | `diagnosticos.ver`; pendientes propios o supervisión explícita |
| PUT | `/api/v1/diagnosticos/{id}` | `diagnosticos.crear`; asignado o supervisor, versión de borrador |
| GET | `/api/v1/diagnosticos/pendientes-cotizacion` | `diagnosticos.ver`; finalizados sin cotización, propios o supervisión |

Total acumulado: **83 operaciones de aplicación**, incluidas `frontend/opciones` y
las ampliaciones previas; las 71 originales permanecen implementadas. No se agregan
respuestas 501. El test de catálogo cubre 82 entradas de negocio y la ruta auxiliar
`frontend/opciones` se verifica por separado.

`POST /recepciones` admite `technicianId` para tipo `Diagnostico`, con permiso de
asignación. `POST /diagnosticos` crea borrador por defecto; `status: Finalizado`
requiere hallazgos, evaluación y recomendaciones. El servidor registra actor y
fechas; `PUT` exige la versión vigente. Finalizados inmutables. Las consultas de
recepciones/diagnósticos y sus opciones auxiliares verifican alcance en Go.
`POST /cotizaciones` rechaza borradores; Directo sigue sin diagnóstico.

[Flujo, compatibilidad de datos, migración y pruebas](FLUJO_DIAGNOSTICOS_ASIGNADOS.md).


Corrección de Cotizaciones: `GET /api/v1/cotizaciones/atenciones`, permiso
`cotizaciones.ver`, búsqueda real normalizada por `q`, `recepcionId` y paginación.
Devuelve orígenes directos o diagnósticos finalizados, con ID de cotización existente.
Son **84 rutas de aplicación** acumuladas (83 de negocio más frontend/opciones),
sin 501. POST cotizaciones rechaza duplicados por atención en transacción; PUT
mantiene el bloqueo si hay OT. Ver [implementación y pruebas](CORRECCION_COTIZACIONES.md).


Auditoría funcional (27/10/2026): PATCH `/api/v1/ordenes/{id}/pagos/{paymentId}`
(pagos.rectificar) y GET `/api/v1/clientes/{id}/historial` (clientes.ver, finanzas
según pagos.ver). Total **86 operaciones de aplicación**: 85 entradas de negocio
y frontend/opciones. Se conservan las 71 originales sin 501. Los contratos existentes
incorporan versión/aprobación, ampliación autorizada, ejecución/conciliación, anticipos
con requestId, reversión/sustitución y conformidad separada de entrega. Ver
[implementación, migraciones 005/006 y pruebas](CORRECCIONES_AUDITORIA_FUNCIONAL.md).


Cotizaciones PDF (27/10/2026): `GET /api/v1/cotizaciones/{id}/pdf`, permiso
`cotizaciones.ver`, devuelve `application/pdf` como archivo adjunto, con caché
privada deshabilitada. Disponible para Pendiente, Aprobada y Rechazada. Solo SELECT:
no modifica cotizaciones, detalles, aprobación, historial ni ficha técnica.
Incluye versión actual y versión de la aprobación registrada; no sustituye una
aprobación invalidada por otra ficticia. Total acumulado: **87 operaciones de
aplicación**, 86 de negocio y frontend/opciones; las 71 originales continúan.

Reportes conserva `GET /api/v1/reportes`: la interfaz envía `period`, `desde` y
`hasta` (exclusivo, zona Lima) usando el contrato existente. Hace una segunda
consulta del período de calendario anterior para comparar ingresos netos y
órdenes finalizadas. No se añadieron columnas, migraciones ni cálculos de venta.
Ver [descarga, organización y pruebas](COTIZACIONES_PDF_REPORTES.md).

Empleados (29/10/2026): GET/POST `/api/v1/empleados` y GET/PATCH
`/api/v1/empleados/{id}`. Total acumulado: **91 operaciones de aplicación**,
90 de negocio y frontend/opciones. POST/PATCH usuarios incorpora `employeeId`
sin reemplazar roles existentes. Ver [contrato y compatibilidad](EMPLEADOS.md).
