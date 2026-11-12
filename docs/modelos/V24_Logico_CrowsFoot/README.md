# V24 actualizado — modelo entidad-relación físico de SQL Server 2025

**Se conserva el diseño original V24**: una sola lámina, las mismas seis áreas, posiciones, colores, 28 tablas, 208 columnas y 50 relaciones Crow's Foot. No se ha vuelto a dibujar ni se ha modificado el backend, los datos o el motor.

## Archivo de trabajo vigente

[Modelo V24 editable en draw.io](El_Bravo_del_JAC_V24_Logico_Completo.drawio)

El modelo físico toma los **tipos declarados en** [database/sqlserver/esquema_entrega.sql](../../../database/sqlserver/esquema_entrega.sql): `BIGINT`, `NVARCHAR(n/MAX)`, `DECIMAL(p,s)`, `DATETIMEOFFSET(6)`, `BIT`, `VARBINARY(MAX)`, etc. El tamaño de algunas cadenas se restringe adicionalmente mediante `CHECK`; esos límites no son el mismo concepto que el tipo declarado.

La relación de `cotizaciones.id_cotizacion_base` con `cotizaciones.id_cotizacion` se conserva porque **ya existe** en la base y permite enlazar revisiones; no es una relación creada al actualizar el dibujo.

### Recorrido breve del negocio

`CLIENTES → VEHÍCULOS → RECEPCIONES`. La recepción guarda `id_cliente_atencion` y `id_vehiculo`, dos claves foráneas distintas. Si la atención requiere diagnóstico, sigue `RECEPCIONES → DIAGNÓSTICOS → COTIZACIONES`; si es servicio directo, sigue `RECEPCIONES → COTIZACIONES`. Después pueden existir `ÓRDENES_TRABAJO → PAGOS / ENTREGAS`. La secuencia es funcional; el diagrama conserva las cardinalidades y relaciones SQL reales.

**Pendiente, no aplicado en SQL Server:** se acordó renombrar `recepciones.id_recepcionista` a `id_usuario_registro`. Hasta que se actualicen SQL Server y Go de forma coordinada, el diagrama físico conserva **`id_recepcionista`**, que apunta a `usuarios.id_usuario` y registra a la persona usuaria que hizo la recepción. No debe interpretarse como una tabla de recepcionistas.

## Modelo lógico asociado

[Diagrama lógico Crow's Foot sin tipos](../Logico_CrowsFoot_Sin_Tipos/El_Bravo_del_JAC_Logico_CrowsFoot.drawio).

## Archivos PDF y PNG anteriores

Los PDF y PNG V24 antiguos **se retiraron de esta rama** porque seguían mostrando PostgreSQL. Sus versiones originales se pueden recuperar desde el historial Git. Para obtener PDF y PNG actualizados, abre el `.drawio` y usa Archivo → Exportar como, manteniendo la página completa.

**No ejecutar scripts SQL del repositorio para “aplicar” cambios del dibujo:** esta revisión es exclusivamente documental.
