package sqlserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

const receptionFields = `
'id':CONVERT(varchar(20),r.id_recepcion), 'attentionType':r.tipo_atencion,
'technicianId':CONVERT(varchar(20),r.id_mecanico_diagnostico),
'technicianName':(SELECT nombre FROM dbo.usuarios WHERE id_usuario=r.id_mecanico_diagnostico),
'assignmentVersion':r.asignacion_version,'assignmentHistory':JSON_QUERY(r.historial_asignaciones),
'evaluationState':CASE WHEN r.tipo_atencion=N'Directo' THEN N'Servicio directo'
 WHEN r.id_mecanico_diagnostico IS NULL THEN N'Sin asignar' ELSE COALESCE((SELECT estado FROM dbo.diagnosticos WHERE id_recepcion=r.id_recepcion),N'Pendiente de diagnóstico') END,
'clientId':CONVERT(varchar(20),r.id_cliente_atencion),'clientName':(SELECT nombre FROM dbo.clientes WHERE id_cliente=r.id_cliente_atencion),
'vehicleId':CONVERT(varchar(20),r.id_vehiculo),'vehiclePlate':(SELECT placa FROM dbo.vehiculos WHERE id_vehiculo=r.id_vehiculo),
'vehicleBrand':(SELECT marca FROM dbo.vehiculos WHERE id_vehiculo=r.id_vehiculo),'vehicleModel':(SELECT modelo FROM dbo.vehiculos WHERE id_vehiculo=r.id_vehiculo),
'mileage':CONVERT(varchar(32),r.km_ingreso),'symptoms':r.sintomas,'createdAt':r.fecha_ingreso,'createdBy':CONVERT(varchar(20),r.id_recepcionista)`
const receptionJSON = "JSON_OBJECT(" + receptionFields + ")"

const diagnosisJSON = `JSON_OBJECT('id':CONVERT(varchar(20),d.id_diagnostico),'receptionId':CONVERT(varchar(20),d.id_recepcion),
'findings':d.resumen,'evaluation':d.evaluacion,'recommendations':d.recomendaciones,'notes':d.observaciones,
'status':d.estado,'version':d.version,'finishedAt':d.finalizado_en,'updatedAt':d.actualizado_en,'history':JSON_QUERY(d.historial),
'technicianName':(SELECT nombre FROM dbo.usuarios WHERE id_usuario=d.id_tecnico),'technicianId':CONVERT(varchar(20),d.id_tecnico),'createdAt':d.fecha)`

const quoteJSON = `JSON_OBJECT('id':CONVERT(varchar(20),q.id_cotizacion),
'clientName':(SELECT c.nombre FROM dbo.recepciones cr JOIN dbo.clientes c ON c.id_cliente=cr.id_cliente_atencion WHERE cr.id_recepcion=q.id_recepcion),
'diagnosisId':CONVERT(varchar(20),q.id_diagnostico),'receptionId':CONVERT(varchar(20),q.id_recepcion),
'orderId':(SELECT CONVERT(varchar(20),id_orden) FROM dbo.ordenes_trabajo WHERE id_cotizacion=q.id_cotizacion),
'orderCode':(SELECT codigo_ot FROM dbo.ordenes_trabajo WHERE id_cotizacion=q.id_cotizacion),
'status':CASE q.estado WHEN N'aprobada' THEN N'Aprobada' WHEN N'rechazada' THEN N'Rechazada' ELSE N'Pendiente' END,
'total':CONVERT(varchar(32),q.total),'createdAt':q.fecha_emision,'approvedAt':q.aprobada_en,
'version':q.version,'approval':JSON_QUERY(q.aprobacion),'revisionHistory':JSON_QUERY(q.historial_revision),
'parentQuoteId':CONVERT(varchar(20),q.id_cotizacion_base),
'items':JSON_QUERY((SELECT CONVERT(varchar(20),l.id_detalle) AS id,CONVERT(varchar(20),l.id_servicio) AS serviceId,
 CONVERT(varchar(20),l.id_repuesto) AS sparePartId,l.descripcion AS concept,
 COALESCE(l.tipo_concepto,CASE WHEN l.id_servicio IS NULL THEN N'Repuesto' ELSE N'Servicio' END) AS type,
 l.tipo_repuesto AS partKind,CONVERT(varchar(32),l.descuento_unitario) AS discount,l.origen_repuesto AS origin,
 CONVERT(varchar(32),l.cantidad) AS quantity,CONVERT(varchar(32),l.precio_unitario) AS unitPrice,
 CONVERT(varchar(32),l.subtotal) AS subtotal FROM dbo.detalle_cotizacion l
 WHERE l.id_cotizacion=q.id_cotizacion ORDER BY l.id_detalle FOR JSON PATH, INCLUDE_NULL_VALUES)))`

