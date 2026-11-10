package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
	"strings"
)

const userJSON = `JSON_OBJECT('id':CONVERT(varchar(20),u.id_usuario),'name':u.nombre,'employeeId':CONVERT(varchar(20),u.id_empleado),
'employeeType':(SELECT e.tipo_empleado FROM dbo.empleados e WHERE e.id_empleado=u.id_empleado),
'employeeActive':(SELECT e.activo FROM dbo.empleados e WHERE e.id_empleado=u.id_empleado),'username':u.usuario,'email':u.correo,'active':u.activo,
'roles':JSON_QUERY((SELECT CONVERT(varchar(20),r.id_rol) AS id,r.nombre_rol AS name FROM dbo.roles r JOIN dbo.usuario_roles ur ON ur.id_rol=r.id_rol
WHERE ur.id_usuario=u.id_usuario ORDER BY r.id_rol FOR JSON PATH, INCLUDE_NULL_VALUES)))`
const roleJSON = `JSON_OBJECT('id':CONVERT(varchar(20),r.id_rol),'name':r.nombre_rol,'active':r.activo,
'permissions':JSON_QUERY(COALESCE((SELECT JSON_OBJECTAGG(x.module:JSON_QUERY(x.actions))
FROM (SELECT LEFT(p.codigo,CHARINDEX(N'.',p.codigo)-1) AS module,
JSON_ARRAYAGG(SUBSTRING(p.codigo,CHARINDEX(N'.',p.codigo)+1,160) ORDER BY p.codigo) AS actions
FROM dbo.rol_permisos rp JOIN dbo.permisos p ON p.id_permiso=rp.id_permiso
WHERE rp.id_rol=r.id_rol AND CHARINDEX(N'.',p.codigo)>0
GROUP BY LEFT(p.codigo,CHARINDEX(N'.',p.codigo)-1)) x),N'{}')))`

func stringsFromRows(rows *sql.Rows) ([]string, error) {
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, translate(err)
		}
		out = append(out, value)
	}
	return out, translate(rows.Err())
}

const activeManager = `SELECT CONVERT(bit,CASE WHEN EXISTS(
SELECT 1 FROM dbo.usuarios u JOIN dbo.usuario_roles ur ON ur.id_usuario=u.id_usuario JOIN dbo.roles r ON r.id_rol=ur.id_rol
WHERE u.activo=1 AND r.activo=1 AND r.nombre_rol=N'Gerente') THEN 1 ELSE 0 END)`

func identityActor(ctx context.Context, tx *sql.Tx, id, permission string) (domain.User, error) {
	if err := acquireLock(ctx, tx, "jac:identity:v1"); err != nil {
		return domain.User{}, err
	}
	u, err := findUser(ctx, tx, "u.id_usuario=@p1", id)
	if errors.Is(err, domain.ErrNotFound) || (err == nil && (!u.Active || !u.Can(permission))) {
		return u, domain.ErrForbidden
	}
	if err != nil {
		return u, err
	}
	err = tx.QueryRowContext(ctx, activeManager).Scan(&u.ManagerRequired)
	return u, translate(err)
}

func hasAdminAccess(ctx context.Context, tx *sql.Tx, id string) (bool, error) {
	var yes bool
	err := tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN COUNT(DISTINCT p.codigo)=3 THEN 1 ELSE 0 END)
