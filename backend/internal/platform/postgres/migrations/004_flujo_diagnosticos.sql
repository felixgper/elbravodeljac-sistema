BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
DO $$ BEGIN IF NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='diagnosticos' AND column_name='estado') THEN
INSERT INTO permisos(codigo) VALUES ('diagnosticos.asignar'),('diagnosticos.supervisar') ON CONFLICT(codigo) DO NOTHING;
INSERT INTO rol_permisos(id_rol,id_permiso) SELECT r.id_rol,p.id_permiso FROM roles r CROSS JOIN permisos p WHERE r.nombre_rol IN ('Propietario','Administrador','Supervisor','Administrativo') AND p.codigo IN ('diagnosticos.ver','diagnosticos.asignar','diagnosticos.supervisar') ON CONFLICT DO NOTHING;
-- Administrative staff may prepare quotes; no user-role relations are replaced.
INSERT INTO rol_permisos(id_rol,id_permiso) SELECT r.id_rol,p.id_permiso FROM roles r CROSS JOIN permisos p WHERE r.nombre_rol='Administrativo' AND p.codigo IN ('cotizaciones.crear','cotizaciones.editar') ON CONFLICT DO NOTHING;
END IF; END $$;
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS id_mecanico_diagnostico bigint REFERENCES usuarios(id_usuario);
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS diagnostico_asignado_por bigint REFERENCES usuarios(id_usuario);
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS diagnostico_asignado_en timestamptz;
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS asignacion_version bigint NOT NULL DEFAULT 0;
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS historial_asignaciones jsonb NOT NULL DEFAULT '[]';
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS estado varchar(15) NOT NULL DEFAULT 'Finalizado' CHECK(estado IN ('Borrador','Finalizado'));
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS evaluacion text NOT NULL DEFAULT '';
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS recomendaciones text NOT NULL DEFAULT '';
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS observaciones text NOT NULL DEFAULT '';
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS version bigint NOT NULL DEFAULT 0;
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS finalizado_en timestamptz;
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS actualizado_en timestamptz;
ALTER TABLE diagnosticos ADD COLUMN IF NOT EXISTS historial jsonb NOT NULL DEFAULT '[]';
UPDATE diagnosticos SET finalizado_en=fecha,actualizado_en=fecha WHERE estado='Finalizado' AND finalizado_en IS NULL;
UPDATE recepciones r SET id_mecanico_diagnostico=d.id_tecnico,diagnostico_asignado_en=d.fecha FROM diagnosticos d WHERE d.id_recepcion=r.id_recepcion AND r.id_mecanico_diagnostico IS NULL;
ALTER TABLE diagnosticos ALTER COLUMN estado SET DEFAULT 'Borrador';
CREATE INDEX IF NOT EXISTS ix_recepciones_mecanico_diagnostico ON recepciones(id_mecanico_diagnostico,id_recepcion);
CREATE INDEX IF NOT EXISTS ix_diagnosticos_estado ON diagnosticos(estado,finalizado_en);
CREATE OR REPLACE FUNCTION jac_validar_diagnostico() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE tipo text; tecnico bigint;
BEGIN
 SELECT tipo_atencion,id_mecanico_diagnostico INTO tipo,tecnico FROM recepciones WHERE id_recepcion=NEW.id_recepcion FOR SHARE;
 IF tipo='Directo' THEN RAISE EXCEPTION 'Servicio directo sin diagnostico' USING ERRCODE='23514'; END IF;
 IF tecnico IS NULL OR tecnico IS DISTINCT FROM NEW.id_tecnico THEN RAISE EXCEPTION 'El diagnostico debe corresponder al mecanico asignado' USING ERRCODE='23514'; END IF;
 IF TG_OP='UPDATE' AND OLD.estado='Finalizado' AND NEW IS DISTINCT FROM OLD THEN RAISE EXCEPTION 'El diagnostico finalizado es inmutable' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_diagnostico_tecnico BEFORE INSERT OR UPDATE ON diagnosticos FOR EACH ROW EXECUTE FUNCTION jac_validar_diagnostico();
CREATE OR REPLACE FUNCTION jac_validar_origen_cotizacion() RETURNS trigger LANGUAGE plpgsql AS $$
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
COMMIT;
