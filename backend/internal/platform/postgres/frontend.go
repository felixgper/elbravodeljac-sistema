package postgres

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func (s *Store) FrontendOptions(ctx context.Context, module, resource string, p domain.Page, u domain.User) (any, error) {
	// Staff updating work sees only entities belonging to their assigned orders.
	// Reception/diagnosis/quotation creation and dispatch need selectable records.
	scoped := (module == "ordenes" || module == "auxilio" || module == "mecanicos") && !orderVisible(u) && !u.Can("auxilio.crear") && !u.Can("mecanicos.ver")
	if resource == "recepciones" || resource == "diagnosticos" {
		return s.WorkshopList(ctx, resource, domain.WorkshopFilter{Page: p}, u)
	}
	var expr, from, where, sort string
	switch resource {
	case "clientes":
		expr = `jsonb_build_object('id',c.id_cliente::text,'name',c.nombre,'documentType',c.tipo_documento,'documentNumber',c.numero_documento)`
		from = "clientes c"
		sort = "c.id_cliente"
		where = "NOT $1 OR c.id_cliente IN (SELECT r.id_cliente_atencion FROM " + orderFrom + " JOIN asignaciones_ot a ON a.id_orden=o.id_orden WHERE a.activo AND a.id_usuario::text=$2)"
	case "vehiculos":
		expr = `jsonb_build_object('id',v0.id_vehiculo::text,'clientId',v0.id_cliente::text,'plate',v0.placa,'brand',v0.marca,'model',v0.modelo,'year',v0.anio,'mileage',v0.kilometraje_actual::text)`
		from = "vehiculos v0"
		sort = "v0.id_vehiculo"
		where = "NOT $1 OR v0.id_vehiculo IN (SELECT r.id_vehiculo FROM " + orderFrom + " JOIN asignaciones_ot a ON a.id_orden=o.id_orden WHERE a.activo AND a.id_usuario::text=$2)"
	case "recepciones":
		expr = receptionJSON
		from = "recepciones r"
		sort = "r.id_recepcion"
		where = "NOT $1 OR r.id_recepcion IN (SELECT q.id_recepcion FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) JOIN asignaciones_ot a USING(id_orden) WHERE a.activo AND a.id_usuario::text=$2)"
	case "diagnosticos":
		expr = diagnosisJSON
		from = "diagnosticos d"
		sort = "d.id_diagnostico"
		where = "NOT $1 OR d.id_diagnostico IN (SELECT q.id_diagnostico FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) JOIN asignaciones_ot a USING(id_orden) WHERE a.activo AND a.id_usuario::text=$2)"
	case "cotizaciones":
		expr = quoteJSON
		from = "cotizaciones q"
		sort = "q.id_cotizacion"
		where = "NOT $1 OR q.id_cotizacion IN (SELECT o.id_cotizacion FROM ordenes_trabajo o JOIN asignaciones_ot a USING(id_orden) WHERE a.activo AND a.id_usuario::text=$2)"
	case "ordenes":
		if module != "mecanicos" {
			return s.WorkshopList(ctx, "ordenes", domain.WorkshopFilter{Page: p}, u)
		}
		// mecanicos.ver already authorizes the assignments directory. Expose
		// order IDs/states and assignment histories needed by that screen.
		expr, from, where, sort = orderJSON, orderFrom, "NOT $1 AND $2<>''", "o.id_orden"
		scoped = false
	case "mecanicos":
		return s.WorkshopList(ctx, "mecanicos", domain.WorkshopFilter{Page: p}, u)
	case "contactos":
		// Contacts are sensitive; only the delivery workflow for a selected order.
		if module != "pagos" || p.ClientID == "" {
			return nil, domain.ErrForbidden
		}
		return nil, domain.Invalid("orderId", "Consulta contactos desde la orden seleccionada.")
	default:
		return nil, domain.ErrForbidden
	}
	if (module == "diagnosticos" || module == "recepciones") && !receptionGlobal(u) {
		switch resource {
		case "clientes":
			where = "c.id_cliente IN (SELECT r.id_cliente_atencion FROM recepciones r WHERE r.id_mecanico_diagnostico::text=$2) AND $1"
		case "vehiculos":
			where = "v0.id_vehiculo IN (SELECT r.id_vehiculo FROM recepciones r WHERE r.id_mecanico_diagnostico::text=$2) AND $1"
		case "mecanicos":
			return s.WorkshopList(ctx, "mecanicos", domain.WorkshopFilter{Page: p}, u)
		}
		scoped = true
	}
	args := []any{scoped, u.ID}
	if resource == "clientes" {
		where = "(" + where + ") AND jac_normalize(c.nombre || ' ' || c.numero_documento) LIKE jac_normalize($3) AND ($4='' OR c.id_cliente::text=$4)"
		args = append(args, "%"+p.Query+"%", p.ClientID)
	} else if resource == "vehiculos" {
		where = "(" + where + ") AND ($3='' OR v0.id_cliente::text=$3) AND jac_normalize(coalesce(v0.placa,'') || ' ' || coalesce(v0.modelo,'')) LIKE jac_normalize($4)"
		args = append(args, p.ClientID, "%"+p.Query+"%")
	}
	out := domain.List[json.RawMessage]{Data: []json.RawMessage{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+from+" WHERE "+where, args...).Scan(&out.Meta.Total); err != nil {
		return nil, translate(err)
	}
	rows, err := tx.Query(ctx, "SELECT "+expr+" FROM "+from+" WHERE "+where+" ORDER BY "+sort+" LIMIT $"+itoa(len(args)+1)+" OFFSET $"+itoa(len(args)+2), append(args, p.Size, (p.Number-1)*p.Size)...)
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
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
