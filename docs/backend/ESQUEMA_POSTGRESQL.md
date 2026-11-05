# Inventario verificado del SQL

Fuente: `database/postgresql/esquema_actual.sql`. SHA-256: `5c4b2b3221c1af2411882d62747332272137ffb1978a1713658579bfdab0de5d`.

Exportación estructural PostgreSQL 18.6. Se contaron **26 tablas y 42 claves foráneas**. No se ejecutó SQL ni se consultó la base viva. No se encontraron bloques COPY ni INSERT INTO en esta exportación.

## Tablas, columnas y restricciones locales

### asignaciones_ot

```sql
id_asignacion bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    rol_en_trabajo character varying(30) NOT NULL,
    fecha_asignacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    activo boolean DEFAULT true NOT NULL
```

### auxilios

```sql
id_auxilio bigint NOT NULL,
    id_vehiculo bigint,
    id_cliente_solicitante bigint,
    id_usuario_registro bigint NOT NULL,
    id_tecnico bigint,
    id_unidad_auxilio bigint,
    id_recepcion bigint,
    ubicacion text NOT NULL,
    tipo_incidente character varying(60),
    descripcion text,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    estado character varying(30) DEFAULT 'pendiente'::character varying NOT NULL
```

### avances_ot

```sql
id_avance bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    observacion text NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
```

### clientes

```sql
id_cliente bigint NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    nombre character varying(150) NOT NULL,
    telefono character varying(20),
    direccion text,
    correo character varying(150)
```

### contactos_cliente

```sql
id_contacto bigint NOT NULL,
    id_cliente bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    telefono character varying(20),
    relacion character varying(50),
    activo boolean DEFAULT true NOT NULL
```

### cotizaciones

```sql
id_cotizacion bigint NOT NULL,
    id_diagnostico bigint NOT NULL,
    estado character varying(20) DEFAULT 'pendiente'::character varying NOT NULL,
    total numeric(12,2) DEFAULT 0 NOT NULL,
    fecha_emision timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    aprobada_en timestamp with time zone,
    CONSTRAINT ck_cotizaciones_total CHECK ((total >= (0)::numeric))
```

### detalle_cotizacion

```sql
id_detalle bigint NOT NULL,
    id_cotizacion bigint NOT NULL,
    id_servicio bigint,
    id_repuesto bigint,
    descripcion text,
    cantidad numeric(10,2) NOT NULL,
    precio_unitario numeric(12,2) NOT NULL,
    subtotal numeric(12,2) NOT NULL,
    origen_repuesto character varying(12),
    CONSTRAINT ck_detalle_cantidad CHECK ((cantidad > (0)::numeric)),
    CONSTRAINT ck_detalle_origen CHECK (((origen_repuesto IS NULL) OR ((origen_repuesto)::text = ANY ((ARRAY['TALLER'::character varying, 'CLIENTE'::character varying])::text[])))),
    CONSTRAINT ck_detalle_precio CHECK ((precio_unitario >= (0)::numeric)),
    CONSTRAINT ck_detalle_referencia CHECK ((((id_servicio IS NOT NULL) AND (id_repuesto IS NULL)) OR ((id_servicio IS NULL) AND (id_repuesto IS NOT NULL)))),
    CONSTRAINT ck_detalle_subtotal CHECK ((subtotal >= (0)::numeric))
```

### diagnosticos

```sql
id_diagnostico bigint NOT NULL,
    id_recepcion bigint NOT NULL,
    id_tecnico bigint NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resumen text NOT NULL
```

### entregas

```sql
id_entrega bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario_entrega bigint NOT NULL,
    id_contacto_receptor bigint,
    fecha_entrega timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    tipo_receptor character varying(30) NOT NULL,
    nombre_receptor_snap character varying(150),
    tipo_doc_receptor_snap character varying(20),
    documento_receptor_snap character varying(30),
    telefono_receptor_snap character varying(20),
    autorizacion_verificada boolean DEFAULT false NOT NULL,
    medio_verificacion character varying(80),
    fecha_verificacion timestamp with time zone,
    id_usuario_verificador bigint NOT NULL
```

### historial_estados_ot

