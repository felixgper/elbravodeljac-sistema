package sqlserver

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"time"
)

func (s *Store) CreateReception(ctx context.Context, v domain.ReceptionInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	vehicle, err := scanVehicle(tx.QueryRowContext(ctx, `SELECT `+vehicleCols+` FROM vehiculos WITH (UPDLOCK,HOLDLOCK) WHERE id_vehiculo=@p1`, v.VehicleID))
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
	if _, err = tx.ExecContext(ctx, `UPDATE vehiculos SET kilometraje_actual=@p2 WHERE id_vehiculo=@p1`, v.VehicleID, v.Mileage); err != nil {
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
	if err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.recepciones_id_recepcion_seq;
INSERT INTO recepciones(id_recepcion,id_vehiculo,id_cliente_atencion,id_recepcionista,km_ingreso,sintomas,tipo_atencion,id_mecanico_diagnostico,diagnostico_asignado_por,diagnostico_asignado_en,asignacion_version,historial_asignaciones) VALUES(@id,@p1,@p2,@p3,@p4,@p5,@p6,NULLIF(@p7,N''),
CASE WHEN @p7<>N'' THEN @p3 END,CASE WHEN @p7<>N'' THEN TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00') END,
CASE WHEN @p7<>N'' THEN 1 ELSE 0 END,
CASE WHEN @p7<>N'' THEN JSON_ARRAY(JSON_QUERY(JSON_OBJECT('technicianId':@p7,'by':@p3,'at':TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00')))) ELSE N'[]' END);
SELECT CONVERT(varchar(20),@id);`, v.VehicleID, v.ClientID, u.ID, v.Mileage, v.Symptoms, v.AttentionType, v.TechnicianID).Scan(&id); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+receptionJSON+" FROM recepciones r WHERE r.id_recepcion=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}

func (s *Store) SaveQuote(ctx context.Context, id string, v domain.QuoteInput, u domain.User) (any, error) {
	wasExisting := id != ""
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	if v.DiagnosisID != "" {
		var reception, state string
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),id_recepcion),estado FROM diagnosticos WITH (HOLDLOCK) WHERE id_diagnostico=@p1`, v.DiagnosisID).Scan(&reception, &state); err != nil {
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
	if err = tx.QueryRowContext(ctx, `SELECT tipo_atencion FROM recepciones WITH (UPDLOCK,HOLDLOCK) WHERE id_recepcion=@p1`, v.ReceptionID).Scan(&attention); err != nil {
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
		if err = tx.QueryRowContext(ctx, `SELECT version,coalesce(contenido_hash,''),estado,aprobacion FROM cotizaciones WITH (UPDLOCK,HOLDLOCK) WHERE id_cotizacion=@p1`, id).Scan(&oldVersion, &oldHash, &oldState, &oldApproval); err != nil {
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
		if err = tx.QueryRowContext(ctx, `SELECT coalesce(CONVERT(varchar(20),id_diagnostico),''),CONVERT(varchar(20),id_recepcion) FROM cotizaciones WITH (UPDLOCK,HOLDLOCK) WHERE id_cotizacion=@p1`, id).Scan(&diagnosis, &reception); err != nil {
			return nil, translate(err)
		}
		if diagnosis != v.DiagnosisID || reception != v.ReceptionID {
			return nil, domain.Invalid("diagnosisId", "No se puede cambiar el diagnóstico de una cotización guardada.")
		}
		var used bool
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM ordenes_trabajo WHERE id_cotizacion=@p1) THEN 1 ELSE 0 END)`, id).Scan(&used); err != nil {
			return nil, err
		}
		if used {
			return nil, domain.ErrConflict
		}
	} else {
		var duplicate bool
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM cotizaciones WHERE id_recepcion=@p1 AND id_cotizacion_base IS NULL) THEN 1 ELSE 0 END)`, v.ReceptionID).Scan(&duplicate); err != nil {
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
			if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM cotizaciones q JOIN ordenes_trabajo o ON o.id_cotizacion=q.id_cotizacion WHERE q.id_cotizacion=@p1 AND q.id_recepcion=@p2 AND q.id_cotizacion_base IS NULL AND o.estado NOT IN ('finalizada','cancelada') AND NOT EXISTS(SELECT 1 FROM entregas e WHERE e.id_orden=o.id_orden)) THEN 1 ELSE 0 END)`, v.ParentQuoteID, v.ReceptionID).Scan(&allowed); err != nil {
				return nil, err
			}
			if !allowed {
				return nil, domain.Invalid("parentQuoteId", "La ampliación requiere una orden original abierta de la misma atención.")
			}
		}
		if err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.cotizaciones_id_cotizacion_seq;
INSERT INTO cotizaciones(id_cotizacion,id_diagnostico,id_recepcion,estado,total,aprobada_en)
VALUES(@id,NULLIF(@p1,N''),@p2,@p3,@p4,CASE WHEN @p3=N'aprobada' THEN TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00') END);
SELECT CONVERT(varchar(20),@id);`, v.DiagnosisID, v.ReceptionID, v.Status, v.Total).Scan(&id); err != nil {
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
		err = tx.QueryRowContext(ctx, "SELECT nombre FROM dbo."+table+" WITH (HOLDLOCK) WHERE "+column+"=@p1 AND activo=1", ref).Scan(&name)
		if errors.Is(err, sql.ErrNoRows) {
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
		if _, err = tx.ExecContext(ctx, `UPDATE q SET historial_revision=JSON_MODIFY(q.historial_revision,'append $',JSON_QUERY(JSON_OBJECT(
'version':q.version,'approval':JSON_QUERY(q.aprobacion),'status':q.estado,'total':q.total,
'items':JSON_QUERY((SELECT d.* FROM detalle_cotizacion d WHERE d.id_cotizacion=q.id_cotizacion ORDER BY d.id_detalle FOR JSON PATH, INCLUDE_NULL_VALUES)),
'by':@p2,'at':TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00'))))
FROM cotizaciones q WHERE q.id_cotizacion=@p1`, id, u.ID); err != nil {
			return nil, translate(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM detalle_cotizacion WHERE id_cotizacion=@p1`, id); err != nil {
		return nil, translate(err)
	}
	for _, line := range v.Items {
		_, err = tx.ExecContext(ctx, `INSERT INTO detalle_cotizacion(id_cotizacion,id_servicio,id_repuesto,descripcion,cantidad,precio_unitario,subtotal,origen_repuesto,tipo_concepto,tipo_repuesto,descuento_unitario) VALUES(@p1,nullif(@p2,''),nullif(@p3,''),@p4,@p5,@p6,@p7,nullif(@p8,''),@p9,nullif(@p10,''),@p11)`, id, line.ServiceID, line.SparePartID, line.Concept, line.Quantity, line.UnitPrice, line.Subtotal, line.Origin, line.Type, line.PartKind, line.Discount)
		if err != nil {
			return nil, translate(err)
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cotizaciones SET estado=@p2,total=@p3,aprobada_en=CASE WHEN @p2='aprobada' THEN coalesce(aprobada_en,TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00')) ELSE NULL END WHERE id_cotizacion=@p1`, id, v.Status, v.Total); err != nil {
		return nil, translate(err)
	}
	// A newly recorded acceptance has its own server timestamp, even when only
	// the approving person/method changed and line content stayed identical.
	if v.Status == "aprobada" && v.Approval != nil {
		if _, err = tx.ExecContext(ctx, `UPDATE cotizaciones SET aprobada_en=TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00') WHERE id_cotizacion=@p1`, id); err != nil {
			return nil, translate(err)
		}
	}
	var approval any
	if v.Status == "aprobada" {
		if v.Approval == nil && !changed && oldState == "aprobada" {
			approval = string(oldApproval)
		} else {
			payload, _ := json.Marshal(map[string]any{"name": v.Approval.Name, "method": v.Approval.Method, "by": u.ID, "version": oldVersion + 1, "items": v.Items, "total": v.Total})
			approval = string(payload)
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE cotizaciones SET version=version+1,contenido_hash=@p2,aprobacion=CASE WHEN @p3 IS NULL THEN NULL ELSE JSON_MODIFY(@p3,'$.at',CONVERT(varchar(40),COALESCE(aprobada_en,TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00')),127)) END,id_cotizacion_base=coalesce(id_cotizacion_base,nullif(@p4,'')) WHERE id_cotizacion=@p1`, id, contentHash, approval, v.ParentQuoteID); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+quoteJSON+" FROM cotizaciones q WHERE q.id_cotizacion=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func technicalUser(ctx context.Context, tx *sql.Tx, id string) error {
	var found string
	err := tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),u.id_usuario) FROM usuarios u WITH (HOLDLOCK) WHERE u.id_usuario=@p1 AND u.activo=1 AND NOT EXISTS(SELECT 1 FROM empleados e WHERE e.id_empleado=u.id_empleado AND e.activo=0) AND EXISTS(SELECT 1 FROM usuario_roles ur JOIN roles r ON r.id_rol=ur.id_rol WHERE ur.id_usuario=u.id_usuario AND r.activo=1 AND LOWER(r.nombre_rol COLLATE Latin1_General_100_CI_AS_SC) IN (N'mecánico',N'mecanico'))`, id).Scan(&found)
	if errors.Is(err, sql.ErrNoRows) {
		return domain.Invalid("mechanicId", "Selecciona un mecánico activo.")
	}
	return translate(err)
}
func insertAssignment(ctx context.Context, tx *sql.Tx, order, id, role string, at time.Time) error {
	_, err := tx.ExecContext(ctx, `INSERT INTO asignaciones_ot(id_orden,id_usuario,rol_en_trabajo,fecha_asignacion) VALUES(@p1,@p2,@p3,@p4)`, order, id, role, at)
	return translate(err)
}
func (s *Store) CreateOrder(ctx context.Context, v domain.CreateOrderInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var state string
	if err = tx.QueryRowContext(ctx, `SELECT CASE WHEN aprobacion IS NULL THEN 'sin_aprobacion' ELSE estado END FROM cotizaciones WITH (UPDLOCK,HOLDLOCK) WHERE id_cotizacion=@p1`, v.QuoteID).Scan(&state); err != nil {
		return nil, translate(err)
	}
	if state != "aprobada" {
		return nil, domain.Invalid("quoteId", "Solo se crea una OT desde una cotización aprobada.")
	}
	if err = technicalUser(ctx, tx, v.MechanicID); err != nil {
		return nil, err
	}
	var id string
	err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.ordenes_trabajo_id_orden_seq;
INSERT INTO ordenes_trabajo(id_orden,id_cotizacion,codigo_ot,id_usuario_creacion,flujo_version)
VALUES(@id,@p1,CONCAT(N'OT-',@id),@p2,1);
SELECT CONVERT(varchar(20),@id);`, v.QuoteID, u.ID).Scan(&id)
	if err != nil {
		return nil, translate(err)
	}
	var at time.Time
	if err = tx.QueryRowContext(ctx, `SELECT TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00')`).Scan(&at); err != nil {
		return nil, err
	}
	if err = insertAssignment(ctx, tx, id, v.MechanicID, "MECANICO", at); err != nil {
		return nil, err
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO historial_estados_ot(id_orden,id_usuario,estado_nuevo,fecha) VALUES(@p1,@p2,'creada',@p3)`, id, u.ID, at); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+orderJSON+" FROM "+orderFrom+" WHERE o.id_orden=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func lockOrder(ctx context.Context, tx *sql.Tx, id string) (string, error) {
	var state string
	err := tx.QueryRowContext(ctx, `SELECT estado FROM ordenes_trabajo WITH (UPDLOCK,HOLDLOCK) WHERE id_orden=@p1`, id).Scan(&state)
	return state, translate(err)
}
func orderClosed(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var closed bool
	err := tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM entregas WHERE id_orden=@p1) THEN 1 ELSE 0 END)`, id).Scan(&closed)
	return closed, err
}
func canEditOrder(ctx context.Context, tx *sql.Tx, id string, u domain.User) error {
	if u.Can("ordenes.asignar") {
		return nil
	}
	var assigned bool
	err := tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM asignaciones_ot WHERE id_orden=@p1 AND id_usuario=@p2 AND activo=1 AND rol_en_trabajo='MECANICO') THEN 1 ELSE 0 END)`, id, u.ID).Scan(&assigned)
	if err != nil {
		return err
	}
	if !assigned {
		return domain.ErrForbidden
	}
	return nil
}
func orderEventTime(ctx context.Context, tx *sql.Tx, id string) (time.Time, error) {
	var at time.Time
	err := tx.QueryRowContext(ctx, `SELECT greatest(TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00'),(SELECT DATEADD(microsecond,1,MAX(fecha)) FROM avances_ot WHERE id_orden=@p1),(SELECT DATEADD(microsecond,1,MAX(fecha)) FROM pruebas_ot WHERE id_orden=@p1),(SELECT DATEADD(microsecond,1,MAX(fecha)) FROM historial_estados_ot WHERE id_orden=@p1))`, id).Scan(&at)
	return at, err
}
func (s *Store) UpdateOrder(ctx context.Context, id string, v domain.UpdateOrderInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if (v.MechanicID != nil) && !u.Can("ordenes.asignar") {
		return nil, domain.ErrForbidden
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
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
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM pagos WHERE id_orden=@p1 AND revertido_en IS NULL) THEN 1 ELSE 0 END)`, id).Scan(&paid); err != nil {
			return nil, err
		}
		if paid {
			return nil, domain.ErrConflict
		}
	}
	var currentMechanic string
	if err = tx.QueryRowContext(ctx, `SELECT coalesce(max(CONVERT(varchar(20),id_usuario)),'') FROM asignaciones_ot WHERE id_orden=@p1 AND activo=1 AND rol_en_trabajo='MECANICO'`, id).Scan(&currentMechanic); err != nil {
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
		if _, err = tx.ExecContext(ctx, `UPDATE asignaciones_ot SET activo=0 WHERE id_orden=@p1 AND rol_en_trabajo='MECANICO' AND activo=1`, id); err != nil {
			return nil, translate(err)
		}
		if err = insertAssignment(ctx, tx, id, *v.MechanicID, "MECANICO", at); err != nil {
			return nil, err
		}
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO avances_ot(id_orden,id_usuario,observacion,fecha) VALUES(@p1,@p2,@p3,@p4)`, id, u.ID, v.Note, at); err != nil {
		return nil, translate(err)
	}
	// Every repair update invalidates earlier tests, including repeated repair state.
	if _, err = tx.ExecContext(ctx, `INSERT INTO historial_estados_ot(id_orden,id_usuario,estado_anterior,estado_nuevo,fecha) VALUES(@p1,@p2,@p3,@p4,@p5)`, id, u.ID, state, v.State, at); err != nil {
		return nil, translate(err)
	}
	if v.Result != "" {
		if _, err = tx.ExecContext(ctx, `INSERT INTO pruebas_ot(id_orden,id_usuario,tipo_prueba,resultado,observaciones,fecha) VALUES(@p1,@p2,@p3,@p4,@p5,@p6)`, id, u.ID, v.Test, v.Result, v.Note, at); err != nil {
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
		err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,COALESCE((SELECT TOP (1) CASE WHEN p.resultado=N'satisfactoria' AND NOT EXISTS(
SELECT 1 FROM historial_estados_ot h WHERE h.id_orden=@p1 AND h.estado_nuevo=N'en_reparacion' AND h.fecha>p.fecha)
THEN 1 ELSE 0 END FROM pruebas_ot p WHERE p.id_orden=@p1 ORDER BY p.fecha DESC,p.id_prueba DESC),0))`, id).Scan(&passed)
		if err != nil {
			return nil, err
		}
		if !passed {
			return nil, domain.Invalid("result", "Finalizar exige la última prueba satisfactoria, posterior a la última reparación.")
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE ordenes_trabajo SET estado=@p2,fecha_inicio=CASE WHEN @p2 IN ('en_reparacion','en_pruebas','finalizada') THEN coalesce(fecha_inicio,@p3) ELSE fecha_inicio END,fecha_fin=CASE WHEN @p2 IN ('finalizada','cancelada') THEN @p3 ELSE NULL END WHERE id_orden=@p1`, id, v.State, at); err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+orderJSON+" FROM "+orderFrom+" WHERE o.id_orden=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) DeleteOrder(ctx context.Context, id string, u domain.User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err = lockOrder(ctx, tx, id); err != nil {
		return err
	}
	var used bool
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM pagos WHERE id_orden=@p1) OR EXISTS(SELECT 1 FROM entregas WHERE id_orden=@p1) OR EXISTS(SELECT 1 FROM avances_ot WHERE id_orden=@p1) OR EXISTS(SELECT 1 FROM pruebas_ot WHERE id_orden=@p1) THEN 1 ELSE 0 END)`, id).Scan(&used); err != nil {
		return err
	}
	if used {
		return domain.ErrConflict
	}
	for _, table := range []string{"historial_estados_ot", "asignaciones_ot", "ordenes_trabajo"} {
		if _, err = tx.ExecContext(ctx, "DELETE FROM "+table+" WHERE id_orden=@p1", id); err != nil {
			return translate(err)
		}
	}
	return translate(tx.Commit())
}
func (s *Store) RegisterPayment(ctx context.Context, id string, v domain.PaymentInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	state, err := lockOrder(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	closed, err := orderClosed(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	var replay []byte
	replayErr := tx.QueryRowContext(ctx, `SELECT `+paymentJSON+` FROM pagos p WHERE id_orden=@p1 AND request_id=@p2`, id, v.RequestID).Scan(&replay)
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
	if !errors.Is(replayErr, sql.ErrNoRows) {
		return nil, translate(replayErr)
	}
	if closed || state == "cancelada" {
		return nil, domain.Invalid("orderId", "No se pagan órdenes canceladas o ya entregadas.")
	}
	var total, paid string
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(42),dbo.jac_importe_orden(o.id_orden)),CONVERT(varchar(32),CAST(COALESCE((SELECT sum(monto) FROM pagos WHERE id_orden=@p1 AND revertido_en IS NULL),0) AS decimal(14,2))) FROM cotizaciones q JOIN ordenes_trabajo o ON o.id_cotizacion=q.id_cotizacion WHERE o.id_orden=@p1`, id).Scan(&total, &paid); err != nil {
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
	out, err := readJSON(ctx, tx, `DECLARE @id bigint = NEXT VALUE FOR dbo.pagos_id_pago_seq;
INSERT INTO pagos(id_pago,id_orden,monto,metodo_pago,comprobante,id_usuario_registro,request_id)
VALUES(@id,@p1,@p2,@p3,NULLIF(@p4,N''),@p5,@p6);
SELECT `+paymentJSON+` FROM dbo.pagos p WHERE p.id_pago=@id;`, id, v.Amount, v.Method, v.Receipt, u.ID, v.RequestID)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) Deliver(ctx context.Context, id string, v domain.DeliveryInput, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
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
	err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),r.id_cliente_atencion),CONVERT(bit,CASE WHEN dbo.jac_importe_orden(o.id_orden)=coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=@p1 AND revertido_en IS NULL),0) THEN 1 ELSE 0 END) FROM ordenes_trabajo o JOIN cotizaciones q ON q.id_cotizacion=o.id_cotizacion JOIN recepciones r ON r.id_recepcion=q.id_recepcion WHERE o.id_orden=@p1`, id).Scan(&clientID, &settled)
	if err != nil {
		return nil, translate(err)
	}
	if !settled {
		return nil, domain.Invalid("balance", "Completa el pago exacto antes de entregar.")
	}
	var unfinished bool
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM ordenes_trabajo x JOIN cotizaciones xq ON xq.id_cotizacion=x.id_cotizacion WHERE xq.id_recepcion=(SELECT q.id_recepcion FROM ordenes_trabajo z JOIN cotizaciones q ON q.id_cotizacion=z.id_cotizacion WHERE z.id_orden=@p1) AND x.id_orden<>@p1 AND x.estado<>'cancelada' AND (x.estado<>'finalizada' OR dbo.jac_importe_orden(x.id_orden)<>coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=x.id_orden AND revertido_en IS NULL),0))) OR EXISTS(SELECT 1 FROM cotizaciones xq WHERE xq.id_recepcion=(SELECT q.id_recepcion FROM ordenes_trabajo z JOIN cotizaciones q ON q.id_cotizacion=z.id_cotizacion WHERE z.id_orden=@p1) AND xq.id_cotizacion_base IS NOT NULL AND xq.estado='aprobada' AND NOT EXISTS(SELECT 1 FROM ordenes_trabajo x WHERE x.id_cotizacion=xq.id_cotizacion)) THEN 1 ELSE 0 END)`, id).Scan(&unfinished); err != nil {
		return nil, err
	}
	if unfinished {
		return nil, domain.Invalid("attention", "Completa y liquida todas las órdenes y ampliaciones aprobadas de esta atención antes de entregar el vehículo.")
	}
	var name, docType, doc, phone string
	if v.Kind == "Tercero" {
		err = tx.QueryRowContext(ctx, `SELECT nombre,tipo_documento,numero_documento,coalesce(telefono,'') FROM contactos_cliente WITH (HOLDLOCK) WHERE id_contacto=@p1 AND id_cliente=@p2 AND activo=1`, v.ContactID, clientID).Scan(&name, &docType, &doc, &phone)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.Invalid("contactId", "El contacto activo debe pertenecer al cliente de la atención.")
		}
	} else {
		err = tx.QueryRowContext(ctx, `SELECT nombre,tipo_documento,numero_documento,coalesce(telefono,'') FROM clientes WITH (HOLDLOCK) WHERE id_cliente=@p1`, clientID).Scan(&name, &docType, &doc, &phone)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, `INSERT INTO entregas(id_orden,id_usuario_entrega,id_contacto_receptor,tipo_receptor,nombre_receptor_snap,tipo_doc_receptor_snap,
documento_receptor_snap,telefono_receptor_snap,autorizacion_verificada,medio_verificacion,fecha_verificacion,id_usuario_verificador)
VALUES(@p1,@p2,NULLIF(@p3,N''),@p4,@p5,@p6,@p7,@p8,1,@p9,TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00'),@p2);
SELECT `+deliveryJSON+` FROM dbo.entregas e WHERE e.id_orden=@p1;`, id, u.ID, v.ContactID, v.Kind, name, docType, doc, phone, v.Method)
	if err != nil {
		return nil, err
	}
	if v.RecipientKind == "Propietario o titular" && docType == "RUC" {
		return nil, domain.Invalid("recipientKind", "Una empresa requiere identificar a su representante físico mediante un contacto.")
	}
	receiver, _ := json.Marshal(map[string]any{"kind": v.RecipientKind, "name": name, "documentType": docType, "documentNumber": doc, "phone": phone})
	conformity, _ := json.Marshal(map[string]any{"accepted": v.Conformity, "notes": v.ConformityNotes, "recordedBy": u.ID})
	if _, err = tx.ExecContext(ctx, `UPDATE entregas SET receptor_real=@p2,conformidad=JSON_MODIFY(@p3,'$.at',CONVERT(varchar(40),TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00'),127)) WHERE id_orden=@p1`, id, string(receiver), string(conformity)); err != nil {
		return nil, translate(err)
	}
	var snapshot []byte
	if err = tx.QueryRowContext(ctx, sheetSnapshotSQL, id).Scan(&snapshot); err != nil {
		return nil, translate(err)
	}
	// Freeze all data even when unusually long content exceeds the fixed template.
	// Printing must not invalidate an otherwise valid vehicle delivery. Downloads
	// report a precise 422 instead of cropping; the UI exposes sheetReady=false.
	pdf, renderErr := technicalsheet.Render(snapshot)
	if renderErr != nil {
		pdf = nil
	}
	if _, err = tx.ExecContext(ctx, `UPDATE entregas SET ficha_snapshot=@p2,ficha_pdf=@p3 WHERE id_orden=@p1`, id, string(snapshot), pdf); err != nil {
		return nil, translate(err)
	}
	out, err = readJSON(ctx, tx, "SELECT "+deliveryJSON+" FROM entregas e WHERE e.id_orden=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
