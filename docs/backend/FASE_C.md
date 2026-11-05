# Fase C: módulos relacionales implementados

Estado: **66 de las 71 operaciones originales implementadas**. Solo las cinco
operaciones MongoDB siguen como 501. La lista completa, una fila por operación,
está en [ESTADO_ENDPOINTS.md](ESTADO_ENDPOINTS.md). Además hay tres operaciones
de apoyo fuera del catálogo original: alta de unidades, reasignación de auxilios
y desvinculación proveedor/repuesto. Total de handlers de negocio reales: 69.

Rama: backend/arquitectura-go-mongodb. No se cambió el modelo de 26 tablas y 42 FK,
el SQL exportado ni dist/. No se conectó a ElBravoDelJac, no se usaron DROP/TRUNCATE,
no se importaron datos reales ni se publicaron secretos. Los DELETE implementan
únicamente solicitudes explícitas de eliminación de registros concretos, con FK
y rollback; no hay borrado masivo, cascadas automáticas ni destrucción del esquema.

## Usuarios y asignaciones N:M

GET/POST /usuarios, GET/PATCH/DELETE /usuarios/{id}; GET /roles y GET /permisos.
Permisos middleware: usuarios.ver/crear/editar/eliminar según operación. Datos de
usuario nunca incluyen password_hash. POST requiere name, username, password,
roleId; email y active opcionales. Nombres 150, usuario 60, correo 150; bcrypt
exige entre 12 y 72 bytes para nuevas contraseñas, sin truncar. Active inicia true.

PATCH acepta los mismos campos opcionales y removeRoleId. **roleId agrega** esa
asignación; no reemplaza las restantes. **removeRoleId retira explícitamente una**
asignación existente. Ambos pueden enviarse juntos para un cambio expresamente
solicitado, siempre que sean distintos. No se infiere qué rol adicional eliminar.
La futura conexión del select visual debe respetar esta semántica.

Roles, activación y restablecimiento de contraseñas requieren también usuarios.roles.
No se permite cambiar estos campos en la propia cuenta ni autoeliminarse. Un editor
sin usuarios.roles tampoco puede modificar una cuenta con acceso administrativo.
No se puede delegar un rol con permisos que el actor no posee. El borrado elimina
sus asignaciones solo dentro de la misma transacción que elimina al usuario; una
FK de operaciones/auditoría devuelve 409 y revierte también esas eliminaciones.

## Matriz de permisos

GET/PUT /roles/{id}/permisos y POST /roles/{id}/permisos/restablecer.
Lectura requiere usuarios.ver; cambios requieren usuarios.roles y revalidación
del actor dentro de la transacción. Payload completo, compatible con la estructura
de la matriz del frontend:

```json
{"permissions":{"clientes":["ver","crear","editar"],"vehiculos":["ver"]}}
```

La matriz reemplaza los permisos del rol solicitado de forma atómica. Rechaza
módulos/acciones desconocidas, duplicados y acciones sin ver; no limpia datos de
entrada silenciosamente. Los códigos deben existir en permisos de PostgreSQL.
Si el rol contiene códigos fuera de la matriz aprobada, la actualización se rechaza
para no borrarlos al reemplazar. Se requiere una decisión explícita para alinearlos.

permission-policy.json es la copia de la política aprobada del frontend, no un
script de sembrado: no crea roles ni permisos en la base. Restablecer aplica esa
política solo a uno de los seis roles conocidos, verificando catálogo y permisos
del actor. Un rol personalizado no tiene valores iniciales inventados.

No se permite editar ningún rol que el propio actor posea. Propietario es inmutable.
Administrador conserva usuarios.ver, usuarios.roles y usuarios.editar. Todo cambio
de identidad se serializa mediante un bloqueo advisory transaccional; se reconsulta
el permiso del actor después de adquirirlo. Antes de confirmar debe existir al
menos una cuenta activa que reúna esos tres permisos, incluso si provienen de
distintos roles. Así se evita perder el último acceso útil de gestión, también
ante peticiones concurrentes. No hay privilegios implícitos por nombre de rol.

Las sesiones se revocan tras cambios de contraseña/estado/asignaciones. El JWT
también está vinculado a una versión derivada del hash de credenciales y se compara
con PostgreSQL en cada petición, cubriendo la carrera entre login y cambio de clave.
Los JWT anteriores al despliegue de esta versión exigen iniciar sesión de nuevo.
Las sesiones siguen siendo de una instancia; compartirlas entre réplicas continúa
pendiente antes de producción.

## Auxilio mecánico y catálogos

GET/POST /auxilios; GET /auxilios/{id}; PATCH /auxilios/{id} adicional para reasignar
technicianId/unitId o cambiar state. Leer requiere auxilio.ver; escribir auxilio.crear,
la acción existente de gestión de solicitudes en la matriz aprobada.

Alta: clientId, vehicleId, technicianId, unitId, location, incident, description;
receptionId opcional. Estado inicial pendiente. Valida propietario, recepción
compatible, técnico activo Mecánico/Supervisor y unidad activa. Patch admite estados
pendiente/en_atencion/finalizado/cancelado; no reabre solicitudes finalizadas/canceladas.
No sobrescribe cliente/vehículo de la atención. Escrituras en transacción.

