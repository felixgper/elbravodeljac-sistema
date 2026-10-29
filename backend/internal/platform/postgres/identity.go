package postgres

import (
	"context"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const userJSON = `jsonb_build_object('id',u.id_usuario::text,'name',u.nombre,'employeeId',u.id_empleado::text,'employeeType',(SELECT e.tipo_empleado FROM empleados e WHERE e.id_empleado=u.id_empleado),'employeeActive',(SELECT e.activo FROM empleados e WHERE e.id_empleado=u.id_empleado),'username',u.usuario,'email',u.correo,'active',u.activo,'roles',coalesce((SELECT jsonb_agg(jsonb_build_object('id',r.id_rol::text,'name',r.nombre_rol) ORDER BY r.id_rol) FROM roles r JOIN usuario_roles ur USING(id_rol) WHERE ur.id_usuario=u.id_usuario),'[]'::jsonb))`
const roleJSON = `jsonb_build_object('id',r.id_rol::text,'name',r.nombre_rol,'permissions',coalesce((SELECT jsonb_object_agg(x.module,x.actions) FROM (SELECT split_part(p.codigo,'.',1) AS module,jsonb_agg(split_part(p.codigo,'.',2) ORDER BY p.codigo) AS actions FROM rol_permisos rp JOIN permisos p USING(id_permiso) WHERE rp.id_rol=r.id_rol GROUP BY split_part(p.codigo,'.',1))x),'{}'::jsonb))`

func identityActor(ctx context.Context, tx pgx.Tx, id, permission string) (domain.User, error) {
	// Serialize every identity mutation and recheck the actor after acquiring it.
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended('jac:identity:v1',0))`); err != nil {
		return domain.User{}, err
	}
	var u domain.User
	u.ID = id
	err := tx.QueryRow(ctx, `SELECT activo,ARRAY(SELECT DISTINCT p.codigo FROM permisos p JOIN rol_permisos rp USING(id_permiso) JOIN usuario_roles ur USING(id_rol) WHERE ur.id_usuario=u.id_usuario) FROM usuarios u WHERE id_usuario=$1`, id).Scan(&u.Active, &u.Permissions)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && (!u.Active || !u.Can(permission))) {
		return u, domain.ErrForbidden
	}
	return u, err
}
func hasAdminAccess(ctx context.Context, tx pgx.Tx, id string) (bool, error) {
	var yes bool
	err := tx.QueryRow(ctx, `SELECT count(DISTINCT p.codigo)=3 FROM usuario_roles ur JOIN rol_permisos rp USING(id_rol) JOIN permisos p USING(id_permiso) WHERE ur.id_usuario=$1 AND p.codigo IN ('usuarios.ver','usuarios.roles','usuarios.editar')`, id).Scan(&yes)
	return yes, err
}
func preserveAdmin(ctx context.Context, tx pgx.Tx) error {
	var yes bool
	err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM usuarios u JOIN usuario_roles ur USING(id_usuario) JOIN rol_permisos rp USING(id_rol) JOIN permisos p USING(id_permiso) WHERE u.activo AND p.codigo IN ('usuarios.ver','usuarios.roles','usuarios.editar') GROUP BY u.id_usuario HAVING count(DISTINCT p.codigo)=3)`).Scan(&yes)
	if err != nil {
		return err
	}
	if !yes {
		return domain.Invalid("administration", "Debe quedar al menos un usuario activo con usuarios.ver, usuarios.roles y usuarios.editar.")
	}
	return nil
}
func grantableRole(ctx context.Context, tx pgx.Tx, id string, actor domain.User) error {
	var name string
	if err := tx.QueryRow(ctx, `SELECT nombre_rol FROM roles WHERE id_rol=$1`, id).Scan(&name); err != nil {
		return translate(err)
	}
	rows, err := tx.Query(ctx, `SELECT p.codigo FROM permisos p JOIN rol_permisos rp USING(id_permiso) WHERE rp.id_rol=$1`, id)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var code string
		if err = rows.Scan(&code); err != nil {
			return err
		}
		if !actor.Can(code) {
			return domain.ErrForbidden
		}
	}
	return rows.Err()
}
func (s *Store) SaveUser(ctx context.Context, id string, v domain.UserInput, actor domain.User) (any, error) {
	create := id == ""
	if err := v.Validate(create); err != nil {
		return nil, err
	}
	var hash []byte
	var err error
	if v.Password != nil {
		hash, err = bcrypt.GenerateFromPassword([]byte(*v.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	permission := "usuarios.editar"
	if create {
		permission = "usuarios.crear"
	}
	actor, err = identityActor(ctx, tx, actor.ID, permission)
	if err != nil {
		return nil, err
	}
	security := v.EmployeeID != nil || v.RoleID != nil || v.RemoveRoleID != nil || v.Password != nil || v.Active != nil
	if security && !actor.Can("usuarios.roles") {
		return nil, domain.ErrForbidden
	}
	if id == actor.ID && security {
		return nil, domain.ErrForbidden
	}
	if v.RoleID != nil {
		if err = grantableRole(ctx, tx, *v.RoleID, actor); err != nil {
			return nil, err
		}
	}
	if create {
		email := ""
		if v.Email != nil {
			email = strings.ToLower(strings.TrimSpace(*v.Email))
		}
		active := true
		if v.Active != nil {
			active = *v.Active
		}
		err = tx.QueryRow(ctx, `INSERT INTO usuarios(nombre,usuario,correo,password_hash,activo) VALUES($1,$2,nullif($3,''),$4,$5) RETURNING id_usuario::text`, *v.Name, *v.Username, email, string(hash), active).Scan(&id)
		if err != nil {
			return nil, translate(err)
		}
	} else {
		var name, username, email string
		var active bool
		err = tx.QueryRow(ctx, `SELECT nombre,usuario,coalesce(correo,''),activo FROM usuarios WHERE id_usuario=$1 FOR UPDATE`, id).Scan(&name, &username, &email, &active)
		if err != nil {
			return nil, translate(err)
		}
		protected, err := hasAdminAccess(ctx, tx, id)
		if err != nil {
			return nil, err
		}
		if protected && !actor.Can("usuarios.roles") {
			return nil, domain.ErrForbidden
		}
		if v.Active != nil && *v.Active {
			rows, err := tx.Query(ctx, `SELECT id_rol::text FROM usuario_roles WHERE id_usuario=$1`, id)
			if err != nil {
				return nil, err
			}
			roles, err := pgx.CollectRows(rows, pgx.RowTo[string])
			if err != nil {
				return nil, err
			}
			for _, role := range roles {
				if err = grantableRole(ctx, tx, role, actor); err != nil {
					return nil, err
				}
			}
		}
		if v.Name != nil {
			var linked bool
			if err = tx.QueryRow(ctx, `SELECT id_empleado IS NOT NULL FROM usuarios WHERE id_usuario=$1`, id).Scan(&linked); err != nil {
				return nil, err
			}
			if linked && *v.Name != name {
				return nil, domain.Invalid("name", "Modifica el nombre principal desde Empleados.")
			}
			name = *v.Name
		}
		if v.Username != nil {
			username = *v.Username
		}
		if v.Email != nil {
			email = strings.ToLower(strings.TrimSpace(*v.Email))
		}
		if v.Active != nil {
			active = *v.Active
		}
		_, err = tx.Exec(ctx, `UPDATE usuarios SET nombre=$2,usuario=$3,correo=nullif($4,''),activo=$5,password_hash=coalesce(nullif($6,''),password_hash) WHERE id_usuario=$1`, id, name, username, email, active, string(hash))
		if err != nil {
			return nil, translate(err)
		}
	}
	if v.EmployeeID != nil {
		if err = bindEmployee(ctx, tx, id, *v.EmployeeID); err != nil {
			return nil, err
		}
	}
	if v.RoleID != nil {
		if _, err = tx.Exec(ctx, `INSERT INTO usuario_roles(id_usuario,id_rol) VALUES($1,$2) ON CONFLICT DO NOTHING`, id, *v.RoleID); err != nil {
			return nil, translate(err)
		}
	}
	if v.RemoveRoleID != nil {
		tag, err := tx.Exec(ctx, `DELETE FROM usuario_roles WHERE id_usuario=$1 AND id_rol=$2`, id, *v.RemoveRoleID)
		if err != nil {
			return nil, translate(err)
		}
		if tag.RowsAffected() != 1 {
			return nil, domain.Invalid("removeRoleId", "La asignación que se pide retirar no existe.")
		}
	}
	if err = preserveAdmin(ctx, tx); err != nil {
		return nil, err
	}
	out, err := readJSON(ctx, tx, "SELECT "+userJSON+" FROM usuarios u WHERE id_usuario=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
func (s *Store) DeleteUser(ctx context.Context, id string, actor domain.User) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	actor, err = identityActor(ctx, tx, actor.ID, "usuarios.eliminar")
	if err != nil {
		return err
	}
	if actor.ID == id {
		return domain.ErrForbidden
	}
	protected, err := hasAdminAccess(ctx, tx, id)
	if err != nil {
		return err
	}
	if protected && !actor.Can("usuarios.roles") {
		return domain.ErrForbidden
	}
	if _, err = tx.Exec(ctx, `DELETE FROM usuario_roles WHERE id_usuario=$1`, id); err != nil {
		return translate(err)
	}
	tag, err := tx.Exec(ctx, `DELETE FROM usuarios WHERE id_usuario=$1`, id)
	if err != nil {
		return translate(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	if err = preserveAdmin(ctx, tx); err != nil {
		return err
	}
	return translate(tx.Commit(ctx))
}
func (s *Store) SetRolePermissions(ctx context.Context, id string, v domain.MatrixInput, reset bool, actor domain.User) (any, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)
	actor, err = identityActor(ctx, tx, actor.ID, "usuarios.roles")
	if err != nil {
		return nil, err
	}
	var role string
	if err = tx.QueryRow(ctx, `SELECT nombre_rol FROM roles WHERE id_rol=$1 FOR UPDATE`, id).Scan(&role); err != nil {
		return nil, translate(err)
	}
	var own bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM usuario_roles WHERE id_usuario=$1 AND id_rol=$2)`, actor.ID, id).Scan(&own); err != nil {
		return nil, err
	}
	if own || role == "Propietario" {
		return nil, domain.ErrForbidden
	}
	if reset {
		var ok bool
		v.Permissions, ok = domain.PermissionPolicy().Defaults[role]
		if !ok {
			return nil, domain.Invalid("role", "No hay política inicial aprobada para este rol.")
		}
	}
	if v.Permissions == nil {
		return nil, domain.Invalid("permissions", "Envía la matriz completa; {} retira los permisos.")
	}
	codes, err := v.Permissions.Codes()
	if err != nil {
		return nil, err
	}
	for _, code := range codes {
		if !actor.Can(code) {
			return nil, domain.ErrForbidden
		}
		var exists bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM permisos WHERE codigo=$1)`, code).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.Invalid("permissions", "El código no existe en el catálogo PostgreSQL; no se inventará.")
		}
	}
	if role == "Administrador" {
		found := map[string]bool{}
		for _, code := range codes {
			found[code] = true
		}
		if !found["usuarios.ver"] || !found["usuarios.roles"] || !found["usuarios.editar"] {
			return nil, domain.Invalid("permissions", "Administrador conserva los permisos mínimos de gestión.")
		}
	}
	known := map[string]bool{}
	for module, actions := range domain.PermissionPolicy().Modules {
		for _, action := range actions {
			known[module+"."+action] = true
		}
	}
	rows, err := tx.Query(ctx, `SELECT p.codigo FROM permisos p JOIN rol_permisos rp USING(id_permiso) WHERE rp.id_rol=$1`, id)
	if err != nil {
		return nil, err
	}
	existing, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return nil, err
	}
	for _, code := range existing {
		if !known[code] {
			return nil, domain.Invalid("permissions", "El rol contiene códigos no representados por la matriz aprobada; no se borrarán silenciosamente.")
		}
	}
	if _, err = tx.Exec(ctx, `DELETE FROM rol_permisos WHERE id_rol=$1`, id); err != nil {
		return nil, translate(err)
	}
	for _, code := range codes {
		if _, err = tx.Exec(ctx, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT $1,id_permiso FROM permisos WHERE codigo=$2`, id, code); err != nil {
			return nil, translate(err)
		}
	}
	if err = preserveAdmin(ctx, tx); err != nil {
		return nil, err
	}
	out, err := readJSON(ctx, tx, "SELECT "+roleJSON+" FROM roles r WHERE id_rol=$1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit(ctx))
}
