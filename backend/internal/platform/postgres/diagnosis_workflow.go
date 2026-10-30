package postgres

import (
	"context"
	"encoding/json"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

func diagnosisSupervisor(u domain.User) bool {
	return u.Can("diagnosticos.supervisar") || u.Can("diagnosticos.asignar")
}
func receptionGlobal(u domain.User) bool {
	return diagnosisSupervisor(u) || u.Can("recepciones.crear") || u.Can("cotizaciones.crear")
}

func (s *Store) AssignDiagnosis(ctx context.Context, id string, v domain.DiagnosisAssignment, u domain.User) (any, error) {
	if !u.Can("diagnosticos.asignar") {
		return nil, domain.ErrForbidden
	}
	if _, err := domain.ID(v.TechnicianID); err != nil {
		return nil, domain.Invalid("technicianId", "Selecciona un mecánico válido.")
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	var attention, current string
	var version int64
	if err = tx.QueryRow(ctx, `SELECT tipo_atencion,coalesce(id_mecanico_diagnostico::text,''),asignacion_version FROM recepciones WHERE id_recepcion=$1 FOR UPDATE`, id).Scan(&attention, &current, &version); err != nil {
		return nil, translate(err)
	}
	if attention != "Diagnostico" {
		return nil, domain.Invalid("attentionType", "El servicio directo no requiere asignación de diagnóstico.")
	}
	if err = technicalUser(ctx, tx, v.TechnicianID); err != nil {
		return nil, err
	}
	var finalized bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM diagnosticos WHERE id_recepcion=$1 AND estado='Finalizado')`, id).Scan(&finalized); err != nil {
		return nil, err
	}
	if finalized {
		return nil, domain.ErrConflict
	}
	if current != v.TechnicianID {
		if version != v.Version {
			return nil, domain.ErrConflict
		}
		_, err = tx.Exec(ctx, `UPDATE recepciones SET id_mecanico_diagnostico=$2::bigint,diagnostico_asignado_por=$3::bigint,diagnostico_asignado_en=clock_timestamp(),asignacion_version=asignacion_version+1,historial_asignaciones=historial_asignaciones||jsonb_build_array(jsonb_build_object('previousId',id_mecanico_diagnostico::text,'technicianId',$2::text,'by',$3::text,'at',clock_timestamp())) WHERE id_recepcion=$1`, id, v.TechnicianID, u.ID)
		if err != nil {
			return nil, translate(err)
		}
		_, err = tx.Exec(ctx, `UPDATE diagnosticos SET id_tecnico=$2::bigint,version=version+1,actualizado_en=clock_timestamp(),historial=historial||jsonb_build_array(jsonb_build_object('event','Reasignación','by',$3::text,'at',clock_timestamp(),'previous',(to_jsonb(diagnosticos)-'historial'))) WHERE id_recepcion=$1 AND estado='Borrador'`, id, v.TechnicianID, u.ID)
		if err != nil {
			return nil, translate(err)
		}
	}
	out, err := readJSON(ctx, tx, "SELECT "+receptionJSON+" FROM recepciones r WHERE r.id_recepcion=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) SaveDiagnosis(ctx context.Context, id string, v domain.DiagnosisInput, u domain.User) (any, error) {
	if v.Status == "" {
		v.Status = "Borrador"
	}
	if err := v.Validate(); err != nil {
		return nil, err
	}
	if v.Status == "" {
		v.Status = "Borrador"
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	// Always lock the reception before the diagnosis, matching reassignment order.
	if id != "" {
		var reception string
		if err = tx.QueryRow(ctx, `SELECT id_recepcion::text FROM diagnosticos WHERE id_diagnostico=$1`, id).Scan(&reception); err != nil {
			return nil, translate(err)
		}
		if reception != v.ReceptionID {
			return nil, domain.Invalid("receptionId", "No se puede cambiar la recepción.")
		}
	}
	var attention, technician string
	err = tx.QueryRow(ctx, `SELECT tipo_atencion,coalesce(id_mecanico_diagnostico::text,'') FROM recepciones WHERE id_recepcion=$1 FOR UPDATE`, v.ReceptionID).Scan(&attention, &technician)
	if err != nil {
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
	event, _ := json.Marshal(map[string]any{"by": u.ID, "action": v.Status, "findings": v.Findings, "evaluation": v.Evaluation, "recommendations": v.Recommendations, "notes": v.Notes})
	if id == "" {
		err = tx.QueryRow(ctx, `INSERT INTO diagnosticos(id_recepcion,id_tecnico,resumen,evaluacion,recomendaciones,observaciones,estado,version,actualizado_en,finalizado_en,historial) VALUES($1,$2,$3,$4,$5,$6,$7::text,1,clock_timestamp(),CASE WHEN $7::text='Finalizado' THEN clock_timestamp() END,jsonb_build_array($8::jsonb||jsonb_build_object('at',clock_timestamp()))) RETURNING id_diagnostico::text`, v.ReceptionID, technician, v.Findings, v.Evaluation, v.Recommendations, v.Notes, v.Status, event).Scan(&id)
	} else {
		var version int64
		var state string
		if err = tx.QueryRow(ctx, `SELECT version,estado FROM diagnosticos WHERE id_diagnostico=$1 FOR UPDATE`, id).Scan(&version, &state); err != nil {
			return nil, translate(err)
		}
		if state == "Finalizado" || version != v.Version {
			return nil, domain.ErrConflict
		}
		_, err = tx.Exec(ctx, `UPDATE diagnosticos SET resumen=$2,evaluacion=$3,recomendaciones=$4,observaciones=$5,estado=$6::text,version=version+1,actualizado_en=clock_timestamp(),finalizado_en=CASE WHEN $6::text='Finalizado' THEN clock_timestamp() END,historial=historial||jsonb_build_array($7::jsonb||jsonb_build_object('at',clock_timestamp())) WHERE id_diagnostico=$1`, id, v.Findings, v.Evaluation, v.Recommendations, v.Notes, v.Status, event)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+diagnosisJSON+" FROM diagnosticos d WHERE d.id_diagnostico=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) DiagnosisQueue(ctx context.Context, ready bool, p domain.Page, u domain.User) (any, error) {
	where := `r.tipo_atencion='Diagnostico' AND (r.id_mecanico_diagnostico::text=$1 OR $2) AND coalesce(d.estado,'Pendiente')<>'Finalizado'`
	if ready {
		where = `d.estado='Finalizado' AND NOT EXISTS(SELECT 1 FROM cotizaciones q WHERE q.id_diagnostico=d.id_diagnostico) AND (r.id_mecanico_diagnostico::text=$1 OR $2)`
	}
	from := `recepciones r LEFT JOIN diagnosticos d USING(id_recepcion)`
	out := domain.List[json.RawMessage]{Data: []json.RawMessage{}, Meta: domain.Meta{Page: p.Number, PageSize: p.Size}}
	tx, err := s.Pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	args := []any{u.ID, diagnosisSupervisor(u)}
	if err = tx.QueryRow(ctx, "SELECT count(*) FROM "+from+" WHERE "+where, args...).Scan(&out.Meta.Total); err != nil {
		return nil, translate(err)
	}
	rows, err := tx.Query(ctx, `SELECT jsonb_build_object('reception',`+receptionJSON+`,'diagnosis',CASE WHEN d.id_diagnostico IS NULL THEN NULL ELSE `+diagnosisJSON+` END) FROM `+from+` WHERE `+where+` ORDER BY r.id_recepcion DESC LIMIT $3 OFFSET $4`, u.ID, diagnosisSupervisor(u), p.Size, (p.Number-1)*p.Size)
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
