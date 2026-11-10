-- EL BRAVO DEL JAC - SQL OFICIAL PARA LA ENTREGA UNIVERSITARIA
-- Estructura real de elbravodeljac_dev, consultada el 2026-11-05.
-- Exportacion pg_dump --schema-only, en modo de solo lectura.
-- No contiene filas, contrasenas, cuentas globales ni valores actuales de secuencias.
-- Modelo verificado: 28 tablas, 208 columnas, 50 FK.
-- Solo para revisar la estructura o cargarla en una base VACIA independiente.
-- No ejecutar sobre bases existentes ni combinar con esquema_actual.sql/migraciones.
-- El arranque Go/Docker conserva su procedimiento tecnico original, sin cambios.

--
-- PostgreSQL database dump
--

\restrict 6gydFGrSzJnXqrIxjJVUHXKV292uXEhHncASRNDV0fNzTYnaHUzmZBj5S3cUjCR

-- Dumped from database version 18.6 (Debian 18.6-1.pgdg13+2)
-- Dumped by pg_dump version 18.6 (Debian 18.6-1.pgdg13+2)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: pg_trgm; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS pg_trgm WITH SCHEMA public;


--
-- Name: unaccent; Type: EXTENSION; Schema: -; Owner: -
--

CREATE EXTENSION IF NOT EXISTS unaccent WITH SCHEMA public;


