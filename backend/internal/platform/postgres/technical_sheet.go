package postgres

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
)

const sheetSnapshotSQL = `SELECT jsonb_build_object('workshop',(SELECT to_jsonb(t) FROM datos_taller_ficha t),'version','original-v1','code',o.codigo_ot,'client',cl.nombre,'document',cl.numero_documento,'phone',coalesce(cl.telefono,''),'address',coalesce(cl.direccion,''),'email',coalesce(cl.correo,''),'plate',v.placa,'vehicle',v.marca||' '||v.modelo,'mileage',r.km_ingreso::text,'symptoms',r.sintomas,'in',r.fecha_ingreso,'out',e.fecha_entrega,'mechanic',coalesce((SELECT u.nombre FROM asignaciones_ot a JOIN usuarios u USING(id_usuario) WHERE a.id_orden=o.id_orden AND a.activo AND a.rol_en_trabajo='MECANICO' ORDER BY a.id_asignacion DESC LIMIT 1),'No registrado'),'total',q.total::text,'lines',coalesce((SELECT jsonb_agg(jsonb_build_object('id',dc.id_detalle::text,'concept',dc.descripcion,'quantity',dc.cantidad::text,'unitPrice',dc.precio_unitario::text,'discount',dc.descuento_unitario::text,'subtotal',dc.subtotal::text) ORDER BY dc.id_detalle) FROM detalle_cotizacion dc WHERE dc.id_cotizacion=q.id_cotizacion),'[]'::jsonb),'advances',coalesce((SELECT jsonb_agg(a.observacion ORDER BY a.id_avance) FROM avances_ot a WHERE a.id_orden=o.id_orden),'[]'::jsonb),'control',o.control_tecnico) FROM ` + orderFrom + ` JOIN entregas e ON e.id_orden=o.id_orden WHERE o.id_orden=$1`

func (s *Store) Inspection(ctx context.Context, id string, u domain.User) (any, error) {
	return readJSON(ctx, s.Pool, `SELECT jsonb_build_object('control',o.control_tecnico,'version',o.control_version,'authorId',o.control_autor::text,'updatedAt',o.control_fecha,'readOnly',o.estado IN ('finalizada','cancelada') OR EXISTS(SELECT 1 FROM entregas e WHERE e.id_orden=o.id_orden),'canEdit',($5 AND ($2 OR EXISTS(SELECT 1 FROM asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.activo AND a.rol_en_trabajo='MECANICO' AND a.id_usuario::text=$3)))) FROM ordenes_trabajo o WHERE o.id_orden=$1 AND ($4 OR EXISTS(SELECT 1 FROM asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.activo AND a.id_usuario::text=$3))`, id, u.Can("ordenes.asignar"), u.ID, orderVisible(u), u.Can("ordenes.editar"))
}
func (s *Store) SaveInspection(ctx context.Context, id string, v technicalsheet.Control, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, domain.Invalid("control", err.Error())
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var state string
	var version int64
	var allowed bool
	err = tx.QueryRow(ctx, `SELECT estado,control_version,($2 OR EXISTS(SELECT 1 FROM asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.activo AND a.rol_en_trabajo='MECANICO' AND a.id_usuario::text=$3)) FROM ordenes_trabajo o WHERE id_orden=$1 FOR UPDATE`, id, u.Can("ordenes.asignar"), u.ID).Scan(&state, &version, &allowed)
	if err != nil {
		return nil, translate(err)
	}
	if !allowed {
		return nil, domain.ErrForbidden
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if closed || state == "finalizada" || state == "cancelada" || version != v.Version {
		return nil, domain.ErrConflict
	}
	if len(v.PerformedItems) > 0 {
		var count int
		err = tx.QueryRow(ctx, `SELECT count(*) FROM detalle_cotizacion dc JOIN ordenes_trabajo o USING(id_cotizacion) WHERE o.id_orden=$1 AND dc.id_detalle::text=ANY($2)`, id, v.PerformedItems).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count != len(v.PerformedItems) {
			return nil, domain.Invalid("performedItems", "Selecciona trabajos de la cotización autorizada sin duplicarlos.")
		}
	}
	v.Version = version + 1
	raw, _ := json.Marshal(v)
	_, err = tx.Exec(ctx, `UPDATE ordenes_trabajo SET control_tecnico=$2,control_version=$3,control_autor=$4,control_fecha=clock_timestamp() WHERE id_orden=$1`, id, raw, v.Version, u.ID)
	if err != nil {
		return nil, translate(err)
	}
	if err = tx.Commit(ctx); err != nil {
		return nil, translate(err)
	}
	return s.Inspection(ctx, id, u)
}
func (s *Store) SheetPDF(ctx context.Context, id string, u domain.User) ([]byte, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var snapshot, pdf []byte
	err = tx.QueryRow(ctx, `SELECT e.ficha_snapshot,e.ficha_pdf FROM entregas e JOIN ordenes_trabajo o USING(id_orden) JOIN cotizaciones q USING(id_cotizacion) WHERE e.id_orden=$1 AND o.estado='finalizada' AND q.estado='aprobada' AND q.total=coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=o.id_orden),0) FOR UPDATE OF e`, id).Scan(&snapshot, &pdf)
	if err != nil {
		return nil, translate(err)
	}
	if len(pdf) > 0 {
		return pdf, nil
	}
	// Legacy deliveries are reconstructed once using available data, then frozen.
	if len(snapshot) == 0 {
		if err = tx.QueryRow(ctx, sheetSnapshotSQL, id).Scan(&snapshot); err != nil {
			return nil, translate(err)
		}
	}
	pdf, err = technicalsheet.Render(snapshot)
	if err != nil {
		return nil, domain.Invalid("ficha", err.Error())
	}
	_, err = tx.Exec(ctx, `UPDATE entregas SET ficha_snapshot=$2,ficha_pdf=$3 WHERE id_orden=$1`, id, snapshot, pdf)
	if err != nil {
		return nil, translate(err)
	}
	return pdf, translate(tx.Commit(ctx))
}