GET/POST /servicios y /repuestos; GET/POST /unidades-auxilio (POST es adicional).
Los GET de servicios/repuestos ya implementados en B requieren cotizaciones.ver;
altas requieren cotizaciones.editar. Unidades exige auxilio.ver para lectura y
auxilio.crear para alta. No se inventan acciones nuevas en la matriz.

- Servicios: name, referencePrice, active opcional; nombre 100.
- Repuestos: name (120), referencePrice, stock opcional (0.00), active opcional.
- Unidades: code (30), description (150), active opcional.

Importes y stock son strings decimales no negativos con máximo dos decimales,
compatibles con NUMERIC(12,2). No se redondea ni recorta entrada inválida. Se respetan
las restricciones únicas/FK existentes. Esta fase no añade ajuste de inventario,
descuento automático por cotización ni una tabla de movimientos no aprobada.

## Proveedores y diferencia de especialidad

GET/POST /proveedores y GET/PATCH/DELETE /proveedores/{id} implementados. Cada uno
requiere la acción correspondiente proveedores.ver/crear/editar/eliminar.
Contrato compatible: name (150), document (20), phone (20), contact (120). ruc es
alias de document para el formulario; si se envían ambos deben coincidir.

**parts/especialidad con contenido produce 422 explícito.** No se almacena en
contacto ni se omite aunque se envíen otros campos válidos: se rechaza la operación
completa, conservando la fila anterior. Los clientes reciben una explicación.

Alternativa compatible sin migración: seleccionar repuestos existentes y guardar
un costo de referencia por relación. Implementado mediante:

- GET /proveedores/{id}/repuestos.
- PUT /proveedores/{id}/repuestos/{repuestoId} con referenceCost decimal no negativo.
- DELETE de esa misma relación (adicional al catálogo), solicitado explícitamente.

PUT actualiza o crea una única relación existente del modelo; no duplica vínculos.
Vincular/desvincular exige proveedores.editar. No se borra un proveedor con vínculos:
la FK devuelve 409 y exige decidir explícitamente su desvinculación.

Esta alternativa modela repuestos concretos, **no equivale a especialidad libre**.
Si se desea conservar ese texto como tal, deberá aprobarse una futura migración
antes de implementarla. No se añadió columna ni se cambió el frontend en esta fase.

## Dashboard y reportes

GET /dashboard requiere inicio.ver; GET /reportes requiere reportes.ver. Consultas
reales y agregados SQL sobre pagos, estados, asignaciones, servicios, auxilios y
registros. Dashboard incluye ingreso mensual, OT activas/finalizadas, auxilios de
hoy, seis meses de ingresos, estados, personal, servicios y conteos del taller.

Reportes acepta period=Este mes o Últimos 6 meses, o desde/hasta RFC3339 juntos
(hasta exclusivo, máximo dos años). Incluye ingresos por mes y total, servicios
de OT finalizadas, auxilios del intervalo y cantidad de OT finalizadas. Calendario
mensual/today en America/Lima; timestamps de API con zona.

Pagos se agregan directamente y por separado de los detalles de cotización y
asignaciones. Una OT con varios conceptos no multiplica el ingreso. Los meses sin
pagos aparecen en cero. Precios/ingresos y cents se devuelven como strings exactos;
los conteos son números. El adaptador del frontend deberá convertir/presentar
esos valores con precisión al activar la conexión; no se sustituyó la demo.

## Verificación ejecutada

Se ejecutó scripts/test-local-postgres.ps1: nuevo clúster, puerto loopback aleatorio,
base exacta elbravodeljac_test, esquema original solo en esa base nueva, registros
sintéticos de pruebas y servidor HTTP/PostgreSQL reales. El clúster se detuvo al
finalizar. No se simularon resultados ni se usaron datos personales reales.

Pasaron go test ./... -count=1 -v (A+B+C), go vet ./..., pruebas unitarias de matriz,
precios/stock y longitudes, y las tres verificaciones existentes del frontend.
TestOriginalCatalogRelationalCoverage comprueba 71 originales, 66 implementadas,
cinco reservas MongoDB, tres ampliaciones y permisos en todos los handlers reales.

La suite C comprobó: CRUD de usuarios; N:M sin sobrescritura; retirada explícita;
rechazo de autoasignación/cambio del propio rol; rollback de matriz inválida;
revocación de JWT tras contraseña; borrado bloqueado por FK sin perder roles;
catálogos y valores inválidos; proveedores/especialidad sin pérdida; vínculos y
costos; auxilios y asignaciones; protección del último administrador y carrera
entre dos administradores. Dashboard/Reportes se contrastaron con SUM(pagos)
independiente usando una OT creada en C con dos detalles y dos pagos parciales.
La base al terminar mantuvo 26 tablas y 42 FK.

No se ejecutó Docker en esta fase: se utilizó la instancia local aislada, con el
mismo script aprobado de A/B. El camino Docker sigue disponible para comprobación
adicional cuando se disponga de su motor. No se afirma haberlo probado aquí.

## MongoDB: cinco operaciones restantes

- GET/POST /diagnosticos/{id}/escaneres.
- GET/POST /diagnosticos/{id}/evidencias.
- GET /diagnosticos/{id}/evidencias/{evidenciaId}/archivo.

Faltan driver, configuración privada, repositorios, validadores/índices, binarios
GridFS o almacenamiento acordado, permisos por diagnóstico, idempotencia,
compensación y pruebas de fallos entre motores. No se moverán transacciones SQL
a MongoDB ni se promete atomicidad común entre los dos motores.
