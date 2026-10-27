package postgres

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

func readJSON(ctx context.Context, q queryer, sql string, args ...any) (any, error) {
	var raw []byte
	err := q.QueryRow(ctx, sql, args...).Scan(&raw)
	if err != nil {
		return nil, translate(err)
	}
	return json.RawMessage(raw), nil
}

const receptionJSON = `jsonb_build_object('id',r.id_recepcion::text,'attentionType',r.tipo_atencion,'technicianId',r.id_mecanico_diagnostico::text,'technicianName',(SELECT u.nombre FROM usuarios u WHERE u.id_usuario=r.id_mecanico_diagnostico),'assignmentVersion',r.asignacion_version,'assignmentHistory',r.historial_asignaciones,'evaluationState',CASE WHEN r.tipo_atencion='Directo' THEN 'Servicio directo' WHEN r.id_mecanico_diagnostico IS NULL THEN 'Sin asignar' ELSE coalesce((SELECT d.estado FROM diagnosticos d WHERE d.id_recepcion=r.id_recepcion),'Pendiente de diagnóstico') END,'clientId',r.id_cliente_atencion::text,'clientName',(SELECT c.nombre FROM clientes c WHERE c.id_cliente=r.id_cliente_atencion),'vehicleId',r.id_vehiculo::text,'vehiclePlate',(SELECT cv.placa FROM vehiculos cv WHERE cv.id_vehiculo=r.id_vehiculo),'vehicleBrand',(SELECT cv.marca FROM vehiculos cv WHERE cv.id_vehiculo=r.id_vehiculo),'vehicleModel',(SELECT cv.modelo FROM vehiculos cv WHERE cv.id_vehiculo=r.id_vehiculo),'mileage',r.km_ingreso::text,'symptoms',r.sintomas,'createdAt',r.fecha_ingreso,'createdBy',r.id_recepcionista::text)`
const diagnosisJSON = `jsonb_build_object('id',d.id_diagnostico::text,'receptionId',d.id_recepcion::text,'findings',d.resumen,'evaluation',d.evaluacion,'recommendations',d.recomendaciones,'notes',d.observaciones,'status',d.estado,'version',d.version,'finishedAt',d.finalizado_en,'updatedAt',d.actualizado_en,'history',d.historial,'technicianName',(SELECT u.nombre FROM usuarios u WHERE u.id_usuario=d.id_tecnico),'technicianId',d.id_tecnico::text,'createdAt',d.fecha)`
const quoteJSON = `jsonb_build_object('id',q.id_cotizacion::text,'diagnosisId',q.id_diagnostico::text,'receptionId',q.id_recepcion::text,'status',CASE q.estado WHEN 'aprobada' THEN 'Aprobada' WHEN 'rechazada' THEN 'Rechazada' ELSE 'Pendiente' END,'total',q.total::text,'createdAt',q.fecha_emision,'approvedAt',q.aprobada_en,
 'items',coalesce((SELECT jsonb_agg(jsonb_build_object('id',l.id_detalle::text,'serviceId',l.id_servicio::text,'sparePartId',l.id_repuesto::text,'concept',l.descripcion,'type',coalesce(l.tipo_concepto,CASE WHEN l.id_servicio IS NULL THEN 'Repuesto' ELSE 'Servicio' END),'partKind',l.tipo_repuesto,'discount',l.descuento_unitario::text,'origin',l.origen_repuesto,'quantity',l.cantidad::text,'unitPrice',l.precio_unitario::text,'subtotal',l.subtotal::text) ORDER BY l.id_detalle) FROM detalle_cotizacion l WHERE l.id_cotizacion=q.id_cotizacion),'[]'::jsonb))`
