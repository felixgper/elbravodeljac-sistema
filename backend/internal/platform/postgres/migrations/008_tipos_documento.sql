-- Incremental and transactional; snapshots of completed deliveries are untouched.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
CREATE TABLE IF NOT EXISTS public.tipos_documento (
 id_tipo_documento bigserial PRIMARY KEY,
 codigo varchar(20) NOT NULL UNIQUE,
 nombre varchar(150) NOT NULL,
 activo boolean NOT NULL DEFAULT true
);
INSERT INTO tipos_documento(codigo,nombre) VALUES
 ('DNI','Documento Nacional de Identidad'),('RUC','Registro Único de Contribuyentes'),
 ('CE','Carné de Extranjería'),('PASAPORTE','Pasaporte') ON CONFLICT(codigo) DO NOTHING;
CREATE OR REPLACE FUNCTION public.jac_documento_codigo(value text) RETURNS text
LANGUAGE sql IMMUTABLE AS $$
 SELECT CASE regexp_replace(jac_normalize(btrim(value)),'[. ]','','g')
 WHEN 'dni' THEN 'DNI' WHEN 'documentonacionaldeidentidad' THEN 'DNI'
 WHEN 'ruc' THEN 'RUC' WHEN 'registrounicodecontribuyentes' THEN 'RUC'
 WHEN 'ce' THEN 'CE' WHEN 'carnedeextranjeria' THEN 'CE' WHEN 'carnetdeextranjeria' THEN 'CE'
 WHEN 'pasaporte' THEN 'PASAPORTE' ELSE NULL END
$$;
-- Unknown historical labels get their own inactive catalogue entry. No guessing.
INSERT INTO tipos_documento(codigo,nombre,activo)
SELECT DISTINCT 'HIST_'||substr(md5(raw),1,15),
 CASE WHEN btrim(raw)='' THEN 'Tipo histórico sin código' ELSE raw END,false
FROM (SELECT tipo_documento AS raw FROM clientes UNION SELECT tipo_documento FROM contactos_cliente) x
WHERE jac_documento_codigo(raw) IS NULL ON CONFLICT(codigo) DO NOTHING;
ALTER TABLE clientes ADD COLUMN IF NOT EXISTS id_tipo_documento bigint REFERENCES tipos_documento(id_tipo_documento) ON DELETE RESTRICT;
ALTER TABLE contactos_cliente ADD COLUMN IF NOT EXISTS id_tipo_documento bigint REFERENCES tipos_documento(id_tipo_documento) ON DELETE RESTRICT;
UPDATE clientes c SET id_tipo_documento=d.id_tipo_documento FROM tipos_documento d
WHERE c.id_tipo_documento IS NULL AND d.codigo=coalesce(jac_documento_codigo(c.tipo_documento),'HIST_'||substr(md5(c.tipo_documento),1,15));
UPDATE contactos_cliente c SET id_tipo_documento=d.id_tipo_documento FROM tipos_documento d
WHERE c.id_tipo_documento IS NULL AND d.codigo=coalesce(jac_documento_codigo(c.tipo_documento),'HIST_'||substr(md5(c.tipo_documento),1,15));
-- An alias collision aborts the entire migration; numbers are never rewritten.
DO $$ BEGIN
 IF EXISTS(SELECT 1 FROM clientes GROUP BY id_tipo_documento,numero_documento HAVING count(*)>1) THEN
  RAISE EXCEPTION 'Tipos equivalentes con números duplicados: revisar antes de migrar; no se fusionaron clientes' USING ERRCODE='23514';
 END IF;
END $$;
CREATE UNIQUE INDEX IF NOT EXISTS uq_clientes_tipo_catalogo_numero ON clientes(id_tipo_documento,numero_documento);
ALTER TABLE clientes ALTER COLUMN id_tipo_documento SET NOT NULL;
ALTER TABLE contactos_cliente ALTER COLUMN id_tipo_documento SET NOT NULL;
UPDATE clientes c SET tipo_documento=CASE d.codigo WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE d.codigo END
FROM tipos_documento d WHERE c.id_tipo_documento=d.id_tipo_documento AND d.codigo IN ('DNI','RUC','CE','PASAPORTE') AND c.tipo_documento IS DISTINCT FROM CASE d.codigo WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE d.codigo END;
UPDATE contactos_cliente c SET tipo_documento=CASE d.codigo WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE d.codigo END
FROM tipos_documento d WHERE c.id_tipo_documento=d.id_tipo_documento AND d.codigo IN ('DNI','RUC','CE','PASAPORTE') AND c.tipo_documento IS DISTINCT FROM CASE d.codigo WHEN 'PASAPORTE' THEN 'Pasaporte' ELSE d.codigo END;
CREATE OR REPLACE FUNCTION jac_persona_tipo_documento() RETURNS trigger LANGUAGE plpgsql AS $$
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
CREATE OR REPLACE TRIGGER jac_cliente_documento BEFORE INSERT OR UPDATE OF tipo_documento,id_tipo_documento,numero_documento ON clientes FOR EACH ROW EXECUTE FUNCTION jac_persona_tipo_documento();
CREATE OR REPLACE TRIGGER jac_contacto_documento BEFORE INSERT OR UPDATE OF tipo_documento,id_tipo_documento,numero_documento ON contactos_cliente FOR EACH ROW EXECUTE FUNCTION jac_persona_tipo_documento();
COMMIT;
