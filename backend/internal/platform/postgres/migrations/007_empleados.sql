-- Additive V20 implementation: no inferred people/jobs or changes to old IDs.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
CREATE TABLE IF NOT EXISTS public.empleados (
 id_empleado bigserial PRIMARY KEY,
 nombre varchar(150) NOT NULL CHECK (btrim(nombre)<>''),
 tipo_empleado varchar(40) NOT NULL CHECK (btrim(tipo_empleado)<>''),
 activo boolean NOT NULL DEFAULT true
);
ALTER TABLE public.usuarios ADD COLUMN IF NOT EXISTS id_empleado bigint REFERENCES public.empleados(id_empleado) ON DELETE RESTRICT;
CREATE UNIQUE INDEX IF NOT EXISTS uq_usuarios_empleado ON public.usuarios(id_empleado);
-- NULL remains valid for legacy/technical accounts; never synthesize employees.
-- Preserve usuarios.nombre for existing APIs, snapshots and actor references.
CREATE OR REPLACE FUNCTION public.jac_usuario_nombre_laboral() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.id_empleado IS NOT NULL THEN
  SELECT nombre INTO NEW.nombre FROM public.empleados WHERE id_empleado=NEW.id_empleado FOR SHARE;
 END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_usuario_nombre_laboral BEFORE INSERT OR UPDATE OF id_empleado,nombre ON public.usuarios FOR EACH ROW EXECUTE FUNCTION public.jac_usuario_nombre_laboral();
CREATE OR REPLACE FUNCTION public.jac_empleado_nombre_compatible() RETURNS trigger LANGUAGE plpgsql AS $$
BEGIN
 IF NEW.nombre IS DISTINCT FROM OLD.nombre THEN
  UPDATE public.usuarios SET nombre=NEW.nombre WHERE id_empleado=NEW.id_empleado;
 END IF;
 RETURN NEW;
END $$;
CREATE OR REPLACE TRIGGER jac_empleado_nombre_compatible AFTER UPDATE OF nombre ON public.empleados FOR EACH ROW EXECUTE FUNCTION public.jac_empleado_nombre_compatible();
COMMIT;
