package sqlserver

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/platform/technicalsheet"
	"time"
)

const sheetSnapshotSQL = `SELECT JSON_OBJECT(
'workshop':JSON_QUERY((SELECT * FROM dbo.datos_taller_ficha FOR JSON PATH, WITHOUT_ARRAY_WRAPPER)),
'version':N'diseno19-v2','code':o.codigo_ot,'client':cl.nombre,'document':CONCAT(cl.tipo_documento,N' ',cl.numero_documento),
'phone':COALESCE(cl.telefono,N''),'address':COALESCE(cl.direccion,N''),'email':COALESCE(cl.correo,N''),
'plate':v.placa,'vehicle':v.marca+N' '+v.modelo+COALESCE(N' · '+CONVERT(varchar(10),v.anio),N''),
'mileage':CONVERT(varchar(32),r.km_ingreso),'symptoms':r.sintomas,'in':r.fecha_ingreso,'out':e.fecha_entrega,
'mechanic':COALESCE((SELECT TOP (1) u.nombre FROM dbo.asignaciones_ot a JOIN dbo.usuarios u ON u.id_usuario=a.id_usuario
 WHERE a.id_orden=o.id_orden AND a.activo=1 AND a.rol_en_trabajo=N'MECANICO' ORDER BY a.id_asignacion DESC),N'No registrado'),
'authorizedTotal':CONVERT(varchar(32),q.total),'total':CONVERT(varchar(42),dbo.jac_importe_orden(o.id_orden)),
'lines':JSON_QUERY((SELECT CONVERT(varchar(20),dc.id_detalle) AS id,
 COALESCE(dc.tipo_concepto,CASE WHEN dc.id_servicio IS NOT NULL THEN N'Servicio' ELSE N'Repuesto' END) AS type,
 JSON_VALUE(o.control_tecnico,CONCAT(N'$.execution."',dc.id_detalle,N'".notes')) AS notes,dc.descripcion AS concept,
 CONVERT(varchar(32),dc.cantidad) AS quantity,CONVERT(varchar(32),dc.precio_unitario) AS unitPrice,
 CONVERT(varchar(32),dc.descuento_unitario) AS discount,CONVERT(varchar(32),dc.subtotal) AS subtotal
 FROM dbo.detalle_cotizacion dc WHERE dc.id_cotizacion=q.id_cotizacion AND
 (JSON_VALUE(o.control_tecnico,CONCAT(N'$.execution."',dc.id_detalle,N'".status'))=N'Realizado'
  OR EXISTS(SELECT 1 FROM OPENJSON(o.control_tecnico,N'$.performedItems') x WHERE x.value=CONVERT(varchar(20),dc.id_detalle)))
 ORDER BY dc.id_detalle FOR JSON PATH, INCLUDE_NULL_VALUES)),
'diagnosis':COALESCE(d.resumen,N''),'evaluation':COALESCE(d.evaluacion,N''),'recommendations':COALESCE(d.recomendaciones,N''),
'diagnosisNotes':COALESCE(d.observaciones,N''),
'tests':JSON_QUERY((SELECT pt.tipo_prueba AS kind,pt.resultado AS result,COALESCE(pt.observaciones,N'') AS notes FROM dbo.pruebas_ot pt
 WHERE pt.id_orden=o.id_orden ORDER BY pt.id_prueba FOR JSON PATH, INCLUDE_NULL_VALUES)),
'advances':JSON_QUERY(COALESCE((SELECT JSON_ARRAYAGG(a.observacion ORDER BY a.id_avance) FROM dbo.avances_ot a WHERE a.id_orden=o.id_orden),N'[]')),
'control':JSON_QUERY(o.control_tecnico)) FROM ` + orderFrom + ` JOIN dbo.entregas e ON e.id_orden=o.id_orden WHERE o.id_orden=@p1`

const assignedMechanic = `EXISTS(SELECT 1 FROM dbo.asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.activo=1
 AND a.rol_en_trabajo=N'MECANICO' AND a.id_usuario=@p3)`

