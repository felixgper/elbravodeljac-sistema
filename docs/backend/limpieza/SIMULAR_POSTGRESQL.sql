-- Rehearsal ONLY on a restored QA copy. Never cleans development.
\set ON_ERROR_STOP on
BEGIN;
SET LOCAL lock_timeout='5s';
DO $$ BEGIN
 IF current_database() !~ '^jac_restore_[0-9]{14}$' THEN
  RAISE EXCEPTION 'Solo se admite una copia restaurada QA; desarrollo está bloqueado';
 END IF;
 IF (SELECT count(*) FROM pg_tables WHERE schemaname='public')<>28 OR
    (SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace)<>50 THEN
  RAISE EXCEPTION 'Esquema diferente: revisar plan antes de continuar';
 END IF;
END $$;
CREATE TEMP TABLE jac_protected_before(name text PRIMARY KEY, fingerprint text) ON COMMIT DROP;
DO $$ DECLARE tab text; hash text; BEGIN
 FOREACH tab IN ARRAY ARRAY['usuarios','empleados','roles','permisos','usuario_roles','rol_permisos','tipos_documento','servicios','repuestos','proveedores','repuesto_proveedor','unidades_auxilio'] LOOP
  EXECUTE format('SELECT md5(coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text)::text,'''')) FROM public.%I t',tab) INTO hash;
  INSERT INTO jac_protected_before VALUES(tab,hash);
 END LOOP;
END $$;
DO $$ DECLARE tab text; removed bigint; BEGIN
 -- Reverse dependency order. No CASCADE, identity reset or security writes.
 FOREACH tab IN ARRAY ARRAY['entregas','pagos','historial_estados_ot','pruebas_ot','avances_ot','asignaciones_ot','ordenes_trabajo','detalle_cotizacion','cotizaciones','diagnosticos','auxilios','recepciones','historial_propietarios','vehiculos','contactos_cliente','clientes'] LOOP
  EXECUTE format('DELETE FROM public.%I',tab);
  GET DIAGNOSTICS removed=ROW_COUNT;
  RAISE NOTICE 'SIMULACIÓN %: % filas',tab,removed;
 END LOOP;
END $$;
DO $$ DECLARE protected record; hash text; BEGIN
 FOR protected IN SELECT * FROM jac_protected_before LOOP
  EXECUTE format('SELECT md5(coalesce(jsonb_agg(to_jsonb(t) ORDER BY to_jsonb(t)::text)::text,'''')) FROM public.%I t',protected.name) INTO hash;
  IF hash IS DISTINCT FROM protected.fingerprint THEN RAISE EXCEPTION 'Se alteró información protegida: %',protected.name; END IF;
 END LOOP;
 IF (SELECT count(*) FROM pg_constraint WHERE contype='f' AND connamespace='public'::regnamespace)<>50 THEN RAISE EXCEPTION 'Cambió el esquema'; END IF;
END $$;
ROLLBACK;
SELECT 'Simulación revertida: no se eliminaron datos persistentes' AS resultado;
