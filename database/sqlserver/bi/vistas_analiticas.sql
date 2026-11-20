-- Propuesta para revisión. Ejecutar separadamente con autorización de DDL.
-- Solo cambia metadatos de seis vistas; no modifica registros ni concede permisos.
-- Ejecutar el archivo completo en una sesión dedicada, preferentemente sqlcmd -b.
-- Fechas analíticas en hora de Lima (UTC-05).
-- Rechazar antes de cualquier SET y sin THROW: RAISERROR no aplica XACT_ABORT,
-- incluso si el llamador lo tiene ON. RETURN impide continuar sin TRY externo.
IF @@TRANCOUNT <> 0
BEGIN
    RAISERROR(N'Utilice una sesión sin transacciones abiertas.', 16, 1);
    RETURN;
END;

SET ANSI_NULLS ON;
SET QUOTED_IDENTIFIER ON;
SET NOCOUNT ON;
SET XACT_ABORT ON;
IF DB_NAME() <> N'elbravodeljac_sqlserver_dev'
    THROW 51000, N'Confirme la base de destino antes de adaptar este script.', 1;
IF SCHEMA_ID(N'bi') IS NULL
    THROW 51000, N'Falta el esquema bi. Revise el destino antes de continuar.', 1;
-- SELECT por vista es suficiente solo con una cadena de propiedad intacta.
-- No cambiar propietarios ni ampliar permisos automáticamente para resolverla.
IF EXISTS (
    SELECT 1 FROM sys.objects AS o
    JOIN sys.schemas AS s ON s.schema_id = o.schema_id
    WHERE ((s.name = N'dbo' AND o.name IN
              (N'pagos', N'ordenes_trabajo', N'cotizaciones', N'recepciones', N'detalle_cotizacion', N'servicios'))
        OR (s.name = N'bi' AND o.name IN
              (N'vw_pagos_vigentes', N'vw_ingresos_mensuales', N'vw_ordenes_analiticas',
               N'vw_cotizaciones_analiticas', N'vw_vehiculos_atendidos', N'vw_servicios_realizados')))
      AND COALESCE(o.principal_id, s.principal_id)
          <> (SELECT principal_id FROM sys.schemas WHERE name = N'bi')
)
    THROW 51000, N'La cadena de propiedad no permite lectura solo por vistas. Revise los propietarios.', 1;

-- SQL dinámico mantiene CREATE VIEW al inicio de su lote y permite capturar
-- errores de compilación de las vistas en un nivel inferior al TRY/CATCH.
BEGIN TRY
    BEGIN TRANSACTION;
    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_pagos_vigentes AS
SELECT p.id_pago, p.id_orden,
       CONVERT(date, SWITCHOFFSET(p.fecha_pago, ''-05:00'')) AS fecha,
       CAST(p.monto AS decimal(12,2)) AS monto
FROM dbo.pagos AS p
WHERE p.revertido_en IS NULL;';

    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_ingresos_mensuales AS
SELECT DATEFROMPARTS(YEAR(fecha), MONTH(fecha), 1) AS mes,
       COUNT_BIG(*) AS pagos,
       CAST(SUM(monto) AS decimal(38,2)) AS ingresos_cobrados
FROM bi.vw_pagos_vigentes
GROUP BY DATEFROMPARTS(YEAR(fecha), MONTH(fecha), 1);';

    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_ordenes_analiticas AS
SELECT o.id_orden, o.id_cotizacion,
       CONVERT(date, SWITCHOFFSET(o.fecha_creacion, ''-05:00'')) AS fecha,
       CASE o.estado
         WHEN N''creada'' THEN N''Pendiente''
         WHEN N''en_reparacion'' THEN N''En reparación''
         WHEN N''en_pruebas'' THEN N''En pruebas''
         WHEN N''finalizada'' THEN N''Finalizada''
         WHEN N''cancelada'' THEN N''Cancelada''
         ELSE o.estado END AS estado
FROM dbo.ordenes_trabajo AS o;';

    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_cotizaciones_analiticas AS
SELECT c.id_cotizacion, c.id_recepcion, c.id_cotizacion_base,
       CONVERT(date, SWITCHOFFSET(c.fecha_emision, ''-05:00'')) AS fecha,
       c.estado, CAST(c.total AS decimal(12,2)) AS total
FROM dbo.cotizaciones AS c;';

    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_vehiculos_atendidos AS
SELECT r.id_recepcion, r.id_vehiculo, r.id_cliente_atencion AS id_cliente,
       CONVERT(date, SWITCHOFFSET(r.fecha_ingreso, ''-05:00'')) AS fecha,
       r.tipo_atencion
FROM dbo.recepciones AS r;';

    EXEC sys.sp_executesql N'CREATE OR ALTER VIEW bi.vw_servicios_realizados AS
SELECT d.id_detalle, o.id_orden, d.id_servicio,
       CONVERT(date, SWITCHOFFSET(o.fecha_fin, ''-05:00'')) AS fecha,
       COALESCE(s.nombre, N''Servicio manual'') AS servicio,
       d.cantidad
FROM dbo.ordenes_trabajo AS o
JOIN dbo.detalle_cotizacion AS d ON d.id_cotizacion = o.id_cotizacion
LEFT JOIN dbo.servicios AS s ON s.id_servicio = d.id_servicio
WHERE o.estado = N''finalizada'' AND o.fecha_fin IS NOT NULL
  AND COALESCE(d.tipo_concepto,
      CASE WHEN d.id_servicio IS NOT NULL THEN N''Servicio'' ELSE N''Repuesto'' END) = N''Servicio''
  AND (JSON_VALUE(o.control_tecnico, CONCAT(N''$.execution."'', d.id_detalle, N''".status'')) = N''Realizado''
       OR EXISTS (SELECT 1 FROM OPENJSON(o.control_tecnico, N''$.performedItems'') AS j
                  WHERE j.value = CONVERT(varchar(20), d.id_detalle)));';
    COMMIT TRANSACTION;
END TRY
BEGIN CATCH
    IF XACT_STATE() <> 0 ROLLBACK TRANSACTION;
    THROW;
END CATCH;
