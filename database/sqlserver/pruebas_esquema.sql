-- Ejecutar una vez con sqlcmd -b en una base QA recién creada con esquema_entrega.sql.
SET NOCOUNT ON;
SET XACT_ABORT ON;
-- Las escrituras sobre índices filtrados requieren estas opciones de sesión.
SET ANSI_NULLS ON;
SET ANSI_PADDING ON;
SET ANSI_WARNINGS ON;
SET ARITHABORT ON;
SET CONCAT_NULL_YIELDS_NULL ON;
SET QUOTED_IDENTIFIER ON;
SET NUMERIC_ROUNDABORT OFF;

IF DB_NAME() <> N'elbravodeljac_sqlserver_qa'
    THROW 52000, N'Las pruebas requieren la base desechable elbravodeljac_sqlserver_qa.', 1;
IF EXISTS (SELECT 1 FROM dbo.usuarios) OR EXISTS (SELECT 1 FROM dbo.clientes)
    THROW 52000, N'Las pruebas requieren una instalación sin datos.', 1;

IF (SELECT COUNT(*) FROM sys.tables WHERE schema_id = SCHEMA_ID(N'dbo')) <> 28
    OR (SELECT COUNT(*) FROM sys.columns c JOIN sys.tables t ON t.object_id = c.object_id
        WHERE t.schema_id = SCHEMA_ID(N'dbo')) <> 208
    OR (SELECT COUNT(*) FROM sys.foreign_keys) <> 50
    OR (SELECT COUNT(*) FROM sys.key_constraints WHERE type = N'PK') <> 28
    OR (SELECT COUNT(*) FROM sys.sequences) <> 26
    OR (SELECT COUNT(*) FROM sys.views WHERE is_ms_shipped = 0) <> 3
    OR (SELECT COUNT(*) FROM sys.objects WHERE type = N'FN' AND is_ms_shipped = 0) <> 3
    OR (SELECT COUNT(*) FROM sys.triggers WHERE parent_class = 1 AND is_ms_shipped = 0) <> 10
    OR (SELECT COUNT(*) FROM sys.indexes i JOIN sys.tables t ON t.object_id = i.object_id
        WHERE t.schema_id = SCHEMA_ID(N'dbo') AND i.index_id > 0) <> 93
    OR EXISTS (SELECT 1 FROM sys.foreign_keys WHERE is_disabled = 1 OR is_not_trusted = 1)
    OR EXISTS (SELECT 1 FROM sys.check_constraints WHERE is_disabled = 1 OR is_not_trusted = 1)
    OR EXISTS (SELECT 1 FROM sys.triggers WHERE parent_class = 1 AND is_disabled = 1)
    THROW 52001, N'Inventario físico o estado de integridad inesperado.', 1;

INSERT dbo.tipos_documento (id_tipo_documento, codigo, nombre, activo)
VALUES (100, N'DNI', N'Documento nacional', 1), (101, N'CE', N'Carné de extranjería', 1),
       (102, N'RUC', N'Inactivo QA', 0);
INSERT dbo.empleados (id_empleado, nombre, tipo_empleado)
VALUES (100, N'Mecánico ágil', N'Mecánico'), (101, N'Recepción QA', N'Recepción');
INSERT dbo.usuarios (id_usuario, nombre, usuario, password_hash, id_empleado)
VALUES (100, N'Normalizar', N'mecanico', N'fixture', 100),
       (101, N'Normalizar', N'recepcion', N'fixture', 101),
       (102, N'Sin empleado 1', N'libre1', N'fixture', NULL),
       (103, N'Sin empleado 2', N'libre2', N'fixture', NULL);
IF NOT EXISTS (SELECT 1 FROM dbo.usuarios WHERE id_usuario = 100 AND nombre = N'Mecánico ágil')
    THROW 52002, N'Nombre laboral no sincronizado en alta múltiple.', 1;

UPDATE dbo.empleados SET nombre = CONCAT(nombre, N' 🚘') WHERE id_empleado IN (100,101);
IF (SELECT COUNT(*) FROM dbo.usuarios u JOIN dbo.empleados e ON e.id_empleado = u.id_empleado
    WHERE u.nombre = e.nombre AND RIGHT(u.nombre COLLATE Latin1_General_100_CI_AS_SC, 1) = N'🚘') <> 2
    THROW 52003, N'Nombre laboral no sincronizado en actualización múltiple.', 1;