const orderFrom = `dbo.ordenes_trabajo o JOIN dbo.cotizaciones q ON q.id_cotizacion=o.id_cotizacion
LEFT JOIN dbo.diagnosticos d ON d.id_diagnostico=q.id_diagnostico
JOIN dbo.recepciones r ON r.id_recepcion=q.id_recepcion JOIN dbo.vehiculos v ON v.id_vehiculo=r.id_vehiculo
JOIN dbo.clientes cl ON cl.id_cliente=r.id_cliente_atencion`

const orderJSON = `JSON_OBJECT('id':CONVERT(varchar(20),o.id_orden),'code':o.codigo_ot,'quoteId':CONVERT(varchar(20),o.id_cotizacion),
'clientId':CONVERT(varchar(20),r.id_cliente_atencion),'clientName':cl.nombre,'vehicleId':CONVERT(varchar(20),r.id_vehiculo),
'vehiclePlate':v.placa,'vehicleBrand':v.marca,'vehicleModel':v.modelo,'amountDue':CONVERT(varchar(42),dbo.jac_importe_orden(o.id_orden)),
'state':CASE o.estado WHEN N'creada' THEN N'Pendiente' WHEN N'en_reparacion' THEN N'En reparación'
 WHEN N'en_pruebas' THEN N'En pruebas' WHEN N'finalizada' THEN N'Finalizada' WHEN N'cancelada' THEN N'Cancelada' ELSE o.estado END,
'createdAt':o.fecha_creacion,'startedAt':o.fecha_inicio,'finishedAt':o.fecha_fin,'createdBy':CONVERT(varchar(20),o.id_usuario_creacion),
'assignments':JSON_QUERY((SELECT CONVERT(varchar(20),a.id_asignacion) AS id,CONVERT(varchar(20),a.id_usuario) AS userId,
 (SELECT nombre FROM dbo.usuarios WHERE id_usuario=a.id_usuario) AS userName,a.rol_en_trabajo AS role,a.activo AS active,a.fecha_asignacion AS [at]
 FROM dbo.asignaciones_ot a WHERE a.id_orden=o.id_orden ORDER BY a.id_asignacion FOR JSON PATH, INCLUDE_NULL_VALUES)),
'advances':JSON_QUERY((SELECT CONVERT(varchar(20),a.id_avance) AS id,a.observacion AS note,CONVERT(varchar(20),a.id_usuario) AS [by],a.fecha AS [at]
 FROM dbo.avances_ot a WHERE a.id_orden=o.id_orden ORDER BY a.fecha,a.id_avance FOR JSON PATH, INCLUDE_NULL_VALUES)),
'tests':JSON_QUERY((SELECT CONVERT(varchar(20),p.id_prueba) AS id,p.tipo_prueba AS test,p.resultado AS result,p.observaciones AS note,
 CONVERT(varchar(20),p.id_usuario) AS [by],p.fecha AS [at] FROM dbo.pruebas_ot p
 WHERE p.id_orden=o.id_orden ORDER BY p.fecha,p.id_prueba FOR JSON PATH, INCLUDE_NULL_VALUES)),
'states':JSON_QUERY((SELECT CONVERT(varchar(20),h.id_historial) AS id,h.estado_anterior AS previous,h.estado_nuevo AS state,
 CONVERT(varchar(20),h.id_usuario) AS [by],h.fecha AS [at] FROM dbo.historial_estados_ot h
 WHERE h.id_orden=o.id_orden ORDER BY h.fecha,h.id_historial FOR JSON PATH, INCLUDE_NULL_VALUES)),
'delivered':CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.entregas e WHERE e.id_orden=o.id_orden) THEN 1 ELSE 0 END))`

