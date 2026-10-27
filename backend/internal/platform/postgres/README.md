# Adaptador PostgreSQL (pendiente)

Implementar pool, consultas parametrizadas, repositorios y unidad transaccional.
Esquema de referencia: database/postgresql/esquema_actual.sql, solo lectura.
No ejecutar ese dump al arrancar ni instalar migraciones automáticamente.
Conectar a ElBravoDelJac solo mediante configuración privada y usuario de mínimo
privilegio. Las pruebas de integración usarán otra base vacía y descartable.

Pagos y entregas bloquean primero la misma fila ordenes_trabajo (FOR UPDATE),
recalculan saldo dentro de la transacción y validan estado. Cotización y creación
de OT comparten bloqueo sobre cotizaciones para impedir ediciones concurrentes.
La transferencia bloquea vehiculos, cierra el historial vigente y abre el nuevo
en la misma transacción. No depender solo de los dos triggers del dump.

No devolver errores SQL ni DSN al cliente. Traducir 23505 a conflicto, 23503 a
referencia inválida/conflicto y 23514 a validación de dominio según contexto.