const orderFrom = `ordenes_trabajo o JOIN cotizaciones q ON q.id_cotizacion=o.id_cotizacion LEFT JOIN diagnosticos d ON d.id_diagnostico=q.id_diagnostico JOIN recepciones r ON r.id_recepcion=q.id_recepcion JOIN vehiculos v ON v.id_vehiculo=r.id_vehiculo JOIN clientes cl ON cl.id_cliente=r.id_cliente_atencion`
const orderJSON = `jsonb_build_object('id',o.id_orden::text,'code',o.codigo_ot,'quoteId',o.id_cotizacion::text,'clientId',r.id_cliente_atencion::text,'clientName',cl.nombre,'vehicleId',r.id_vehiculo::text,'vehiclePlate',v.placa,'vehicleBrand',v.marca,'vehicleModel',v.modelo,'state',CASE o.estado WHEN 'creada' THEN 'Pendiente' WHEN 'en_reparacion' THEN 'En reparación' WHEN 'en_pruebas' THEN 'En pruebas' WHEN 'finalizada' THEN 'Finalizada' WHEN 'cancelada' THEN 'Cancelada' ELSE o.estado END,'createdAt',o.fecha_creacion,'startedAt',o.fecha_inicio,'finishedAt',o.fecha_fin,'createdBy',o.id_usuario_creacion::text,
 'assignments',coalesce((SELECT jsonb_agg(jsonb_build_object('id',a.id_asignacion::text,'userId',a.id_usuario::text,'role',a.rol_en_trabajo,'active',a.activo,'at',a.fecha_asignacion) ORDER BY a.id_asignacion) FROM asignaciones_ot a WHERE a.id_orden=o.id_orden),'[]'::jsonb),
 'advances',coalesce((SELECT jsonb_agg(jsonb_build_object('id',a.id_avance::text,'note',a.observacion,'by',a.id_usuario::text,'at',a.fecha) ORDER BY a.fecha,a.id_avance) FROM avances_ot a WHERE a.id_orden=o.id_orden),'[]'::jsonb),
 'tests',coalesce((SELECT jsonb_agg(jsonb_build_object('id',p.id_prueba::text,'test',p.tipo_prueba,'result',p.resultado,'note',p.observaciones,'by',p.id_usuario::text,'at',p.fecha) ORDER BY p.fecha,p.id_prueba) FROM pruebas_ot p WHERE p.id_orden=o.id_orden),'[]'::jsonb),
 'states',coalesce((SELECT jsonb_agg(jsonb_build_object('id',h.id_historial::text,'previous',h.estado_anterior,'state',h.estado_nuevo,'by',h.id_usuario::text,'at',h.fecha) ORDER BY h.fecha,h.id_historial) FROM historial_estados_ot h WHERE h.id_orden=o.id_orden),'[]'::jsonb),
 'delivered',EXISTS(SELECT 1 FROM entregas e WHERE e.id_orden=o.id_orden))`
const paymentJSON = `jsonb_build_object('id',p.id_pago::text,'orderId',p.id_orden::text,'amount',p.monto::text,'method',p.metodo_pago,'comprobante',p.comprobante,'createdAt',p.fecha_pago,'createdBy',p.id_usuario_registro::text)`
const deliveryJSON = `jsonb_build_object('id',e.id_entrega::text,'orderId',e.id_orden::text,'sheetReady',e.ficha_pdf IS NOT NULL,'kind',e.tipo_receptor,'contactId',e.id_contacto_receptor::text,'method',e.medio_verificacion,'verified',e.autorizacion_verificada,'name',e.nombre_receptor_snap,'documentType',e.tipo_doc_receptor_snap,'documentNumber',e.documento_receptor_snap,'phone',e.telefono_receptor_snap,'verifiedAt',e.fecha_verificacion,'verifiedBy',e.id_usuario_verificador::text,'createdAt',e.fecha_entrega,'createdBy',e.id_usuario_entrega::text)`