INSERT dbo.clientes (id_cliente, tipo_documento, numero_documento, nombre, id_tipo_documento)
VALUES (100, N'Ignorar', N'100', N'Cliente Perú', 100), (101, N'Ignorar', N'101', N'Cliente QA', 101);
INSERT dbo.contactos_cliente (id_contacto, id_cliente, nombre, tipo_documento, numero_documento, id_tipo_documento)
VALUES (100, 100, N'Contacto uno', N'Ignorar', N'200', 100), (101, 101, N'Contacto dos', N'Ignorar', N'201', 101);
UPDATE dbo.contactos_cliente SET tipo_documento = N'Carné de extranjería' WHERE id_contacto = 100;
IF NOT EXISTS (SELECT 1 FROM dbo.clientes WHERE id_cliente = 100 AND tipo_documento = N'DNI')
    OR NOT EXISTS (SELECT 1 FROM dbo.contactos_cliente WHERE id_contacto = 100 AND tipo_documento = N'CE' AND id_tipo_documento = 101)
    THROW 52004, N'Catálogo de documentos no aplicado.', 1;
BEGIN TRY
    INSERT dbo.clientes (id_cliente, tipo_documento, numero_documento, nombre, id_tipo_documento)
    VALUES (102, N'DNI', N'102', N'Válido', 100), (103, N'RUC', N'103', N'Inactivo', 102);
    THROW 52005, N'Se admitió documento inactivo.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51002 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.clientes WHERE id_cliente IN (102,103))
    THROW 52006, N'El rechazo documental no fue atómico.', 1;

INSERT dbo.vehiculos (id_vehiculo, id_cliente, placa, marca)
VALUES (100, 100, N'QA-001', N'CITROEN'), (101, 101, N'QA-002', N'Marca QA');
IF (SELECT COUNT(*) FROM dbo.marcas_vehiculos) <> 35
    OR NOT EXISTS (SELECT 1 FROM dbo.marcas_vehiculos WHERE nombre = N'Citroën')
    THROW 52007, N'La vista de marcas no preserva la prioridad ni los acentos.', 1;
BEGIN TRY
    INSERT dbo.vehiculos (id_vehiculo, id_cliente, placa) VALUES (102, 9999, N'QA-003');
    THROW 52008, N'Se admitió un vehículo huérfano.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 547 THROW;
END CATCH;

INSERT dbo.recepciones (id_recepcion, id_vehiculo, id_cliente_atencion, id_recepcionista, tipo_atencion, id_mecanico_diagnostico)
VALUES (100, 100, 100, 101, N'Diagnostico', 100), (101, 101, 101, 101, N'Directo', 100),
       (102, 100, 100, 101, N'Diagnostico', 100), (103, 101, 101, 101, N'Diagnostico', 100);
BEGIN TRY
    INSERT dbo.diagnosticos (id_diagnostico, id_recepcion, id_tecnico, resumen)
    VALUES (100, 100, 100, N'Válido'), (101, 101, 100, N'Atención directa');
    THROW 52009, N'Se admitió diagnóstico para atención directa.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51006 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.diagnosticos) THROW 52010, N'El rechazo de diagnóstico no fue atómico.', 1;

INSERT dbo.diagnosticos (id_diagnostico, id_recepcion, id_tecnico, resumen)
VALUES (100, 100, 100, N'Evaluación ágil'), (102, 102, 100, N'Borrador QA');
UPDATE dbo.diagnosticos SET estado = N'Finalizado', finalizado_en = SYSDATETIMEOFFSET() WHERE id_diagnostico = 100;
BEGIN TRY
    UPDATE dbo.diagnosticos SET resumen = N'Reemplazo' WHERE id_diagnostico IN (100,102);
    THROW 52011, N'Se modificó un diagnóstico finalizado.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51007 THROW;
END CATCH;
IF NOT EXISTS (SELECT 1 FROM dbo.diagnosticos WHERE id_diagnostico = 102 AND resumen = N'Borrador QA')
    THROW 52012, N'El borrador cambió pese al rechazo del lote.', 1;
BEGIN TRY
    UPDATE dbo.recepciones SET tipo_atencion = N'Directo' WHERE id_recepcion IN (100,103);
    THROW 52013, N'Se cambió la atención con historial.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51005 THROW;
END CATCH;
IF NOT EXISTS (SELECT 1 FROM dbo.recepciones WHERE id_recepcion = 103 AND tipo_atencion = N'Diagnostico')
    THROW 52014, N'El cambio de atención no se revirtió en todas las filas.', 1;

BEGIN TRY
    INSERT dbo.cotizaciones (id_cotizacion, id_recepcion, id_diagnostico, estado, total)
    VALUES (100, 100, 100, N'aprobada', 10.10), (102, 102, 102, N'aprobada', 20.20);
    THROW 52015, N'Se cotizó un diagnóstico en borrador.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51004 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.cotizaciones) THROW 52016, N'El rechazo de cotización no fue atómico.', 1;
