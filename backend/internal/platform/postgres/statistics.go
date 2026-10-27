package postgres

import (
	"context"
	"time"
)

// Payments are aggregated directly, never joined to quote details or assignments.
func (s *Store) Statistics(ctx context.Context, dashboard bool, start, end time.Time) (any, error) {
	cte := `WITH months AS (
 SELECT generate_series(date_trunc('month',$1::timestamptz AT TIME ZONE 'America/Lima'),date_trunc('month',($2::timestamptz-interval '1 microsecond') AT TIME ZONE 'America/Lima'),interval '1 month') AS bucket
), pay_month AS (
 SELECT date_trunc('month',fecha_pago AT TIME ZONE 'America/Lima') AS bucket,sum(monto) AS amount FROM pagos WHERE fecha_pago >= $1 AND fecha_pago < $2 GROUP BY 1
), income AS (
 SELECT m.bucket,coalesce(p.amount,0)::numeric(18,2) AS amount FROM months m LEFT JOIN pay_month p USING(bucket)
), services AS (
 SELECT s.id_servicio::text AS id,s.nombre AS label,count(*) AS count FROM ordenes_trabajo o JOIN detalle_cotizacion d USING(id_cotizacion) JOIN servicios s USING(id_servicio) WHERE o.estado='finalizada' AND ($3::boolean OR (o.fecha_fin >= $1 AND o.fecha_fin < $2)) GROUP BY s.id_servicio,s.nombre
), rescues AS (
 SELECT count(*) AS count FROM auxilios WHERE fecha >= $1 AND fecha < $2
), states AS (
 SELECT t.key,t.label,t.color,(SELECT count(*) FROM ordenes_trabajo o WHERE o.estado=t.state) AS count FROM (VALUES ('pending','Pendientes','#d89a2a','creada'),('repair','En reparación','#791c2b','en_reparacion'),('testing','En pruebas','#467bbb','en_pruebas'),('completed','Finalizadas','#24906d','finalizada'),('cancelled','Canceladas','#b85858','cancelada')) t(key,label,color,state)
), staff AS (
 SELECT u.id_usuario::text AS id,u.nombre AS label,count(DISTINCT o.id_orden) AS count FROM usuarios u LEFT JOIN asignaciones_ot a ON a.id_usuario=u.id_usuario AND a.activo AND a.rol_en_trabajo='MECANICO' LEFT JOIN ordenes_trabajo o ON o.id_orden=a.id_orden AND o.estado IN ('creada','en_reparacion','en_pruebas') WHERE EXISTS(SELECT 1 FROM usuario_roles ur JOIN roles r USING(id_rol) WHERE ur.id_usuario=u.id_usuario AND lower(r.nombre_rol) IN ('mecánico','mecanico','supervisor')) GROUP BY u.id_usuario,u.nombre ORDER BY count DESC,u.id_usuario LIMIT 4
) `
	projection := `jsonb_build_object(
 'monthlyIncome',(SELECT amount::text FROM income ORDER BY bucket DESC LIMIT 1),
 'monthlyIncomeCents',(SELECT (amount*100)::bigint::text FROM income ORDER BY bucket DESC LIMIT 1),
 'activeOrders',(SELECT count(*) FROM ordenes_trabajo WHERE estado IN ('creada','en_reparacion','en_pruebas')),
 'completedServices',(SELECT count(*) FROM ordenes_trabajo WHERE estado='finalizada'),
 'rescuesToday',(SELECT count(*) FROM auxilios WHERE (fecha AT TIME ZONE 'America/Lima')::date=(statement_timestamp() AT TIME ZONE 'America/Lima')::date),
 'income',(SELECT coalesce(jsonb_agg(jsonb_build_object('month',to_char(bucket,'YYYY-MM'),'amount',amount::text,'cents',(amount*100)::bigint::text) ORDER BY bucket),'[]'::jsonb) FROM income),
 'states',(SELECT jsonb_agg(to_jsonb(states)) FROM states),
 'mechanics',(SELECT coalesce(jsonb_agg(to_jsonb(staff)),'[]'::jsonb) FROM staff),
 'services',(SELECT coalesce(jsonb_agg(to_jsonb(t)),'[]'::jsonb) FROM (SELECT * FROM services ORDER BY count DESC,label LIMIT 4)t),
 'currentMonth',to_char(($2::timestamptz-interval '1 microsecond') AT TIME ZONE 'America/Lima','YYYY-MM'),
 'from',$1::timestamptz,'until',$2::timestamptz,
 'counts',jsonb_build_object('clients',(SELECT count(*) FROM clientes),'vehicles',(SELECT count(*) FROM vehiculos),'receptions',(SELECT count(*) FROM recepciones),'diagnoses',(SELECT count(*) FROM diagnosticos),'quotes',(SELECT count(*) FROM cotizaciones),'deliveries',(SELECT count(*) FROM entregas)))`
	if !dashboard {
		projection = `jsonb_build_object(
 'services',(SELECT coalesce(jsonb_agg(to_jsonb(t) ORDER BY t.label),'[]'::jsonb) FROM (SELECT * FROM services UNION ALL SELECT NULL::text,'Auxilio mecánico',count FROM rescues WHERE count>0)t),
 'months',(SELECT coalesce(jsonb_agg(jsonb_build_object('label',to_char(bucket,'YYYY-MM'),'amount',amount::text,'cents',(amount*100)::bigint::text) ORDER BY bucket),'[]'::jsonb) FROM income),
 'orders',(SELECT count(*) FROM ordenes_trabajo WHERE estado='finalizada' AND fecha_fin >= $1 AND fecha_fin < $2),
 'income',(SELECT coalesce(sum(amount),0)::numeric(18,2)::text FROM income),
 'incomeCents',(SELECT (coalesce(sum(amount),0)*100)::bigint::text FROM income),
 'from',$1::timestamptz,'until',$2::timestamptz)`
	}
	return readJSON(ctx, s.Pool, cte+"SELECT "+projection, start, end, dashboard)
}