func orderVisible(u domain.User) bool {
	return u.Can("ordenes.asignar") || u.Can("ordenes.crear") || u.Can("pagos.ver") || u.Can("ordenes.eliminar")
}
func (s *Store) WorkshopList(ctx context.Context, resource string, f domain.WorkshopFilter, u domain.User) (any, error) {
	var expr, from, where, sort string
	args := []any{}
	switch resource {
	case "recepciones":
		expr, from, where, sort = receptionJSON, "recepciones r", "($1='' OR r.id_cliente_atencion::text=$1) AND (NOT $2 OR (r.tipo_atencion='Diagnostico' AND NOT EXISTS(SELECT 1 FROM diagnosticos d WHERE d.id_recepcion=r.id_recepcion)))", "r.id_recepcion"
		args = []any{f.ClientID, f.Pending}
	case "diagnosticos":
		expr, from, where, sort = diagnosisJSON, "diagnosticos d", "($1='' OR d.id_recepcion::text=$1)", "d.id_diagnostico"
		args = []any{f.ParentID}
	case "cotizaciones":
		expr, from, where, sort = quoteJSON, "cotizaciones q", "($1='' OR q.id_diagnostico::text=$1) AND ($2='' OR q.estado=$2) AND (NOT $3 OR NOT EXISTS(SELECT 1 FROM ordenes_trabajo o WHERE o.id_cotizacion=q.id_cotizacion))", "q.id_cotizacion"
		args = []any{f.ParentID, f.State, f.WithoutOrder}
	case "ordenes", "tareas":
		expr, from, where, sort = orderJSON, orderFrom, "($1='' OR o.estado=$1) AND (o.codigo_ot ILIKE $2 OR cl.nombre ILIKE $2 OR v.placa ILIKE $2) AND ($3 OR EXISTS(SELECT 1 FROM asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.id_usuario::text=$4 AND a.activo))", "o.id_orden"
		global := orderVisible(u)
		if resource == "tareas" {
			global = false
			where += ` AND o.estado IN ('creada','en_reparacion','en_pruebas')`
		}
		args = []any{f.State, "%" + f.Query + "%", global, u.ID}
	case "servicios":
		expr, from, where, sort = `jsonb_build_object('id',c.id_servicio::text,'name',c.nombre,'referencePrice',c.precio_referencia::text,'active',c.activo)`, "servicios c", "c.activo AND c.nombre ILIKE $1", "c.id_servicio"
		args = []any{"%" + f.Query + "%"}
	case "repuestos":
		expr, from, where, sort = `jsonb_build_object('id',c.id_repuesto::text,'name',c.nombre,'referencePrice',c.precio_referencia::text,'stock',c.stock::text,'active',c.activo)`, "repuestos c", "c.activo AND c.nombre ILIKE $1", "c.id_repuesto"
		args = []any{"%" + f.Query + "%"}
	case "mecanicos":
		expr, from, where, sort = `jsonb_build_object('id',t.id_usuario::text,'name',t.nombre,'active',t.activo,'roles',ARRAY(SELECT rr.nombre_rol FROM roles rr JOIN usuario_roles ur USING(id_rol) WHERE ur.id_usuario=t.id_usuario),'activeOrders',(SELECT count(DISTINCT a.id_orden) FROM asignaciones_ot a JOIN ordenes_trabajo o USING(id_orden) WHERE a.id_usuario=t.id_usuario AND a.activo AND o.estado IN ('creada','en_reparacion','en_pruebas')))`, "usuarios t", `t.activo AND EXISTS(SELECT 1 FROM usuario_roles ur JOIN roles rr USING(id_rol) WHERE ur.id_usuario=t.id_usuario AND lower(rr.nombre_rol) IN ('mecánico','mecanico','supervisor','auxiliar'))`, "t.id_usuario"
	case "asignaciones":
		expr, from, where, sort = `jsonb_build_object('id',a.id_asignacion::text,'orderId',a.id_orden::text,'userId',a.id_usuario::text,'role',a.rol_en_trabajo,'at',a.fecha_asignacion)`, "asignaciones_ot a", "a.activo AND ($1='' OR a.id_orden::text=$1) AND ($2='' OR a.id_usuario::text=$2)", "a.id_asignacion"
		args = []any{f.ParentID, f.ClientID}
	default:
		return nil, domain.ErrNotFound
	}
	if resource == "recepciones" {
		where += " AND " + receptionScope("$"+itoa(len(args)+1), "$"+itoa(len(args)+2))
		args = append(args, receptionGlobal(u), u.ID)
	}
	if resource == "diagnosticos" {
		where += " AND " + diagnosisScope("$"+itoa(len(args)+1), "$"+itoa(len(args)+2))
		args = append(args, diagnosisSupervisor(u), u.ID)
	}
	out := domain.List[json.RawMessage]{Data: []json.RawMessage{}, Meta: domain.Meta{Page: f.Number, PageSize: f.Size}}
	// Count and rows share one read snapshot, including nested JSON projections.
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+from+" WHERE "+where, args...).Scan(&out.Meta.Total); err != nil {
		return nil, translate(err)
	}
	query := "SELECT " + expr + " FROM " + from + " WHERE " + where + " ORDER BY " + sort + " DESC LIMIT $" + itoa(len(args)+1) + " OFFSET $" + itoa(len(args)+2)
	args = append(args, f.Size, (f.Number-1)*f.Size)
	rows, err := tx.Query(ctx, query, args...)
	if err != nil {
		return nil, translate(err)
	}
	for rows.Next() {
		var raw []byte
		if err = rows.Scan(&raw); err != nil {
			rows.Close()
			return nil, err
		}
		out.Data = append(out.Data, json.RawMessage(raw))
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, translate(err)
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) WorkshopItem(ctx context.Context, resource, id string, u domain.User) (any, error) {
	switch resource {
	case "recepciones":
		return readJSON(ctx, s.Pool, "SELECT "+receptionJSON+" FROM recepciones r WHERE r.id_recepcion=$1 AND "+receptionScope("$2", "$3"), id, receptionGlobal(u), u.ID)
	case "diagnosticos":
		return readJSON(ctx, s.Pool, "SELECT "+diagnosisJSON+" FROM diagnosticos d WHERE d.id_diagnostico=$1 AND "+diagnosisScope("$2", "$3"), id, diagnosisSupervisor(u), u.ID)
	case "cotizaciones":
		return readJSON(ctx, s.Pool, "SELECT "+quoteJSON+" FROM cotizaciones q WHERE q.id_cotizacion=$1", id)
	case "ordenes":
		return readJSON(ctx, s.Pool, "SELECT "+orderJSON+" FROM "+orderFrom+" WHERE o.id_orden=$1 AND ($2 OR EXISTS(SELECT 1 FROM asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.id_usuario::text=$3 AND a.activo))", id, orderVisible(u), u.ID)
	case "pagos":
		return readJSON(ctx, s.Pool, `SELECT jsonb_build_object('orderId',o.id_orden::text,'total',q.total::text,'paid',coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=o.id_orden),0)::numeric(14,2)::text,'balance',(q.total-coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=o.id_orden),0))::numeric(14,2)::text,'payments',coalesce((SELECT jsonb_agg(`+paymentJSON+` ORDER BY p.id_pago) FROM pagos p WHERE p.id_orden=o.id_orden),'[]'::jsonb)) FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) WHERE o.id_orden=$1`, id)
	case "entrega":
		return readJSON(ctx, s.Pool, "SELECT "+deliveryJSON+" FROM entregas e WHERE e.id_orden=$1", id)
	}
	return nil, domain.ErrNotFound
}

func receptionScope(global, actor string) string {
	return "(" + global + " OR r.id_mecanico_diagnostico::text=" + actor + " OR EXISTS(SELECT 1 FROM cotizaciones cq JOIN ordenes_trabajo co USING(id_cotizacion) JOIN asignaciones_ot ca USING(id_orden) WHERE cq.id_recepcion=r.id_recepcion AND ca.activo AND ca.id_usuario::text=" + actor + "))"
}
func diagnosisScope(global, actor string) string {
	return "(" + global + " OR d.id_tecnico::text=" + actor + " OR EXISTS(SELECT 1 FROM cotizaciones cq JOIN ordenes_trabajo co USING(id_cotizacion) JOIN asignaciones_ot ca USING(id_orden) WHERE cq.id_diagnostico=d.id_diagnostico AND ca.activo AND ca.id_usuario::text=" + actor + "))"
}