INSERT dbo.cotizaciones (id_cotizacion, id_recepcion, id_diagnostico, estado, total)
VALUES (100, 100, 100, N'aprobada', 10.10), (101, 101, NULL, N'pendiente', 20.20);
BEGIN TRY
    INSERT dbo.ordenes_trabajo (id_orden, id_cotizacion, codigo_ot, estado, id_usuario_creacion)
    VALUES (100, 100, N'QA-OT-100', N'finalizada', 101), (101, 101, N'QA-OT-101', N'finalizada', 101);
    THROW 52017, N'Se creó una orden con cotización pendiente.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51009 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.ordenes_trabajo) THROW 52018, N'El rechazo de orden no fue atómico.', 1;
UPDATE dbo.cotizaciones SET estado = N'aprobada' WHERE id_cotizacion = 101;
INSERT dbo.ordenes_trabajo (id_orden, id_cotizacion, codigo_ot, estado, id_usuario_creacion)
VALUES (100, 100, N'QA-OT-100', N'finalizada', 101), (101, 101, N'QA-OT-101', N'finalizada', 101);
BEGIN TRY
    INSERT dbo.asignaciones_ot (id_orden, id_usuario, rol_en_trabajo)
    VALUES (100, 100, N'MECANICO'), (101, 100, N'AUXILIAR');
    THROW 52019, N'Se creó una asignación auxiliar.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51008 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.asignaciones_ot) THROW 52020, N'El rechazo de asignación no fue atómico.', 1;

INSERT dbo.detalle_cotizacion (id_detalle, id_cotizacion, descripcion, cantidad, precio_unitario, subtotal, tipo_concepto)
VALUES (100, 101, N'Importe a cobrar', 1, 20.20, 20.20, N'Servicio'),
       (101, 101, N'Trabajo no realizado', 1, 15, 15, N'Servicio');
UPDATE dbo.ordenes_trabajo
SET flujo_version = 1, control_tecnico = N'{"reconciliation":{},"execution":{"101":{"status":"No realizado"}}}'
WHERE id_orden = 101;
IF dbo.jac_importe_orden(101) <> 20.20 THROW 52021, N'El importe conciliado no coincide.', 1;

INSERT dbo.pagos (id_orden, monto, metodo_pago, id_usuario_registro, request_id, revertido_en)
VALUES (100, 10.10, N'Efectivo', 101, N'pagado', NULL), (101, 10.20, N'Efectivo', 101, N'parcial', NULL),
       (101, 50.00, N'Efectivo', 101, N'revertido', SYSDATETIMEOFFSET());
BEGIN TRY
    INSERT dbo.pagos (id_orden, monto, metodo_pago, id_usuario_registro, request_id)
    VALUES (101, 1, N'Efectivo', 101, N'parcial');
    THROW 52022, N'Se admitió una clave de idempotencia duplicada.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() NOT IN (2601,2627) THROW;
END CATCH;
BEGIN TRY
    INSERT dbo.entregas (id_orden, id_usuario_entrega, tipo_receptor, autorizacion_verificada, medio_verificacion, id_usuario_verificador)
    VALUES (100, 101, N'Titular', 1, N'Documento', 101), (101, 101, N'Titular', 1, N'Documento', 101);
    THROW 52023, N'Se admitió una entrega con saldo pendiente.', 1;
END TRY
BEGIN CATCH
    IF ERROR_NUMBER() <> 51012 THROW;
END CATCH;
IF EXISTS (SELECT 1 FROM dbo.entregas) THROW 52024, N'El rechazo de entrega no fue atómico.', 1;
INSERT dbo.pagos (id_orden, monto, metodo_pago, id_usuario_registro) VALUES (101, 10.00, N'Efectivo', 101);
INSERT dbo.entregas (id_orden, id_usuario_entrega, tipo_receptor, autorizacion_verificada, medio_verificacion,
                     id_usuario_verificador, ficha_snapshot, ficha_pdf)
VALUES (100, 101, N'Titular', 1, N'Documento', 101, N'{}', 0x25504446),
       (101, 101, N'Titular', 1, N'Documento', 101, NULL, NULL);
IF (SELECT SUM(monto) FROM dbo.jac_movimientos_pagos WHERE id_orden = 101) <> 20.20
    OR NOT EXISTS (SELECT 1 FROM dbo.entregas WHERE id_orden = 100 AND ficha_pdf = 0x25504446)
    THROW 52025, N'El saldo neto o los bytes archivados no coinciden.', 1;
PRINT N'PASS: inventario, integridad, lotes, documentos, diagnósticos, órdenes, pagos y entregas.';
