-- One database catalogue without adding tables or changing existing vehicles/FK.
-- Base brands are available even when no vehicle has been registered for them.
BEGIN;
SELECT pg_advisory_xact_lock(hashtextextended('jac:migrations:v1',0));
CREATE OR REPLACE VIEW public.marcas_vehiculos AS
SELECT nombre FROM (
 SELECT DISTINCT ON (public.jac_normalize(btrim(nombre))) btrim(nombre)::varchar(40) AS nombre,
        prioridad
 FROM (
  SELECT nombre,0 AS prioridad FROM (VALUES
   ('JAC'),
   ('Toyota'),
   ('Nissan'),
   ('Hyundai'),
   ('Kia'),
   ('Chevrolet'),
   ('Ford'),
   ('Volkswagen'),
   ('Suzuki'),
   ('Mitsubishi'),
   ('Mazda'),
   ('Renault'),
   ('Chery'),
   ('Changan'),
   ('Haval'),
   ('BYD'),
   ('MG'),
   ('Great Wall'),
   ('BMW'),
   ('Mercedes-Benz'),
   ('Audi'),
   ('Volvo'),
   ('Peugeot'),
   ('Citroën'),
   ('BAIC'),
   ('Jinbei'),
   ('DFM'),
   ('Brilliance'),
   ('Changhe'),
   ('Mahindra'),
   ('SsangYong'),
   ('Daewoo'),
   ('Land Rover'),
   ('Honda')
  ) base(nombre)
  UNION ALL
  SELECT marca,1 FROM public.vehiculos WHERE btrim(coalesce(marca,''))<>''
 ) catalogo
 ORDER BY public.jac_normalize(btrim(nombre)),prioridad,nombre
) unicas;
COMMENT ON VIEW public.marcas_vehiculos IS '34 marcas aprobadas y marcas adicionales persistidas en vehiculos; lectura única para las APIs, sin modificar registros existentes';
COMMIT;
