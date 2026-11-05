BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
DO $$ BEGIN IF NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name='pagos' AND column_name='request_id') THEN
 INSERT INTO permisos(codigo) VALUES('pagos.rectificar') ON CONFLICT DO NOTHING;
 INSERT INTO rol_permisos(id_rol,id_permiso) SELECT r.id_rol,p.id_permiso FROM roles r CROSS JOIN permisos p WHERE r.nombre_rol IN ('Propietario','Administrador') AND p.codigo='pagos.rectificar' ON CONFLICT DO NOTHING;
END IF; END $$;
ALTER TABLE pagos ADD COLUMN IF NOT EXISTS request_id varchar(100);
ALTER TABLE pagos ADD COLUMN IF NOT EXISTS revertido_en timestamptz;
ALTER TABLE pagos ADD COLUMN IF NOT EXISTS rectificacion jsonb;
CREATE UNIQUE INDEX IF NOT EXISTS ix_pago_request ON pagos(id_orden,request_id) WHERE request_id IS NOT NULL;
ALTER TABLE entregas ADD COLUMN IF NOT EXISTS receptor_real jsonb;
ALTER TABLE entregas ADD COLUMN IF NOT EXISTS conformidad jsonb;
CREATE OR REPLACE FUNCTION jac_importe_orden(oid bigint) RETURNS numeric LANGUAGE sql STABLE AS $$
 SELECT CASE WHEN o.flujo_version=0 OR o.control_tecnico->'reconciliation' IS NULL OR o.control_tecnico->'reconciliation'='null'::jsonb THEN q.total
 ELSE coalesce((SELECT sum(d.subtotal) FROM detalle_cotizacion d WHERE d.id_cotizacion=q.id_cotizacion AND coalesce(o.control_tecnico->'execution'->d.id_detalle::text->>'status','Pendiente')<>'No realizado'),0) END FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) WHERE o.id_orden=oid
$$;
CREATE OR REPLACE VIEW jac_movimientos_pagos AS
 SELECT id_pago,id_orden,monto,fecha_pago AS fecha FROM pagos
 UNION ALL SELECT id_pago,id_orden,-monto,revertido_en FROM pagos WHERE revertido_en IS NOT NULL;
CREATE OR REPLACE FUNCTION fn_bravo_entrega_validada() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE total numeric; paid numeric; state text;
BEGIN
 IF NEW.autorizacion_verificada IS DISTINCT FROM TRUE OR NEW.id_usuario_verificador IS NULL OR NEW.medio_verificacion IS NULL THEN RAISE EXCEPTION 'Verifica identidad y autorización antes de entregar' USING ERRCODE='23514'; END IF;
 SELECT jac_importe_orden(id_orden),estado INTO total,state FROM ordenes_trabajo WHERE id_orden=NEW.id_orden FOR UPDATE;
 IF state<>'finalizada' THEN RAISE EXCEPTION 'La orden debe estar finalizada' USING ERRCODE='23514'; END IF;
 SELECT coalesce(sum(monto),0) INTO paid FROM pagos WHERE id_orden=NEW.id_orden AND revertido_en IS NULL;
 IF total IS NULL OR total<>paid THEN RAISE EXCEPTION 'La entrega requiere saldo neto cero' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
COMMIT;
