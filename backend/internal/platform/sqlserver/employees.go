package sqlserver

import (
	"context"
	"database/sql"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
)

const employeeJSON = `JSON_OBJECT('id':CONVERT(varchar(20),e.id_empleado),'name':e.nombre,'employeeType':e.tipo_empleado,'active':e.activo,
'userId':(SELECT CONVERT(varchar(20),u.id_usuario) FROM dbo.usuarios u WHERE u.id_empleado=e.id_empleado),
'username':(SELECT u.usuario FROM dbo.usuarios u WHERE u.id_empleado=e.id_empleado))`

func (s *Store) SaveEmployee(ctx context.Context, id string, v domain.EmployeeInput, actor domain.User) (any, error) {
	if err := v.Validate(id == ""); err != nil {
		return nil, err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
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
		err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.empleados_id_empleado_seq;
INSERT dbo.empleados(id_empleado,nombre,tipo_empleado,activo) VALUES(@id,@p1,@p2,@p3);
SELECT CONVERT(varchar(20),@id);`, *v.Name, *v.Type, active).Scan(&id)
	} else {
		var name, kind, linked string
		var active bool
		err = tx.QueryRowContext(ctx, `SELECT e.nombre,e.tipo_empleado,e.activo,coalesce((SELECT CONVERT(varchar(32),u.id_usuario) FROM usuarios u WITH (UPDLOCK,HOLDLOCK) WHERE u.id_empleado=e.id_empleado),'') FROM empleados e WHERE id_empleado=@p1`, id).Scan(&name, &kind, &active, &linked)
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
		_, err = tx.ExecContext(ctx, `UPDATE empleados SET nombre=@p2,tipo_empleado=@p3,activo=@p4 WHERE id_empleado=@p1`, id, name, kind, active)
	}
	if err != nil {
		return nil, translate(err)
	}
	out, err := readJSON(ctx, tx, "SELECT "+employeeJSON+" FROM empleados e WHERE id_empleado=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}

// Called only inside the existing serialized identity transaction. Binding is
// atomic with user creation/roles; an occupied employee rolls everything back.
func bindEmployee(ctx context.Context, tx *sql.Tx, user, employee string) error {
	var current string
	if err := tx.QueryRowContext(ctx, `SELECT coalesce(CONVERT(varchar(32),id_empleado),'') FROM usuarios WITH (UPDLOCK,HOLDLOCK) WHERE id_usuario=@p1`, user).Scan(&current); err != nil {
		return translate(err)
	}
	if current == employee {
		return nil
	}
	if current != "" {
		return domain.Invalid("employeeId", "La cuenta ya está vinculada; no se reemplaza su empleado silenciosamente.")
	}
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT activo FROM empleados WITH (UPDLOCK,HOLDLOCK) WHERE id_empleado=@p1`, employee).Scan(&active); err != nil {
		return translate(err)
	}
	if !active {
		return domain.Invalid("employeeId", "No se vincula una cuenta nueva a un empleado inactivo.")
	}
	_, err := tx.ExecContext(ctx, `UPDATE usuarios SET id_empleado=@p2 WHERE id_usuario=@p1`, user, employee)
	return translate(err)
}