const paymentColumns = `CONVERT(varchar(20),p.id_pago) AS id,CONVERT(varchar(20),p.id_orden) AS orderId,
CONVERT(varchar(32),p.monto) AS amount,p.metodo_pago AS method,COALESCE(p.comprobante,N'') AS comprobante,
(SELECT nombre FROM dbo.usuarios WHERE id_usuario=p.id_usuario_registro) AS createdByName,p.request_id AS requestId,
p.revertido_en AS reversedAt,JSON_QUERY(p.rectificacion) AS correction,p.fecha_pago AS createdAt,
CONVERT(varchar(20),p.id_usuario_registro) AS createdBy`
const paymentJSON = "(SELECT " + paymentColumns + " FOR JSON PATH, INCLUDE_NULL_VALUES, WITHOUT_ARRAY_WRAPPER)"
const deliveryJSON = `JSON_OBJECT('id':CONVERT(varchar(20),e.id_entrega),'orderId':CONVERT(varchar(20),e.id_orden),
'sheetReady':CONVERT(bit,CASE WHEN e.ficha_pdf IS NOT NULL THEN 1 ELSE 0 END),
'recipientKind':JSON_VALUE(e.receptor_real,N'$.kind'),'receiver':JSON_QUERY(e.receptor_real),'conformity':JSON_QUERY(e.conformidad),
'kind':e.tipo_receptor,'contactId':CONVERT(varchar(20),e.id_contacto_receptor),'method':e.medio_verificacion,
'verified':e.autorizacion_verificada,'name':e.nombre_receptor_snap,'documentType':e.tipo_doc_receptor_snap,
'documentNumber':e.documento_receptor_snap,'phone':e.telefono_receptor_snap,'verifiedAt':e.fecha_verificacion,
'verifiedBy':CONVERT(varchar(20),e.id_usuario_verificador),'createdAt':e.fecha_entrega,'createdBy':CONVERT(varchar(20),e.id_usuario_entrega))`

func orderVisible(u domain.User) bool {
	return u.Can("ordenes.asignar") || u.Can("ordenes.crear") || u.Can("pagos.ver") || u.Can("ordenes.eliminar")
}
func receptionGlobal(u domain.User) bool {
	return diagnosisSupervisor(u) || u.Can("recepciones.crear") || u.Can("cotizaciones.crear")
}
func receptionScope(global, actor string) string {
	return scope(global, actor, "r.id_mecanico_diagnostico", "cq.id_recepcion=r.id_recepcion")
}
func diagnosisScope(global, actor string) string {
	return scope(global, actor, "d.id_tecnico", "cq.id_diagnostico=d.id_diagnostico")
}
func scope(global, actor, owner, link string) string {
	return fmt.Sprintf(`(%s=1 OR %s=TRY_CONVERT(bigint,%s) OR EXISTS(
SELECT 1 FROM dbo.cotizaciones cq JOIN dbo.ordenes_trabajo co ON co.id_cotizacion=cq.id_cotizacion
JOIN dbo.asignaciones_ot ca ON ca.id_orden=co.id_orden WHERE %s AND ca.activo=1 AND ca.id_usuario=TRY_CONVERT(bigint,%s)))`, global, owner, actor, link, actor)
}
func param(n int) string { return fmt.Sprintf("@p%d", n) }

