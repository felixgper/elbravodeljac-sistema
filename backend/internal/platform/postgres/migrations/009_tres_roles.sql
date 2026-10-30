-- Preserve membership IDs/permissions. Unclassified accounts remain pending.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
DO $migration$ BEGIN
IF NOT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_schema='public' AND table_name='roles' AND column_name='activo') THEN
IF EXISTS(SELECT 1 FROM roles WHERE nombre_rol IN ('Gerente','Administrador (legado)')) THEN
 RAISE EXCEPTION 'Ya existe un rol de destino; revisar correspondencias antes de migrar';
END IF;
ALTER TABLE roles ADD COLUMN activo boolean NOT NULL DEFAULT true;
INSERT INTO permisos(codigo) VALUES ('inicio.ver'),('clientes.ver'),('clientes.crear'),('clientes.editar'),('clientes.eliminar'),('vehiculos.ver'),('vehiculos.crear'),('vehiculos.editar'),('vehiculos.eliminar'),('recepciones.ver'),('recepciones.crear'),('diagnosticos.ver'),('diagnosticos.crear'),('diagnosticos.asignar'),('diagnosticos.supervisar'),('cotizaciones.ver'),('cotizaciones.crear'),('cotizaciones.editar'),('ordenes.ver'),('ordenes.crear'),('ordenes.editar'),('ordenes.eliminar'),('ordenes.asignar'),('auxilio.ver'),('auxilio.crear'),('pagos.ver'),('pagos.registrar'),('pagos.entregar'),('pagos.rectificar'),('mecanicos.ver'),('proveedores.ver'),('proveedores.crear'),('proveedores.editar'),('proveedores.eliminar'),('reportes.ver'),('usuarios.ver'),('usuarios.crear'),('usuarios.editar'),('usuarios.eliminar'),('usuarios.roles') ON CONFLICT(codigo) DO NOTHING;

-- One-time transition, detected by the new policy's canonical Gerente row.
DO $$ DECLARE rid bigint; BEGIN
 IF NOT EXISTS(SELECT 1 FROM roles WHERE nombre_rol='Gerente') THEN
  -- Owner -> manager retains role ID, all memberships and effective permissions.
  UPDATE roles SET nombre_rol='Gerente' WHERE nombre_rol='Propietario';
  IF NOT EXISTS(SELECT 1 FROM roles WHERE nombre_rol='Gerente') THEN
   INSERT INTO roles(nombre_rol) VALUES('Gerente');
  END IF;
  -- Old Administrator had full access: do NOT overwrite its permissions or
  -- silently turn every account into a new operational Administrator.
  UPDATE roles SET nombre_rol='Administrador (legado)',activo=false WHERE nombre_rol='Administrador';
  UPDATE roles SET activo=false WHERE nombre_rol NOT IN ('Gerente','Mecánico');
 END IF;
END $$;
INSERT INTO roles(nombre_rol) VALUES('Administrador'),('Mecánico') ON CONFLICT(nombre_rol) DO NOTHING;
UPDATE roles SET activo=false WHERE nombre_rol IN ('Propietario','Supervisor','Administrativo','Auxiliar');
-- Provision ONLY empty canonical roles. Repeated migration never restores a
-- permission that an authorized manager intentionally removed later.
DO $$ DECLARE r record; p text; codes text[]; BEGIN
 FOR r IN SELECT id_rol,nombre_rol FROM roles WHERE nombre_rol IN ('Gerente','Administrador','Mecánico') LOOP
  IF NOT EXISTS(SELECT 1 FROM rol_permisos WHERE id_rol=r.id_rol) THEN
   IF r.nombre_rol='Gerente' THEN
    SELECT array_agg(codigo) INTO codes FROM permisos;
   ELSIF r.nombre_rol='Administrador' THEN
    SELECT array_agg(codigo) INTO codes FROM permisos WHERE split_part(codigo,'.',1) NOT IN ('usuarios') AND codigo<>'pagos.rectificar';
   ELSE
    codes=ARRAY['inicio.ver','vehiculos.ver','recepciones.ver','diagnosticos.ver','diagnosticos.crear','ordenes.ver','ordenes.editar','auxilio.ver'];
   END IF;
   FOREACH p IN ARRAY coalesce(codes,ARRAY[]::text[]) LOOP
    INSERT INTO rol_permisos(id_rol,id_permiso) SELECT r.id_rol,id_permiso FROM permisos WHERE codigo=p ON CONFLICT DO NOTHING;
   END LOOP;
  END IF;
 END LOOP;
END $$;
END IF;
END $migration$;
-- Historical AUXILIAR assignments remain readable and are never updated here.
CREATE OR REPLACE FUNCTION jac_no_nuevo_auxiliar() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.rol_en_trabajo='AUXILIAR' AND (TG_OP='INSERT' OR NEW.id_orden IS DISTINCT FROM OLD.id_orden OR NEW.id_usuario IS DISTINCT FROM OLD.id_usuario OR (NEW.activo AND NOT OLD.activo)) THEN
  RAISE EXCEPTION 'No se crean ni reactivan asignaciones auxiliares; se conserva el historial' USING ERRCODE='23514';
 END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_sin_auxiliar_nuevo BEFORE INSERT OR UPDATE ON asignaciones_ot FOR EACH ROW EXECUTE FUNCTION jac_no_nuevo_auxiliar();
COMMIT;
