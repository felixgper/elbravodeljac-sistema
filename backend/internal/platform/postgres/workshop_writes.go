package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
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
	if v.TechnicianID != "" {
		if v.AttentionType != "Diagnostico" {
			return nil, domain.Invalid("technicianId", "El servicio directo no lleva asignación de diagnóstico.")
		}
		if !u.Can("diagnosticos.asignar") {
			return nil, domain.ErrForbidden
		}
		if err = technicalUser(ctx, tx, v.TechnicianID); err != nil {
			return nil, err
		}
	}
	var id string
	if err = tx.QueryRow(ctx, `INSERT INTO recepciones(id_vehiculo,id_cliente_atencion,id_recepcionista,km_ingreso,sintomas,tipo_atencion,id_mecanico_diagnostico,diagnostico_asignado_por,diagnostico_asignado_en,asignacion_version,historial_asignaciones) VALUES($1,$2,$3::bigint,$4,$5,$6,nullif($7,'')::bigint,CASE WHEN $7<>'' THEN $3::bigint END,CASE WHEN $7<>'' THEN clock_timestamp() END,CASE WHEN $7<>'' THEN 1 ELSE 0 END,CASE WHEN $7<>'' THEN jsonb_build_array(jsonb_build_object('technicianId',$7::text,'by',$3::text,'at',clock_timestamp())) ELSE '[]'::jsonb END) RETURNING id_recepcion::text`, v.VehicleID, v.ClientID, u.ID, v.Mileage, v.Symptoms, v.AttentionType, v.TechnicianID).Scan(&id); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+receptionJSON+" FROM recepciones r WHERE r.id_recepcion=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) CreateDiagnosis(ctx context.Context, v domain.DiagnosisInput, u domain.User) (any, error) {
	return s.SaveDiagnosis(ctx, "", v, u)
}

