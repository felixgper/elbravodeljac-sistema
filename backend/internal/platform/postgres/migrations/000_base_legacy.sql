--
-- PostgreSQL database dump
--

\restrict Lu9Bnerz4NmBJ4i1sLHJed93WIM8eE1Mngrd3BUaWAN38egpOcvmS5GZbhfmr9o

-- Dumped from database version 18.6
-- Dumped by pg_dump version 18.6

-- Started on 2026-10-26 21:46:04

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
-- TOC entry 269 (class 1255 OID 17158)
-- Name: fn_bravo_cotizacion_aprobada(); Type: FUNCTION; Schema: public; Owner: postgres
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


ALTER FUNCTION public.fn_bravo_cotizacion_aprobada() OWNER TO postgres;

--
-- TOC entry 270 (class 1255 OID 17160)
-- Name: fn_bravo_entrega_validada(); Type: FUNCTION; Schema: public; Owner: postgres
--

CREATE FUNCTION public.fn_bravo_entrega_validada() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
DECLARE v_total numeric(12,2); v_pagado numeric(14,2); v_estado text;
BEGIN
  IF NEW.autorizacion_verificada IS DISTINCT FROM TRUE OR NEW.id_usuario_verificador IS NULL OR NEW.medio_verificacion IS NULL THEN
    RAISE EXCEPTION 'Debe comprobarse la autorización antes de entregar el vehículo' USING ERRCODE='23514';
  END IF;
  SELECT c.total,ot.estado INTO v_total,v_estado FROM public.ordenes_trabajo ot JOIN public.cotizaciones c ON ot.id_cotizacion=c.id_cotizacion WHERE ot.id_orden=NEW.id_orden;
  IF lower(coalesce(v_estado,'')) NOT IN ('finalizada','finalizado','terminada','terminado','entregada','entregado') THEN
    RAISE EXCEPTION 'La orden debe estar finalizada para confirmar la entrega' USING ERRCODE='23514';
  END IF;
  SELECT coalesce(sum(monto),0) INTO v_pagado FROM public.pagos WHERE id_orden=NEW.id_orden;
  IF v_total IS NULL OR v_pagado < v_total THEN
    RAISE EXCEPTION 'Pago insuficiente para entrega: pagado %, total %',v_pagado,v_total USING ERRCODE='23514';
  END IF;
  RETURN NEW;
END $$;


ALTER FUNCTION public.fn_bravo_entrega_validada() OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- TOC entry 250 (class 1259 OID 16761)
-- Name: asignaciones_ot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.asignaciones_ot (
    id_asignacion bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    rol_en_trabajo character varying(30) NOT NULL,
    fecha_asignacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    activo boolean DEFAULT true NOT NULL
);


ALTER TABLE public.asignaciones_ot OWNER TO postgres;

--
-- TOC entry 249 (class 1259 OID 16760)
-- Name: asignaciones_ot_id_asignacion_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.asignaciones_ot_id_asignacion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.asignaciones_ot_id_asignacion_seq OWNER TO postgres;

--
-- TOC entry 5357 (class 0 OID 0)
-- Dependencies: 249
-- Name: asignaciones_ot_id_asignacion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.asignaciones_ot_id_asignacion_seq OWNED BY public.asignaciones_ot.id_asignacion;


--
-- TOC entry 262 (class 1259 OID 16850)
-- Name: auxilios; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.auxilios OWNER TO postgres;

--
-- TOC entry 261 (class 1259 OID 16849)
-- Name: auxilios_id_auxilio_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.auxilios_id_auxilio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.auxilios_id_auxilio_seq OWNER TO postgres;

--
-- TOC entry 5358 (class 0 OID 0)
-- Dependencies: 261
-- Name: auxilios_id_auxilio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.auxilios_id_auxilio_seq OWNED BY public.auxilios.id_auxilio;


--
-- TOC entry 252 (class 1259 OID 16776)
-- Name: avances_ot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.avances_ot (
    id_avance bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    observacion text NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.avances_ot OWNER TO postgres;

--
-- TOC entry 251 (class 1259 OID 16775)
-- Name: avances_ot_id_avance_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.avances_ot_id_avance_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.avances_ot_id_avance_seq OWNER TO postgres;

--
-- TOC entry 5359 (class 0 OID 0)
-- Dependencies: 251
-- Name: avances_ot_id_avance_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.avances_ot_id_avance_seq OWNED BY public.avances_ot.id_avance;


--
-- TOC entry 220 (class 1259 OID 16505)
-- Name: clientes; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.clientes (
    id_cliente bigint NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    nombre character varying(150) NOT NULL,
    telefono character varying(20),
    direccion text,
    correo character varying(150)
);


ALTER TABLE public.clientes OWNER TO postgres;

--
-- TOC entry 219 (class 1259 OID 16504)
-- Name: clientes_id_cliente_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.clientes_id_cliente_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.clientes_id_cliente_seq OWNER TO postgres;

--
-- TOC entry 5360 (class 0 OID 0)
-- Dependencies: 219
-- Name: clientes_id_cliente_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.clientes_id_cliente_seq OWNED BY public.clientes.id_cliente;


--
-- TOC entry 234 (class 1259 OID 16647)
-- Name: contactos_cliente; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.contactos_cliente (
    id_contacto bigint NOT NULL,
    id_cliente bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    tipo_documento character varying(20) NOT NULL,
    numero_documento character varying(20) NOT NULL,
    telefono character varying(20),
    relacion character varying(50),
    activo boolean DEFAULT true NOT NULL
);


ALTER TABLE public.contactos_cliente OWNER TO postgres;

--
-- TOC entry 233 (class 1259 OID 16646)
-- Name: contactos_cliente_id_contacto_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.contactos_cliente_id_contacto_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.contactos_cliente_id_contacto_seq OWNER TO postgres;

--
-- TOC entry 5361 (class 0 OID 0)
-- Dependencies: 233
-- Name: contactos_cliente_id_contacto_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.contactos_cliente_id_contacto_seq OWNED BY public.contactos_cliente.id_contacto;


--
-- TOC entry 240 (class 1259 OID 16691)
-- Name: cotizaciones; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.cotizaciones (
    id_cotizacion bigint NOT NULL,
    id_diagnostico bigint NOT NULL,
    estado character varying(20) DEFAULT 'pendiente'::character varying NOT NULL,
    total numeric(12,2) DEFAULT 0 NOT NULL,
    fecha_emision timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    aprobada_en timestamp with time zone,
    CONSTRAINT ck_cotizaciones_total CHECK ((total >= (0)::numeric))
);


ALTER TABLE public.cotizaciones OWNER TO postgres;

--
-- TOC entry 239 (class 1259 OID 16690)
-- Name: cotizaciones_id_cotizacion_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.cotizaciones_id_cotizacion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.cotizaciones_id_cotizacion_seq OWNER TO postgres;

--
-- TOC entry 5362 (class 0 OID 0)
-- Dependencies: 239
-- Name: cotizaciones_id_cotizacion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.cotizaciones_id_cotizacion_seq OWNED BY public.cotizaciones.id_cotizacion;


--
-- TOC entry 242 (class 1259 OID 16706)
-- Name: detalle_cotizacion; Type: TABLE; Schema: public; Owner: postgres
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
    CONSTRAINT ck_detalle_cantidad CHECK ((cantidad > (0)::numeric)),
    CONSTRAINT ck_detalle_origen CHECK (((origen_repuesto IS NULL) OR ((origen_repuesto)::text = ANY ((ARRAY['TALLER'::character varying, 'CLIENTE'::character varying])::text[])))),
    CONSTRAINT ck_detalle_precio CHECK ((precio_unitario >= (0)::numeric)),
    CONSTRAINT ck_detalle_referencia CHECK ((((id_servicio IS NOT NULL) AND (id_repuesto IS NULL)) OR ((id_servicio IS NULL) AND (id_repuesto IS NOT NULL)))),
    CONSTRAINT ck_detalle_subtotal CHECK ((subtotal >= (0)::numeric))
);


