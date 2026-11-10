package sqlserver

import (
	"context"
	"strings"
	"time"
)

// Join payment movements only to calendar buckets. Joining to work items or staff duplicates revenue.
func (s *Store) Statistics(ctx context.Context, dashboard bool, start, end time.Time) (any, error) {
	cte := `WITH limits AS (
 SELECT CAST(DATETRUNC(month,@p1 AT TIME ZONE 'SA Pacific Standard Time') AS date) AS first,
 CAST(DATETRUNC(month,DATEADD(microsecond,-1,@p2) AT TIME ZONE 'SA Pacific Standard Time') AS date) AS last
), buckets AS (
 SELECT first AS bucket,last FROM limits
 UNION ALL SELECT DATEADD(month,1,bucket),last FROM buckets WHERE bucket<last
), pay_bucket AS (
 SELECT CAST(DATETRUNC(month,fecha AT TIME ZONE 'SA Pacific Standard Time') AS date) AS bucket,SUM(monto) AS amount
 FROM dbo.jac_movimientos_pagos WHERE fecha>=@p1 AND fecha<@p2
 GROUP BY CAST(DATETRUNC(month,fecha AT TIME ZONE 'SA Pacific Standard Time') AS date)
), income AS (
 SELECT b.bucket,CAST(COALESCE(p.amount,0) AS decimal(18,2)) AS amount FROM buckets b LEFT JOIN pay_bucket p ON p.bucket=b.bucket
), services AS (
 SELECT CONVERT(varchar(20),s.id_servicio) AS id,d.descripcion AS label,SUM(d.cantidad) AS count
 FROM dbo.ordenes_trabajo o JOIN dbo.detalle_cotizacion d ON d.id_cotizacion=o.id_cotizacion
 LEFT JOIN dbo.servicios s ON s.id_servicio=d.id_servicio
 WHERE COALESCE(d.tipo_concepto,CASE WHEN d.id_servicio IS NOT NULL THEN N'Servicio' ELSE N'Repuesto' END)=N'Servicio'
 AND (JSON_VALUE(o.control_tecnico,CONCAT(N'$.execution."',d.id_detalle,N'".status'))=N'Realizado'
 OR EXISTS(SELECT 1 FROM OPENJSON(o.control_tecnico,N'$.performedItems') j WHERE j.value=CONVERT(varchar(20),d.id_detalle)))
 AND o.estado=N'finalizada' AND o.fecha_fin>=@p1 AND o.fecha_fin<@p2 GROUP BY s.id_servicio,d.descripcion
), rescues AS (
 SELECT COUNT(*) AS count FROM dbo.auxilios WHERE fecha>=@p1 AND fecha<@p2
), states AS (
 SELECT t.pos,t.[key],t.label,t.color,
 (SELECT COUNT(*) FROM dbo.ordenes_trabajo o WHERE o.estado=t.state AND
 (o.estado IN(N'creada',N'en_reparacion',N'en_pruebas') OR (o.fecha_fin>=@p1 AND o.fecha_fin<@p2))) AS count
 FROM (VALUES(1,N'pending',N'Pendientes',N'#d89a2a',N'creada'),(2,N'repair',N'En reparación',N'#791c2b',N'en_reparacion'),
 (3,N'testing',N'En pruebas',N'#467bbb',N'en_pruebas'),(4,N'completed',N'Finalizadas',N'#24906d',N'finalizada'),
 (5,N'cancelled',N'Canceladas',N'#b85858',N'cancelada')) t(pos,[key],label,color,state)
), staff AS (
 SELECT TOP (4) u.id_usuario, u.nombre AS label,COUNT(DISTINCT o.id_orden) AS count
 FROM dbo.usuarios u LEFT JOIN dbo.asignaciones_ot a ON a.id_usuario=u.id_usuario AND a.activo=1 AND a.rol_en_trabajo=N'MECANICO'
 LEFT JOIN dbo.ordenes_trabajo o ON o.id_orden=a.id_orden AND o.estado IN(N'creada',N'en_reparacion',N'en_pruebas')
 WHERE EXISTS(SELECT 1 FROM dbo.usuario_roles ur JOIN dbo.roles r ON r.id_rol=ur.id_rol WHERE ur.id_usuario=u.id_usuario
 AND r.nombre_rol COLLATE Latin1_General_100_CI_AS_SC IN(N'mecánico',N'mecanico',N'supervisor'))
 GROUP BY u.id_usuario,u.nombre ORDER BY count DESC,u.id_usuario
) `
	projection := `JSON_OBJECT(
'monthlyIncome':(SELECT CONVERT(varchar(32),CAST(COALESCE(SUM(amount),0) AS decimal(18,2))) FROM income),
'monthlyIncomeCents':(SELECT CONVERT(varchar(32),CAST(COALESCE(SUM(amount),0)*100 AS bigint)) FROM income),
'activeOrders':(SELECT COUNT(*) FROM dbo.ordenes_trabajo WHERE estado IN(N'creada',N'en_reparacion',N'en_pruebas')),
'completedServices':(SELECT COALESCE(SUM(count),0) FROM services),'rescuesToday':(SELECT count FROM rescues),
'income':JSON_QUERY((SELECT CONVERT(char(7),bucket,23) AS month,CONVERT(varchar(32),amount) AS amount,
 CONVERT(varchar(32),CAST(amount*100 AS bigint)) AS cents FROM income ORDER BY bucket FOR JSON PATH)),
'states':JSON_QUERY((SELECT [key],label,color,count FROM states ORDER BY pos FOR JSON PATH)),
'mechanics':JSON_QUERY((SELECT CONVERT(varchar(20),id_usuario) AS id,label,count FROM staff ORDER BY count DESC,id_usuario FOR JSON PATH)),
'services':JSON_QUERY((SELECT TOP (4) id,label,count FROM services ORDER BY count DESC,label FOR JSON PATH, INCLUDE_NULL_VALUES)),
'currentMonth':(SELECT CONVERT(char(7),last,23) FROM limits),
'dashboard':CONVERT(bit,@p3),'from':@p1,'until':@p2,
'counts':JSON_QUERY(JSON_OBJECT(
 'clients':(SELECT COUNT(DISTINCT id_cliente_atencion) FROM dbo.recepciones WHERE fecha_ingreso>=@p1 AND fecha_ingreso<@p2),
 'vehicles':(SELECT COUNT(DISTINCT id_vehiculo) FROM dbo.recepciones WHERE fecha_ingreso>=@p1 AND fecha_ingreso<@p2),
 'receptions':(SELECT COUNT(*) FROM dbo.recepciones WHERE fecha_ingreso>=@p1 AND fecha_ingreso<@p2),
 'diagnoses':(SELECT COUNT(*) FROM dbo.diagnosticos WHERE fecha>=@p1 AND fecha<@p2),
 'quotes':(SELECT COUNT(*) FROM dbo.cotizaciones WHERE fecha_emision>=@p1 AND fecha_emision<@p2),
 'deliveries':(SELECT COUNT(*) FROM dbo.entregas WHERE fecha_entrega>=@p1 AND fecha_entrega<@p2))))`
	if !dashboard {
		projection = `JSON_OBJECT(
'services':JSON_QUERY((SELECT id,label,count FROM (
 SELECT id,label,count FROM services UNION ALL SELECT NULL,N'Auxilio mecánico',count FROM rescues WHERE count>0
 ) x ORDER BY label FOR JSON PATH, INCLUDE_NULL_VALUES)),
'months':JSON_QUERY((SELECT CONVERT(char(7),bucket,23) AS label,CONVERT(varchar(32),amount) AS amount,
 CONVERT(varchar(32),CAST(amount*100 AS bigint)) AS cents FROM income ORDER BY bucket FOR JSON PATH)),
'orders':(SELECT COUNT(*) FROM dbo.ordenes_trabajo WHERE estado=N'finalizada' AND fecha_fin>=@p1 AND fecha_fin<@p2),
'income':(SELECT CONVERT(varchar(32),CAST(COALESCE(SUM(amount),0) AS decimal(18,2))) FROM income),
'incomeCents':(SELECT CONVERT(varchar(32),CAST(COALESCE(SUM(amount),0)*100 AS bigint)) FROM income),
'dashboard':CONVERT(bit,@p3),'from':@p1,'until':@p2)`
	}
	if end.Sub(start) <= 31*24*time.Hour {
		cte = strings.ReplaceAll(cte, "(month,", "(day,")
		projection = strings.ReplaceAll(projection, "char(7),bucket", "char(10),bucket")
	}
	return readJSON(ctx, s.db, cte+"SELECT "+projection+" OPTION (MAXRECURSION 366)", start, end, dashboard)
}