--
-- Name: fn_bravo_cotizacion_aprobada(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.fn_bravo_cotizacion_aprobada() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE v_estado text;
BEGIN
  SELECT estado INTO v_estado FROM public.cotizaciones WHERE id_cotizacion=NEW.id_cotizacion;
  IF lower(coalesce(v_estado,'')) <> 'aprobada' THEN
    RAISE EXCEPTION 'Solo puede crearse una orden con cotización aprobada (id=%)',NEW.id_cotizacion USING ERRCODE='23514';
  END IF;
  RETURN NEW;
END $$;


--
-- Name: fn_bravo_entrega_validada(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.fn_bravo_entrega_validada() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE total numeric; paid numeric; state text;
BEGIN
 IF NEW.autorizacion_verificada IS DISTINCT FROM TRUE OR NEW.id_usuario_verificador IS NULL OR NEW.medio_verificacion IS NULL THEN RAISE EXCEPTION 'Verifica identidad y autorización antes de entregar' USING ERRCODE='23514'; END IF;
 SELECT jac_importe_orden(id_orden),estado INTO total,state FROM ordenes_trabajo WHERE id_orden=NEW.id_orden FOR UPDATE;
 IF state<>'finalizada' THEN RAISE EXCEPTION 'La orden debe estar finalizada' USING ERRCODE='23514'; END IF;
 SELECT coalesce(sum(monto),0) INTO paid FROM pagos WHERE id_orden=NEW.id_orden AND revertido_en IS NULL;
 IF total IS NULL OR total<>paid THEN RAISE EXCEPTION 'La entrega requiere saldo neto cero' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_documento_codigo(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_documento_codigo(value text) RETURNS text
    LANGUAGE sql IMMUTABLE
    AS $$
 SELECT CASE regexp_replace(jac_normalize(btrim(value)),'[. ]','','g')
 WHEN 'dni' THEN 'DNI' WHEN 'documentonacionaldeidentidad' THEN 'DNI'
 WHEN 'ruc' THEN 'RUC' WHEN 'registrounicodecontribuyentes' THEN 'RUC'
 WHEN 'ce' THEN 'CE' WHEN 'carnedeextranjeria' THEN 'CE' WHEN 'carnetdeextranjeria' THEN 'CE'
 WHEN 'pasaporte' THEN 'PASAPORTE' ELSE NULL END
$$;


--
-- Name: jac_empleado_nombre_compatible(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_empleado_nombre_compatible() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
 IF NEW.nombre IS DISTINCT FROM OLD.nombre THEN
  UPDATE public.usuarios SET nombre=NEW.nombre WHERE id_empleado=NEW.id_empleado;
 END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_importe_orden(bigint); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_importe_orden(oid bigint) RETURNS numeric
    LANGUAGE sql STABLE
    AS $$
 SELECT CASE WHEN o.flujo_version=0 OR o.control_tecnico->'reconciliation' IS NULL OR o.control_tecnico->'reconciliation'='null'::jsonb THEN q.total
 ELSE coalesce((SELECT sum(d.subtotal) FROM detalle_cotizacion d WHERE d.id_cotizacion=q.id_cotizacion AND coalesce(o.control_tecnico->'execution'->d.id_detalle::text->>'status','Pendiente')<>'No realizado'),0) END FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) WHERE o.id_orden=oid
$$;


--
-- Name: jac_no_nuevo_auxiliar(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_no_nuevo_auxiliar() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
 IF NEW.rol_en_trabajo='AUXILIAR' AND (TG_OP='INSERT' OR NEW.id_orden IS DISTINCT FROM OLD.id_orden OR NEW.id_usuario IS DISTINCT FROM OLD.id_usuario OR (NEW.activo AND NOT OLD.activo)) THEN
  RAISE EXCEPTION 'No se crean ni reactivan asignaciones auxiliares; se conserva el historial' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_normalize(text); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_normalize(value text) RETURNS text
    LANGUAGE sql IMMUTABLE PARALLEL SAFE
    AS $$ SELECT lower(public.unaccent('public.unaccent', coalesce(value,''))) $$;


--
-- Name: jac_persona_tipo_documento(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_persona_tipo_documento() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE doc tipos_documento%ROWTYPE;
BEGIN
 IF TG_OP='UPDATE' AND NEW.tipo_documento IS NOT DISTINCT FROM OLD.tipo_documento AND NEW.id_tipo_documento IS NOT DISTINCT FROM OLD.id_tipo_documento AND NEW.numero_documento IS NOT DISTINCT FROM OLD.numero_documento THEN
  RETURN NEW; -- Retain unknown/inactive historical identity during ordinary edits.
 END IF;
 IF NEW.id_tipo_documento IS NOT NULL AND (TG_OP='INSERT' OR NEW.id_tipo_documento IS DISTINCT FROM OLD.id_tipo_documento) THEN
  SELECT * INTO doc FROM tipos_documento WHERE id_tipo_documento=NEW.id_tipo_documento FOR SHARE;
 ELSE
  SELECT * INTO doc FROM tipos_documento WHERE codigo=jac_documento_codigo(NEW.tipo_documento) FOR SHARE;
 END IF;
 IF doc.id_tipo_documento IS NULL OR NOT doc.activo THEN
  RAISE EXCEPTION 'Tipo de documento inexistente o inactivo' USING ERRCODE='23514';
 END IF;
 NEW.id_tipo_documento=doc.id_tipo_documento;
 NEW.tipo_documento=CASE doc.codigo WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE doc.codigo END;
 RETURN NEW;
END $$;


--
-- Name: jac_proteger_tipo_atencion(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_proteger_tipo_atencion() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
 IF OLD.tipo_atencion IS DISTINCT FROM NEW.tipo_atencion AND (EXISTS(SELECT 1 FROM diagnosticos WHERE id_recepcion=OLD.id_recepcion) OR EXISTS(SELECT 1 FROM cotizaciones WHERE id_recepcion=OLD.id_recepcion)) THEN RAISE EXCEPTION 'No cambiar tipo de una atencion con historial' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_usuario_nombre_laboral(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_usuario_nombre_laboral() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
 IF NEW.id_empleado IS NOT NULL THEN
  SELECT nombre INTO NEW.nombre FROM public.empleados WHERE id_empleado=NEW.id_empleado FOR SHARE;
 END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_validar_diagnostico(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_validar_diagnostico() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE tipo text; tecnico bigint;
BEGIN
 SELECT tipo_atencion,id_mecanico_diagnostico INTO tipo,tecnico FROM recepciones WHERE id_recepcion=NEW.id_recepcion FOR SHARE;
 IF tipo='Directo' THEN RAISE EXCEPTION 'Servicio directo sin diagnostico' USING ERRCODE='23514'; END IF;
 IF tecnico IS NULL OR tecnico IS DISTINCT FROM NEW.id_tecnico THEN RAISE EXCEPTION 'El diagnostico debe corresponder al mecanico asignado' USING ERRCODE='23514'; END IF;
 IF TG_OP='UPDATE' AND OLD.estado='Finalizado' AND NEW IS DISTINCT FROM OLD THEN RAISE EXCEPTION 'El diagnostico finalizado es inmutable' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;


--
-- Name: jac_validar_origen_cotizacion(); Type: FUNCTION; Schema: public; Owner: -
--

CREATE FUNCTION public.jac_validar_origen_cotizacion() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE tipo text; origen bigint; estado_d text;
BEGIN
 IF NEW.id_recepcion IS NULL AND NEW.id_diagnostico IS NOT NULL THEN SELECT id_recepcion INTO NEW.id_recepcion FROM diagnosticos WHERE id_diagnostico=NEW.id_diagnostico; END IF;
 SELECT tipo_atencion INTO tipo FROM recepciones WHERE id_recepcion=NEW.id_recepcion FOR SHARE;
 IF tipo IS NULL THEN RAISE EXCEPTION 'Recepcion inexistente' USING ERRCODE='23514'; END IF;
 IF tipo='Directo' AND NEW.id_diagnostico IS NOT NULL THEN RAISE EXCEPTION 'Servicio directo sin diagnostico' USING ERRCODE='23514'; END IF;
 IF tipo='Diagnostico' THEN
  SELECT id_recepcion,estado INTO origen,estado_d FROM diagnosticos WHERE id_diagnostico=NEW.id_diagnostico;
  IF origen IS DISTINCT FROM NEW.id_recepcion OR estado_d IS DISTINCT FROM 'Finalizado' THEN RAISE EXCEPTION 'Se requiere diagnostico finalizado de la recepcion' USING ERRCODE='23514'; END IF;
 END IF;
 RETURN NEW;
END $$;


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: asignaciones_ot; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.asignaciones_ot (
    id_asignacion bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    rol_en_trabajo character varying(30) NOT NULL,
    fecha_asignacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    activo boolean DEFAULT true NOT NULL
);


--
-- Name: asignaciones_ot_id_asignacion_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.asignaciones_ot_id_asignacion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: asignaciones_ot_id_asignacion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.asignaciones_ot_id_asignacion_seq OWNED BY public.asignaciones_ot.id_asignacion;


--
-- Name: auxilios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.auxilios (
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
);


--
-- Name: auxilios_id_auxilio_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.auxilios_id_auxilio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: auxilios_id_auxilio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.auxilios_id_auxilio_seq OWNED BY public.auxilios.id_auxilio;


--
-- Name: avances_ot; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.avances_ot (
    id_avance bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    observacion text NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: avances_ot_id_avance_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.avances_ot_id_avance_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: avances_ot_id_avance_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.avances_ot_id_avance_seq OWNED BY public.avances_ot.id_avance;


--
-- Name: clientes; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clientes (
    id_cliente bigint NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    nombre character varying(150) NOT NULL,
    telefono character varying(20),
    direccion text,
    correo character varying(150),
    id_tipo_documento bigint NOT NULL
);


--
-- Name: clientes_id_cliente_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.clientes_id_cliente_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: clientes_id_cliente_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.clientes_id_cliente_seq OWNED BY public.clientes.id_cliente;


--
-- Name: contactos_cliente; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.contactos_cliente (
    id_contacto bigint NOT NULL,
    id_cliente bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    telefono character varying(20),
    relacion character varying(50),
    activo boolean DEFAULT true NOT NULL,
    id_tipo_documento bigint NOT NULL
);


--
-- Name: contactos_cliente_id_contacto_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.contactos_cliente_id_contacto_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: contactos_cliente_id_contacto_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.contactos_cliente_id_contacto_seq OWNED BY public.contactos_cliente.id_contacto;


--
-- Name: cotizaciones; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cotizaciones (
    id_cotizacion bigint NOT NULL,
    id_diagnostico bigint,
    estado character varying(20) DEFAULT 'pendiente'::character varying NOT NULL,
    total numeric(12,2) DEFAULT 0 NOT NULL,
    fecha_emision timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    aprobada_en timestamp with time zone,
    id_recepcion bigint NOT NULL,
    version bigint DEFAULT 0 NOT NULL,
    contenido_hash text,
    aprobacion jsonb,
    historial_revision jsonb DEFAULT '[]'::jsonb NOT NULL,
    id_cotizacion_base bigint,
    CONSTRAINT ck_cotizaciones_total CHECK ((total >= (0)::numeric))
);


--
-- Name: cotizaciones_id_cotizacion_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.cotizaciones_id_cotizacion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: cotizaciones_id_cotizacion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.cotizaciones_id_cotizacion_seq OWNED BY public.cotizaciones.id_cotizacion;


--
-- Name: datos_taller_ficha; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.datos_taller_ficha AS
 SELECT 'EL BRAVO DEL JAC'::text AS name,
    '10400848209'::text AS ruc,
    'Calle Urano 381, Los Olivos'::text AS address;


--
-- Name: detalle_cotizacion; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.detalle_cotizacion (
    id_detalle bigint NOT NULL,
    id_cotizacion bigint NOT NULL,
    id_servicio bigint,
    id_repuesto bigint,
    descripcion text,
    cantidad numeric(10,2) NOT NULL,
    precio_unitario numeric(12,2) NOT NULL,
    subtotal numeric(12,2) NOT NULL,
    origen_repuesto character varying(12),
    tipo_concepto character varying(10),
    tipo_repuesto character varying(12),
    descuento_unitario numeric(12,2) DEFAULT 0 NOT NULL,
    CONSTRAINT ck_detalle_cantidad CHECK ((cantidad > (0)::numeric)),
    CONSTRAINT ck_detalle_descuento CHECK (((descuento_unitario >= (0)::numeric) AND (descuento_unitario <= precio_unitario))),
    CONSTRAINT ck_detalle_origen CHECK (((origen_repuesto IS NULL) OR ((origen_repuesto)::text = ANY (ARRAY[('TALLER'::character varying)::text, ('CLIENTE'::character varying)::text])))),
    CONSTRAINT ck_detalle_precio CHECK ((precio_unitario >= (0)::numeric)),
    CONSTRAINT ck_detalle_referencia CHECK ((((id_servicio IS NOT NULL) AND (id_repuesto IS NULL) AND ((tipo_concepto IS NULL) OR ((tipo_concepto)::text = 'Servicio'::text))) OR ((id_repuesto IS NOT NULL) AND (id_servicio IS NULL) AND ((tipo_concepto IS NULL) OR ((tipo_concepto)::text = 'Repuesto'::text))) OR ((id_servicio IS NULL) AND (id_repuesto IS NULL) AND (tipo_concepto IS NOT NULL) AND ((tipo_concepto)::text = ANY ((ARRAY['Servicio'::character varying, 'Repuesto'::character varying])::text[])) AND (descripcion IS NOT NULL) AND (btrim(descripcion) <> ''::text)))),
    CONSTRAINT ck_detalle_subtotal CHECK ((subtotal >= (0)::numeric)),
    CONSTRAINT ck_detalle_tipo_repuesto CHECK (((tipo_repuesto IS NULL) OR (((tipo_repuesto)::text = ANY ((ARRAY['Original'::character varying, 'Alternativo'::character varying])::text[])) AND ((COALESCE(tipo_concepto, (
CASE
    WHEN (id_servicio IS NOT NULL) THEN 'Servicio'::text
    ELSE 'Repuesto'::text
END)::character varying))::text = 'Repuesto'::text))))
);


--
-- Name: detalle_cotizacion_id_detalle_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.detalle_cotizacion_id_detalle_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: detalle_cotizacion_id_detalle_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.detalle_cotizacion_id_detalle_seq OWNED BY public.detalle_cotizacion.id_detalle;


--
-- Name: diagnosticos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.diagnosticos (
    id_diagnostico bigint NOT NULL,
    id_recepcion bigint NOT NULL,
    id_tecnico bigint NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resumen text NOT NULL,
    estado character varying(15) DEFAULT 'Borrador'::character varying NOT NULL,
    evaluacion text DEFAULT ''::text NOT NULL,
    recomendaciones text DEFAULT ''::text NOT NULL,
    observaciones text DEFAULT ''::text NOT NULL,
    version bigint DEFAULT 0 NOT NULL,
    finalizado_en timestamp with time zone,
    actualizado_en timestamp with time zone,
    historial jsonb DEFAULT '[]'::jsonb NOT NULL,
    CONSTRAINT diagnosticos_estado_check CHECK (((estado)::text = ANY ((ARRAY['Borrador'::character varying, 'Finalizado'::character varying])::text[])))
);


--
-- Name: diagnosticos_id_diagnostico_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.diagnosticos_id_diagnostico_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: diagnosticos_id_diagnostico_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.diagnosticos_id_diagnostico_seq OWNED BY public.diagnosticos.id_diagnostico;


--
-- Name: empleados; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.empleados (
    id_empleado bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    tipo_empleado character varying(40) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT empleados_nombre_check CHECK ((btrim((nombre)::text) <> ''::text)),
    CONSTRAINT empleados_tipo_empleado_check CHECK ((btrim((tipo_empleado)::text) <> ''::text))
);


--
-- Name: empleados_id_empleado_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.empleados_id_empleado_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: empleados_id_empleado_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.empleados_id_empleado_seq OWNED BY public.empleados.id_empleado;


--
-- Name: entregas; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.entregas (
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
    id_usuario_verificador bigint NOT NULL,
    ficha_snapshot jsonb,
    ficha_pdf bytea,
    receptor_real jsonb,
    conformidad jsonb
);


--
-- Name: entregas_id_entrega_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.entregas_id_entrega_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: entregas_id_entrega_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.entregas_id_entrega_seq OWNED BY public.entregas.id_entrega;


--
-- Name: historial_estados_ot; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.historial_estados_ot (
    id_historial bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    estado_anterior character varying(30),
    estado_nuevo character varying(30) NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: historial_estados_ot_id_historial_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.historial_estados_ot_id_historial_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: historial_estados_ot_id_historial_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.historial_estados_ot_id_historial_seq OWNED BY public.historial_estados_ot.id_historial;


--
-- Name: historial_propietarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.historial_propietarios (
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
);


--
-- Name: historial_propietarios_id_historial_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.historial_propietarios_id_historial_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: historial_propietarios_id_historial_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.historial_propietarios_id_historial_seq OWNED BY public.historial_propietarios.id_historial;


--
-- Name: pagos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pagos (
    id_pago bigint NOT NULL,
    id_orden bigint NOT NULL,
    monto numeric(12,2) NOT NULL,
    metodo_pago character varying(30) NOT NULL,
    fecha_pago timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    comprobante character varying(50),
    id_usuario_registro bigint NOT NULL,
    request_id character varying(100),
    revertido_en timestamp with time zone,
    rectificacion jsonb,
    CONSTRAINT ck_pagos_monto CHECK ((monto > (0)::numeric))
);


--
-- Name: jac_movimientos_pagos; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.jac_movimientos_pagos AS
 SELECT pagos.id_pago,
    pagos.id_orden,
    pagos.monto,
    pagos.fecha_pago AS fecha
   FROM public.pagos
UNION ALL
 SELECT pagos.id_pago,
    pagos.id_orden,
    (- pagos.monto) AS monto,
    pagos.revertido_en AS fecha
   FROM public.pagos
  WHERE (pagos.revertido_en IS NOT NULL);


--
-- Name: vehiculos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.vehiculos (
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
);


--
-- Name: marcas_vehiculos; Type: VIEW; Schema: public; Owner: -
--

CREATE VIEW public.marcas_vehiculos AS
 SELECT nombre
   FROM ( SELECT DISTINCT ON ((public.jac_normalize(btrim(catalogo.nombre)))) (btrim(catalogo.nombre))::character varying(40) AS nombre,
            catalogo.prioridad
           FROM ( SELECT base.nombre,
                    0 AS prioridad
                   FROM ( VALUES ('JAC'::text), ('Toyota'::text), ('Nissan'::text), ('Hyundai'::text), ('Kia'::text), ('Chevrolet'::text), ('Ford'::text), ('Volkswagen'::text), ('Suzuki'::text), ('Mitsubishi'::text), ('Mazda'::text), ('Renault'::text), ('Chery'::text), ('Changan'::text), ('Haval'::text), ('BYD'::text), ('MG'::text), ('Great Wall'::text), ('BMW'::text), ('Mercedes-Benz'::text), ('Audi'::text), ('Volvo'::text), ('Peugeot'::text), ('Citroën'::text), ('BAIC'::text), ('Jinbei'::text), ('DFM'::text), ('Brilliance'::text), ('Changhe'::text), ('Mahindra'::text), ('SsangYong'::text), ('Daewoo'::text), ('Land Rover'::text), ('Honda'::text)) base(nombre)
                UNION ALL
                 SELECT vehiculos.marca,
                    1
                   FROM public.vehiculos
                  WHERE (btrim((COALESCE(vehiculos.marca, ''::character varying))::text) <> ''::text)) catalogo
          ORDER BY (public.jac_normalize(btrim(catalogo.nombre))), catalogo.prioridad, ((btrim(catalogo.nombre))::character varying(40))) unicas;


--
-- Name: ordenes_trabajo; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.ordenes_trabajo (
    id_orden bigint NOT NULL,
    id_cotizacion bigint NOT NULL,
    codigo_ot character varying(24) NOT NULL,
    estado character varying(30) DEFAULT 'creada'::character varying NOT NULL,
    id_usuario_creacion bigint NOT NULL,
    fecha_creacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    fecha_inicio timestamp with time zone,
    fecha_fin timestamp with time zone,
    control_tecnico jsonb DEFAULT '{}'::jsonb NOT NULL,
    control_version bigint DEFAULT 0 NOT NULL,
    control_autor bigint,
    control_fecha timestamp with time zone,
    flujo_version integer DEFAULT 0 NOT NULL,
    control_historial jsonb DEFAULT '[]'::jsonb NOT NULL
);


--
-- Name: ordenes_trabajo_id_orden_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.ordenes_trabajo_id_orden_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: ordenes_trabajo_id_orden_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.ordenes_trabajo_id_orden_seq OWNED BY public.ordenes_trabajo.id_orden;


--
-- Name: pagos_id_pago_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pagos_id_pago_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pagos_id_pago_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pagos_id_pago_seq OWNED BY public.pagos.id_pago;


--
-- Name: permisos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permisos (
    id_permiso bigint NOT NULL,
    codigo character varying(80) NOT NULL,
    descripcion character varying(150)
);


--
-- Name: permisos_id_permiso_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permisos_id_permiso_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permisos_id_permiso_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permisos_id_permiso_seq OWNED BY public.permisos.id_permiso;


--
-- Name: proveedores; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.proveedores (
    id_proveedor bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    documento character varying(20) NOT NULL,
    telefono character varying(20),
    contacto character varying(120)
);


--
-- Name: proveedores_id_proveedor_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.proveedores_id_proveedor_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: proveedores_id_proveedor_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.proveedores_id_proveedor_seq OWNED BY public.proveedores.id_proveedor;


--
-- Name: pruebas_ot; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.pruebas_ot (
    id_prueba bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    tipo_prueba character varying(100) NOT NULL,
    resultado character varying(20) NOT NULL,
    observaciones text,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: pruebas_ot_id_prueba_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.pruebas_ot_id_prueba_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: pruebas_ot_id_prueba_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.pruebas_ot_id_prueba_seq OWNED BY public.pruebas_ot.id_prueba;


--
-- Name: recepciones; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.recepciones (
    id_recepcion bigint NOT NULL,
    id_vehiculo bigint NOT NULL,
    id_cliente_atencion bigint NOT NULL,
    id_recepcionista bigint NOT NULL,
    fecha_ingreso timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    km_ingreso numeric(10,2),
    sintomas text,
    tipo_atencion character varying(15) DEFAULT 'Diagnostico'::character varying NOT NULL,
    id_mecanico_diagnostico bigint,
    diagnostico_asignado_por bigint,
    diagnostico_asignado_en timestamp with time zone,
    asignacion_version bigint DEFAULT 0 NOT NULL,
    historial_asignaciones jsonb DEFAULT '[]'::jsonb NOT NULL,
    CONSTRAINT ck_recepciones_km CHECK (((km_ingreso IS NULL) OR (km_ingreso >= (0)::numeric))),
    CONSTRAINT recepciones_tipo_atencion_check CHECK (((tipo_atencion)::text = ANY ((ARRAY['Directo'::character varying, 'Diagnostico'::character varying])::text[])))
);


--
-- Name: recepciones_id_recepcion_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.recepciones_id_recepcion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: recepciones_id_recepcion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.recepciones_id_recepcion_seq OWNED BY public.recepciones.id_recepcion;


--
-- Name: repuesto_proveedor; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.repuesto_proveedor (
    id bigint NOT NULL,
    id_repuesto bigint NOT NULL,
    id_proveedor bigint NOT NULL,
    costo_referencia numeric(12,2) NOT NULL,
    CONSTRAINT ck_rp_costo CHECK ((costo_referencia >= (0)::numeric))
);


--
-- Name: repuesto_proveedor_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.repuesto_proveedor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: repuesto_proveedor_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.repuesto_proveedor_id_seq OWNED BY public.repuesto_proveedor.id;


--
-- Name: repuestos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.repuestos (
    id_repuesto bigint NOT NULL,
    nombre character varying(120) NOT NULL,
    stock numeric(12,2) DEFAULT 0 NOT NULL,
    precio_referencia numeric(12,2) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT ck_repuestos_precio CHECK ((precio_referencia >= (0)::numeric)),
    CONSTRAINT ck_repuestos_stock CHECK ((stock >= (0)::numeric))
);


--
-- Name: repuestos_id_repuesto_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.repuestos_id_repuesto_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: repuestos_id_repuesto_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.repuestos_id_repuesto_seq OWNED BY public.repuestos.id_repuesto;


--
-- Name: rol_permisos; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.rol_permisos (
    id_rol bigint NOT NULL,
    id_permiso bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id_rol bigint NOT NULL,
    nombre_rol character varying(60) NOT NULL,
    activo boolean DEFAULT true NOT NULL
);


--
-- Name: roles_id_rol_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_rol_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_rol_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_rol_seq OWNED BY public.roles.id_rol;


--
-- Name: servicios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.servicios (
    id_servicio bigint NOT NULL,
    nombre character varying(100) NOT NULL,
    precio_referencia numeric(12,2) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT ck_servicios_precio CHECK ((precio_referencia >= (0)::numeric))
);


--
-- Name: servicios_id_servicio_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.servicios_id_servicio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: servicios_id_servicio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.servicios_id_servicio_seq OWNED BY public.servicios.id_servicio;


--
-- Name: tipos_documento; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.tipos_documento (
    id_tipo_documento bigint NOT NULL,
    codigo character varying(20) NOT NULL,
    nombre character varying(150) NOT NULL,
    activo boolean DEFAULT true NOT NULL
);


--
-- Name: tipos_documento_id_tipo_documento_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.tipos_documento_id_tipo_documento_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: tipos_documento_id_tipo_documento_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.tipos_documento_id_tipo_documento_seq OWNED BY public.tipos_documento.id_tipo_documento;


--
-- Name: unidades_auxilio; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.unidades_auxilio (
    id_unidad_auxilio bigint NOT NULL,
    codigo character varying(30) NOT NULL,
    descripcion character varying(150),
    activo boolean DEFAULT true NOT NULL
);


--
-- Name: unidades_auxilio_id_unidad_auxilio_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.unidades_auxilio_id_unidad_auxilio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: unidades_auxilio_id_unidad_auxilio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.unidades_auxilio_id_unidad_auxilio_seq OWNED BY public.unidades_auxilio.id_unidad_auxilio;


--
-- Name: usuario_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.usuario_roles (
    id_usuario bigint NOT NULL,
    id_rol bigint NOT NULL
);


--
-- Name: usuarios; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.usuarios (
    id_usuario bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    usuario character varying(60) NOT NULL,
    correo character varying(150),
    password_hash text NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    id_empleado bigint
);


--
-- Name: usuarios_id_usuario_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.usuarios_id_usuario_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: usuarios_id_usuario_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.usuarios_id_usuario_seq OWNED BY public.usuarios.id_usuario;


--
-- Name: vehiculos_id_vehiculo_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.vehiculos_id_vehiculo_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: vehiculos_id_vehiculo_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.vehiculos_id_vehiculo_seq OWNED BY public.vehiculos.id_vehiculo;


--
-- Name: asignaciones_ot id_asignacion; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asignaciones_ot ALTER COLUMN id_asignacion SET DEFAULT nextval('public.asignaciones_ot_id_asignacion_seq'::regclass);


--
-- Name: auxilios id_auxilio; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios ALTER COLUMN id_auxilio SET DEFAULT nextval('public.auxilios_id_auxilio_seq'::regclass);


--
-- Name: avances_ot id_avance; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avances_ot ALTER COLUMN id_avance SET DEFAULT nextval('public.avances_ot_id_avance_seq'::regclass);


--
-- Name: clientes id_cliente; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clientes ALTER COLUMN id_cliente SET DEFAULT nextval('public.clientes_id_cliente_seq'::regclass);


--
-- Name: contactos_cliente id_contacto; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contactos_cliente ALTER COLUMN id_contacto SET DEFAULT nextval('public.contactos_cliente_id_contacto_seq'::regclass);


--
-- Name: cotizaciones id_cotizacion; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cotizaciones ALTER COLUMN id_cotizacion SET DEFAULT nextval('public.cotizaciones_id_cotizacion_seq'::regclass);


--
-- Name: detalle_cotizacion id_detalle; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detalle_cotizacion ALTER COLUMN id_detalle SET DEFAULT nextval('public.detalle_cotizacion_id_detalle_seq'::regclass);


--
-- Name: diagnosticos id_diagnostico; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.diagnosticos ALTER COLUMN id_diagnostico SET DEFAULT nextval('public.diagnosticos_id_diagnostico_seq'::regclass);


--
-- Name: empleados id_empleado; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.empleados ALTER COLUMN id_empleado SET DEFAULT nextval('public.empleados_id_empleado_seq'::regclass);


--
-- Name: entregas id_entrega; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas ALTER COLUMN id_entrega SET DEFAULT nextval('public.entregas_id_entrega_seq'::regclass);


--
-- Name: historial_estados_ot id_historial; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_estados_ot ALTER COLUMN id_historial SET DEFAULT nextval('public.historial_estados_ot_id_historial_seq'::regclass);


--
-- Name: historial_propietarios id_historial; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_propietarios ALTER COLUMN id_historial SET DEFAULT nextval('public.historial_propietarios_id_historial_seq'::regclass);


--
-- Name: ordenes_trabajo id_orden; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ordenes_trabajo ALTER COLUMN id_orden SET DEFAULT nextval('public.ordenes_trabajo_id_orden_seq'::regclass);


--
-- Name: pagos id_pago; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pagos ALTER COLUMN id_pago SET DEFAULT nextval('public.pagos_id_pago_seq'::regclass);


--
-- Name: permisos id_permiso; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permisos ALTER COLUMN id_permiso SET DEFAULT nextval('public.permisos_id_permiso_seq'::regclass);


--
-- Name: proveedores id_proveedor; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proveedores ALTER COLUMN id_proveedor SET DEFAULT nextval('public.proveedores_id_proveedor_seq'::regclass);


--
-- Name: pruebas_ot id_prueba; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pruebas_ot ALTER COLUMN id_prueba SET DEFAULT nextval('public.pruebas_ot_id_prueba_seq'::regclass);


--
-- Name: recepciones id_recepcion; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones ALTER COLUMN id_recepcion SET DEFAULT nextval('public.recepciones_id_recepcion_seq'::regclass);


--
-- Name: repuesto_proveedor id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuesto_proveedor ALTER COLUMN id SET DEFAULT nextval('public.repuesto_proveedor_id_seq'::regclass);


--
-- Name: repuestos id_repuesto; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuestos ALTER COLUMN id_repuesto SET DEFAULT nextval('public.repuestos_id_repuesto_seq'::regclass);


--
-- Name: roles id_rol; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id_rol SET DEFAULT nextval('public.roles_id_rol_seq'::regclass);


--
-- Name: servicios id_servicio; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.servicios ALTER COLUMN id_servicio SET DEFAULT nextval('public.servicios_id_servicio_seq'::regclass);


--
-- Name: tipos_documento id_tipo_documento; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tipos_documento ALTER COLUMN id_tipo_documento SET DEFAULT nextval('public.tipos_documento_id_tipo_documento_seq'::regclass);


--
-- Name: unidades_auxilio id_unidad_auxilio; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.unidades_auxilio ALTER COLUMN id_unidad_auxilio SET DEFAULT nextval('public.unidades_auxilio_id_unidad_auxilio_seq'::regclass);


--
-- Name: usuarios id_usuario; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios ALTER COLUMN id_usuario SET DEFAULT nextval('public.usuarios_id_usuario_seq'::regclass);


--
-- Name: vehiculos id_vehiculo; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vehiculos ALTER COLUMN id_vehiculo SET DEFAULT nextval('public.vehiculos_id_vehiculo_seq'::regclass);


--
-- Name: asignaciones_ot asignaciones_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT asignaciones_ot_pkey PRIMARY KEY (id_asignacion);


--
-- Name: auxilios auxilios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT auxilios_pkey PRIMARY KEY (id_auxilio);


--
-- Name: avances_ot avances_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT avances_ot_pkey PRIMARY KEY (id_avance);


--
-- Name: clientes clientes_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT clientes_pkey PRIMARY KEY (id_cliente);


--
-- Name: contactos_cliente contactos_cliente_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contactos_cliente
    ADD CONSTRAINT contactos_cliente_pkey PRIMARY KEY (id_contacto);


--
-- Name: cotizaciones cotizaciones_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT cotizaciones_pkey PRIMARY KEY (id_cotizacion);


--
-- Name: detalle_cotizacion detalle_cotizacion_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT detalle_cotizacion_pkey PRIMARY KEY (id_detalle);


--
-- Name: diagnosticos diagnosticos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT diagnosticos_pkey PRIMARY KEY (id_diagnostico);


--
-- Name: empleados empleados_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.empleados
    ADD CONSTRAINT empleados_pkey PRIMARY KEY (id_empleado);


--
-- Name: entregas entregas_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT entregas_pkey PRIMARY KEY (id_entrega);


--
-- Name: historial_estados_ot historial_estados_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT historial_estados_ot_pkey PRIMARY KEY (id_historial);


--
-- Name: historial_propietarios historial_propietarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT historial_propietarios_pkey PRIMARY KEY (id_historial);


--
-- Name: ordenes_trabajo ordenes_trabajo_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT ordenes_trabajo_pkey PRIMARY KEY (id_orden);


--
-- Name: pagos pagos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT pagos_pkey PRIMARY KEY (id_pago);


--
-- Name: permisos permisos_codigo_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_codigo_key UNIQUE (codigo);


--
-- Name: permisos permisos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_pkey PRIMARY KEY (id_permiso);


--
-- Name: rol_permisos pk_rol_permisos; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT pk_rol_permisos PRIMARY KEY (id_rol, id_permiso);


--
-- Name: usuario_roles pk_usuario_roles; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT pk_usuario_roles PRIMARY KEY (id_usuario, id_rol);


--
-- Name: proveedores proveedores_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.proveedores
    ADD CONSTRAINT proveedores_pkey PRIMARY KEY (id_proveedor);


--
-- Name: pruebas_ot pruebas_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT pruebas_ot_pkey PRIMARY KEY (id_prueba);


--
-- Name: recepciones recepciones_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT recepciones_pkey PRIMARY KEY (id_recepcion);


--
-- Name: repuesto_proveedor repuesto_proveedor_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT repuesto_proveedor_pkey PRIMARY KEY (id);


--
-- Name: repuestos repuestos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuestos
    ADD CONSTRAINT repuestos_pkey PRIMARY KEY (id_repuesto);


--
-- Name: roles roles_nombre_rol_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_nombre_rol_key UNIQUE (nombre_rol);


--
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id_rol);


--
-- Name: servicios servicios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.servicios
    ADD CONSTRAINT servicios_pkey PRIMARY KEY (id_servicio);


--
-- Name: tipos_documento tipos_documento_codigo_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tipos_documento
    ADD CONSTRAINT tipos_documento_codigo_key UNIQUE (codigo);


--
-- Name: tipos_documento tipos_documento_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.tipos_documento
    ADD CONSTRAINT tipos_documento_pkey PRIMARY KEY (id_tipo_documento);


--
-- Name: unidades_auxilio unidades_auxilio_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.unidades_auxilio
    ADD CONSTRAINT unidades_auxilio_pkey PRIMARY KEY (id_unidad_auxilio);


--
-- Name: clientes uq_cliente_documento; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT uq_cliente_documento UNIQUE (tipo_documento, numero_documento);


--
-- Name: usuarios usuarios_correo_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_correo_key UNIQUE (correo);


--
-- Name: usuarios usuarios_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_pkey PRIMARY KEY (id_usuario);


--
-- Name: usuarios usuarios_usuario_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_usuario_key UNIQUE (usuario);


--
-- Name: vehiculos vehiculos_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT vehiculos_pkey PRIMARY KEY (id_vehiculo);


--
-- Name: vehiculos vehiculos_placa_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT vehiculos_placa_key UNIQUE (placa);


--
-- Name: ix_asignaciones_ot_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_asignaciones_ot_id_orden ON public.asignaciones_ot USING btree (id_orden);


--
-- Name: ix_asignaciones_ot_id_usuario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_asignaciones_ot_id_usuario ON public.asignaciones_ot USING btree (id_usuario);


--
-- Name: ix_auxilios_id_cliente_solicitante; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_cliente_solicitante ON public.auxilios USING btree (id_cliente_solicitante);


--
-- Name: ix_auxilios_id_recepcion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_recepcion ON public.auxilios USING btree (id_recepcion);


--
-- Name: ix_auxilios_id_tecnico; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_tecnico ON public.auxilios USING btree (id_tecnico);


--
-- Name: ix_auxilios_id_unidad_auxilio; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_unidad_auxilio ON public.auxilios USING btree (id_unidad_auxilio);


--
-- Name: ix_auxilios_id_usuario_registro; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_usuario_registro ON public.auxilios USING btree (id_usuario_registro);


--
-- Name: ix_auxilios_id_vehiculo; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_auxilios_id_vehiculo ON public.auxilios USING btree (id_vehiculo);


--
-- Name: ix_avances_ot_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_avances_ot_id_orden ON public.avances_ot USING btree (id_orden);


--
-- Name: ix_avances_ot_id_usuario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_avances_ot_id_usuario ON public.avances_ot USING btree (id_usuario);


--
-- Name: ix_clientes_busqueda; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_clientes_busqueda ON public.clientes USING gin (public.jac_normalize((((nombre)::text || ' '::text) || (numero_documento)::text)) public.gin_trgm_ops);


--
-- Name: ix_contactos_cliente_id_cliente; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_contactos_cliente_id_cliente ON public.contactos_cliente USING btree (id_cliente);


--
-- Name: ix_cotizaciones_id_diagnostico; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_cotizaciones_id_diagnostico ON public.cotizaciones USING btree (id_diagnostico);


--
-- Name: ix_cotizaciones_recepcion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_cotizaciones_recepcion ON public.cotizaciones USING btree (id_recepcion);


--
-- Name: ix_detalle_cotizacion_id_cotizacion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_detalle_cotizacion_id_cotizacion ON public.detalle_cotizacion USING btree (id_cotizacion);


--
-- Name: ix_detalle_cotizacion_id_repuesto; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_detalle_cotizacion_id_repuesto ON public.detalle_cotizacion USING btree (id_repuesto);


--
-- Name: ix_detalle_cotizacion_id_servicio; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_detalle_cotizacion_id_servicio ON public.detalle_cotizacion USING btree (id_servicio);


--
-- Name: ix_diagnosticos_estado; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_diagnosticos_estado ON public.diagnosticos USING btree (estado, finalizado_en);


--
-- Name: ix_diagnosticos_id_recepcion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_diagnosticos_id_recepcion ON public.diagnosticos USING btree (id_recepcion);


--
-- Name: ix_diagnosticos_id_tecnico; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_diagnosticos_id_tecnico ON public.diagnosticos USING btree (id_tecnico);


--
-- Name: ix_entregas_id_contacto_receptor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_entregas_id_contacto_receptor ON public.entregas USING btree (id_contacto_receptor);


--
-- Name: ix_entregas_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_entregas_id_orden ON public.entregas USING btree (id_orden);


--
-- Name: ix_entregas_id_usuario_entrega; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_entregas_id_usuario_entrega ON public.entregas USING btree (id_usuario_entrega);


--
-- Name: ix_entregas_id_usuario_verificador; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_entregas_id_usuario_verificador ON public.entregas USING btree (id_usuario_verificador);


--
-- Name: ix_historial_estados_ot_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_historial_estados_ot_id_orden ON public.historial_estados_ot USING btree (id_orden);


--
-- Name: ix_historial_estados_ot_id_usuario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_historial_estados_ot_id_usuario ON public.historial_estados_ot USING btree (id_usuario);


--
-- Name: ix_historial_propietarios_id_cliente; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_historial_propietarios_id_cliente ON public.historial_propietarios USING btree (id_cliente);


--
-- Name: ix_historial_propietarios_id_usuario_registro; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_historial_propietarios_id_usuario_registro ON public.historial_propietarios USING btree (id_usuario_registro);


--
-- Name: ix_historial_propietarios_id_vehiculo; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_historial_propietarios_id_vehiculo ON public.historial_propietarios USING btree (id_vehiculo);


--
-- Name: ix_ordenes_trabajo_id_cotizacion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_ordenes_trabajo_id_cotizacion ON public.ordenes_trabajo USING btree (id_cotizacion);


--
-- Name: ix_ordenes_trabajo_id_usuario_creacion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_ordenes_trabajo_id_usuario_creacion ON public.ordenes_trabajo USING btree (id_usuario_creacion);


--
-- Name: ix_pago_request; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX ix_pago_request ON public.pagos USING btree (id_orden, request_id) WHERE (request_id IS NOT NULL);


--
-- Name: ix_pagos_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_pagos_id_orden ON public.pagos USING btree (id_orden);


--
-- Name: ix_pagos_id_usuario_registro; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_pagos_id_usuario_registro ON public.pagos USING btree (id_usuario_registro);


--
-- Name: ix_pruebas_ot_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_pruebas_ot_id_orden ON public.pruebas_ot USING btree (id_orden);


--
-- Name: ix_pruebas_ot_id_usuario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_pruebas_ot_id_usuario ON public.pruebas_ot USING btree (id_usuario);


--
-- Name: ix_recepciones_id_cliente_atencion; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_recepciones_id_cliente_atencion ON public.recepciones USING btree (id_cliente_atencion);


--
-- Name: ix_recepciones_id_recepcionista; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_recepciones_id_recepcionista ON public.recepciones USING btree (id_recepcionista);


--
-- Name: ix_recepciones_id_vehiculo; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_recepciones_id_vehiculo ON public.recepciones USING btree (id_vehiculo);


--
-- Name: ix_recepciones_mecanico_diagnostico; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_recepciones_mecanico_diagnostico ON public.recepciones USING btree (id_mecanico_diagnostico, id_recepcion);


--
-- Name: ix_repuesto_proveedor_id_proveedor; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_repuesto_proveedor_id_proveedor ON public.repuesto_proveedor USING btree (id_proveedor);


--
-- Name: ix_repuesto_proveedor_id_repuesto; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_repuesto_proveedor_id_repuesto ON public.repuesto_proveedor USING btree (id_repuesto);


--
-- Name: ix_rol_permisos_id_permiso; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_rol_permisos_id_permiso ON public.rol_permisos USING btree (id_permiso);


--
-- Name: ix_rol_permisos_id_rol; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_rol_permisos_id_rol ON public.rol_permisos USING btree (id_rol);


--
-- Name: ix_usuario_roles_id_rol; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_usuario_roles_id_rol ON public.usuario_roles USING btree (id_rol);


--
-- Name: ix_usuario_roles_id_usuario; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_usuario_roles_id_usuario ON public.usuario_roles USING btree (id_usuario);


--
-- Name: ix_vehiculos_id_cliente; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX ix_vehiculos_id_cliente ON public.vehiculos USING btree (id_cliente);


--
-- Name: uq_clientes_tipo_catalogo_numero; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_clientes_tipo_catalogo_numero ON public.clientes USING btree (id_tipo_documento, numero_documento);


--
-- Name: uq_diagnosticos_id_recepcion; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_diagnosticos_id_recepcion ON public.diagnosticos USING btree (id_recepcion);


--
-- Name: uq_entregas_id_orden; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_entregas_id_orden ON public.entregas USING btree (id_orden);


--
-- Name: uq_historial_propietario_vigente; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_historial_propietario_vigente ON public.historial_propietarios USING btree (id_vehiculo) WHERE (fecha_fin IS NULL);


--
-- Name: uq_ordenes_trabajo_codigo_ot; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_ordenes_trabajo_codigo_ot ON public.ordenes_trabajo USING btree (codigo_ot);


--
-- Name: uq_ordenes_trabajo_id_cotizacion; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_ordenes_trabajo_id_cotizacion ON public.ordenes_trabajo USING btree (id_cotizacion);


--
-- Name: uq_proveedores_documento; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_proveedores_documento ON public.proveedores USING btree (documento);


--
-- Name: uq_repuesto_proveedor_id_repuesto_id_proveedor; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_repuesto_proveedor_id_repuesto_id_proveedor ON public.repuesto_proveedor USING btree (id_repuesto, id_proveedor);


--
-- Name: uq_servicios_nombre; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_servicios_nombre ON public.servicios USING btree (nombre);


--
-- Name: uq_unidades_auxilio_codigo; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_unidades_auxilio_codigo ON public.unidades_auxilio USING btree (codigo);


--
-- Name: uq_usuarios_empleado; Type: INDEX; Schema: public; Owner: -
--

CREATE UNIQUE INDEX uq_usuarios_empleado ON public.usuarios USING btree (id_empleado);


--
-- Name: clientes jac_cliente_documento; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_cliente_documento BEFORE INSERT OR UPDATE OF tipo_documento, id_tipo_documento, numero_documento ON public.clientes FOR EACH ROW EXECUTE FUNCTION public.jac_persona_tipo_documento();


--
-- Name: contactos_cliente jac_contacto_documento; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_contacto_documento BEFORE INSERT OR UPDATE OF tipo_documento, id_tipo_documento, numero_documento ON public.contactos_cliente FOR EACH ROW EXECUTE FUNCTION public.jac_persona_tipo_documento();


--
-- Name: diagnosticos jac_diagnostico_tecnico; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_diagnostico_tecnico BEFORE INSERT OR UPDATE ON public.diagnosticos FOR EACH ROW EXECUTE FUNCTION public.jac_validar_diagnostico();


--
-- Name: empleados jac_empleado_nombre_compatible; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_empleado_nombre_compatible AFTER UPDATE OF nombre ON public.empleados FOR EACH ROW EXECUTE FUNCTION public.jac_empleado_nombre_compatible();


--
-- Name: cotizaciones jac_origen_cotizacion; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_origen_cotizacion BEFORE INSERT OR UPDATE OF id_recepcion, id_diagnostico ON public.cotizaciones FOR EACH ROW EXECUTE FUNCTION public.jac_validar_origen_cotizacion();


--
-- Name: asignaciones_ot jac_sin_auxiliar_nuevo; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_sin_auxiliar_nuevo BEFORE INSERT OR UPDATE ON public.asignaciones_ot FOR EACH ROW EXECUTE FUNCTION public.jac_no_nuevo_auxiliar();


--
-- Name: recepciones jac_tipo_atencion; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_tipo_atencion BEFORE UPDATE OF tipo_atencion ON public.recepciones FOR EACH ROW EXECUTE FUNCTION public.jac_proteger_tipo_atencion();


--
-- Name: usuarios jac_usuario_nombre_laboral; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER jac_usuario_nombre_laboral BEFORE INSERT OR UPDATE OF id_empleado, nombre ON public.usuarios FOR EACH ROW EXECUTE FUNCTION public.jac_usuario_nombre_laboral();


--
-- Name: ordenes_trabajo trg_bravo_cotizacion_aprobada; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_bravo_cotizacion_aprobada BEFORE INSERT OR UPDATE OF id_cotizacion ON public.ordenes_trabajo FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_cotizacion_aprobada();


--
-- Name: entregas trg_bravo_entrega_validada; Type: TRIGGER; Schema: public; Owner: -
--

CREATE TRIGGER trg_bravo_entrega_validada BEFORE INSERT OR UPDATE ON public.entregas FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_entrega_validada();


--
-- Name: clientes clientes_id_tipo_documento_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT clientes_id_tipo_documento_fkey FOREIGN KEY (id_tipo_documento) REFERENCES public.tipos_documento(id_tipo_documento) ON DELETE RESTRICT;


--
-- Name: contactos_cliente contactos_cliente_id_tipo_documento_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contactos_cliente
    ADD CONSTRAINT contactos_cliente_id_tipo_documento_fkey FOREIGN KEY (id_tipo_documento) REFERENCES public.tipos_documento(id_tipo_documento) ON DELETE RESTRICT;


--
-- Name: cotizaciones cotizaciones_id_cotizacion_base_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT cotizaciones_id_cotizacion_base_fkey FOREIGN KEY (id_cotizacion_base) REFERENCES public.cotizaciones(id_cotizacion);


--
-- Name: cotizaciones cotizaciones_id_recepcion_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT cotizaciones_id_recepcion_fkey FOREIGN KEY (id_recepcion) REFERENCES public.recepciones(id_recepcion);


--
-- Name: asignaciones_ot fk_asignaciones_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT fk_asignaciones_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: asignaciones_ot fk_asignaciones_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT fk_asignaciones_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- Name: auxilios fk_auxilios_id_cliente_solicitante; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_cliente_solicitante FOREIGN KEY (id_cliente_solicitante) REFERENCES public.clientes(id_cliente);


--
-- Name: auxilios fk_auxilios_id_recepcion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES public.recepciones(id_recepcion);


--
-- Name: auxilios fk_auxilios_id_tecnico; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES public.usuarios(id_usuario);


--
-- Name: auxilios fk_auxilios_id_unidad_auxilio; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_unidad_auxilio FOREIGN KEY (id_unidad_auxilio) REFERENCES public.unidades_auxilio(id_unidad_auxilio);


--
-- Name: auxilios fk_auxilios_id_usuario_registro; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- Name: auxilios fk_auxilios_id_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- Name: avances_ot fk_avances_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT fk_avances_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: avances_ot fk_avances_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT fk_avances_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- Name: contactos_cliente fk_contactos_cliente_id_cliente; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.contactos_cliente
    ADD CONSTRAINT fk_contactos_cliente_id_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


--
-- Name: cotizaciones fk_cotizaciones_id_diagnostico; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT fk_cotizaciones_id_diagnostico FOREIGN KEY (id_diagnostico) REFERENCES public.diagnosticos(id_diagnostico);


--
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_cotizacion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES public.cotizaciones(id_cotizacion);


--
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_repuesto; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES public.repuestos(id_repuesto);


--
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_servicio; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_servicio FOREIGN KEY (id_servicio) REFERENCES public.servicios(id_servicio);


--
-- Name: diagnosticos fk_diagnosticos_id_recepcion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT fk_diagnosticos_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES public.recepciones(id_recepcion);


--
-- Name: diagnosticos fk_diagnosticos_id_tecnico; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT fk_diagnosticos_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES public.usuarios(id_usuario);


--
-- Name: entregas fk_entregas_id_contacto_receptor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_contacto_receptor FOREIGN KEY (id_contacto_receptor) REFERENCES public.contactos_cliente(id_contacto);


--
-- Name: entregas fk_entregas_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: entregas fk_entregas_id_usuario_entrega; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_usuario_entrega FOREIGN KEY (id_usuario_entrega) REFERENCES public.usuarios(id_usuario);


--
-- Name: entregas fk_entregas_id_usuario_verificador; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_usuario_verificador FOREIGN KEY (id_usuario_verificador) REFERENCES public.usuarios(id_usuario);


--
-- Name: historial_propietarios fk_historial_cliente; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


--
-- Name: historial_estados_ot fk_historial_estados_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT fk_historial_estados_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: historial_estados_ot fk_historial_estados_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT fk_historial_estados_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- Name: historial_propietarios fk_historial_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_usuario FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- Name: historial_propietarios fk_historial_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- Name: ordenes_trabajo fk_ordenes_trabajo_id_cotizacion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT fk_ordenes_trabajo_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES public.cotizaciones(id_cotizacion);


--
-- Name: ordenes_trabajo fk_ordenes_trabajo_id_usuario_creacion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT fk_ordenes_trabajo_id_usuario_creacion FOREIGN KEY (id_usuario_creacion) REFERENCES public.usuarios(id_usuario);


--
-- Name: pagos fk_pagos_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT fk_pagos_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: pagos fk_pagos_id_usuario_registro; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT fk_pagos_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- Name: pruebas_ot fk_pruebas_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT fk_pruebas_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- Name: pruebas_ot fk_pruebas_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT fk_pruebas_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- Name: recepciones fk_recepciones_id_cliente_atencion; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_cliente_atencion FOREIGN KEY (id_cliente_atencion) REFERENCES public.clientes(id_cliente);


--
-- Name: recepciones fk_recepciones_id_recepcionista; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_recepcionista FOREIGN KEY (id_recepcionista) REFERENCES public.usuarios(id_usuario);


--
-- Name: recepciones fk_recepciones_id_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- Name: repuesto_proveedor fk_repuesto_proveedor_id_proveedor; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT fk_repuesto_proveedor_id_proveedor FOREIGN KEY (id_proveedor) REFERENCES public.proveedores(id_proveedor);


--
-- Name: repuesto_proveedor fk_repuesto_proveedor_id_repuesto; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT fk_repuesto_proveedor_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES public.repuestos(id_repuesto);


--
-- Name: rol_permisos fk_rol_permisos_permiso; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT fk_rol_permisos_permiso FOREIGN KEY (id_permiso) REFERENCES public.permisos(id_permiso);


--
-- Name: rol_permisos fk_rol_permisos_rol; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT fk_rol_permisos_rol FOREIGN KEY (id_rol) REFERENCES public.roles(id_rol);


--
-- Name: usuario_roles fk_usuario_roles_rol; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT fk_usuario_roles_rol FOREIGN KEY (id_rol) REFERENCES public.roles(id_rol);


--
-- Name: usuario_roles fk_usuario_roles_usuario; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT fk_usuario_roles_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- Name: vehiculos fk_vehiculos_cliente; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT fk_vehiculos_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


--
-- Name: ordenes_trabajo ordenes_trabajo_control_autor_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT ordenes_trabajo_control_autor_fkey FOREIGN KEY (control_autor) REFERENCES public.usuarios(id_usuario);


--
-- Name: recepciones recepciones_diagnostico_asignado_por_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT recepciones_diagnostico_asignado_por_fkey FOREIGN KEY (diagnostico_asignado_por) REFERENCES public.usuarios(id_usuario);


--
-- Name: recepciones recepciones_id_mecanico_diagnostico_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT recepciones_id_mecanico_diagnostico_fkey FOREIGN KEY (id_mecanico_diagnostico) REFERENCES public.usuarios(id_usuario);


--
-- Name: usuarios usuarios_id_empleado_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_id_empleado_fkey FOREIGN KEY (id_empleado) REFERENCES public.empleados(id_empleado) ON DELETE RESTRICT;


--
-- Marca de instalacion COMPLETA desde el SQL consolidado; no es una migracion.
COMMENT ON SCHEMA public IS 'elbravodeljac:esquema_entrega:v1';

-- PostgreSQL database dump complete
--

\unrestrict 6gydFGrSzJnXqrIxjJVUHXKV292uXEhHncASRNDV0fNzTYnaHUzmZBj5S3cUjCR