ALTER TABLE public.detalle_cotizacion OWNER TO postgres;

--
-- TOC entry 241 (class 1259 OID 16705)
-- Name: detalle_cotizacion_id_detalle_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.detalle_cotizacion_id_detalle_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.detalle_cotizacion_id_detalle_seq OWNER TO postgres;

--
-- TOC entry 5363 (class 0 OID 0)
-- Dependencies: 241
-- Name: detalle_cotizacion_id_detalle_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.detalle_cotizacion_id_detalle_seq OWNED BY public.detalle_cotizacion.id_detalle;


--
-- TOC entry 238 (class 1259 OID 16676)
-- Name: diagnosticos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.diagnosticos (
    id_diagnostico bigint NOT NULL,
    id_recepcion bigint NOT NULL,
    id_tecnico bigint NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    resumen text NOT NULL
);


ALTER TABLE public.diagnosticos OWNER TO postgres;

--
-- TOC entry 237 (class 1259 OID 16675)
-- Name: diagnosticos_id_diagnostico_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.diagnosticos_id_diagnostico_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.diagnosticos_id_diagnostico_seq OWNER TO postgres;

--
-- TOC entry 5364 (class 0 OID 0)
-- Dependencies: 237
-- Name: diagnosticos_id_diagnostico_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.diagnosticos_id_diagnostico_seq OWNED BY public.diagnosticos.id_diagnostico;


--
-- TOC entry 260 (class 1259 OID 16834)
-- Name: entregas; Type: TABLE; Schema: public; Owner: postgres
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
    id_usuario_verificador bigint NOT NULL
);


ALTER TABLE public.entregas OWNER TO postgres;

--
-- TOC entry 259 (class 1259 OID 16833)
-- Name: entregas_id_entrega_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.entregas_id_entrega_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.entregas_id_entrega_seq OWNER TO postgres;

--
-- TOC entry 5365 (class 0 OID 0)
-- Dependencies: 259
-- Name: entregas_id_entrega_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.entregas_id_entrega_seq OWNED BY public.entregas.id_entrega;


