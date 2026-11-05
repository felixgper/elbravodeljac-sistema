-- Additive, repeatable migration. No data deletion, no FK/table removal.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
CREATE EXTENSION IF NOT EXISTS unaccent;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE OR REPLACE FUNCTION public.jac_normalize(value text) RETURNS text
LANGUAGE sql IMMUTABLE PARALLEL SAFE AS
$$ SELECT lower(public.unaccent('public.unaccent', coalesce(value,''))) $$;
CREATE INDEX IF NOT EXISTS ix_clientes_busqueda ON public.clientes
USING gin (public.jac_normalize(nombre || ' ' || numero_documento) gin_trgm_ops);
ALTER TABLE public.detalle_cotizacion ADD COLUMN IF NOT EXISTS tipo_concepto varchar(10);
ALTER TABLE public.detalle_cotizacion ADD COLUMN IF NOT EXISTS tipo_repuesto varchar(12);
ALTER TABLE public.detalle_cotizacion ADD COLUMN IF NOT EXISTS descuento_unitario numeric(12,2) NOT NULL DEFAULT 0;
ALTER TABLE public.detalle_cotizacion DROP CONSTRAINT IF EXISTS ck_detalle_referencia;
ALTER TABLE public.detalle_cotizacion ADD CONSTRAINT ck_detalle_referencia CHECK (
 (id_servicio IS NOT NULL AND id_repuesto IS NULL AND (tipo_concepto IS NULL OR tipo_concepto='Servicio')) OR
 (id_repuesto IS NOT NULL AND id_servicio IS NULL AND (tipo_concepto IS NULL OR tipo_concepto='Repuesto')) OR
 (id_servicio IS NULL AND id_repuesto IS NULL AND tipo_concepto IS NOT NULL AND tipo_concepto IN ('Servicio','Repuesto') AND descripcion IS NOT NULL AND btrim(descripcion)<>'')
);
DO $$ BEGIN
 IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.detalle_cotizacion'::regclass AND conname='ck_detalle_descuento') THEN
  ALTER TABLE public.detalle_cotizacion ADD CONSTRAINT ck_detalle_descuento CHECK(descuento_unitario>=0 AND descuento_unitario<=precio_unitario);
 END IF;
 IF NOT EXISTS(SELECT 1 FROM pg_constraint WHERE conrelid='public.detalle_cotizacion'::regclass AND conname='ck_detalle_tipo_repuesto') THEN
  ALTER TABLE public.detalle_cotizacion ADD CONSTRAINT ck_detalle_tipo_repuesto CHECK(tipo_repuesto IS NULL OR (tipo_repuesto IN ('Original','Alternativo') AND coalesce(tipo_concepto,CASE WHEN id_servicio IS NOT NULL THEN 'Servicio' ELSE 'Repuesto' END)='Repuesto'));
 END IF;
END $$;
COMMIT;