func (s *Store) Inspection(ctx context.Context, id string, u domain.User) (any, error) {
	return readJSON(ctx, s.db, `SELECT JSON_OBJECT('control':JSON_QUERY(o.control_tecnico),'version':o.control_version,
'authorId':CONVERT(varchar(20),o.control_autor),'updatedAt':o.control_fecha,
'readOnly':CONVERT(bit,CASE WHEN o.estado IN(N'finalizada',N'cancelada') OR EXISTS(SELECT 1 FROM dbo.entregas e WHERE e.id_orden=o.id_orden) THEN 1 ELSE 0 END),
'canReconcile':CONVERT(bit,@p6),'canEdit':CONVERT(bit,CASE WHEN @p5=1 AND (@p2=1 OR `+assignedMechanic+`) THEN 1 ELSE 0 END))
FROM dbo.ordenes_trabajo o WHERE o.id_orden=@p1 AND (@p4=1 OR EXISTS(
 SELECT 1 FROM dbo.asignaciones_ot a WHERE a.id_orden=o.id_orden AND a.activo=1 AND a.id_usuario=@p3))`,
		id, u.Can("ordenes.asignar"), u.ID, orderVisible(u), u.Can("ordenes.editar"), u.Can("cotizaciones.editar"))
}

func (s *Store) SaveInspection(ctx context.Context, id string, v technicalsheet.Control, u domain.User) (any, error) {
	if err := v.Validate(); err != nil {
		return nil, domain.Invalid("control", err.Error())
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var state string
	var version int64
	var allowed bool
	err = tx.QueryRowContext(ctx, `SELECT estado,control_version,CONVERT(bit,CASE WHEN @p2=1 OR `+assignedMechanic+` THEN 1 ELSE 0 END)
FROM dbo.ordenes_trabajo o WITH (UPDLOCK,HOLDLOCK) WHERE id_orden=@p1`, id, u.Can("ordenes.asignar"), u.ID).Scan(&state, &version, &allowed)
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
		performed, _ := json.Marshal(v.PerformedItems)
		var count int
		err = tx.QueryRowContext(ctx, `SELECT count(*) FROM detalle_cotizacion dc JOIN ordenes_trabajo o ON o.id_cotizacion=dc.id_cotizacion WHERE o.id_orden=@p1 AND CONVERT(varchar(20),dc.id_detalle) IN(SELECT value FROM OPENJSON(@p2))`, id, string(performed)).Scan(&count)
		if err != nil {
			return nil, err
		}
		if count != len(v.PerformedItems) {
			return nil, domain.Invalid("performedItems", "Selecciona trabajos de la cotización autorizada sin duplicarlos.")
		}
	}
	var oldRaw []byte
	if err = tx.QueryRowContext(ctx, `SELECT control_tecnico FROM ordenes_trabajo WHERE id_orden=@p1`, id).Scan(&oldRaw); err != nil {
		return nil, err
	}
	var old technicalsheet.Control
	json.Unmarshal(oldRaw, &old)
	var validIDs []string
	rows, err := tx.QueryContext(ctx, `SELECT CONVERT(varchar(20),dc.id_detalle) FROM detalle_cotizacion dc JOIN ordenes_trabajo o ON o.id_cotizacion=dc.id_cotizacion WHERE o.id_orden=@p1`, id)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var key string
		if err = rows.Scan(&key); err != nil {
			rows.Close()
			return nil, err
		}
		validIDs = append(validIDs, key)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}
	if v.Execution == nil {
		v.Execution = old.Execution
		if len(v.PerformedItems) > 0 {
			v.Execution = map[string]technicalsheet.Execution{}
			for _, key := range validIDs {
				v.Execution[key] = technicalsheet.Execution{Status: "Pendiente"}
			}
			for _, key := range v.PerformedItems {
				v.Execution[key] = technicalsheet.Execution{Status: "Realizado"}
			}
		}
	}
	allowedIDs := map[string]bool{}
	for _, key := range validIDs {
		allowedIDs[key] = true
	}
	for key := range v.Execution {
		if !allowedIDs[key] {
			return nil, domain.Invalid("execution", "Trabajo ajeno a la cotización autorizada.")
		}
	}
	v.PerformedItems = nil
	for _, key := range validIDs {
		if v.Execution[key].Status == "Realizado" {
			v.PerformedItems = append(v.PerformedItems, key)
		}
	}
	rawExecution, _ := json.Marshal(v.Execution)
	hash := fmt.Sprintf("%x", sha256.Sum256(rawExecution))
	v.Reconciliation = nil
	if old.Reconciliation != nil && old.Reconciliation.Hash == hash {
		v.Reconciliation = old.Reconciliation
	}
	if v.AdjustmentReason != "" {
		if !u.Can("cotizaciones.editar") {
			return nil, domain.ErrForbidden
		}
		if len(v.AdjustmentReason) > 1200 {
			return nil, domain.Invalid("adjustmentReason", "Máximo 1200 caracteres.")
		}
		v.Reconciliation = &technicalsheet.Reconciliation{By: u.ID, At: time.Now().UTC().Format(time.RFC3339Nano), Hash: hash, Reason: v.AdjustmentReason}
	}
	v.AdjustmentReason = ""
	v.Version = version + 1
	raw, _ := json.Marshal(v)
	_, err = tx.ExecContext(ctx, `UPDATE ordenes_trabajo SET control_historial=JSON_MODIFY(control_historial,'append $',JSON_QUERY(JSON_OBJECT('previous':JSON_QUERY(control_tecnico),'version':control_version,'by':@p4,'at':SYSUTCDATETIME()))),control_tecnico=@p2,control_version=@p3,control_autor=@p4,control_fecha=TODATETIMEOFFSET(SYSUTCDATETIME(),'+00:00') WHERE id_orden=@p1`, id, string(raw), v.Version, u.ID)
	if err != nil {
		return nil, translate(err)
	}
	if err = tx.Commit(); err != nil {
		return nil, translate(err)
	}
	return s.Inspection(ctx, id, u)
}
func (s *Store) SheetPDF(ctx context.Context, id string, u domain.User) ([]byte, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	var snapshot, pdf []byte
	err = tx.QueryRowContext(ctx, `SELECT e.ficha_snapshot,e.ficha_pdf FROM dbo.entregas e WITH (UPDLOCK,HOLDLOCK) JOIN dbo.ordenes_trabajo o ON o.id_orden=e.id_orden JOIN dbo.cotizaciones q ON q.id_cotizacion=o.id_cotizacion WHERE e.id_orden=@p1 AND o.estado='finalizada' AND q.estado='aprobada' AND dbo.jac_importe_orden(o.id_orden)=coalesce((SELECT sum(monto) FROM pagos WHERE id_orden=o.id_orden AND revertido_en IS NULL),0)`, id).Scan(&snapshot, &pdf)
	if err != nil {
		return nil, translate(err)
	}
	if len(pdf) > 0 {
		// Issued documents are immutable, even when a snapshot is available.
		// Only deliveries without an archived PDF enter the compatibility renderer.
		return pdf, nil
	}
	// Legacy deliveries are reconstructed once using available data, then frozen.
	if len(snapshot) == 0 {
		if err = tx.QueryRowContext(ctx, sheetSnapshotSQL, id).Scan(&snapshot); err != nil {
			return nil, translate(err)
		}
	}
	pdf, err = technicalsheet.Render(snapshot)
	if err != nil {
		return nil, domain.Invalid("ficha", err.Error())
	}
	_, err = tx.ExecContext(ctx, `UPDATE entregas SET ficha_snapshot=@p2,ficha_pdf=@p3 WHERE id_orden=@p1`, id, string(snapshot), pdf)
	if err != nil {
		return nil, translate(err)
	}
	return pdf, translate(tx.Commit())
}
