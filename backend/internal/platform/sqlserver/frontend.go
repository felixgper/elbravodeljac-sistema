package sqlserver

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
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
		expr = `JSON_OBJECT('id':CONVERT(varchar(20),c.id_cliente),'name':c.nombre,'documentType':c.tipo_documento,'documentNumber':c.numero_documento)`
		from = "clientes c"
		sort = "c.id_cliente"
		where = "@p1=0 OR c.id_cliente IN (SELECT r.id_cliente_atencion FROM " + orderFrom + " JOIN asignaciones_ot a ON a.id_orden=o.id_orden WHERE a.activo=1 AND CONVERT(varchar(32),a.id_usuario)=@p2)"
	case "vehiculos":
		expr = `JSON_OBJECT('id':CONVERT(varchar(20),v0.id_vehiculo),'clientId':CONVERT(varchar(20),v0.id_cliente),'plate':v0.placa,'brand':v0.marca,'model':v0.modelo,'year':v0.anio,'mileage':CONVERT(varchar(32),v0.kilometraje_actual))`
		from = "vehiculos v0"
		sort = "v0.id_vehiculo"
		where = "@p1=0 OR v0.id_vehiculo IN (SELECT r.id_vehiculo FROM " + orderFrom + " JOIN asignaciones_ot a ON a.id_orden=o.id_orden WHERE a.activo=1 AND CONVERT(varchar(32),a.id_usuario)=@p2)"
	case "cotizaciones":
		if module == "ordenes" && u.Can("ordenes.crear") {
			return s.WorkshopList(ctx, "cotizaciones", domain.WorkshopFilter{Page: p, State: "aprobada", WithoutOrder: true}, u)
		}
		expr = quoteJSON
		from = "cotizaciones q"
		sort = "q.id_cotizacion"
		where = "@p1=0 OR q.id_cotizacion IN (SELECT o.id_cotizacion FROM ordenes_trabajo o JOIN dbo.asignaciones_ot a ON a.id_orden=o.id_orden WHERE a.activo=1 AND CONVERT(varchar(32),a.id_usuario)=@p2)"
	case "ordenes":
		if module != "mecanicos" {
			return s.WorkshopList(ctx, "ordenes", domain.WorkshopFilter{Page: p}, u)
		}
		// mecanicos.ver already authorizes the assignments directory. Expose
		// order IDs/states and assignment histories needed by that screen.
		expr, from, where, sort = orderJSON, orderFrom, "@p1=0 AND @p2<>''", "o.id_orden"
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
			where = "c.id_cliente IN (SELECT r.id_cliente_atencion FROM recepciones r WHERE CONVERT(varchar(32),r.id_mecanico_diagnostico)=@p2) AND @p1=1"
		case "vehiculos":
			where = "v0.id_vehiculo IN (SELECT r.id_vehiculo FROM recepciones r WHERE CONVERT(varchar(32),r.id_mecanico_diagnostico)=@p2) AND @p1=1"
		case "mecanicos":
			return s.WorkshopList(ctx, "mecanicos", domain.WorkshopFilter{Page: p}, u)
		}
		scoped = true
	}
	args := []any{scoped, u.ID}
	if resource == "clientes" {
		where = "(" + where + ") AND CONCAT(c.nombre,N' ',c.numero_documento) COLLATE Latin1_General_100_CI_AI_SC LIKE @p3 AND (@p4='' OR CONVERT(varchar(32),c.id_cliente)=@p4)"
		args = append(args, "%"+p.Query+"%", p.ClientID)
	} else if resource == "vehiculos" {
		where = "(" + where + ") AND (@p3='' OR CONVERT(varchar(32),v0.id_cliente)=@p3) AND CONCAT(v0.placa,N' ',v0.modelo) COLLATE Latin1_General_100_CI_AI_SC LIKE @p4"
		args = append(args, p.ClientID, "%"+p.Query+"%")
	}
	return s.jsonPage(ctx, p, expr, from, where, sort, args...)
}
