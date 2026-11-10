-- El Bravo del JAC. Esquema físico SQL Server 2025.
-- Ejecutar en una base vacía previamente creada, con sqlcmd -b.
-- No crea bases, cuentas, catálogos iniciales ni datos de demostración.
SET ANSI_NULLS ON;
SET ANSI_PADDING ON;
SET ANSI_WARNINGS ON;
SET ARITHABORT ON;
SET CONCAT_NULL_YIELDS_NULL ON;
SET QUOTED_IDENTIFIER ON;
SET NUMERIC_ROUNDABORT OFF;
SET XACT_ABORT ON;

IF DB_NAME() IN (N'master', N'model', N'msdb', N'tempdb')
    OR EXISTS (SELECT 1 FROM sys.tables WHERE is_ms_shipped = 0)
    THROW 51000, N'Seleccione una base de usuario vacía.', 1;

BEGIN TRANSACTION;
GO

-- Secuencias independientes para permitir cargar identificadores históricos.
CREATE SEQUENCE dbo.asignaciones_ot_id_asignacion_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.auxilios_id_auxilio_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.avances_ot_id_avance_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.clientes_id_cliente_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.contactos_cliente_id_contacto_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.cotizaciones_id_cotizacion_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.detalle_cotizacion_id_detalle_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.diagnosticos_id_diagnostico_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.empleados_id_empleado_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.entregas_id_entrega_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.historial_estados_ot_id_historial_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.historial_propietarios_id_historial_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.ordenes_trabajo_id_orden_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.pagos_id_pago_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.permisos_id_permiso_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.proveedores_id_proveedor_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.pruebas_ot_id_prueba_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.recepciones_id_recepcion_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.repuesto_proveedor_id_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.repuestos_id_repuesto_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.roles_id_rol_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.servicios_id_servicio_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.tipos_documento_id_tipo_documento_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.unidades_auxilio_id_unidad_auxilio_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.usuarios_id_usuario_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
CREATE SEQUENCE dbo.vehiculos_id_vehiculo_seq AS bigint START WITH 1 INCREMENT BY 1 NO CACHE;
GO

