package postgres

import (
	"context"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
)

const employeeJSON = `jsonb_build_object('id',e.id_empleado::text,'name',e.nombre,'employeeType',e.tipo_empleado,'active',e.activo,'userId',(SELECT u.id_usuario::text FROM usuarios u WHERE u.id_empleado=e.id_empleado),'username',(SELECT u.usuario FROM usuarios u WHERE u.id_empleado=e.id_empleado))`

func (s *Store) SaveEmployee(ctx context.Context, id string, v domain.EmployeeInput, actor domain.User) (any, error) {
	if err := v.Validate(id == ""); err != nil {
		return nil, err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	permission := "usuarios.editar"
	if id == "" {
		permission = "usuarios.crear"
	}
	actor, err = identityActor(ctx, tx, actor.ID, permission)
	if err != nil {
		return nil, err
	}
	if id == "" {
		active := true
		if v.Active != nil {
			active = *v.Active
		}
		err = tx.QueryRow(ctx, `INSERT INTO empleados(nombre,tipo_empleado,activo) VALUES($1,$2,$3) RETURNING id_empleado::text`, *v.Name, *v.Type, active).Scan(&id)
	} else {
		var name, kind, linked string
		var active bool
		err = tx.QueryRow(ctx, `SELECT e.nombre,e.tipo_empleado,e.activo,coalesce((SELECT u.id_usuario::text FROM usuarios u WHERE u.id_empleado=e.id_empleado),'') FROM empleados e WHERE id_empleado=$1 FOR UPDATE`, id).Scan(&name, &kind, &active, &linked)
		if err != nil {
			return nil, translate(err)
		}
		if linked != "" {
			protected, e := hasAdminAccess(ctx, tx, linked)
			if e != nil {
				return nil, e
			}
			if protected && !actor.Can("usuarios.roles") {
				return nil, domain.ErrForbidden
			}
		}
		if v.Name != nil {
			name = *v.Name
		}
		if v.Type != nil {
			kind = *v.Type
		}
		if v.Active != nil {
			active = *v.Active
		}
		_, err = tx.Exec(ctx, `UPDATE empleados SET nombre=$2,tipo_empleado=$3,activo=$4 WHERE id_empleado=$1`, id, name, kind, active)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+employeeJSON+" FROM empleados e WHERE id_empleado=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}

// Called only inside the existing serialized identity transaction. Binding is
// atomic with user creation/roles; an occupied employee rolls everything back.
func bindEmployee(ctx context.Context, tx pgx.Tx, user, employee string) error {
	var current string
	if err := tx.QueryRow(ctx, `SELECT coalesce(id_empleado::text,'') FROM usuarios WHERE id_usuario=$1 FOR UPDATE`, user).Scan(&current); err != nil {
		return translate(err)
	}
	if current == employee {
		return nil
	}
	if current != "" {
		return domain.Invalid("employeeId", "La cuenta ya está vinculada; no se reemplaza su empleado silenciosamente.")
	}
	var active bool
	if err := tx.QueryRow(ctx, `SELECT activo FROM empleados WHERE id_empleado=$1 FOR UPDATE`, employee).Scan(&active); err != nil {
		return translate(err)
	}
	if !active {
		return domain.Invalid("employeeId", "No se vincula una cuenta nueva a un empleado inactivo.")
	}
	_, err := tx.Exec(ctx, `UPDATE usuarios SET id_empleado=$2 WHERE id_usuario=$1`, user, employee)
	return translate(err)
}
