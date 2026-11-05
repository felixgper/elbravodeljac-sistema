-- Additive evolution: existing 26 tables/42 FK retained; two new references.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
ALTER TABLE recepciones ADD COLUMN IF NOT EXISTS tipo_atencion varchar(15) NOT NULL DEFAULT 'Diagnostico' CHECK(tipo_atencion IN ('Directo','Diagnostico'));
ALTER TABLE cotizaciones ADD COLUMN IF NOT EXISTS id_recepcion bigint REFERENCES recepciones(id_recepcion);
UPDATE cotizaciones q SET id_recepcion=d.id_recepcion FROM diagnosticos d WHERE q.id_diagnostico=d.id_diagnostico AND q.id_recepcion IS NULL;
ALTER TABLE cotizaciones ALTER COLUMN id_recepcion SET NOT NULL;
ALTER TABLE cotizaciones ALTER COLUMN id_diagnostico DROP NOT NULL;
CREATE INDEX IF NOT EXISTS ix_cotizaciones_recepcion ON cotizaciones(id_recepcion);
ALTER TABLE ordenes_trabajo ADD COLUMN IF NOT EXISTS control_tecnico jsonb NOT NULL DEFAULT '{}';
ALTER TABLE ordenes_trabajo ADD COLUMN IF NOT EXISTS control_version bigint NOT NULL DEFAULT 0;
ALTER TABLE ordenes_trabajo ADD COLUMN IF NOT EXISTS control_autor bigint REFERENCES usuarios(id_usuario);
ALTER TABLE ordenes_trabajo ADD COLUMN IF NOT EXISTS control_fecha timestamptz;
ALTER TABLE entregas ADD COLUMN IF NOT EXISTS ficha_snapshot jsonb;
ALTER TABLE entregas ADD COLUMN IF NOT EXISTS ficha_pdf bytea;
CREATE OR REPLACE FUNCTION jac_validar_origen_cotizacion() RETURNS trigger LANGUAGE plpgsql AS $$
DECLARE tipo text; recepcion_diagnostico bigint;
BEGIN
 -- Compatibility for existing SQL callers; no synthetic diagnosis is inserted.
 IF NEW.id_recepcion IS NULL AND NEW.id_diagnostico IS NOT NULL THEN
  SELECT id_recepcion INTO NEW.id_recepcion FROM diagnosticos WHERE id_diagnostico=NEW.id_diagnostico;
 END IF;
 SELECT tipo_atencion INTO tipo FROM recepciones WHERE id_recepcion=NEW.id_recepcion FOR SHARE;
 IF tipo IS NULL THEN RAISE EXCEPTION 'Recepcion inexistente' USING ERRCODE='23514'; END IF;
 IF tipo='Directo' AND NEW.id_diagnostico IS NOT NULL THEN RAISE EXCEPTION 'Servicio directo no requiere diagnostico' USING ERRCODE='23514'; END IF;
 IF tipo='Diagnostico' THEN
  SELECT id_recepcion INTO recepcion_diagnostico FROM diagnosticos WHERE id_diagnostico=NEW.id_diagnostico;
  IF recepcion_diagnostico IS DISTINCT FROM NEW.id_recepcion THEN RAISE EXCEPTION 'Diagnostico obligatorio y vinculado a recepcion' USING ERRCODE='23514'; END IF;
 END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_origen_cotizacion BEFORE INSERT OR UPDATE OF id_recepcion,id_diagnostico ON cotizaciones FOR EACH ROW EXECUTE FUNCTION jac_validar_origen_cotizacion();
CREATE OR REPLACE FUNCTION jac_validar_diagnostico() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF (SELECT tipo_atencion FROM recepciones WHERE id_recepcion=NEW.id_recepcion FOR SHARE)='Directo' THEN RAISE EXCEPTION 'No crear diagnosticos para servicio directo' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_diagnostico_tecnico BEFORE INSERT OR UPDATE OF id_recepcion ON diagnosticos FOR EACH ROW EXECUTE FUNCTION jac_validar_diagnostico();
CREATE OR REPLACE FUNCTION jac_proteger_tipo_atencion() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF OLD.tipo_atencion IS DISTINCT FROM NEW.tipo_atencion AND (EXISTS(SELECT 1 FROM diagnosticos WHERE id_recepcion=OLD.id_recepcion) OR EXISTS(SELECT 1 FROM cotizaciones WHERE id_recepcion=OLD.id_recepcion)) THEN RAISE EXCEPTION 'No cambiar tipo de una atencion con historial' USING ERRCODE='23514'; END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_tipo_atencion BEFORE UPDATE OF tipo_atencion ON recepciones FOR EACH ROW EXECUTE FUNCTION jac_proteger_tipo_atencion();
CREATE OR REPLACE VIEW public.datos_taller_ficha AS SELECT 'EL BRAVO DEL JAC'::text AS name,'10400848209'::text AS ruc,'Calle Urano 381, Los Olivos'::text AS address;
COMMIT;
