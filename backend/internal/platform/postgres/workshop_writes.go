package postgres

import (
	"context"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"strconv"
	"time"
)

func itoa(n int) string { return strconv.Itoa(n) }
func (s *Store) CreateReception(ctx context.Context, v domain.ReceptionInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	vehicle, err := scanVehicle(tx.QueryRow(ctx, `SELECT `+vehicleCols+` FROM vehiculos WHERE id_vehiculo=$1 FOR UPDATE`, v.VehicleID))
	if err != nil {
		return nil, err
	}
	if vehicle.ClientID != v.ClientID {
		return nil, domain.Invalid("clientId", "El vehículo debe pertenecer al cliente de esta recepción.")
	}
	before, _ := domain.Mileage(vehicle.Mileage)
	after, _ := domain.Mileage(v.Mileage)
	if after < before {
		return nil, domain.Invalid("mileage", "El kilometraje no puede disminuir.")
	}
	if _, err = tx.Exec(ctx, `UPDATE vehiculos SET kilometraje_actual=$2 WHERE id_vehiculo=$1`, v.VehicleID, v.Mileage); err != nil {
		return nil, translate(err)
	}
	var id string
	if err = tx.QueryRow(ctx, `INSERT INTO recepciones(id_vehiculo,id_cliente_atencion,id_recepcionista,km_ingreso,sintomas) VALUES($1,$2,$3,$4,$5) RETURNING id_recepcion::text`, v.VehicleID, v.ClientID, u.ID, v.Mileage, v.Symptoms).Scan(&id); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+receptionJSON+" FROM recepciones r WHERE r.id_recepcion=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) CreateDiagnosis(ctx context.Context, v domain.DiagnosisInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	return readJSON(ctx, s.Pool, `WITH inserted AS (INSERT INTO diagnosticos(id_recepcion,id_tecnico,resumen) VALUES($1,$2,$3) RETURNING *) SELECT `+diagnosisJSON+` FROM inserted d`, v.ReceptionID, u.ID, v.Findings)
}
func (s *Store) SaveQuote(ctx context.Context, id string, v domain.QuoteInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if id != "" {
		var diagnosis string
		if err = tx.QueryRow(ctx, `SELECT id_diagnostico::text FROM cotizaciones WHERE id_cotizacion=$1 FOR UPDATE`, id).Scan(&diagnosis); err != nil {
			return nil, translate(err)
		}
		if diagnosis != v.DiagnosisID {
			return nil, domain.Invalid("diagnosisId", "No se puede cambiar el diagnóstico de una cotización guardada.")
		}
		var used bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ordenes_trabajo WHERE id_cotizacion=$1)`, id).Scan(&used); err != nil {
			return nil, err
		}
		if used {
			return nil, domain.ErrConflict
		}
	} else {
		var diagnosis string
		if err = tx.QueryRow(ctx, `SELECT id_diagnostico::text FROM diagnosticos WHERE id_diagnostico=$1 FOR KEY SHARE`, v.DiagnosisID).Scan(&diagnosis); err != nil {
			return nil, translate(err)
		}
		if err = tx.QueryRow(ctx, `INSERT INTO cotizaciones(id_diagnostico,estado,total,aprobada_en) VALUES($1,$2::text,$3,CASE WHEN $2::text='aprobada' THEN clock_timestamp() END) RETURNING id_cotizacion::text`, v.DiagnosisID, v.Status, v.Total).Scan(&id); err != nil {
			return nil, translate(err)
		}
	}
	for i := range v.Items {
		line := &v.Items[i]
		if line.ServiceID == "" && line.SparePartID == "" {
			continue
		}
		var name string
		var table, column, ref string
		if line.ServiceID != "" {
			table, column, ref = "servicios", "id_servicio", line.ServiceID
		} else {
			table, column, ref = "repuestos", "id_repuesto", line.SparePartID
		}
		err = tx.QueryRow(ctx, "SELECT nombre FROM "+table+" WHERE "+column+"=$1 AND activo FOR SHARE", ref).Scan(&name)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Invalid("items", "Servicio o repuesto inexistente/inactivo.")
		}
		if err != nil {
			return nil, translate(err)
		}
		if line.Concept == "" {
			line.Concept = name
		}
	}
	if _, err = tx.Exec(ctx, `DELETE FROM detalle_cotizacion WHERE id_cotizacion=$1`, id); err != nil {
		return nil, translate(err)
	}
	for _, line := range v.Items {
		_, err = tx.Exec(ctx, `INSERT INTO detalle_cotizacion(id_cotizacion,id_servicio,id_repuesto,descripcion,cantidad,precio_unitario,subtotal,origen_repuesto,tipo_concepto,tipo_repuesto,descuento_unitario) VALUES($1,nullif($2,'')::bigint,nullif($3,'')::bigint,$4,$5,$6,$7,nullif($8,''),$9,nullif($10,''),$11)`, id, line.ServiceID, line.SparePartID, line.Concept, line.Quantity, line.UnitPrice, line.Subtotal, line.Origin, line.Type, line.PartKind, line.Discount)
		if err != nil {
			return nil, translate(err)
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE cotizaciones SET estado=$2::text,total=$3,aprobada_en=CASE WHEN $2::text='aprobada' THEN coalesce(aprobada_en,clock_timestamp()) ELSE NULL END WHERE id_cotizacion=$1`, id, v.Status, v.Total); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+quoteJSON+" FROM cotizaciones q WHERE q.id_cotizacion=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func technicalUser(ctx context.Context, tx pgx.Tx, id string, assistant bool) error {
	var found string
	roles := []string{"mecánico", "mecanico", "supervisor"}
	field := "mechanicId"
	if assistant {
		roles = []string{"auxiliar"}
		field = "assistantId"
	}
	err := tx.QueryRow(ctx, `SELECT u.id_usuario::text FROM usuarios u WHERE u.id_usuario=$1 AND u.activo AND EXISTS(SELECT 1 FROM usuario_roles ur JOIN roles r USING(id_rol) WHERE ur.id_usuario=u.id_usuario AND lower(r.nombre_rol)=ANY($2)) FOR SHARE OF u`, id, roles).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Invalid(field, "Selecciona personal activo con rol técnico compatible.")
	}
	return translate(err)
}
func insertAssignment(ctx context.Context, tx pgx.Tx, order, id, role string, at time.Time) error {
	_, err := tx.Exec(ctx, `INSERT INTO asignaciones_ot(id_orden,id_usuario,rol_en_trabajo,fecha_asignacion) VALUES($1,$2,$3,$4)`, order, id, role, at)
	return translate(err)
}
func (s *Store) CreateOrder(ctx context.Context, v domain.CreateOrderInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var state string
	if err = tx.QueryRow(ctx, `SELECT estado FROM cotizaciones WHERE id_cotizacion=$1 FOR UPDATE`, v.QuoteID).Scan(&state); err != nil {
		return nil, translate(err)
	}
	if state != "aprobada" {
		return nil, domain.Invalid("quoteId", "Solo se crea una OT desde una cotización aprobada.")
	}
	if err = technicalUser(ctx, tx, v.MechanicID, false); err != nil {
		return nil, err
	}
	if v.AssistantID != "" {
		if err = technicalUser(ctx, tx, v.AssistantID, true); err != nil {
			return nil, err
		}
	}
	var id string
	err = tx.QueryRow(ctx, `WITH n AS (SELECT nextval(pg_get_serial_sequence('ordenes_trabajo','id_orden')) AS id) INSERT INTO ordenes_trabajo(id_orden,id_cotizacion,codigo_ot,id_usuario_creacion) SELECT id,$1,'OT-'||id::text,$2 FROM n RETURNING id_orden::text`, v.QuoteID, u.ID).Scan(&id)
	if err != nil {
		return nil, translate(err)
	}
	var at time.Time
	if err = tx.QueryRow(ctx, `SELECT clock_timestamp()`).Scan(&at); err != nil {
		return nil, err
	}
	if err = insertAssignment(ctx, tx, id, v.MechanicID, "MECANICO", at); err != nil {
		return nil, err
	}
	if v.AssistantID != "" {
		if err = insertAssignment(ctx, tx, id, v.AssistantID, "AUXILIAR", at); err != nil {
			return nil, err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO historial_estados_ot(id_orden,id_usuario,estado_nuevo,fecha) VALUES($1,$2,'creada',$3)`, id, u.ID, at); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+orderJSON+" FROM "+orderFrom+" WHERE o.id_orden=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func lockOrder(ctx context.Context, tx pgx.Tx, id string) (string, error) {
	var state string
	err := tx.QueryRow(ctx, `SELECT estado FROM ordenes_trabajo WHERE id_orden=$1 FOR UPDATE`, id).Scan(&state)
	return state, translate(err)
}
func orderClosed(ctx context.Context, tx pgx.Tx, id string) (bool, error) {
	var closed bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM entregas WHERE id_orden=$1)`, id).Scan(&closed)
	return closed, err
}
func canEditOrder(ctx context.Context, tx pgx.Tx, id string, u domain.User) error {
	if u.Can("ordenes.asignar") {
		return nil
	}
	var assigned bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM asignaciones_ot WHERE id_orden=$1 AND id_usuario=$2 AND activo)`, id, u.ID).Scan(&assigned)
	if err != nil {
		return err
	}
	if !assigned {
		return domain.ErrForbidden
	}
	return nil
}
func orderEventTime(ctx context.Context, tx pgx.Tx, id string) (time.Time, error) {
	var at time.Time
	err := tx.QueryRow(ctx, `SELECT greatest(clock_timestamp(),(SELECT max(fecha)+interval '1 microsecond' FROM avances_ot WHERE id_orden=$1),(SELECT max(fecha)+interval '1 microsecond' FROM pruebas_ot WHERE id_orden=$1),(SELECT max(fecha)+interval '1 microsecond' FROM historial_estados_ot WHERE id_orden=$1))`, id).Scan(&at)
	return at, err
}
func (s *Store) UpdateOrder(ctx context.Context, id string, v domain.UpdateOrderInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if (v.MechanicID != nil || v.AssistantID != nil) && !u.Can("ordenes.asignar") {
		return nil, domain.ErrForbidden
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	state, err := lockOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if err = canEditOrder(ctx, tx, id, u); err != nil {
		return nil, err
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if closed || state == "cancelada" || state == "finalizada" {
		return nil, domain.ErrConflict
	}
	if v.State == "cancelada" {
		var paid bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pagos WHERE id_orden=$1)`, id).Scan(&paid); err != nil {
			return nil, err
		}
		if paid {
			return nil, domain.ErrConflict
		}
	}
	var currentMechanic, currentAssistant string
	if err = tx.QueryRow(ctx, `SELECT coalesce(max(id_usuario::text) FILTER(WHERE rol_en_trabajo='MECANICO'),''),coalesce(max(id_usuario::text) FILTER(WHERE rol_en_trabajo='AUXILIAR'),'') FROM asignaciones_ot WHERE id_orden=$1 AND activo`, id).Scan(&currentMechanic, &currentAssistant); err != nil {
		return nil, err
	}
	nextMechanic, nextAssistant := currentMechanic, currentAssistant
	if v.MechanicID != nil {
		nextMechanic = *v.MechanicID
	}
	if v.AssistantID != nil {
		nextAssistant = *v.AssistantID
	}
	if nextAssistant != "" && nextAssistant == nextMechanic {
		return nil, domain.Invalid("assistantId", "Mecánico responsable y auxiliar deben ser personas distintas.")
	}
	if v.MechanicID != nil {
		if err = technicalUser(ctx, tx, *v.MechanicID, false); err != nil {
			return nil, err
		}
	}
	if v.AssistantID != nil && *v.AssistantID != "" {
		if err = technicalUser(ctx, tx, *v.AssistantID, true); err != nil {
			return nil, err
		}
	}
	at, err := orderEventTime(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	for _, assignment := range []struct {
		id   *string
		role string
	}{{v.MechanicID, "MECANICO"}, {v.AssistantID, "AUXILIAR"}} {
		if assignment.id == nil || (assignment.role == "MECANICO" && *assignment.id == currentMechanic) || (assignment.role == "AUXILIAR" && *assignment.id == currentAssistant) {
			continue
		}
		if _, err = tx.Exec(ctx, `UPDATE asignaciones_ot SET activo=false WHERE id_orden=$1 AND rol_en_trabajo=$2 AND activo`, id, assignment.role); err != nil {
			return nil, translate(err)
		}
		if *assignment.id != "" {
			if err = insertAssignment(ctx, tx, id, *assignment.id, assignment.role, at); err != nil {
				return nil, err
			}
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO avances_ot(id_orden,id_usuario,observacion,fecha) VALUES($1,$2,$3,$4)`, id, u.ID, v.Note, at); err != nil {
		return nil, translate(err)
	}
	// Every repair update invalidates earlier tests, including repeated repair state.
	if _, err = tx.Exec(ctx, `INSERT INTO historial_estados_ot(id_orden,id_usuario,estado_anterior,estado_nuevo,fecha) VALUES($1,$2,$3,$4,$5)`, id, u.ID, state, v.State, at); err != nil {
		return nil, translate(err)
	}
	if v.Result != "" {
		if _, err = tx.Exec(ctx, `INSERT INTO pruebas_ot(id_orden,id_usuario,tipo_prueba,resultado,observaciones,fecha) VALUES($1,$2,$3,$4,$5,$6)`, id, u.ID, v.Test, v.Result, v.Note, at); err != nil {
			return nil, translate(err)
		}
	}
	if v.State == "finalizada" {
		var passed bool
		err = tx.QueryRow(ctx, `SELECT coalesce((SELECT p.resultado='satisfactoria' AND p.fecha>=coalesce((SELECT max(h.fecha) FROM historial_estados_ot h WHERE h.id_orden=$1 AND h.estado_nuevo='en_reparacion'),'-infinity'::timestamptz) FROM pruebas_ot p WHERE p.id_orden=$1 ORDER BY p.fecha DESC,p.id_prueba DESC LIMIT 1),false)`, id).Scan(&passed)
		if err != nil {
			return nil, err
		}
		if !passed {
			return nil, domain.Invalid("result", "Finalizar exige la última prueba satisfactoria, posterior a la última reparación.")
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE ordenes_trabajo SET estado=$2::text,fecha_inicio=CASE WHEN $2::text IN ('en_reparacion','en_pruebas','finalizada') THEN coalesce(fecha_inicio,$3) ELSE fecha_inicio END,fecha_fin=CASE WHEN $2::text IN ('finalizada','cancelada') THEN $3 ELSE NULL END WHERE id_orden=$1`, id, v.State, at); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+orderJSON+" FROM "+orderFrom+" WHERE o.id_orden=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) DeleteOrder(ctx context.Context, id string, u domain.User) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = lockOrder(ctx, tx, id); err != nil {
		return err
	}
	var used bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pagos WHERE id_orden=$1) OR EXISTS(SELECT 1 FROM entregas WHERE id_orden=$1) OR EXISTS(SELECT 1 FROM avances_ot WHERE id_orden=$1) OR EXISTS(SELECT 1 FROM pruebas_ot WHERE id_orden=$1)`, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return domain.ErrConflict
	}
	for _, table := range []string{"historial_estados_ot", "asignaciones_ot", "ordenes_trabajo"} {
		if _, err = tx.Exec(ctx, "DELETE FROM "+table+" WHERE id_orden=$1", id); err != nil {
			return translate(err)
		}
	}
	return translate(tx.Commit(ctx))
}
func (s *Store) RegisterPayment(ctx context.Context, id string, v domain.PaymentInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	state, err := lockOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if closed || state != "finalizada" {
		return nil, domain.Invalid("orderId", "Solo se pagan órdenes finalizadas y aún no entregadas.")
	}
	var total, paid string
	if err = tx.QueryRow(ctx, `SELECT q.total::text,coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=$1),0)::numeric(14,2)::text FROM cotizaciones q JOIN ordenes_trabajo o USING(id_cotizacion) WHERE o.id_orden=$1`, id).Scan(&total, &paid); err != nil {
		return nil, translate(err)
	}
	totalCents, err := domain.Money(total, "total")
	if err != nil {
		return nil, err
	}
	paidCents, err := domain.Money(paid, "paid")
	if err != nil {
		return nil, err
	}
	amount, _ := domain.Money(v.Amount, "amount")
	if amount > totalCents-paidCents {
		return nil, domain.Invalid("amount", "El pago supera el saldo pendiente.")
	}
	out, err := readJSON(ctx, tx, `WITH inserted AS (INSERT INTO pagos(id_orden,monto,metodo_pago,comprobante,id_usuario_registro) VALUES($1,$2,$3,nullif($4,''),$5) RETURNING *) SELECT `+paymentJSON+` FROM inserted p`, id, v.Amount, v.Method, v.Receipt, u.ID)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) Deliver(ctx context.Context, id string, v domain.DeliveryInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	state, err := lockOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if closed {
		return nil, domain.ErrConflict
	}
	if state != "finalizada" {
		return nil, domain.Invalid("orderId", "La orden debe estar finalizada.")
	}
	var clientID string
	var settled bool
	err = tx.QueryRow(ctx, `SELECT r.id_cliente_atencion::text,q.total=coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=$1),0) FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) JOIN diagnosticos d USING(id_diagnostico) JOIN recepciones r USING(id_recepcion) WHERE o.id_orden=$1`, id).Scan(&clientID, &settled)
	if err != nil {
		return nil, translate(err)
	}
	if !settled {
		return nil, domain.Invalid("balance", "Completa el pago exacto antes de entregar.")
	}
	var name, docType, doc, phone string
	if v.Kind == "Tercero" {
		err = tx.QueryRow(ctx, `SELECT nombre,tipo_documento,numero_documento,coalesce(telefono,'') FROM contactos_cliente WHERE id_contacto=$1 AND id_cliente=$2 AND activo FOR SHARE`, v.ContactID, clientID).Scan(&name, &docType, &doc, &phone)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.Invalid("contactId", "El contacto activo debe pertenecer al cliente de la atención.")
		}
	} else {
		err = tx.QueryRow(ctx, `SELECT nombre,tipo_documento,numero_documento,coalesce(telefono,'') FROM clientes WHERE id_cliente=$1 FOR SHARE`, clientID).Scan(&name, &docType, &doc, &phone)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, `WITH inserted AS (INSERT INTO entregas(id_orden,id_usuario_entrega,id_contacto_receptor,tipo_receptor,nombre_receptor_snap,tipo_doc_receptor_snap,documento_receptor_snap,telefono_receptor_snap,autorizacion_verificada,medio_verificacion,fecha_verificacion,id_usuario_verificador) VALUES($1,$2,nullif($3,'')::bigint,$4,$5,$6,$7,$8,true,$9,clock_timestamp(),$2) RETURNING *) SELECT `+deliveryJSON+` FROM inserted e`, id, u.ID, v.ContactID, v.Kind, name, docType, doc, phone, v.Method)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
