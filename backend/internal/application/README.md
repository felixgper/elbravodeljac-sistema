# Casos de uso

Implementar aquí servicios de aplicación por módulo: identidad, clientes, vehículos,
recepciones, diagnósticos, cotizaciones, órdenes, pagos/entregas, auxilios y consultas.
Reciben contexto, actor autenticado, DTO validado e interfaces de repositorio.
No reciben http.Request ni dependen de controladores concretos de bases.

Una operación de negocio completa delimita la transacción PostgreSQL. No repartir
crear/actualizar una OT en escrituras HTTP independientes que puedan quedar a medias.
Consultar docs/backend/CORRESPONDENCIA.md antes de definir los DTO definitivos.
