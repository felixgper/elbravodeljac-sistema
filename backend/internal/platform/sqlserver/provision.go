package sqlserver

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/felixgper/elbravodeljac-sistema/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

// Provision creates catalogs and the initial manager in an empty development or QA
// database. API startup never calls it.
func (s *Store) Provision(ctx context.Context, target, username, password string) error {
	if target != "elbravodeljac_sqlserver_dev" && target != "elbravodeljac_sqlserver_qa" {
		return errors.New("Confirma la base SQL Server de desarrollo o QA por su nombre exacto")
	}
	name := "Administración de pruebas"
	placeholderRole := "1"
	if len(password) < 12 {
		return errors.New("La contraseña inicial requiere al menos 12 bytes")
	}
	if err := (domain.UserInput{Name: &name, Username: &username, Password: &password, RoleID: &placeholderRole}).Validate(true); err != nil {
		return err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return translate(err)
	}
	defer tx.Rollback()
	if err := acquireLock(ctx, tx, "jac:sqlserver:provision"); err != nil {
		return err
	}

	var nameDB string
	var tables, columns, foreignKeys int
	err = tx.QueryRowContext(ctx, `SELECT DB_NAME(),
      (SELECT COUNT(*) FROM sys.tables WHERE schema_id=SCHEMA_ID(N'dbo')),
      (SELECT COUNT(*) FROM sys.columns c JOIN sys.tables t ON t.object_id=c.object_id WHERE t.schema_id=SCHEMA_ID(N'dbo')),
      (SELECT COUNT(*) FROM sys.foreign_keys WHERE schema_id=SCHEMA_ID(N'dbo') AND is_disabled=0 AND is_not_trusted=0)`).
		Scan(&nameDB, &tables, &columns, &foreignKeys)
	if err != nil {
		return translate(err)
	}
	if nameDB != target || tables != 28 || columns != 208 || foreignKeys != 50 {
		return errors.New("La base y el esquema de instalación no coinciden")
	}
	// A separate, explicit provisioning command must never augment an existing installation.
	var occupied int
	err = tx.QueryRowContext(ctx, `
SELECT CASE WHEN EXISTS(SELECT 1 FROM dbo.usuarios)
    OR EXISTS(SELECT 1 FROM dbo.roles)
    OR EXISTS(SELECT 1 FROM dbo.permisos)
    OR EXISTS(SELECT 1 FROM dbo.tipos_documento)
    OR EXISTS(SELECT 1 FROM dbo.clientes)
    OR EXISTS(SELECT 1 FROM dbo.empleados)
THEN 1 ELSE 0 END`).Scan(&occupied)
	if err != nil {
		return translate(err)
	}
	if occupied != 0 {
		return errors.New("La base SQL Server ya contiene datos")
	}

	_, err = tx.ExecContext(ctx, `INSERT dbo.tipos_documento(codigo,nombre) VALUES
(N'DNI',N'Documento Nacional de Identidad'),(N'RUC',N'Registro Único de Contribuyentes'),
(N'CE',N'Carné de Extranjería'),(N'PASAPORTE',N'Pasaporte')`)
	if err != nil {
		return translate(err)
	}
	policy := domain.PermissionPolicy()
	for module, actions := range policy.Modules {
		for _, action := range actions {
			if _, err := tx.ExecContext(ctx, `INSERT dbo.permisos(codigo) VALUES(@p1)`, module+"."+action); err != nil {
				return translate(err)
			}
		}
	}
	for role, matrix := range policy.Defaults {
		var id int64
		if err := tx.QueryRowContext(ctx, `INSERT dbo.roles(nombre_rol) OUTPUT inserted.id_rol VALUES(@p1)`, role).Scan(&id); err != nil {
			return translate(err)
		}
		codes, err := matrix.Codes()
		if err != nil {
			return err
		}
		for _, code := range codes {
			if _, err := tx.ExecContext(ctx, `INSERT dbo.rol_permisos(id_rol,id_permiso)
SELECT @p1,id_permiso FROM dbo.permisos WHERE codigo=@p2`, id, code); err != nil {
				return translate(err)
			}
		}
	}
	_, err = tx.ExecContext(ctx, `
DECLARE @user bigint = NEXT VALUE FOR dbo.usuarios_id_usuario_seq;
INSERT dbo.usuarios(id_usuario,nombre,usuario,password_hash) VALUES(@user,@p1,@p2,@p3);
INSERT dbo.usuario_roles(id_usuario,id_rol) SELECT @user,id_rol FROM dbo.roles WHERE nombre_rol=N'Gerente';`, name, strings.TrimSpace(username), string(hash))
	if err != nil {
		return translate(err)
	}
	return translate(tx.Commit())
}

var _ queryer = (*sql.Tx)(nil)