func (s *Store) jsonPage(ctx context.Context, p domain.Page, expr, from, where, sort string, args ...any) (any, error) {
	out := domain.List[json.RawMessage]{Data: []json.RawMessage{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return nil, translate(err)
	}
	defer tx.Rollback()
	if err = tx.QueryRowContext(ctx, "SELECT COUNT(*) FROM "+from+" WHERE "+where, args...).Scan(&out.Meta.Total); err != nil {
		return nil, translate(err)
	}
	query := "SELECT " + expr + " FROM " + from + " WHERE " + where + " ORDER BY " + sort + " OFFSET " + param(len(args)+1) + " ROWS FETCH NEXT " + param(len(args)+2) + " ROWS ONLY"
	rows, err := tx.QueryContext(ctx, query, append(args, (p.Number-1)*p.Size, p.Size)...)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			return nil, translate(err)
		}
		out.Data = append(out.Data, json.RawMessage(raw))
	}
	if err = rows.Err(); err != nil {
		return nil, translate(err)
	}
	return out, translate(tx.Commit())
}

func (s *Store) WorkshopItem(ctx context.Context, resource, id string, u domain.User) (any, error) {
	switch resource {
	case "recepciones":
		return readJSON(ctx, s.db, "SELECT "+receptionJSON+" FROM dbo.recepciones r WHERE r.id_recepcion=@p1 AND "+receptionScope("@p2", "@p3"), id, receptionGlobal(u), u.ID)
	case "diagnosticos":
		return readJSON(ctx, s.db, "SELECT "+diagnosisJSON+" FROM dbo.diagnosticos d WHERE d.id_diagnostico=@p1 AND "+diagnosisScope("@p2", "@p3"), id, diagnosisSupervisor(u), u.ID)
	case "cotizaciones":
		return readJSON(ctx, s.db, "SELECT "+quoteJSON+" FROM dbo.cotizaciones q WHERE q.id_cotizacion=@p1", id)
	case "ordenes":
		return readJSON(ctx, s.db, "SELECT "+orderJSON+" FROM "+orderFrom+` WHERE o.id_orden=@p1 AND
(@p2=1 OR EXISTS(SELECT 1 FROM dbo.asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.id_usuario=TRY_CONVERT(bigint,@p3) AND a.activo=1))`, id, orderVisible(u), u.ID)
	case "pagos":
		return readJSON(ctx, s.db, `SELECT JSON_OBJECT('orderId':CONVERT(varchar(20),o.id_orden),
'total':CONVERT(varchar(42),dbo.jac_importe_orden(o.id_orden)),'paid':CONVERT(varchar(32),CAST(x.paid AS decimal(14,2))),
'balance':CONVERT(varchar(32),CAST(CAST(dbo.jac_importe_orden(o.id_orden) AS decimal(14,2))-CAST(x.paid AS decimal(14,2)) AS decimal(14,2))),
'payments':JSON_QUERY(COALESCE((SELECT `+paymentColumns+` FROM dbo.pagos p WHERE p.id_orden=o.id_orden ORDER BY p.id_pago FOR JSON PATH, INCLUDE_NULL_VALUES),N'[]')))
FROM dbo.ordenes_trabajo o CROSS APPLY(SELECT COALESCE(SUM(monto),0) AS paid FROM dbo.pagos WHERE id_orden=o.id_orden AND revertido_en IS NULL)x
WHERE o.id_orden=@p1`, id)
	case "entrega":
		return readJSON(ctx, s.db, "SELECT "+deliveryJSON+" FROM dbo.entregas e WHERE e.id_orden=@p1", id)
	default:
		return nil, domain.ErrNotFound
	}
}