-- Tablas y validaciones locales.
CREATE TABLE dbo.asignaciones_ot (
    id_asignacion bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.asignaciones_ot_id_asignacion_seq),
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    rol_en_trabajo nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL,
    fecha_asignacion datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_asignaciones_ot_rol_en_trabajo_longitud CHECK (LEN((rol_en_trabajo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30)
);

CREATE TABLE dbo.auxilios (
    id_auxilio bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.auxilios_id_auxilio_seq),
    id_vehiculo bigint NULL,
    id_cliente_solicitante bigint NULL,
    id_usuario_registro bigint NOT NULL,
    id_tecnico bigint NULL,
    id_unidad_auxilio bigint NULL,
    id_recepcion bigint NULL,
    ubicacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL,
    tipo_incidente nvarchar(120) COLLATE Latin1_General_100_BIN2 NULL,
    descripcion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    fecha datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    estado nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'pendiente'),
    CONSTRAINT ck_auxilios_tipo_incidente_longitud CHECK (LEN((tipo_incidente COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 60),
    CONSTRAINT ck_auxilios_estado_longitud CHECK (LEN((estado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30)
);

CREATE TABLE dbo.avances_ot (
    id_avance bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.avances_ot_id_avance_seq),
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    observacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL,
    fecha datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00'))
);

CREATE TABLE dbo.clientes (
    id_cliente bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.clientes_id_cliente_seq),
    tipo_documento nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    numero_documento nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    telefono nvarchar(40) COLLATE Latin1_General_100_BIN2 NULL,
    direccion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    correo nvarchar(300) COLLATE Latin1_General_100_BIN2 NULL,
    id_tipo_documento bigint NOT NULL,
    CONSTRAINT ck_clientes_tipo_documento_longitud CHECK (LEN((tipo_documento COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_clientes_numero_documento_longitud CHECK (LEN((numero_documento COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_clientes_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_clientes_telefono_longitud CHECK (LEN((telefono COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_clientes_correo_longitud CHECK (LEN((correo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150)
);

CREATE TABLE dbo.contactos_cliente (
    id_contacto bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.contactos_cliente_id_contacto_seq),
    id_cliente bigint NOT NULL,
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    tipo_documento nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    numero_documento nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    telefono nvarchar(40) COLLATE Latin1_General_100_BIN2 NULL,
    relacion nvarchar(100) COLLATE Latin1_General_100_BIN2 NULL,
    activo bit NOT NULL DEFAULT (1),
    id_tipo_documento bigint NOT NULL,
    CONSTRAINT ck_contactos_cliente_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_contactos_cliente_tipo_documento_longitud CHECK (LEN((tipo_documento COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_contactos_cliente_numero_documento_longitud CHECK (LEN((numero_documento COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_contactos_cliente_telefono_longitud CHECK (LEN((telefono COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_contactos_cliente_relacion_longitud CHECK (LEN((relacion COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 50)
);

CREATE TABLE dbo.cotizaciones (
    id_cotizacion bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.cotizaciones_id_cotizacion_seq),
    id_diagnostico bigint NULL,
    estado nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'pendiente'),
    total decimal(12,2) NOT NULL DEFAULT (0),
    fecha_emision datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    aprobada_en datetimeoffset(6) NULL,
    id_recepcion bigint NOT NULL,
    version bigint NOT NULL DEFAULT (0),
    contenido_hash nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    aprobacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    historial_revision nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'[]'),
    id_cotizacion_base bigint NULL,
    CONSTRAINT ck_cotizaciones_total CHECK (((total >= 0))),
    CONSTRAINT ck_cotizaciones_estado_longitud CHECK (LEN((estado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_cotizaciones_aprobacion_json CHECK (ISJSON(aprobacion, VALUE) = 1),
    CONSTRAINT ck_cotizaciones_historial_revision_json CHECK (ISJSON(historial_revision, VALUE) = 1)
);

CREATE TABLE dbo.detalle_cotizacion (
    id_detalle bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.detalle_cotizacion_id_detalle_seq),
    id_cotizacion bigint NOT NULL,
    id_servicio bigint NULL,
    id_repuesto bigint NULL,
    descripcion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    cantidad decimal(10,2) NOT NULL,
    precio_unitario decimal(12,2) NOT NULL,
    subtotal decimal(12,2) NOT NULL,
    origen_repuesto nvarchar(24) COLLATE Latin1_General_100_BIN2 NULL,
    tipo_concepto nvarchar(20) COLLATE Latin1_General_100_BIN2 NULL,
    tipo_repuesto nvarchar(24) COLLATE Latin1_General_100_BIN2 NULL,
    descuento_unitario decimal(12,2) NOT NULL DEFAULT (0),
    CONSTRAINT ck_detalle_cantidad CHECK (((cantidad > 0))),
    CONSTRAINT ck_detalle_descuento CHECK ((((descuento_unitario >= 0) AND (descuento_unitario <= precio_unitario)))),
    CONSTRAINT ck_detalle_origen CHECK ((DATALENGTH(origen_repuesto) = DATALENGTH(RTRIM(origen_repuesto)) AND origen_repuesto IN (N'TALLER', N'CLIENTE'))),
    CONSTRAINT ck_detalle_precio CHECK (((precio_unitario >= 0))),
    CONSTRAINT ck_detalle_referencia CHECK ((id_servicio IS NOT NULL AND id_repuesto IS NULL AND (tipo_concepto IS NULL OR (tipo_concepto = N'Servicio' AND DATALENGTH(tipo_concepto) = DATALENGTH(N'Servicio'))))
        OR (id_repuesto IS NOT NULL AND id_servicio IS NULL AND (tipo_concepto IS NULL OR (tipo_concepto = N'Repuesto' AND DATALENGTH(tipo_concepto) = DATALENGTH(N'Repuesto'))))
        OR (id_servicio IS NULL AND id_repuesto IS NULL AND tipo_concepto IS NOT NULL
            AND (DATALENGTH(tipo_concepto) = DATALENGTH(RTRIM(tipo_concepto)) AND tipo_concepto IN (N'Servicio', N'Repuesto')) AND descripcion IS NOT NULL AND TRIM(descripcion) <> N'')),
    CONSTRAINT ck_detalle_subtotal CHECK (((subtotal >= 0))),
    CONSTRAINT ck_detalle_tipo_repuesto CHECK (tipo_repuesto IS NULL OR ((DATALENGTH(tipo_repuesto) = DATALENGTH(RTRIM(tipo_repuesto)) AND tipo_repuesto IN (N'Original', N'Alternativo'))
        AND COALESCE(tipo_concepto, CASE WHEN id_servicio IS NOT NULL THEN N'Servicio' ELSE N'Repuesto' END) = N'Repuesto'
        AND (tipo_concepto IS NULL OR DATALENGTH(tipo_concepto) = DATALENGTH(N'Repuesto')))),
    CONSTRAINT ck_detalle_cotizacion_origen_repuesto_longitud CHECK (LEN((origen_repuesto COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 12),
    CONSTRAINT ck_detalle_cotizacion_tipo_concepto_longitud CHECK (LEN((tipo_concepto COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 10),
    CONSTRAINT ck_detalle_cotizacion_tipo_repuesto_longitud CHECK (LEN((tipo_repuesto COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 12)
);

CREATE TABLE dbo.diagnosticos (
    id_diagnostico bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.diagnosticos_id_diagnostico_seq),
    id_recepcion bigint NOT NULL,
    id_tecnico bigint NOT NULL,
    fecha datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    resumen nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL,
    estado nvarchar(30) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'Borrador'),
    evaluacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N''),
    recomendaciones nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N''),
    observaciones nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N''),
    version bigint NOT NULL DEFAULT (0),
    finalizado_en datetimeoffset(6) NULL,
    actualizado_en datetimeoffset(6) NULL,
    historial nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'[]'),
    CONSTRAINT diagnosticos_estado_check CHECK ((DATALENGTH(estado) = DATALENGTH(RTRIM(estado)) AND estado IN (N'Borrador', N'Finalizado'))),
    CONSTRAINT ck_diagnosticos_estado_longitud CHECK (LEN((estado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 15),
    CONSTRAINT ck_diagnosticos_historial_json CHECK (ISJSON(historial, VALUE) = 1)
);

CREATE TABLE dbo.empleados (
    id_empleado bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.empleados_id_empleado_seq),
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    tipo_empleado nvarchar(80) COLLATE Latin1_General_100_BIN2 NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT empleados_nombre_check CHECK (TRIM(nombre) <> N''),
    CONSTRAINT empleados_tipo_empleado_check CHECK (TRIM(tipo_empleado) <> N''),
    CONSTRAINT ck_empleados_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_empleados_tipo_empleado_longitud CHECK (LEN((tipo_empleado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 40)
);

CREATE TABLE dbo.entregas (
    id_entrega bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.entregas_id_entrega_seq),
    id_orden bigint NOT NULL,
    id_usuario_entrega bigint NOT NULL,
    id_contacto_receptor bigint NULL,
    fecha_entrega datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    tipo_receptor nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL,
    nombre_receptor_snap nvarchar(300) COLLATE Latin1_General_100_BIN2 NULL,
    tipo_doc_receptor_snap nvarchar(40) COLLATE Latin1_General_100_BIN2 NULL,
    documento_receptor_snap nvarchar(60) COLLATE Latin1_General_100_BIN2 NULL,
    telefono_receptor_snap nvarchar(40) COLLATE Latin1_General_100_BIN2 NULL,
    autorizacion_verificada bit NOT NULL DEFAULT (0),
    medio_verificacion nvarchar(160) COLLATE Latin1_General_100_BIN2 NULL,
    fecha_verificacion datetimeoffset(6) NULL,
    id_usuario_verificador bigint NOT NULL,
    ficha_snapshot nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    ficha_pdf varbinary(max) NULL,
    receptor_real nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    conformidad nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT ck_entregas_tipo_receptor_longitud CHECK (LEN((tipo_receptor COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_entregas_nombre_receptor_snap_longitud CHECK (LEN((nombre_receptor_snap COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_entregas_tipo_doc_receptor_snap_longitud CHECK (LEN((tipo_doc_receptor_snap COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_entregas_documento_receptor_snap_longitud CHECK (LEN((documento_receptor_snap COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_entregas_telefono_receptor_snap_longitud CHECK (LEN((telefono_receptor_snap COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_entregas_medio_verificacion_longitud CHECK (LEN((medio_verificacion COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 80),
    CONSTRAINT ck_entregas_ficha_snapshot_json CHECK (ISJSON(ficha_snapshot, VALUE) = 1),
    CONSTRAINT ck_entregas_receptor_real_json CHECK (ISJSON(receptor_real, VALUE) = 1),
    CONSTRAINT ck_entregas_conformidad_json CHECK (ISJSON(conformidad, VALUE) = 1)
);

CREATE TABLE dbo.historial_estados_ot (
    id_historial bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.historial_estados_ot_id_historial_seq),
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    estado_anterior nvarchar(60) COLLATE Latin1_General_100_BIN2 NULL,
    estado_nuevo nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL,
    fecha datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    CONSTRAINT ck_historial_estados_ot_estado_anterior_longitud CHECK (LEN((estado_anterior COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_historial_estados_ot_estado_nuevo_longitud CHECK (LEN((estado_nuevo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30)
);

CREATE TABLE dbo.historial_propietarios (
    id_historial bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.historial_propietarios_id_historial_seq),
    id_vehiculo bigint NOT NULL,
    id_cliente bigint NOT NULL,
    fecha_inicio datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    fecha_fin datetimeoffset(6) NULL,
    id_usuario_registro bigint NOT NULL,
    verificado bit NOT NULL DEFAULT (0),
    observacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT chk_fechas_propietario CHECK ((((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio)))),
    CONSTRAINT ck_historial_fechas CHECK ((((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio))))
);

CREATE TABLE dbo.pagos (
    id_pago bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.pagos_id_pago_seq),
    id_orden bigint NOT NULL,
    monto decimal(12,2) NOT NULL,
    metodo_pago nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL,
    fecha_pago datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    comprobante nvarchar(100) COLLATE Latin1_General_100_BIN2 NULL,
    id_usuario_registro bigint NOT NULL,
    request_id nvarchar(200) COLLATE Latin1_General_100_BIN2 NULL,
    revertido_en datetimeoffset(6) NULL,
    rectificacion nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT ck_pagos_monto CHECK (((monto > 0))),
    CONSTRAINT ck_pagos_metodo_pago_longitud CHECK (LEN((metodo_pago COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_pagos_comprobante_longitud CHECK (LEN((comprobante COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 50),
    CONSTRAINT ck_pagos_request_id_longitud CHECK (LEN((request_id COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 100),
    CONSTRAINT ck_pagos_rectificacion_json CHECK (ISJSON(rectificacion, VALUE) = 1)
);

CREATE TABLE dbo.vehiculos (
    id_vehiculo bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.vehiculos_id_vehiculo_seq),
    id_cliente bigint NOT NULL,
    placa nvarchar(24) COLLATE Latin1_General_100_BIN2 NOT NULL,
    marca nvarchar(80) COLLATE Latin1_General_100_BIN2 NULL,
    modelo nvarchar(120) COLLATE Latin1_General_100_BIN2 NULL,
    anio smallint NULL,
    kilometraje_actual decimal(10,2) NULL,
    observaciones nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT ck_vehiculos_anio CHECK ((((anio IS NULL) OR ((anio >= 1900) AND (anio <= 2100))))),
    CONSTRAINT ck_vehiculos_km CHECK ((((kilometraje_actual IS NULL) OR (kilometraje_actual >= 0)))),
    CONSTRAINT ck_vehiculos_placa_longitud CHECK (LEN((placa COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 12),
    CONSTRAINT ck_vehiculos_marca_longitud CHECK (LEN((marca COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 40),
    CONSTRAINT ck_vehiculos_modelo_longitud CHECK (LEN((modelo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 60)
);

CREATE TABLE dbo.ordenes_trabajo (
    id_orden bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.ordenes_trabajo_id_orden_seq),
    id_cotizacion bigint NOT NULL,
    codigo_ot nvarchar(48) COLLATE Latin1_General_100_BIN2 NOT NULL,
    estado nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'creada'),
    id_usuario_creacion bigint NOT NULL,
    fecha_creacion datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    fecha_inicio datetimeoffset(6) NULL,
    fecha_fin datetimeoffset(6) NULL,
    control_tecnico nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'{}'),
    control_version bigint NOT NULL DEFAULT (0),
    control_autor bigint NULL,
    control_fecha datetimeoffset(6) NULL,
    flujo_version int NOT NULL DEFAULT (0),
    control_historial nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'[]'),
    CONSTRAINT ck_ordenes_trabajo_codigo_ot_longitud CHECK (LEN((codigo_ot COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 24),
    CONSTRAINT ck_ordenes_trabajo_estado_longitud CHECK (LEN((estado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_ordenes_trabajo_control_tecnico_json CHECK (ISJSON(control_tecnico, VALUE) = 1),
    CONSTRAINT ck_ordenes_trabajo_control_historial_json CHECK (ISJSON(control_historial, VALUE) = 1)
);

CREATE TABLE dbo.permisos (
    id_permiso bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.permisos_id_permiso_seq),
    codigo nvarchar(160) COLLATE Latin1_General_100_BIN2 NOT NULL,
    descripcion nvarchar(300) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT ck_permisos_codigo_longitud CHECK (LEN((codigo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 80),
    CONSTRAINT ck_permisos_descripcion_longitud CHECK (LEN((descripcion COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150)
);

CREATE TABLE dbo.proveedores (
    id_proveedor bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.proveedores_id_proveedor_seq),
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    documento nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    telefono nvarchar(40) COLLATE Latin1_General_100_BIN2 NULL,
    contacto nvarchar(240) COLLATE Latin1_General_100_BIN2 NULL,
    CONSTRAINT ck_proveedores_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_proveedores_documento_longitud CHECK (LEN((documento COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_proveedores_telefono_longitud CHECK (LEN((telefono COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_proveedores_contacto_longitud CHECK (LEN((contacto COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 120)
);

CREATE TABLE dbo.pruebas_ot (
    id_prueba bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.pruebas_ot_id_prueba_seq),
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    tipo_prueba nvarchar(200) COLLATE Latin1_General_100_BIN2 NOT NULL,
    resultado nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    observaciones nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    fecha datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    CONSTRAINT ck_pruebas_ot_tipo_prueba_longitud CHECK (LEN((tipo_prueba COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 100),
    CONSTRAINT ck_pruebas_ot_resultado_longitud CHECK (LEN((resultado COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20)
);

CREATE TABLE dbo.recepciones (
    id_recepcion bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.recepciones_id_recepcion_seq),
    id_vehiculo bigint NOT NULL,
    id_cliente_atencion bigint NOT NULL,
    id_recepcionista bigint NOT NULL,
    fecha_ingreso datetimeoffset(6) NOT NULL DEFAULT (TODATETIMEOFFSET(SYSUTCDATETIME(), '+00:00')),
    km_ingreso decimal(10,2) NULL,
    sintomas nvarchar(max) COLLATE Latin1_General_100_BIN2 NULL,
    tipo_atencion nvarchar(30) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'Diagnostico'),
    id_mecanico_diagnostico bigint NULL,
    diagnostico_asignado_por bigint NULL,
    diagnostico_asignado_en datetimeoffset(6) NULL,
    asignacion_version bigint NOT NULL DEFAULT (0),
    historial_asignaciones nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL DEFAULT (N'[]'),
    CONSTRAINT ck_recepciones_km CHECK ((((km_ingreso IS NULL) OR (km_ingreso >= 0)))),
    CONSTRAINT recepciones_tipo_atencion_check CHECK ((DATALENGTH(tipo_atencion) = DATALENGTH(RTRIM(tipo_atencion)) AND tipo_atencion IN (N'Directo', N'Diagnostico'))),
    CONSTRAINT ck_recepciones_tipo_atencion_longitud CHECK (LEN((tipo_atencion COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 15),
    CONSTRAINT ck_recepciones_historial_asignaciones_json CHECK (ISJSON(historial_asignaciones, VALUE) = 1)
);

CREATE TABLE dbo.repuesto_proveedor (
    id bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.repuesto_proveedor_id_seq),
    id_repuesto bigint NOT NULL,
    id_proveedor bigint NOT NULL,
    costo_referencia decimal(12,2) NOT NULL,
    CONSTRAINT ck_rp_costo CHECK (((costo_referencia >= 0)))
);

CREATE TABLE dbo.repuestos (
    id_repuesto bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.repuestos_id_repuesto_seq),
    nombre nvarchar(240) COLLATE Latin1_General_100_BIN2 NOT NULL,
    stock decimal(12,2) NOT NULL DEFAULT (0),
    precio_referencia decimal(12,2) NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_repuestos_precio CHECK (((precio_referencia >= 0))),
    CONSTRAINT ck_repuestos_stock CHECK (((stock >= 0))),
    CONSTRAINT ck_repuestos_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 120)
);

CREATE TABLE dbo.rol_permisos (
    id_rol bigint NOT NULL,
    id_permiso bigint NOT NULL
);

CREATE TABLE dbo.roles (
    id_rol bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.roles_id_rol_seq),
    nombre_rol nvarchar(120) COLLATE Latin1_General_100_BIN2 NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_roles_nombre_rol_longitud CHECK (LEN((nombre_rol COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 60)
);

CREATE TABLE dbo.servicios (
    id_servicio bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.servicios_id_servicio_seq),
    nombre nvarchar(200) COLLATE Latin1_General_100_BIN2 NOT NULL,
    precio_referencia decimal(12,2) NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_servicios_precio CHECK (((precio_referencia >= 0))),
    CONSTRAINT ck_servicios_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 100)
);

CREATE TABLE dbo.tipos_documento (
    id_tipo_documento bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.tipos_documento_id_tipo_documento_seq),
    codigo nvarchar(40) COLLATE Latin1_General_100_BIN2 NOT NULL,
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_tipos_documento_codigo_longitud CHECK (LEN((codigo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 20),
    CONSTRAINT ck_tipos_documento_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150)
);

CREATE TABLE dbo.unidades_auxilio (
    id_unidad_auxilio bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.unidades_auxilio_id_unidad_auxilio_seq),
    codigo nvarchar(60) COLLATE Latin1_General_100_BIN2 NOT NULL,
    descripcion nvarchar(300) COLLATE Latin1_General_100_BIN2 NULL,
    activo bit NOT NULL DEFAULT (1),
    CONSTRAINT ck_unidades_auxilio_codigo_longitud CHECK (LEN((codigo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 30),
    CONSTRAINT ck_unidades_auxilio_descripcion_longitud CHECK (LEN((descripcion COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150)
);

CREATE TABLE dbo.usuario_roles (
    id_usuario bigint NOT NULL,
    id_rol bigint NOT NULL
);

CREATE TABLE dbo.usuarios (
    id_usuario bigint NOT NULL DEFAULT (NEXT VALUE FOR dbo.usuarios_id_usuario_seq),
    nombre nvarchar(300) COLLATE Latin1_General_100_BIN2 NOT NULL,
    usuario nvarchar(120) COLLATE Latin1_General_100_BIN2 NOT NULL,
    correo nvarchar(300) COLLATE Latin1_General_100_BIN2 NULL,
    password_hash nvarchar(max) COLLATE Latin1_General_100_BIN2 NOT NULL,
    activo bit NOT NULL DEFAULT (1),
    id_empleado bigint NULL,
    CONSTRAINT ck_usuarios_nombre_longitud CHECK (LEN((nombre COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150),
    CONSTRAINT ck_usuarios_usuario_longitud CHECK (LEN((usuario COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 60),
    CONSTRAINT ck_usuarios_correo_longitud CHECK (LEN((correo COLLATE Latin1_General_100_CI_AS_SC) + N'#') - 1 <= 150)
);

GO

-- Claves primarias, unicidad y referencias.
ALTER TABLE dbo.asignaciones_ot ADD CONSTRAINT asignaciones_ot_pkey PRIMARY KEY (id_asignacion);
ALTER TABLE dbo.auxilios ADD CONSTRAINT auxilios_pkey PRIMARY KEY (id_auxilio);
ALTER TABLE dbo.avances_ot ADD CONSTRAINT avances_ot_pkey PRIMARY KEY (id_avance);
ALTER TABLE dbo.clientes ADD CONSTRAINT clientes_pkey PRIMARY KEY (id_cliente);
ALTER TABLE dbo.contactos_cliente ADD CONSTRAINT contactos_cliente_pkey PRIMARY KEY (id_contacto);
ALTER TABLE dbo.cotizaciones ADD CONSTRAINT cotizaciones_pkey PRIMARY KEY (id_cotizacion);
ALTER TABLE dbo.detalle_cotizacion ADD CONSTRAINT detalle_cotizacion_pkey PRIMARY KEY (id_detalle);
ALTER TABLE dbo.diagnosticos ADD CONSTRAINT diagnosticos_pkey PRIMARY KEY (id_diagnostico);
ALTER TABLE dbo.empleados ADD CONSTRAINT empleados_pkey PRIMARY KEY (id_empleado);
ALTER TABLE dbo.entregas ADD CONSTRAINT entregas_pkey PRIMARY KEY (id_entrega);
ALTER TABLE dbo.historial_estados_ot ADD CONSTRAINT historial_estados_ot_pkey PRIMARY KEY (id_historial);
ALTER TABLE dbo.historial_propietarios ADD CONSTRAINT historial_propietarios_pkey PRIMARY KEY (id_historial);
ALTER TABLE dbo.ordenes_trabajo ADD CONSTRAINT ordenes_trabajo_pkey PRIMARY KEY (id_orden);
ALTER TABLE dbo.pagos ADD CONSTRAINT pagos_pkey PRIMARY KEY (id_pago);
ALTER TABLE dbo.permisos ADD CONSTRAINT permisos_codigo_key UNIQUE (codigo);
ALTER TABLE dbo.permisos ADD CONSTRAINT permisos_pkey PRIMARY KEY (id_permiso);
ALTER TABLE dbo.rol_permisos ADD CONSTRAINT pk_rol_permisos PRIMARY KEY (id_rol, id_permiso);
ALTER TABLE dbo.usuario_roles ADD CONSTRAINT pk_usuario_roles PRIMARY KEY (id_usuario, id_rol);
ALTER TABLE dbo.proveedores ADD CONSTRAINT proveedores_pkey PRIMARY KEY (id_proveedor);
ALTER TABLE dbo.pruebas_ot ADD CONSTRAINT pruebas_ot_pkey PRIMARY KEY (id_prueba);
ALTER TABLE dbo.recepciones ADD CONSTRAINT recepciones_pkey PRIMARY KEY (id_recepcion);
ALTER TABLE dbo.repuesto_proveedor ADD CONSTRAINT repuesto_proveedor_pkey PRIMARY KEY (id);
ALTER TABLE dbo.repuestos ADD CONSTRAINT repuestos_pkey PRIMARY KEY (id_repuesto);
ALTER TABLE dbo.roles ADD CONSTRAINT roles_nombre_rol_key UNIQUE (nombre_rol);
ALTER TABLE dbo.roles ADD CONSTRAINT roles_pkey PRIMARY KEY (id_rol);
ALTER TABLE dbo.servicios ADD CONSTRAINT servicios_pkey PRIMARY KEY (id_servicio);
ALTER TABLE dbo.tipos_documento ADD CONSTRAINT tipos_documento_codigo_key UNIQUE (codigo);
ALTER TABLE dbo.tipos_documento ADD CONSTRAINT tipos_documento_pkey PRIMARY KEY (id_tipo_documento);
ALTER TABLE dbo.unidades_auxilio ADD CONSTRAINT unidades_auxilio_pkey PRIMARY KEY (id_unidad_auxilio);
ALTER TABLE dbo.clientes ADD CONSTRAINT uq_cliente_documento UNIQUE (tipo_documento, numero_documento);
ALTER TABLE dbo.usuarios ADD CONSTRAINT usuarios_pkey PRIMARY KEY (id_usuario);
ALTER TABLE dbo.usuarios ADD CONSTRAINT usuarios_usuario_key UNIQUE (usuario);
ALTER TABLE dbo.vehiculos ADD CONSTRAINT vehiculos_pkey PRIMARY KEY (id_vehiculo);
ALTER TABLE dbo.vehiculos ADD CONSTRAINT vehiculos_placa_key UNIQUE (placa);
ALTER TABLE dbo.clientes ADD CONSTRAINT clientes_id_tipo_documento_fkey FOREIGN KEY (id_tipo_documento) REFERENCES dbo.tipos_documento(id_tipo_documento) ON DELETE NO ACTION;
ALTER TABLE dbo.contactos_cliente ADD CONSTRAINT contactos_cliente_id_tipo_documento_fkey FOREIGN KEY (id_tipo_documento) REFERENCES dbo.tipos_documento(id_tipo_documento) ON DELETE NO ACTION;
ALTER TABLE dbo.cotizaciones ADD CONSTRAINT cotizaciones_id_cotizacion_base_fkey FOREIGN KEY (id_cotizacion_base) REFERENCES dbo.cotizaciones(id_cotizacion);
ALTER TABLE dbo.cotizaciones ADD CONSTRAINT cotizaciones_id_recepcion_fkey FOREIGN KEY (id_recepcion) REFERENCES dbo.recepciones(id_recepcion);
ALTER TABLE dbo.asignaciones_ot ADD CONSTRAINT fk_asignaciones_ot_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.asignaciones_ot ADD CONSTRAINT fk_asignaciones_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_cliente_solicitante FOREIGN KEY (id_cliente_solicitante) REFERENCES dbo.clientes(id_cliente);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES dbo.recepciones(id_recepcion);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_unidad_auxilio FOREIGN KEY (id_unidad_auxilio) REFERENCES dbo.unidades_auxilio(id_unidad_auxilio);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.auxilios ADD CONSTRAINT fk_auxilios_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES dbo.vehiculos(id_vehiculo);
ALTER TABLE dbo.avances_ot ADD CONSTRAINT fk_avances_ot_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.avances_ot ADD CONSTRAINT fk_avances_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.contactos_cliente ADD CONSTRAINT fk_contactos_cliente_id_cliente FOREIGN KEY (id_cliente) REFERENCES dbo.clientes(id_cliente);
ALTER TABLE dbo.cotizaciones ADD CONSTRAINT fk_cotizaciones_id_diagnostico FOREIGN KEY (id_diagnostico) REFERENCES dbo.diagnosticos(id_diagnostico);
ALTER TABLE dbo.detalle_cotizacion ADD CONSTRAINT fk_detalle_cotizacion_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES dbo.cotizaciones(id_cotizacion);
ALTER TABLE dbo.detalle_cotizacion ADD CONSTRAINT fk_detalle_cotizacion_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES dbo.repuestos(id_repuesto);
ALTER TABLE dbo.detalle_cotizacion ADD CONSTRAINT fk_detalle_cotizacion_id_servicio FOREIGN KEY (id_servicio) REFERENCES dbo.servicios(id_servicio);
ALTER TABLE dbo.diagnosticos ADD CONSTRAINT fk_diagnosticos_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES dbo.recepciones(id_recepcion);
ALTER TABLE dbo.diagnosticos ADD CONSTRAINT fk_diagnosticos_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.entregas ADD CONSTRAINT fk_entregas_id_contacto_receptor FOREIGN KEY (id_contacto_receptor) REFERENCES dbo.contactos_cliente(id_contacto);
ALTER TABLE dbo.entregas ADD CONSTRAINT fk_entregas_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.entregas ADD CONSTRAINT fk_entregas_id_usuario_entrega FOREIGN KEY (id_usuario_entrega) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.entregas ADD CONSTRAINT fk_entregas_id_usuario_verificador FOREIGN KEY (id_usuario_verificador) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.historial_propietarios ADD CONSTRAINT fk_historial_cliente FOREIGN KEY (id_cliente) REFERENCES dbo.clientes(id_cliente);
ALTER TABLE dbo.historial_estados_ot ADD CONSTRAINT fk_historial_estados_ot_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.historial_estados_ot ADD CONSTRAINT fk_historial_estados_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.historial_propietarios ADD CONSTRAINT fk_historial_usuario FOREIGN KEY (id_usuario_registro) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.historial_propietarios ADD CONSTRAINT fk_historial_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES dbo.vehiculos(id_vehiculo);
ALTER TABLE dbo.ordenes_trabajo ADD CONSTRAINT fk_ordenes_trabajo_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES dbo.cotizaciones(id_cotizacion);
ALTER TABLE dbo.ordenes_trabajo ADD CONSTRAINT fk_ordenes_trabajo_id_usuario_creacion FOREIGN KEY (id_usuario_creacion) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.pagos ADD CONSTRAINT fk_pagos_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.pagos ADD CONSTRAINT fk_pagos_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.pruebas_ot ADD CONSTRAINT fk_pruebas_ot_id_orden FOREIGN KEY (id_orden) REFERENCES dbo.ordenes_trabajo(id_orden);
ALTER TABLE dbo.pruebas_ot ADD CONSTRAINT fk_pruebas_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.recepciones ADD CONSTRAINT fk_recepciones_id_cliente_atencion FOREIGN KEY (id_cliente_atencion) REFERENCES dbo.clientes(id_cliente);
ALTER TABLE dbo.recepciones ADD CONSTRAINT fk_recepciones_id_recepcionista FOREIGN KEY (id_recepcionista) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.recepciones ADD CONSTRAINT fk_recepciones_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES dbo.vehiculos(id_vehiculo);
ALTER TABLE dbo.repuesto_proveedor ADD CONSTRAINT fk_repuesto_proveedor_id_proveedor FOREIGN KEY (id_proveedor) REFERENCES dbo.proveedores(id_proveedor);
ALTER TABLE dbo.repuesto_proveedor ADD CONSTRAINT fk_repuesto_proveedor_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES dbo.repuestos(id_repuesto);
ALTER TABLE dbo.rol_permisos ADD CONSTRAINT fk_rol_permisos_permiso FOREIGN KEY (id_permiso) REFERENCES dbo.permisos(id_permiso);
ALTER TABLE dbo.rol_permisos ADD CONSTRAINT fk_rol_permisos_rol FOREIGN KEY (id_rol) REFERENCES dbo.roles(id_rol);
ALTER TABLE dbo.usuario_roles ADD CONSTRAINT fk_usuario_roles_rol FOREIGN KEY (id_rol) REFERENCES dbo.roles(id_rol);
ALTER TABLE dbo.usuario_roles ADD CONSTRAINT fk_usuario_roles_usuario FOREIGN KEY (id_usuario) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.vehiculos ADD CONSTRAINT fk_vehiculos_cliente FOREIGN KEY (id_cliente) REFERENCES dbo.clientes(id_cliente);
ALTER TABLE dbo.ordenes_trabajo ADD CONSTRAINT ordenes_trabajo_control_autor_fkey FOREIGN KEY (control_autor) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.recepciones ADD CONSTRAINT recepciones_diagnostico_asignado_por_fkey FOREIGN KEY (diagnostico_asignado_por) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.recepciones ADD CONSTRAINT recepciones_id_mecanico_diagnostico_fkey FOREIGN KEY (id_mecanico_diagnostico) REFERENCES dbo.usuarios(id_usuario);
ALTER TABLE dbo.usuarios ADD CONSTRAINT usuarios_id_empleado_fkey FOREIGN KEY (id_empleado) REFERENCES dbo.empleados(id_empleado) ON DELETE NO ACTION;
GO

-- Índices relacionales y unicidad filtrada (NULL no compite con NULL).
CREATE UNIQUE INDEX usuarios_correo_key ON dbo.usuarios (correo) WHERE correo IS NOT NULL;
CREATE INDEX ix_asignaciones_ot_id_orden ON dbo.asignaciones_ot (id_orden);
CREATE INDEX ix_asignaciones_ot_id_usuario ON dbo.asignaciones_ot (id_usuario);
CREATE INDEX ix_auxilios_id_cliente_solicitante ON dbo.auxilios (id_cliente_solicitante);
CREATE INDEX ix_auxilios_id_recepcion ON dbo.auxilios (id_recepcion);
CREATE INDEX ix_auxilios_id_tecnico ON dbo.auxilios (id_tecnico);
CREATE INDEX ix_auxilios_id_unidad_auxilio ON dbo.auxilios (id_unidad_auxilio);
CREATE INDEX ix_auxilios_id_usuario_registro ON dbo.auxilios (id_usuario_registro);
CREATE INDEX ix_auxilios_id_vehiculo ON dbo.auxilios (id_vehiculo);
CREATE INDEX ix_avances_ot_id_orden ON dbo.avances_ot (id_orden);
CREATE INDEX ix_avances_ot_id_usuario ON dbo.avances_ot (id_usuario);
CREATE INDEX ix_clientes_busqueda ON dbo.clientes (nombre) INCLUDE (numero_documento);
CREATE INDEX ix_contactos_cliente_id_cliente ON dbo.contactos_cliente (id_cliente);
CREATE INDEX ix_cotizaciones_id_diagnostico ON dbo.cotizaciones (id_diagnostico);
CREATE INDEX ix_cotizaciones_recepcion ON dbo.cotizaciones (id_recepcion);
CREATE INDEX ix_detalle_cotizacion_id_cotizacion ON dbo.detalle_cotizacion (id_cotizacion);
CREATE INDEX ix_detalle_cotizacion_id_repuesto ON dbo.detalle_cotizacion (id_repuesto);
CREATE INDEX ix_detalle_cotizacion_id_servicio ON dbo.detalle_cotizacion (id_servicio);
CREATE INDEX ix_diagnosticos_estado ON dbo.diagnosticos (estado, finalizado_en);
CREATE INDEX ix_diagnosticos_id_recepcion ON dbo.diagnosticos (id_recepcion);
CREATE INDEX ix_diagnosticos_id_tecnico ON dbo.diagnosticos (id_tecnico);
CREATE INDEX ix_entregas_id_contacto_receptor ON dbo.entregas (id_contacto_receptor);
CREATE INDEX ix_entregas_id_orden ON dbo.entregas (id_orden);
CREATE INDEX ix_entregas_id_usuario_entrega ON dbo.entregas (id_usuario_entrega);
CREATE INDEX ix_entregas_id_usuario_verificador ON dbo.entregas (id_usuario_verificador);
CREATE INDEX ix_historial_estados_ot_id_orden ON dbo.historial_estados_ot (id_orden);
CREATE INDEX ix_historial_estados_ot_id_usuario ON dbo.historial_estados_ot (id_usuario);
CREATE INDEX ix_historial_propietarios_id_cliente ON dbo.historial_propietarios (id_cliente);
CREATE INDEX ix_historial_propietarios_id_usuario_registro ON dbo.historial_propietarios (id_usuario_registro);
CREATE INDEX ix_historial_propietarios_id_vehiculo ON dbo.historial_propietarios (id_vehiculo);
CREATE INDEX ix_ordenes_trabajo_id_cotizacion ON dbo.ordenes_trabajo (id_cotizacion);
CREATE INDEX ix_ordenes_trabajo_id_usuario_creacion ON dbo.ordenes_trabajo (id_usuario_creacion);
CREATE UNIQUE INDEX ix_pago_request ON dbo.pagos (id_orden, request_id) WHERE (request_id IS NOT NULL);
CREATE INDEX ix_pagos_id_orden ON dbo.pagos (id_orden);
CREATE INDEX ix_pagos_id_usuario_registro ON dbo.pagos (id_usuario_registro);
CREATE INDEX ix_pruebas_ot_id_orden ON dbo.pruebas_ot (id_orden);
CREATE INDEX ix_pruebas_ot_id_usuario ON dbo.pruebas_ot (id_usuario);
CREATE INDEX ix_recepciones_id_cliente_atencion ON dbo.recepciones (id_cliente_atencion);
CREATE INDEX ix_recepciones_id_recepcionista ON dbo.recepciones (id_recepcionista);
CREATE INDEX ix_recepciones_id_vehiculo ON dbo.recepciones (id_vehiculo);
CREATE INDEX ix_recepciones_mecanico_diagnostico ON dbo.recepciones (id_mecanico_diagnostico, id_recepcion);
CREATE INDEX ix_repuesto_proveedor_id_proveedor ON dbo.repuesto_proveedor (id_proveedor);
CREATE INDEX ix_repuesto_proveedor_id_repuesto ON dbo.repuesto_proveedor (id_repuesto);
CREATE INDEX ix_rol_permisos_id_permiso ON dbo.rol_permisos (id_permiso);
CREATE INDEX ix_rol_permisos_id_rol ON dbo.rol_permisos (id_rol);
CREATE INDEX ix_usuario_roles_id_rol ON dbo.usuario_roles (id_rol);
CREATE INDEX ix_usuario_roles_id_usuario ON dbo.usuario_roles (id_usuario);
CREATE INDEX ix_vehiculos_id_cliente ON dbo.vehiculos (id_cliente);
CREATE UNIQUE INDEX uq_clientes_tipo_catalogo_numero ON dbo.clientes (id_tipo_documento, numero_documento);
CREATE UNIQUE INDEX uq_diagnosticos_id_recepcion ON dbo.diagnosticos (id_recepcion);
CREATE UNIQUE INDEX uq_entregas_id_orden ON dbo.entregas (id_orden);
CREATE UNIQUE INDEX uq_historial_propietario_vigente ON dbo.historial_propietarios (id_vehiculo) INCLUDE (fecha_fin) WHERE (fecha_fin IS NULL);
CREATE UNIQUE INDEX uq_ordenes_trabajo_codigo_ot ON dbo.ordenes_trabajo (codigo_ot);
CREATE UNIQUE INDEX uq_ordenes_trabajo_id_cotizacion ON dbo.ordenes_trabajo (id_cotizacion);
CREATE UNIQUE INDEX uq_proveedores_documento ON dbo.proveedores (documento);
CREATE UNIQUE INDEX uq_repuesto_proveedor_id_repuesto_id_proveedor ON dbo.repuesto_proveedor (id_repuesto, id_proveedor);
CREATE UNIQUE INDEX uq_servicios_nombre ON dbo.servicios (nombre);
CREATE UNIQUE INDEX uq_unidades_auxilio_codigo ON dbo.unidades_auxilio (codigo);
CREATE UNIQUE INDEX uq_usuarios_empleado ON dbo.usuarios (id_empleado) WHERE id_empleado IS NOT NULL;
GO


-- Funciones de lectura y vistas.
CREATE FUNCTION dbo.jac_normalize(@value nvarchar(max))
RETURNS nvarchar(max)
AS
BEGIN
    -- La comparación de búsqueda usa una colación sin distinción de acentos.
    RETURN LOWER(COALESCE(@value, N'') COLLATE Latin1_General_100_CI_AI_SC);
END;
GO

CREATE FUNCTION dbo.jac_documento_codigo(@value nvarchar(max))
RETURNS nvarchar(40)
AS
BEGIN
    RETURN CASE REPLACE(REPLACE(dbo.jac_normalize(TRIM(@value)), N'.', N''), N' ', N'')
                COLLATE Latin1_General_100_CI_AI_SC
        WHEN N'dni' THEN N'DNI'
        WHEN N'documentonacionaldeidentidad' THEN N'DNI'
        WHEN N'ruc' THEN N'RUC'
        WHEN N'registrounicodecontribuyentes' THEN N'RUC'
        WHEN N'ce' THEN N'CE'
        WHEN N'carnedeextranjeria' THEN N'CE'
        WHEN N'carnetdeextranjeria' THEN N'CE'
        WHEN N'pasaporte' THEN N'PASAPORTE'
        ELSE NULL END;
END;
GO

CREATE FUNCTION dbo.jac_importe_orden(@id_orden bigint)
RETURNS decimal(38,2)
AS
BEGIN
    DECLARE @importe decimal(38,2);

    SELECT @importe = CASE
        WHEN o.flujo_version = 0 OR NOT EXISTS (
            SELECT 1 FROM OPENJSON(CASE WHEN ISJSON(o.control_tecnico) = 1 THEN o.control_tecnico ELSE N'{}' END) AS j
            WHERE j.[key] = N'reconciliation' AND j.[type] <> 0
        ) THEN c.total
        ELSE COALESCE((
            SELECT SUM(d.subtotal) FROM dbo.detalle_cotizacion AS d
            WHERE d.id_cotizacion = c.id_cotizacion
              AND COALESCE(JSON_VALUE(o.control_tecnico, CONCAT(N'$.execution."', d.id_detalle, N'".status')), N'Pendiente') <> N'No realizado'
        ), 0) END
    FROM dbo.ordenes_trabajo AS o
    JOIN dbo.cotizaciones AS c ON c.id_cotizacion = o.id_cotizacion
    WHERE o.id_orden = @id_orden;

    RETURN @importe;
END;
GO

CREATE VIEW dbo.datos_taller_ficha AS
    SELECT N'EL BRAVO DEL JAC' AS name, N'10400848209' AS ruc,
           N'Calle Urano 381, Los Olivos' AS address;
GO

CREATE VIEW dbo.jac_movimientos_pagos AS
    SELECT id_pago, id_orden, monto, fecha_pago AS fecha FROM dbo.pagos
    UNION ALL
    SELECT id_pago, id_orden, -monto, revertido_en FROM dbo.pagos
    WHERE revertido_en IS NOT NULL;
GO

CREATE VIEW dbo.marcas_vehiculos AS
    WITH catalogo AS (
        SELECT CONVERT(nvarchar(80), base.nombre) COLLATE Latin1_General_100_BIN2 AS nombre, 0 AS prioridad
        FROM (VALUES
            (N'JAC'), (N'Toyota'), (N'Nissan'), (N'Hyundai'), (N'Kia'), (N'Chevrolet'),
            (N'Ford'), (N'Volkswagen'), (N'Suzuki'), (N'Mitsubishi'), (N'Mazda'), (N'Renault'),
            (N'Chery'), (N'Changan'), (N'Haval'), (N'BYD'), (N'MG'), (N'Great Wall'),
            (N'BMW'), (N'Mercedes-Benz'), (N'Audi'), (N'Volvo'), (N'Peugeot'), (N'Citroën'),
            (N'BAIC'), (N'Jinbei'), (N'DFM'), (N'Brilliance'), (N'Changhe'), (N'Mahindra'),
            (N'SsangYong'), (N'Daewoo'), (N'Land Rover'), (N'Honda')
        ) AS base(nombre)
        UNION ALL
        SELECT TRIM(marca), 1 FROM dbo.vehiculos WHERE TRIM(marca) <> N''
    ), ranked AS (
        SELECT nombre, ROW_NUMBER() OVER (
            PARTITION BY dbo.jac_normalize(nombre) COLLATE Latin1_General_100_CI_AI_SC
            ORDER BY prioridad, nombre COLLATE Latin1_General_100_BIN2
        ) AS posicion
        FROM catalogo
    )
    SELECT nombre FROM ranked WHERE posicion = 1;
GO

-- Normalizadores de escritura. Las columnas NOT NULL deben estar presentes en
-- la sentencia externa (la API futura resolverá los valores omitidos).

CREATE TRIGGER dbo.jac_cliente_documento ON dbo.clientes
INSTEAD OF INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_cliente)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;

    SELECT i.id_cliente AS id_cliente, CASE WHEN c.changed = 0 THEN i.tipo_documento WHEN d.codigo = N'PASAPORTE' THEN N'Pasaporte' ELSE d.codigo END AS tipo_documento,
           i.numero_documento AS numero_documento,
           i.nombre AS nombre,
           i.telefono AS telefono,
           i.direccion AS direccion,
           i.correo AS correo, CASE WHEN c.changed = 1 THEN d.id_tipo_documento ELSE i.id_tipo_documento END AS id_tipo_documento
    INTO #normalizadas
    FROM inserted AS i
    LEFT JOIN deleted AS o ON o.id_cliente = i.id_cliente
    CROSS APPLY (SELECT CASE WHEN o.id_cliente IS NULL OR (o.id_tipo_documento IS DISTINCT FROM i.id_tipo_documento
        OR CONVERT(varbinary(max), o.tipo_documento) IS DISTINCT FROM CONVERT(varbinary(max), i.tipo_documento)
        OR CONVERT(varbinary(max), o.numero_documento) IS DISTINCT FROM CONVERT(varbinary(max), i.numero_documento))
                             THEN 1 ELSE 0 END AS changed) AS c
    LEFT JOIN dbo.tipos_documento AS d WITH (HOLDLOCK) ON c.changed = 1 AND d.activo = 1
        AND ((i.id_tipo_documento IS NOT NULL AND (o.id_cliente IS NULL OR o.id_tipo_documento IS DISTINCT FROM i.id_tipo_documento)
              AND d.id_tipo_documento = i.id_tipo_documento)
          OR ((i.id_tipo_documento IS NULL OR (o.id_cliente IS NOT NULL AND o.id_tipo_documento = i.id_tipo_documento))
              AND d.codigo = dbo.jac_documento_codigo(i.tipo_documento)));
    IF EXISTS (SELECT 1 FROM #normalizadas WHERE id_tipo_documento IS NULL)
        THROW 51002, N'Tipo de documento inexistente o inactivo.', 1;

    IF EXISTS (SELECT 1 FROM deleted)
        UPDATE t SET tipo_documento = n.tipo_documento,
                     numero_documento = n.numero_documento,
                     nombre = n.nombre,
                     telefono = n.telefono,
                     direccion = n.direccion,
                     correo = n.correo,
                     id_tipo_documento = n.id_tipo_documento
        FROM dbo.clientes AS t JOIN #normalizadas AS n ON n.id_cliente = t.id_cliente;
    ELSE
        INSERT dbo.clientes (id_cliente, tipo_documento, numero_documento, nombre, telefono, direccion, correo, id_tipo_documento)
        SELECT id_cliente, tipo_documento, numero_documento, nombre, telefono, direccion, correo, id_tipo_documento FROM #normalizadas;
END;
GO

CREATE TRIGGER dbo.jac_contacto_documento ON dbo.contactos_cliente
INSTEAD OF INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_contacto)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;

    SELECT i.id_contacto AS id_contacto,
           i.id_cliente AS id_cliente,
           i.nombre AS nombre, CASE WHEN c.changed = 0 THEN i.tipo_documento WHEN d.codigo = N'PASAPORTE' THEN N'Pasaporte' ELSE d.codigo END AS tipo_documento,
           i.numero_documento AS numero_documento,
           i.telefono AS telefono,
           i.relacion AS relacion,
           i.activo AS activo, CASE WHEN c.changed = 1 THEN d.id_tipo_documento ELSE i.id_tipo_documento END AS id_tipo_documento
    INTO #normalizadas
    FROM inserted AS i
    LEFT JOIN deleted AS o ON o.id_contacto = i.id_contacto
    CROSS APPLY (SELECT CASE WHEN o.id_contacto IS NULL OR (o.id_tipo_documento IS DISTINCT FROM i.id_tipo_documento
        OR CONVERT(varbinary(max), o.tipo_documento) IS DISTINCT FROM CONVERT(varbinary(max), i.tipo_documento)
        OR CONVERT(varbinary(max), o.numero_documento) IS DISTINCT FROM CONVERT(varbinary(max), i.numero_documento))
                             THEN 1 ELSE 0 END AS changed) AS c
    LEFT JOIN dbo.tipos_documento AS d WITH (HOLDLOCK) ON c.changed = 1 AND d.activo = 1
        AND ((i.id_tipo_documento IS NOT NULL AND (o.id_contacto IS NULL OR o.id_tipo_documento IS DISTINCT FROM i.id_tipo_documento)
              AND d.id_tipo_documento = i.id_tipo_documento)
          OR ((i.id_tipo_documento IS NULL OR (o.id_contacto IS NOT NULL AND o.id_tipo_documento = i.id_tipo_documento))
              AND d.codigo = dbo.jac_documento_codigo(i.tipo_documento)));
    IF EXISTS (SELECT 1 FROM #normalizadas WHERE id_tipo_documento IS NULL)
        THROW 51002, N'Tipo de documento inexistente o inactivo.', 1;

    IF EXISTS (SELECT 1 FROM deleted)
        UPDATE t SET id_cliente = n.id_cliente,
                     nombre = n.nombre,
                     tipo_documento = n.tipo_documento,
                     numero_documento = n.numero_documento,
                     telefono = n.telefono,
                     relacion = n.relacion,
                     activo = n.activo,
                     id_tipo_documento = n.id_tipo_documento
        FROM dbo.contactos_cliente AS t JOIN #normalizadas AS n ON n.id_contacto = t.id_contacto;
    ELSE
        INSERT dbo.contactos_cliente (id_contacto, id_cliente, nombre, tipo_documento, numero_documento, telefono, relacion, activo, id_tipo_documento)
        SELECT id_contacto, id_cliente, nombre, tipo_documento, numero_documento, telefono, relacion, activo, id_tipo_documento FROM #normalizadas;
END;
GO

CREATE TRIGGER dbo.jac_usuario_nombre_laboral ON dbo.usuarios
INSTEAD OF INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_usuario)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;
    DECLARE @normalizar bit = 0;
    IF UPDATE(id_empleado) OR UPDATE(nombre) SET @normalizar = 1;
    SELECT i.id_usuario AS id_usuario, CASE WHEN i.id_empleado IS NOT NULL AND (NOT EXISTS (SELECT 1 FROM deleted) OR @normalizar = 1) THEN e.nombre ELSE i.nombre END AS nombre,
           i.usuario AS usuario,
           i.correo AS correo,
           i.password_hash AS password_hash,
           i.activo AS activo,
           i.id_empleado AS id_empleado
    INTO #normalizadas
    FROM inserted AS i
    LEFT JOIN dbo.empleados AS e WITH (HOLDLOCK) ON e.id_empleado = i.id_empleado;
    IF EXISTS (SELECT 1 FROM #normalizadas WHERE nombre IS NULL)
        THROW 51003, N'Empleado inexistente o nombre no disponible.', 1;

    IF EXISTS (SELECT 1 FROM deleted)
        UPDATE t SET nombre = n.nombre,
                     usuario = n.usuario,
                     correo = n.correo,
                     password_hash = n.password_hash,
                     activo = n.activo,
                     id_empleado = n.id_empleado
        FROM dbo.usuarios AS t JOIN #normalizadas AS n ON n.id_usuario = t.id_usuario;
    ELSE
        INSERT dbo.usuarios (id_usuario, nombre, usuario, correo, password_hash, activo, id_empleado)
        SELECT id_usuario, nombre, usuario, correo, password_hash, activo, id_empleado FROM #normalizadas;
END;
GO

CREATE TRIGGER dbo.jac_origen_cotizacion ON dbo.cotizaciones
INSTEAD OF INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_cotizacion)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;
    DECLARE @validar bit = 0;
    IF NOT EXISTS (SELECT 1 FROM deleted) OR UPDATE(id_recepcion) OR UPDATE(id_diagnostico) SET @validar = 1;
    SELECT i.id_cotizacion AS id_cotizacion,
           i.id_diagnostico AS id_diagnostico,
           i.estado AS estado,
           i.total AS total,
           i.fecha_emision AS fecha_emision,
           i.aprobada_en AS aprobada_en, CASE WHEN @validar = 1 THEN COALESCE(i.id_recepcion, d.id_recepcion) ELSE i.id_recepcion END AS id_recepcion,
           i.version AS version,
           i.contenido_hash AS contenido_hash,
           i.aprobacion AS aprobacion,
           i.historial_revision AS historial_revision,
           i.id_cotizacion_base AS id_cotizacion_base
    INTO #normalizadas
    FROM inserted AS i
    LEFT JOIN dbo.diagnosticos AS d WITH (HOLDLOCK) ON @validar = 1 AND d.id_diagnostico = i.id_diagnostico;
    IF @validar = 1 AND EXISTS (
        SELECT 1 FROM #normalizadas AS n
        LEFT JOIN dbo.recepciones AS r WITH (UPDLOCK, HOLDLOCK) ON r.id_recepcion = n.id_recepcion
        LEFT JOIN dbo.diagnosticos AS d WITH (HOLDLOCK) ON d.id_diagnostico = n.id_diagnostico
        WHERE r.id_recepcion IS NULL
           OR (r.tipo_atencion = N'Directo' AND n.id_diagnostico IS NOT NULL)
           OR (r.tipo_atencion = N'Diagnostico' AND
               (d.id_recepcion IS DISTINCT FROM n.id_recepcion OR d.estado IS DISTINCT FROM N'Finalizado'))
    )
        THROW 51004, N'La cotización requiere una recepción válida y, cuando corresponde, su diagnóstico finalizado.', 1;

    IF EXISTS (SELECT 1 FROM deleted)
        UPDATE t SET id_diagnostico = n.id_diagnostico,
                     estado = n.estado,
                     total = n.total,
                     fecha_emision = n.fecha_emision,
                     aprobada_en = n.aprobada_en,
                     id_recepcion = n.id_recepcion,
                     version = n.version,
                     contenido_hash = n.contenido_hash,
                     aprobacion = n.aprobacion,
                     historial_revision = n.historial_revision,
                     id_cotizacion_base = n.id_cotizacion_base
        FROM dbo.cotizaciones AS t JOIN #normalizadas AS n ON n.id_cotizacion = t.id_cotizacion;
    ELSE
        INSERT dbo.cotizaciones (id_cotizacion, id_diagnostico, estado, total, fecha_emision, aprobada_en, id_recepcion, version, contenido_hash, aprobacion, historial_revision, id_cotizacion_base)
        SELECT id_cotizacion, id_diagnostico, estado, total, fecha_emision, aprobada_en, id_recepcion, version, contenido_hash, aprobacion, historial_revision, id_cotizacion_base FROM #normalizadas;
END;
GO
-- Reglas cruzadas. Una fila inválida cancela la sentencia completa.
CREATE TRIGGER dbo.jac_empleado_nombre_compatible ON dbo.empleados
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT UPDATE(nombre) RETURN;

    UPDATE u SET nombre = i.nombre
    FROM dbo.usuarios AS u
    JOIN inserted AS i ON i.id_empleado = u.id_empleado
    JOIN deleted AS d ON d.id_empleado = i.id_empleado
    WHERE CONVERT(varbinary(max), i.nombre) <> CONVERT(varbinary(max), d.nombre);
END;
GO

CREATE TRIGGER dbo.jac_tipo_atencion ON dbo.recepciones
AFTER UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT UPDATE(tipo_atencion) RETURN;

    IF EXISTS (
        SELECT 1 FROM inserted AS i
        JOIN deleted AS d ON d.id_recepcion = i.id_recepcion
        WHERE i.tipo_atencion <> d.tipo_atencion AND (
            EXISTS (SELECT 1 FROM dbo.diagnosticos AS t WITH (UPDLOCK, HOLDLOCK) WHERE t.id_recepcion = i.id_recepcion)
            OR EXISTS (SELECT 1 FROM dbo.cotizaciones AS c WITH (UPDLOCK, HOLDLOCK) WHERE c.id_recepcion = i.id_recepcion)
        )
    )
        THROW 51005, N'No se puede cambiar el tipo de una atención con historial.', 1;
END;
GO

CREATE TRIGGER dbo.jac_diagnostico_tecnico ON dbo.diagnosticos
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_diagnostico)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted AS i
        JOIN dbo.recepciones AS r WITH (UPDLOCK, HOLDLOCK) ON r.id_recepcion = i.id_recepcion
        WHERE r.tipo_atencion = N'Directo' OR r.id_mecanico_diagnostico IS DISTINCT FROM i.id_tecnico
    )
        THROW 51006, N'El diagnóstico requiere atención técnica y debe corresponder al mecánico asignado.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted AS i JOIN deleted AS d ON d.id_diagnostico = i.id_diagnostico
        WHERE d.estado = N'Finalizado' AND EXISTS (
            SELECT i.id_recepcion, i.id_tecnico, i.fecha,
                   CONVERT(varbinary(max), i.resumen), i.estado,
                   CONVERT(varbinary(max), i.evaluacion), CONVERT(varbinary(max), i.recomendaciones),
                   CONVERT(varbinary(max), i.observaciones), i.version, i.finalizado_en, i.actualizado_en,
                   CONVERT(varbinary(max), i.historial)
            EXCEPT
            SELECT d.id_recepcion, d.id_tecnico, d.fecha,
                   CONVERT(varbinary(max), d.resumen), d.estado,
                   CONVERT(varbinary(max), d.evaluacion), CONVERT(varbinary(max), d.recomendaciones),
                   CONVERT(varbinary(max), d.observaciones), d.version, d.finalizado_en, d.actualizado_en,
                   CONVERT(varbinary(max), d.historial)
        )
    )
        THROW 51007, N'El diagnóstico finalizado es inmutable.', 1;
END;
GO

CREATE TRIGGER dbo.jac_sin_auxiliar_nuevo ON dbo.asignaciones_ot
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;
    IF EXISTS (SELECT 1 FROM deleted) AND UPDATE(id_asignacion)
        THROW 51001, N'El identificador de la fila es inmutable.', 1;

    IF EXISTS (
        SELECT 1 FROM inserted AS i
        LEFT JOIN deleted AS d ON d.id_asignacion = i.id_asignacion
        WHERE CONVERT(varbinary(max), i.rol_en_trabajo) = CONVERT(varbinary(max), N'AUXILIAR') AND
            (d.id_asignacion IS NULL OR i.id_orden <> d.id_orden OR i.id_usuario <> d.id_usuario
             OR (i.activo = 1 AND d.activo = 0))
    )
        THROW 51008, N'No se crean ni reactivan asignaciones auxiliares; se conserva el historial.', 1;
END;
GO

CREATE TRIGGER dbo.trg_bravo_cotizacion_aprobada ON dbo.ordenes_trabajo
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF EXISTS (SELECT 1 FROM deleted) AND NOT UPDATE(id_cotizacion) RETURN;

    IF EXISTS (
        SELECT 1 FROM inserted AS i
        JOIN dbo.cotizaciones AS c WITH (UPDLOCK, HOLDLOCK) ON c.id_cotizacion = i.id_cotizacion
        WHERE CONVERT(varbinary(max), LOWER(c.estado COLLATE Latin1_General_100_CI_AS_SC))
           <> CONVERT(varbinary(max), N'aprobada')
    )
        THROW 51009, N'Solo puede crearse una orden con cotización aprobada.', 1;
END;
GO

CREATE TRIGGER dbo.trg_bravo_entrega_validada ON dbo.entregas
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;
    IF NOT EXISTS (SELECT 1 FROM inserted) RETURN;

    IF EXISTS (
        SELECT 1 FROM inserted WHERE autorizacion_verificada <> 1
            OR id_usuario_verificador IS NULL OR medio_verificacion IS NULL
    )
        THROW 51010, N'Verifique identidad y autorización antes de entregar.', 1;

    -- Toma los bloqueos de orden antes de leer pagos; el adaptador debe usar
    -- el mismo orden de adquisición para pagos, rectificaciones y entregas.
    DECLARE @ordenes table (id bigint PRIMARY KEY, estado nvarchar(60) NOT NULL);
    INSERT @ordenes (id, estado)
    SELECT o.id_orden, o.estado FROM dbo.ordenes_trabajo AS o WITH (UPDLOCK, HOLDLOCK)
    JOIN (SELECT DISTINCT id_orden FROM inserted) AS i ON i.id_orden = o.id_orden;

    IF EXISTS (
        SELECT 1 FROM @ordenes
        WHERE CONVERT(varbinary(max), estado) <> CONVERT(varbinary(max), N'finalizada')
    )
        THROW 51011, N'La orden debe estar finalizada.', 1;

    IF EXISTS (
        SELECT 1 FROM @ordenes AS o
        OUTER APPLY (
            SELECT SUM(p.monto) AS total FROM dbo.pagos AS p WITH (UPDLOCK, HOLDLOCK)
            WHERE p.id_orden = o.id AND p.revertido_en IS NULL
        ) AS pago
        WHERE dbo.jac_importe_orden(o.id) IS DISTINCT FROM COALESCE(pago.total, 0)
    )
        THROW 51012, N'La entrega requiere saldo neto cero.', 1;
END;
GO

COMMIT TRANSACTION;
GO
