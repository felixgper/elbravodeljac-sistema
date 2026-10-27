package postgres

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

func (s *Store) AuthorizeDiagnosis(ctx context.Context, id, actor string, write bool) error {
	var technician string
	if err := s.Pool.QueryRow(ctx, `SELECT id_tecnico::text FROM diagnosticos WHERE id_diagnostico=$1`, id).Scan(&technician); err != nil {
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
	err = s.Pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM cotizaciones q JOIN ordenes_trabajo o USING(id_cotizacion) JOIN asignaciones_ot a USING(id_orden) WHERE q.id_diagnostico=$1 AND a.id_usuario=$2 AND a.activo)`, id, actor).Scan(&assigned)
	if err != nil {
		return err
	}
	if !assigned {
		return domain.ErrForbidden
	}
	return nil
}