func (s *Store) SaveQuote(ctx context.Context, id string, v domain.QuoteInput, u domain.User) (any, error) {
	wasExisting := id != ""
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	if v.DiagnosisID != "" {
		var reception, state string
		if err = tx.QueryRow(ctx, `SELECT id_recepcion::text,estado FROM diagnosticos WHERE id_diagnostico=$1 FOR KEY SHARE`, v.DiagnosisID).Scan(&reception, &state); err != nil {
			return nil, translate(err)
		}
		if state != "Finalizado" {
			return nil, domain.Invalid("diagnosisId", "Finaliza el diagnóstico antes de cotizar.")
		}
		if v.ReceptionID != "" && v.ReceptionID != reception {
			return nil, domain.Invalid("receptionId", "Diagnóstico y recepción no corresponden.")
		}
		v.ReceptionID = reception
	}
	var attention string
	if err = tx.QueryRow(ctx, `SELECT tipo_atencion FROM recepciones WHERE id_recepcion=$1 FOR UPDATE`, v.ReceptionID).Scan(&attention); err != nil {
		return nil, translate(err)
	}
	if (attention == "Directo") != (v.DiagnosisID == "") {
		return nil, domain.Invalid("diagnosisId", "El tipo de atención determina si se requiere diagnóstico.")
	}
	rawContent, _ := json.Marshal(v.Items)
	contentHash := fmt.Sprintf("%x", sha256.Sum256(rawContent))
	var oldHash, oldState string
	var oldVersion int64
	var oldApproval []byte
	if id != "" {
		if err = tx.QueryRow(ctx, `SELECT version,coalesce(contenido_hash,''),estado,aprobacion FROM cotizaciones WHERE id_cotizacion=$1 FOR UPDATE`, id).Scan(&oldVersion, &oldHash, &oldState, &oldApproval); err != nil {
			return nil, translate(err)
		}
		if oldVersion != v.Version {
			return nil, domain.Invalid("version", "La cotización cambió. Recarga antes de editar.")
		}
	}
	changed := id != "" && oldHash != contentHash
	if changed && oldState == "aprobada" {
		v.Status = "pendiente"
		v.Approval = nil
	}
	if v.Status == "aprobada" && (id == "" || changed || oldState != "aprobada" || len(oldApproval) == 0) {
		if v.Approval == nil {
			return nil, domain.Invalid("approval", "Registra la persona y el medio de aprobación del cliente.")
		}
		if err = v.Approval.Validate(); err != nil {
			return nil, err
		}
	}
	if id != "" {
		var diagnosis, reception string
		if err = tx.QueryRow(ctx, `SELECT coalesce(id_diagnostico::text,''),id_recepcion::text FROM cotizaciones WHERE id_cotizacion=$1 FOR UPDATE`, id).Scan(&diagnosis, &reception); err != nil {
			return nil, translate(err)
		}
		if diagnosis != v.DiagnosisID || reception != v.ReceptionID {
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
		var duplicate bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cotizaciones WHERE id_recepcion=$1 AND id_cotizacion_base IS NULL)`, v.ReceptionID).Scan(&duplicate); err != nil {
			return nil, err
		}
		if duplicate && v.ParentQuoteID == "" {
			return nil, domain.Invalid("receptionId", "Esta atención ya tiene una cotización. Abre la existente; no se puede duplicar sin un flujo autorizado de rectificación.")
		}
		if v.ParentQuoteID != "" {
			if !u.Can("cotizaciones.editar") {
				return nil, domain.ErrForbidden
			}
			var allowed bool
			if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cotizaciones q JOIN ordenes_trabajo o USING(id_cotizacion) WHERE q.id_cotizacion=$1 AND q.id_recepcion=$2 AND q.id_cotizacion_base IS NULL AND o.estado NOT IN ('finalizada','cancelada') AND NOT EXISTS(SELECT 1 FROM entregas e WHERE e.id_orden=o.id_orden))`, v.ParentQuoteID, v.ReceptionID).Scan(&allowed); err != nil {
				return nil, err
			}
			if !allowed {
				return nil, domain.Invalid("parentQuoteId", "La ampliación requiere una orden original abierta de la misma atención.")
			}
		}
		if err = tx.QueryRow(ctx, `INSERT INTO cotizaciones(id_diagnostico,id_recepcion,estado,total,aprobada_en) VALUES(nullif($1,'')::bigint,$2,$3::text,$4,CASE WHEN $3::text='aprobada' THEN clock_timestamp() END) RETURNING id_cotizacion::text`, v.DiagnosisID, v.ReceptionID, v.Status, v.Total).Scan(&id); err != nil {
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
	if wasExisting {
		if _, err = tx.Exec(ctx, `UPDATE cotizaciones q SET historial_revision=historial_revision||jsonb_build_array(jsonb_build_object('version',version,'approval',aprobacion,'status',estado,'total',total,'items',(SELECT jsonb_agg(to_jsonb(d) ORDER BY id_detalle) FROM detalle_cotizacion d WHERE d.id_cotizacion=q.id_cotizacion),'by',$2::text,'at',clock_timestamp())) WHERE id_cotizacion=$1`, id, u.ID); err != nil {
			return nil, translate(err)
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
	// A newly recorded acceptance has its own server timestamp, even when only
	// the approving person/method changed and line content stayed identical.
	if v.Status == "aprobada" && v.Approval != nil {
		if _, err = tx.Exec(ctx, `UPDATE cotizaciones SET aprobada_en=clock_timestamp() WHERE id_cotizacion=$1`, id); err != nil {
			return nil, translate(err)
		}
	}
	var approval any
	if v.Status == "aprobada" {
		if v.Approval == nil && !changed && oldState == "aprobada" {
			approval = oldApproval
		} else {
			approval, _ = json.Marshal(map[string]any{"name": v.Approval.Name, "method": v.Approval.Method, "by": u.ID, "version": oldVersion + 1, "items": v.Items, "total": v.Total})
		}
	}
	if _, err = tx.Exec(ctx, `UPDATE cotizaciones SET version=version+1,contenido_hash=$2,aprobacion=CASE WHEN $3::jsonb IS NULL THEN NULL ELSE $3::jsonb||jsonb_build_object('at',coalesce(aprobada_en,clock_timestamp())) END,id_cotizacion_base=coalesce(id_cotizacion_base,nullif($4,'')::bigint) WHERE id_cotizacion=$1`, id, contentHash, approval, v.ParentQuoteID); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+quoteJSON+" FROM cotizaciones q WHERE q.id_cotizacion=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func technicalUser(ctx context.Context, tx pgx.Tx, id string) error {
	var found string
	roles := []string{"mecánico", "mecanico"}
	err := tx.QueryRow(ctx, `SELECT u.id_usuario::text FROM usuarios u WHERE u.id_usuario=$1 AND u.activo AND EXISTS(SELECT 1 FROM usuario_roles ur JOIN roles r USING(id_rol) WHERE ur.id_usuario=u.id_usuario AND r.activo AND lower(r.nombre_rol)=ANY($2)) FOR SHARE OF u`, id, roles).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Invalid("mechanicId", "Selecciona un mecánico activo.")
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
	if err = tx.QueryRow(ctx, `SELECT CASE WHEN aprobacion IS NULL THEN 'sin_aprobacion' ELSE estado END FROM cotizaciones WHERE id_cotizacion=$1 FOR UPDATE`, v.QuoteID).Scan(&state); err != nil {
		return nil, translate(err)
	}
	if state != "aprobada" {
		return nil, domain.Invalid("quoteId", "Solo se crea una OT desde una cotización aprobada.")
	}
	if err = technicalUser(ctx, tx, v.MechanicID); err != nil {
		return nil, err
	}
	var id string
	err = tx.QueryRow(ctx, `WITH n AS (SELECT nextval(pg_get_serial_sequence('ordenes_trabajo','id_orden')) AS id) INSERT INTO ordenes_trabajo(id_orden,id_cotizacion,codigo_ot,id_usuario_creacion,flujo_version) SELECT id,$1,'OT-'||id::text,$2,1 FROM n RETURNING id_orden::text`, v.QuoteID, u.ID).Scan(&id)
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
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM asignaciones_ot WHERE id_orden=$1 AND id_usuario=$2 AND activo AND rol_en_trabajo='MECANICO')`, id, u.ID).Scan(&assigned)
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
	if (v.MechanicID != nil) && !u.Can("ordenes.asignar") {
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
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM pagos WHERE id_orden=$1 AND revertido_en IS NULL)`, id).Scan(&paid); err != nil {
			return nil, err
		}
		if paid {
			return nil, domain.ErrConflict
		}
	}
	var currentMechanic string
	if err = tx.QueryRow(ctx, `SELECT coalesce(max(id_usuario::text),'') FROM asignaciones_ot WHERE id_orden=$1 AND activo AND rol_en_trabajo='MECANICO'`, id).Scan(&currentMechanic); err != nil {
		return nil, err
	}
	if v.MechanicID != nil {
		if err = technicalUser(ctx, tx, *v.MechanicID); err != nil {
			return nil, err
		}
	}
	at, err := orderEventTime(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if v.MechanicID != nil && *v.MechanicID != currentMechanic {
		if _, err = tx.Exec(ctx, `UPDATE asignaciones_ot SET activo=false WHERE id_orden=$1 AND rol_en_trabajo='MECANICO' AND activo`, id); err != nil {
			return nil, translate(err)
		}
		if err = insertAssignment(ctx, tx, id, *v.MechanicID, "MECANICO", at); err != nil {
			return nil, err
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
		if err = validateExecution(ctx, tx, id); err != nil {
			return nil, err
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
	var replay []byte
	replayErr := tx.QueryRow(ctx, `SELECT `+paymentJSON+` FROM pagos p WHERE id_orden=$1 AND request_id=$2`, id, v.RequestID).Scan(&replay)
	if replayErr == nil {
		var previous map[string]any
		json.Unmarshal(replay, &previous)
		if previous["reversedAt"] != nil {
			return nil, domain.ErrConflict
		}
		if previous["amount"] != v.Amount || previous["method"] != v.Method || previous["comprobante"] != v.Receipt {
			return nil, domain.ErrConflict
		}
		return json.RawMessage(replay), nil
	}
	if !errors.Is(replayErr, pgx.ErrNoRows) {
		return nil, translate(replayErr)
	}
	if closed || state == "cancelada" {
		return nil, domain.Invalid("orderId", "No se pagan órdenes canceladas o ya entregadas.")
	}
	var total, paid string
	if err = tx.QueryRow(ctx, `SELECT jac_importe_orden(o.id_orden)::text,coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=$1 AND revertido_en IS NULL),0)::numeric(14,2)::text FROM cotizaciones q JOIN ordenes_trabajo o USING(id_cotizacion) WHERE o.id_orden=$1`, id).Scan(&total, &paid); err != nil {
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
	out, err := readJSON(ctx, tx, `WITH inserted AS (INSERT INTO pagos(id_orden,monto,metodo_pago,comprobante,id_usuario_registro,request_id) VALUES($1,$2,$3,nullif($4,''),$5,$6) RETURNING *) SELECT `+paymentJSON+` FROM inserted p`, id, v.Amount, v.Method, v.Receipt, u.ID, v.RequestID)
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
	err = tx.QueryRow(ctx, `SELECT r.id_cliente_atencion::text,jac_importe_orden(o.id_orden)=coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=$1 AND revertido_en IS NULL),0) FROM ordenes_trabajo o JOIN cotizaciones q USING(id_cotizacion) JOIN recepciones r ON r.id_recepcion=q.id_recepcion WHERE o.id_orden=$1`, id).Scan(&clientID, &settled)
	if err != nil {
		return nil, translate(err)
	}
	if !settled {
		return nil, domain.Invalid("balance", "Completa el pago exacto antes de entregar.")
	}
	var unfinished bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM ordenes_trabajo x JOIN cotizaciones xq USING(id_cotizacion) WHERE xq.id_recepcion=(SELECT q.id_recepcion FROM ordenes_trabajo z JOIN cotizaciones q USING(id_cotizacion) WHERE z.id_orden=$1) AND x.id_orden<>$1 AND x.estado<>'cancelada' AND (x.estado<>'finalizada' OR jac_importe_orden(x.id_orden)<>coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=x.id_orden AND revertido_en IS NULL),0))) OR EXISTS(SELECT 1 FROM cotizaciones xq WHERE xq.id_recepcion=(SELECT q.id_recepcion FROM ordenes_trabajo z JOIN cotizaciones q USING(id_cotizacion) WHERE z.id_orden=$1) AND xq.id_cotizacion_base IS NOT NULL AND xq.estado='aprobada' AND NOT EXISTS(SELECT 1 FROM ordenes_trabajo x WHERE x.id_cotizacion=xq.id_cotizacion))`, id).Scan(&unfinished); err != nil {
		return nil, err
	}
	if unfinished {
		return nil, domain.Invalid("attention", "Completa y liquida todas las órdenes y ampliaciones aprobadas de esta atención antes de entregar el vehículo.")
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
	if v.RecipientKind == "Propietario o titular" && docType == "RUC" {
		return nil, domain.Invalid("recipientKind", "Una empresa requiere identificar a su representante físico mediante un contacto.")
	}
	receiver, _ := json.Marshal(map[string]any{"kind": v.RecipientKind, "name": name, "documentType": docType, "documentNumber": doc, "phone": phone})
	conformity, _ := json.Marshal(map[string]any{"accepted": v.Conformity, "notes": v.ConformityNotes, "recordedBy": u.ID})
	if _, err = tx.Exec(ctx, `UPDATE entregas SET receptor_real=$2,conformidad=$3::jsonb||jsonb_build_object('at',clock_timestamp()) WHERE id_orden=$1`, id, receiver, conformity); err != nil {
		return nil, translate(err)
	}
	var snapshot []byte
	if err = tx.QueryRow(ctx, sheetSnapshotSQL, id).Scan(&snapshot); err != nil {
		return nil, translate(err)
	}
	// Freeze all data even when unusually long content exceeds the fixed template.
	// Printing must not invalidate an otherwise valid vehicle delivery. Downloads
	// report a precise 422 instead of cropping; the UI exposes sheetReady=false.
	pdf, renderErr := technicalsheet.Render(snapshot)
	if renderErr != nil {
		pdf = nil
	}
	if _, err = tx.Exec(ctx, `UPDATE entregas SET ficha_snapshot=$2,ficha_pdf=$3 WHERE id_orden=$1`, id, snapshot, pdf); err != nil {
		return nil, translate(err)
	}
	out, err = readJSON(ctx, tx, "SELECT "+deliveryJSON+" FROM entregas e WHERE e.id_orden=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
