-- Solo lectura. Ejecutar después de la instalación autorizada de vistas_analiticas.sql.
SET NOCOUNT ON;
IF (SELECT COUNT(*) FROM bi.vw_pagos_vigentes)
    <> (SELECT COUNT(*) FROM dbo.pagos WHERE revertido_en IS NULL)
    THROW 51000, N'El inventario de pagos vigentes no coincide.', 1;
IF EXISTS (SELECT id_pago FROM bi.vw_pagos_vigentes GROUP BY id_pago HAVING COUNT(*) <> 1)
    THROW 51000, N'Pago duplicado en la fuente analítica.', 1;
IF COALESCE((SELECT SUM(monto) FROM bi.vw_pagos_vigentes),0)
    <> COALESCE((SELECT SUM(monto) FROM dbo.pagos WHERE revertido_en IS NULL),0)
    THROW 51000, N'La suma de cobros no coincide.', 1;
IF COALESCE((SELECT SUM(ingresos_cobrados) FROM bi.vw_ingresos_mensuales),0)
    <> COALESCE((SELECT SUM(monto) FROM bi.vw_pagos_vigentes),0)
    THROW 51000, N'La agregación mensual duplica u omite importes.', 1;
IF EXISTS (SELECT 1 FROM bi.vw_pagos_vigentes v JOIN dbo.pagos p ON p.id_pago=v.id_pago
           WHERE p.revertido_en IS NOT NULL)
    THROW 51000, N'Hay pagos revertidos en la fuente analítica.', 1;
IF (SELECT COUNT(*) FROM bi.vw_ordenes_analiticas) <> (SELECT COUNT(*) FROM dbo.ordenes_trabajo)
    THROW 51000, N'Las órdenes no coinciden.', 1;
IF (SELECT COUNT(*) FROM bi.vw_cotizaciones_analiticas) <> (SELECT COUNT(*) FROM dbo.cotizaciones)
    THROW 51000, N'Las cotizaciones no coinciden.', 1;
IF (SELECT COUNT(*) FROM bi.vw_vehiculos_atendidos) <> (SELECT COUNT(*) FROM dbo.recepciones)
    THROW 51000, N'Las recepciones no coinciden.', 1;
IF EXISTS (SELECT id_detalle FROM bi.vw_servicios_realizados GROUP BY id_detalle HAVING COUNT(*) <> 1)
    THROW 51000, N'Concepto de servicio duplicado.', 1;
SELECT N'PASS' AS conciliacion;