```sql
id_historial bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    estado_anterior character varying(30),
    estado_nuevo character varying(30) NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
```

### historial_propietarios

```sql
id_historial bigint NOT NULL,
    id_vehiculo bigint NOT NULL,
    id_cliente bigint NOT NULL,
    fecha_inicio timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    fecha_fin timestamp with time zone,
    id_usuario_registro bigint NOT NULL,
    verificado boolean DEFAULT false NOT NULL,
    observacion text,
    CONSTRAINT chk_fechas_propietario CHECK (((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio))),
    CONSTRAINT ck_historial_fechas CHECK (((fecha_fin IS NULL) OR (fecha_fin >= fecha_inicio)))
```

### ordenes_trabajo

```sql
id_orden bigint NOT NULL,
    id_cotizacion bigint NOT NULL,
    codigo_ot character varying(24) NOT NULL,
    estado character varying(30) DEFAULT 'creada'::character varying NOT NULL,
    id_usuario_creacion bigint NOT NULL,
    fecha_creacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    fecha_inicio timestamp with time zone,
    fecha_fin timestamp with time zone
```

### pagos

```sql
id_pago bigint NOT NULL,
    id_orden bigint NOT NULL,
    monto numeric(12,2) NOT NULL,
    metodo_pago character varying(30) NOT NULL,
    fecha_pago timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    comprobante character varying(50),
    id_usuario_registro bigint NOT NULL,
    CONSTRAINT ck_pagos_monto CHECK ((monto > (0)::numeric))
```

### permisos

```sql
id_permiso bigint NOT NULL,
    codigo character varying(80) NOT NULL,
    descripcion character varying(150)
```

### proveedores

```sql
id_proveedor bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    documento character varying(20) NOT NULL,
    telefono character varying(20),
    contacto character varying(120)
```

### pruebas_ot

```sql
id_prueba bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    tipo_prueba character varying(100) NOT NULL,
    resultado character varying(20) NOT NULL,
    observaciones text,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
```

### recepciones

```sql
id_recepcion bigint NOT NULL,
    id_vehiculo bigint NOT NULL,
    id_cliente_atencion bigint NOT NULL,
    id_recepcionista bigint NOT NULL,
    fecha_ingreso timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    km_ingreso numeric(10,2),
    sintomas text,
    CONSTRAINT ck_recepciones_km CHECK (((km_ingreso IS NULL) OR (km_ingreso >= (0)::numeric)))
```

### repuesto_proveedor

```sql
id bigint NOT NULL,
    id_repuesto bigint NOT NULL,
    id_proveedor bigint NOT NULL,
    costo_referencia numeric(12,2) NOT NULL,
    CONSTRAINT ck_rp_costo CHECK ((costo_referencia >= (0)::numeric))
```

### repuestos

```sql
id_repuesto bigint NOT NULL,
    nombre character varying(120) NOT NULL,
    stock numeric(12,2) DEFAULT 0 NOT NULL,
    precio_referencia numeric(12,2) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT ck_repuestos_precio CHECK ((precio_referencia >= (0)::numeric)),
    CONSTRAINT ck_repuestos_stock CHECK ((stock >= (0)::numeric))
```

### rol_permisos

```sql
id_rol bigint NOT NULL,
    id_permiso bigint NOT NULL
```

### roles

```sql
id_rol bigint NOT NULL,
    nombre_rol character varying(60) NOT NULL
```

### servicios

```sql
id_servicio bigint NOT NULL,
    nombre character varying(100) NOT NULL,
    precio_referencia numeric(12,2) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT ck_servicios_precio CHECK ((precio_referencia >= (0)::numeric))
```

### unidades_auxilio

```sql
id_unidad_auxilio bigint NOT NULL,
    codigo character varying(30) NOT NULL,
    descripcion character varying(150),
    activo boolean DEFAULT true NOT NULL
```

### usuario_roles

```sql
id_usuario bigint NOT NULL,
    id_rol bigint NOT NULL
```

### usuarios

```sql
id_usuario bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    usuario character varying(60) NOT NULL,
    correo character varying(150),
    password_hash text NOT NULL,
    activo boolean DEFAULT true NOT NULL
```

### vehiculos