func (s *Store) WorkshopList(ctx context.Context, resource string, f domain.WorkshopFilter, u domain.User) (any, error) {
	var expr, from, where, sort string
	var args []any
	like := "%" + f.Query + "%"
	switch resource {
	case "recepciones":
		expr, from, sort = receptionJSON, "dbo.recepciones r", "r.id_recepcion DESC"
		where = `(@p1=N'' OR r.id_cliente_atencion=TRY_CONVERT(bigint,@p1))
AND (@p2=0 OR (r.tipo_atencion=N'Diagnostico' AND NOT EXISTS(SELECT 1 FROM dbo.diagnosticos d WHERE d.id_recepcion=r.id_recepcion)))`
		args = []any{f.ClientID, f.Pending}
	case "atenciones":
		expr = "JSON_OBJECT(" + receptionFields + `,'clientDocumentType':c.tipo_documento,'clientDocumentNumber':c.numero_documento,
'diagnosis':JSON_QUERY(CASE WHEN d.id_diagnostico IS NULL THEN NULL ELSE ` + diagnosisJSON + ` END),
'quoteId':(SELECT TOP (1) CONVERT(varchar(20),cq.id_cotizacion) FROM dbo.cotizaciones cq WHERE cq.id_recepcion=r.id_recepcion ORDER BY cq.id_cotizacion))`
		from = `dbo.recepciones r JOIN dbo.clientes c ON c.id_cliente=r.id_cliente_atencion
JOIN dbo.vehiculos v ON v.id_vehiculo=r.id_vehiculo LEFT JOIN dbo.diagnosticos d ON d.id_recepcion=r.id_recepcion`
		where = `(r.tipo_atencion=N'Directo' OR d.estado=N'Finalizado') AND (@p1=N'' OR r.id_recepcion=TRY_CONVERT(bigint,@p1))
AND CONCAT(c.nombre,N' ',c.numero_documento,N' ',v.placa,N' ',r.id_recepcion) COLLATE Latin1_General_100_CI_AI_SC LIKE @p2`
		sort = "r.id_recepcion DESC"
		args = []any{f.ParentID, like}
	case "diagnosticos":
		expr, from, where, sort = diagnosisJSON, "dbo.diagnosticos d", "(@p1=N'' OR d.id_recepcion=TRY_CONVERT(bigint,@p1))", "d.id_diagnostico DESC"
		args = []any{f.ParentID}
	case "cotizaciones":
		expr, from, sort = quoteJSON, "dbo.cotizaciones q", "q.id_cotizacion DESC"
		where = `(@p1=N'' OR q.id_diagnostico=TRY_CONVERT(bigint,@p1)) AND (@p2=N'' OR q.estado=@p2)
AND (@p3=0 OR (q.aprobacion IS NOT NULL AND NOT EXISTS(SELECT 1 FROM dbo.ordenes_trabajo o WHERE o.id_cotizacion=q.id_cotizacion)))
AND EXISTS(SELECT 1 FROM dbo.recepciones sr JOIN dbo.clientes sc ON sc.id_cliente=sr.id_cliente_atencion
JOIN dbo.vehiculos sv ON sv.id_vehiculo=sr.id_vehiculo WHERE sr.id_recepcion=q.id_recepcion
AND CONCAT(sc.nombre,N' ',sc.numero_documento,N' ',sv.placa,N' ',q.id_cotizacion) COLLATE Latin1_General_100_CI_AI_SC LIKE @p4)`
		args = []any{f.ParentID, f.State, f.WithoutOrder, like}
	case "ordenes", "tareas":
		expr, from, sort = orderJSON, orderFrom, "o.id_orden DESC"
		where = `(@p1=N'' OR o.estado=@p1)
AND CONCAT(o.codigo_ot,N' ',cl.nombre,N' ',cl.numero_documento,N' ',v.placa,N' ',o.id_orden) COLLATE Latin1_General_100_CI_AI_SC LIKE @p2
AND (@p3=1 OR EXISTS(SELECT 1 FROM dbo.asignaciones_ot a WHERE a.id_orden=o.id_orden
AND a.id_usuario=TRY_CONVERT(bigint,@p4) AND a.activo=1 AND a.rol_en_trabajo=N'MECANICO'))`
		global := orderVisible(u)
		if resource == "tareas" {
			global = false
			where += " AND o.estado IN(N'creada',N'en_reparacion',N'en_pruebas')"
		}
		args = []any{f.State, like, global, u.ID}
	case "servicios":
		expr, from, where, sort = `JSON_OBJECT('id':CONVERT(varchar(20),c.id_servicio),'name':c.nombre,
'referencePrice':CONVERT(varchar(32),c.precio_referencia),'active':c.activo)`, "dbo.servicios c",
			"c.activo=1 AND c.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1", "c.id_servicio DESC"
		args = []any{like}
	case "repuestos":
		expr, from, where, sort = `JSON_OBJECT('id':CONVERT(varchar(20),c.id_repuesto),'name':c.nombre,
'referencePrice':CONVERT(varchar(32),c.precio_referencia),'stock':CONVERT(varchar(32),c.stock),'active':c.activo)`, "dbo.repuestos c",
			"c.activo=1 AND c.nombre COLLATE Latin1_General_100_CI_AS_SC LIKE @p1", "c.id_repuesto DESC"
		args = []any{like}
	case "mecanicos":
		expr = `JSON_OBJECT('id':CONVERT(varchar(20),t.id_usuario),'name':t.nombre,'active':t.activo,
'roles':JSON_QUERY(COALESCE((SELECT JSON_ARRAYAGG(rr.nombre_rol ORDER BY rr.id_rol) FROM dbo.roles rr
 JOIN dbo.usuario_roles ur ON ur.id_rol=rr.id_rol WHERE ur.id_usuario=t.id_usuario),N'[]')),
'activeOrders':(SELECT COUNT(DISTINCT a.id_orden) FROM dbo.asignaciones_ot a JOIN dbo.ordenes_trabajo o ON o.id_orden=a.id_orden
 WHERE a.id_usuario=t.id_usuario AND a.activo=1 AND o.estado IN(N'creada',N'en_reparacion',N'en_pruebas')))`
		from, where, sort = "dbo.usuarios t", `t.activo=1 AND NOT EXISTS(SELECT 1 FROM dbo.empleados e WHERE e.id_empleado=t.id_empleado AND e.activo=0)
AND EXISTS(SELECT 1 FROM dbo.usuario_roles ur JOIN dbo.roles rr ON rr.id_rol=ur.id_rol WHERE ur.id_usuario=t.id_usuario
AND rr.activo=1 AND rr.nombre_rol COLLATE Latin1_General_100_CI_AS_SC IN(N'mecánico',N'mecanico'))`, "t.id_usuario DESC"
	case "asignaciones":
		expr = `JSON_OBJECT('id':CONVERT(varchar(20),a.id_asignacion),'orderId':CONVERT(varchar(20),a.id_orden),'userId':CONVERT(varchar(20),a.id_usuario),
'userName':(SELECT au.nombre FROM dbo.usuarios au WHERE au.id_usuario=a.id_usuario),'role':a.rol_en_trabajo,'at':a.fecha_asignacion)`
		from, where, sort = "dbo.asignaciones_ot a", `a.activo=1 AND (@p1=N'' OR a.id_orden=TRY_CONVERT(bigint,@p1))
AND (@p2=N'' OR a.id_usuario=TRY_CONVERT(bigint,@p2))`, "a.id_asignacion DESC"
		args = []any{f.ParentID, f.ClientID}
	default:
		return nil, domain.ErrNotFound
	}
	if resource == "recepciones" || resource == "atenciones" {
		where += " AND " + receptionScope(param(len(args)+1), param(len(args)+2))
		args = append(args, receptionGlobal(u) || (resource == "atenciones" && u.Can("cotizaciones.ver")), u.ID)
	}
	if resource == "diagnosticos" {
		where += " AND " + diagnosisScope(param(len(args)+1), param(len(args)+2))
		args = append(args, diagnosisSupervisor(u), u.ID)
	}
	return s.jsonPage(ctx, f.Page, expr, from, where, sort, args...)
}
