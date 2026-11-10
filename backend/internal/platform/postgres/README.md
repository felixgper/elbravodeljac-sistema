# Adaptador PostgreSQL

`store.go` implementa acceso/roles/permisos, Clientes, contactos y Vehículos con
pgxpool y transacciones. Las demás operaciones de este documento son el plan
para fases posteriores. No se ejecuta SQL de inicialización desde el servidor.

Implementar pool, consultas parametrizadas, repositorios y unidad transaccional.
Modelo físico único: database/postgresql/esquema_entrega.sql.
Docker lo utiliza solo para bases nuevas. Las bases ya existentes siguen su ruta histórica de migraciones; la API no ejecuta el esquema final sobre ellas.
Conectar a ElBravoDelJac solo mediante configuración privada y usuario de mínimo
privilegio. Las pruebas de integración usarán otra base vacía y descartable.

Pagos y entregas bloquean primero la misma fila ordenes_trabajo (FOR UPDATE),
recalculan saldo dentro de la transacción y validan estado. Cotización y creación
de OT comparten bloqueo sobre cotizaciones para impedir ediciones concurrentes.
La transferencia bloquea vehiculos, cierra el historial vigente y abre el nuevo
en la misma transacción. No depender solo de los dos triggers del dump.

No devolver errores SQL ni DSN al cliente. Traducir 23505 a conflicto, 23503 a
referencia inválida/conflicto y 23514 a validación de dominio según contexto.
