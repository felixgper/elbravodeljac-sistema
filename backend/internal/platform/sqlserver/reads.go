package sqlserver

import (
	"context"
	"encoding/json"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

func readJSON(ctx context.Context, q queryer, statement string, args ...any) (any, error) {
	var payload []byte
	if err := q.QueryRowContext(ctx, statement, args...).Scan(&payload); err != nil {
		return nil, translate(err)
	}
	return json.RawMessage(payload), nil
}

func (s *Store) Brands(ctx context.Context) ([]string, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT nombre FROM dbo.marcas_vehiculos
ORDER BY nombre COLLATE Latin1_General_100_CI_AI_SC,nombre`)
	if err != nil {
		return nil, translate(err)
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, translate(err)
		}
		out = append(out, name)
	}
	return out, translate(rows.Err())
}

func (s *Store) DocumentTypes(ctx context.Context) (any, error) {
	return readJSON(ctx, s.db, `SELECT (SELECT CONVERT(varchar(20),id_tipo_documento) AS id,
codigo AS code,nombre AS name,activo AS active,
CASE codigo WHEN N'PASAPORTE' THEN N'Pasaporte' WHEN N'DNI' THEN N'DNI' WHEN N'RUC' THEN N'RUC'
            WHEN N'CE' THEN N'CE' ELSE nombre END AS documentType,
CASE codigo WHEN N'CE' THEN N'Carné de Extranjería' WHEN N'PASAPORTE' THEN N'Pasaporte' ELSE codigo END AS label
FROM dbo.tipos_documento
ORDER BY CASE codigo WHEN N'DNI' THEN 0 WHEN N'RUC' THEN 1 WHEN N'CE' THEN 2 WHEN N'PASAPORTE' THEN 3 ELSE 4 END,id_tipo_documento
FOR JSON PATH, INCLUDE_NULL_VALUES)`)
}

func (s *Store) VehicleHistory(ctx context.Context, id string) (any, error) {
	return readJSON(ctx, s.db, `SELECT (
SELECT CONVERT(varchar(20),v.id_vehiculo) AS vehicleId,
    JSON_QUERY((SELECT CONVERT(varchar(20),h.id_historial) AS id,CONVERT(varchar(20),h.id_cliente) AS clientId,
        h.fecha_inicio AS [from],h.fecha_fin AS [until],h.verificado AS verified,
        CONVERT(varchar(20),h.id_usuario_registro) AS actorId,h.observacion AS observation
        FROM dbo.historial_propietarios h WHERE h.id_vehiculo=v.id_vehiculo
        ORDER BY h.fecha_inicio,h.id_historial FOR JSON PATH, INCLUDE_NULL_VALUES)) AS transfers,
    JSON_QUERY((SELECT CONVERT(varchar(20),r.id_recepcion) AS id,
        CONVERT(varchar(20),r.id_cliente_atencion) AS clientId,r.fecha_ingreso AS createdAt,
        r.sintomas AS symptoms,r.tipo_atencion AS attentionType,
        CONVERT(varchar(20),d.id_diagnostico) AS diagnosisId,d.resumen AS findings,
        JSON_QUERY((SELECT CONVERT(varchar(20),c.id_cotizacion) AS id,CONVERT(varchar(20),o.id_orden) AS orderId,
            o.codigo_ot AS code,o.estado AS state,
            CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM dbo.entregas e WHERE e.id_orden=o.id_orden) THEN 1 ELSE 0 END) AS delivered
            FROM dbo.cotizaciones c LEFT JOIN dbo.ordenes_trabajo o ON o.id_cotizacion=c.id_cotizacion
            WHERE c.id_recepcion=r.id_recepcion ORDER BY c.id_cotizacion FOR JSON PATH, INCLUDE_NULL_VALUES)) AS quotes
        FROM dbo.recepciones r LEFT JOIN dbo.diagnosticos d ON d.id_recepcion=r.id_recepcion
        WHERE r.id_vehiculo=v.id_vehiculo ORDER BY r.fecha_ingreso DESC,r.id_recepcion DESC
        FOR JSON PATH, INCLUDE_NULL_VALUES)) AS receptions
FOR JSON PATH, INCLUDE_NULL_VALUES, WITHOUT_ARRAY_WRAPPER)
FROM dbo.vehiculos v WHERE v.id_vehiculo=@p1`, id)
}

func diagnosisSupervisor(u domain.User) bool {
	return u.Can("diagnosticos.supervisar") || u.Can("diagnosticos.asignar")
}

func (s *Store) AuthorizeDiagnosis(ctx context.Context, id, actor string, write bool) error {
	var technician string
	if err := s.db.QueryRowContext(ctx, `SELECT CONVERT(varchar(20),id_tecnico) FROM dbo.diagnosticos WHERE id_diagnostico=@p1`, id).Scan(&technician); err != nil {
		return translate(err)
	}
	u, err := s.User(ctx, actor)
	if err != nil {
		return err
	}
	permission := "diagnosticos.ver"
	if write {
		permission = "diagnosticos.crear"
	}
	if !u.Active || !u.Can(permission) {
		return domain.ErrForbidden
	}
	if technician == actor || diagnosisSupervisor(u) || u.Can("usuarios.roles") {
		return nil
	}
	var assigned bool
	err = s.db.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(
SELECT 1 FROM dbo.cotizaciones q JOIN dbo.ordenes_trabajo o ON o.id_cotizacion=q.id_cotizacion
JOIN dbo.asignaciones_ot a ON a.id_orden=o.id_orden WHERE q.id_diagnostico=@p1
AND a.id_usuario=@p2 AND a.activo=1 AND (@p3=0 OR a.rol_en_trabajo=N'MECANICO'))
THEN 1 ELSE 0 END)`, id, actor, write).Scan(&assigned)
	if err != nil {
		return translate(err)
	}
	if !assigned {
		return domain.ErrForbidden
	}
	return nil
}
