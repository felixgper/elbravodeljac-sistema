package sqlserver

import (
	"context"
	"encoding/json"
	"time"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

func (s *Store) AssignDiagnosis(ctx context.Context, id string, v domain.DiagnosisAssignment, u domain.User) (any, error) {
	if !u.Can("diagnosticos.asignar") {
		return nil, domain.ErrForbidden
	}
	if _, err := domain.ID(v.TechnicianID); err != nil {
		return nil, domain.Invalid("technicianId", "Selecciona un mecánico válido.")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, translate(err)
	}
	defer tx.Rollback()
	var attention, current string
	var version int64
	if err = tx.QueryRowContext(ctx, `SELECT tipo_atencion,COALESCE(CONVERT(varchar(20),id_mecanico_diagnostico),''),asignacion_version
FROM dbo.recepciones WITH (UPDLOCK,HOLDLOCK) WHERE id_recepcion=@p1`, id).Scan(&attention, &current, &version); err != nil {
		return nil, translate(err)
	}
	if attention != "Diagnostico" {
		return nil, domain.Invalid("attentionType", "El servicio directo no requiere asignación de diagnóstico.")
	}
	if err = technicalUser(ctx, tx, v.TechnicianID); err != nil {
		return nil, err
	}
	var finalized bool
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(
SELECT 1 FROM dbo.diagnosticos WHERE id_recepcion=@p1 AND estado=N'Finalizado') THEN 1 ELSE 0 END)`, id).Scan(&finalized); err != nil {
		return nil, translate(err)
	}
	if finalized {
		return nil, domain.ErrConflict
	}
	if current != v.TechnicianID {
		if version != v.Version {
			return nil, domain.ErrConflict
		}
		at := time.Now().UTC()
		event := map[string]any{"previousId": nil, "technicianId": v.TechnicianID, "by": u.ID, "at": at}
		if current != "" {
			event["previousId"] = current
		}
		history, _ := json.Marshal(event)
		if _, err = tx.ExecContext(ctx, `UPDATE dbo.recepciones SET id_mecanico_diagnostico=@p2,diagnostico_asignado_por=@p3,
diagnostico_asignado_en=@p4,asignacion_version=asignacion_version+1,
historial_asignaciones=JSON_MODIFY(historial_asignaciones,'append $',JSON_QUERY(@p5))
WHERE id_recepcion=@p1`, id, v.TechnicianID, u.ID, at, string(history)); err != nil {
			return nil, translate(err)
		}
		if _, err = tx.ExecContext(ctx, `UPDATE d SET id_tecnico=@p2,version=version+1,actualizado_en=@p4,
historial=JSON_MODIFY(d.historial,'append $',JSON_QUERY(JSON_OBJECT('event':N'Reasignación','by':@p3,'at':@p4,
'previous':JSON_QUERY((SELECT d.id_diagnostico,d.id_recepcion,d.id_tecnico,d.fecha,d.resumen,d.estado,d.evaluacion,
d.recomendaciones,d.observaciones,d.version,d.finalizado_en,d.actualizado_en FOR JSON PATH, INCLUDE_NULL_VALUES, WITHOUT_ARRAY_WRAPPER)))))
FROM dbo.diagnosticos d WHERE d.id_recepcion=@p1 AND d.estado=N'Borrador'`, id, v.TechnicianID, u.ID, at); err != nil {
			return nil, translate(err)
		}
	}
	out, err := readJSON(ctx, tx, "SELECT "+receptionJSON+" FROM dbo.recepciones r WHERE r.id_recepcion=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}

func (s *Store) CreateDiagnosis(ctx context.Context, v domain.DiagnosisInput, u domain.User) (any, error) {
	return s.SaveDiagnosis(ctx, "", v, u)
}

func (s *Store) SaveDiagnosis(ctx context.Context, id string, v domain.DiagnosisInput, u domain.User) (any, error) {
	if v.Status == "" {
		v.Status = "Borrador"
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if v.TechnicalDiagnosis != nil {
		v.Findings = *v.TechnicalDiagnosis
		v.Evaluation = ""
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, translate(err)
	}
	defer tx.Rollback()
	if id != "" {
		var reception string
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),id_recepcion) FROM dbo.diagnosticos WHERE id_diagnostico=@p1`, id).Scan(&reception); err != nil {
			return nil, translate(err)
		}
		if reception != v.ReceptionID {
			return nil, domain.Invalid("receptionId", "No se puede cambiar la recepción.")
		}
	}
	var attention, technician string
	if err = tx.QueryRowContext(ctx, `SELECT tipo_atencion,COALESCE(CONVERT(varchar(20),id_mecanico_diagnostico),'')
FROM dbo.recepciones WITH (UPDLOCK,HOLDLOCK) WHERE id_recepcion=@p1`, v.ReceptionID).Scan(&attention, &technician); err != nil {
		return nil, translate(err)
	}
	if attention != "Diagnostico" {
		return nil, domain.Invalid("attentionType", "Servicio directo sin diagnóstico.")
	}
	if technician == "" {
		return nil, domain.Invalid("technicianId", "Asigna un mecánico a la recepción antes de registrar el diagnóstico.")
	}
	if technician != u.ID && !diagnosisSupervisor(u) {
		return nil, domain.ErrForbidden
	}
	if v.TechnicianID != "" && v.TechnicianID != technician {
		return nil, domain.Invalid("technicianId", "El responsable corresponde a la asignación de recepción.")
	}
	if err = technicalUser(ctx, tx, technician); err != nil {
		return nil, err
	}

	at := time.Now().UTC()
	event := map[string]any{"by": u.ID, "action": v.Status, "findings": v.Findings, "evaluation": v.Evaluation, "recommendations": v.Recommendations, "notes": v.Notes, "at": at}
	if id == "" {
		raw, _ := json.Marshal([]any{event})
		err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.diagnosticos_id_diagnostico_seq;
INSERT dbo.diagnosticos(id_diagnostico,id_recepcion,id_tecnico,resumen,evaluacion,recomendaciones,observaciones,estado,
version,actualizado_en,finalizado_en,historial) VALUES(@id,@p1,@p2,@p3,@p4,@p5,@p6,@p7,1,@p8,CASE WHEN @p7=N'Finalizado' THEN @p8 END,@p9);
SELECT CONVERT(varchar(20),@id);`, v.ReceptionID, technician, v.Findings, v.Evaluation, v.Recommendations, v.Notes, v.Status, at, string(raw)).Scan(&id)
	} else {
		var version int64
		var state, findings, evaluation, recommendations, notes string
		if err = tx.QueryRowContext(ctx, `SELECT version,estado,resumen,evaluacion,recomendaciones,observaciones
FROM dbo.diagnosticos WITH (UPDLOCK,HOLDLOCK) WHERE id_diagnostico=@p1`, id).Scan(&version, &state, &findings, &evaluation, &recommendations, &notes); err != nil {
			return nil, translate(err)
		}
		if state == "Finalizado" || version != v.Version {
			return nil, domain.ErrConflict
		}
		event["previous"] = map[string]any{"version": version, "status": state, "findings": findings, "evaluation": evaluation, "recommendations": recommendations, "notes": notes}
		raw, _ := json.Marshal(event)
		_, err = tx.ExecContext(ctx, `UPDATE dbo.diagnosticos SET resumen=@p2,evaluacion=@p3,recomendaciones=@p4,observaciones=@p5,
estado=@p6,version=version+1,actualizado_en=@p7,finalizado_en=CASE WHEN @p6=N'Finalizado' THEN @p7 END,
historial=JSON_MODIFY(historial,'append $',JSON_QUERY(@p8)) WHERE id_diagnostico=@p1`, id, v.Findings, v.Evaluation, v.Recommendations, v.Notes, v.Status, at, string(raw))
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+diagnosisJSON+" FROM dbo.diagnosticos d WHERE d.id_diagnostico=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}

func (s *Store) DiagnosisQueue(ctx context.Context, ready bool, p domain.Page, u domain.User) (any, error) {
	where := `r.tipo_atencion=N'Diagnostico' AND (r.id_mecanico_diagnostico=@p1 OR @p2=1) AND COALESCE(d.estado,N'Pendiente')<>N'Finalizado'`
	if ready {
		where = `d.estado=N'Finalizado' AND NOT EXISTS(SELECT 1 FROM dbo.cotizaciones q WHERE q.id_diagnostico=d.id_diagnostico)
AND (r.id_mecanico_diagnostico=@p1 OR @p2=1)`
	}
	return s.jsonPage(ctx, p, "JSON_OBJECT('reception':JSON_QUERY("+receptionJSON+"),'diagnosis':JSON_QUERY(CASE WHEN d.id_diagnostico IS NULL THEN NULL ELSE "+diagnosisJSON+" END))",
		"dbo.recepciones r LEFT JOIN dbo.diagnosticos d ON d.id_recepcion=r.id_recepcion", where, "r.id_recepcion DESC", u.ID, diagnosisSupervisor(u))
}