```sql
id_vehiculo bigint NOT NULL,
    id_cliente bigint NOT NULL,
    placa character varying(12) NOT NULL,
    marca character varying(40),
    modelo character varying(60),
    anio smallint,
    kilometraje_actual numeric(10,2),
    observaciones text,
    CONSTRAINT ck_vehiculos_anio CHECK (((anio IS NULL) OR ((anio >= 1900) AND (anio <= 2100)))),
    CONSTRAINT ck_vehiculos_km CHECK (((kilometraje_actual IS NULL) OR (kilometraje_actual >= (0)::numeric)))
```

## Relaciones completas

| Tabla origen | Columna | Tabla destino | Columna destino | Restricción |
|---|---|---|---|---|
| asignaciones_ot | id_orden | ordenes_trabajo | id_orden | fk_asignaciones_ot_id_orden |
| asignaciones_ot | id_usuario | usuarios | id_usuario | fk_asignaciones_ot_id_usuario |
| auxilios | id_cliente_solicitante | clientes | id_cliente | fk_auxilios_id_cliente_solicitante |
| auxilios | id_recepcion | recepciones | id_recepcion | fk_auxilios_id_recepcion |
| auxilios | id_tecnico | usuarios | id_usuario | fk_auxilios_id_tecnico |
| auxilios | id_unidad_auxilio | unidades_auxilio | id_unidad_auxilio | fk_auxilios_id_unidad_auxilio |
| auxilios | id_usuario_registro | usuarios | id_usuario | fk_auxilios_id_usuario_registro |
| auxilios | id_vehiculo | vehiculos | id_vehiculo | fk_auxilios_id_vehiculo |
| avances_ot | id_orden | ordenes_trabajo | id_orden | fk_avances_ot_id_orden |
| avances_ot | id_usuario | usuarios | id_usuario | fk_avances_ot_id_usuario |
| contactos_cliente | id_cliente | clientes | id_cliente | fk_contactos_cliente_id_cliente |
| cotizaciones | id_diagnostico | diagnosticos | id_diagnostico | fk_cotizaciones_id_diagnostico |
| detalle_cotizacion | id_cotizacion | cotizaciones | id_cotizacion | fk_detalle_cotizacion_id_cotizacion |
| detalle_cotizacion | id_repuesto | repuestos | id_repuesto | fk_detalle_cotizacion_id_repuesto |
| detalle_cotizacion | id_servicio | servicios | id_servicio | fk_detalle_cotizacion_id_servicio |
| diagnosticos | id_recepcion | recepciones | id_recepcion | fk_diagnosticos_id_recepcion |
| diagnosticos | id_tecnico | usuarios | id_usuario | fk_diagnosticos_id_tecnico |
| entregas | id_contacto_receptor | contactos_cliente | id_contacto | fk_entregas_id_contacto_receptor |
| entregas | id_orden | ordenes_trabajo | id_orden | fk_entregas_id_orden |
| entregas | id_usuario_entrega | usuarios | id_usuario | fk_entregas_id_usuario_entrega |
| entregas | id_usuario_verificador | usuarios | id_usuario | fk_entregas_id_usuario_verificador |
| historial_propietarios | id_cliente | clientes | id_cliente | fk_historial_cliente |
| historial_estados_ot | id_orden | ordenes_trabajo | id_orden | fk_historial_estados_ot_id_orden |
| historial_estados_ot | id_usuario | usuarios | id_usuario | fk_historial_estados_ot_id_usuario |
| historial_propietarios | id_usuario_registro | usuarios | id_usuario | fk_historial_usuario |
| historial_propietarios | id_vehiculo | vehiculos | id_vehiculo | fk_historial_vehiculo |
| ordenes_trabajo | id_cotizacion | cotizaciones | id_cotizacion | fk_ordenes_trabajo_id_cotizacion |
| ordenes_trabajo | id_usuario_creacion | usuarios | id_usuario | fk_ordenes_trabajo_id_usuario_creacion |
| pagos | id_orden | ordenes_trabajo | id_orden | fk_pagos_id_orden |
| pagos | id_usuario_registro | usuarios | id_usuario | fk_pagos_id_usuario_registro |
| pruebas_ot | id_orden | ordenes_trabajo | id_orden | fk_pruebas_ot_id_orden |
| pruebas_ot | id_usuario | usuarios | id_usuario | fk_pruebas_ot_id_usuario |
| recepciones | id_cliente_atencion | clientes | id_cliente | fk_recepciones_id_cliente_atencion |
| recepciones | id_recepcionista | usuarios | id_usuario | fk_recepciones_id_recepcionista |
| recepciones | id_vehiculo | vehiculos | id_vehiculo | fk_recepciones_id_vehiculo |
| repuesto_proveedor | id_proveedor | proveedores | id_proveedor | fk_repuesto_proveedor_id_proveedor |
| repuesto_proveedor | id_repuesto | repuestos | id_repuesto | fk_repuesto_proveedor_id_repuesto |
| rol_permisos | id_permiso | permisos | id_permiso | fk_rol_permisos_permiso |
| rol_permisos | id_rol | roles | id_rol | fk_rol_permisos_rol |
| usuario_roles | id_rol | roles | id_rol | fk_usuario_roles_rol |
| usuario_roles | id_usuario | usuarios | id_usuario | fk_usuario_roles_usuario |
| vehiculos | id_cliente | clientes | id_cliente | fk_vehiculos_cliente |