FROM dbo.usuario_roles ur JOIN dbo.rol_permisos rp ON rp.id_rol=ur.id_rol JOIN dbo.permisos p ON p.id_permiso=rp.id_permiso
WHERE ur.id_usuario=@p1 AND p.codigo IN (N'usuarios.ver',N'usuarios.roles',N'usuarios.editar')`, id).Scan(&yes)
	return yes, translate(err)
}

func preserveAdmin(ctx context.Context, tx *sql.Tx, managerRequired bool) error {
	var yes bool
	if err := tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(
SELECT 1 FROM dbo.usuarios u JOIN dbo.usuario_roles ur ON ur.id_usuario=u.id_usuario
JOIN dbo.rol_permisos rp ON rp.id_rol=ur.id_rol JOIN dbo.permisos p ON p.id_permiso=rp.id_permiso
WHERE u.activo=1 AND p.codigo IN(N'usuarios.ver',N'usuarios.roles',N'usuarios.editar')
GROUP BY u.id_usuario HAVING COUNT(DISTINCT p.codigo)=3) THEN 1 ELSE 0 END)`).Scan(&yes); err != nil {
		return translate(err)
	}
	if !yes {
		return domain.Invalid("administration", "Debe quedar al menos un usuario activo con usuarios.ver, usuarios.roles y usuarios.editar.")
	}
	if managerRequired {
		if err := tx.QueryRowContext(ctx, activeManager).Scan(&yes); err != nil {
			return translate(err)
		}
		if !yes {
			return domain.Invalid("administration", "Debe conservarse al menos un Gerente activo.")
		}
	}
	return nil
}
func grantableRole(ctx context.Context, tx *sql.Tx, id string, actor domain.User) error {
	var name string
	var active bool
	if err := tx.QueryRowContext(ctx, `SELECT nombre_rol,activo FROM roles WHERE id_rol=@p1`, id).Scan(&name, &active); err != nil {
		return translate(err)
	}
	if !active || name == "Auxiliar" || name == "Supervisor" || name == "Administrativo" || name == "Propietario" {
		return domain.Invalid("roleId", "Rol retirado: conserva sus referencias históricas, pero no admite nuevas asignaciones.")
	}
	rows, err := tx.QueryContext(ctx, `SELECT p.codigo FROM permisos p JOIN rol_permisos rp ON rp.id_permiso=p.id_permiso WHERE rp.id_rol=@p1`, id)
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
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
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
		err = tx.QueryRowContext(ctx, `DECLARE @id bigint = NEXT VALUE FOR dbo.usuarios_id_usuario_seq;
INSERT dbo.usuarios(id_usuario,nombre,usuario,correo,password_hash,activo) VALUES(@id,@p1,@p2,NULLIF(@p3,N''),@p4,@p5);
SELECT CONVERT(varchar(20),@id);`, *v.Name, *v.Username, email, string(hash), active).Scan(&id)
		if err != nil {
			return nil, translate(err)
		}
	} else {
		var name, username, email string
		var active bool
		err = tx.QueryRowContext(ctx, `SELECT nombre,usuario,coalesce(correo,''),activo FROM usuarios WITH (UPDLOCK,HOLDLOCK) WHERE id_usuario=@p1`, id).Scan(&name, &username, &email, &active)
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
		if v.Active != nil && *v.Active && !active {
			rows, err := tx.QueryContext(ctx, `SELECT CONVERT(varchar(32),id_rol) FROM usuario_roles WHERE id_usuario=@p1`, id)
			if err != nil {
				return nil, err
			}
			roles, err := stringsFromRows(rows)
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
			if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN id_empleado IS NOT NULL THEN 1 ELSE 0 END) FROM usuarios WHERE id_usuario=@p1`, id).Scan(&linked); err != nil {
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
		_, err = tx.ExecContext(ctx, `UPDATE usuarios SET nombre=@p2,usuario=@p3,correo=nullif(@p4,''),activo=@p5,password_hash=coalesce(nullif(@p6,''),password_hash) WHERE id_usuario=@p1`, id, name, username, email, active, string(hash))
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
		if _, err = tx.ExecContext(ctx, `INSERT INTO usuario_roles(id_usuario,id_rol) SELECT @p1,@p2 WHERE NOT EXISTS(SELECT 1 FROM dbo.usuario_roles WHERE id_usuario=@p1 AND id_rol=@p2)`, id, *v.RoleID); err != nil {
			return nil, translate(err)
		}
	}
	if v.RemoveRoleID != nil {
		tag, err := tx.ExecContext(ctx, `DELETE FROM usuario_roles WHERE id_usuario=@p1 AND id_rol=@p2`, id, *v.RemoveRoleID)
		if err != nil {
			return nil, translate(err)
		}
		affected, err := tag.RowsAffected()
		if err != nil {
			return nil, translate(err)
		}
		if affected != 1 {
			return nil, domain.Invalid("removeRoleId", "La asignación que se pide retirar no existe.")
		}
	}
	if err = preserveAdmin(ctx, tx, actor.ManagerRequired); err != nil {
		return nil, err
	}
	out, err := readJSON(ctx, tx, "SELECT "+userJSON+" FROM usuarios u WHERE id_usuario=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
func (s *Store) DeleteUser(ctx context.Context, id string, actor domain.User) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
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
	if _, err = tx.ExecContext(ctx, `DELETE FROM usuario_roles WHERE id_usuario=@p1`, id); err != nil {
		return translate(err)
	}
	tag, err := tx.ExecContext(ctx, `DELETE FROM usuarios WHERE id_usuario=@p1`, id)
	if err != nil {
		return translate(err)
	}
	affected, err := tag.RowsAffected()
	if err != nil {
		return translate(err)
	}
	if affected == 0 {
		return domain.ErrNotFound
	}
	if err = preserveAdmin(ctx, tx, actor.ManagerRequired); err != nil {
		return err
	}
	return translate(tx.Commit())
}
func (s *Store) SetRolePermissions(ctx context.Context, id string, v domain.MatrixInput, reset bool, actor domain.User) (any, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	actor, err = identityActor(ctx, tx, actor.ID, "usuarios.roles")
	if err != nil {
		return nil, err
	}
	var role string
	if err = tx.QueryRowContext(ctx, `SELECT nombre_rol FROM roles WITH (UPDLOCK,HOLDLOCK) WHERE id_rol=@p1`, id).Scan(&role); err != nil {
		return nil, translate(err)
	}
	var active bool
	if err = tx.QueryRowContext(ctx, `SELECT activo FROM roles WHERE id_rol=@p1`, id).Scan(&active); err != nil {
		return nil, err
	}
	if !active {
		return nil, domain.Invalid("role", "La matriz histórica de un rol retirado no se modifica.")
	}
	var own bool
	if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM usuario_roles WHERE id_usuario=@p1 AND id_rol=@p2) THEN 1 ELSE 0 END)`, actor.ID, id).Scan(&own); err != nil {
		return nil, err
	}
	if own || role == "Gerente" || role == "Propietario" {
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
		if err = tx.QueryRowContext(ctx, `SELECT CONVERT(bit,CASE WHEN EXISTS(SELECT 1 FROM permisos WHERE codigo=@p1) THEN 1 ELSE 0 END)`, code).Scan(&exists); err != nil {
			return nil, err
		}
		if !exists {
			return nil, domain.Invalid("permissions", "El código no existe en el catálogo de permisos; no se inventará.")
		}
	}
	known := map[string]bool{}
	for module, actions := range domain.PermissionPolicy().Modules {
		for _, action := range actions {
			known[module+"."+action] = true
		}
	}
	rows, err := tx.QueryContext(ctx, `SELECT p.codigo FROM permisos p JOIN rol_permisos rp ON rp.id_permiso=p.id_permiso WHERE rp.id_rol=@p1`, id)
	if err != nil {
		return nil, err
	}
	existing, err := stringsFromRows(rows)
	if err != nil {
		return nil, err
	}
	for _, code := range existing {
		if !known[code] {
			return nil, domain.Invalid("permissions", "El rol contiene códigos no representados por la matriz aprobada; no se borrarán silenciosamente.")
		}
	}
	if _, err = tx.ExecContext(ctx, `DELETE FROM rol_permisos WHERE id_rol=@p1`, id); err != nil {
		return nil, translate(err)
	}
	for _, code := range codes {
		if _, err = tx.ExecContext(ctx, `INSERT INTO rol_permisos(id_rol,id_permiso) SELECT @p1,id_permiso FROM permisos WHERE codigo=@p2`, id, code); err != nil {
			return nil, translate(err)
		}
	}
	if err = preserveAdmin(ctx, tx, actor.ManagerRequired); err != nil {
		return nil, err
	}
	out, err := readJSON(ctx, tx, "SELECT "+roleJSON+" FROM roles r WHERE id_rol=@p1", id)
	if err != nil {
		return nil, err
	}
	return out, translate(tx.Commit())
}