--
-- TOC entry 256 (class 1259 OID 16807)
-- Name: historial_estados_ot; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.historial_estados_ot (
    id_historial bigint NOT NULL,
    id_orden bigint NOT NULL,
    id_usuario bigint NOT NULL,
    estado_anterior character varying(30),
    estado_nuevo character varying(30) NOT NULL,
    fecha timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


ALTER TABLE public.historial_estados_ot OWNER TO postgres;

--
-- TOC entry 255 (class 1259 OID 16806)
-- Name: historial_estados_ot_id_historial_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.historial_estados_ot_id_historial_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.historial_estados_ot_id_historial_seq OWNER TO postgres;

--
-- TOC entry 5366 (class 0 OID 0)
-- Dependencies: 255
-- Name: historial_estados_ot_id_historial_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.historial_estados_ot_id_historial_seq OWNED BY public.historial_estados_ot.id_historial;


--
-- TOC entry 226 (class 1259 OID 16558)
-- Name: historial_propietarios; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.historial_propietarios OWNER TO postgres;

--
-- TOC entry 225 (class 1259 OID 16557)
-- Name: historial_propietarios_id_historial_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.historial_propietarios_id_historial_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.historial_propietarios_id_historial_seq OWNER TO postgres;

--
-- TOC entry 5367 (class 0 OID 0)
-- Dependencies: 225
-- Name: historial_propietarios_id_historial_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.historial_propietarios_id_historial_seq OWNED BY public.historial_propietarios.id_historial;


--
-- TOC entry 248 (class 1259 OID 16746)
-- Name: ordenes_trabajo; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.ordenes_trabajo (
    id_orden bigint NOT NULL,
    id_cotizacion bigint NOT NULL,
    codigo_ot character varying(24) NOT NULL,
    estado character varying(30) DEFAULT 'creada'::character varying NOT NULL,
    id_usuario_creacion bigint NOT NULL,
    fecha_creacion timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    fecha_inicio timestamp with time zone,
    fecha_fin timestamp with time zone
);


ALTER TABLE public.ordenes_trabajo OWNER TO postgres;

--
-- TOC entry 247 (class 1259 OID 16745)
-- Name: ordenes_trabajo_id_orden_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.ordenes_trabajo_id_orden_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.ordenes_trabajo_id_orden_seq OWNER TO postgres;

--
-- TOC entry 5368 (class 0 OID 0)
-- Dependencies: 247
-- Name: ordenes_trabajo_id_orden_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.ordenes_trabajo_id_orden_seq OWNED BY public.ordenes_trabajo.id_orden;


--
-- TOC entry 258 (class 1259 OID 16820)
-- Name: pagos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.pagos (
    id_pago bigint NOT NULL,
    id_orden bigint NOT NULL,
    monto numeric(12,2) NOT NULL,
    metodo_pago character varying(30) NOT NULL,
    fecha_pago timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    comprobante character varying(50),
    id_usuario_registro bigint NOT NULL,
    CONSTRAINT ck_pagos_monto CHECK ((monto > (0)::numeric))
);


ALTER TABLE public.pagos OWNER TO postgres;

--
-- TOC entry 257 (class 1259 OID 16819)
-- Name: pagos_id_pago_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pagos_id_pago_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pagos_id_pago_seq OWNER TO postgres;

--
-- TOC entry 5369 (class 0 OID 0)
-- Dependencies: 257
-- Name: pagos_id_pago_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pagos_id_pago_seq OWNED BY public.pagos.id_pago;


--
-- TOC entry 230 (class 1259 OID 16602)
-- Name: permisos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.permisos (
    id_permiso bigint NOT NULL,
    codigo character varying(80) NOT NULL,
    descripcion character varying(150)
);


ALTER TABLE public.permisos OWNER TO postgres;

--
-- TOC entry 229 (class 1259 OID 16601)
-- Name: permisos_id_permiso_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.permisos_id_permiso_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.permisos_id_permiso_seq OWNER TO postgres;

--
-- TOC entry 5370 (class 0 OID 0)
-- Dependencies: 229
-- Name: permisos_id_permiso_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.permisos_id_permiso_seq OWNED BY public.permisos.id_permiso;


--
-- TOC entry 266 (class 1259 OID 16877)
-- Name: proveedores; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.proveedores (
    id_proveedor bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    documento character varying(20) NOT NULL,
    telefono character varying(20),
    contacto character varying(120)
);


ALTER TABLE public.proveedores OWNER TO postgres;

--
-- TOC entry 265 (class 1259 OID 16876)
-- Name: proveedores_id_proveedor_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.proveedores_id_proveedor_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.proveedores_id_proveedor_seq OWNER TO postgres;

--
-- TOC entry 5371 (class 0 OID 0)
-- Dependencies: 265
-- Name: proveedores_id_proveedor_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.proveedores_id_proveedor_seq OWNED BY public.proveedores.id_proveedor;


--
-- TOC entry 254 (class 1259 OID 16791)
-- Name: pruebas_ot; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.pruebas_ot OWNER TO postgres;

--
-- TOC entry 253 (class 1259 OID 16790)
-- Name: pruebas_ot_id_prueba_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.pruebas_ot_id_prueba_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.pruebas_ot_id_prueba_seq OWNER TO postgres;

--
-- TOC entry 5372 (class 0 OID 0)
-- Dependencies: 253
-- Name: pruebas_ot_id_prueba_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.pruebas_ot_id_prueba_seq OWNED BY public.pruebas_ot.id_prueba;


--
-- TOC entry 236 (class 1259 OID 16661)
-- Name: recepciones; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.recepciones (
    id_recepcion bigint NOT NULL,
    id_vehiculo bigint NOT NULL,
    id_cliente_atencion bigint NOT NULL,
    id_recepcionista bigint NOT NULL,
    fecha_ingreso timestamp with time zone DEFAULT CURRENT_TIMESTAMP NOT NULL,
    km_ingreso numeric(10,2),
    sintomas text,
    CONSTRAINT ck_recepciones_km CHECK (((km_ingreso IS NULL) OR (km_ingreso >= (0)::numeric)))
);


ALTER TABLE public.recepciones OWNER TO postgres;

--
-- TOC entry 235 (class 1259 OID 16660)
-- Name: recepciones_id_recepcion_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.recepciones_id_recepcion_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.recepciones_id_recepcion_seq OWNER TO postgres;

--
-- TOC entry 5373 (class 0 OID 0)
-- Dependencies: 235
-- Name: recepciones_id_recepcion_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.recepciones_id_recepcion_seq OWNED BY public.recepciones.id_recepcion;


--
-- TOC entry 268 (class 1259 OID 16887)
-- Name: repuesto_proveedor; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.repuesto_proveedor (
    id bigint NOT NULL,
    id_repuesto bigint NOT NULL,
    id_proveedor bigint NOT NULL,
    costo_referencia numeric(12,2) NOT NULL,
    CONSTRAINT ck_rp_costo CHECK ((costo_referencia >= (0)::numeric))
);


ALTER TABLE public.repuesto_proveedor OWNER TO postgres;

--
-- TOC entry 267 (class 1259 OID 16886)
-- Name: repuesto_proveedor_id_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.repuesto_proveedor_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.repuesto_proveedor_id_seq OWNER TO postgres;

--
-- TOC entry 5374 (class 0 OID 0)
-- Dependencies: 267
-- Name: repuesto_proveedor_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.repuesto_proveedor_id_seq OWNED BY public.repuesto_proveedor.id;


--
-- TOC entry 246 (class 1259 OID 16732)
-- Name: repuestos; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.repuestos OWNER TO postgres;

--
-- TOC entry 245 (class 1259 OID 16731)
-- Name: repuestos_id_repuesto_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.repuestos_id_repuesto_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.repuestos_id_repuesto_seq OWNER TO postgres;

--
-- TOC entry 5375 (class 0 OID 0)
-- Dependencies: 245
-- Name: repuestos_id_repuesto_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.repuestos_id_repuesto_seq OWNED BY public.repuestos.id_repuesto;


--
-- TOC entry 232 (class 1259 OID 16629)
-- Name: rol_permisos; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.rol_permisos (
    id_rol bigint NOT NULL,
    id_permiso bigint NOT NULL
);


ALTER TABLE public.rol_permisos OWNER TO postgres;

--
-- TOC entry 228 (class 1259 OID 16591)
-- Name: roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.roles (
    id_rol bigint NOT NULL,
    nombre_rol character varying(60) NOT NULL
);


ALTER TABLE public.roles OWNER TO postgres;

--
-- TOC entry 227 (class 1259 OID 16590)
-- Name: roles_id_rol_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.roles_id_rol_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.roles_id_rol_seq OWNER TO postgres;

--
-- TOC entry 5376 (class 0 OID 0)
-- Dependencies: 227
-- Name: roles_id_rol_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.roles_id_rol_seq OWNED BY public.roles.id_rol;


--
-- TOC entry 244 (class 1259 OID 16720)
-- Name: servicios; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.servicios (
    id_servicio bigint NOT NULL,
    nombre character varying(100) NOT NULL,
    precio_referencia numeric(12,2) NOT NULL,
    activo boolean DEFAULT true NOT NULL,
    CONSTRAINT ck_servicios_precio CHECK ((precio_referencia >= (0)::numeric))
);


ALTER TABLE public.servicios OWNER TO postgres;

--
-- TOC entry 243 (class 1259 OID 16719)
-- Name: servicios_id_servicio_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.servicios_id_servicio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.servicios_id_servicio_seq OWNER TO postgres;

--
-- TOC entry 5377 (class 0 OID 0)
-- Dependencies: 243
-- Name: servicios_id_servicio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.servicios_id_servicio_seq OWNED BY public.servicios.id_servicio;


--
-- TOC entry 264 (class 1259 OID 16866)
-- Name: unidades_auxilio; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.unidades_auxilio (
    id_unidad_auxilio bigint NOT NULL,
    codigo character varying(30) NOT NULL,
    descripcion character varying(150),
    activo boolean DEFAULT true NOT NULL
);


ALTER TABLE public.unidades_auxilio OWNER TO postgres;

--
-- TOC entry 263 (class 1259 OID 16865)
-- Name: unidades_auxilio_id_unidad_auxilio_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.unidades_auxilio_id_unidad_auxilio_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.unidades_auxilio_id_unidad_auxilio_seq OWNER TO postgres;

--
-- TOC entry 5378 (class 0 OID 0)
-- Dependencies: 263
-- Name: unidades_auxilio_id_unidad_auxilio_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.unidades_auxilio_id_unidad_auxilio_seq OWNED BY public.unidades_auxilio.id_unidad_auxilio;


--
-- TOC entry 231 (class 1259 OID 16612)
-- Name: usuario_roles; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.usuario_roles (
    id_usuario bigint NOT NULL,
    id_rol bigint NOT NULL
);


ALTER TABLE public.usuario_roles OWNER TO postgres;

--
-- TOC entry 224 (class 1259 OID 16539)
-- Name: usuarios; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.usuarios (
    id_usuario bigint NOT NULL,
    nombre character varying(150) NOT NULL,
    usuario character varying(60) NOT NULL,
    correo character varying(150),
    password_hash text NOT NULL,
    activo boolean DEFAULT true NOT NULL
);


ALTER TABLE public.usuarios OWNER TO postgres;

--
-- TOC entry 223 (class 1259 OID 16538)
-- Name: usuarios_id_usuario_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.usuarios_id_usuario_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.usuarios_id_usuario_seq OWNER TO postgres;

--
-- TOC entry 5379 (class 0 OID 0)
-- Dependencies: 223
-- Name: usuarios_id_usuario_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.usuarios_id_usuario_seq OWNED BY public.usuarios.id_usuario;


--
-- TOC entry 222 (class 1259 OID 16520)
-- Name: vehiculos; Type: TABLE; Schema: public; Owner: postgres
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


ALTER TABLE public.vehiculos OWNER TO postgres;

--
-- TOC entry 221 (class 1259 OID 16519)
-- Name: vehiculos_id_vehiculo_seq; Type: SEQUENCE; Schema: public; Owner: postgres
--

CREATE SEQUENCE public.vehiculos_id_vehiculo_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


ALTER SEQUENCE public.vehiculos_id_vehiculo_seq OWNER TO postgres;

--
-- TOC entry 5380 (class 0 OID 0)
-- Dependencies: 221
-- Name: vehiculos_id_vehiculo_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: postgres
--

ALTER SEQUENCE public.vehiculos_id_vehiculo_seq OWNED BY public.vehiculos.id_vehiculo;


--
-- TOC entry 5009 (class 2604 OID 16764)
-- Name: asignaciones_ot id_asignacion; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.asignaciones_ot ALTER COLUMN id_asignacion SET DEFAULT nextval('public.asignaciones_ot_id_asignacion_seq'::regclass);


--
-- TOC entry 5023 (class 2604 OID 16853)
-- Name: auxilios id_auxilio; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios ALTER COLUMN id_auxilio SET DEFAULT nextval('public.auxilios_id_auxilio_seq'::regclass);


--
-- TOC entry 5012 (class 2604 OID 16779)
-- Name: avances_ot id_avance; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.avances_ot ALTER COLUMN id_avance SET DEFAULT nextval('public.avances_ot_id_avance_seq'::regclass);


--
-- TOC entry 4981 (class 2604 OID 16508)
-- Name: clientes id_cliente; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.clientes ALTER COLUMN id_cliente SET DEFAULT nextval('public.clientes_id_cliente_seq'::regclass);


--
-- TOC entry 4990 (class 2604 OID 16650)
-- Name: contactos_cliente id_contacto; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contactos_cliente ALTER COLUMN id_contacto SET DEFAULT nextval('public.contactos_cliente_id_contacto_seq'::regclass);


--
-- TOC entry 4996 (class 2604 OID 16694)
-- Name: cotizaciones id_cotizacion; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cotizaciones ALTER COLUMN id_cotizacion SET DEFAULT nextval('public.cotizaciones_id_cotizacion_seq'::regclass);


--
-- TOC entry 5000 (class 2604 OID 16709)
-- Name: detalle_cotizacion id_detalle; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.detalle_cotizacion ALTER COLUMN id_detalle SET DEFAULT nextval('public.detalle_cotizacion_id_detalle_seq'::regclass);


--
-- TOC entry 4994 (class 2604 OID 16679)
-- Name: diagnosticos id_diagnostico; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.diagnosticos ALTER COLUMN id_diagnostico SET DEFAULT nextval('public.diagnosticos_id_diagnostico_seq'::regclass);


--
-- TOC entry 5020 (class 2604 OID 16837)
-- Name: entregas id_entrega; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas ALTER COLUMN id_entrega SET DEFAULT nextval('public.entregas_id_entrega_seq'::regclass);


--
-- TOC entry 5016 (class 2604 OID 16810)
-- Name: historial_estados_ot id_historial; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_estados_ot ALTER COLUMN id_historial SET DEFAULT nextval('public.historial_estados_ot_id_historial_seq'::regclass);


--
-- TOC entry 4985 (class 2604 OID 16561)
-- Name: historial_propietarios id_historial; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_propietarios ALTER COLUMN id_historial SET DEFAULT nextval('public.historial_propietarios_id_historial_seq'::regclass);


--
-- TOC entry 5006 (class 2604 OID 16749)
-- Name: ordenes_trabajo id_orden; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordenes_trabajo ALTER COLUMN id_orden SET DEFAULT nextval('public.ordenes_trabajo_id_orden_seq'::regclass);


--
-- TOC entry 5018 (class 2604 OID 16823)
-- Name: pagos id_pago; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pagos ALTER COLUMN id_pago SET DEFAULT nextval('public.pagos_id_pago_seq'::regclass);


--
-- TOC entry 4989 (class 2604 OID 16605)
-- Name: permisos id_permiso; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permisos ALTER COLUMN id_permiso SET DEFAULT nextval('public.permisos_id_permiso_seq'::regclass);


--
-- TOC entry 5028 (class 2604 OID 16880)
-- Name: proveedores id_proveedor; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.proveedores ALTER COLUMN id_proveedor SET DEFAULT nextval('public.proveedores_id_proveedor_seq'::regclass);


--
-- TOC entry 5014 (class 2604 OID 16794)
-- Name: pruebas_ot id_prueba; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pruebas_ot ALTER COLUMN id_prueba SET DEFAULT nextval('public.pruebas_ot_id_prueba_seq'::regclass);


--
-- TOC entry 4992 (class 2604 OID 16664)
-- Name: recepciones id_recepcion; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recepciones ALTER COLUMN id_recepcion SET DEFAULT nextval('public.recepciones_id_recepcion_seq'::regclass);


--
-- TOC entry 5029 (class 2604 OID 16890)
-- Name: repuesto_proveedor id; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuesto_proveedor ALTER COLUMN id SET DEFAULT nextval('public.repuesto_proveedor_id_seq'::regclass);


--
-- TOC entry 5003 (class 2604 OID 16735)
-- Name: repuestos id_repuesto; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuestos ALTER COLUMN id_repuesto SET DEFAULT nextval('public.repuestos_id_repuesto_seq'::regclass);


--
-- TOC entry 4988 (class 2604 OID 16594)
-- Name: roles id_rol; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles ALTER COLUMN id_rol SET DEFAULT nextval('public.roles_id_rol_seq'::regclass);


--
-- TOC entry 5001 (class 2604 OID 16723)
-- Name: servicios id_servicio; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.servicios ALTER COLUMN id_servicio SET DEFAULT nextval('public.servicios_id_servicio_seq'::regclass);


--
-- TOC entry 5026 (class 2604 OID 16869)
-- Name: unidades_auxilio id_unidad_auxilio; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.unidades_auxilio ALTER COLUMN id_unidad_auxilio SET DEFAULT nextval('public.unidades_auxilio_id_unidad_auxilio_seq'::regclass);


--
-- TOC entry 4983 (class 2604 OID 16542)
-- Name: usuarios id_usuario; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuarios ALTER COLUMN id_usuario SET DEFAULT nextval('public.usuarios_id_usuario_seq'::regclass);


--
-- TOC entry 4982 (class 2604 OID 16523)
-- Name: vehiculos id_vehiculo; Type: DEFAULT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vehiculos ALTER COLUMN id_vehiculo SET DEFAULT nextval('public.vehiculos_id_vehiculo_seq'::regclass);


--
-- TOC entry 5116 (class 2606 OID 16774)
-- Name: asignaciones_ot asignaciones_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT asignaciones_ot_pkey PRIMARY KEY (id_asignacion);


--
-- TOC entry 5143 (class 2606 OID 16864)
-- Name: auxilios auxilios_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT auxilios_pkey PRIMARY KEY (id_auxilio);


--
-- TOC entry 5120 (class 2606 OID 16789)
-- Name: avances_ot avances_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT avances_ot_pkey PRIMARY KEY (id_avance);


--
-- TOC entry 5047 (class 2606 OID 16516)
-- Name: clientes clientes_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT clientes_pkey PRIMARY KEY (id_cliente);


--
-- TOC entry 5084 (class 2606 OID 16659)
-- Name: contactos_cliente contactos_cliente_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contactos_cliente
    ADD CONSTRAINT contactos_cliente_pkey PRIMARY KEY (id_contacto);


--
-- TOC entry 5097 (class 2606 OID 16704)
-- Name: cotizaciones cotizaciones_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT cotizaciones_pkey PRIMARY KEY (id_cotizacion);


--
-- TOC entry 5100 (class 2606 OID 16718)
-- Name: detalle_cotizacion detalle_cotizacion_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT detalle_cotizacion_pkey PRIMARY KEY (id_detalle);


--
-- TOC entry 5092 (class 2606 OID 16689)
-- Name: diagnosticos diagnosticos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT diagnosticos_pkey PRIMARY KEY (id_diagnostico);


--
-- TOC entry 5136 (class 2606 OID 16848)
-- Name: entregas entregas_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT entregas_pkey PRIMARY KEY (id_entrega);


--
-- TOC entry 5128 (class 2606 OID 16818)
-- Name: historial_estados_ot historial_estados_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT historial_estados_ot_pkey PRIMARY KEY (id_historial);


--
-- TOC entry 5062 (class 2606 OID 16574)
-- Name: historial_propietarios historial_propietarios_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT historial_propietarios_pkey PRIMARY KEY (id_historial);


--
-- TOC entry 5112 (class 2606 OID 16759)
-- Name: ordenes_trabajo ordenes_trabajo_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT ordenes_trabajo_pkey PRIMARY KEY (id_orden);


--
-- TOC entry 5134 (class 2606 OID 16832)
-- Name: pagos pagos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT pagos_pkey PRIMARY KEY (id_pago);


--
-- TOC entry 5072 (class 2606 OID 16611)
-- Name: permisos permisos_codigo_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_codigo_key UNIQUE (codigo);


--
-- TOC entry 5074 (class 2606 OID 16609)
-- Name: permisos permisos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.permisos
    ADD CONSTRAINT permisos_pkey PRIMARY KEY (id_permiso);


--
-- TOC entry 5082 (class 2606 OID 16635)
-- Name: rol_permisos pk_rol_permisos; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT pk_rol_permisos PRIMARY KEY (id_rol, id_permiso);


--
-- TOC entry 5078 (class 2606 OID 16618)
-- Name: usuario_roles pk_usuario_roles; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT pk_usuario_roles PRIMARY KEY (id_usuario, id_rol);


--
-- TOC entry 5154 (class 2606 OID 16885)
-- Name: proveedores proveedores_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.proveedores
    ADD CONSTRAINT proveedores_pkey PRIMARY KEY (id_proveedor);


--
-- TOC entry 5126 (class 2606 OID 16805)
-- Name: pruebas_ot pruebas_ot_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT pruebas_ot_pkey PRIMARY KEY (id_prueba);


--
-- TOC entry 5090 (class 2606 OID 16674)
-- Name: recepciones recepciones_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT recepciones_pkey PRIMARY KEY (id_recepcion);


--
-- TOC entry 5159 (class 2606 OID 16896)
-- Name: repuesto_proveedor repuesto_proveedor_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT repuesto_proveedor_pkey PRIMARY KEY (id);


--
-- TOC entry 5108 (class 2606 OID 16744)
-- Name: repuestos repuestos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuestos
    ADD CONSTRAINT repuestos_pkey PRIMARY KEY (id_repuesto);


--
-- TOC entry 5068 (class 2606 OID 16600)
-- Name: roles roles_nombre_rol_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_nombre_rol_key UNIQUE (nombre_rol);


--
-- TOC entry 5070 (class 2606 OID 16598)
-- Name: roles roles_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey PRIMARY KEY (id_rol);


--
-- TOC entry 5105 (class 2606 OID 16730)
-- Name: servicios servicios_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.servicios
    ADD CONSTRAINT servicios_pkey PRIMARY KEY (id_servicio);


--
-- TOC entry 5151 (class 2606 OID 16875)
-- Name: unidades_auxilio unidades_auxilio_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.unidades_auxilio
    ADD CONSTRAINT unidades_auxilio_pkey PRIMARY KEY (id_unidad_auxilio);


--
-- TOC entry 5049 (class 2606 OID 16518)
-- Name: clientes uq_cliente_documento; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.clientes
    ADD CONSTRAINT uq_cliente_documento UNIQUE (tipo_documento, numero_documento);


--
-- TOC entry 5056 (class 2606 OID 16556)
-- Name: usuarios usuarios_correo_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_correo_key UNIQUE (correo);


--
-- TOC entry 5058 (class 2606 OID 16552)
-- Name: usuarios usuarios_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_pkey PRIMARY KEY (id_usuario);


--
-- TOC entry 5060 (class 2606 OID 16554)
-- Name: usuarios usuarios_usuario_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuarios
    ADD CONSTRAINT usuarios_usuario_key UNIQUE (usuario);


--
-- TOC entry 5052 (class 2606 OID 16530)
-- Name: vehiculos vehiculos_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT vehiculos_pkey PRIMARY KEY (id_vehiculo);


--
-- TOC entry 5054 (class 2606 OID 16532)
-- Name: vehiculos vehiculos_placa_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT vehiculos_placa_key UNIQUE (placa);


--
-- TOC entry 5117 (class 1259 OID 17132)
-- Name: ix_asignaciones_ot_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_asignaciones_ot_id_orden ON public.asignaciones_ot USING btree (id_orden);


--
-- TOC entry 5118 (class 1259 OID 17133)
-- Name: ix_asignaciones_ot_id_usuario; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_asignaciones_ot_id_usuario ON public.asignaciones_ot USING btree (id_usuario);


--
-- TOC entry 5144 (class 1259 OID 17147)
-- Name: ix_auxilios_id_cliente_solicitante; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_cliente_solicitante ON public.auxilios USING btree (id_cliente_solicitante);


--
-- TOC entry 5145 (class 1259 OID 17151)
-- Name: ix_auxilios_id_recepcion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_recepcion ON public.auxilios USING btree (id_recepcion);


--
-- TOC entry 5146 (class 1259 OID 17149)
-- Name: ix_auxilios_id_tecnico; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_tecnico ON public.auxilios USING btree (id_tecnico);


--
-- TOC entry 5147 (class 1259 OID 17150)
-- Name: ix_auxilios_id_unidad_auxilio; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_unidad_auxilio ON public.auxilios USING btree (id_unidad_auxilio);


--
-- TOC entry 5148 (class 1259 OID 17148)
-- Name: ix_auxilios_id_usuario_registro; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_usuario_registro ON public.auxilios USING btree (id_usuario_registro);


--
-- TOC entry 5149 (class 1259 OID 17146)
-- Name: ix_auxilios_id_vehiculo; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_auxilios_id_vehiculo ON public.auxilios USING btree (id_vehiculo);


--
-- TOC entry 5121 (class 1259 OID 17134)
-- Name: ix_avances_ot_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_avances_ot_id_orden ON public.avances_ot USING btree (id_orden);


--
-- TOC entry 5122 (class 1259 OID 17135)
-- Name: ix_avances_ot_id_usuario; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_avances_ot_id_usuario ON public.avances_ot USING btree (id_usuario);


--
-- TOC entry 5085 (class 1259 OID 17120)
-- Name: ix_contactos_cliente_id_cliente; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_contactos_cliente_id_cliente ON public.contactos_cliente USING btree (id_cliente);


--
-- TOC entry 5098 (class 1259 OID 17126)
-- Name: ix_cotizaciones_id_diagnostico; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_cotizaciones_id_diagnostico ON public.cotizaciones USING btree (id_diagnostico);


--
-- TOC entry 5101 (class 1259 OID 17127)
-- Name: ix_detalle_cotizacion_id_cotizacion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_detalle_cotizacion_id_cotizacion ON public.detalle_cotizacion USING btree (id_cotizacion);


--
-- TOC entry 5102 (class 1259 OID 17129)
-- Name: ix_detalle_cotizacion_id_repuesto; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_detalle_cotizacion_id_repuesto ON public.detalle_cotizacion USING btree (id_repuesto);


--
-- TOC entry 5103 (class 1259 OID 17128)
-- Name: ix_detalle_cotizacion_id_servicio; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_detalle_cotizacion_id_servicio ON public.detalle_cotizacion USING btree (id_servicio);


--
-- TOC entry 5093 (class 1259 OID 17124)
-- Name: ix_diagnosticos_id_recepcion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_diagnosticos_id_recepcion ON public.diagnosticos USING btree (id_recepcion);


--
-- TOC entry 5094 (class 1259 OID 17125)
-- Name: ix_diagnosticos_id_tecnico; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_diagnosticos_id_tecnico ON public.diagnosticos USING btree (id_tecnico);


--
-- TOC entry 5137 (class 1259 OID 17145)
-- Name: ix_entregas_id_contacto_receptor; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_entregas_id_contacto_receptor ON public.entregas USING btree (id_contacto_receptor);


--
-- TOC entry 5138 (class 1259 OID 17142)
-- Name: ix_entregas_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_entregas_id_orden ON public.entregas USING btree (id_orden);


--
-- TOC entry 5139 (class 1259 OID 17143)
-- Name: ix_entregas_id_usuario_entrega; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_entregas_id_usuario_entrega ON public.entregas USING btree (id_usuario_entrega);


--
-- TOC entry 5140 (class 1259 OID 17144)
-- Name: ix_entregas_id_usuario_verificador; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_entregas_id_usuario_verificador ON public.entregas USING btree (id_usuario_verificador);


--
-- TOC entry 5129 (class 1259 OID 17138)
-- Name: ix_historial_estados_ot_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_historial_estados_ot_id_orden ON public.historial_estados_ot USING btree (id_orden);


--
-- TOC entry 5130 (class 1259 OID 17139)
-- Name: ix_historial_estados_ot_id_usuario; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_historial_estados_ot_id_usuario ON public.historial_estados_ot USING btree (id_usuario);


--
-- TOC entry 5063 (class 1259 OID 17118)
-- Name: ix_historial_propietarios_id_cliente; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_historial_propietarios_id_cliente ON public.historial_propietarios USING btree (id_cliente);


--
-- TOC entry 5064 (class 1259 OID 17119)
-- Name: ix_historial_propietarios_id_usuario_registro; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_historial_propietarios_id_usuario_registro ON public.historial_propietarios USING btree (id_usuario_registro);


--
-- TOC entry 5065 (class 1259 OID 17117)
-- Name: ix_historial_propietarios_id_vehiculo; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_historial_propietarios_id_vehiculo ON public.historial_propietarios USING btree (id_vehiculo);


--
-- TOC entry 5109 (class 1259 OID 17130)
-- Name: ix_ordenes_trabajo_id_cotizacion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_ordenes_trabajo_id_cotizacion ON public.ordenes_trabajo USING btree (id_cotizacion);


--
-- TOC entry 5110 (class 1259 OID 17131)
-- Name: ix_ordenes_trabajo_id_usuario_creacion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_ordenes_trabajo_id_usuario_creacion ON public.ordenes_trabajo USING btree (id_usuario_creacion);


--
-- TOC entry 5131 (class 1259 OID 17140)
-- Name: ix_pagos_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_pagos_id_orden ON public.pagos USING btree (id_orden);


--
-- TOC entry 5132 (class 1259 OID 17141)
-- Name: ix_pagos_id_usuario_registro; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_pagos_id_usuario_registro ON public.pagos USING btree (id_usuario_registro);


--
-- TOC entry 5123 (class 1259 OID 17136)
-- Name: ix_pruebas_ot_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_pruebas_ot_id_orden ON public.pruebas_ot USING btree (id_orden);


--
-- TOC entry 5124 (class 1259 OID 17137)
-- Name: ix_pruebas_ot_id_usuario; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_pruebas_ot_id_usuario ON public.pruebas_ot USING btree (id_usuario);


--
-- TOC entry 5086 (class 1259 OID 17122)
-- Name: ix_recepciones_id_cliente_atencion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_recepciones_id_cliente_atencion ON public.recepciones USING btree (id_cliente_atencion);


--
-- TOC entry 5087 (class 1259 OID 17123)
-- Name: ix_recepciones_id_recepcionista; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_recepciones_id_recepcionista ON public.recepciones USING btree (id_recepcionista);


--
-- TOC entry 5088 (class 1259 OID 17121)
-- Name: ix_recepciones_id_vehiculo; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_recepciones_id_vehiculo ON public.recepciones USING btree (id_vehiculo);


--
-- TOC entry 5156 (class 1259 OID 17153)
-- Name: ix_repuesto_proveedor_id_proveedor; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_repuesto_proveedor_id_proveedor ON public.repuesto_proveedor USING btree (id_proveedor);


--
-- TOC entry 5157 (class 1259 OID 17152)
-- Name: ix_repuesto_proveedor_id_repuesto; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_repuesto_proveedor_id_repuesto ON public.repuesto_proveedor USING btree (id_repuesto);


--
-- TOC entry 5079 (class 1259 OID 17157)
-- Name: ix_rol_permisos_id_permiso; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_rol_permisos_id_permiso ON public.rol_permisos USING btree (id_permiso);


--
-- TOC entry 5080 (class 1259 OID 17156)
-- Name: ix_rol_permisos_id_rol; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_rol_permisos_id_rol ON public.rol_permisos USING btree (id_rol);


--
-- TOC entry 5075 (class 1259 OID 17155)
-- Name: ix_usuario_roles_id_rol; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_usuario_roles_id_rol ON public.usuario_roles USING btree (id_rol);


--
-- TOC entry 5076 (class 1259 OID 17154)
-- Name: ix_usuario_roles_id_usuario; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_usuario_roles_id_usuario ON public.usuario_roles USING btree (id_usuario);


--
-- TOC entry 5050 (class 1259 OID 17116)
-- Name: ix_vehiculos_id_cliente; Type: INDEX; Schema: public; Owner: postgres
--

CREATE INDEX ix_vehiculos_id_cliente ON public.vehiculos USING btree (id_cliente);


--
-- TOC entry 5095 (class 1259 OID 16922)
-- Name: uq_diagnosticos_id_recepcion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_diagnosticos_id_recepcion ON public.diagnosticos USING btree (id_recepcion);


--
-- TOC entry 5141 (class 1259 OID 16925)
-- Name: uq_entregas_id_orden; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_entregas_id_orden ON public.entregas USING btree (id_orden);


--
-- TOC entry 5066 (class 1259 OID 17115)
-- Name: uq_historial_propietario_vigente; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_historial_propietario_vigente ON public.historial_propietarios USING btree (id_vehiculo) WHERE (fecha_fin IS NULL);


--
-- TOC entry 5113 (class 1259 OID 16924)
-- Name: uq_ordenes_trabajo_codigo_ot; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_ordenes_trabajo_codigo_ot ON public.ordenes_trabajo USING btree (codigo_ot);


--
-- TOC entry 5114 (class 1259 OID 16923)
-- Name: uq_ordenes_trabajo_id_cotizacion; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_ordenes_trabajo_id_cotizacion ON public.ordenes_trabajo USING btree (id_cotizacion);


--
-- TOC entry 5155 (class 1259 OID 16928)
-- Name: uq_proveedores_documento; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_proveedores_documento ON public.proveedores USING btree (documento);


--
-- TOC entry 5160 (class 1259 OID 16929)
-- Name: uq_repuesto_proveedor_id_repuesto_id_proveedor; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_repuesto_proveedor_id_repuesto_id_proveedor ON public.repuesto_proveedor USING btree (id_repuesto, id_proveedor);


--
-- TOC entry 5106 (class 1259 OID 16926)
-- Name: uq_servicios_nombre; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_servicios_nombre ON public.servicios USING btree (nombre);


--
-- TOC entry 5152 (class 1259 OID 16927)
-- Name: uq_unidades_auxilio_codigo; Type: INDEX; Schema: public; Owner: postgres
--

CREATE UNIQUE INDEX uq_unidades_auxilio_codigo ON public.unidades_auxilio USING btree (codigo);


--
-- TOC entry 5203 (class 2620 OID 17159)
-- Name: ordenes_trabajo trg_bravo_cotizacion_aprobada; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_bravo_cotizacion_aprobada BEFORE INSERT OR UPDATE OF id_cotizacion ON public.ordenes_trabajo FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_cotizacion_aprobada();


--
-- TOC entry 5204 (class 2620 OID 17161)
-- Name: entregas trg_bravo_entrega_validada; Type: TRIGGER; Schema: public; Owner: postgres
--

CREATE TRIGGER trg_bravo_entrega_validada BEFORE INSERT OR UPDATE ON public.entregas FOR EACH ROW EXECUTE FUNCTION public.fn_bravo_entrega_validada();


--
-- TOC entry 5181 (class 2606 OID 16990)
-- Name: asignaciones_ot fk_asignaciones_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT fk_asignaciones_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5182 (class 2606 OID 16995)
-- Name: asignaciones_ot fk_asignaciones_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.asignaciones_ot
    ADD CONSTRAINT fk_asignaciones_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5195 (class 2606 OID 17065)
-- Name: auxilios fk_auxilios_id_cliente_solicitante; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_cliente_solicitante FOREIGN KEY (id_cliente_solicitante) REFERENCES public.clientes(id_cliente);


--
-- TOC entry 5196 (class 2606 OID 17085)
-- Name: auxilios fk_auxilios_id_recepcion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES public.recepciones(id_recepcion);


--
-- TOC entry 5197 (class 2606 OID 17075)
-- Name: auxilios fk_auxilios_id_tecnico; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5198 (class 2606 OID 17080)
-- Name: auxilios fk_auxilios_id_unidad_auxilio; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_unidad_auxilio FOREIGN KEY (id_unidad_auxilio) REFERENCES public.unidades_auxilio(id_unidad_auxilio);


--
-- TOC entry 5199 (class 2606 OID 17070)
-- Name: auxilios fk_auxilios_id_usuario_registro; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5200 (class 2606 OID 17060)
-- Name: auxilios fk_auxilios_id_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.auxilios
    ADD CONSTRAINT fk_auxilios_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- TOC entry 5183 (class 2606 OID 17000)
-- Name: avances_ot fk_avances_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT fk_avances_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5184 (class 2606 OID 17005)
-- Name: avances_ot fk_avances_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.avances_ot
    ADD CONSTRAINT fk_avances_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5169 (class 2606 OID 16930)
-- Name: contactos_cliente fk_contactos_cliente_id_cliente; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.contactos_cliente
    ADD CONSTRAINT fk_contactos_cliente_id_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


--
-- TOC entry 5175 (class 2606 OID 16960)
-- Name: cotizaciones fk_cotizaciones_id_diagnostico; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.cotizaciones
    ADD CONSTRAINT fk_cotizaciones_id_diagnostico FOREIGN KEY (id_diagnostico) REFERENCES public.diagnosticos(id_diagnostico);


--
-- TOC entry 5176 (class 2606 OID 16965)
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_cotizacion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES public.cotizaciones(id_cotizacion);


--
-- TOC entry 5177 (class 2606 OID 16975)
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_repuesto; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES public.repuestos(id_repuesto);


--
-- TOC entry 5178 (class 2606 OID 16970)
-- Name: detalle_cotizacion fk_detalle_cotizacion_id_servicio; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.detalle_cotizacion
    ADD CONSTRAINT fk_detalle_cotizacion_id_servicio FOREIGN KEY (id_servicio) REFERENCES public.servicios(id_servicio);


--
-- TOC entry 5173 (class 2606 OID 16950)
-- Name: diagnosticos fk_diagnosticos_id_recepcion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT fk_diagnosticos_id_recepcion FOREIGN KEY (id_recepcion) REFERENCES public.recepciones(id_recepcion);


--
-- TOC entry 5174 (class 2606 OID 16955)
-- Name: diagnosticos fk_diagnosticos_id_tecnico; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.diagnosticos
    ADD CONSTRAINT fk_diagnosticos_id_tecnico FOREIGN KEY (id_tecnico) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5191 (class 2606 OID 17055)
-- Name: entregas fk_entregas_id_contacto_receptor; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_contacto_receptor FOREIGN KEY (id_contacto_receptor) REFERENCES public.contactos_cliente(id_contacto);


--
-- TOC entry 5192 (class 2606 OID 17040)
-- Name: entregas fk_entregas_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5193 (class 2606 OID 17045)
-- Name: entregas fk_entregas_id_usuario_entrega; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_usuario_entrega FOREIGN KEY (id_usuario_entrega) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5194 (class 2606 OID 17050)
-- Name: entregas fk_entregas_id_usuario_verificador; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.entregas
    ADD CONSTRAINT fk_entregas_id_usuario_verificador FOREIGN KEY (id_usuario_verificador) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5162 (class 2606 OID 16580)
-- Name: historial_propietarios fk_historial_cliente; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


--
-- TOC entry 5187 (class 2606 OID 17020)
-- Name: historial_estados_ot fk_historial_estados_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT fk_historial_estados_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5188 (class 2606 OID 17025)
-- Name: historial_estados_ot fk_historial_estados_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_estados_ot
    ADD CONSTRAINT fk_historial_estados_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5163 (class 2606 OID 16585)
-- Name: historial_propietarios fk_historial_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_usuario FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5164 (class 2606 OID 16575)
-- Name: historial_propietarios fk_historial_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.historial_propietarios
    ADD CONSTRAINT fk_historial_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- TOC entry 5179 (class 2606 OID 16980)
-- Name: ordenes_trabajo fk_ordenes_trabajo_id_cotizacion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT fk_ordenes_trabajo_id_cotizacion FOREIGN KEY (id_cotizacion) REFERENCES public.cotizaciones(id_cotizacion);


--
-- TOC entry 5180 (class 2606 OID 16985)
-- Name: ordenes_trabajo fk_ordenes_trabajo_id_usuario_creacion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.ordenes_trabajo
    ADD CONSTRAINT fk_ordenes_trabajo_id_usuario_creacion FOREIGN KEY (id_usuario_creacion) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5189 (class 2606 OID 17030)
-- Name: pagos fk_pagos_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT fk_pagos_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5190 (class 2606 OID 17035)
-- Name: pagos fk_pagos_id_usuario_registro; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pagos
    ADD CONSTRAINT fk_pagos_id_usuario_registro FOREIGN KEY (id_usuario_registro) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5185 (class 2606 OID 17010)
-- Name: pruebas_ot fk_pruebas_ot_id_orden; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT fk_pruebas_ot_id_orden FOREIGN KEY (id_orden) REFERENCES public.ordenes_trabajo(id_orden);


--
-- TOC entry 5186 (class 2606 OID 17015)
-- Name: pruebas_ot fk_pruebas_ot_id_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.pruebas_ot
    ADD CONSTRAINT fk_pruebas_ot_id_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5170 (class 2606 OID 16940)
-- Name: recepciones fk_recepciones_id_cliente_atencion; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_cliente_atencion FOREIGN KEY (id_cliente_atencion) REFERENCES public.clientes(id_cliente);


--
-- TOC entry 5171 (class 2606 OID 16945)
-- Name: recepciones fk_recepciones_id_recepcionista; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_recepcionista FOREIGN KEY (id_recepcionista) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5172 (class 2606 OID 16935)
-- Name: recepciones fk_recepciones_id_vehiculo; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.recepciones
    ADD CONSTRAINT fk_recepciones_id_vehiculo FOREIGN KEY (id_vehiculo) REFERENCES public.vehiculos(id_vehiculo);


--
-- TOC entry 5201 (class 2606 OID 17095)
-- Name: repuesto_proveedor fk_repuesto_proveedor_id_proveedor; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT fk_repuesto_proveedor_id_proveedor FOREIGN KEY (id_proveedor) REFERENCES public.proveedores(id_proveedor);


--
-- TOC entry 5202 (class 2606 OID 17090)
-- Name: repuesto_proveedor fk_repuesto_proveedor_id_repuesto; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.repuesto_proveedor
    ADD CONSTRAINT fk_repuesto_proveedor_id_repuesto FOREIGN KEY (id_repuesto) REFERENCES public.repuestos(id_repuesto);


--
-- TOC entry 5167 (class 2606 OID 16641)
-- Name: rol_permisos fk_rol_permisos_permiso; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT fk_rol_permisos_permiso FOREIGN KEY (id_permiso) REFERENCES public.permisos(id_permiso);


--
-- TOC entry 5168 (class 2606 OID 16636)
-- Name: rol_permisos fk_rol_permisos_rol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.rol_permisos
    ADD CONSTRAINT fk_rol_permisos_rol FOREIGN KEY (id_rol) REFERENCES public.roles(id_rol);


--
-- TOC entry 5165 (class 2606 OID 16624)
-- Name: usuario_roles fk_usuario_roles_rol; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT fk_usuario_roles_rol FOREIGN KEY (id_rol) REFERENCES public.roles(id_rol);


--
-- TOC entry 5166 (class 2606 OID 16619)
-- Name: usuario_roles fk_usuario_roles_usuario; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.usuario_roles
    ADD CONSTRAINT fk_usuario_roles_usuario FOREIGN KEY (id_usuario) REFERENCES public.usuarios(id_usuario);


--
-- TOC entry 5161 (class 2606 OID 16533)
-- Name: vehiculos fk_vehiculos_cliente; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.vehiculos
    ADD CONSTRAINT fk_vehiculos_cliente FOREIGN KEY (id_cliente) REFERENCES public.clientes(id_cliente);


-- Completed on 2026-10-26 21:46:04

--
-- PostgreSQL database dump complete
--

\unrestrict Lu9Bnerz4NmBJ4i1sLHJed93WIM8eE1Mngrd3BUaWAN38egpOcvmS5GZbhfmr9o