## Cardinalidades y restricciones adicionales

Clientes→vehículos/recepciones/contactos; vehículo→historial de propietarios y recepciones; recepción→diagnóstico (máximo uno); diagnóstico→cotizaciones; cotización→detalle y OT (máximo una); OT→asignaciones/avances/pruebas/estados/pagos y entrega (máximo una). Usuarios↔roles↔permisos y repuestos↔proveedores son N:M. El historial de propietarios admite solo un propietario vigente por vehículo. Las FK no declaran ON DELETE CASCADE.

Las cardinalidades de máximo uno se derivan de índices únicos; no garantizan que exista el registro dependiente. Usuarios es destino de auditoría de múltiples tablas.

Restricciones e índices únicos y triggers extraídos del archivo:

```sql
ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_codigo_key UNIQUE (codigo);
ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_nombre_rol_key UNIQUE (nombre_rol);
ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT uq_cliente_documento UNIQUE (tipo_documento, numero_documento);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_correo_key UNIQUE (correo);
ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_usuario_key UNIQUE (usuario);
ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT vehiculos_placa_key UNIQUE (placa);
CREATE UNIQUE INDEX uq_diagnosticos_id_recepcion ON public.diagnosticos USING btree (id_recepcion);
CREATE UNIQUE INDEX uq_entregas_id_orden ON public.entregas USING btree (id_orden);
CREATE UNIQUE INDEX uq_historial_propietario_vigente ON public.historial_propietarios USING btree (id_vehiculo) WHERE (fecha_fin IS NULL);
CREATE UNIQUE INDEX uq_ordenes_trabajo_codigo_ot ON public.ordenes_trabajo USING btree (codigo_ot);
CREATE UNIQUE INDEX uq_ordenes_trabajo_id_cotizacion ON public.ordenes_trabajo USING btree (id_cotizacion);
CREATE UNIQUE INDEX uq_proveedores_documento ON public.proveedores USING btree (documento);
CREATE UNIQUE INDEX uq_repuesto_proveedor_id_repuesto_id_proveedor ON public.repuesto_proveedor USING btree (id_repuesto, id_proveedor);
CREATE UNIQUE INDEX uq_servicios_nombre ON public.servicios USING btree (nombre);
CREATE UNIQUE INDEX uq_unidades_auxilio_codigo ON public.unidades_auxilio USING btree (codigo);
CREATE TRIGGER trg_bravo_cotizacion_aprobada BEFORE INSERT OR UPDATE OF id_cotizacion ON public.ordenes_trabajo FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_cotizacion_aprobada();
CREATE TRIGGER trg_bravo_entrega_validada BEFORE INSERT OR UPDATE ON public.entregas FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_entrega_validada();
```

Los CHECK de km/precios no sustituyen reglas de negocio ni controles de concurrencia. Dos CHECK equivalentes de fechas en historial_propietarios son redundantes; se documentan sin modificar el esquema.
